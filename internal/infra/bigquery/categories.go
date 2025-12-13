package bigquery

import "cloud.google.com/go/bigquery"

// CategoryRow represents a denormalized category-subcategory pair
type CategoryRow struct {
	CategoryID      string              `bigquery:"category_id"`      // REQUIRED
	CategoryName    string              `bigquery:"category_name"`    // REQUIRED
	SubcategoryName bigquery.NullString `bigquery:"subcategory_name"` // NULLABLE

	Slug string `bigquery:"slug"` // REQUIRED

	Description bigquery.NullString `bigquery:"description"` // NULLABLE
	IsActive    bigquery.NullBool   `bigquery:"is_active"`   // NULLABLE

	CreatedTS bigquery.NullTimestamp `bigquery:"created_ts"` // NULLABLE (defaults to CURRENT_TIMESTAMP())
	RetiredTS bigquery.NullTimestamp `bigquery:"retired_ts"` // NULLABLE

	Metadata bigquery.NullJSON `bigquery:"metadata"` // NULLABLE (JSON)
}
