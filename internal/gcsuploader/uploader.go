package gcsuploader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

// UploadFile uploads a local file to a GCS bucket under the given object name.
// It assumes Application Default Credentials are configured (gcloud auth application-default login).
func UploadFile(ctx context.Context, bucketName, objectName, filePath string) error {
	// Open local file
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file %q: %w", filePath, err)
	}
	defer f.Close()

	// Create storage client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("create storage client: %w", err)
	}
	defer client.Close()

	// Get bucket handle
	bkt := client.Bucket(bucketName)

	// Optional: you can set a timeout per upload
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// Get object handle
	obj := bkt.Object(objectName)

	// Create writer
	w := obj.NewWriter(ctx)
	defer func() {
		// Ensure the writer is closed even on early returns
		_ = w.Close()
	}()

	// Copy file content into writer
	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("copy file to GCS writer: %w", err)
	}

	// Close to finalize the upload
	if err := w.Close(); err != nil {
		return fmt.Errorf("finalize upload: %w", err)
	}

	return nil
}

// FetchFromGCS downloads the file bytes from the given GCS URI.
func FetchFromGCS(ctx context.Context, gcsURI string) ([]byte, error) {
	// gcsURI example: gs://my-bucket/path/to/file.pdf
	if !strings.HasPrefix(gcsURI, "gs://") {
		return nil, fmt.Errorf("invalid GCS URI: %s", gcsURI)
	}

	trimmed := strings.TrimPrefix(gcsURI, "gs://")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid GCS URI (no object path): %s", gcsURI)
	}

	bucketName := parts[0]
	objectPath := parts[1]

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetchFromGCS: creating storage client: %w", err)
	}
	defer storageClient.Close()

	rc, err := storageClient.Bucket(bucketName).Object(objectPath).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetchFromGCS: reading object %s/%s: %w", bucketName, objectPath, err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("fetchFromGCS: reading bytes: %w", err)
	}

	return data, nil
}

// ExtractFilenameFromGCSURI extracts the filename from a GCS URI.
// e.g., "gs://bucket/folder/file.pdf" → "file.pdf"
func ExtractFilenameFromGCSURI(uri string) string {
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
