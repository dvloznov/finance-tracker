package infra

import "cloud.google.com/go/bigquery"

type AccountRow struct {
	AccountID string `bigquery:"account_id"` // REQUIRED

	UserID        string `bigquery:"user_id"`        // NULLABLE (empty string → "")
	InstitutionID string `bigquery:"institution_id"` // NULLABLE
	AccountName   string `bigquery:"account_name"`   // NULLABLE
	AccountNumber string `bigquery:"account_number"` // NULLABLE
	SortCode      string `bigquery:"sort_code"`      // NULLABLE
	IBAN          string `bigquery:"iban"`           // NULLABLE
	Currency      string `bigquery:"currency"`       // NULLABLE
	AccountType   string `bigquery:"account_type"`   // NULLABLE

	OpenedDate bigquery.NullDate      `bigquery:"opened_date"` // DATE, NULLABLE
	ClosedDate bigquery.NullDate      `bigquery:"closed_date"` // DATE, NULLABLE
	IsPrimary  bigquery.NullBool      `bigquery:"is_primary"`  // BOOLEAN, NULLABLE
	Metadata   bigquery.NullJSON      `bigquery:"metadata"`    // JSON, NULLABLE
	CreatedTS  bigquery.NullTimestamp `bigquery:"created_ts"`  // TIMESTAMP, NULLABLE (default CURRENT_TIMESTAMP())
	UpdatedTS  bigquery.NullTimestamp `bigquery:"updated_ts"`  // TIMESTAMP, NULLABLE
}
