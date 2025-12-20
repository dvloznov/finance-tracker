package pipeline

import (
	"fmt"
	"strings"
	"time"

	bigquerylib "cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"github.com/dvloznov/finance-tracker/internal/bigquery"
)

// transformModelOutputToTransactions converts raw model output into normalized transaction structs.
func transformModelOutputToTransactions(
	rawOutput map[string]interface{},
) ([]*Transaction, error) {
	// Expect top-level: { "transactions": [...] }
	txAny, ok := rawOutput["transactions"]
	if !ok {
		return nil, fmt.Errorf("transformModelOutputToTransactions: missing 'transactions' key in model output")
	}

	txSlice, ok := txAny.([]interface{})
	if !ok {
		return nil, fmt.Errorf("transformModelOutputToTransactions: 'transactions' is %T, want []interface{}", txAny)
	}

	result := make([]*Transaction, 0, len(txSlice))

	for i, item := range txSlice {
		obj, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("transformModelOutputToTransactions: element %d is %T, want map[string]interface{}", i, item)
		}

		// Required fields
		dateStr, err := getStringField(obj, "date", true)
		if err != nil {
			return nil, fmt.Errorf("transaction %d: %w", i, err)
		}
		desc, err := getStringField(obj, "description", true)
		if err != nil {
			return nil, fmt.Errorf("transaction %d: %w", i, err)
		}
		currency, err := getStringField(obj, "currency", true)
		if err != nil {
			return nil, fmt.Errorf("transaction %d: %w", i, err)
		}
		category, err := getStringField(obj, "category", true)
		if err != nil {
			return nil, fmt.Errorf("transaction %d: %w", i, err)
		}
		subcategoryPtr, err := getOptionalStringField(obj, "subcategory")
		if err != nil {
			return nil, fmt.Errorf("transaction %d: %w", i, err)
		}
		subcategory := ""
		if subcategoryPtr != nil {
			subcategory = *subcategoryPtr
		}

		amount, err := getFloat64Field(obj, "amount", true)
		if err != nil {
			return nil, fmt.Errorf("transaction %d: %w", i, err)
		}

		// Parse date string YYYY-MM-DD
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("transaction %d: invalid date %q: %w", i, dateStr, err)
		}

		// Optional fields
		balanceAfter, err := getOptionalFloat64Field(obj, "balance_after")
		if err != nil {
			return nil, fmt.Errorf("transaction %d: %w", i, err)
		}

		t := &Transaction{
			Date:         date,
			Description:  desc,
			Amount:       amount,
			Currency:     currency,
			BalanceAfter: balanceAfter,
			Category:     category,
			Subcategory:  subcategory,
		}

		result = append(result, t)
	}

	return result, nil
}

func getStringField(m map[string]interface{}, key string, required bool) (string, error) {
	v, ok := m[key]
	if !ok {
		if required {
			return "", fmt.Errorf("missing required field %q", key)
		}
		return "", nil
	}
	switch val := v.(type) {
	case string:
		if required && strings.TrimSpace(val) == "" {
			return "", fmt.Errorf("required field %q is empty", key)
		}
		return val, nil
	default:
		return "", fmt.Errorf("field %q has type %T, want string", key, v)
	}
}

func getOptionalStringField(m map[string]interface{}, key string) (*string, error) {
	v, ok := m[key]
	if !ok || v == nil {
		return nil, nil
	}
	switch val := v.(type) {
	case string:
		s := strings.TrimSpace(val)
		if s == "" {
			return nil, nil
		}
		return &s, nil
	default:
		return nil, fmt.Errorf("field %q has type %T, want string or null", key, v)
	}
}

func getFloat64Field(m map[string]interface{}, key string, required bool) (float64, error) {
	v, ok := m[key]
	if !ok {
		if required {
			return 0, fmt.Errorf("missing required field %q", key)
		}
		return 0, nil
	}
	switch val := v.(type) {
	case float64:
		return val, nil
	case int: // unlikely from encoding/json, but harmless to support
		return float64(val), nil
	default:
		return 0, fmt.Errorf("field %q has type %T, want number", key, v)
	}
}

