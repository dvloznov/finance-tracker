package pipeline

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	bigquerylib "cloud.google.com/go/bigquery"
	infra "github.com/dvloznov/finance-tracker/internal/infra/bigquery"
	"github.com/google/uuid"
)

// createDocument inserts a row into the documents table for this file.
func createDocument(ctx context.Context, gcsURI string) (string, error) {
	// Generate a UUID for this document
	documentID := uuid.NewString()

	// Extract filename from GCS URI
	// e.g. "gs://bucket/folder/file.pdf" → "file.pdf"
	filename := extractFilenameFromGCSURI(gcsURI)

	// Prepare row to insert
	row := &infra.DocumentRow{
		DocumentID:       documentID,
		UserID:           DefaultUserID,
		GCSURI:           gcsURI,
		DocumentType:     DefaultDocumentType,
		SourceSystem:     DefaultSourceSystem,
		InstitutionID:    "", // Can be filled later
		AccountID:        "", // Can be filled later
		ParsingStatus:    "PENDING",
		UploadTS:         time.Now(),
		OriginalFilename: filename,
		FileMimeType:     "",                                 // Fill later if you detect MIME
		Metadata:         bigquerylib.NullJSON{Valid: false}, // NULL for now
	}

	if err := infra.InsertDocument(ctx, row); err != nil {
		return "", fmt.Errorf("createDocument: inserting row: %w", err)
	}

	return documentID, nil
}

// extractFilenameFromGCSURI extracts the filename from a GCS URI.
// e.g., "gs://bucket/folder/file.pdf" → "file.pdf"
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
