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
	"github.com/dvloznov/finance-tracker/internal/logger"
)

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	var (
		logFilePath   = flag.String("log-file", getEnvOrDefault("LOG_FILE", "api.log"), "Log file path (truncated on each startup; JSON format)")
		port          = flag.String("port", "8080", "HTTP server port")
		bucket        = flag.String("bucket", os.Getenv("GCS_BUCKET"), "GCS bucket name for document uploads (or GCS_BUCKET env)")
		pubsubProject = flag.String("pubsub-project", os.Getenv("PUBSUB_PROJECT"), "Pub/Sub project ID (or PUBSUB_PROJECT env)")
		pubsubTopic   = flag.String("pubsub-topic", os.Getenv("PUBSUB_TOPIC"), "Pub/Sub topic ID or full name (or PUBSUB_TOPIC env)")
	)
	flag.Parse()

	log, logCloser, err := logger.New(*logFilePath)
	if err != nil {
		fallback := logger.NewConsoleOnly()
		fallback.Fatal().Err(err).Str("log_file", *logFilePath).Msg("Failed to open log file")
	}
	defer logCloser.Close()

	if *bucket == "" {
		log.Warn().Msg("No GCS bucket configured - document uploads will be disabled")
	}
	if *pubsubProject == "" {
		*pubsubProject = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}

	ctx := context.Background()

	docRepo, err := infraBQ.NewBigQueryDocumentRepository(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create document repository")
	}
	defer docRepo.Close()

	institutionRepo, err := infraBQ.NewBigQueryInstitutionRepository(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create institution repository")
	}
	defer institutionRepo.Close()

	var eventPublisher events.Publisher
	if *pubsubProject == "" || *pubsubTopic == "" {
		log.Warn().Msg("Pub/Sub not configured (PUBSUB_PROJECT/PUBSUB_TOPIC) - document events will be disabled")
	} else {
		publisher, err := events.NewPubSubPublisher(ctx, *pubsubProject, *pubsubTopic)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create Pub/Sub publisher")
		}
		eventPublisher = publisher
		defer eventPublisher.Close()
	}

	documentsHandler := handlers.NewDocumentsHandler(docRepo, eventPublisher, *bucket, log)
	transactionsHandler := handlers.NewTransactionsHandler(docRepo, log)
	categoriesHandler := handlers.NewCategoriesHandler(docRepo, log)
	accountsHandler := handlers.NewAccountsHandler(docRepo, log)
	institutionsHandler := handlers.NewInstitutionsHandler(institutionRepo, log)

	mux := http.NewServeMux()

	// Documents
	mux.HandleFunc("GET /api/documents", documentsHandler.ListDocuments)
	mux.HandleFunc("POST /api/documents/upload-url", documentsHandler.CreateUploadURL)
	mux.HandleFunc("POST /api/documents/parse", documentsHandler.EnqueueParsing)
	mux.HandleFunc("POST /api/documents/upload/{documentID}", func(w http.ResponseWriter, r *http.Request) {
		documentsHandler.UploadDocument(w, r, r.PathValue("documentID"))
	})
	mux.HandleFunc("PUT /api/documents/upload/{documentID}", func(w http.ResponseWriter, r *http.Request) {
		documentsHandler.UploadDocument(w, r, r.PathValue("documentID"))
	})
	mux.HandleFunc("DELETE /api/documents/{documentID}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("documentID")
		if strings.Contains(id, "/") {
			middleware.WriteError(w, http.StatusBadRequest, "Invalid document ID")
			return
		}
		documentsHandler.DeleteDocument(w, r, id)
	})

	// Transactions
	mux.HandleFunc("GET /api/transactions", transactionsHandler.ListTransactions)

	// Categories
	mux.HandleFunc("GET /api/categories", categoriesHandler.ListCategories)

	// Accounts
	mux.HandleFunc("GET /api/accounts", accountsHandler.ListAccounts)

	// Institutions
	mux.HandleFunc("GET /api/institutions", institutionsHandler.ListInstitutions)

	// Health
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		middleware.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	handler := middleware.Recovery(log)(
		middleware.Logger(log)(
			middleware.RequestID(
				middleware.CORS(
					middleware.Auth(mux),
				),
			),
		),
	)

	server := &http.Server{
		Addr:           ":" + *port,
		Handler:        handler,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	go func() {
		log.Info().Str("port", *port).Msg("Starting API server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
