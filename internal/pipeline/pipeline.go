package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/genai"
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

type CategoryRow struct {
	CategoryID       string              `bigquery:"category_id"`
	ParentCategoryID bigquery.NullString `bigquery:"parent_category_id"`
	Depth            int64               `bigquery:"depth"`
	Name             string              `bigquery:"name"`
	IsActive         bool                `bigquery:"is_active"`
}

type ModelOutputRow struct {
	OutputID     string `bigquery:"output_id"`
	ParsingRunID string `bigquery:"parsing_run_id"`
	DocumentID   string `bigquery:"document_id"`

	ModelName    string              `bigquery:"model_name"`
	ModelVersion bigquery.NullString `bigquery:"model_version"`

	RawJSON bigquery.NullJSON `bigquery:"raw_json"`

	ExtractedText bigquery.NullString `bigquery:"extracted_text"`
	Notes         bigquery.NullString `bigquery:"notes"`

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
	// gcsURI example: gs://my-bucket/path/to/file.pdf
	if !strings.HasPrefix(gcsURI, "gs://") {
		return nil, fmt.Errorf("invalid GCS URI: %s", gcsURI)
	}

	trimmed := strings.TrimPrefix(gcsURI, "gs://")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid GCS URI (no object path): %s", gcsURI)
	}

	bucketName := parts[0]
	objectPath := parts[1]

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetchFromGCS: creating storage client: %w", err)
	}
	defer storageClient.Close()

	rc, err := storageClient.Bucket(bucketName).Object(objectPath).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetchFromGCS: reading object %s/%s: %w", bucketName, objectPath, err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("fetchFromGCS: reading bytes: %w", err)
	}

	return data, nil
}

func buildCategoriesPrompt(ctx context.Context) (string, error) {
	client, err := bigquery.NewClient(ctx, "studious-union-470122-v7")
	if err != nil {
		return "", fmt.Errorf("buildCategoriesPrompt: bigquery client: %w", err)
	}
	defer client.Close()

	q := client.Query(`
		SELECT
		  category_id,
		  parent_category_id,
		  depth,
		  name,
		  is_active
		FROM finance.categories
		WHERE is_active = TRUE
		ORDER BY depth, parent_category_id, name
	`)

	it, err := q.Read(ctx)
	if err != nil {
		return "", fmt.Errorf("buildCategoriesPrompt: query read: %w", err)
	}

	var rows []CategoryRow

	for {
		var r CategoryRow
		err := it.Next(&r)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return "", fmt.Errorf("buildCategoriesPrompt: iter next: %w", err)
		}
		rows = append(rows, r)
	}

	if len(rows) == 0 {
		return "", fmt.Errorf("buildCategoriesPrompt: no active categories found")
	}

	// Separate parents and children.
	type parentInfo struct {
		ID   string
		Name string
	}
	var parentsOrder []parentInfo
	parentNameByID := make(map[string]string)
	childrenByParent := make(map[string][]string)

	for _, r := range rows {
		if r.Depth == 1 {
			parentsOrder = append(parentsOrder, parentInfo{ID: r.CategoryID, Name: r.Name})
			parentNameByID[r.CategoryID] = r.Name
			if _, ok := childrenByParent[r.CategoryID]; !ok {
				childrenByParent[r.CategoryID] = []string{}
			}
		}
	}

	for _, r := range rows {
		if r.Depth == 2 && r.ParentCategoryID.Valid {
			parentID := r.ParentCategoryID.StringVal
			childrenByParent[parentID] = append(childrenByParent[parentID], r.Name)
		}
	}

	var b strings.Builder
	b.WriteString("Use ONLY the following Categories and Subcategories:\n\n")

	for _, p := range parentsOrder {
		b.WriteString(p.Name + ":\n")
		subs := childrenByParent[p.ID]
		if len(subs) == 0 {
			// no subcategories defined – still list a placeholder so the model knows.
			b.WriteString("  - Other\n\n")
			continue
		}
		for _, s := range subs {
			b.WriteString("  - " + s + "\n")
		}
		b.WriteString("\n")
	}

	// Additionally, constrain what the model is allowed to output.
	b.WriteString("Category must be exactly one of the category names shown above.\n")
	b.WriteString("Subcategory must be exactly one of the subcategory names listed under that category.\n")
	b.WriteString("If you are unsure, default to category \"OTHER\" with subcategory \"Other\" if it exists.\n")

	return b.String(), nil
}

const modelName = "gemini-2.5-flash"

