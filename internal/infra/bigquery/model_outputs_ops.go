package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
)

const (
	moProjectID       = "studious-union-470122-v7"
	moDatasetID       = "finance"
	modelOutputsTable = "model_outputs"
)

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
// using the provided BigQuery client. Uses DML INSERT to avoid streaming buffer issues.
func InsertModelOutputWithClient(ctx context.Context, client *bigquery.Client, row *ModelOutputRow) error {
	q := client.Query(`
		INSERT INTO ` + "`" + moProjectID + "." + moDatasetID + ".model_outputs" + "`" + ` (
			output_id, parsing_run_id, document_id,
			model_name, model_version, raw_json,
			extracted_text, created_ts, notes, metadata
		)
		VALUES (
			@output_id, @parsing_run_id, @document_id,
			@model_name, @model_version, @raw_json,
			@extracted_text, @created_ts, @notes, @metadata
		)
	`)

	q.Parameters = []bigquery.QueryParameter{
		{Name: "output_id", Value: row.OutputID},
		{Name: "parsing_run_id", Value: row.ParsingRunID},
		{Name: "document_id", Value: row.DocumentID},
		{Name: "model_name", Value: row.ModelName},
		{Name: "model_version", Value: row.ModelVersion},
		{Name: "raw_json", Value: row.RawJSON},
		{Name: "extracted_text", Value: row.ExtractedText},
		{Name: "created_ts", Value: row.CreatedTS},
		{Name: "notes", Value: row.Notes},
		{Name: "metadata", Value: row.Metadata},
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
