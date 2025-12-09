package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/dvloznov/finance-tracker/internal/logger"
	"github.com/dvloznov/finance-tracker/internal/pipeline"
)

func main() {
	// Initialize structured logger
	log := logger.New()

	// Parse CLI flags
	gcsURI := flag.String("gcs-uri", "", "GCS URI of the statement PDF (e.g. gs://bucket/file.pdf)")
	flag.Parse()

	if *gcsURI == "" {
		log.Fatal().Msg("Error: --gcs-uri is required")
	}

	// Create context with timeout so CLI doesn't hang
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Add logger to context
	ctx = logger.WithContext(ctx, log)

	log.Info().Str("gcs_uri", *gcsURI).Msg("Starting ingestion")

	if err := pipeline.IngestStatementFromGCS(ctx, *gcsURI); err != nil {
		log.Fatal().Err(err).Msg("Ingestion failed")
	}

	fmt.Println("Ingestion completed successfully.")
}
