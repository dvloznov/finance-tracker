package pipeline

import (
	"context"
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
)

type DocumentRow struct {
	DocumentID       string                 `bigquery:"document_id"`
	UserID           string                 `bigquery:"user_id"`
	GCSURI           string                 `bigquery:"gcs_uri"`
	DocumentType     string                 `bigquery:"document_type"`
	SourceSystem     string                 `bigquery:"source_system"`
	InstitutionID    string                 `bigquery:"institution_id"`
	AccountID        string                 `bigquery:"account_id"`
	StatementStart   bigquery.NullDate      `bigquery:"statement_start_date"`
	StatementEnd     bigquery.NullDate      `bigquery:"statement_end_date"`
	UploadTS         time.Time              `bigquery:"upload_ts"`
	ProcessedTS      bigquery.NullTimestamp `bigquery:"processed_ts"`
	ParsingStatus    string                 `bigquery:"parsing_status"`
	OriginalFilename string                 `bigquery:"original_filename"`
	FileMimeType     string                 `bigquery:"file_mime_type"`
	TextGCSURI       string                 `bigquery:"text_gcs_uri"`
	ChecksumSHA256   string                 `bigquery:"checksum_sha256"`
	Metadata         bigquery.NullJSON      `bigquery:"metadata"`
}

type ParsingRunRow struct {
	ParsingRunID string `bigquery:"parsing_run_id"`
	DocumentID   string `bigquery:"document_id"`

	StartedAt  time.Time              `bigquery:"started_ts"`  // note: _ts
	FinishedAt bigquery.NullTimestamp `bigquery:"finished_ts"` // note: _ts

	ParserType    string `bigquery:"parser_type"`    // e.g. GEMINI_VISION
	ParserVersion string `bigquery:"parser_version"` // e.g. v1

	Status       string `bigquery:"status"`
	ErrorMessage string `bigquery:"error_message"`

	// Optional metrics (can be NULL)
	TokensInput  bigquery.NullInt64 `bigquery:"tokens_input"`
	TokensOutput bigquery.NullInt64 `bigquery:"tokens_output"`

	Metadata bigquery.NullJSON `bigquery:"metadata"`
}

// IngestStatementFromGCS processes a single bank statement PDF stored in GCS.
// gcsURI should look like: "gs://bucket/path/to/statement.pdf".
func IngestStatementFromGCS(ctx context.Context, gcsURI string) error {
	// 1. Create a document record for this file.
	documentID, err := createDocument(ctx, gcsURI)
	if err != nil {
		return err
	}

	// 2. Start a parsing run (status=RUNNING).
	parsingRunID, err := startParsingRun(ctx, documentID)
	if err != nil {
		return err
	}

	// 3. Fetch the PDF bytes from GCS.
	pdfBytes, err := fetchFromGCS(ctx, gcsURI)
	if err != nil {
		markParsingRunFailed(ctx, parsingRunID, err)
		return err
	}

	// 4. Call the statement parser (Gemini) with the PDF.
	rawModelOutput, err := parseStatementWithModel(ctx, pdfBytes)
	if err != nil {
		markParsingRunFailed(ctx, parsingRunID, err)
		return err
	}

	// 5. Store raw model output in model_outputs.
	_, err = storeModelOutput(ctx, parsingRunID, documentID, rawModelOutput)
	if err != nil {
		markParsingRunFailed(ctx, parsingRunID, err)
		return err
	}

	// 6. Transform raw model output into normalized transactions.
	txs, err := transformModelOutputToTransactions(rawModelOutput)
	if err != nil {
		markParsingRunFailed(ctx, parsingRunID, err)
		return err
	}

	// 7. Insert transactions into the transactions table.
	if err := insertTransactions(ctx, documentID, parsingRunID, txs); err != nil {
		markParsingRunFailed(ctx, parsingRunID, err)
		return err
	}

	// 8. Mark parsing run as SUCCESS.
	if err := markParsingRunSucceeded(ctx, parsingRunID); err != nil {
		return err
	}

	return nil
}

//
// ──────────────────────────────────────────────────────────────
//  Helper function skeletons (generic, no bank-specific naming)
// ──────────────────────────────────────────────────────────────
//

// createDocument inserts a row into the documents table for this file.
func createDocument(ctx context.Context, gcsURI string) (string, error) {
	// BigQuery client
	client, err := bigquery.NewClient(ctx, "studious-union-470122-v7")
	if err != nil {
		return "", fmt.Errorf("createDocument: bigquery client: %w", err)
	}
	defer client.Close()

	// Generate a UUID for this document
	documentID := uuid.NewString()

	// Extract filename from GCS URI
	// e.g. "gs://bucket/folder/file.pdf" → "file.pdf"
	filename := extractFilenameFromGCSURI(gcsURI)

	// Prepare row to insert
	row := &DocumentRow{
		DocumentID:       documentID,
		UserID:           "denis", // You can generalize this later
		GCSURI:           gcsURI,
		DocumentType:     "BANK_STATEMENT", // For now we assume this
		SourceSystem:     "BARCLAYS",       // Later: detect automatically
		InstitutionID:    "",               // Can be filled later
		AccountID:        "",               // Can be filled later
		ParsingStatus:    "PENDING",
		UploadTS:         time.Now(),
		OriginalFilename: filename,
		FileMimeType:     "",                              // Fill later if you detect MIME
		Metadata:         bigquery.NullJSON{Valid: false}, // NULL for now
	}

	inserter := client.Dataset("finance").Table("documents").Inserter()

	if err := inserter.Put(ctx, row); err != nil {
		return "", fmt.Errorf("createDocument: inserting row: %w", err)
	}

	return documentID, nil
}

