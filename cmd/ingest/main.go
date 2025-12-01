package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/dvloznov/finance-tracker/internal/pipeline"
)

func main() {
	// Parse CLI flags
	gcsURI := flag.String("gcs-uri", "", "GCS URI of the statement PDF (e.g. gs://bucket/file.pdf)")
	flag.Parse()

	if *gcsURI == "" {
		log.Fatal("Error: --gcs-uri is required")
	}

	// Create context with timeout so CLI doesn't hang
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Printf("Starting ingestion for %s\n", *gcsURI)

	if err := pipeline.IngestStatementFromGCS(ctx, *gcsURI); err != nil {
		log.Fatalf("Ingestion failed: %v", err)
	}

	fmt.Println("Ingestion completed successfully.")
}
