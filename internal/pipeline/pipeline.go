package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	bigquerylib "cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	infra "github.com/dvloznov/finance-tracker/internal/infra/bigquery"
	"github.com/google/uuid"
)

// IngestStatementFromGCS processes a single bank statement PDF stored in GCS.
// gcsURI should look like: "gs://bucket/path/to/statement.pdf".
func IngestStatementFromGCS(ctx context.Context, gcsURI string) error {
	// Initialize pipeline state
	state := &PipelineState{
		GCSURI: gcsURI,
	}

	// Create and execute the standard ingestion pipeline
	pipeline := NewStatementIngestionPipeline()
	return pipeline.Execute(ctx, state)
}

//
// ──────────────────────────────────────────────────────────────
//  Helper function skeletons (generic, no bank-specific naming)
// ──────────────────────────────────────────────────────────────
//

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

		ModelName: DefaultModelName,
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

			UserID:    DefaultUserID,
			AccountID: "", // can map accounts later

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
