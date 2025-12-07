package infra

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
)

type DocumentRow struct {
	DocumentID string `bigquery:"document_id"` // REQUIRED
	UserID     string `bigquery:"user_id"`     // NULLABLE
	GCSURI     string `bigquery:"gcs_uri"`     // REQUIRED

	DocumentType string `bigquery:"document_type"` // REQUIRED
	SourceSystem string `bigquery:"source_system"` // NULLABLE

	InstitutionID string `bigquery:"institution_id"` // NULLABLE
	AccountID     string `bigquery:"account_id"`     // NULLABLE

	StatementStartDate bigquery.NullDate `bigquery:"statement_start_date"` // NULLABLE
	StatementEndDate   bigquery.NullDate `bigquery:"statement_end_date"`   // NULLABLE

	UploadTS    time.Time              `bigquery:"upload_ts"`    // REQUIRED
	ProcessedTS bigquery.NullTimestamp `bigquery:"processed_ts"` // NULLABLE

	ParsingStatus string `bigquery:"parsing_status"` // NULLABLE

	OriginalFilename string `bigquery:"original_filename"` // NULLABLE
	FileMimeType     string `bigquery:"file_mime_type"`    // NULLABLE

	TextGCSURI string `bigquery:"text_gcs_uri"` // NULLABLE

	ChecksumSHA256 string `bigquery:"checksum_sha256"` // NULLABLE

	Metadata bigquery.NullJSON `bigquery:"metadata"` // NULLABLE
}

// CreateDocument inserts a row into the documents table for this file.
func CreateDocument(ctx context.Context, gcsURI string) (string, error) {
	// BigQuery client
	client, err := bigquery.NewClient(ctx, "studious-union-470122-v7")
	if err != nil {
		return "", fmt.Errorf("createDocument: bigquery client: %w", err)
	}
	defer client.Close()

	// Generate a UUID for this document
	documentID := uuid.NewString()

	// Extract filename from GCS URI
	// e.g. "gs://bucket/folder/file.pdf" → "file.pdf"
	filename := extractFilenameFromGCSURI(gcsURI)

	// Prepare row to insert
	row := &DocumentRow{
		DocumentID:       documentID,
		UserID:           "denis", // You can generalize this later
		GCSURI:           gcsURI,
		DocumentType:     "BANK_STATEMENT", // For now we assume this
		SourceSystem:     "BARCLAYS",       // Later: detect automatically
		InstitutionID:    "",               // Can be filled later
		AccountID:        "",               // Can be filled later
		ParsingStatus:    "PENDING",
		UploadTS:         time.Now(),
		OriginalFilename: filename,
		FileMimeType:     "",                              // Fill later if you detect MIME
		Metadata:         bigquery.NullJSON{Valid: false}, // NULL for now
	}

	inserter := client.Dataset("finance").Table("documents").Inserter()

	if err := inserter.Put(ctx, row); err != nil {
		return "", fmt.Errorf("createDocument: inserting row: %w", err)
	}

	return documentID, nil
}

func extractFilenameFromGCSURI(uri string) string {
	// Remove "gs://"
	trimmed := strings.TrimPrefix(uri, "gs://")

	// Remove bucket name
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) < 2 {
		return trimmed
	}

	// Extract actual filename
	return path.Base(parts[1])
}
