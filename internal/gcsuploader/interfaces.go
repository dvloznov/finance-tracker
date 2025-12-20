package gcsuploader

import (
	"context"

	"github.com/dvloznov/finance-tracker/internal/gcs"
)

// Re-export interface from shared package for backward compatibility
type StorageService = gcs.StorageService

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
