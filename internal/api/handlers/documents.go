package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/dvloznov/finance-tracker/internal/api/middleware"
	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/dvloznov/finance-tracker/internal/events"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// defaultUserID is a placeholder until real authentication is implemented.
// TODO: Replace with user ID extracted from auth context.
const defaultUserID = "denis"

// DocumentsHandler handles document-related endpoints.
type DocumentsHandler struct {
	repo           bigquery.DocumentRepository
	accountRepo    bigquery.AccountRepository
	eventPublisher events.Publisher
	bucket         string
	log            zerolog.Logger
}

// NewDocumentsHandler creates a new documents handler.
func NewDocumentsHandler(repo bigquery.DocumentRepository, accountRepo bigquery.AccountRepository, publisher events.Publisher, bucket string, log zerolog.Logger) *DocumentsHandler {
	return &DocumentsHandler{
		repo:           repo,
		accountRepo:    accountRepo,
		eventPublisher: publisher,
		bucket:         bucket,
		log:            log,
	}
}

// ListDocuments handles GET /api/documents
func (h *DocumentsHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := r.URL.Query()
	institutionID := strings.TrimSpace(query.Get("institution_id"))
	accountID := strings.TrimSpace(query.Get("account_id"))

	documents, err := h.repo.ListAllDocuments(ctx)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to list documents")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to list documents")
		return
	}

	if institutionID != "" || accountID != "" {
		filtered := make([]*bigquery.DocumentRow, 0, len(documents))
		for _, doc := range documents {
			if institutionID != "" && doc.InstitutionID != institutionID {
				continue
			}
			if accountID != "" && doc.AccountID != accountID {
				continue
			}
			filtered = append(filtered, doc)
		}
		documents = filtered
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"documents": documents,
		"count":     len(documents),
	})
}

// CreateUploadURL handles POST /api/documents/upload-url
func (h *DocumentsHandler) CreateUploadURL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Filename    string  `json:"filename"`
		ContentType string  `json:"content_type"`
		AccountID   *string `json:"account_id"`
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

	// Build upload URL with optional account_id for document assignment
	uploadURL := fmt.Sprintf("/api/documents/upload/%s?object_name=%s&filename=%s", documentID, url.QueryEscape(objectName), url.QueryEscape(req.Filename))
	if req.AccountID != nil && *req.AccountID != "" {
		uploadURL += "&account_id=" + url.QueryEscape(*req.AccountID)
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"upload_url":  uploadURL,
		"gcs_uri":     gcsURI,
		"object_name": objectName,
		"document_id": documentID,
		"account_id":  req.AccountID,
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
		UserID:           defaultUserID,
		OriginalFilename: filename,
		GCSURI:           gcsURI,
		UploadTS:         time.Now(),
		ParsingStatus:    "PENDING",
		FileMimeType:     contentType,
	}

	// Optional: assign to account if account_id provided in query
	if accountID := strings.TrimSpace(r.URL.Query().Get("account_id")); accountID != "" {
		account, err := h.accountRepo.GetAccountByID(ctx, accountID)
		if err != nil {
			h.log.Error().Err(err).Str("account_id", accountID).Msg("Failed to retrieve account for document assignment")
			middleware.WriteError(w, http.StatusInternalServerError, "Failed to retrieve account")
			return
		}
		if account != nil {
			doc.AccountID = account.AccountID
			doc.InstitutionID = account.InstitutionID
		}
	}

	if err := h.repo.InsertDocument(ctx, doc); err != nil {
		h.log.Error().Err(err).Msg("Failed to insert document metadata")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to save document metadata")
		return
	}

	if h.eventPublisher == nil {
		h.log.Error().Msg("Document event publisher is not configured")
		middleware.WriteError(w, http.StatusServiceUnavailable, "Document event publisher is not configured")
		return
	}

	event := events.DocumentUploadedEvent{
		EventID:    uuid.NewString(),
		Type:       events.DocumentUploadedEventType,
		DocumentID: documentID,
		GCSURI:     gcsURI,
		Filename:   filename,
		UploadedAt: time.Now().UTC(),
		Source:     "api",
	}

	if err := h.eventPublisher.PublishDocumentUploaded(ctx, event); err != nil {
		h.log.Error().Err(err).Str("document_id", documentID).Msg("Failed to publish document uploaded event")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to publish document event")
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]string{
		"document_id": documentID,
		"gcs_uri":     gcsURI,
		"status":      "uploaded",
		"event_id":    event.EventID,
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

	if h.eventPublisher == nil {
		h.log.Error().Msg("Document event publisher is not configured")
		middleware.WriteError(w, http.StatusServiceUnavailable, "Document event publisher is not configured")
		return
	}

	event := events.DocumentUploadedEvent{
		EventID:    uuid.NewString(),
		Type:       events.DocumentUploadedEventType,
		DocumentID: req.DocumentID,
		GCSURI:     req.GCSURI,
		Filename:   "",
		UploadedAt: time.Now().UTC(),
		Source:     "api",
	}

	if err := h.eventPublisher.PublishDocumentUploaded(ctx, event); err != nil {
		h.log.Error().Err(err).Str("document_id", req.DocumentID).Msg("Failed to enqueue parsing via Pub/Sub")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to enqueue parsing")
		return
	}

	h.log.Info().Str("event_id", event.EventID).Str("document_id", req.DocumentID).Msg("Parsing event published")

	middleware.WriteJSON(w, http.StatusAccepted, map[string]string{
		"event_id":    event.EventID,
		"document_id": req.DocumentID,
		"status":      "queued",
	})
}

