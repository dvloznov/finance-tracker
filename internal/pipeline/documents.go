package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/google/uuid"
)

// createDocumentWithRepo inserts a row into the documents table using the provided repository.
func createDocumentWithRepo(ctx context.Context, gcsURI string, repo bigquery.DocumentRepository, storage StorageService) (string, error) {
	// Extract filename from GCS URI
	// e.g. "gs://bucket/folder/file.pdf" → "file.pdf"
	filename := storage.ExtractFilenameFromGCSURI(gcsURI)

	// Generate a UUID for this document (will be used if no duplicate found)
	documentID := uuid.NewString()

	// Prepare row to insert
	row := &bigquery.DocumentRow{
		DocumentID:       documentID,
		UserID:           DefaultUserID,
		GCSURI:           gcsURI,
		InstitutionID:    "", // Can be filled later
		AccountID:        "", // Can be filled later
		ParsingStatus:    "PENDING",
		UploadTS:         time.Now(),
		OriginalFilename: filename,
		FileMimeType:     "", // Fill later if you detect MIME
	}

	if err := repo.InsertDocument(ctx, row); err != nil {
		return "", fmt.Errorf("createDocument: inserting row: %w", err)
	}

	return documentID, nil
}
