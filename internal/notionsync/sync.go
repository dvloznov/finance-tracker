package notionsync

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dvloznov/finance-tracker/internal/infra/bigquery"
	"github.com/dvloznov/finance-tracker/internal/logger"
	"github.com/jomei/notionapi"
)

const (
	// BatchSize defines the number of transactions to process in a single batch
	BatchSize = 100
)

// SyncTransactions syncs transactions from BigQuery to Notion within the specified date range.
// It queries BigQuery for transactions, batches them, and creates/updates corresponding Notion pages.
// The external_reference field on transactions is used to track Notion page IDs for idempotency.
// Deprecated: Use SyncTransactionsWithCategories instead.
func SyncTransactions(ctx context.Context, repo bigquery.DocumentRepository, notionClient NotionService, notionDBID string, startDate, endDate time.Time, dryRun bool) error {
	return SyncTransactionsWithCategories(ctx, repo, notionClient, notionDBID, startDate, endDate, nil, dryRun)
}

// SyncTransactionsWithCategories syncs transactions from BigQuery to Notion with category relations.
// categoryPageIDs maps category_id -> Notion page ID for creating category relations.
// This function:
// 1. Queries all existing Notion transactions
// 2. Deletes stale transactions (not in BigQuery active set)
// 3. Creates/updates current transactions from BigQuery
func SyncTransactionsWithCategories(ctx context.Context, repo bigquery.DocumentRepository, notionClient NotionService, notionDBID string, startDate, endDate time.Time, categoryPageIDs map[string]string, dryRun bool) error {
	log := logger.FromContext(ctx)

	log.Info().
		Time("start_date", startDate).
		Time("end_date", endDate).
		Bool("dry_run", dryRun).
		Int("category_mappings", len(categoryPageIDs)).
		Msg("Starting transaction sync to Notion")

	// Query transactions from BigQuery (already filtered to active parsing runs only)
	transactions, err := repo.QueryTransactionsByDateRange(ctx, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to query transactions: %w", err)
	}

	log.Info().Int("transaction_count", len(transactions)).Msg("Retrieved transactions from BigQuery")

	// Build set of valid transaction IDs from BigQuery
	validTransactionIDs := make(map[string]bool)
	for _, tx := range transactions {
		validTransactionIDs[tx.TransactionID] = true
	}

	// Query all existing transactions from Notion
	log.Info().Msg("Querying existing transactions from Notion")
	notionPages, err := queryAllNotionPages(ctx, notionClient, notionDBID)
	if err != nil {
		return fmt.Errorf("failed to query Notion pages: %w", err)
	}

	log.Info().Int("notion_page_count", len(notionPages)).Msg("Retrieved existing Notion pages")

	// Build map of existing transaction IDs in Notion (for deduplication)
	existingTransactionIDs := make(map[string]bool)
	for _, page := range notionPages {
		txID := extractTransactionID(page)
		if txID != "" {
			existingTransactionIDs[txID] = true
		}
	}

	// Delete stale transactions from Notion (those not in the valid set)
	var deleted int
	for _, page := range notionPages {
		txID := extractTransactionID(page)
		
		// Delete pages without Transaction ID (from old sync) or not in valid set
		if txID == "" || !validTransactionIDs[txID] {
			if dryRun {
				log.Info().
					Str("transaction_id", txID).
					Str("page_id", string(page.ID)).
					Msg("[DRY RUN] Would delete stale Notion page")
				deleted++
			} else {
				if err := notionClient.DeletePage(ctx, string(page.ID)); err != nil {
					log.Warn().
						Err(err).
						Str("transaction_id", txID).
						Str("page_id", string(page.ID)).
						Msg("Failed to delete stale Notion page")
					continue
				}
				log.Info().
					Str("transaction_id", txID).
					Str("page_id", string(page.ID)).
					Msg("Deleted stale Notion page")
				deleted++
			}
		}
	}

	if deleted > 0 {
		log.Info().Int("deleted", deleted).Msg("Deleted stale transactions from Notion")
	}

	// Process transactions in batches
	var created, updated, skipped int
	for i := 0; i < len(transactions); i += BatchSize {
		end := i + BatchSize
		if end > len(transactions) {
			end = len(transactions)
		}

		batch := transactions[i:end]
		log.Info().
			Int("batch_start", i).
			Int("batch_end", end).
			Int("batch_size", len(batch)).
			Msg("Processing batch")

		for _, tx := range batch {
			// Skip if already exists in Notion
			if existingTransactionIDs[tx.TransactionID] {
				skipped++
				continue
			}

			// Check if transaction already has a Notion page ID
			existingPageID := GetNotionPageIDFromTransaction(tx)

			if dryRun {
				if existingPageID != "" {
					log.Info().
						Str("transaction_id", tx.TransactionID).
						Str("existing_page_id", extractPageID(existingPageID)).
						Msg("[DRY RUN] Would update existing Notion page")
					updated++
				} else {
					log.Info().
						Str("transaction_id", tx.TransactionID).
						Msg("[DRY RUN] Would create new Notion page")
					created++
				}
				continue
			}

			// Convert transaction to Notion properties
			props := TransactionToNotionPropertiesWithCategories(tx, categoryPageIDs)

			if existingPageID != "" {
				// Update existing page
				pageID := extractPageID(existingPageID)
				_, err := notionClient.UpdatePage(ctx, pageID, props)
				if err != nil {
					log.Warn().
						Err(err).
						Str("transaction_id", tx.TransactionID).
						Str("page_id", pageID).
						Msg("Failed to update Notion page")
					// Continue processing other transactions
					continue
				}
				log.Info().
					Str("transaction_id", tx.TransactionID).
					Str("page_id", pageID).
					Msg("Updated Notion page")
				updated++
			} else {
				// Create new page
				page, err := notionClient.CreatePage(ctx, notionDBID, props)
				if err != nil {
					log.Warn().
						Err(err).
						Str("transaction_id", tx.TransactionID).
						Msg("Failed to create Notion page")
					// Continue processing other transactions
					continue
				}
				log.Info().
					Str("transaction_id", tx.TransactionID).
					Str("page_id", string(page.ID)).
					Msg("Created Notion page")
				created++

				// Note: In a production system, we would update the external_reference field
				// in BigQuery here to store the Notion page ID for future syncs.
				// This is omitted for simplicity in this implementation.
			}
		}
	}

	log.Info().
		Int("deleted", deleted).
		Int("created", created).
		Int("updated", updated).
		Int("total", len(transactions)).
		Msg("Transaction sync completed")

	return nil
}

