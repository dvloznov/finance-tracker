package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dvloznov/finance-tracker/internal/api/middleware"
	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/rs/zerolog"
)

// MerchantsHandler handles merchant-related endpoints.
type MerchantsHandler struct {
	repo bigquery.DocumentRepository
	log  zerolog.Logger
}

// NewMerchantsHandler creates a new merchants handler.
func NewMerchantsHandler(repo bigquery.DocumentRepository, log zerolog.Logger) *MerchantsHandler {
	return &MerchantsHandler{
		repo: repo,
		log:  log,
	}
}

type merchantResponse struct {
	MerchantID       string `json:"merchant_id"`
	MerchantName     string `json:"merchant_name"`
	NormalizedName   string `json:"normalized_name"`
	CategoryID       string `json:"category_id"`
	CategoryName     string `json:"category_name,omitempty"`
	SubcategoryName  string `json:"subcategory_name,omitempty"`
	TransactionCount int64  `json:"transaction_count"`
}

// ListMerchants handles GET /api/merchants
// Returns all merchants with resolved category names and transaction counts,
// ordered by transaction count descending.
func (h *MerchantsHandler) ListMerchants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	merchants, err := h.repo.ListMerchants(ctx)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to list merchants")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to list merchants")
		return
	}

	categories, err := h.repo.ListActiveCategories(ctx)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to list categories for merchant join")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to list categories")
		return
	}

	categoryLookup := make(map[string]bigquery.CategoryRow, len(categories))
	for _, cat := range categories {
		categoryLookup[cat.CategoryID] = cat
	}

	responses := make([]merchantResponse, 0, len(merchants))
	for _, m := range merchants {
		resp := merchantResponse{
			MerchantID:       m.MerchantID,
			MerchantName:     m.MerchantName,
			NormalizedName:   m.NormalizedName,
			CategoryID:       m.CategoryID,
			TransactionCount: m.TransactionCount,
		}
		if cat, ok := categoryLookup[m.CategoryID]; ok {
			resp.CategoryName = cat.CategoryName
			if cat.SubcategoryName.Valid {
				resp.SubcategoryName = cat.SubcategoryName.StringVal
			}
		}
		responses = append(responses, resp)
	}

	middleware.WriteJSON(w, http.StatusOK, responses)
}

// UpdateMerchantCategory handles PUT /api/merchants/{merchantID}/category
// Body: { "category_id": "..." }
func (h *MerchantsHandler) UpdateMerchantCategory(w http.ResponseWriter, r *http.Request, merchantID string) {
	ctx := r.Context()

	var body struct {
		CategoryID string `json:"category_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if body.CategoryID == "" {
		middleware.WriteError(w, http.StatusBadRequest, "category_id is required")
		return
	}

	if err := h.repo.UpdateMerchantCategory(ctx, merchantID, body.CategoryID); err != nil {
		h.log.Error().Err(err).Str("merchant_id", merchantID).Msg("Failed to update merchant category")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to update merchant category")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
