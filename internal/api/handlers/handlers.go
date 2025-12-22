package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/dvloznov/finance-tracker/internal/api/middleware"
	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/dvloznov/finance-tracker/internal/jobs"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// DocumentsHandler handles document-related endpoints.
type DocumentsHandler struct {
	repo      bigquery.DocumentRepository
	publisher jobs.Publisher
	bucket    string
	log       zerolog.Logger
}

// NewDocumentsHandler creates a new documents handler.
func NewDocumentsHandler(repo bigquery.DocumentRepository, publisher jobs.Publisher, bucket string, log zerolog.Logger) *DocumentsHandler {
	return &DocumentsHandler{
		repo:      repo,
		publisher: publisher,
		bucket:    bucket,
		log:       log,
	}
}

// ListDocuments handles GET /api/documents
func (h *DocumentsHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	documents, err := h.repo.ListAllDocuments(ctx)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to list documents")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to list documents")
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"documents": documents,
		"count":     len(documents),
	})
}

// CreateUploadURL handles POST /api/documents/upload-url
func (h *DocumentsHandler) CreateUploadURL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Filename    string `json:"filename"`
		ContentType string `json:"content_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Filename == "" {
		middleware.WriteError(w, http.StatusBadRequest, "Filename is required")
		return
	}

	// Generate unique object name
	objectName := fmt.Sprintf("uploads/%s/%s", time.Now().Format("2006/01/02"), uuid.New().String()+"-"+req.Filename)
	gcsURI := fmt.Sprintf("gs://%s/%s", h.bucket, objectName)
	documentID := uuid.New().String()

	// For local development with user credentials, return direct upload URL
	// In production with service accounts, this would use signed URLs
	uploadURL := fmt.Sprintf("/api/documents/upload/%s?object_name=%s&filename=%s", documentID, url.QueryEscape(objectName), url.QueryEscape(req.Filename))

	middleware.WriteJSON(w, http.StatusOK, map[string]string{
		"upload_url":  uploadURL,
		"gcs_uri":     gcsURI,
		"object_name": objectName,
		"document_id": documentID,
	})
}

// UploadDocument handles POST /api/documents/upload/:documentId
// Direct upload endpoint for local development with user credentials
func (h *DocumentsHandler) UploadDocument(w http.ResponseWriter, r *http.Request, documentID string) {
	ctx := r.Context()

	// Get object name from query parameter (passed from CreateUploadURL)
	objectName := r.URL.Query().Get("object_name")
	if objectName == "" {
		middleware.WriteError(w, http.StatusBadRequest, "object_name is required")
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/pdf"
	}

	gcsURI := fmt.Sprintf("gs://%s/%s", h.bucket, objectName)

	// Upload to GCS
	client, err := storage.NewClient(ctx)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to create storage client")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to upload file")
		return
	}
	defer client.Close()

	wc := client.Bucket(h.bucket).Object(objectName).NewWriter(ctx)
	wc.ContentType = contentType

	// Copy request body directly to GCS
	written, err := io.Copy(wc, r.Body)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to write to GCS")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to upload file")
		return
	}

	if err := wc.Close(); err != nil {
		h.log.Error().Err(err).Msg("Failed to close GCS writer")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to upload file")
		return
	}

	h.log.Info().
		Str("document_id", documentID).
		Str("gcs_uri", gcsURI).
		Int64("bytes", written).
		Msg("File uploaded successfully")

	// Save document metadata to BigQuery
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		filename = "document.pdf"
	}
	// Clean filename - remove any path or query parameters
	if idx := strings.Index(filename, "?"); idx > 0 {
		filename = filename[:idx]
	}
	filename = filepath.Base(filename)

	doc := &bigquery.DocumentRow{
		DocumentID:       documentID,
		OriginalFilename: filename,
		GCSURI:           gcsURI,
		UploadTS:         time.Now(),
		ParsingStatus:    "PENDING",
		FileMimeType:     contentType,
	}

	if err := h.repo.InsertDocument(ctx, doc); err != nil {
		h.log.Error().Err(err).Msg("Failed to insert document metadata")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to save document metadata")
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]string{
		"document_id": documentID,
		"gcs_uri":     gcsURI,
		"status":      "uploaded",
	})
}

// EnqueueParsing handles POST /api/documents/parse
func (h *DocumentsHandler) EnqueueParsing(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DocumentID string `json:"document_id"`
		GCSURI     string `json:"gcs_uri"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.DocumentID == "" || req.GCSURI == "" {
		middleware.WriteError(w, http.StatusBadRequest, "document_id and gcs_uri are required")
		return
	}

	ctx := r.Context()

	// Create parse job
	job := &jobs.ParseDocumentJob{
		DocumentID: req.DocumentID,
		GCSURI:     req.GCSURI,
	}

	// Publish job
	if err := h.publisher.PublishParseDocument(ctx, job); err != nil {
		h.log.Error().Err(err).Msg("Failed to enqueue parsing job")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to enqueue parsing job")
		return
	}

	h.log.Info().Str("job_id", job.JobID).Str("document_id", req.DocumentID).Msg("Parsing job enqueued")

	middleware.WriteJSON(w, http.StatusAccepted, map[string]string{
		"job_id":      job.JobID,
		"document_id": req.DocumentID,
		"status":      string(job.Status),
	})
}

