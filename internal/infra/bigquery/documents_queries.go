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

// FindDocumentByChecksum retrieves a document by its SHA-256 checksum.
// Returns nil if no document with the given checksum exists.
func FindDocumentByChecksum(ctx context.Context, checksum string) (*DocumentRow, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("FindDocumentByChecksum: creating client: %w", err)
	}
	defer client.Close()

	return FindDocumentByChecksumWithClient(ctx, client, checksum)
}

// FindDocumentByChecksumWithClient retrieves a document by checksum using the provided BigQuery client.
func FindDocumentByChecksumWithClient(ctx context.Context, client *bigquery.Client, checksum string) (*DocumentRow, error) {
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
		WHERE checksum_sha256 = @checksum
		LIMIT 1
	`, projectID, datasetID)

	q := client.Query(query)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "checksum", Value: checksum},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("FindDocumentByChecksumWithClient: reading query: %w", err)
	}

	var row DocumentRow
	err = it.Next(&row)
	if err == iterator.Done {
		// No document found with this checksum
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("FindDocumentByChecksumWithClient: reading row: %w", err)
	}

	return &row, nil
}