// UpdateDocument handles PATCH /api/documents/:documentId
// Reassigns a document to an account. Body: { "account_id": "..." } or { "account_id": null } to unassign.
func (h *DocumentsHandler) UpdateDocument(w http.ResponseWriter, r *http.Request, documentID string) {
	ctx := r.Context()

	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		middleware.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		AccountID *string `json:"account_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	doc, err := h.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to retrieve document")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to retrieve document")
		return
	}
	if doc == nil {
		middleware.WriteError(w, http.StatusNotFound, "Document not found")
		return
	}

	accountID := ""
	institutionID := ""
	if req.AccountID != nil && *req.AccountID != "" {
		account, err := h.accountRepo.GetAccountByID(ctx, *req.AccountID)
		if err != nil {
			h.log.Error().Err(err).Str("account_id", *req.AccountID).Msg("Failed to retrieve account")
			middleware.WriteError(w, http.StatusInternalServerError, "Failed to retrieve account")
			return
		}
		if account == nil {
			middleware.WriteError(w, http.StatusNotFound, "Account not found")
			return
		}
		accountID = account.AccountID
		institutionID = account.InstitutionID
	}

	if err := h.repo.UpdateDocumentAccountAndInstitution(ctx, documentID, accountID, institutionID); err != nil {
		h.log.Error().Err(err).Str("document_id", documentID).Msg("Failed to update document")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to update document")
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"document_id":     documentID,
		"account_id":      accountID,
		"institution_id":  institutionID,
		"status":          "updated",
	})
}

// DeleteDocument handles DELETE /api/documents/:documentId
// Deletes the document and all related data (transactions, parsing runs, model outputs, GCS file)
func (h *DocumentsHandler) DeleteDocument(w http.ResponseWriter, r *http.Request, documentID string) {
	ctx := r.Context()

	doc, err := h.repo.GetDocumentByID(ctx, documentID)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to retrieve document")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to retrieve document")
		return
	}
	if doc == nil {
		middleware.WriteError(w, http.StatusNotFound, "Document not found")
		return
	}

	gcsURI := doc.GCSURI

	if err := h.repo.DeleteDocument(ctx, documentID); err != nil {
		h.log.Error().Err(err).Str("document_id", documentID).Msg("Failed to delete document from BigQuery")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to delete document")
		return
	}

	// Delete from GCS
	if gcsURI != "" {
		if err := h.deleteFromGCS(ctx, gcsURI); err != nil {
			h.log.Warn().Err(err).Str("gcs_uri", gcsURI).Msg("Failed to delete file from GCS (document already deleted from database)")
			// Continue anyway - document is deleted from DB
		}
	}

	h.log.Info().
		Str("document_id", documentID).
		Str("gcs_uri", gcsURI).
		Msg("Document deleted successfully")

	middleware.WriteJSON(w, http.StatusOK, map[string]string{
		"document_id": documentID,
		"status":      "deleted",
	})
}

// deleteFromGCS deletes a file from GCS given its gs:// URI
func (h *DocumentsHandler) deleteFromGCS(ctx context.Context, gcsURI string) error {
	// Parse gs://bucket/path format
	if !strings.HasPrefix(gcsURI, "gs://") {
		return fmt.Errorf("invalid GCS URI format: %s", gcsURI)
	}

	path := strings.TrimPrefix(gcsURI, "gs://")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid GCS URI format: %s", gcsURI)
	}

	bucket := parts[0]
	objectName := parts[1]

	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	obj := client.Bucket(bucket).Object(objectName)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}
