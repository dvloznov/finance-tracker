package bigquery

import "cloud.google.com/go/bigquery"

type ModelOutputRow struct {
	OutputID     string `bigquery:"output_id"`      // REQUIRED
	ParsingRunID string `bigquery:"parsing_run_id"` // REQUIRED
	DocumentID   string `bigquery:"document_id"`    // REQUIRED

	ModelName    string              `bigquery:"model_name"`    // REQUIRED
	ModelVersion bigquery.NullString `bigquery:"model_version"` // NULLABLE

	RawJSON       bigquery.NullJSON   `bigquery:"raw_json"`       // REQUIRED (JSON)
	ExtractedText bigquery.NullString `bigquery:"extracted_text"` // NULLABLE

	CreatedTS bigquery.NullTimestamp `bigquery:"created_ts"` // REQUIRED (default CURRENT_TIMESTAMP)
	Notes     bigquery.NullString    `bigquery:"notes"`      // NULLABLE

	Metadata bigquery.NullJSON `bigquery:"metadata"` // NULLABLE
}