// extractPageID extracts the page ID from the external_reference format "notion:page_id"
func extractPageID(externalRef string) string {
	if strings.HasPrefix(externalRef, "notion:") {
		return strings.TrimPrefix(externalRef, "notion:")
	}
	return externalRef
}

// SyncAccounts syncs all accounts from BigQuery to Notion.
// Deletes stale accounts and creates/updates current ones.
func SyncAccounts(ctx context.Context, repo bigquery.DocumentRepository, notionClient NotionService, notionDBID string, dryRun bool) error {
	log := logger.FromContext(ctx)

	log.Info().
		Bool("dry_run", dryRun).
		Msg("Starting accounts sync to Notion")

	// Query all accounts from BigQuery
	accounts, err := repo.ListAllAccounts(ctx)
	if err != nil {
		return fmt.Errorf("failed to query accounts: %w", err)
	}

	log.Info().Int("account_count", len(accounts)).Msg("Retrieved accounts from BigQuery")

	// Build set of valid account IDs from BigQuery
	validAccountIDs := make(map[string]bool)
	for _, acc := range accounts {
		validAccountIDs[acc.AccountID] = true
	}

	// Query all existing accounts from Notion
	log.Info().Msg("Querying existing accounts from Notion")
	notionPages, err := queryAllNotionPages(ctx, notionClient, notionDBID)
	if err != nil {
		return fmt.Errorf("failed to query Notion pages: %w", err)
	}

	log.Info().Int("notion_page_count", len(notionPages)).Msg("Retrieved existing Notion pages")

	// Build map of existing account IDs in Notion (for deduplication)
	existingAccountIDs := make(map[string]bool)
	for _, page := range notionPages {
		accID := extractAccountID(page)
		if accID != "" {
			existingAccountIDs[accID] = true
		}
	}

	// Delete stale accounts from Notion
	var deleted int
	for _, page := range notionPages {
		accID := extractAccountID(page)
		
		// Delete pages without Account ID (from old sync) or not in valid set
		if accID == "" || !validAccountIDs[accID] {
			if dryRun {
				log.Info().
					Str("account_id", accID).
					Str("page_id", string(page.ID)).
					Msg("[DRY RUN] Would delete stale Notion page")
				deleted++
			} else {
				if err := notionClient.DeletePage(ctx, string(page.ID)); err != nil {
					log.Warn().
						Err(err).
						Str("account_id", accID).
						Str("page_id", string(page.ID)).
						Msg("Failed to delete stale Notion page")
					continue
				}
				log.Info().
					Str("account_id", accID).
					Str("page_id", string(page.ID)).
					Msg("Deleted stale Notion page")
				deleted++
			}
		}
	}

	// Sync accounts
	var created, skipped int
	for _, acc := range accounts {
		// Skip if already exists in Notion
		if existingAccountIDs[acc.AccountID] {
			skipped++
			continue
		}

		if dryRun {
			log.Info().
				Str("account_id", acc.AccountID).
				Msg("[DRY RUN] Would create/update Notion page for account")
			created++
			continue
		}

		// Convert account to Notion properties
		props := AccountToNotionProperties(acc)

		// Create new page
		page, err := notionClient.CreatePage(ctx, notionDBID, props)
		if err != nil {
			log.Warn().
				Err(err).
				Str("account_id", acc.AccountID).
				Msg("Failed to create Notion page for account")
			continue
		}

		log.Info().
			Str("account_id", acc.AccountID).
			Str("page_id", string(page.ID)).
			Msg("Created Notion page for account")
		created++
	}

	log.Info().
		Int("deleted", deleted).
		Int("created", created).
		Int("total", len(accounts)).
		Msg("Accounts sync completed")

	return nil
}

