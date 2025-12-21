package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dvloznov/finance-tracker/internal/jobs"
	"github.com/dvloznov/finance-tracker/internal/jobs/inmemory"
	"github.com/dvloznov/finance-tracker/internal/logger"
	"github.com/dvloznov/finance-tracker/internal/pipeline"
)

func main() {
	// Initialize logger
	log := logger.New()

	// Initialize job store and queue
	// In production, this would be replaced with Cloud Tasks or Pub/Sub
	jobStore := inmemory.NewStore()
	jobQueue := inmemory.NewQueue(100, jobStore)

	log.Info().Msg("Starting worker service")

	// Create context that cancels on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create job handler that processes parse jobs
	handler := func(ctx context.Context, job jobs.Job) error {
		parseJob, ok := job.(*jobs.ParseDocumentJob)
		if !ok {
			return fmt.Errorf("unexpected job type: %T", job)
		}

		log.Info().
			Str("job_id", parseJob.JobID).
			Str("document_id", parseJob.DocumentID).
			Str("gcs_uri", parseJob.GCSURI).
			Msg("Processing parse job")

		// Execute the pipeline
		err := pipeline.IngestStatementFromGCS(ctx, parseJob.GCSURI)
		if err != nil {
			log.Error().
				Err(err).
				Str("job_id", parseJob.JobID).
				Str("document_id", parseJob.DocumentID).
				Msg("Pipeline execution failed")
			return err
		}

		log.Info().
			Str("job_id", parseJob.JobID).
			Str("document_id", parseJob.DocumentID).
			Msg("Pipeline execution completed successfully")

		return nil
	}

	// Start consuming jobs
	if err := jobQueue.Start(ctx, handler); err != nil {
		log.Fatal().Err(err).Msg("Failed to start job consumer")
	}

	log.Info().Msg("Worker service started, waiting for jobs...")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down worker service...")

	// Cancel context to stop workers
	cancel()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop the queue and wait for in-flight jobs
	if err := jobQueue.Stop(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Error during graceful shutdown")
	}

	// Close the queue
	if err := jobQueue.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close job queue")
	}

	log.Info().Msg("Worker service exited")
}
