package gcsuploader

import (
	"context"
	"fmt"
	"io"
	"os"
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