// SyncCategories syncs all active categories from BigQuery to Notion.
// Deletes stale categories and creates current ones.
// Returns a map of category_id -> Notion page ID for use in transaction sync.
func SyncCategories(ctx context.Context, repo bigquery.DocumentRepository, notionClient NotionService, notionDBID string, dryRun bool) (map[string]string, error) {
	log := logger.FromContext(ctx)

	log.Info().
		Bool("dry_run", dryRun).
		Msg("Starting categories sync to Notion")

	// Query all active categories from BigQuery
	categories, err := repo.ListActiveCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}

	log.Info().Int("category_count", len(categories)).Msg("Retrieved categories from BigQuery")

	// Build set of valid category slugs from BigQuery
	validCategorySlugs := make(map[string]bool)
	for _, cat := range categories {
		validCategorySlugs[cat.Slug] = true
	}

	// Query all existing categories from Notion
	log.Info().Msg("Querying existing categories from Notion")
	notionPages, err := queryAllNotionPages(ctx, notionClient, notionDBID)
	if err != nil {
		return nil, fmt.Errorf("failed to query Notion pages: %w", err)
	}

	log.Info().Int("notion_page_count", len(notionPages)).Msg("Retrieved existing Notion pages")

	// Build map of existing slugs in Notion (for deduplication)
	existingSlugs := make(map[string]bool)
	for _, page := range notionPages {
		slug := extractCategorySlug(page)
		if slug != "" {
			existingSlugs[slug] = true
		}
	}

	// Delete stale categories from Notion
	var deleted int
	for _, page := range notionPages {
		slug := extractCategorySlug(page)
		
		// Delete pages without Slug (from old sync) or not in valid set
		if slug == "" || !validCategorySlugs[slug] {
			if dryRun {
				log.Info().
					Str("slug", slug).
					Str("page_id", string(page.ID)).
					Msg("[DRY RUN] Would delete stale Notion page")
				deleted++
			} else {
				if err := notionClient.DeletePage(ctx, string(page.ID)); err != nil {
					log.Warn().
						Err(err).
						Str("slug", slug).
						Str("page_id", string(page.ID)).
						Msg("Failed to delete stale Notion page")
					continue
				}
				log.Info().
					Str("slug", slug).
					Str("page_id", string(page.ID)).
					Msg("Deleted stale Notion page")
				deleted++
			}
		}
	}

	// Map to track category_id -> Notion page ID
	categoryPageIDs := make(map[string]string)

	var created, skipped int
	for _, cat := range categories {
		// Skip if already exists in Notion
		if existingSlugs[cat.Slug] {
			skipped++
			continue
		}

		if dryRun {
			log.Info().
				Str("category_id", cat.CategoryID).
				Str("category_name", cat.CategoryName).
				Msg("[DRY RUN] Would create Notion page for category")
			created++
			continue
		}

		// Convert category to Notion properties
		props := CategoryToNotionProperties(&cat)

		// Create the page
		page, err := notionClient.CreatePage(ctx, notionDBID, props)
		if err != nil {
			log.Warn().
				Err(err).
				Str("category_id", cat.CategoryID).
				Str("category_name", cat.CategoryName).
				Msg("Failed to create Notion page for category")
			continue
		}

		// Store the mapping
		categoryPageIDs[cat.CategoryID] = string(page.ID)

		subcatInfo := ""
		if cat.SubcategoryName.Valid && cat.SubcategoryName.StringVal != "" {
			subcatInfo = " -> " + cat.SubcategoryName.StringVal
		}

		log.Info().
			Str("category_id", cat.CategoryID).
			Str("category_name", cat.CategoryName+subcatInfo).
			Str("page_id", string(page.ID)).
			Msg("Created Notion page for category")
		created++
	}

	log.Info().
		Int("deleted", deleted).
		Int("created", created).
		Int("skipped", skipped).
		Int("total", len(categories)).
		Msg("Categories sync completed")

	return categoryPageIDs, nil
}

