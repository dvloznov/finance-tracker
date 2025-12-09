package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/dvloznov/finance-tracker/internal/gcsuploader"
	"github.com/dvloznov/finance-tracker/internal/logger"
)

func main() {
	// Initialize structured logger
	log := logger.New()

	var (
		bucketName string
		objectName string
		filePath   string
	)

	flag.StringVar(&bucketName, "bucket", "", "GCS bucket name (required)")
	flag.StringVar(&objectName, "object", "", "GCS object name (optional; defaults to file name)")
	flag.StringVar(&filePath, "file", "", "Path to local PDF file (required)")
	flag.Parse()

	if bucketName == "" || filePath == "" {
		log.Fatal().Msg("Usage: upload-pdf -bucket BUCKET_NAME -file /path/to/file.pdf [-object OBJECT_NAME]")
	}

	if objectName == "" {
		objectName = filepath.Base(filePath)
	}

	ctx := context.Background()
	ctx = logger.WithContext(ctx, log)

	log.Info().
		Str("bucket", bucketName).
		Str("object", objectName).
		Str("file", filePath).
		Msg("Uploading file to GCS")

	if err := gcsuploader.UploadFile(ctx, bucketName, objectName, filePath); err != nil {
		log.Fatal().Err(err).Msg("Upload failed")
	}

	fmt.Printf("Uploaded %s to gs://%s/%s\n", filePath, bucketName, objectName)
}
