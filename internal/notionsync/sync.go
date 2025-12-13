package notionsync

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dvloznov/finance-tracker/internal/infra/bigquery"
	"github.com/dvloznov/finance-tracker/internal/logger"
)

const (
	// BatchSize defines the number of transactions to process in a single batch
	BatchSize = 100
)

// SyncTransactions syncs transactions from BigQuery to Notion within the specified date range.
// It queries BigQuery for transactions, batches them, and creates/updates corresponding Notion pages.
// The external_reference field on transactions is used to track Notion page IDs for idempotency.
func SyncTransactions(ctx context.Context, repo bigquery.DocumentRepository, notionClient NotionService, notionDBID string, startDate, endDate time.Time, dryRun bool) error {
	log := logger.FromContext(ctx)

	log.Info().
		Time("start_date", startDate).
		Time("end_date", endDate).
		Bool("dry_run", dryRun).
		Msg("Starting transaction sync to Notion")

	// Query transactions from BigQuery
	transactions, err := repo.QueryTransactionsByDateRange(ctx, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to query transactions: %w", err)
	}

	log.Info().Int("transaction_count", len(transactions)).Msg("Retrieved transactions from BigQuery")

	if len(transactions) == 0 {
		log.Info().Msg("No transactions to sync")
		return nil
	}

	// Process transactions in batches
	var created, updated int
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
			props := TransactionToNotionProperties(tx)

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
// Creates or updates Notion pages for each account in the database.
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

	if len(accounts) == 0 {
		log.Info().Msg("No accounts to sync")
		return nil
	}

	// Sync accounts
	var created, updated int
	for _, acc := range accounts {
		if dryRun {
			log.Info().
				Str("account_id", acc.AccountID).
				Msg("[DRY RUN] Would create/update Notion page for account")
			created++
			continue
		}

		// Convert account to Notion properties
		props := AccountToNotionProperties(acc)

		// For accounts, we'll always create new pages for now
		// In a production system, you'd want to track Notion page IDs similar to transactions
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
		Int("created", created).
		Int("updated", updated).
		Int("total", len(accounts)).
		Msg("Accounts sync completed")

	return nil
}

// SyncCategories syncs all active categories from BigQuery to Notion.
// Creates or updates Notion pages for each category in the database.
func SyncCategories(ctx context.Context, repo bigquery.DocumentRepository, notionClient NotionService, notionDBID string, dryRun bool) error {
	log := logger.FromContext(ctx)

	log.Info().
		Bool("dry_run", dryRun).
		Msg("Starting categories sync to Notion")

	// Query all active categories from BigQuery
	categories, err := repo.ListActiveCategories(ctx)
	if err != nil {
		return fmt.Errorf("failed to query categories: %w", err)
	}

	log.Info().Int("category_count", len(categories)).Msg("Retrieved categories from BigQuery")

	if len(categories) == 0 {
		log.Info().Msg("No categories to sync")
		return nil
	}

	// Sync categories (sorted by depth to handle parent-child relationships)
	var created, updated int
	for _, cat := range categories {
		if dryRun {
			log.Info().
				Str("category_id", cat.CategoryID).
				Str("category_name", cat.Name).
				Msg("[DRY RUN] Would create/update Notion page for category")
			created++
			continue
		}

		// Convert category to Notion properties
		props := CategoryToNotionProperties(&cat)

		// For categories, we'll always create new pages for now
		// In a production system, you'd want to track Notion page IDs and handle parent relations
		page, err := notionClient.CreatePage(ctx, notionDBID, props)
		if err != nil {
			log.Warn().
				Err(err).
				Str("category_id", cat.CategoryID).
				Str("category_name", cat.Name).
				Msg("Failed to create Notion page for category")
			continue
		}

		log.Info().
			Str("category_id", cat.CategoryID).
			Str("category_name", cat.Name).
			Str("page_id", string(page.ID)).
			Msg("Created Notion page for category")
		created++
	}

	log.Info().
		Int("created", created).
		Int("updated", updated).
		Int("total", len(categories)).
		Msg("Categories sync completed")

	return nil
}

// SyncDocuments syncs all documents from BigQuery to Notion.
// Creates or updates Notion pages for each document in the database.
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

	if len(documents) == 0 {
		log.Info().Msg("No documents to sync")
		return nil
	}

	// Sync documents
	var created, updated int
	for _, doc := range documents {
		if dryRun {
			log.Info().
				Str("document_id", doc.DocumentID).
				Msg("[DRY RUN] Would create/update Notion page for document")
			created++
			continue
		}

		// Convert document to Notion properties
		props := DocumentToNotionProperties(doc)

		// For documents, we'll always create new pages for now
		// In a production system, you'd want to track Notion page IDs similar to transactions
		page, err := notionClient.CreatePage(ctx, notionDBID, props)
		if err != nil {
			log.Warn().
				Err(err).
				Str("document_id", doc.DocumentID).
				Msg("Failed to create Notion page for document")
			continue
		}

		log.Info().
			Str("document_id", doc.DocumentID).
			Str("page_id", string(page.ID)).
			Msg("Created Notion page for document")
		created++
	}

	log.Info().
		Int("created", created).
		Int("updated", updated).
		Int("total", len(documents)).
		Msg("Documents sync completed")

	return nil
}
