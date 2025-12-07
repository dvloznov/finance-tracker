package bigquery

import (
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
)

type TransactionRow struct {
	TransactionID string `bigquery:"transaction_id"` // REQUIRED

	UserID    string `bigquery:"user_id"`    // NULLABLE
	AccountID string `bigquery:"account_id"` // NULLABLE

	DocumentID   string `bigquery:"document_id"`    // NULLABLE
	ParsingRunID string `bigquery:"parsing_run_id"` // NULLABLE

	TransactionDate civil.Date            `bigquery:"transaction_date"` // REQUIRED in schema
	PostingDate     bigquery.NullDate     `bigquery:"posting_date"`     // NULLABLE
	BookingDatetime bigquery.NullDateTime `bigquery:"booking_datetime"` // NULLABLE

	Amount   float64 `bigquery:"amount"`   // REQUIRED NUMERIC
	Currency string  `bigquery:"currency"` // REQUIRED STRING

	BalanceAfter bigquery.NullFloat64 `bigquery:"balance_after"` // NULLABLE NUMERIC

	Direction string `bigquery:"direction"` // NULLABLE

	RawDescription        string              `bigquery:"raw_description"`        // REQUIRED STRING
	NormalizedDescription bigquery.NullString `bigquery:"normalized_description"` // NULLABLE STRING

	MerchantID      string `bigquery:"merchant_id"`      // NULLABLE
	MerchantName    string `bigquery:"merchant_name"`    // NULLABLE
	MerchantCountry string `bigquery:"merchant_country"` // NULLABLE

	CategoryID      string              `bigquery:"category_id"`      // NULLABLE
	SubcategoryID   string              `bigquery:"subcategory_id"`   // NULLABLE
	CategoryName    bigquery.NullString `bigquery:"category_name"`    // NULLABLE
	SubcategoryName bigquery.NullString `bigquery:"subcategory_name"` // NULLABLE

	StatementLineNo bigquery.NullInt64 `bigquery:"statement_line_no"` // NULLABLE
	StatementPageNo bigquery.NullInt64 `bigquery:"statement_page_no"` // NULLABLE

	IsPending          bigquery.NullBool `bigquery:"is_pending"`
	IsRefund           bigquery.NullBool `bigquery:"is_refund"`
	IsInternalTransfer bigquery.NullBool `bigquery:"is_internal_transfer"`
	IsSplitParent      bigquery.NullBool `bigquery:"is_split_parent"`
	IsSplitChild       bigquery.NullBool `bigquery:"is_split_child"`

	ExternalReference string `bigquery:"external_reference"` // NULLABLE

	Tags []string `bigquery:"tags"` // REPEATED STRING

	CreatedTS time.Time              `bigquery:"created_ts"` // REQUIRED (default CURRENT_TIMESTAMP)
	UpdatedTS bigquery.NullTimestamp `bigquery:"updated_ts"` // NULLABLE

	Extra bigquery.NullJSON `bigquery:"extra"` // NULLABLE JSON
}
