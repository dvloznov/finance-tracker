package pipeline

import "time"

// Transaction represents one normalized transaction produced by the model.
// This is a domain struct, not a BigQuery row; insertTransactions will map it
// into the finance.transactions table schema.
// Note: AccountName and AccountNumber fields have been removed as accounts are
// now extracted separately from the statement header.
type Transaction struct {
	Date            time.Time // parsed from "date" (YYYY-MM-DD)
	StatementDate   time.Time // statement date (defaults to Date for now)
	Description     string    // from "description"
	MerchantName    string    // from "merchant_name" (or Description)
	TransactionType string    // from "transaction_type" (optional)
	Amount          float64   // from "amount" (IN = positive, OUT = negative)
	Currency        string    // from "currency"
	BalanceAfter    *float64  // from "balance_after" or nil

	Category    string // from "category" (kept for backward compatibility)
	Subcategory string // from "subcategory" (kept for backward compatibility)
	CategoryID  string // populated during validation - links to categories table
}
