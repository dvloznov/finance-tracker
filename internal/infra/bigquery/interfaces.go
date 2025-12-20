package bigquery

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	bq "github.com/dvloznov/finance-tracker/internal/bigquery"
)

// Re-export interfaces from shared package for backward compatibility
type DocumentRepository = bq.DocumentRepository
type AccountRepository = bq.AccountRepository
type CategoryRepository = bq.CategoryRepository

// BigQueryAccountRepository is the concrete implementation of AccountRepository
// that interacts with BigQuery.
type BigQueryAccountRepository struct {
	client *bigquery.Client
}

// NewBigQueryAccountRepository creates a new instance of BigQueryAccountRepository
// with a shared BigQuery client.
func NewBigQueryAccountRepository(ctx context.Context) (*BigQueryAccountRepository, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("NewBigQueryAccountRepository: creating client: %w", err)
	}
	return &BigQueryAccountRepository{
		client: client,
	}, nil
}

// Close closes the BigQuery client connection.
func (r *BigQueryAccountRepository) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// UpsertAccount delegates to the existing UpsertAccount function with the shared client.
func (r *BigQueryAccountRepository) UpsertAccount(ctx context.Context, row *AccountRow) (string, error) {
	return UpsertAccountWithClient(ctx, r.client, row)
}

// FindAccountByNumberAndCurrency delegates to the existing function with the shared client.
func (r *BigQueryAccountRepository) FindAccountByNumberAndCurrency(ctx context.Context, accountNumber, currency string) (*AccountRow, error) {
	return FindAccountByNumberAndCurrencyWithClient(ctx, r.client, accountNumber, currency)
}

// ListAllAccounts delegates to the existing ListAllAccounts function with the shared client.
func (r *BigQueryAccountRepository) ListAllAccounts(ctx context.Context) ([]*AccountRow, error) {
	return ListAllAccountsWithClient(ctx, r.client)
}

// BigQueryDocumentRepository is the concrete implementation of DocumentRepository
// that interacts with BigQuery. It holds a shared BigQuery client to avoid
// creating a new connection for each operation.
type BigQueryDocumentRepository struct {
	client *bigquery.Client
}

// NewBigQueryDocumentRepository creates a new instance of BigQueryDocumentRepository
// with a shared BigQuery client.
func NewBigQueryDocumentRepository(ctx context.Context) (*BigQueryDocumentRepository, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("NewBigQueryDocumentRepository: creating client: %w", err)
	}
	return &BigQueryDocumentRepository{
		client: client,
	}, nil
}

// Close closes the BigQuery client connection. This should be called when
// the repository is no longer needed to release resources.
func (r *BigQueryDocumentRepository) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// InsertDocument delegates to the existing InsertDocument function with the shared client.
func (r *BigQueryDocumentRepository) InsertDocument(ctx context.Context, row *DocumentRow) error {
	return InsertDocumentWithClient(ctx, r.client, row)
}

// InsertTransactions delegates to the existing InsertTransactions function with the shared client.
func (r *BigQueryDocumentRepository) InsertTransactions(ctx context.Context, rows []*TransactionRow) error {
	return InsertTransactionsWithClient(ctx, r.client, rows)
}

// InsertModelOutput delegates to the existing InsertModelOutput function with the shared client.
func (r *BigQueryDocumentRepository) InsertModelOutput(ctx context.Context, row *ModelOutputRow) error {
	return InsertModelOutputWithClient(ctx, r.client, row)
}

// StartParsingRun delegates to the existing StartParsingRun function with the shared client.
func (r *BigQueryDocumentRepository) StartParsingRun(ctx context.Context, documentID string) (string, error) {
	return StartParsingRunWithClient(ctx, r.client, documentID)
}

// MarkParsingRunFailed delegates to the existing MarkParsingRunFailed function with the shared client.
func (r *BigQueryDocumentRepository) MarkParsingRunFailed(ctx context.Context, parsingRunID string, parseErr error) {
	MarkParsingRunFailedWithClient(ctx, r.client, parsingRunID, parseErr)
}

// MarkParsingRunSucceeded delegates to the existing MarkParsingRunSucceeded function with the shared client.
func (r *BigQueryDocumentRepository) MarkParsingRunSucceeded(ctx context.Context, parsingRunID string) error {
	return MarkParsingRunSucceededWithClient(ctx, r.client, parsingRunID)
}

// ListActiveCategories delegates to the existing ListActiveCategories function with the shared client.
func (r *BigQueryDocumentRepository) ListActiveCategories(ctx context.Context) ([]CategoryRow, error) {
	return ListActiveCategoriesWithClient(ctx, r.client)
}

// QueryTransactionsByDateRange delegates to the existing QueryTransactionsByDateRange function with the shared client.
func (r *BigQueryDocumentRepository) QueryTransactionsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*TransactionRow, error) {
	return QueryTransactionsByDateRangeWithClient(ctx, r.client, startDate, endDate)
}

// ListAllAccounts delegates to the existing ListAllAccounts function with the shared client.
func (r *BigQueryDocumentRepository) ListAllAccounts(ctx context.Context) ([]*AccountRow, error) {
	return ListAllAccountsWithClient(ctx, r.client)
}

// ListAllDocuments delegates to the existing ListAllDocuments function with the shared client.
func (r *BigQueryDocumentRepository) ListAllDocuments(ctx context.Context) ([]*DocumentRow, error) {
	return ListAllDocumentsWithClient(ctx, r.client)
}

// FindDocumentByChecksum delegates to the existing FindDocumentByChecksum function with the shared client.
func (r *BigQueryDocumentRepository) FindDocumentByChecksum(ctx context.Context, checksum string) (*DocumentRow, error) {
	return FindDocumentByChecksumWithClient(ctx, r.client, checksum)
}

// MarkParsingRunsAsSuperseded delegates to the existing MarkParsingRunsAsSuperseded function with the shared client.
func (r *BigQueryDocumentRepository) MarkParsingRunsAsSuperseded(ctx context.Context, documentID string) error {
	return MarkParsingRunsAsSupersededWithClient(ctx, r.client, documentID)
}
