package handlers

import (
	"net/http"

	"github.com/dvloznov/finance-tracker/internal/api/middleware"
	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/rs/zerolog"
)

// AccountsHandler handles account-related endpoints.
type AccountsHandler struct {
	repo bigquery.DocumentRepository
	log  zerolog.Logger
}

// NewAccountsHandler creates a new accounts handler.
func NewAccountsHandler(repo bigquery.DocumentRepository, log zerolog.Logger) *AccountsHandler {
	return &AccountsHandler{
		repo: repo,
		log:  log,
	}
}

// ListAccounts handles GET /api/accounts
func (h *AccountsHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	accounts, err := h.repo.ListAllAccounts(ctx)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to list accounts")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to list accounts")
		return
	}

	if accounts == nil {
		accounts = []*bigquery.AccountRow{}
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"accounts": accounts,
		"count":    len(accounts),
	})
}
