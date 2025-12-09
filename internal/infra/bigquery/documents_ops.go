package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
)

const documentsTable = "documents"

// InsertDocument inserts a single DocumentRow into finance.documents.
func InsertDocument(ctx context.Context, row *DocumentRow) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("InsertDocument: bigquery client: %w", err)
	}
	defer client.Close()

	return InsertDocumentWithClient(ctx, client, row)
}

// InsertDocumentWithClient inserts a single DocumentRow into finance.documents
// using the provided BigQuery client.
func InsertDocumentWithClient(ctx context.Context, client *bigquery.Client, row *DocumentRow) error {
	inserter := client.Dataset(datasetID).Table(documentsTable).Inserter()
	if err := inserter.Put(ctx, row); err != nil {
		return fmt.Errorf("InsertDocument: inserting row: %w", err)
	}

	return nil
}
