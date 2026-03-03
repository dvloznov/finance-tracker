package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

// DeleteAccount deletes an account and all associated documents (with their transactions,
// parsing runs, model outputs). Uses the same cascade as DeleteDocument for each document.
func DeleteAccount(ctx context.Context, accountID string) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("DeleteAccount: bigquery client: %w", err)
	}
	defer client.Close()

	return DeleteAccountWithClient(ctx, client, accountID)
}

// DeleteAccountWithClient deletes an account and all associated data using the provided client.
func DeleteAccountWithClient(ctx context.Context, client *bigquery.Client, accountID string) error {
	// 1. Get all documents for this account
	docIDs, err := listDocumentIDsByAccountID(ctx, client, accountID)
	if err != nil {
		return fmt.Errorf("DeleteAccount: listing documents: %w", err)
	}

	// 2. Delete each document (cascades to transactions, model_outputs, parsing_runs)
	for _, docID := range docIDs {
		if err := DeleteDocumentWithClient(ctx, client, docID); err != nil {
			return fmt.Errorf("DeleteAccount: deleting document %s: %w", docID, err)
		}
	}

	// 3. Delete the account
	q := client.Query(`
		DELETE FROM ` + "`" + projectID + "." + datasetID + ".accounts" + "`" + `
		WHERE account_id = @account_id
	`)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "account_id", Value: accountID},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("DeleteAccount: run query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("DeleteAccount: wait for job: %w", err)
	}

	if err := status.Err(); err != nil {
		return fmt.Errorf("DeleteAccount: job error: %w", err)
	}

	return nil
}

func listDocumentIDsByAccountID(ctx context.Context, client *bigquery.Client, accountID string) ([]string, error) {
	q := client.Query(`
		SELECT document_id
		FROM ` + "`" + projectID + "." + datasetID + ".documents" + "`" + `
		WHERE account_id = @account_id
	`)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "account_id", Value: accountID},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("listDocumentIDsByAccountID: %w", err)
	}

	var ids []string
	for {
		var row struct {
			DocumentID string `bigquery:"document_id"`
		}
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("listDocumentIDsByAccountID: iterating: %w", err)
		}
		ids = append(ids, row.DocumentID)
	}

	return ids, nil
}
