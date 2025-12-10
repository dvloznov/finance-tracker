package bigquery

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
)

// DocumentRepository provides an interface for document-related database operations.
// This interface enables mocking and testing of document storage functionality.
type DocumentRepository interface {
	// InsertDocument inserts a single DocumentRow into the database.
	InsertDocument(ctx context.Context, row *DocumentRow) error

	// InsertTransactions inserts a batch of TransactionRow into the database.
	InsertTransactions(ctx context.Context, rows []*TransactionRow) error

	// InsertModelOutput inserts a single ModelOutputRow into the database.
	InsertModelOutput(ctx context.Context, row *ModelOutputRow) error

	// StartParsingRun inserts a new parsing run with status=RUNNING and returns the parsing_run_id.
	StartParsingRun(ctx context.Context, documentID string) (string, error)

	// MarkParsingRunFailed sets status=FAILED, finished_ts and error_message for a parsing run.
	// Note: This method does not return an error. Failures are logged but not propagated
	// to prevent cascading errors during error handling.
	MarkParsingRunFailed(ctx context.Context, parsingRunID string, parseErr error)

	// MarkParsingRunSucceeded sets status=SUCCESS and finished_ts for a parsing run.
	MarkParsingRunSucceeded(ctx context.Context, parsingRunID string) error

	// ListActiveCategories retrieves all active categories from the database.
	ListActiveCategories(ctx context.Context) ([]CategoryRow, error)

	// QueryTransactionsByDateRange queries transactions within the specified date range.
	QueryTransactionsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*TransactionRow, error)
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
