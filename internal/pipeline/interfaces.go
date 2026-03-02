package pipeline

import (
	"context"

	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/dvloznov/finance-tracker/internal/gcs"
)

// StorageService is an interface for storage operations.
type StorageService = gcs.StorageService

// CategoryRepository is an interface for category-related database operations.
type CategoryRepository = bigquery.CategoryRepository

// AIParser provides an interface for AI-powered document parsing operations.
// This interface enables mocking and testing of AI parsing functionality.
// All methods return the prompt sent to the model as the second value for audit logging.
type AIParser interface {
	// ParseStatement sends PDF bytes to an AI model and returns parsed JSON output and the prompt used.
	ParseStatement(ctx context.Context, pdfBytes []byte) (map[string]interface{}, string, error)

	// ExtractAccountHeader sends PDF bytes to an AI model to extract account metadata from the header.
	// Returns the parsed output and the prompt used.
	ExtractAccountHeader(ctx context.Context, pdfBytes []byte) (map[string]interface{}, string, error)

	// CategorizeMerchants assigns category_ids to merchants using the provided categories list.
	// Returns merchant_name->category_id map, the prompt used, and error.
	CategorizeMerchants(ctx context.Context, merchantNames []string, categories []bigquery.CategoryRow) (map[string]string, string, error)
}

// GeminiAIParser is the concrete implementation of AIParser that uses Gemini AI.
type GeminiAIParser struct {
	repo     CategoryRepository
	bankName string // e.g. "Barclays UK" — used to build parser prompts
}

// NewGeminiAIParser creates a new instance of GeminiAIParser.
// bankName is used in the statement-parsing prompt; set via BQ_BANK_NAME env var or leave empty
// to use the DefaultBankName constant.
func NewGeminiAIParser(repo CategoryRepository) *GeminiAIParser {
	return &GeminiAIParser{
		repo:     repo,
		bankName: DefaultBankName,
	}
}

// ParseStatement delegates to the existing parseStatementWithModel function.
func (p *GeminiAIParser) ParseStatement(ctx context.Context, pdfBytes []byte) (map[string]interface{}, string, error) {
	return parseStatementWithModel(ctx, pdfBytes, p.repo, p.bankName)
}

// ExtractAccountHeader calls the AI model to extract account metadata from the statement header.
func (p *GeminiAIParser) ExtractAccountHeader(ctx context.Context, pdfBytes []byte) (map[string]interface{}, string, error) {
	return extractAccountHeaderWithModel(ctx, pdfBytes)
}

// CategorizeMerchants calls the AI model to classify merchants into category_ids.
func (p *GeminiAIParser) CategorizeMerchants(ctx context.Context, merchantNames []string, categories []bigquery.CategoryRow) (map[string]string, string, error) {
	return categorizeMerchantsWithModel(ctx, merchantNames, categories)
}
