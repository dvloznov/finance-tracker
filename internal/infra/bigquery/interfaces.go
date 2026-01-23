package bigquery

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	bq "github.com/dvloznov/finance-tracker/internal/bigquery"
)

// Removed re-exported interfaces for backward compatibility

// BigQueryAccountRepository is the concrete implementation of AccountRepository
// that interacts with BigQuery.
type BigQueryAccountRepository struct {
	client *bigquery.Client
}

// BigQueryInstitutionRepository is the concrete implementation of InstitutionRepository
// that interacts with BigQuery.
type BigQueryInstitutionRepository struct {
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

// NewBigQueryInstitutionRepository creates a new instance of BigQueryInstitutionRepository
// with a shared BigQuery client.
func NewBigQueryInstitutionRepository(ctx context.Context) (*BigQueryInstitutionRepository, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("NewBigQueryInstitutionRepository: creating client: %w", err)
	}
	return &BigQueryInstitutionRepository{
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

// Close closes the BigQuery client connection.
func (r *BigQueryInstitutionRepository) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// UpsertAccount delegates to the existing UpsertAccount function with the shared client.
func (r *BigQueryAccountRepository) UpsertAccount(ctx context.Context, row *bq.AccountRow) (string, error) {
	return UpsertAccountWithClient(ctx, r.client, row)
}

// FindAccountByNumberAndCurrency delegates to the existing function with the shared client.
func (r *BigQueryAccountRepository) FindAccountByNumberAndCurrency(ctx context.Context, accountNumber, currency, institutionID string) (*bq.AccountRow, error) {
	return FindAccountByNumberAndCurrencyWithClient(ctx, r.client, accountNumber, currency, institutionID)
}

// ListAllAccounts delegates to the existing ListAllAccounts function with the shared client.
func (r *BigQueryAccountRepository) ListAllAccounts(ctx context.Context) ([]*bq.AccountRow, error) {
	return ListAllAccountsWithClient(ctx, r.client)
}

// UpsertInstitution delegates to the existing UpsertInstitution function with the shared client.
func (r *BigQueryInstitutionRepository) UpsertInstitution(ctx context.Context, row *bq.InstitutionRow) (string, error) {
	return UpsertInstitutionWithClient(ctx, r.client, row)
}

// ListAllInstitutions delegates to the existing ListAllInstitutions function with the shared client.
func (r *BigQueryInstitutionRepository) ListAllInstitutions(ctx context.Context) ([]*bq.InstitutionRow, error) {
	return ListAllInstitutionsWithClient(ctx, r.client)
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
func (r *BigQueryDocumentRepository) InsertDocument(ctx context.Context, row *bq.DocumentRow) error {
	return InsertDocumentWithClient(ctx, r.client, row)
}

// InsertTransactions delegates to the existing InsertTransactions function with the shared client.
func (r *BigQueryDocumentRepository) InsertTransactions(ctx context.Context, rows []*bq.TransactionRow) error {
	return InsertTransactionsWithClient(ctx, r.client, rows)
}

// InsertModelOutput delegates to the existing InsertModelOutput function with the shared client.
func (r *BigQueryDocumentRepository) InsertModelOutput(ctx context.Context, row *bq.ModelOutputRow) error {
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
func (r *BigQueryDocumentRepository) ListActiveCategories(ctx context.Context) ([]bq.CategoryRow, error) {
	return ListActiveCategoriesWithClient(ctx, r.client)
}

// QueryTransactionsByDateRange delegates to the existing QueryTransactionsByDateRange function with the shared client.
func (r *BigQueryDocumentRepository) QueryTransactionsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*bq.TransactionRow, error) {
	return QueryTransactionsByDateRangeWithClient(ctx, r.client, startDate, endDate)
}

// ListAllAccounts delegates to the existing ListAllAccounts function with the shared client.
func (r *BigQueryDocumentRepository) ListAllAccounts(ctx context.Context) ([]*bq.AccountRow, error) {
	return ListAllAccountsWithClient(ctx, r.client)
}

// ListAllDocuments delegates to the existing ListAllDocuments function with the shared client.
func (r *BigQueryDocumentRepository) ListAllDocuments(ctx context.Context) ([]*bq.DocumentRow, error) {
	return ListAllDocumentsWithClient(ctx, r.client)
}

// MarkParsingRunsAsSuperseded delegates to the existing MarkParsingRunsAsSuperseded function with the shared client.
func (r *BigQueryDocumentRepository) MarkParsingRunsAsSuperseded(ctx context.Context, documentID string) error {
	return MarkParsingRunsAsSupersededWithClient(ctx, r.client, documentID)
}

// UpdateDocumentAccountAndInstitution delegates to the existing UpdateDocumentAccountAndInstitution function with the shared client.
func (r *BigQueryDocumentRepository) UpdateDocumentAccountAndInstitution(ctx context.Context, documentID, accountID, institutionID string) error {
	return UpdateDocumentAccountAndInstitutionWithClient(ctx, r.client, documentID, accountID, institutionID)
}
