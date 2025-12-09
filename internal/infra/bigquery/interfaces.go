package bigquery

import (
	"context"
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
}

// BigQueryDocumentRepository is the concrete implementation of DocumentRepository
// that interacts with BigQuery.
type BigQueryDocumentRepository struct{}

// NewBigQueryDocumentRepository creates a new instance of BigQueryDocumentRepository.
func NewBigQueryDocumentRepository() *BigQueryDocumentRepository {
	return &BigQueryDocumentRepository{}
}

// InsertDocument delegates to the existing InsertDocument function.
func (r *BigQueryDocumentRepository) InsertDocument(ctx context.Context, row *DocumentRow) error {
	return InsertDocument(ctx, row)
}

// InsertTransactions delegates to the existing InsertTransactions function.
func (r *BigQueryDocumentRepository) InsertTransactions(ctx context.Context, rows []*TransactionRow) error {
	return InsertTransactions(ctx, rows)
}

// InsertModelOutput delegates to the existing InsertModelOutput function.
func (r *BigQueryDocumentRepository) InsertModelOutput(ctx context.Context, row *ModelOutputRow) error {
	return InsertModelOutput(ctx, row)
}

// StartParsingRun delegates to the existing StartParsingRun function.
func (r *BigQueryDocumentRepository) StartParsingRun(ctx context.Context, documentID string) (string, error) {
	return StartParsingRun(ctx, documentID)
}

// MarkParsingRunFailed delegates to the existing MarkParsingRunFailed function.
func (r *BigQueryDocumentRepository) MarkParsingRunFailed(ctx context.Context, parsingRunID string, parseErr error) {
	MarkParsingRunFailed(ctx, parsingRunID, parseErr)
}

// MarkParsingRunSucceeded delegates to the existing MarkParsingRunSucceeded function.
func (r *BigQueryDocumentRepository) MarkParsingRunSucceeded(ctx context.Context, parsingRunID string) error {
	return MarkParsingRunSucceeded(ctx, parsingRunID)
}

// ListActiveCategories delegates to the existing ListActiveCategories function.
func (r *BigQueryDocumentRepository) ListActiveCategories(ctx context.Context) ([]CategoryRow, error) {
	return ListActiveCategories(ctx)
}
