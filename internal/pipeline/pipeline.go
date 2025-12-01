package pipeline

import (
	"context"
	"fmt"
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
func startParsingRun(ctx context.Context, documentID string) (parsingRunID string, err error) {
	// TODO: insert into finance.parsing_runs and return parsing_run_id
	return "", nil
}

// markParsingRunFailed updates a parsing_runs row to status=FAILED.
func markParsingRunFailed(ctx context.Context, parsingRunID string, parseErr error) {
	// TODO: update finance.parsing_runs set status='FAILED', error_message=...
}

// markParsingRunSucceeded updates a parsing_runs row to status=SUCCESS.
func markParsingRunSucceeded(ctx context.Context, parsingRunID string) error {
	// TODO: update finance.parsing_runs set status='SUCCESS'
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
