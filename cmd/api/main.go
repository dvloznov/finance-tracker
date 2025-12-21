package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dvloznov/finance-tracker/internal/api/handlers"
	"github.com/dvloznov/finance-tracker/internal/api/middleware"
	infraBQ "github.com/dvloznov/finance-tracker/internal/infra/bigquery"
	"github.com/dvloznov/finance-tracker/internal/jobs"
	"github.com/dvloznov/finance-tracker/internal/jobs/inmemory"
	"github.com/dvloznov/finance-tracker/internal/logger"
	"github.com/dvloznov/finance-tracker/internal/pipeline"
)

func main() {
	// Parse command-line flags
	var (
		port   = flag.String("port", "8080", "HTTP server port")
		bucket = flag.String("bucket", os.Getenv("GCS_BUCKET"), "GCS bucket name for document uploads (or set GCS_BUCKET env)")
	)
	flag.Parse()

	// Initialize logger
	log := logger.New()

	if *bucket == "" {
		log.Warn().Msg("No GCS bucket configured - document uploads will be disabled")
	}

	// Initialize repositories
	ctx := context.Background()

	docRepo, err := infraBQ.NewBigQueryDocumentRepository(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create document repository")
	}
	defer docRepo.Close()

	// Initialize job infrastructure
	jobStore := inmemory.NewStore()
	jobQueue := inmemory.NewQueue(100, jobStore)

	// Start worker in background to process jobs
	workerCtx, cancelWorker := context.WithCancel(ctx)
	defer cancelWorker()

	// Create job handler for processing parse jobs
	jobHandler := func(ctx context.Context, job jobs.Job) error {
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

	// Start job consumer in background
	go func() {
		log.Info().Msg("Starting job worker")
		if err := jobQueue.Start(workerCtx, jobHandler); err != nil {
			log.Error().Err(err).Msg("Job worker stopped with error")
		}
	}()

	// Initialize handlers
	documentsHandler := handlers.NewDocumentsHandler(docRepo, jobQueue, *bucket, log)
	transactionsHandler := handlers.NewTransactionsHandler(docRepo, log)
	categoriesHandler := handlers.NewCategoriesHandler(docRepo, log)
	jobsHandler := handlers.NewJobsHandler(jobStore, log)

	// Create router
	mux := http.NewServeMux()

	// Documents endpoints
	mux.HandleFunc("/api/documents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			documentsHandler.ListDocuments(w, r)
		} else {
			middleware.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/api/documents/upload-url", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			documentsHandler.CreateUploadURL(w, r)
		} else {
			middleware.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/api/documents/upload/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			// Extract document ID from path
			documentID := strings.TrimPrefix(r.URL.Path, "/api/documents/upload/")
			if documentID == "" {
				middleware.WriteError(w, http.StatusBadRequest, "Document ID is required")
				return
			}
			documentsHandler.UploadDocument(w, r, documentID)
		} else {
			middleware.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/api/documents/parse", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			documentsHandler.EnqueueParsing(w, r)
		} else {
			middleware.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Transactions endpoints
	mux.HandleFunc("/api/transactions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			transactionsHandler.ListTransactions(w, r)
		} else {
			middleware.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Categories endpoints
	mux.HandleFunc("/api/categories", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			categoriesHandler.ListCategories(w, r)
		} else {
			middleware.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Jobs endpoints
	mux.HandleFunc("/api/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			jobsHandler.ListJobs(w, r)
		} else {
			middleware.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/api/jobs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// Extract job ID from path
			jobID := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
			if jobID == "" {
				middleware.WriteError(w, http.StatusBadRequest, "Job ID is required")
				return
			}
			jobsHandler.GetJob(w, r, jobID)
		} else {
			middleware.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		middleware.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Apply middleware
	handler := middleware.Recovery(log)(
		middleware.Logger(log)(
			middleware.RequestID(
				middleware.CORS(
					middleware.Auth(mux),
				),
			),
		),
	)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info().Str("port", *port).Msg("Starting API server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Cancel worker context
	cancelWorker()

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	// Stop job queue and wait for in-flight jobs
	if err := jobQueue.Stop(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Error stopping job queue")
	}

	// Close job queue
	if err := jobQueue.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close job queue")
	}

	log.Info().Msg("Server exited")
}
