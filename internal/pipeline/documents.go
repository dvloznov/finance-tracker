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
	repo, err := infra.NewBigQueryDocumentRepository(ctx)
	if err != nil {
		return "", fmt.Errorf("createDocument: creating BigQuery repository: %w", err)
	}
	defer repo.Close()

	storage := gcsuploader.NewGCSStorageService()
	return createDocumentWithRepo(ctx, gcsURI, repo, storage)
}

// createDocumentWithRepo inserts a row into the documents table using the provided repository.
// If a document with the same checksum already exists, it returns the existing document ID
// and sets state.IsReparse to true.
func createDocumentWithRepo(ctx context.Context, gcsURI string, repo infra.DocumentRepository, storage StorageService) (string, error) {
	// First, check if we have a checksum to search for duplicates
	// Note: The checksum should be set by CalculateChecksumStep before this step
	// For now, we'll skip checksum lookup if it's not available (backward compatibility)
	
	// Extract filename from GCS URI
	// e.g. "gs://bucket/folder/file.pdf" → "file.pdf"
	filename := storage.ExtractFilenameFromGCSURI(gcsURI)

	// Generate a UUID for this document (will be used if no duplicate found)
	documentID := uuid.NewString()

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

// createDocumentWithChecksumRepo inserts a row into the documents table with checksum.
func createDocumentWithChecksumRepo(ctx context.Context, gcsURI string, checksum string, repo infra.DocumentRepository, storage StorageService) (string, error) {
	// Generate a UUID for this document
	documentID := uuid.NewString()

	// Extract filename from GCS URI
	filename := storage.ExtractFilenameFromGCSURI(gcsURI)

	// Prepare row to insert with checksum
	row := &infra.DocumentRow{
		DocumentID:       documentID,
		UserID:           DefaultUserID,
		GCSURI:           gcsURI,
		DocumentType:     DefaultDocumentType,
		SourceSystem:     DefaultSourceSystem,
		InstitutionID:    "",
		AccountID:        "",
		ParsingStatus:    "PENDING",
		UploadTS:         time.Now(),
		OriginalFilename: filename,
		FileMimeType:     "",
		ChecksumSHA256:   checksum, // Set the calculated checksum
		Metadata:         bigquerylib.NullJSON{Valid: false},
	}

	if err := repo.InsertDocument(ctx, row); err != nil {
		return "", fmt.Errorf("createDocumentWithChecksum: inserting row: %w", err)
	}

	return documentID, nil
}

// extractFilenameFromGCSURI extracts the filename from a GCS URI.
// e.g., "gs://bucket/folder/file.pdf" → "file.pdf"
// DEPRECATED: Use StorageService.ExtractFilenameFromGCSURI instead.
func extractFilenameFromGCSURI(uri string) string {
	return gcsuploader.ExtractFilenameFromGCSURI(uri)
}


