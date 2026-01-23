package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	bq "github.com/dvloznov/finance-tracker/internal/bigquery"
)

const documentsTable = "documents"

// InsertDocument inserts a single DocumentRow into finance.documents.
func InsertDocument(ctx context.Context, row *bq.DocumentRow) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("InsertDocument: bigquery client: %w", err)
	}
	defer client.Close()

	return InsertDocumentWithClient(ctx, client, row)
}

// InsertDocumentWithClient inserts a single DocumentRow into finance.documents
// using the provided BigQuery client.
// Uses INSERT query instead of streaming API to allow immediate UPDATEs.
func InsertDocumentWithClient(ctx context.Context, client *bigquery.Client, row *bq.DocumentRow) error {
	q := client.Query(fmt.Sprintf(`
		INSERT %s.%s (
			document_id,
			user_id,
			gcs_uri,
			institution_id,
			account_id,
			upload_ts,
			parsing_status,
			original_filename,
			file_mime_type
		)
		VALUES (
			@document_id,
			@user_id,
			@gcs_uri,
			@institution_id,
			@account_id,
			@upload_ts,
			@parsing_status,
			@original_filename,
			@file_mime_type
		)
	`, datasetID, documentsTable))

	q.Parameters = []bigquery.QueryParameter{
		{Name: "document_id", Value: row.DocumentID},
		{Name: "user_id", Value: row.UserID},
		{Name: "gcs_uri", Value: row.GCSURI},
		{Name: "institution_id", Value: row.InstitutionID},
		{Name: "account_id", Value: row.AccountID},
		{Name: "upload_ts", Value: row.UploadTS},
		{Name: "parsing_status", Value: row.ParsingStatus},
		{Name: "original_filename", Value: row.OriginalFilename},
		{Name: "file_mime_type", Value: row.FileMimeType},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("InsertDocument: running insert query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("InsertDocument: waiting for job: %w", err)
	}
	if err := status.Err(); err != nil {
		return fmt.Errorf("InsertDocument: job error: %w", err)
	}

	return nil
}

// UpdateDocumentParsingStatus updates the parsing_status field for a document.
func UpdateDocumentParsingStatus(ctx context.Context, documentID, status string) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("UpdateDocumentParsingStatus: bigquery client: %w", err)
	}
	defer client.Close()

	return UpdateDocumentParsingStatusWithClient(ctx, client, documentID, status)
}

// UpdateDocumentParsingStatusWithClient updates the parsing_status field for a document
// using the provided BigQuery client.
func UpdateDocumentParsingStatusWithClient(ctx context.Context, client *bigquery.Client, documentID, status string) error {
	query := client.Query(`
		UPDATE ` + "`" + projectID + "." + datasetID + "." + documentsTable + "`" + `
		SET parsing_status = @status
		WHERE document_id = @document_id
	`)
	query.Parameters = []bigquery.QueryParameter{
		{Name: "status", Value: status},
		{Name: "document_id", Value: documentID},
	}

	job, err := query.Run(ctx)
	if err != nil {
		return fmt.Errorf("UpdateDocumentParsingStatus: query run: %w", err)
	}

	status2, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("UpdateDocumentParsingStatus: job wait: %w", err)
	}

	if status2.Err() != nil {
		return fmt.Errorf("UpdateDocumentParsingStatus: job error: %w", status2.Err())
	}

	return nil
}

// UpdateDocumentAccountAndInstitution updates the account_id and institution_id fields for a document.
func UpdateDocumentAccountAndInstitution(ctx context.Context, documentID, accountID, institutionID string) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("UpdateDocumentAccountAndInstitution: bigquery client: %w", err)
	}
	defer client.Close()

	return UpdateDocumentAccountAndInstitutionWithClient(ctx, client, documentID, accountID, institutionID)
}

// UpdateDocumentAccountAndInstitutionWithClient updates the account_id and institution_id fields for a document
// using the provided BigQuery client.
func UpdateDocumentAccountAndInstitutionWithClient(ctx context.Context, client *bigquery.Client, documentID, accountID, institutionID string) error {
	query := client.Query(`
		UPDATE ` + "`" + projectID + "." + datasetID + "." + documentsTable + "`" + `
		SET account_id = @account_id,
			institution_id = @institution_id
		WHERE document_id = @document_id
	`)
	query.Parameters = []bigquery.QueryParameter{
		{Name: "account_id", Value: accountID},
		{Name: "institution_id", Value: institutionID},
		{Name: "document_id", Value: documentID},
	}

	job, err := query.Run(ctx)
	if err != nil {
		return fmt.Errorf("UpdateDocumentAccountAndInstitution: query run: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("UpdateDocumentAccountAndInstitution: job wait: %w", err)
	}

	if status.Err() != nil {
		return fmt.Errorf("UpdateDocumentAccountAndInstitution: job error: %w", status.Err())
	}

	return nil
}
