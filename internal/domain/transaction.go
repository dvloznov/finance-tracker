package domain

import (
	"time"
)

// Transaction represents one normalized transaction produced by the model.
// This is a domain struct, not a BigQuery row; insertTransactions will map it
// into the finance.transactions table schema.
// Note: AccountName and AccountNumber fields have been removed as accounts are
// now extracted separately from the statement header.
type Transaction struct {
	Date         time.Time // parsed from "date" (YYYY-MM-DD)
	Description  string    // from "description"
	Amount       float64   // from "amount" (IN = positive, OUT = negative)
	Currency     string    // from "currency"
	BalanceAfter *float64  // from "balance_after" or nil

	Category    string // from "category" (kept for backward compatibility)
	Subcategory string // from "subcategory" (kept for backward compatibility)
	CategoryID  string // populated during validation - links to categories table
}
