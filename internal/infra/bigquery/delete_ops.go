package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
)

// DeleteDocument deletes a document and all its related data (transactions, parsing runs, model outputs).
func DeleteDocument(ctx context.Context, documentID string) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("DeleteDocument: bigquery client: %w", err)
	}
	defer client.Close()

	// Delete in order: transactions, model_outputs, parsing_runs, then document
	// This ensures foreign key constraints are respected

	// 1. Delete transactions
	if err := deleteTransactions(ctx, client, documentID); err != nil {
		return fmt.Errorf("deleting transactions: %w", err)
	}

	// 2. Delete model outputs
	if err := deleteModelOutputs(ctx, client, documentID); err != nil {
		return fmt.Errorf("deleting model outputs: %w", err)
	}

	// 3. Delete parsing runs
	if err := deleteParsingRuns(ctx, client, documentID); err != nil {
		return fmt.Errorf("deleting parsing runs: %w", err)
	}

	// 4. Delete document
	if err := deleteDocumentRecord(ctx, client, documentID); err != nil {
		return fmt.Errorf("deleting document: %w", err)
	}

	return nil
}

func deleteTransactions(ctx context.Context, client *bigquery.Client, documentID string) error {
	q := client.Query(`
		DELETE FROM ` + "`" + projectID + "." + datasetID + ".transactions" + "`" + `
		WHERE document_id = @document_id
	`)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "document_id", Value: documentID},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("run query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("wait for job: %w", err)
	}

	if err := status.Err(); err != nil {
		return fmt.Errorf("job error: %w", err)
	}

	return nil
}

func deleteModelOutputs(ctx context.Context, client *bigquery.Client, documentID string) error {
	q := client.Query(`
		DELETE FROM ` + "`" + projectID + "." + datasetID + ".model_outputs" + "`" + `
		WHERE document_id = @document_id
	`)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "document_id", Value: documentID},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("run query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("wait for job: %w", err)
	}

	if err := status.Err(); err != nil {
		return fmt.Errorf("job error: %w", err)
	}

	return nil
}

func deleteParsingRuns(ctx context.Context, client *bigquery.Client, documentID string) error {
	q := client.Query(`
		DELETE FROM ` + "`" + projectID + "." + datasetID + ".parsing_runs" + "`" + `
		WHERE document_id = @document_id
	`)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "document_id", Value: documentID},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("run query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("wait for job: %w", err)
	}

	if err := status.Err(); err != nil {
		return fmt.Errorf("job error: %w", err)
	}

	return nil
}

func deleteDocumentRecord(ctx context.Context, client *bigquery.Client, documentID string) error {
	q := client.Query(`
		DELETE FROM ` + "`" + projectID + "." + datasetID + ".documents" + "`" + `
		WHERE document_id = @document_id
	`)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "document_id", Value: documentID},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("run query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("wait for job: %w", err)
	}

	if err := status.Err(); err != nil {
		return fmt.Errorf("job error: %w", err)
	}

	return nil
}
