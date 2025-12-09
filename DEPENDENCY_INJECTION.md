# Dependency Injection and Testing Guide

This document describes the interfaces introduced for dependency injection and how to use them for testing.

## Overview

Three main interfaces have been extracted to enable mocking and testing:

1. **DocumentRepository** - BigQuery document operations
2. **StorageService** - GCS storage operations  
3. **AIParser** - AI-powered document parsing

## Interfaces

### DocumentRepository

Located in `internal/infra/bigquery/interfaces.go`

```go
type DocumentRepository interface {
    InsertDocument(ctx context.Context, row *DocumentRow) error
    InsertTransactions(ctx context.Context, rows []*TransactionRow) error
    InsertModelOutput(ctx context.Context, row *ModelOutputRow) error
    StartParsingRun(ctx context.Context, documentID string) (string, error)
    MarkParsingRunFailed(ctx context.Context, parsingRunID string, parseErr error)
    MarkParsingRunSucceeded(ctx context.Context, parsingRunID string) error
    ListActiveCategories(ctx context.Context) ([]CategoryRow, error)
}
```

**Concrete Implementation:** `BigQueryDocumentRepository`

### StorageService

Located in `internal/gcsuploader/interfaces.go`

```go
type StorageService interface {
    UploadFile(ctx context.Context, bucketName, objectName, filePath string) error
    FetchFromGCS(ctx context.Context, gcsURI string) ([]byte, error)
    ExtractFilenameFromGCSURI(uri string) string
}
```

**Concrete Implementation:** `GCSStorageService`

### AIParser

Located in `internal/pipeline/interfaces.go`

```go
type AIParser interface {
    ParseStatement(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error)
}
```

**Concrete Implementation:** `GeminiAIParser`

## Usage

### Production Code

The existing `IngestStatementFromGCS` function remains unchanged for production use:

```go
err := pipeline.IngestStatementFromGCS(ctx, "gs://bucket/file.pdf")
```

### Testing with Dependency Injection

Use `IngestStatementFromGCSWithDeps` to inject mock implementations:

```go
// Create mocks
mockRepo := &MockDocumentRepository{...}
mockStorage := &MockStorageService{...}
mockParser := &MockAIParser{...}

// Call with dependencies
err := pipeline.IngestStatementFromGCSWithDeps(
    ctx,
    "gs://bucket/file.pdf",
    mockRepo,
    mockStorage,
    mockParser,
)
```

### Example Mock Implementation

See `internal/pipeline/example_test.go` for complete mock implementations:

```go
type MockStorageService struct {
    FetchFromGCSFunc func(ctx context.Context, gcsURI string) ([]byte, error)
}

func (m *MockStorageService) FetchFromGCS(ctx context.Context, gcsURI string) ([]byte, error) {
    if m.FetchFromGCSFunc != nil {
        return m.FetchFromGCSFunc(ctx, gcsURI)
    }
    return []byte("mock data"), nil
}
```

## Benefits

1. **Testability** - Easy to create unit tests without external dependencies
2. **Mocking** - Simple to mock behavior for different test scenarios
3. **Loose Coupling** - Components depend on interfaces, not concrete implementations
4. **Backward Compatibility** - Existing code continues to work without changes

## Pipeline State

The `PipelineState` struct now holds the injected dependencies:

```go
type PipelineState struct {
    // ... existing fields
    
    // Injected dependencies
    DocumentRepo   infra.DocumentRepository
    StorageService StorageService
    AIParser       AIParser
}
```

Each pipeline step uses these dependencies instead of calling concrete implementations directly.
