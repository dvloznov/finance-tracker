package bigquery

import (
	"time"

	"cloud.google.com/go/bigquery"
)

type DocumentRow struct {
	DocumentID string `bigquery:"document_id"` // REQUIRED
	UserID     string `bigquery:"user_id"`     // NULLABLE
	GCSURI     string `bigquery:"gcs_uri"`     // REQUIRED

	DocumentType string `bigquery:"document_type"` // REQUIRED
	SourceSystem string `bigquery:"source_system"` // NULLABLE

	InstitutionID string `bigquery:"institution_id"` // NULLABLE
	AccountID     string `bigquery:"account_id"`     // NULLABLE

	StatementStartDate bigquery.NullDate `bigquery:"statement_start_date"` // NULLABLE
	StatementEndDate   bigquery.NullDate `bigquery:"statement_end_date"`   // NULLABLE

	UploadTS    time.Time              `bigquery:"upload_ts"`    // REQUIRED
	ProcessedTS bigquery.NullTimestamp `bigquery:"processed_ts"` // NULLABLE

	ParsingStatus string `bigquery:"parsing_status"` // NULLABLE

	OriginalFilename string `bigquery:"original_filename"` // NULLABLE
	FileMimeType     string `bigquery:"file_mime_type"`    // NULLABLE

	TextGCSURI string `bigquery:"text_gcs_uri"` // NULLABLE

	ChecksumSHA256 string `bigquery:"checksum_sha256"` // NULLABLE

	Metadata bigquery.NullJSON `bigquery:"metadata"` // NULLABLE
}
