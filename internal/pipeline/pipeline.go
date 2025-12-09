package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	bigquerylib "cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"github.com/dvloznov/finance-tracker/internal/gcsuploader"
	infra "github.com/dvloznov/finance-tracker/internal/infra/bigquery"
	"github.com/google/uuid"
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

// IngestStatementFromGCS processes a single bank statement PDF stored in GCS.
// gcsURI should look like: "gs://bucket/path/to/statement.pdf".
func IngestStatementFromGCS(ctx context.Context, gcsURI string) error {
	// 1. Create a document record for this file.
	documentID, err := createDocument(ctx, gcsURI)
	if err != nil {
		return err
	}

	// 2. Start a parsing run (status=RUNNING).
	parsingRunID, err := infra.StartParsingRun(ctx, documentID)
	if err != nil {
		return err
	}

	// 3. Fetch the PDF bytes from GCS.
	pdfBytes, err := gcsuploader.FetchFromGCS(ctx, gcsURI)
	if err != nil {
		infra.MarkParsingRunFailed(ctx, parsingRunID, err)
		return err
	}

	// 4. Call the statement parser (Gemini) with the PDF.
	rawModelOutput, err := parseStatementWithModel(ctx, pdfBytes)
	if err != nil {
		infra.MarkParsingRunFailed(ctx, parsingRunID, err)
		return err
	}

	// 5. Store raw model output in model_outputs.
	_, err = storeModelOutput(ctx, parsingRunID, documentID, rawModelOutput)
	if err != nil {
		infra.MarkParsingRunFailed(ctx, parsingRunID, err)
		return err
	}

	// 6. Transform raw model output into normalized transactions.
	txs, err := transformModelOutputToTransactions(rawModelOutput)
	if err != nil {
		infra.MarkParsingRunFailed(ctx, parsingRunID, err)
		return err
	}

	// 7. Insert transactions into the transactions table.
	if err := insertTransactions(ctx, documentID, parsingRunID, txs); err != nil {
		infra.MarkParsingRunFailed(ctx, parsingRunID, err)
		return err
	}

	// 8. Mark parsing run as SUCCESS.
	if err := infra.MarkParsingRunSucceeded(ctx, parsingRunID); err != nil {
		return err
	}

	return nil
}

//
// ──────────────────────────────────────────────────────────────
//  Helper function skeletons (generic, no bank-specific naming)
// ──────────────────────────────────────────────────────────────
//

// createDocument inserts a row into the documents table for this file.
func createDocument(ctx context.Context, gcsURI string) (string, error) {
	// Generate a UUID for this document
	documentID := uuid.NewString()

	// Extract filename from GCS URI
	// e.g. "gs://bucket/folder/file.pdf" → "file.pdf"
	filename := gcsuploader.ExtractFilenameFromGCSURI(gcsURI)

	// Prepare row to insert
	row := &infra.DocumentRow{
		DocumentID:       documentID,
		UserID:           "denis", // You can generalize this later
		GCSURI:           gcsURI,
		DocumentType:     "BANK_STATEMENT", // For now we assume this
		SourceSystem:     "BARCLAYS",       // Later: detect automatically
		InstitutionID:    "",               // Can be filled later
		AccountID:        "",               // Can be filled later
		ParsingStatus:    "PENDING",
		UploadTS:         time.Now(),
		OriginalFilename: filename,
		FileMimeType:     "",                                 // Fill later if you detect MIME
		Metadata:         bigquerylib.NullJSON{Valid: false}, // NULL for now
	}

	if err := infra.InsertDocument(ctx, row); err != nil {
		return "", fmt.Errorf("createDocument: inserting row: %w", err)
	}

	return documentID, nil
}

// storeModelOutput inserts raw model output into the model_outputs table.
func storeModelOutput(
	ctx context.Context,
	parsingRunID string,
	documentID string,
	rawOutput map[string]interface{},
) (string, error) {
	outputID := uuid.NewString()

	jsonBytes, err := json.Marshal(rawOutput)
	if err != nil {
		return "", fmt.Errorf("storeModelOutput: marshal rawOutput: %w", err)
	}

	row := &infra.ModelOutputRow{
		OutputID:     outputID,
		ParsingRunID: parsingRunID,
		DocumentID:   documentID,

		ModelName: "gemini-2.5-flash",
		ModelVersion: bigquerylib.NullString{
			Valid: false,
		},

		CreatedTS: bigquerylib.NullTimestamp{
			Timestamp: time.Now(),
			Valid:     true,
		},

		RawJSON: bigquerylib.NullJSON{
			JSONVal: string(jsonBytes), // <<<< correct
			Valid:   true,
		},

		ExtractedText: bigquerylib.NullString{Valid: false},
		Notes:         bigquerylib.NullString{Valid: false},

		Metadata: bigquerylib.NullJSON{
			Valid: false,
		},
	}

	if err := infra.InsertModelOutput(ctx, row); err != nil {
		return "", fmt.Errorf("storeModelOutput: inserting row: %w", err)
	}

	return outputID, nil
}

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

// insertTransactions writes a batch of transactions to the transactions table.
func insertTransactions(
	ctx context.Context,
	documentID string,
	parsingRunID string,
	txs []*Transaction,
) error {
	if len(txs) == 0 {
		return nil
	}

	rows := make([]*infra.TransactionRow, 0, len(txs))

	for _, t := range txs {
		// Determine direction based on sign of amount
		dir := ""
		if t.Amount > 0 {
			dir = "IN"
		} else if t.Amount < 0 {
			dir = "OUT"
		}

		txDate := civil.DateOf(t.Date)

		var balanceAfter bigquerylib.NullFloat64
		if t.BalanceAfter != nil {
			balanceAfter = bigquerylib.NullFloat64{
				Float64: *t.BalanceAfter,
				Valid:   true,
			}
		}

		var normalizedDescription bigquerylib.NullString
		if t.Description != "" {
			normalizedDescription = bigquerylib.NullString{
				StringVal: t.Description,
				Valid:     true,
			}
		}

		var categoryName bigquerylib.NullString
		if strings.TrimSpace(t.Category) != "" {
			categoryName = bigquerylib.NullString{
				StringVal: t.Category,
				Valid:     true,
			}
		}

		var subcategoryName bigquerylib.NullString
		if strings.TrimSpace(t.Subcategory) != "" {
			subcategoryName = bigquerylib.NullString{
				StringVal: t.Subcategory,
				Valid:     true,
			}
		}

		row := &infra.TransactionRow{
			TransactionID: uuid.NewString(),

			UserID:    "denis", // same as in documents
			AccountID: "",      // can map accounts later

			DocumentID:   documentID,
			ParsingRunID: parsingRunID,

			TransactionDate: txDate,

			Amount:   t.Amount,
			Currency: t.Currency,

			BalanceAfter: balanceAfter,

			Direction: dir,

			RawDescription:        t.Description,
			NormalizedDescription: normalizedDescription,

			CategoryName:    categoryName,
			SubcategoryName: subcategoryName,
		}

		rows = append(rows, row)
	}

	if err := infra.InsertTransactions(ctx, rows); err != nil {
		return fmt.Errorf("insertTransactions: inserting rows: %w", err)
	}

	return nil
}
