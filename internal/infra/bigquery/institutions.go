package bigquery

import "cloud.google.com/go/bigquery"

type InstitutionRow struct {
	InstitutionID string `bigquery:"institution_id"` // REQUIRED
	Name          string `bigquery:"name"`           // REQUIRED

	Type     string            `bigquery:"type"`     // NULLABLE (STRING)
	Country  string            `bigquery:"country"`  // NULLABLE (STRING)
	Metadata bigquery.NullJSON `bigquery:"metadata"` // NULLABLE (JSON)

	CreatedTS bigquery.NullTimestamp `bigquery:"created_ts"` // NULLABLE (default CURRENT_TIMESTAMP())
}
