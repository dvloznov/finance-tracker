package gcsuploader

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

// GCSStorageService is the concrete implementation of StorageService
// that interacts with Google Cloud Storage.
type GCSStorageService struct{}

// NewGCSStorageService creates a new instance of GCSStorageService.
func NewGCSStorageService() *GCSStorageService {
	return &GCSStorageService{}
}

// UploadFile delegates to the existing UploadFile function.
func (s *GCSStorageService) UploadFile(ctx context.Context, bucketName, objectName, filePath string) error {
	return UploadFile(ctx, bucketName, objectName, filePath)
}

// FetchFromGCS delegates to the existing FetchFromGCS function.
func (s *GCSStorageService) FetchFromGCS(ctx context.Context, gcsURI string) ([]byte, error) {
	return FetchFromGCS(ctx, gcsURI)
}

// ExtractFilenameFromGCSURI delegates to the existing ExtractFilenameFromGCSURI function.
func (s *GCSStorageService) ExtractFilenameFromGCSURI(uri string) string {
	return ExtractFilenameFromGCSURI(uri)
}
