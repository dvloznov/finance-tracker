package pipeline

import (
	"context"

	infra "github.com/dvloznov/finance-tracker/internal/infra/bigquery"
)

// StorageService is an interface for storage operations.
type StorageService interface {
	FetchFromGCS(ctx context.Context, gcsURI string) ([]byte, error)
	ExtractFilenameFromGCSURI(uri string) string
}

// AIParser provides an interface for AI-powered document parsing operations.
// This interface enables mocking and testing of AI parsing functionality.
type AIParser interface {
	// ParseStatement sends PDF bytes to an AI model and returns parsed JSON output.
	ParseStatement(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error)

	// BuildCategoriesPrompt constructs a prompt string containing all active categories
	// and subcategories from the database, formatted for LLM consumption.
	BuildCategoriesPrompt(ctx context.Context) (string, error)
}

// GeminiAIParser is the concrete implementation of AIParser that uses Gemini AI.
type GeminiAIParser struct {
	repo DocumentRepository
}

// NewGeminiAIParser creates a new instance of GeminiAIParser.
func NewGeminiAIParser(repo DocumentRepository) *GeminiAIParser {
	return &GeminiAIParser{
		repo: repo,
	}
}

// ParseStatement delegates to the existing parseStatementWithModel function.
func (p *GeminiAIParser) ParseStatement(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error) {
	return parseStatementWithModel(ctx, pdfBytes, p.repo)
}

// BuildCategoriesPrompt delegates to the existing buildCategoriesPrompt function.
func (p *GeminiAIParser) BuildCategoriesPrompt(ctx context.Context) (string, error) {
	return buildCategoriesPromptWithRepo(ctx, p.repo)
}

// DocumentRepository is an interface for document-related database operations.
// This is used by the AIParser to avoid circular dependencies.
type DocumentRepository interface {
	ListActiveCategories(ctx context.Context) ([]infra.CategoryRow, error)
}