// parseStatementWithModel sends the PDF to Gemini and returns the parsed JSON output.
// It expects the model to return a STRICT JSON array of transactions.
func parseStatementWithModel(ctx context.Context, pdfBytes []byte) (map[string]interface{}, error) {
	// 1) Build category prompt from BigQuery taxonomy.
	catPrompt, err := buildCategoriesPrompt(ctx)
	if err != nil {
		return nil, fmt.Errorf("parseStatementWithModel: loading categories: %w", err)
	}

	// 2) Base instructions (very close to your test code).
	basePrompt :=
		"You are a financial statement parser for Barclays UK PDF bank statements.\n\n" +
			"Task:\n" +
			"- Parse ALL transactions in the attached Barclays statement.\n" +
			"- Output STRICT JSON only (no comments, no trailing commas, no extra text).\n" +
			"- Output a JSON array of objects.\n\n" +
			"Each object must have these fields:\n" +
			"- \"account_name\": string or null\n" +
			"- \"account_number\": string or null\n" +
			"- \"date\": string, ISO format \"YYYY-MM-DD\"\n" +
			"- \"description\": string\n" +
			"- \"amount\": number (positive for money IN, negative for money OUT)\n" +
			"- \"currency\": string (e.g. \"GBP\")\n" +
			"- \"balance_after\": number or null\n" +
			"- \"category\": string (one of the predefined categories)\n" +
			"- \"subcategory\": string (one of the predefined subcategories below)\n\n"

	rulesPrompt :=
		"Rules:\n" +
			"- Classify each transaction into the most appropriate category/subcategory.\n" +
			"- If the statement has separate \"paid out\" / \"paid in\" columns, convert to a single signed \"amount\".\n" +
			"- If the running balance is missing, set \"balance_after\" to null.\n" +
			"- If account name or number cannot be determined, set them to null.\n" +
			"- If the PDF contains multiple accounts, attribute transactions correctly.\n\n" +
			"Return ONLY valid raw JSON.\n" +
			"Do NOT wrap the response in code fences.\n" +
			"Do NOT use ```json or any Markdown.\n" +
			"Output must begin with \"[\" and end with \"]\".\n"

	fullPrompt := basePrompt + "\n" + catPrompt + "\n\n" + rulesPrompt

	// 3) Create GenAI client (same style as your test program).
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		HTTPOptions: genai.HTTPOptions{APIVersion: "v1"},
	})
	if err != nil {
		return nil, fmt.Errorf("parseStatementWithModel: create genai client: %w", err)
	}

	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: fullPrompt},
				{
					InlineData: &genai.Blob{
						MIMEType: "application/pdf",
						Data:     pdfBytes,
					},
				},
			},
		},
	}

	resp, err := client.Models.GenerateContent(ctx, modelName, contents, nil)
	if err != nil {
		return nil, fmt.Errorf("parseStatementWithModel: generate content: %w", err)
	}

	rawText := resp.Text()
	if rawText == "" {
		return nil, fmt.Errorf("parseStatementWithModel: empty response from model")
	}

	// Clean up Markdown fences / extra text if the model ignored instructions.
	clean := cleanModelJSON(rawText)

	// 4) Parse JSON into a generic value.
	var parsed interface{}
	if err := json.Unmarshal([]byte(clean), &parsed); err != nil {
		return nil, fmt.Errorf("parseStatementWithModel: unmarshal JSON: %w\nraw response: %s", err, rawText)
	}

	// Expect top-level array; for flexibility we just wrap it under "transactions".
	return map[string]interface{}{
		"transactions": parsed,
	}, nil
}

func cleanModelJSON(raw string) string {
	s := strings.TrimSpace(raw)

	// Handle ```json ... ``` or ``` ... ``` wrappers.
	if strings.HasPrefix(s, "```") {
		// Drop the first line (``` or ```json).
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		} else {
			// Single-line weirdness; just return as-is.
			return s
		}
		s = strings.TrimSpace(s)
	}

	// Remove trailing ``` if present.
	if idx := strings.LastIndex(s, "```"); idx != -1 {
		s = s[:idx]
	}

	s = strings.TrimSpace(s)

	// Extra safety: if there's still junk around the JSON array,
	// try to keep only from the first '[' to the last ']'.
	if start := strings.Index(s, "["); start != -1 {
		if end := strings.LastIndex(s, "]"); end != -1 && end > start {
			s = s[start : end+1]
			s = strings.TrimSpace(s)
		}
	}

	return s
}

// storeModelOutput inserts raw model output into the model_outputs table.
func storeModelOutput(
	ctx context.Context,
	parsingRunID string,
	documentID string,
	rawOutput map[string]interface{},
) (string, error) {
	const (
		projectID = "studious-union-470122-v7"
		datasetID = "finance"
		tableID   = "model_outputs"
	)

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("storeModelOutput: bigquery client: %w", err)
	}
	defer client.Close()

	outputID := uuid.NewString()

	jsonBytes, err := json.Marshal(rawOutput)
	if err != nil {
		return "", fmt.Errorf("storeModelOutput: marshal rawOutput: %w", err)
	}

	row := &ModelOutputRow{
		OutputID:     outputID,
		ParsingRunID: parsingRunID,
		DocumentID:   documentID,

		ModelName: "gemini-2.5-flash",
		ModelVersion: bigquery.NullString{
			Valid: false,
		},

		RawJSON: bigquery.NullJSON{
			JSONVal: string(jsonBytes), // <<<< correct
			Valid:   true,
		},

		ExtractedText: bigquery.NullString{Valid: false},
		Notes:         bigquery.NullString{Valid: false},

		Metadata: bigquery.NullJSON{
			Valid: false,
		},
	}

	inserter := client.Dataset(datasetID).Table(tableID).Inserter()
	if err := inserter.Put(ctx, row); err != nil {
		return "", fmt.Errorf("storeModelOutput: inserting row: %w", err)
	}

	return outputID, nil
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
