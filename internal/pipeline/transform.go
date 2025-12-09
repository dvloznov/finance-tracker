package pipeline

import (
	"fmt"
	"strings"
	"time"
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
		subcategory, err := getStringField(obj, "subcategory", true)
		if err != nil {
			return nil, fmt.Errorf("transaction %d: %w", i, err)
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
		accountName, err := getOptionalStringField(obj, "account_name")
		if err != nil {
			return nil, fmt.Errorf("transaction %d: %w", i, err)
		}
		accountNumber, err := getOptionalStringField(obj, "account_number")
		if err != nil {
			return nil, fmt.Errorf("transaction %d: %w", i, err)
		}
		balanceAfter, err := getOptionalFloat64Field(obj, "balance_after")
		if err != nil {
			return nil, fmt.Errorf("transaction %d: %w", i, err)
		}

		t := &Transaction{
			AccountName:   accountName,
			AccountNumber: accountNumber,
			Date:          date,
			Description:   desc,
			Amount:        amount,
			Currency:      currency,
			BalanceAfter:  balanceAfter,
			Category:      category,
			Subcategory:   subcategory,
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
