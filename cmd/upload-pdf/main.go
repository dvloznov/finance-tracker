package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/dvloznov/finance-tracker/internal/gcsuploader"
)

func main() {
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
		log.Fatalf("Usage: upload-pdf -bucket BUCKET_NAME -file /path/to/file.pdf [-object OBJECT_NAME]")
	}

	if objectName == "" {
		objectName = filepath.Base(filePath)
	}

	ctx := context.Background()

	if err := gcsuploader.UploadFile(ctx, bucketName, objectName, filePath); err != nil {
		log.Fatalf("upload failed: %v", err)
	}

	fmt.Printf("Uploaded %s to gs://%s/%s\n", filePath, bucketName, objectName)
}
