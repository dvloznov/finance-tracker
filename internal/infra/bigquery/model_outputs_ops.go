package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	bq "github.com/dvloznov/finance-tracker/internal/bigquery"
)

const modelOutputsTable = "model_outputs"

// InsertModelOutput inserts a single ModelOutputRow into finance.model_outputs.
func InsertModelOutput(ctx context.Context, row *bq.ModelOutputRow) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("InsertModelOutput: bigquery client: %w", err)
	}
	defer client.Close()

	return InsertModelOutputWithClient(ctx, client, row)
}

// InsertModelOutputWithClient inserts a single ModelOutputRow into finance.model_outputs
// using the provided BigQuery client. Uses DML INSERT to avoid streaming buffer issues.
func InsertModelOutputWithClient(ctx context.Context, client *bigquery.Client, row *bq.ModelOutputRow) error {
	q := client.Query(`
		INSERT INTO ` + "`" + projectID + "." + datasetID + ".model_outputs" + "`" + ` (
			output_id, parsing_run_id, document_id,
			model_name, raw_json,
			created_ts
		)
		VALUES (
			@output_id, @parsing_run_id, @document_id,
			@model_name, @raw_json,
			@created_ts
		)
	`)

	q.Parameters = []bigquery.QueryParameter{
		{Name: "output_id", Value: row.OutputID},
		{Name: "parsing_run_id", Value: row.ParsingRunID},
		{Name: "document_id", Value: row.DocumentID},
		{Name: "model_name", Value: row.ModelName},
		{Name: "raw_json", Value: row.RawJSON},
		{Name: "created_ts", Value: row.CreatedTS},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("InsertModelOutput: running insert query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("InsertModelOutput: waiting for job: %w", err)
	}
	if err := status.Err(); err != nil {
		return fmt.Errorf("InsertModelOutput: job error: %w", err)
	}

	return nil
}
