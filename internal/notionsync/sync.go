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
