package pipeline_test

import (
	"context"
	"testing"

	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/dvloznov/finance-tracker/internal/pipeline"
)

// MockDocumentRepository is a mock implementation of DocumentRepository for testing.
type MockDocumentRepository struct {
	InsertDocumentFunc          func(ctx context.Context, row interface{}) error
	InsertTransactionsFunc      func(ctx context.Context, rows interface{}) error
	InsertModelOutputFunc       func(ctx context.Context, row interface{}) error
	StartParsingRunFunc         func(ctx context.Context, documentID string) (string, error)
	MarkParsingRunFailedFunc    func(ctx context.Context, parsingRunID string, parseErr error)
	MarkParsingRunSucceededFunc func(ctx context.Context, parsingRunID string) error
	ListActiveCategoriesFunc    func(ctx context.Context) (interface{}, error)
}

// MockStorageService is a mock implementation of StorageService for testing.
type MockStorageService struct {
	UploadFileFunc                func(ctx context.Context, bucketName, objectName, filePath string) error
	FetchFromGCSFunc              func(ctx context.Context, gcsURI string) ([]byte, error)
	ExtractFilenameFromGCSURIFunc func(uri string) string
}

func (m *MockStorageService) UploadFile(ctx context.Context, bucketName, objectName, filePath string) error {
	if m.UploadFileFunc != nil {
		return m.UploadFileFunc(ctx, bucketName, objectName, filePath)
	}
	return nil
}

func (m *MockStorageService) FetchFromGCS(ctx context.Context, gcsURI string) ([]byte, error) {
	if m.FetchFromGCSFunc != nil {
		return m.FetchFromGCSFunc(ctx, gcsURI)
	}
	return []byte("mock pdf data"), nil
}

func (m *MockStorageService) ExtractFilenameFromGCSURI(uri string) string {
	if m.ExtractFilenameFromGCSURIFunc != nil {
		return m.ExtractFilenameFromGCSURIFunc(uri)
	}
	return "mock-file.pdf"
}

// MockAccountRepository is a mock implementation of AccountRepository for testing.
type MockAccountRepository struct {
	UpsertAccountFunc                  func(ctx context.Context, row *bigquery.AccountRow) (string, error)
	FindAccountByNumberAndCurrencyFunc func(ctx context.Context, accountNumber, currency string) (*bigquery.AccountRow, error)
	ListAllAccountsFunc                func(ctx context.Context) ([]*bigquery.AccountRow, error)
}

func (m *MockAccountRepository) UpsertAccount(ctx context.Context, row *bigquery.AccountRow) (string, error) {
	if m.UpsertAccountFunc != nil {
		return m.UpsertAccountFunc(ctx, row)
	}
	return "mock-account-id", nil
}

func (m *MockAccountRepository) FindAccountByNumberAndCurrency(ctx context.Context, accountNumber, currency string) (*bigquery.AccountRow, error) {
	if m.FindAccountByNumberAndCurrencyFunc != nil {
		return m.FindAccountByNumberAndCurrencyFunc(ctx, accountNumber, currency)
	}
	return nil, nil
}

func (m *MockAccountRepository) ListAllAccounts(ctx context.Context) ([]*bigquery.AccountRow, error) {
	if m.ListAllAccountsFunc != nil {
		return m.ListAllAccountsFunc(ctx)
	}
	return []*bigquery.AccountRow{}, nil
}

// MockAIParser is a mock implementation of AIParser for testing.
type MockAIParser struct {
	ParseStatementFunc       func(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error)
	ExtractAccountHeaderFunc func(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error)
}

func (m *MockAIParser) ParseStatement(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error) {
	if m.ParseStatementFunc != nil {
		return m.ParseStatementFunc(ctx, pdfBytes)
	}
	return map[string]interface{}{
		"transactions": []interface{}{},
	}, nil
}

func (m *MockAIParser) ExtractAccountHeader(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error) {
	if m.ExtractAccountHeaderFunc != nil {
		return m.ExtractAccountHeaderFunc(ctx, pdfBytes)
	}
	// Default mock account data
	return map[string]interface{}{
		"account_number": "12345678",
		"currency":       "GBP",
		"account_name":   "Current Account",
		"account_type":   "CURRENT",
	}, nil
}

// TestPipelineWithMocks demonstrates how to use the interfaces for testing.
// This test shows the structure but doesn't actually test the full pipeline
// to keep it simple and avoid BigQuery dependencies.
func TestPipelineWithMocks(t *testing.T) {
	// Create mock dependencies
	mockStorage := &MockStorageService{}
	mockAIParser := &MockAIParser{}

	// Verify that our mock types implement the required interfaces
	var _ pipeline.StorageService = mockStorage
	var _ pipeline.AIParser = mockAIParser

	// This demonstrates that the interfaces can be used for testing.
	// In a real test, you would:
	// 1. Create a MockDocumentRepository
	// 2. Set up the mock behavior
	// 3. Call IngestStatementFromGCSWithDeps with your mocks
	// 4. Assert the expected behavior

	t.Log("Mock implementations successfully created")
	t.Log("Interfaces can now be used for testing with custom mocks")
}
