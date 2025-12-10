package bigquery

import (
	"math/big"
	"time"

	"cloud.google.com/go/bigquery"
)

type ReceiptRow struct {
	ReceiptID string `bigquery:"receipt_id"` // REQUIRED
	UserID    string `bigquery:"user_id"`    // NULLABLE

	DocumentID   string `bigquery:"document_id"`    // REQUIRED
	ParsingRunID string `bigquery:"parsing_run_id"` // NULLABLE

	MerchantID   string `bigquery:"merchant_id"`   // NULLABLE
	MerchantName string `bigquery:"merchant_name"` // NULLABLE

	PurchaseDateTime bigquery.NullDateTime `bigquery:"purchase_datetime"` // DATETIME, NULLABLE
	PurchaseDate     bigquery.NullDate     `bigquery:"purchase_date"`     // DATE, NULLABLE

	TotalAmount    *big.Rat `bigquery:"total_amount"`    // NUMERIC, REQUIRED
	SubtotalAmount *big.Rat `bigquery:"subtotal_amount"` // NUMERIC, NULLABLE
	TaxAmount      *big.Rat `bigquery:"tax_amount"`      // NUMERIC, NULLABLE
	TipAmount      *big.Rat `bigquery:"tip_amount"`      // NUMERIC, NULLABLE

	Currency string `bigquery:"currency"` // REQUIRED

	PaymentMethod       string `bigquery:"payment_method"`        // NULLABLE
	CardLast4           string `bigquery:"card_last4"`            // NULLABLE
	LinkedTransactionID string `bigquery:"linked_transaction_id"` // NULLABLE

	CreatedTS time.Time              `bigquery:"created_ts"` // REQUIRED (default CURRENT_TIMESTAMP())
	UpdatedTS bigquery.NullTimestamp `bigquery:"updated_ts"` // NULLABLE

	Metadata bigquery.NullJSON `bigquery:"metadata"` // JSON, NULLABLE
}
