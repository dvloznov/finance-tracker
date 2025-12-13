package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

// ListAllDocuments retrieves all documents from the database.
func ListAllDocuments(ctx context.Context) ([]*DocumentRow, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ListAllDocuments: creating client: %w", err)
	}
	defer client.Close()

	return ListAllDocumentsWithClient(ctx, client)
}

// ListAllDocumentsWithClient retrieves all documents using the provided BigQuery client.
func ListAllDocumentsWithClient(ctx context.Context, client *bigquery.Client) ([]*DocumentRow, error) {
	query := fmt.Sprintf(`
		SELECT
			document_id,
			user_id,
			gcs_uri,
			document_type,
			source_system,
			institution_id,
			account_id,
			statement_start_date,
			statement_end_date,
			upload_ts,
			processed_ts,
			parsing_status,
			original_filename,
			file_mime_type,
			text_gcs_uri,
			checksum_sha256,
			metadata
		FROM `+"`%s.%s.documents`"+`
		ORDER BY upload_ts DESC
	`, projectID, datasetID)

	q := client.Query(query)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListAllDocumentsWithClient: reading query: %w", err)
	}

	var documents []*DocumentRow
	for {
		var row DocumentRow
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
