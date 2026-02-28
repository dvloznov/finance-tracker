package pipeline_test

import (
	"context"
	"errors"
	"testing"

	bigquerylib "cloud.google.com/go/bigquery"
	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/dvloznov/finance-tracker/internal/pipeline"
)

// TestPipelineWithoutCategoryInference tests the full pipeline without category inference/validation
func TestPipelineWithoutCategoryInference(t *testing.T) {
	mockCategories := []bigquery.CategoryRow{
		{CategoryID: "cat1-sub1", CategoryName: "Food & Dining", SubcategoryName: bigquerylib.NullString{StringVal: "Groceries", Valid: true}},
		{CategoryID: "cat_uncategorized_other", CategoryName: "Uncategorized", SubcategoryName: bigquerylib.NullString{StringVal: "Other", Valid: true}},
	}

	var insertedDocuments int
	var startedParsingRuns int
	var insertedTransactions int
	var succeededParsingRuns int

	mockRepo := &MockDocumentRepository{
		InsertDocumentFunc: func(ctx context.Context, row interface{}) error {
			insertedDocuments++
			return nil
		},
		StartParsingRunFunc: func(ctx context.Context, documentID string) (string, error) {
			startedParsingRuns++
			return "test-parsing-run-id", nil
		},
		InsertModelOutputFunc: func(ctx context.Context, row interface{}) error {
			return nil
		},
		InsertTransactionsFunc: func(ctx context.Context, rows interface{}) error {
			insertedTransactions++
			return nil
		},
		MarkParsingRunSucceededFunc: func(ctx context.Context, parsingRunID string) error {
			succeededParsingRuns++
			return nil
		},
		MarkParsingRunFailedFunc: func(ctx context.Context, parsingRunID string, parseErr error) error {
			return nil
		},
		ListActiveCategoriesFunc: func(ctx context.Context) (interface{}, error) {
			return mockCategories, nil
		},
	}

	mockStorage := &MockStorageService{
		FetchFromGCSFunc: func(ctx context.Context, gcsURI string) ([]byte, error) {
			return []byte("mock pdf data"), nil
		},
		ExtractFilenameFromGCSURIFunc: func(uri string) string {
			return "test.pdf"
		},
	}

	t.Run("NoCategoriesProvided", func(t *testing.T) {
		insertedDocuments = 0
		startedParsingRuns = 0
		insertedTransactions = 0
		succeededParsingRuns = 0

		mockAIParser := &MockAIParser{
			ParseStatementFunc: func(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error) {
				return map[string]interface{}{
					"transactions": []interface{}{
						map[string]interface{}{
							"date":          "2024-01-01",
							"description":   "Test transaction",
							"amount":        -10.50,
							"currency":      "GBP",
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
			CategorizeMerchantsFunc: func(ctx context.Context, merchantNames []string, categories []bigquery.CategoryRow) (map[string]string, error) {
				result := make(map[string]string, len(merchantNames))
				for _, name := range merchantNames {
					result[name] = pipeline.DefaultCategoryID
				}
				return result, nil
			},
		}

		mockAccountRepo := &MockAccountRepository{
			UpsertAccountFunc: func(ctx context.Context, row *bigquery.AccountRow) (string, error) {
				return "test-account-id", nil
			},
		}

		mockInstitutionRepo := &MockInstitutionRepository{
			UpsertInstitutionFunc: func(ctx context.Context, row *bigquery.InstitutionRow) (string, error) {
				return "test-institution-id", nil
			},
		}

		mockMerchantRepo := &MockMerchantRepository{}

		repo := &mockDocumentRepo{MockDocumentRepository: mockRepo}
		err := pipeline.IngestStatementFromGCSWithDeps(
			context.Background(),
			"gs://test-bucket/test.pdf",
			"",
			repo,
			mockAccountRepo,
			mockInstitutionRepo,
			mockMerchantRepo,
			mockStorage,
			mockAIParser,
		)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if insertedDocuments != 1 {
			t.Errorf("expected 1 document inserted, got %d", insertedDocuments)
		}
		if startedParsingRuns != 1 {
			t.Errorf("expected 1 parsing run started, got %d", startedParsingRuns)
		}
		if insertedTransactions != 1 {
			t.Errorf("expected 1 transaction batch inserted, got %d", insertedTransactions)
		}
		if succeededParsingRuns != 1 {
			t.Errorf("expected 1 parsing run marked succeeded, got %d", succeededParsingRuns)
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

func (m *mockDocumentRepo) MarkParsingRunFailed(ctx context.Context, parsingRunID string, parseErr error) error {
	if m.MarkParsingRunFailedFunc != nil {
		m.MarkParsingRunFailedFunc(ctx, parsingRunID, parseErr)
	}
	return nil
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

func (m *mockDocumentRepo) QueryTransactions(ctx context.Context, opts bigquery.TransactionQuery) ([]*bigquery.TransactionRow, error) {
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

func (m *mockDocumentRepo) MarkParsingRunsAsSuperseded(ctx context.Context, documentID string) error {
	// For tests, just return success
	return nil
}

func (m *mockDocumentRepo) UpdateDocumentAccountAndInstitution(ctx context.Context, documentID, accountID, institutionID string) error {
	return nil
}

func (m *mockDocumentRepo) GetDocumentByID(ctx context.Context, documentID string) (*bigquery.DocumentRow, error) {
	return nil, nil
}

func (m *mockDocumentRepo) DeleteDocument(ctx context.Context, documentID string) error {
	return nil
}

func (m *mockDocumentRepo) UpdateDocumentParsingStatus(ctx context.Context, documentID, status string) error {
	return nil
}

func (m *mockDocumentRepo) Close() error {
	return nil
}

// MockMerchantRepository is a mock implementation of MerchantRepository for testing.
type MockMerchantRepository struct {
	FindMerchantByNormalizedNameFunc func(ctx context.Context, normalizedName string) (*bigquery.MerchantRow, error)
	InsertMerchantFunc               func(ctx context.Context, row *bigquery.MerchantRow) (string, error)
}

func (m *MockMerchantRepository) FindMerchantByNormalizedName(ctx context.Context, normalizedName string) (*bigquery.MerchantRow, error) {
	if m.FindMerchantByNormalizedNameFunc != nil {
		return m.FindMerchantByNormalizedNameFunc(ctx, normalizedName)
	}
	return nil, nil
}

func (m *MockMerchantRepository) InsertMerchant(ctx context.Context, row *bigquery.MerchantRow) (string, error) {
	if m.InsertMerchantFunc != nil {
		return m.InsertMerchantFunc(ctx, row)
	}
	if row.MerchantID != "" {
		return row.MerchantID, nil
	}
	return "test-merchant-id", nil
}
