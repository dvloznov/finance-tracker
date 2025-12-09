package pipeline

import (
	"context"
	"fmt"
	"time"

	bigquerylib "cloud.google.com/go/bigquery"
	"github.com/dvloznov/finance-tracker/internal/gcsuploader"
	infra "github.com/dvloznov/finance-tracker/internal/infra/bigquery"
	"github.com/google/uuid"
)

// createDocument inserts a row into the documents table for this file.
func createDocument(ctx context.Context, gcsURI string) (string, error) {
	repo := infra.NewBigQueryDocumentRepository()
	storage := gcsuploader.NewGCSStorageService()
	return createDocumentWithRepo(ctx, gcsURI, repo, storage)
}

// createDocumentWithRepo inserts a row into the documents table using the provided repository.
func createDocumentWithRepo(ctx context.Context, gcsURI string, repo infra.DocumentRepository, storage StorageService) (string, error) {
	// Generate a UUID for this document
	documentID := uuid.NewString()

	// Extract filename from GCS URI
	// e.g. "gs://bucket/folder/file.pdf" → "file.pdf"
	filename := storage.ExtractFilenameFromGCSURI(gcsURI)

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

	if err := repo.InsertDocument(ctx, row); err != nil {
		return "", fmt.Errorf("createDocument: inserting row: %w", err)
	}

	return documentID, nil
}

// extractFilenameFromGCSURI extracts the filename from a GCS URI.
// e.g., "gs://bucket/folder/file.pdf" → "file.pdf"
// DEPRECATED: Use StorageService.ExtractFilenameFromGCSURI instead.
func extractFilenameFromGCSURI(uri string) string {
	return gcsuploader.ExtractFilenameFromGCSURI(uri)
}