// SyncDocuments syncs all documents from BigQuery to Notion.
// Deletes stale documents and creates/updates current ones.
func SyncDocuments(ctx context.Context, repo bigquery.DocumentRepository, notionClient NotionService, notionDBID string, dryRun bool) error {
	log := logger.FromContext(ctx)

	log.Info().
		Bool("dry_run", dryRun).
		Msg("Starting documents sync to Notion")

	// Query all documents from BigQuery
	documents, err := repo.ListAllDocuments(ctx)
	if err != nil {
		return fmt.Errorf("failed to query documents: %w", err)
	}

	log.Info().Int("document_count", len(documents)).Msg("Retrieved documents from BigQuery")

	// Build set of valid document IDs from BigQuery
	validDocumentIDs := make(map[string]bool)
	for _, doc := range documents {
		validDocumentIDs[doc.DocumentID] = true
	}

	// Query all existing documents from Notion
	log.Info().Msg("Querying existing documents from Notion")
	notionPages, err := queryAllNotionPages(ctx, notionClient, notionDBID)
	if err != nil {
		return fmt.Errorf("failed to query Notion pages: %w", err)
	}

	log.Info().Int("notion_page_count", len(notionPages)).Msg("Retrieved existing Notion pages")

	// Build map of existing document IDs in Notion (for deduplication)
	existingDocumentIDs := make(map[string]bool)
	for _, page := range notionPages {
		docID := extractDocumentID(page)
		if docID != "" {
			existingDocumentIDs[docID] = true
		}
	}

	// Delete stale documents from Notion
	var deleted int
	for _, page := range notionPages {
		docID := extractDocumentID(page)
		
		// Delete pages without Document ID (from old sync) or not in valid set
		if docID == "" || !validDocumentIDs[docID] {
			if dryRun {
				log.Info().
					Str("document_id", docID).
					Str("page_id", string(page.ID)).
					Msg("[DRY RUN] Would delete stale Notion page")
				deleted++
			} else {
				if err := notionClient.DeletePage(ctx, string(page.ID)); err != nil {
					log.Warn().
						Err(err).
						Str("document_id", docID).
						Str("page_id", string(page.ID)).
						Msg("Failed to delete stale Notion page")
					continue
				}
				log.Info().
					Str("document_id", docID).
					Str("page_id", string(page.ID)).
					Msg("Deleted stale Notion page")
				deleted++
			}
		}
	}

	// Sync documents
	var created, skipped int
	for _, doc := range documents {
		// Skip if already exists in Notion
		if existingDocumentIDs[doc.DocumentID] {
			skipped++
			continue
		}

		if dryRun {
			log.Info().
				Str("document_id", doc.DocumentID).
				Msg("[DRY RUN] Would create/update Notion page for document")
			created++
			continue
		}

		// Convert document to Notion properties
		props := DocumentToNotionProperties(doc)

		// Create new page
		page, err := notionClient.CreatePage(ctx, notionDBID, props)
		if err != nil {
			log.Warn().
				Err(err).
				Str("document_id", doc.DocumentID).
				Msg("Failed to create Notion page for document")
			continue
		}

		// Update document status to SYNCED in BigQuery
		if err := bigquery.UpdateDocumentParsingStatus(ctx, doc.DocumentID, "SYNCED"); err != nil {
			log.Warn().
				Err(err).
				Str("document_id", doc.DocumentID).
				Msg("Failed to update document status to SYNCED in BigQuery")
			// Don't fail the sync, just log the warning
		} else {
			// Also update the Notion page's Processing Status
			updateProps := notionapi.Properties{
				"Processing Status": notionapi.SelectProperty{
					Select: notionapi.Option{
						Name: "SYNCED",
					},
				},
			}
			if _, err := notionClient.UpdatePage(ctx, string(page.ID), updateProps); err != nil {
				log.Warn().
					Err(err).
					Str("document_id", doc.DocumentID).
					Str("page_id", string(page.ID)).
					Msg("Failed to update Processing Status in Notion")
			}
		}

		log.Info().
			Str("document_id", doc.DocumentID).
			Str("page_id", string(page.ID)).
			Msg("Created Notion page for document")
		created++
	}

	log.Info().
		Int("created", created).
		Int("deleted", deleted).
		Int("skipped", skipped).
		Int("total", len(documents)).
		Msg("Documents sync completed")

	return nil
}

