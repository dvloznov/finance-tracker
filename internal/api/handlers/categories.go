package handlers

import (
	"net/http"

	"github.com/dvloznov/finance-tracker/internal/api/middleware"
	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/rs/zerolog"
)

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
