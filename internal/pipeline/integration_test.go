package pipeline_test

import (
	"context"
	"errors"
	"testing"
	"time"

	bigquerylib "cloud.google.com/go/bigquery"
	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/dvloznov/finance-tracker/internal/pipeline"
)

// TestPipelineWithCategoryValidation tests the full pipeline with category validation
func TestPipelineWithCategoryValidation(t *testing.T) {
	// Setup mock categories
	mockCategories := []bigquery.CategoryRow{
		{CategoryID: "cat1-sub1", CategoryName: "Food & Dining", SubcategoryName: bigquerylib.NullString{StringVal: "Groceries", Valid: true}},
		{CategoryID: "cat2-sub1", CategoryName: "Transportation", SubcategoryName: bigquerylib.NullString{StringVal: "Fuel", Valid: true}},
		{CategoryID: "cat_healthcare", CategoryName: "Healthcare", SubcategoryName: bigquerylib.NullString{Valid: false}},
	}

	// Setup mock document repository
	mockRepo := &MockDocumentRepository{
		InsertDocumentFunc: func(ctx context.Context, row interface{}) error {
			return nil
		},
		StartParsingRunFunc: func(ctx context.Context, documentID string) (string, error) {
			return "test-parsing-run-id", nil
		},
		InsertModelOutputFunc: func(ctx context.Context, row interface{}) error {
			return nil
		},
		InsertTransactionsFunc: func(ctx context.Context, rows interface{}) error {
			return nil
		},
		MarkParsingRunSucceededFunc: func(ctx context.Context, parsingRunID string) error {
			return nil
		},
		MarkParsingRunFailedFunc: func(ctx context.Context, parsingRunID string, parseErr error) {
			// Track failures if needed
		},
		ListActiveCategoriesFunc: func(ctx context.Context) (interface{}, error) {
			return mockCategories, nil
		},
	}

	// Setup mock storage
	mockStorage := &MockStorageService{
		FetchFromGCSFunc: func(ctx context.Context, gcsURI string) ([]byte, error) {
			return []byte("mock pdf data"), nil
		},
		ExtractFilenameFromGCSURIFunc: func(uri string) string {
			return "test.pdf"
		},
	}

	// Test case 1: Valid categories
	t.Run("ValidCategories", func(t *testing.T) {
		mockAIParser := &MockAIParser{
			ParseStatementFunc: func(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error) {
				return map[string]interface{}{
					"transactions": []interface{}{
						map[string]interface{}{
							"date":          "2024-01-01",
							"description":   "Test transaction",
							"amount":        -10.50,
							"currency":      "GBP",
							"category":      "Food & Dining",
							"subcategory":   "Groceries",
							"balance_after": 100.0,
						},
					},
				}, nil
			},
			ExtractAccountHeaderFunc: func(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error) {
				return map[string]interface{}{
					"account_number": "12345678",
					"currency":       "GBP",
				}, nil
			},
		}

		mockAccountRepo := &MockAccountRepository{
			UpsertAccountFunc: func(ctx context.Context, row *bigquery.AccountRow) (string, error) {
				return "test-account-id", nil
			},
		}

		repo := &mockDocumentRepo{MockDocumentRepository: mockRepo}
		err := pipeline.IngestStatementFromGCSWithDeps(
			context.Background(),
			"gs://test-bucket/test.pdf",
			repo,
			mockAccountRepo,
			mockStorage,
			mockAIParser,
		)

		if err != nil {
			t.Errorf("Expected no error with valid categories, got: %v", err)
		}
	})

	// Test case 2: Invalid category
	t.Run("InvalidCategory", func(t *testing.T) {
		mockAIParser := &MockAIParser{
			ParseStatementFunc: func(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error) {
				return map[string]interface{}{
					"transactions": []interface{}{
						map[string]interface{}{
							"date":          "2024-01-01",
							"description":   "Test transaction",
							"amount":        -10.50,
							"currency":      "GBP",
							"category":      "INVALID_CATEGORY",
							"subcategory":   "Groceries",
							"balance_after": 100.0,
						},
					},
				}, nil
			},
			ExtractAccountHeaderFunc: func(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error) {
				return map[string]interface{}{
					"account_number": "12345678",
					"currency":       "GBP",
				}, nil
			},
		}

		mockAccountRepo := &MockAccountRepository{
			UpsertAccountFunc: func(ctx context.Context, row *bigquery.AccountRow) (string, error) {
				return "test-account-id", nil
			},
		}

		repo := &mockDocumentRepo{MockDocumentRepository: mockRepo}
		err := pipeline.IngestStatementFromGCSWithDeps(
			context.Background(),
			"gs://test-bucket/test.pdf",
			repo,
			mockAccountRepo,
			mockStorage,
			mockAIParser,
		)

		if err == nil {
			t.Error("Expected error with invalid category, got nil")
		}
	})

	// Test case 3: Invalid subcategory
	t.Run("InvalidSubcategory", func(t *testing.T) {
		mockAIParser := &MockAIParser{
			ParseStatementFunc: func(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error) {
				return map[string]interface{}{
					"transactions": []interface{}{
						map[string]interface{}{
							"date":          "2024-01-01",
							"description":   "Test transaction",
							"amount":        -10.50,
							"currency":      "GBP",
							"category":      "Food & Dining",
							"subcategory":   "Fuel", // Wrong subcategory for Food & Dining
							"balance_after": 100.0,
						},
					},
				}, nil
			},
			ExtractAccountHeaderFunc: func(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error) {
				return map[string]interface{}{
					"account_number": "12345678",
					"currency":       "GBP",
				}, nil
			},
		}

		mockAccountRepo := &MockAccountRepository{
			UpsertAccountFunc: func(ctx context.Context, row *bigquery.AccountRow) (string, error) {
				return "test-account-id", nil
			},
		}

		repo := &mockDocumentRepo{MockDocumentRepository: mockRepo}
		err := pipeline.IngestStatementFromGCSWithDeps(
			context.Background(),
			"gs://test-bucket/test.pdf",
			repo,
			mockAccountRepo,
			mockStorage,
			mockAIParser,
		)

		if err == nil {
			t.Error("Expected error with invalid subcategory, got nil")
		}
	})
}

