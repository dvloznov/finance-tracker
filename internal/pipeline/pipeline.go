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
// documentID is optional - if provided, it will use the existing document record instead of creating a new one.
func IngestStatementFromGCS(ctx context.Context, gcsURI string, documentID ...string) error {
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

	institutionRepo, err := infraBQ.NewBigQueryInstitutionRepository(ctx)
	if err != nil {
		return fmt.Errorf("IngestStatementFromGCS: creating BigQuery institution repository: %w", err)
	}
	defer institutionRepo.Close()

	merchantRepo, err := infraBQ.NewBigQueryMerchantRepository(ctx)
	if err != nil {
		return fmt.Errorf("IngestStatementFromGCS: creating BigQuery merchant repository: %w", err)
	}
	defer merchantRepo.Close()

	storage := &gcsuploader.GCSStorageService{}
	aiParser := NewGeminiAIParser(repo)

	// Use provided documentID if available
	var docID string
	if len(documentID) > 0 && documentID[0] != "" {
		docID = documentID[0]
	}

	return IngestStatementFromGCSWithDeps(ctx, gcsURI, docID, repo, accountRepo, institutionRepo, merchantRepo, storage, aiParser)
}

// IngestStatementFromGCSWithDeps processes a single bank statement PDF stored in GCS
// using the provided dependencies. This enables dependency injection for testing.
// If documentID is provided, it will use that existing document instead of creating a new one.
func IngestStatementFromGCSWithDeps(
	ctx context.Context,
	gcsURI string,
	documentID string,
	repo bigquery.DocumentRepository,
	accountRepo bigquery.AccountRepository,
	institutionRepo bigquery.InstitutionRepository,
	merchantRepo bigquery.MerchantRepository,
	storage StorageService,
	aiParser AIParser,
) error {
	// Initialize pipeline state
	state := &PipelineState{
		GCSURI:          gcsURI,
		DocumentID:      documentID, // Set documentID if provided
		DocumentRepo:    repo,
		AccountRepo:     accountRepo,
		InstitutionRepo: institutionRepo,
		MerchantRepo:    merchantRepo,
		StorageService:  storage,
		AIParser:        aiParser,
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

// storeModelOutputWithRepo inserts raw model output into the model_outputs table.
// operation: e.g. extract_account_header, parse_statement, categorize_merchants
func storeModelOutputWithRepo(
	ctx context.Context,
	parsingRunID string,
	documentID string,
	operation string,
	prompt string,
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
		Operation:    operation,
		Prompt:       prompt,
		ModelName:    DefaultModelName,
		CreatedTS: bigquerylib.NullTimestamp{
			Timestamp: time.Now(),
			Valid:     true,
		},
		RawJSON: bigquerylib.NullJSON{
			JSONVal: string(jsonBytes),
			Valid:   true,
		},
	}

	if err := repo.InsertModelOutput(ctx, row); err != nil {
		return "", fmt.Errorf("storeModelOutput: inserting row: %w", err)
	}

	return outputID, nil
}

// insertTransactionsWithRepo writes a batch of transactions to the transactions table using the provided repository.
func insertTransactionsWithRepo(
	ctx context.Context,
	documentID string,
	parsingRunID string,
	accountID string,
	institutionID string,
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
		statementDate := civil.DateOf(t.StatementDate)

		var balanceAfter *big.Rat
		if t.BalanceAfter != nil {
			balanceAfter = new(big.Rat).SetFloat64(*t.BalanceAfter)
		}

		var transactionType bigquerylib.NullString
		if strings.TrimSpace(t.TransactionType) != "" {
			transactionType = bigquerylib.NullString{
				StringVal: t.TransactionType,
				Valid:     true,
			}
		}

		merchantID := strings.TrimSpace(t.MerchantID)
		if merchantID == "" {
			return fmt.Errorf("insertTransactions: missing merchant_id for transaction %s", t.Description)
		}

		row := &bigquery.TransactionRow{
			TransactionID: uuid.NewString(),
			UserID:        DefaultUserID,
			AccountID:     accountID,
			InstitutionID: institutionID,
			DocumentID:    documentID,
			ParsingRunID:  parsingRunID,

			TransactionDate: txDate,
			StatementDate:   statementDate,
			TransactionType: transactionType,

			Amount:   new(big.Rat).SetFloat64(t.Amount),
			Currency: t.Currency,

			BalanceAfter: balanceAfter,

			Direction: dir,

			RawDescription: t.Description,
			MerchantID:     merchantID,

			CreatedTS: time.Now(),
		}

		rows = append(rows, row)
	}

	if err := repo.InsertTransactions(ctx, rows); err != nil {
		return fmt.Errorf("insertTransactions: inserting rows: %w", err)
	}

	return nil
}
