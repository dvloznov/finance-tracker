package bigquery

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/dvloznov/finance-tracker/internal/logger"
	"github.com/google/uuid"
)

const (
	projectID        = "studious-union-470122-v7"
	datasetID        = "finance"
	parsingRunsTable = "parsing_runs"
)

// StartParsingRun inserts a new row into finance.parsing_runs with status=RUNNING
// and returns the generated parsing_run_id.
func StartParsingRun(ctx context.Context, documentID string) (string, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("StartParsingRun: bigquery client: %w", err)
	}
	defer client.Close()

	return StartParsingRunWithClient(ctx, client, documentID)
}

// StartParsingRunWithClient inserts a new row into finance.parsing_runs with status=RUNNING
// and returns the generated parsing_run_id using the provided BigQuery client.
func StartParsingRunWithClient(ctx context.Context, client *bigquery.Client, documentID string) (string, error) {
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
	`, datasetID, parsingRunsTable))

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
		return "", fmt.Errorf("StartParsingRun: running insert query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("StartParsingRun: waiting for job: %w", err)
	}
	if err := status.Err(); err != nil {
		return "", fmt.Errorf("StartParsingRun: job error: %w", err)
	}

	return parsingRunID, nil
}

// MarkParsingRunFailed sets status=FAILED, finished_ts and error_message.
func MarkParsingRunFailed(ctx context.Context, parsingRunID string, parseErr error) {
	log := logger.FromContext(ctx)
	
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Error().
			Err(err).
			Str("parsing_run_id", parsingRunID).
			Msg("MarkParsingRunFailed: bigquery client error")
		return
	}
	defer client.Close()

	MarkParsingRunFailedWithClient(ctx, client, parsingRunID, parseErr)
}

// MarkParsingRunFailedWithClient sets status=FAILED, finished_ts and error_message
// using the provided BigQuery client.
func MarkParsingRunFailedWithClient(ctx context.Context, client *bigquery.Client, parsingRunID string, parseErr error) {
	log := logger.FromContext(ctx)
	
	errMsg := ""
	if parseErr != nil {
		errMsg = parseErr.Error()
		const maxLen = 2000
		if len(errMsg) > maxLen {
			errMsg = errMsg[:maxLen]
		}
	}

	q := client.Query(fmt.Sprintf(`
		UPDATE %s.%s
		SET status = @status,
		    finished_ts = @finished_ts,
		    error_message = @error_message
		WHERE parsing_run_id = @parsing_run_id
	`, datasetID, parsingRunsTable))

	q.Parameters = []bigquery.QueryParameter{
		{Name: "status", Value: "FAILED"},
		{Name: "finished_ts", Value: time.Now()},
		{Name: "error_message", Value: errMsg},
		{Name: "parsing_run_id", Value: parsingRunID},
	}

	job, err := q.Run(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Str("parsing_run_id", parsingRunID).
			Msg("MarkParsingRunFailed: running update query")
		return
	}

	status, err := job.Wait(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Str("parsing_run_id", parsingRunID).
			Msg("MarkParsingRunFailed: waiting for job")
		return
	}
	if err := status.Err(); err != nil {
		log.Error().
			Err(err).
			Str("parsing_run_id", parsingRunID).
			Msg("MarkParsingRunFailed: job completed with error")
	}
}

// MarkParsingRunSucceeded sets status=SUCCESS and finished_ts, clears error_message.
func MarkParsingRunSucceeded(ctx context.Context, parsingRunID string) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("MarkParsingRunSucceeded: bigquery client: %w", err)
	}
	defer client.Close()

	return MarkParsingRunSucceededWithClient(ctx, client, parsingRunID)
}

// MarkParsingRunSucceededWithClient sets status=SUCCESS and finished_ts, clears error_message
// using the provided BigQuery client.
func MarkParsingRunSucceededWithClient(ctx context.Context, client *bigquery.Client, parsingRunID string) error {
	q := client.Query(fmt.Sprintf(`
		UPDATE %s.%s
		SET status = @status,
		    finished_ts = @finished_ts,
		    error_message = ""
		WHERE parsing_run_id = @parsing_run_id
	`, datasetID, parsingRunsTable))

	q.Parameters = []bigquery.QueryParameter{
		{Name: "status", Value: "SUCCESS"},
		{Name: "finished_ts", Value: time.Now()},
		{Name: "parsing_run_id", Value: parsingRunID},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("MarkParsingRunSucceeded: running update query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("MarkParsingRunSucceeded: waiting for job: %w", err)
	}
	if err := status.Err(); err != nil {
		return fmt.Errorf("MarkParsingRunSucceeded: job error: %w", err)
	}

	return nil
}
