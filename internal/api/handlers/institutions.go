package handlers

import (
	"net/http"

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
