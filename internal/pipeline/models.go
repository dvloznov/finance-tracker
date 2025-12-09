package pipeline

import (
	"time"
)

// Transaction represents one normalized transaction produced by the model.
// This is a domain struct, not a BigQuery row; insertTransactions will map it
// into the finance.transactions table schema.
type Transaction struct {
	AccountName   *string // from "account_name" or nil
	AccountNumber *string // from "account_number" or nil

	Date         time.Time // parsed from "date" (YYYY-MM-DD)
	Description  string    // from "description"
	Amount       float64   // from "amount" (IN = positive, OUT = negative)
	Currency     string    // from "currency"
	BalanceAfter *float64  // from "balance_after" or nil

	Category    string // from "category"
	Subcategory string // from "subcategory"
}
