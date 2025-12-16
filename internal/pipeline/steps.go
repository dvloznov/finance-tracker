package pipeline

import (
	"context"
	"crypto/sha256"
	"fmt"

	infra "github.com/dvloznov/finance-tracker/internal/infra/bigquery"
)

// PipelineStep represents a single step in the ingestion pipeline.
type PipelineStep interface {
	Execute(ctx context.Context, state *PipelineState) error
	Name() string
}

// PipelineState holds the shared state across all pipeline steps.
type PipelineState struct {
	GCSURI         string
	DocumentID     string
	ParsingRunID   string
	PDFBytes       []byte
	Checksum       string // SHA-256 checksum of the PDF file
	RawModelOutput map[string]interface{}
	Transactions   []*Transaction
	IsReparse      bool // True if we're re-parsing an existing document

	// Injected dependencies
	DocumentRepo      infra.DocumentRepository
	StorageService    StorageService
	AIParser          AIParser
	CategoryValidator *CategoryValidator
}

// Step 1: CreateDocumentStep creates a document record for the file.
type CreateDocumentStep struct{}

func (s *CreateDocumentStep) Name() string {
	return "CreateDocument"
}

func (s *CreateDocumentStep) Execute(ctx context.Context, state *PipelineState) error {
	// Check if a document with this checksum already exists
	if state.Checksum != "" {
		existingDoc, err := state.DocumentRepo.FindDocumentByChecksum(ctx, state.Checksum)
		if err != nil {
			return fmt.Errorf("CreateDocument: checking for duplicate: %w", err)
		}

		if existingDoc != nil {
			// Document already exists - reuse it
			state.DocumentID = existingDoc.DocumentID
			state.IsReparse = true
			return nil
		}
	}

	// No duplicate found - create new document with checksum
	documentID, err := createDocumentWithChecksumRepo(ctx, state.GCSURI, state.Checksum, state.DocumentRepo, state.StorageService)
	if err != nil {
		return err
	}
	state.DocumentID = documentID
	state.IsReparse = false
	return nil
}

// Step 1a: SupersedeOldParsingRunsStep marks old parsing runs as SUPERSEDED if re-parsing.
type SupersedeOldParsingRunsStep struct{}

func (s *SupersedeOldParsingRunsStep) Name() string {
	return "SupersedeOldParsingRuns"
}

func (s *SupersedeOldParsingRunsStep) Execute(ctx context.Context, state *PipelineState) error {
	// Only supersede if this is a re-parse (document already existed)
	if !state.IsReparse {
		return nil
	}

	if err := state.DocumentRepo.MarkParsingRunsAsSuperseded(ctx, state.DocumentID); err != nil {
		return fmt.Errorf("SupersedeOldParsingRuns: %w", err)
	}
	return nil
}

// Step 2: StartParsingRunStep starts a parsing run (status=RUNNING).
type StartParsingRunStep struct{}

func (s *StartParsingRunStep) Name() string {
	return "StartParsingRun"
}

func (s *StartParsingRunStep) Execute(ctx context.Context, state *PipelineState) error {
	parsingRunID, err := state.DocumentRepo.StartParsingRun(ctx, state.DocumentID)
	if err != nil {
		return err
	}
	state.ParsingRunID = parsingRunID
	return nil
}

// Step 3: FetchPDFStep fetches the PDF bytes from GCS.
type FetchPDFStep struct{}

func (s *FetchPDFStep) Name() string {
	return "FetchPDF"
}

func (s *FetchPDFStep) Execute(ctx context.Context, state *PipelineState) error {
	pdfBytes, err := state.StorageService.FetchFromGCS(ctx, state.GCSURI)
	if err != nil {
		// Only mark parsing run as failed if it exists
		if state.ParsingRunID != "" {
			state.DocumentRepo.MarkParsingRunFailed(ctx, state.ParsingRunID, err)
		}
		return err
	}
	state.PDFBytes = pdfBytes
	return nil
}

// Step 3a: CalculateChecksumStep calculates the SHA-256 checksum of the PDF.
type CalculateChecksumStep struct{}

func (s *CalculateChecksumStep) Name() string {
	return "CalculateChecksum"
}

func (s *CalculateChecksumStep) Execute(ctx context.Context, state *PipelineState) error {
	if len(state.PDFBytes) == 0 {
		return fmt.Errorf("CalculateChecksum: PDF bytes not available")
	}
	// Calculate SHA-256 hash
	hash := sha256.Sum256(state.PDFBytes)
	state.Checksum = fmt.Sprintf("%x", hash[:])
	return nil
}

// Step 4: ParseStatementStep calls the statement parser (Gemini) with the PDF.
type ParseStatementStep struct{}

func (s *ParseStatementStep) Name() string {
	return "ParseStatement"
}

func (s *ParseStatementStep) Execute(ctx context.Context, state *PipelineState) error {
	rawModelOutput, err := state.AIParser.ParseStatement(ctx, state.PDFBytes)
	if err != nil {
		state.DocumentRepo.MarkParsingRunFailed(ctx, state.ParsingRunID, err)
		return err
	}
	state.RawModelOutput = rawModelOutput
	return nil
}

// Step 5: StoreModelOutputStep stores raw model output in model_outputs.
type StoreModelOutputStep struct{}

