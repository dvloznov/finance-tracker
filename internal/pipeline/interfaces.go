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
}

// GeminiAIParser is the concrete implementation of AIParser that uses Gemini AI.
type GeminiAIParser struct {
	repo CategoryRepository
}

// NewGeminiAIParser creates a new instance of GeminiAIParser.
func NewGeminiAIParser(repo CategoryRepository) *GeminiAIParser {
	return &GeminiAIParser{
		repo: repo,
	}
}

// ParseStatement delegates to the existing parseStatementWithModel function.
func (p *GeminiAIParser) ParseStatement(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error) {
	return parseStatementWithModel(ctx, pdfBytes, p.repo)
}

// CategoryRepository is an interface for category-related database operations.
// This is a minimal interface used by the AIParser to avoid circular dependencies.
// For full document repository operations, see infra.DocumentRepository.
type CategoryRepository interface {
	ListActiveCategories(ctx context.Context) ([]infra.CategoryRow, error)
}

