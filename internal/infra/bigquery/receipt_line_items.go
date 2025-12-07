package bigquery

import "cloud.google.com/go/bigquery"

type ReceiptLineItemRow struct {
	LineItemID string `bigquery:"line_item_id"` // REQUIRED
	ReceiptID  string `bigquery:"receipt_id"`   // REQUIRED

	LineIndex int64 `bigquery:"line_index"` // NULLABLE (INTEGER → int64)

	Description string `bigquery:"description"` // REQUIRED

	Quantity   bigquery.NullFloat64 `bigquery:"quantity"`    // NULLABLE (NUMERIC)
	UnitPrice  bigquery.NullFloat64 `bigquery:"unit_price"`  // NULLABLE (NUMERIC)
	TotalPrice bigquery.NullFloat64 `bigquery:"total_price"` // NULLABLE (NUMERIC)

	CategoryID      string `bigquery:"category_id"`      // NULLABLE
	SubcategoryID   string `bigquery:"subcategory_id"`   // NULLABLE
	CategoryName    string `bigquery:"category_name"`    // NULLABLE
	SubcategoryName string `bigquery:"subcategory_name"` // NULLABLE

	SKU string `bigquery:"sku"` // NULLABLE

	Metadata bigquery.NullJSON `bigquery:"metadata"` // NULLABLE
}
