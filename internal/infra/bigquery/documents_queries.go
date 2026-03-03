package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	bq "github.com/dvloznov/finance-tracker/internal/bigquery"
	"google.golang.org/api/iterator"
)


// GetDocumentByID retrieves a single document by its ID.
func GetDocumentByID(ctx context.Context, documentID string) (*bq.DocumentRow, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("GetDocumentByID: creating client: %w", err)
	}
	defer client.Close()

	return GetDocumentByIDWithClient(ctx, client, documentID)
}

// GetDocumentByIDWithClient retrieves a single document by its ID using the provided BigQuery client.
func GetDocumentByIDWithClient(ctx context.Context, client *bigquery.Client, documentID string) (*bq.DocumentRow, error) {
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
			file_mime_type,
			COALESCE(FORMAT_DATE('%%Y-%%m-%%d', statement_start_date), '') AS statement_start_date,
			COALESCE(FORMAT_DATE('%%Y-%%m-%%d', statement_end_date), '') AS statement_end_date
		FROM `+"`%s.%s.documents`"+`
		WHERE document_id = @document_id
		LIMIT 1
	`, projectID, datasetID)

	q := client.Query(query)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "document_id", Value: documentID},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetDocumentByIDWithClient: reading query: %w", err)
	}

	var row bq.DocumentRow
	if err := it.Next(&row); err != nil {
		if err == iterator.Done {
			return nil, nil
		}
		return nil, fmt.Errorf("GetDocumentByIDWithClient: iterating: %w", err)
	}

	return &row, nil
}

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
// Includes error_message from the most recent failed parsing run when available.
func ListAllDocumentsWithClient(ctx context.Context, client *bigquery.Client) ([]*bq.DocumentRow, error) {
	query := fmt.Sprintf(`
		SELECT
			d.document_id,
			d.user_id,
			d.gcs_uri,
			d.institution_id,
			d.account_id,
			d.upload_ts,
			d.parsing_status,
			d.original_filename,
			d.file_mime_type,
			COALESCE(FORMAT_DATE('%%Y-%%m-%%d', d.statement_start_date), '') AS statement_start_date,
			COALESCE(FORMAT_DATE('%%Y-%%m-%%d', d.statement_end_date), '') AS statement_end_date,
			COALESCE(pr.error_message, '') AS error_message
		FROM `+"`%s.%s.documents`"+` d
		LEFT JOIN (
			SELECT document_id, error_message
			FROM (
				SELECT document_id, error_message,
					ROW_NUMBER() OVER (PARTITION BY document_id ORDER BY finished_ts DESC NULLS LAST) AS rn
				FROM `+"`%s.%s.parsing_runs`"+`
				WHERE status = 'FAILED' AND error_message IS NOT NULL AND error_message != ''
			)
			WHERE rn = 1
		) pr ON d.document_id = pr.document_id
		ORDER BY d.upload_ts DESC
	`, projectID, datasetID, projectID, datasetID)

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