func extractFilenameFromGCSURI(uri string) string {
	// Remove "gs://"
	trimmed := strings.TrimPrefix(uri, "gs://")

	// Remove bucket name
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) < 2 {
		return trimmed
	}

	// Extract actual filename
	return path.Base(parts[1])
}

// startParsingRun creates a parsing_runs row with status=RUNNING.
func startParsingRun(ctx context.Context, documentID string) (string, error) {
	const (
		projectID = "studious-union-470122-v7"
		datasetID = "finance"
		tableID   = "parsing_runs"
	)

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("startParsingRun: bigquery client: %w", err)
	}
	defer client.Close()

	parsingRunID := uuid.NewString()
	started := time.Now()

	q := client.Query(fmt.Sprintf(`
		INSERT %s.%s (
			parsing_run_id,
			document_id,
			started_ts,
			parser_type,
			parser_version,
			status
		)
		VALUES (
			@parsing_run_id,
			@document_id,
			@started_ts,
			@parser_type,
			@parser_version,
			@status
		)
	`, datasetID, tableID))

	q.Parameters = []bigquery.QueryParameter{
		{Name: "parsing_run_id", Value: parsingRunID},
		{Name: "document_id", Value: documentID},
		{Name: "started_ts", Value: started},
		{Name: "parser_type", Value: "GEMINI_VISION"},
		{Name: "parser_version", Value: "v1"},
		{Name: "status", Value: "RUNNING"},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return "", fmt.Errorf("startParsingRun: running insert query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("startParsingRun: waiting for job: %w", err)
	}
	if err := status.Err(); err != nil {
		return "", fmt.Errorf("startParsingRun: job error: %w", err)
	}

	return parsingRunID, nil
}

// markParsingRunFailed updates a parsing_runs row to status=FAILED.
func markParsingRunFailed(ctx context.Context, parsingRunID string, parseErr error) {
	const (
		projectID = "studious-union-470122-v7"
		datasetID = "finance"
	)

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("markParsingRunFailed: bigquery client error for run %s: %v", parsingRunID, err)
		return
	}
	defer client.Close()

	errMsg := ""
	if parseErr != nil {
		errMsg = parseErr.Error()
		const maxLen = 2000
		if len(errMsg) > maxLen {
			errMsg = errMsg[:maxLen]
		}
	}

	q := client.Query(fmt.Sprintf(`
		UPDATE %s.parsing_runs
		SET status = @status,
		    finished_ts = @finished_ts,
		    error_message = @error_message
		WHERE parsing_run_id = @parsing_run_id
	`, datasetID))

	q.Parameters = []bigquery.QueryParameter{
		{Name: "status", Value: "FAILED"},
		{Name: "finished_ts", Value: time.Now()},
		{Name: "error_message", Value: errMsg},
		{Name: "parsing_run_id", Value: parsingRunID},
	}

	job, err := q.Run(ctx)
	if err != nil {
		log.Printf("markParsingRunFailed: running update query for run %s: %v", parsingRunID, err)
		return
	}

	status, err := job.Wait(ctx)
	if err != nil {
		log.Printf("markParsingRunFailed: waiting for job for run %s: %v", parsingRunID, err)
		return
	}
	if err := status.Err(); err != nil {
		log.Printf("markParsingRunFailed: job completed with error for run %s: %v", parsingRunID, err)
	}
}

// markParsingRunSucceeded updates a parsing_runs row to status=SUCCESS.
func markParsingRunSucceeded(ctx context.Context, parsingRunID string) error {
	const (
		projectID = "studious-union-470122-v7"
		datasetID = "finance"
	)

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("markParsingRunSucceeded: bigquery client: %w", err)
	}
	defer client.Close()

	q := client.Query(fmt.Sprintf(`
		UPDATE %s.parsing_runs
		SET status = @status,
		    finished_ts = @finished_ts,
		    error_message = ""
		WHERE parsing_run_id = @parsing_run_id
	`, datasetID))

	q.Parameters = []bigquery.QueryParameter{
		{Name: "status", Value: "SUCCESS"},
		{Name: "finished_ts", Value: time.Now()},
		{Name: "parsing_run_id", Value: parsingRunID},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("markParsingRunSucceeded: running update query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("markParsingRunSucceeded: waiting for job: %w", err)
	}
	if err := status.Err(); err != nil {
		return fmt.Errorf("markParsingRunSucceeded: job error: %w", err)
	}

	return nil
}

// fetchFromGCS downloads the file bytes from the given GCS URI.
func fetchFromGCS(ctx context.Context, gcsURI string) ([]byte, error) {
	// TODO: download bytes from Google Cloud Storage
	return nil, nil
}

// parseStatementWithModel sends the PDF to the model (e.g. Gemini) and returns raw output.
func parseStatementWithModel(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error) {
	// TODO: call LLM / vision model and return raw parsed JSON as a generic map
	return nil, nil
}

// storeModelOutput inserts raw model output into the model_outputs table.
func storeModelOutput(
	ctx context.Context,
	parsingRunID string,
	documentID string,
	rawOutput map[string]interface{},
) (outputID string, err error) {
	// TODO: insert into finance.model_outputs and return output_id
	return "", nil
}

// transformModelOutputToTransactions converts raw model output into normalized transaction structs.
func transformModelOutputToTransactions(
	rawOutput map[string]interface{},
) ([]*Transaction, error) {
	// TODO: map model JSON → Transaction structs
	return nil, nil
}

// insertTransactions writes a batch of transactions to the transactions table.
func insertTransactions(
	ctx context.Context,
	documentID string,
	parsingRunID string,
	txs []*Transaction,
) error {
	// TODO: insert into finance.transactions (streaming or batch)
	return nil
}

// Transaction represents one normalized transaction ready to be stored.
type Transaction struct {
	// TODO: add fields: Date, Amount, Currency, Description, CategoryID, etc.
}
