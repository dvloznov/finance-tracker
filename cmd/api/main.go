package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dvloznov/finance-tracker/internal/api/handlers"
	"github.com/dvloznov/finance-tracker/internal/api/middleware"
	infraBQ "github.com/dvloznov/finance-tracker/internal/infra/bigquery"
	"github.com/dvloznov/finance-tracker/internal/jobs/inmemory"
	"github.com/dvloznov/finance-tracker/internal/logger"
)

func main() {
	// Parse command-line flags
	var (
		port   = flag.String("port", "8080", "HTTP server port")
		bucket = flag.String("bucket", "", "GCS bucket name for document uploads")
	)
	flag.Parse()

	// Initialize logger
	log := logger.New()

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

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	// Close job queue
	if err := jobQueue.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close job queue")
	}

	log.Info().Msg("Server exited")
}
