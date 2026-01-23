package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	bq "github.com/dvloznov/finance-tracker/internal/bigquery"
	"google.golang.org/api/iterator"
)

// ListAllDocuments retrieves all documents from the database.
func ListAllDocuments(ctx context.Context) ([]*bq.DocumentRow, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ListAllDocuments: creating client: %w", err)
	}
	defer client.Close()

	return ListAllDocumentsWithClient(ctx, client)
}

// ListAllDocumentsWithClient retrieves all documents using the provided BigQuery client.
func ListAllDocumentsWithClient(ctx context.Context, client *bigquery.Client) ([]*bq.DocumentRow, error) {
	query := fmt.Sprintf(`
		SELECT
			document_id,
			user_id,
			gcs_uri,
			institution_id,
			account_id,
			upload_ts,
			parsing_status,
			original_filename,
			file_mime_type
		FROM `+"`%s.%s.documents`"+`
		ORDER BY upload_ts DESC
	`, projectID, datasetID)

	q := client.Query(query)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListAllDocumentsWithClient: reading query: %w", err)
	}

	var documents []*bq.DocumentRow
	for {
		var row bq.DocumentRow
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ListAllDocumentsWithClient: iterating: %w", err)
		}
		documents = append(documents, &row)
	}

	return documents, nil
}
