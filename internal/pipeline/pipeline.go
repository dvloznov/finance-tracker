package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	bigquerylib "cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/dvloznov/finance-tracker/internal/gcsuploader"
	infraBQ "github.com/dvloznov/finance-tracker/internal/infra/bigquery"
	"github.com/google/uuid"
)

// IngestStatementFromGCS processes a single bank statement PDF stored in GCS.
// gcsURI should look like: "gs://bucket/path/to/statement.pdf".
func IngestStatementFromGCS(ctx context.Context, gcsURI string) error {
	// Initialize concrete dependencies
	repo, err := infraBQ.NewBigQueryDocumentRepository(ctx)
	if err != nil {
		return fmt.Errorf("IngestStatementFromGCS: creating BigQuery repository: %w", err)
	}
	defer repo.Close()

	accountRepo, err := infraBQ.NewBigQueryAccountRepository(ctx)
	if err != nil {
		return fmt.Errorf("IngestStatementFromGCS: creating BigQuery account repository: %w", err)
	}
	defer accountRepo.Close()

	storage := &gcsuploader.GCSStorageService{}
	aiParser := NewGeminiAIParser(repo)

	return IngestStatementFromGCSWithDeps(ctx, gcsURI, repo, accountRepo, storage, aiParser)
}

// IngestStatementFromGCSWithDeps processes a single bank statement PDF stored in GCS
// using the provided dependencies. This enables dependency injection for testing.
func IngestStatementFromGCSWithDeps(
	ctx context.Context,
	gcsURI string,
	repo bigquery.DocumentRepository,
	accountRepo bigquery.AccountRepository,
	storage StorageService,
	aiParser AIParser,
) error {
	// Initialize pipeline state
	state := &PipelineState{
		GCSURI:         gcsURI,
		DocumentRepo:   repo,
		AccountRepo:    accountRepo,
		StorageService: storage,
		AIParser:       aiParser,
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
	repo, err := infraBQ.NewBigQueryDocumentRepository(ctx)
	if err != nil {
		return "", fmt.Errorf("storeModelOutput: creating BigQuery repository: %w", err)
	}
	defer repo.Close()

	return storeModelOutputWithRepo(ctx, parsingRunID, documentID, rawOutput, repo)
}

// storeModelOutputWithRepo inserts raw model output into the model_outputs table using the provided repository.
func storeModelOutputWithRepo(
	ctx context.Context,
	parsingRunID string,
	documentID string,
	rawOutput map[string]interface{},
	repo bigquery.DocumentRepository,
) (string, error) {
	outputID := uuid.NewString()

	jsonBytes, err := json.Marshal(rawOutput)
	if err != nil {
		return "", fmt.Errorf("storeModelOutput: marshal rawOutput: %w", err)
	}

	row := &bigquery.ModelOutputRow{
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

	if err := repo.InsertModelOutput(ctx, row); err != nil {
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
	repo, err := infraBQ.NewBigQueryDocumentRepository(ctx)
	if err != nil {
		return fmt.Errorf("insertTransactions: creating BigQuery repository: %w", err)
	}
	defer repo.Close()

	return insertTransactionsWithRepo(ctx, documentID, parsingRunID, "", txs, repo)
}

// insertTransactionsWithRepo writes a batch of transactions to the transactions table using the provided repository.
func insertTransactionsWithRepo(
	ctx context.Context,
	documentID string,
	parsingRunID string,
	accountID string,
	txs []*Transaction,
	repo bigquery.DocumentRepository,
) error {
	if len(txs) == 0 {
		return nil
	}

	rows := make([]*bigquery.TransactionRow, 0, len(txs))

	for _, t := range txs {
		// Determine direction based on sign of amount
		var dir bigquerylib.NullString
		if t.Amount > 0 {
			dir = bigquerylib.NullString{StringVal: "IN", Valid: true}
		} else if t.Amount < 0 {
			dir = bigquerylib.NullString{StringVal: "OUT", Valid: true}
		}

		txDate := civil.DateOf(t.Date)

		var balanceAfter *big.Rat
		if t.BalanceAfter != nil {
			balanceAfter = new(big.Rat).SetFloat64(*t.BalanceAfter)
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

		var categoryID bigquerylib.NullString
		if strings.TrimSpace(t.CategoryID) != "" {
			categoryID = bigquerylib.NullString{
				StringVal: t.CategoryID,
				Valid:     true,
			}
		}

		row := &bigquery.TransactionRow{
			TransactionID: uuid.NewString(),

			UserID:    DefaultUserID,
			AccountID: accountID, // Link transaction to account

			DocumentID:   documentID,
			ParsingRunID: parsingRunID,

			TransactionDate: txDate,

			Amount:   new(big.Rat).SetFloat64(t.Amount),
			Currency: t.Currency,

			BalanceAfter: balanceAfter,

			Direction: dir,

			RawDescription:        t.Description,
			NormalizedDescription: normalizedDescription,

			CategoryID:      categoryID,
			CategoryName:    categoryName,
			SubcategoryName: subcategoryName,
		}

		rows = append(rows, row)
	}

	if err := repo.InsertTransactions(ctx, rows); err != nil {
		return fmt.Errorf("insertTransactions: inserting rows: %w", err)
	}

	return nil
}