// generateSignedURL generates a signed URL for uploading to GCS.
func (h *DocumentsHandler) generateSignedURL(ctx context.Context, bucket, object, contentType string) (string, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	opts := &storage.SignedURLOptions{
		Method:      "PUT",
		Expires:     time.Now().Add(15 * time.Minute),
		ContentType: contentType,
		Scheme:      storage.SigningSchemeV4,
	}

	url, err := client.Bucket(bucket).SignedURL(object, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return url, nil
}

// TransactionsHandler handles transaction-related endpoints.
type TransactionsHandler struct {
	repo bigquery.DocumentRepository
	log  zerolog.Logger
}

// NewTransactionsHandler creates a new transactions handler.
func NewTransactionsHandler(repo bigquery.DocumentRepository, log zerolog.Logger) *TransactionsHandler {
	return &TransactionsHandler{
		repo: repo,
		log:  log,
	}
}

// ListTransactions handles GET /api/transactions
func (h *TransactionsHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	query := r.URL.Query()
	startDateStr := query.Get("start_date")
	endDateStr := query.Get("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "Invalid start_date format")
			return
		}
	} else {
		startDate = time.Now().AddDate(-1, 0, 0) // 1 year ago
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "Invalid end_date format")
			return
		}
	} else {
		endDate = time.Now()
	}

	transactions, err := h.repo.QueryTransactionsByDateRange(ctx, startDate, endDate)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to query transactions")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to query transactions")
		return
	}

	// Return array directly for frontend compatibility
	if transactions == nil {
		transactions = []*bigquery.TransactionRow{}
	}
	middleware.WriteJSON(w, http.StatusOK, transactions)
}

// CategoriesHandler handles category-related endpoints.
type CategoriesHandler struct {
	repo bigquery.DocumentRepository
	log  zerolog.Logger
}

// NewCategoriesHandler creates a new categories handler.
func NewCategoriesHandler(repo bigquery.DocumentRepository, log zerolog.Logger) *CategoriesHandler {
	return &CategoriesHandler{
		repo: repo,
		log:  log,
	}
}

// ListCategories handles GET /api/categories
func (h *CategoriesHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	categories, err := h.repo.ListActiveCategories(ctx)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to list categories")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to list categories")
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"categories": categories,
		"count":      len(categories),
	})
}

// JobsHandler handles job-related endpoints.
type JobsHandler struct {
	store jobs.JobStore
	log   zerolog.Logger
}

// NewJobsHandler creates a new jobs handler.
func NewJobsHandler(store jobs.JobStore, log zerolog.Logger) *JobsHandler {
	return &JobsHandler{
		store: store,
		log:   log,
	}
}

// GetJob handles GET /api/jobs/{id}
func (h *JobsHandler) GetJob(w http.ResponseWriter, r *http.Request, jobID string) {
	ctx := r.Context()

	job, err := h.store.GetJob(ctx, jobID)
	if err != nil {
		h.log.Error().Err(err).Str("job_id", jobID).Msg("Failed to get job")
		middleware.WriteError(w, http.StatusNotFound, "Job not found")
		return
	}

	middleware.WriteJSON(w, http.StatusOK, job)
}

// ListJobs handles GET /api/jobs
func (h *JobsHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	query := r.URL.Query()
	filter := jobs.JobFilter{
		DocumentID: query.Get("document_id"),
		Status:     jobs.JobStatus(query.Get("status")),
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	jobsList, err := h.store.ListJobs(ctx, filter)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to list jobs")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to list jobs")
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"jobs":  jobsList,
		"count": len(jobsList),
	})
}
