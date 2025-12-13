package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/dvloznov/finance-tracker/internal/infra/bigquery"
	"github.com/dvloznov/finance-tracker/internal/logger"
	"github.com/dvloznov/finance-tracker/internal/notionsync"
)

func main() {
	// Initialize structured logger
	log := logger.New()

	// Parse CLI flags
	syncType := flag.String("type", "transactions", "Type of data to sync: transactions, accounts, categories, documents, or all")
	startDateStr := flag.String("start-date", "", "Start date in YYYY-MM-DD format (required for transactions sync)")
	endDateStr := flag.String("end-date", "", "End date in YYYY-MM-DD format (required for transactions sync)")
	notionToken := flag.String("notion-token", "", "Notion API token (required)")
	notionTransactionsDBID := flag.String("notion-transactions-db-id", "", "Notion transactions database ID")
	notionAccountsDBID := flag.String("notion-accounts-db-id", "", "Notion accounts database ID")
	notionCategoriesDBID := flag.String("notion-categories-db-id", "", "Notion categories database ID")
	notionDocumentsDBID := flag.String("notion-documents-db-id", "", "Notion documents database ID")
	dryRun := flag.Bool("dry-run", false, "Dry run mode - preview changes without syncing")
	flag.Parse()

	// Validate required flags
	if *notionToken == "" {
		log.Fatal().Msg("Error: --notion-token is required")
	}

	// Validate sync type
	validTypes := map[string]bool{
		"transactions": true,
		"accounts":     true,
		"categories":   true,
		"documents":    true,
		"all":          true,
	}
	if !validTypes[*syncType] {
		log.Fatal().Str("sync_type", *syncType).Msg("Error: invalid --type. Must be one of: transactions, accounts, categories, documents, all")
	}

	// Validate date parameters for transactions sync
	if (*syncType == "transactions" || *syncType == "all") && (*startDateStr == "" || *endDateStr == "") {
		log.Fatal().Msg("Error: --start-date and --end-date are required for transactions sync")
	}

	// Parse dates if provided
	var startDate, endDate time.Time
	var err error
	if *startDateStr != "" && *endDateStr != "" {
		startDate, err = time.Parse("2006-01-02", *startDateStr)
		if err != nil {
			log.Fatal().Err(err).Str("start_date", *startDateStr).Msg("Error: invalid start-date format, expected YYYY-MM-DD")
		}

		endDate, err = time.Parse("2006-01-02", *endDateStr)
		if err != nil {
			log.Fatal().Err(err).Str("end_date", *endDateStr).Msg("Error: invalid end-date format, expected YYYY-MM-DD")
		}

		// Validate date range
		if endDate.Before(startDate) {
			log.Fatal().
				Time("start_date", startDate).
				Time("end_date", endDate).
				Msg("Error: end-date must be after start-date")
		}
	}

	// Create context with timeout so CLI doesn't hang
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Add logger to context
	ctx = logger.WithContext(ctx, log)

	log.Info().
		Str("sync_type", *syncType).
		Bool("dry_run", *dryRun).
		Msg("Starting Notion sync")

	// Initialize BigQuery repository
	repo, err := bigquery.NewBigQueryDocumentRepository(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize BigQuery repository")
	}
	defer repo.Close()

	// Initialize Notion client
	notionClient := notionsync.NewNotionClient(*notionToken)

	// Perform sync based on type
	switch *syncType {
	case "transactions":
		if *notionTransactionsDBID == "" {
			log.Fatal().Msg("Error: --notion-transactions-db-id is required for transactions sync")
		}
		if err := notionsync.SyncTransactions(ctx, repo, notionClient, *notionTransactionsDBID, startDate, endDate, *dryRun); err != nil {
			log.Fatal().Err(err).Msg("Transactions sync failed")
		}

	case "accounts":
		if *notionAccountsDBID == "" {
			log.Fatal().Msg("Error: --notion-accounts-db-id is required for accounts sync")
		}
		if err := notionsync.SyncAccounts(ctx, repo, notionClient, *notionAccountsDBID, *dryRun); err != nil {
			log.Fatal().Err(err).Msg("Accounts sync failed")
		}

	case "categories":
		if *notionCategoriesDBID == "" {
			log.Fatal().Msg("Error: --notion-categories-db-id is required for categories sync")
		}
		if _, err := notionsync.SyncCategories(ctx, repo, notionClient, *notionCategoriesDBID, *dryRun); err != nil {
			log.Fatal().Err(err).Msg("Categories sync failed")
		}

	case "documents":
		if *notionDocumentsDBID == "" {
			log.Fatal().Msg("Error: --notion-documents-db-id is required for documents sync")
		}
		if err := notionsync.SyncDocuments(ctx, repo, notionClient, *notionDocumentsDBID, *dryRun); err != nil {
			log.Fatal().Err(err).Msg("Documents sync failed")
		}

	case "all":
		// Sync all tables
		var categoryPageIDs map[string]string

		if *notionAccountsDBID != "" {
			log.Info().Msg("Syncing accounts...")
			if err := notionsync.SyncAccounts(ctx, repo, notionClient, *notionAccountsDBID, *dryRun); err != nil {
				log.Error().Err(err).Msg("Accounts sync failed")
			}
		}

		if *notionCategoriesDBID != "" {
			log.Info().Msg("Syncing categories...")
			var err error
			categoryPageIDs, err = notionsync.SyncCategories(ctx, repo, notionClient, *notionCategoriesDBID, *dryRun)
			if err != nil {
				log.Error().Err(err).Msg("Categories sync failed")
			}
		}

		if *notionDocumentsDBID != "" {
			log.Info().Msg("Syncing documents...")
			if err := notionsync.SyncDocuments(ctx, repo, notionClient, *notionDocumentsDBID, *dryRun); err != nil {
				log.Error().Err(err).Msg("Documents sync failed")
			}
		}

		if *notionTransactionsDBID != "" {
			log.Info().Msg("Syncing transactions...")
			if err := notionsync.SyncTransactionsWithCategories(ctx, repo, notionClient, *notionTransactionsDBID, startDate, endDate, categoryPageIDs, *dryRun); err != nil {
				log.Error().Err(err).Msg("Transactions sync failed")
			}
		}
	}

	fmt.Println("Sync completed successfully.")
}