func getOptionalFloat64Field(m map[string]interface{}, key string) (*float64, error) {
	v, ok := m[key]
	if !ok || v == nil {
		return nil, nil
	}
	switch val := v.(type) {
	case float64:
		f := val
		return &f, nil
	case int:
		f := float64(val)
		return &f, nil
	default:
		return nil, fmt.Errorf("field %q has type %T, want number or null", key, v)
	}
}

// transformAccountInfo converts raw LLM account extraction output into an AccountRow.
// Returns nil if the extraction failed or data is invalid.
func transformAccountInfo(rawOutput map[string]interface{}, documentID string) (*bigquery.AccountRow, error) {
	// Extract optional fields
	accountNumber, err := getOptionalStringField(rawOutput, "account_number")
	if err != nil {
		return nil, fmt.Errorf("transformAccountInfo: %w", err)
	}
	iban, err := getOptionalStringField(rawOutput, "iban")
	if err != nil {
		return nil, fmt.Errorf("transformAccountInfo: %w", err)
	}
	sortCode, err := getOptionalStringField(rawOutput, "sort_code")
	if err != nil {
		return nil, fmt.Errorf("transformAccountInfo: %w", err)
	}
	accountName, err := getOptionalStringField(rawOutput, "account_name")
	if err != nil {
		return nil, fmt.Errorf("transformAccountInfo: %w", err)
	}
	accountType, err := getOptionalStringField(rawOutput, "account_type")
	if err != nil {
		return nil, fmt.Errorf("transformAccountInfo: %w", err)
	}
	currency, err := getOptionalStringField(rawOutput, "currency")
	if err != nil {
		return nil, fmt.Errorf("transformAccountInfo: %w", err)
	}
	institutionID, err := getOptionalStringField(rawOutput, "institution_id")
	if err != nil {
		return nil, fmt.Errorf("transformAccountInfo: %w", err)
	}
	openedDateStr, err := getOptionalStringField(rawOutput, "opened_date")
	if err != nil {
		return nil, fmt.Errorf("transformAccountInfo: %w", err)
	}

	// Parse opened_date if present
	var openedDate civil.Date
	var hasOpenedDate bool
	if openedDateStr != nil {
		parsed, err := time.Parse("2006-01-02", *openedDateStr)
		if err != nil {
			return nil, fmt.Errorf("transformAccountInfo: invalid opened_date %q: %w", *openedDateStr, err)
		}
		openedDate = civil.DateOf(parsed)
		hasOpenedDate = true
	}

	// Build account row
	row := &bigquery.AccountRow{
		UserID: DefaultUserID,
	}

	if accountNumber != nil {
		row.AccountNumber = *accountNumber
	}
	if iban != nil {
		row.IBAN = *iban
	}
	if sortCode != nil {
		row.SortCode = *sortCode
	}
	if accountName != nil {
		row.AccountName = *accountName
	}
	if accountType != nil {
		row.AccountType = strings.ToUpper(*accountType)
	}
	if currency != nil {
		row.Currency = strings.ToUpper(*currency)
	}
	if institutionID != nil {
		row.InstitutionID = strings.ToUpper(*institutionID)
	} else {
		// Default to BARCLAYS if not extracted
		row.InstitutionID = DefaultSourceSystem
	}
	if hasOpenedDate {
		row.OpenedDate = bigquerylib.NullDate{Date: openedDate, Valid: true}
	}

	// If we got nothing useful, return nil to signal we should use default
	if row.AccountNumber == "" && row.IBAN == "" && row.SortCode == "" {
		return nil, nil
	}

	return row, nil
}

// generateDefaultAccount creates a document-scoped fallback account when
// extraction fails or returns no account identifiers.
func generateDefaultAccount(documentID string) *bigquery.AccountRow {
	// Generate synthetic account number from document ID
	accountNumber := fmt.Sprintf("DOC-%s", documentID[:8])

	return &bigquery.AccountRow{
		UserID:        DefaultUserID,
		InstitutionID: DefaultSourceSystem,
		AccountNumber: accountNumber,
		AccountName:   fmt.Sprintf("Barclays Current Account (%s)", documentID[:8]),
		AccountType:   "CURRENT",
		Currency:      "GBP", // Default to GBP for UK statements
	}
}
