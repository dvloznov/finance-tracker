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
type AIParser interface {
	// ParseStatement sends PDF bytes to an AI model and returns parsed JSON output.
	ParseStatement(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error)

	// ExtractAccountHeader sends PDF bytes to an AI model to extract account metadata from the header.
	ExtractAccountHeader(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error)
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

// ExtractAccountHeader calls the AI model to extract account metadata from the statement header.
func (p *GeminiAIParser) ExtractAccountHeader(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error) {
	return extractAccountHeaderWithModel(ctx, pdfBytes)
}
