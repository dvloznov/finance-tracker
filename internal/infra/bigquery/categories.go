package bigquery

import "cloud.google.com/go/bigquery"

type CategoryRow struct {
	CategoryID       string              `bigquery:"category_id"`        // REQUIRED
	ParentCategoryID bigquery.NullString `bigquery:"parent_category_id"` // NULLABLE

	Depth int64 `bigquery:"depth"` // REQUIRED (INTEGER in BQ maps to int64)

	Slug string `bigquery:"slug"` // REQUIRED
	Name string `bigquery:"name"` // REQUIRED

	Description bigquery.NullString `bigquery:"description"` // NULLABLE
	IsActive    bigquery.NullBool   `bigquery:"is_active"`   // NULLABLE

	CreatedTS bigquery.NullTimestamp `bigquery:"created_ts"` // NULLABLE (defaults to CURRENT_TIMESTAMP())
	RetiredTS bigquery.NullTimestamp `bigquery:"retired_ts"` // NULLABLE

	Metadata bigquery.NullJSON `bigquery:"metadata"` // NULLABLE (JSON)
}
