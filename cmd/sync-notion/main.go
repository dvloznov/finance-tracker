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
	startDateStr := flag.String("start-date", "", "Start date in YYYY-MM-DD format (required)")
	endDateStr := flag.String("end-date", "", "End date in YYYY-MM-DD format (required)")
	notionToken := flag.String("notion-token", "", "Notion API token (required)")
	notionDBID := flag.String("notion-db-id", "", "Notion database ID (required)")
	dryRun := flag.Bool("dry-run", false, "Dry run mode - preview changes without syncing")
	flag.Parse()

	// Validate required flags
	if *startDateStr == "" {
		log.Fatal().Msg("Error: --start-date is required")
	}
	if *endDateStr == "" {
		log.Fatal().Msg("Error: --end-date is required")
	}
	if *notionToken == "" {
		log.Fatal().Msg("Error: --notion-token is required")
	}
	if *notionDBID == "" {
		log.Fatal().Msg("Error: --notion-db-id is required")
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", *startDateStr)
	if err != nil {
		log.Fatal().Err(err).Str("start_date", *startDateStr).Msg("Error: invalid start-date format, expected YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", *endDateStr)
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

	// Create context with timeout so CLI doesn't hang
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Add logger to context
	ctx = logger.WithContext(ctx, log)

	log.Info().
		Str("start_date", *startDateStr).
		Str("end_date", *endDateStr).
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

	// Sync transactions
	if err := notionsync.SyncTransactions(ctx, repo, notionClient, *notionDBID, startDate, endDate, *dryRun); err != nil {
		log.Fatal().Err(err).Msg("Sync failed")
	}

	fmt.Println("Sync completed successfully.")
}
