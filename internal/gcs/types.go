package gcs

import (
	"context"
)

// StorageService provides an interface for cloud storage operations.
// This interface enables mocking and testing of storage functionality.
type StorageService interface {
	// UploadFile uploads a local file to a storage bucket under the given object name.
	UploadFile(ctx context.Context, bucketName, objectName, filePath string) error

	// FetchFromGCS downloads file bytes from the given storage URI.
	FetchFromGCS(ctx context.Context, gcsURI string) ([]byte, error)

	// ExtractFilenameFromGCSURI extracts the filename from a storage URI.
	ExtractFilenameFromGCSURI(uri string) string
}
