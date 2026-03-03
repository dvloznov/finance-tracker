package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

// DeleteInstitution deletes an institution and all associated accounts (with their documents,
// transactions, parsing runs, model outputs), plus any orphan documents with that institution_id.
func DeleteInstitution(ctx context.Context, institutionID string) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("DeleteInstitution: bigquery client: %w", err)
	}
	defer client.Close()

	return DeleteInstitutionWithClient(ctx, client, institutionID)
}

// DeleteInstitutionWithClient deletes an institution and all associated data using the provided client.
func DeleteInstitutionWithClient(ctx context.Context, client *bigquery.Client, institutionID string) error {
	// 1. Delete all accounts for this institution (cascades to their documents)
	accountIDs, err := listAccountIDsByInstitutionID(ctx, client, institutionID)
	if err != nil {
		return fmt.Errorf("DeleteInstitution: listing accounts: %w", err)
	}

	for _, accID := range accountIDs {
		if err := DeleteAccountWithClient(ctx, client, accID); err != nil {
			return fmt.Errorf("DeleteInstitution: deleting account %s: %w", accID, err)
		}
	}

	// 2. Delete orphan documents (institution_id set but account_id null, or from failed parsing)
	docIDs, err := listDocumentIDsByInstitutionID(ctx, client, institutionID)
	if err != nil {
		return fmt.Errorf("DeleteInstitution: listing orphan documents: %w", err)
	}

	for _, docID := range docIDs {
		if err := DeleteDocumentWithClient(ctx, client, docID); err != nil {
			return fmt.Errorf("DeleteInstitution: deleting document %s: %w", docID, err)
		}
	}

	// 3. Delete the institution
	q := client.Query(`
		DELETE FROM ` + "`" + projectID + "." + datasetID + ".institutions" + "`" + `
		WHERE institution_id = @institution_id
	`)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "institution_id", Value: institutionID},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("DeleteInstitution: run query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("DeleteInstitution: wait for job: %w", err)
	}

	if err := status.Err(); err != nil {
		return fmt.Errorf("DeleteInstitution: job error: %w", err)
	}

	return nil
}

func listAccountIDsByInstitutionID(ctx context.Context, client *bigquery.Client, institutionID string) ([]string, error) {
	q := client.Query(`
		SELECT account_id
		FROM ` + "`" + projectID + "." + datasetID + ".accounts" + "`" + `
		WHERE institution_id = @institution_id
	`)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "institution_id", Value: institutionID},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("listAccountIDsByInstitutionID: %w", err)
	}

	var ids []string
	for {
		var row struct {
			AccountID string `bigquery:"account_id"`
		}
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("listAccountIDsByInstitutionID: iterating: %w", err)
		}
		ids = append(ids, row.AccountID)
	}

	return ids, nil
}

func listDocumentIDsByInstitutionID(ctx context.Context, client *bigquery.Client, institutionID string) ([]string, error) {
	q := client.Query(`
		SELECT document_id
		FROM ` + "`" + projectID + "." + datasetID + ".documents" + "`" + `
		WHERE institution_id = @institution_id
	`)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "institution_id", Value: institutionID},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("listDocumentIDsByInstitutionID: %w", err)
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
			return nil, fmt.Errorf("listDocumentIDsByInstitutionID: iterating: %w", err)
		}
		ids = append(ids, row.DocumentID)
	}

	return ids, nil
}
