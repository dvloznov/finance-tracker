package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
)

const modelOutputsTable = "model_outputs"

// InsertModelOutput inserts a single ModelOutputRow into finance.model_outputs.
func InsertModelOutput(ctx context.Context, row *ModelOutputRow) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("InsertModelOutput: bigquery client: %w", err)
	}
	defer client.Close()

	return InsertModelOutputWithClient(ctx, client, row)
}

// InsertModelOutputWithClient inserts a single ModelOutputRow into finance.model_outputs
// using the provided BigQuery client.
func InsertModelOutputWithClient(ctx context.Context, client *bigquery.Client, row *ModelOutputRow) error {
	inserter := client.Dataset(datasetID).Table(modelOutputsTable).Inserter()
	if err := inserter.Put(ctx, row); err != nil {
		return fmt.Errorf("InsertModelOutput: inserting row: %w", err)
	}

	return nil
}
