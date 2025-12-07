package bigquery

import (
	"time"

	"cloud.google.com/go/bigquery"
)

type ParsingRunRow struct {
	ParsingRunID string `bigquery:"parsing_run_id"` // REQUIRED
	DocumentID   string `bigquery:"document_id"`    // REQUIRED

	StartedTS  time.Time              `bigquery:"started_ts"`  // REQUIRED
	FinishedTS bigquery.NullTimestamp `bigquery:"finished_ts"` // NULLABLE

	ParserType    string `bigquery:"parser_type"`    // NULLABLE
	ParserVersion string `bigquery:"parser_version"` // NULLABLE

	Status       string `bigquery:"status"`        // NULLABLE
	ErrorMessage string `bigquery:"error_message"` // NULLABLE

	TokensInput  bigquery.NullInt64 `bigquery:"tokens_input"`  // NULLABLE
	TokensOutput bigquery.NullInt64 `bigquery:"tokens_output"` // NULLABLE

	Metadata bigquery.NullJSON `bigquery:"metadata"` // NULLABLE
}