func (s *StoreModelOutputStep) Name() string {
	return "StoreModelOutput"
}

func (s *StoreModelOutputStep) Execute(ctx context.Context, state *PipelineState) error {
	_, err := storeModelOutputWithRepo(ctx, state.ParsingRunID, state.DocumentID, state.RawModelOutput, state.DocumentRepo)
	if err != nil {
		state.DocumentRepo.MarkParsingRunFailed(ctx, state.ParsingRunID, err)
		return err
	}
	return nil
}

// Step 6: TransformTransactionsStep transforms raw model output into normalized transactions.
type TransformTransactionsStep struct{}

func (s *TransformTransactionsStep) Name() string {
	return "TransformTransactions"
}

func (s *TransformTransactionsStep) Execute(ctx context.Context, state *PipelineState) error {
	txs, err := transformModelOutputToTransactions(state.RawModelOutput)
	if err != nil {
		state.DocumentRepo.MarkParsingRunFailed(ctx, state.ParsingRunID, err)
		return err
	}
	state.Transactions = txs
	return nil
}

// Step 6a: CreateCategoryValidatorStep creates a category validator from the taxonomy.
type CreateCategoryValidatorStep struct{}

func (s *CreateCategoryValidatorStep) Name() string {
	return "CreateCategoryValidator"
}

func (s *CreateCategoryValidatorStep) Execute(ctx context.Context, state *PipelineState) error {
	validator, err := NewCategoryValidator(ctx, state.DocumentRepo)
	if err != nil {
		return fmt.Errorf("CreateCategoryValidator: %w", err)
	}
	state.CategoryValidator = validator
	return nil
}

// Step 6b: ValidateCategoriesStep validates all transaction categories against the taxonomy.
type ValidateCategoriesStep struct{}

func (s *ValidateCategoriesStep) Name() string {
	return "ValidateCategories"
}

func (s *ValidateCategoriesStep) Execute(ctx context.Context, state *PipelineState) error {
	if state.CategoryValidator == nil {
		return fmt.Errorf("ValidateCategories: category validator not initialized")
	}

	var validationErrors []string
	for i, tx := range state.Transactions {
		categoryID, err := state.CategoryValidator.ValidateCategory(tx.Category, tx.Subcategory)
		if err != nil {
			validationErrors = append(validationErrors,
				fmt.Sprintf("transaction %d (date: %s, desc: %s): %v",
					i, tx.Date.Format("2006-01-02"), tx.Description, err))
		} else {
			// Store the validated category_id back in the transaction
			tx.CategoryID = categoryID
		}
	}

	if len(validationErrors) > 0 {
		err := fmt.Errorf("category validation failed:\n  - %s",
			fmt.Sprintf("%v", validationErrors))
		state.DocumentRepo.MarkParsingRunFailed(ctx, state.ParsingRunID, err)
		return err
	}

	return nil
}

// Step 7: InsertTransactionsStep inserts transactions into the transactions table.
type InsertTransactionsStep struct{}

func (s *InsertTransactionsStep) Name() string {
	return "InsertTransactions"
}

func (s *InsertTransactionsStep) Execute(ctx context.Context, state *PipelineState) error {
	if err := insertTransactionsWithRepo(ctx, state.DocumentID, state.ParsingRunID, state.Transactions, state.DocumentRepo); err != nil {
		state.DocumentRepo.MarkParsingRunFailed(ctx, state.ParsingRunID, err)
		return err
	}
	return nil
}

// Step 8: MarkSuccessStep marks the parsing run as SUCCESS.
type MarkSuccessStep struct{}

func (s *MarkSuccessStep) Name() string {
	return "MarkSuccess"
}

func (s *MarkSuccessStep) Execute(ctx context.Context, state *PipelineState) error {
	if err := state.DocumentRepo.MarkParsingRunSucceeded(ctx, state.ParsingRunID); err != nil {
		return err
	}
	return nil
}

// Pipeline executes a sequence of steps in order.
type Pipeline struct {
	steps []PipelineStep
}

// NewPipeline creates a new pipeline with the given steps.
func NewPipeline(steps ...PipelineStep) *Pipeline {
	return &Pipeline{steps: steps}
}

// Execute runs all steps in the pipeline sequentially.
func (p *Pipeline) Execute(ctx context.Context, state *PipelineState) error {
	for i, step := range p.steps {
		if err := step.Execute(ctx, state); err != nil {
			return fmt.Errorf("pipeline step %d (%s) failed: %w", i+1, step.Name(), err)
		}
	}
	return nil
}

// NewStatementIngestionPipeline creates the standard pipeline for ingesting statements.
func NewStatementIngestionPipeline() *Pipeline {
	return NewPipeline(
		&FetchPDFStep{},
		&CalculateChecksumStep{},
		&CreateDocumentStep{},
		&SupersedeOldParsingRunsStep{},
		&StartParsingRunStep{},
		&ParseStatementStep{},
		&StoreModelOutputStep{},
		&TransformTransactionsStep{},
		&CreateCategoryValidatorStep{},
		&ValidateCategoriesStep{},
		&InsertTransactionsStep{},
		&MarkSuccessStep{},
	)
}
