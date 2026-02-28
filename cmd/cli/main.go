package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bq "github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/dvloznov/finance-tracker/internal/gcsuploader"
	infraBQ "github.com/dvloznov/finance-tracker/internal/infra/bigquery"
	"github.com/dvloznov/finance-tracker/internal/logger"
	"github.com/dvloznov/finance-tracker/internal/pipeline"
	"github.com/rs/zerolog"
)

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// --log-file must be parsed before subcommand flags so we accept it as a
	// global flag placed before the subcommand (e.g. cli --log-file=x.log ingest ...).
	logFilePath := flag.String("log-file", getEnvOrDefault("LOG_FILE", "cli.log"), "Log file path (truncated on each startup; JSON format)")
	flag.Parse()

	log, logCloser, err := logger.New(*logFilePath)
	if err != nil {
		fallback := logger.NewConsoleOnly()
		fallback.Fatal().Err(err).Str("log_file", *logFilePath).Msg("Failed to open log file")
	}
	defer logCloser.Close()

	// After flag.Parse the remaining args are available via flag.Args().
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	switch args[0] {
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
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
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

func buildPipelineDeps(ctx context.Context, log zerolog.Logger) (
	*infraBQ.BigQueryDocumentRepository,
	*infraBQ.BigQueryAccountRepository,
	*infraBQ.BigQueryInstitutionRepository,
	*infraBQ.BigQueryMerchantRepository,
	pipeline.StorageService,
	pipeline.AIParser,
) {
	docRepo, err := infraBQ.NewBigQueryDocumentRepository(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create document repository")
	}
	accountRepo, err := infraBQ.NewBigQueryAccountRepository(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create account repository")
	}
	institutionRepo, err := infraBQ.NewBigQueryInstitutionRepository(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create institution repository")
	}
	merchantRepo, err := infraBQ.NewBigQueryMerchantRepository(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create merchant repository")
	}
	storageService := &gcsuploader.GCSStorageService{}
	aiParser := pipeline.NewGeminiAIParser(docRepo)
	return docRepo, accountRepo, institutionRepo, merchantRepo, storageService, aiParser
}

func runIngest(log zerolog.Logger) {
	fs := flag.NewFlagSet("ingest", flag.ExitOnError)
	gcsURI := fs.String("gcs-uri", "", "GCS URI of the statement PDF")
	fs.Parse(flag.Args()[1:])

	if *gcsURI == "" {
		log.Fatal().Msg("Error: --gcs-uri is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	ctx = logger.WithContext(ctx, log)

	docRepo, accountRepo, institutionRepo, merchantRepo, storageService, aiParser := buildPipelineDeps(ctx, log)
	defer docRepo.Close()
	defer accountRepo.Close()
	defer institutionRepo.Close()
	defer merchantRepo.Close()

	log.Info().Str("gcs_uri", *gcsURI).Msg("Starting ingestion")

	if err := pipeline.IngestStatementFromGCSWithDeps(ctx, *gcsURI, "", docRepo, accountRepo, institutionRepo, merchantRepo, storageService, aiParser); err != nil {
		log.Fatal().Err(err).Msg("Ingestion failed")
	}

	fmt.Println("Ingestion completed successfully.")
}

func runUpload(log zerolog.Logger) {
	fs := flag.NewFlagSet("upload", flag.ExitOnError)
	bucketName := fs.String("bucket", "", "GCS bucket name")
	objectName := fs.String("object", "", "GCS object name (defaults to filename)")
	filePath := fs.String("file", "", "Path to local PDF file")
	fs.Parse(flag.Args()[1:])

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
	fs.Parse(flag.Args()[1:])

	if *documentID == "" {
		log.Fatal().Msg("Error: --document-id is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	ctx = logger.WithContext(ctx, log)

	docRepo, accountRepo, institutionRepo, merchantRepo, storageService, aiParser := buildPipelineDeps(ctx, log)
	defer docRepo.Close()
	defer accountRepo.Close()
	defer institutionRepo.Close()
	defer merchantRepo.Close()

	log.Info().Str("document_id", *documentID).Msg("Starting re-parse")

	doc, err := docRepo.GetDocumentByID(ctx, *documentID)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to retrieve document")
	}
	if doc == nil {
		log.Fatal().Msg("Document not found")
	}
	if doc.GCSURI == "" {
		log.Fatal().Msg("Document has no GCS URI")
	}

	log.Info().Str("gcs_uri", doc.GCSURI).Msg("Re-parsing document")

	if err := pipeline.IngestStatementFromGCSWithDeps(ctx, doc.GCSURI, *documentID, docRepo, accountRepo, institutionRepo, merchantRepo, storageService, aiParser); err != nil {
		log.Fatal().Err(err).Msg("Re-parse failed")
	}

	fmt.Println("Re-parse completed successfully.")
}

func runInspect(log zerolog.Logger) {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	documentID := fs.String("document-id", "", "Document ID to inspect")
	fs.Parse(flag.Args()[1:])

	if *documentID == "" {
		log.Fatal().Msg("Error: --document-id is required")
	}

	ctx := context.Background()
	ctx = logger.WithContext(ctx, log)

	repo, err := infraBQ.NewBigQueryDocumentRepository(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create repository")
	}
	defer repo.Close()

	doc, err := repo.GetDocumentByID(ctx, *documentID)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to retrieve document")
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

	allTxns, err := repo.QueryTransactions(ctx, bq.TransactionQuery{
		StartDate: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Now(),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to query transactions")
	}

	var txCount int
	fmt.Printf("\n=== Transactions ===\n")
	for i, txn := range allTxns {
		if txn.DocumentID != *documentID {
			continue
		}
		txCount++
		fmt.Printf("\n%d. %s\n", i+1, txn.RawDescription)
		fmt.Printf("   Date:     %s\n", txn.TransactionDate)
		fmt.Printf("   Amount:   %s %s\n", txn.Amount.FloatString(2), txn.Currency)
		if txn.BalanceAfter != nil {
			fmt.Printf("   Balance:  %s\n", txn.BalanceAfter.FloatString(2))
		}
	}
	fmt.Printf("\nTotal: %d transactions\n", txCount)
}
