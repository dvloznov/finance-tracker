package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"golang.org/x/sync/errgroup"
)

// DeleteDocument deletes a document and all its related data (transactions, parsing runs, model outputs).
func DeleteDocument(ctx context.Context, documentID string) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("DeleteDocument: bigquery client: %w", err)
	}
	defer client.Close()

	return DeleteDocumentWithClient(ctx, client, documentID)
}

// DeleteDocumentWithClient deletes a document and all related data using the provided BigQuery client.
// Related rows (transactions, model_outputs, parsing_runs) are deleted concurrently; the document
// record itself is deleted last once all child deletes succeed.
func DeleteDocumentWithClient(ctx context.Context, client *bigquery.Client, documentID string) error {
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := deleteTransactions(gCtx, client, documentID); err != nil {
			return fmt.Errorf("deleting transactions: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		if err := deleteModelOutputs(gCtx, client, documentID); err != nil {
			return fmt.Errorf("deleting model outputs: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		if err := deleteParsingRuns(gCtx, client, documentID); err != nil {
			return fmt.Errorf("deleting parsing runs: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

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
