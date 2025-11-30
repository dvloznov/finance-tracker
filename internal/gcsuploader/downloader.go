package gcsuploader

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

func DownloadFile(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	client, err := storage.NewClient(ctx)

	if err != nil {
		return nil, fmt.Errorf("create storage client: %w", err)
	}
	defer client.Close()

	bkt := client.Bucket(bucketName)
	obj := bkt.Object(objectName)

	r, err := obj.NewReader(ctx)

	if err != nil {
		return nil, fmt.Errorf("open GCS object reader: %w", err)
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read GCS object: %w", err)
	}

	return data, nil
}
