package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dvloznov/finance-tracker/internal/gcsuploader"
	infraBQ "github.com/dvloznov/finance-tracker/internal/infra/bigquery"
	"github.com/dvloznov/finance-tracker/internal/logger"
	"github.com/dvloznov/finance-tracker/internal/pipeline"
	"github.com/rs/zerolog"
)

func main() {
	log := logger.New()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "ingest":
		runIngest(log)
	case "upload":
		runUpload(log)
	case "reparse":
		runReparse(log)
	case "inspect":
		runInspect(log)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Finance Tracker CLI")
	fmt.Println("\nUsage:")
	fmt.Println("  cli <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  ingest    Parse and ingest a bank statement from GCS")
	fmt.Println("  upload    Upload a PDF file to GCS")
	fmt.Println("  reparse   Re-parse an existing document by ID")
	fmt.Println("  inspect   Inspect a document and its transactions")
	fmt.Println("  help      Show this help message")
	fmt.Println("\nRun 'cli <command> -h' for more information on a command.")
}

func runIngest(log zerolog.Logger) {
	fs := flag.NewFlagSet("ingest", flag.ExitOnError)
	gcsURI := fs.String("gcs-uri", "", "GCS URI of the statement PDF")
	fs.Parse(os.Args[2:])

	if *gcsURI == "" {
		log.Fatal().Msg("Error: --gcs-uri is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	ctx = logger.WithContext(ctx, log)

	log.Info().Str("gcs_uri", *gcsURI).Msg("Starting ingestion")

	if err := pipeline.IngestStatementFromGCS(ctx, *gcsURI); err != nil {
		log.Fatal().Err(err).Msg("Ingestion failed")
	}

	fmt.Println("Ingestion completed successfully.")
}

func runUpload(log zerolog.Logger) {
	fs := flag.NewFlagSet("upload", flag.ExitOnError)
	bucketName := fs.String("bucket", "", "GCS bucket name")
	objectName := fs.String("object", "", "GCS object name (defaults to filename)")
	filePath := fs.String("file", "", "Path to local PDF file")
	fs.Parse(os.Args[2:])

	if *bucketName == "" || *filePath == "" {
		log.Fatal().Msg("Usage: cli upload -bucket NAME -file PATH")
	}

	if *objectName == "" {
		*objectName = filepath.Base(*filePath)
	}

	ctx := context.Background()
	ctx = logger.WithContext(ctx, log)

	log.Info().
		Str("bucket", *bucketName).
		Str("object", *objectName).
		Str("file", *filePath).
		Msg("Uploading file to GCS")

	if err := gcsuploader.UploadFile(ctx, *bucketName, *objectName, *filePath); err != nil {
		log.Fatal().Err(err).Msg("Upload failed")
	}

	fmt.Printf("Uploaded %s to gs://%s/%s\n", *filePath, *bucketName, *objectName)
}

func runReparse(log zerolog.Logger) {
	fs := flag.NewFlagSet("reparse", flag.ExitOnError)
	documentID := fs.String("document-id", "", "Document ID to re-parse")
	fs.Parse(os.Args[2:])

	if *documentID == "" {
		log.Fatal().Msg("Error: --document-id is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	ctx = logger.WithContext(ctx, log)

	log.Info().Str("document_id", *documentID).Msg("Starting re-parse")

	// Get all documents and find the one with matching ID
	docs, err := infraBQ.ListAllDocuments(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to list documents")
	}

	var doc *infraBQ.DocumentRow
	for _, d := range docs {
		if d.DocumentID == *documentID {
			doc = d
			break
		}
	}

	if doc == nil {
		log.Fatal().Msg("Document not found")
	}

	if doc.GCSURI == "" {
		log.Fatal().Msg("Document has no GCS URI")
	}

	log.Info().Str("gcs_uri", doc.GCSURI).Msg("Re-parsing document")

	if err := pipeline.IngestStatementFromGCS(ctx, doc.GCSURI); err != nil {
		log.Fatal().Err(err).Msg("Re-parse failed")
	}

	fmt.Println("Re-parse completed successfully.")
}

func runInspect(log zerolog.Logger) {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	documentID := fs.String("document-id", "", "Document ID to inspect")
	fs.Parse(os.Args[2:])

	if *documentID == "" {
		log.Fatal().Msg("Error: --document-id is required")
	}

	ctx := context.Background()
	ctx = logger.WithContext(ctx, log)

	// Get all documents and find the one with matching ID
	docs, err := infraBQ.ListAllDocuments(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to list documents")
	}

	var doc *infraBQ.DocumentRow
	for _, d := range docs {
		if d.DocumentID == *documentID {
			doc = d
			break
		}
	}

	if doc == nil {
		log.Fatal().Msg("Document not found")
	}

	fmt.Println("\n=== Document Details ===")
	fmt.Printf("ID:         %s\n", doc.DocumentID)
	fmt.Printf("Account ID: %s\n", doc.AccountID)
	fmt.Printf("GCS URI:    %s\n", doc.GCSURI)
	fmt.Printf("Created:    %s\n", doc.UploadTS)
	fmt.Printf("Status:     %s\n", doc.ParsingStatus)

	// Get all transactions (we'll filter by document ID)
	// Query for a wide date range to get all transactions
	startDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Now().AddDate(1, 0, 0)

	repo, err := infraBQ.NewBigQueryDocumentRepository(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create repository")
	}
	defer repo.Close()

	allTxns, err := repo.QueryTransactionsByDateRange(ctx, startDate, endDate)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to query transactions")
	}

	// Filter by document ID
	var transactions []*infraBQ.TransactionRow
	for _, txn := range allTxns {
		if txn.DocumentID == *documentID {
			transactions = append(transactions, txn)
		}
	}

	fmt.Printf("\n=== Transactions (%d) ===\n", len(transactions))
	for i, txn := range transactions {
		fmt.Printf("\n%d. %s\n", i+1, txn.RawDescription)
		fmt.Printf("   Date:     %s\n", txn.TransactionDate)
		fmt.Printf("   Amount:   %s %s\n", txn.Amount.FloatString(2), txn.Currency)
		if txn.CategoryName.Valid {
			fmt.Printf("   Category: %s\n", txn.CategoryName.StringVal)
		}
		if txn.BalanceAfter != nil {
			fmt.Printf("   Balance:  %s\n", txn.BalanceAfter.FloatString(2))
		}
	}
	fmt.Println()
}
