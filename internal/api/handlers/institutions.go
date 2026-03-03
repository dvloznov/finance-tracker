package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dvloznov/finance-tracker/internal/api/middleware"
	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/rs/zerolog"
)

// InstitutionsHandler handles institution-related endpoints.
type InstitutionsHandler struct {
	repo bigquery.InstitutionRepository
	log  zerolog.Logger
}

// NewInstitutionsHandler creates a new institutions handler.
func NewInstitutionsHandler(repo bigquery.InstitutionRepository, log zerolog.Logger) *InstitutionsHandler {
	return &InstitutionsHandler{
		repo: repo,
		log:  log,
	}
}

// ListInstitutions handles GET /api/institutions
func (h *InstitutionsHandler) ListInstitutions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	institutions, err := h.repo.ListAllInstitutions(ctx)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to list institutions")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to list institutions")
		return
	}

	if institutions == nil {
		institutions = []*bigquery.InstitutionRow{}
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"institutions": institutions,
		"count":        len(institutions),
	})
}

// CreateInstitution handles POST /api/institutions
func (h *InstitutionsHandler) CreateInstitution(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		middleware.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	row := &bigquery.InstitutionRow{Name: name}
	institutionID, err := h.repo.CreateInstitution(ctx, row)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to create institution")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to create institution")
		return
	}

	middleware.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"institution_id": institutionID,
		"name":           name,
	})
}

// UpdateInstitution handles PATCH /api/institutions/:institutionId
func (h *InstitutionsHandler) UpdateInstitution(w http.ResponseWriter, r *http.Request, institutionID string) {
	ctx := r.Context()

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		middleware.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	existing, err := h.repo.GetInstitutionByID(ctx, institutionID)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to retrieve institution")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to update institution")
		return
	}
	if existing == nil {
		middleware.WriteError(w, http.StatusNotFound, "Institution not found")
		return
	}

	if err := h.repo.UpdateInstitution(ctx, institutionID, name); err != nil {
		h.log.Error().Err(err).Msg("Failed to update institution")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to update institution")
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"institution_id": institutionID,
		"status":         "updated",
	})
}

// DeleteInstitution handles DELETE /api/institutions/:institutionId
func (h *InstitutionsHandler) DeleteInstitution(w http.ResponseWriter, r *http.Request, institutionID string) {
	ctx := r.Context()

	existing, err := h.repo.GetInstitutionByID(ctx, institutionID)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to retrieve institution")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to delete institution")
		return
	}
	if existing == nil {
		middleware.WriteError(w, http.StatusNotFound, "Institution not found")
		return
	}

	if err := h.repo.DeleteInstitution(ctx, institutionID); err != nil {
		h.log.Error().Err(err).Msg("Failed to delete institution")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to delete institution")
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"institution_id": institutionID,
		"status":         "deleted",
	})
}
