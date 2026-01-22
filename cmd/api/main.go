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
	"github.com/dvloznov/finance-tracker/internal/events"
	infraBQ "github.com/dvloznov/finance-tracker/internal/infra/bigquery"
	"github.com/dvloznov/finance-tracker/internal/jobs/inmemory"
	"github.com/dvloznov/finance-tracker/internal/logger"
)

func main() {
	// Parse command-line flags
	var (
		port          = flag.String("port", "8080", "HTTP server port")
		bucket        = flag.String("bucket", os.Getenv("GCS_BUCKET"), "GCS bucket name for document uploads (or set GCS_BUCKET env)")
		pubsubProject = flag.String("pubsub-project", os.Getenv("PUBSUB_PROJECT"), "Pub/Sub project ID (or PUBSUB_PROJECT env)")
		pubsubTopic   = flag.String("pubsub-topic", os.Getenv("PUBSUB_TOPIC"), "Pub/Sub topic ID or full name for document events (or PUBSUB_TOPIC env)")
	)
	flag.Parse()

	// Initialize logger
	log := logger.New()

	if *bucket == "" {
		log.Warn().Msg("No GCS bucket configured - document uploads will be disabled")
	}
	if *pubsubProject == "" {
		*pubsubProject = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}

	// Initialize repositories
	ctx := context.Background()

	docRepo, err := infraBQ.NewBigQueryDocumentRepository(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create document repository")
	}
	defer docRepo.Close()

	// Initialize job infrastructure (used for /api/jobs endpoints)
	jobStore := inmemory.NewStore()

	// Initialize Pub/Sub publisher for document events
	var eventPublisher events.Publisher
	if *pubsubProject == "" || *pubsubTopic == "" {
		log.Warn().Msg("Pub/Sub not configured (PUBSUB_PROJECT/PUBSUB_TOPIC) - document events will fail")
	} else {
		publisher, err := events.NewPubSubPublisher(ctx, *pubsubProject, *pubsubTopic)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create Pub/Sub publisher")
		}
		eventPublisher = publisher
		defer eventPublisher.Close()
	}

	// Initialize handlers
	documentsHandler := handlers.NewDocumentsHandler(docRepo, eventPublisher, *bucket, log)
	transactionsHandler := handlers.NewTransactionsHandler(docRepo, log)
	categoriesHandler := handlers.NewCategoriesHandler(docRepo, log)
	accountsHandler := handlers.NewAccountsHandler(docRepo, log)

	institutionRepo, err := infraBQ.NewBigQueryInstitutionRepository(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create institution repository")
	}
	defer institutionRepo.Close()
	institutionsHandler := handlers.NewInstitutionsHandler(institutionRepo, log)
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

	mux.HandleFunc("/api/documents/", func(w http.ResponseWriter, r *http.Request) {
		// Handle DELETE /api/documents/:id
		if r.Method == http.MethodDelete {
			documentID := strings.TrimPrefix(r.URL.Path, "/api/documents/")
			documentID = strings.TrimSuffix(documentID, "/")
			if documentID == "" || strings.Contains(documentID, "/") {
				middleware.WriteError(w, http.StatusBadRequest, "Invalid document ID")
				return
			}
			documentsHandler.DeleteDocument(w, r, documentID)
			return
		}
		middleware.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
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

	// Accounts endpoints
	mux.HandleFunc("/api/accounts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			accountsHandler.ListAccounts(w, r)
		} else {
			middleware.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	// Institutions endpoints
	mux.HandleFunc("/api/institutions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			institutionsHandler.ListInstitutions(w, r)
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

	log.Info().Msg("Server exited")
}