// queryAllNotionPages queries all pages from a Notion database and returns them.
// Handles pagination automatically.
func queryAllNotionPages(ctx context.Context, notionClient NotionService, databaseID string) ([]notionapi.Page, error) {
	var allPages []notionapi.Page
	var cursor notionapi.Cursor

	for {
		req := &notionapi.DatabaseQueryRequest{
			PageSize: 100,
		}

		// Only set StartCursor if we have a cursor value
		if cursor != "" {
			req.StartCursor = cursor
		}

		resp, err := notionClient.QueryDatabase(ctx, databaseID, req)
		if err != nil {
			return nil, fmt.Errorf("queryAllNotionPages: %w", err)
		}

		allPages = append(allPages, resp.Results...)

		if !resp.HasMore {
			break
		}
		cursor = resp.NextCursor
	}

	return allPages, nil
}

// extractTransactionID extracts the transaction ID from a Notion page's properties.
// Returns empty string if not found.
func extractTransactionID(page notionapi.Page) string {
	// Check if there's a Transaction ID property
	if prop, ok := page.Properties["Transaction ID"]; ok {
		if richText, ok := prop.(*notionapi.RichTextProperty); ok {
			if len(richText.RichText) > 0 {
				return richText.RichText[0].PlainText
			}
		}
	}
	return ""
}

// extractAccountID extracts the account ID from a Notion page's properties.
// Returns empty string if not found.
func extractAccountID(page notionapi.Page) string {
	if prop, ok := page.Properties["Account ID"]; ok {
		if title, ok := prop.(*notionapi.TitleProperty); ok {
			if len(title.Title) > 0 {
				return title.Title[0].PlainText
			}
		}
	}
	return ""
}

// extractCategorySlug extracts the category slug from a Notion page's properties.
// Returns empty string if not found.
func extractCategorySlug(page notionapi.Page) string {
	if prop, ok := page.Properties["Slug"]; ok {
		if richText, ok := prop.(*notionapi.RichTextProperty); ok {
			if len(richText.RichText) > 0 {
				return richText.RichText[0].PlainText
			}
		}
	}
	return ""
}

// extractDocumentID extracts the document ID from a Notion page's properties.
// Returns empty string if not found.
func extractDocumentID(page notionapi.Page) string {
	if prop, ok := page.Properties["Document ID"]; ok {
		if title, ok := prop.(*notionapi.TitleProperty); ok {
			if len(title.Title) > 0 {
				return title.Title[0].PlainText
			}
		}
	}
	return ""
}