// mockDocumentRepo implements both DocumentRepository and CategoryRepository interfaces
type mockDocumentRepo struct {
	*MockDocumentRepository
}

func (m *mockDocumentRepo) InsertDocument(ctx context.Context, row *bigquery.DocumentRow) error {
	if m.InsertDocumentFunc != nil {
		return m.InsertDocumentFunc(ctx, row)
	}
	return nil
}

func (m *mockDocumentRepo) InsertTransactions(ctx context.Context, rows []*bigquery.TransactionRow) error {
	if m.InsertTransactionsFunc != nil {
		return m.InsertTransactionsFunc(ctx, rows)
	}
	return nil
}

func (m *mockDocumentRepo) InsertModelOutput(ctx context.Context, row *bigquery.ModelOutputRow) error {
	if m.InsertModelOutputFunc != nil {
		return m.InsertModelOutputFunc(ctx, row)
	}
	return nil
}

func (m *mockDocumentRepo) StartParsingRun(ctx context.Context, documentID string) (string, error) {
	if m.StartParsingRunFunc != nil {
		return m.StartParsingRunFunc(ctx, documentID)
	}
	return "test-run-id", nil
}

func (m *mockDocumentRepo) MarkParsingRunFailed(ctx context.Context, parsingRunID string, parseErr error) {
	if m.MarkParsingRunFailedFunc != nil {
		m.MarkParsingRunFailedFunc(ctx, parsingRunID, parseErr)
	}
}

func (m *mockDocumentRepo) MarkParsingRunSucceeded(ctx context.Context, parsingRunID string) error {
	if m.MarkParsingRunSucceededFunc != nil {
		return m.MarkParsingRunSucceededFunc(ctx, parsingRunID)
	}
	return nil
}

func (m *mockDocumentRepo) ListActiveCategories(ctx context.Context) ([]bigquery.CategoryRow, error) {
	if m.ListActiveCategoriesFunc != nil {
		result, err := m.ListActiveCategoriesFunc(ctx)
		if err != nil {
			return nil, err
		}
		if categories, ok := result.([]bigquery.CategoryRow); ok {
			return categories, nil
		}
		return nil, errors.New("invalid categories type")
	}
	return nil, nil
}

func (m *mockDocumentRepo) QueryTransactionsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*bigquery.TransactionRow, error) {
	// Not needed for pipeline tests, return empty slice
	return []*bigquery.TransactionRow{}, nil
}

func (m *mockDocumentRepo) ListAllAccounts(ctx context.Context) ([]*bigquery.AccountRow, error) {
	// Not needed for pipeline tests, return empty slice
	return []*bigquery.AccountRow{}, nil
}

func (m *mockDocumentRepo) ListAllDocuments(ctx context.Context) ([]*bigquery.DocumentRow, error) {
	// Not needed for pipeline tests, return empty slice
	return []*bigquery.DocumentRow{}, nil
}

func (m *mockDocumentRepo) FindDocumentByChecksum(ctx context.Context, checksum string) (*bigquery.DocumentRow, error) {
	// For tests, return nil to indicate no duplicate found
	return nil, nil
}

func (m *mockDocumentRepo) MarkParsingRunsAsSuperseded(ctx context.Context, documentID string) error {
	// For tests, just return success
	return nil
}

func (m *mockDocumentRepo) Close() error {
	return nil
}
