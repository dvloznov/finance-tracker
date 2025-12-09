package pipeline

import (
	"context"
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
	RawModelOutput map[string]interface{}
	Transactions   []*Transaction

	// Injected dependencies
	DocumentRepo   infra.DocumentRepository
	StorageService StorageService
	AIParser       AIParser
}

// Step 1: CreateDocumentStep creates a document record for the file.
type CreateDocumentStep struct{}

func (s *CreateDocumentStep) Name() string {
	return "CreateDocument"
}

func (s *CreateDocumentStep) Execute(ctx context.Context, state *PipelineState) error {
	documentID, err := createDocumentWithRepo(ctx, state.GCSURI, state.DocumentRepo, state.StorageService)
	if err != nil {
		return err
	}
	state.DocumentID = documentID
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
		state.DocumentRepo.MarkParsingRunFailed(ctx, state.ParsingRunID, err)
		return err
	}
	state.PDFBytes = pdfBytes
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

// NewStatementIngestionPipeline creates the standard 8-step pipeline for ingesting statements.
func NewStatementIngestionPipeline() *Pipeline {
	return NewPipeline(
		&CreateDocumentStep{},
		&StartParsingRunStep{},
		&FetchPDFStep{},
		&ParseStatementStep{},
		&StoreModelOutputStep{},
		&TransformTransactionsStep{},
		&InsertTransactionsStep{},
		&MarkSuccessStep{},
	)
}
