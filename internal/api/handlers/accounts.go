package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dvloznov/finance-tracker/internal/api/middleware"
	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/rs/zerolog"
)

// AccountsHandler handles account-related endpoints.
type AccountsHandler struct {
	repo bigquery.DocumentRepository
	acc  bigquery.AccountRepository
	inst bigquery.InstitutionRepository
	log  zerolog.Logger
}

// NewAccountsHandler creates a new accounts handler.
func NewAccountsHandler(repo bigquery.DocumentRepository, acc bigquery.AccountRepository, inst bigquery.InstitutionRepository, log zerolog.Logger) *AccountsHandler {
	return &AccountsHandler{
		repo: repo,
		acc:  acc,
		inst: inst,
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

// CreateAccount handles POST /api/accounts
func (h *AccountsHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		InstitutionID string `json:"institution_id"`
		AccountName   string `json:"account_name"`
		AccountNumber string `json:"account_number"`
		SortCode      string `json:"sort_code"`
		IBAN          string `json:"iban"`
		Currency      string `json:"currency"`
		AccountType   string `json:"account_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.InstitutionID) == "" {
		middleware.WriteError(w, http.StatusBadRequest, "institution_id is required")
		return
	}

	// Verify institution exists
	inst, err := h.inst.GetInstitutionByID(ctx, strings.TrimSpace(req.InstitutionID))
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to verify institution")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to create account")
		return
	}
	if inst == nil {
		middleware.WriteError(w, http.StatusBadRequest, "Institution not found")
		return
	}

	row := &bigquery.AccountRow{
		InstitutionID: strings.TrimSpace(req.InstitutionID),
		AccountName:   strings.TrimSpace(req.AccountName),
		AccountNumber: strings.TrimSpace(req.AccountNumber),
		SortCode:      strings.TrimSpace(req.SortCode),
		IBAN:          strings.TrimSpace(req.IBAN),
		Currency:      strings.TrimSpace(req.Currency),
		AccountType:   strings.TrimSpace(req.AccountType),
	}
	if row.Currency == "" {
		row.Currency = "GBP"
	}

	accountID, err := h.acc.CreateAccount(ctx, row)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to create account")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to create account")
		return
	}

	middleware.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"account_id": accountID,
		"account":    row,
	})
}

// UpdateAccount handles PATCH /api/accounts/:accountId
func (h *AccountsHandler) UpdateAccount(w http.ResponseWriter, r *http.Request, accountID string) {
	ctx := r.Context()

	var req struct {
		InstitutionID string `json:"institution_id"`
		AccountName   string `json:"account_name"`
		AccountNumber string `json:"account_number"`
		SortCode      string `json:"sort_code"`
		IBAN          string `json:"iban"`
		Currency      string `json:"currency"`
		AccountType   string `json:"account_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	existing, err := h.acc.GetAccountByID(ctx, accountID)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to retrieve account")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to update account")
		return
	}
	if existing == nil {
		middleware.WriteError(w, http.StatusNotFound, "Account not found")
		return
	}

	row := &bigquery.AccountRow{
		InstitutionID: existing.InstitutionID,
		AccountName:   existing.AccountName,
		AccountNumber: existing.AccountNumber,
		SortCode:      existing.SortCode,
		IBAN:          existing.IBAN,
		Currency:      existing.Currency,
		AccountType:   existing.AccountType,
	}
	if req.InstitutionID != "" {
		row.InstitutionID = strings.TrimSpace(req.InstitutionID)
	}
	if req.AccountName != "" {
		row.AccountName = strings.TrimSpace(req.AccountName)
	}
	if req.AccountNumber != "" {
		row.AccountNumber = strings.TrimSpace(req.AccountNumber)
	}
	if req.SortCode != "" {
		row.SortCode = strings.TrimSpace(req.SortCode)
	}
	if req.IBAN != "" {
		row.IBAN = strings.TrimSpace(req.IBAN)
	}
	if req.Currency != "" {
		row.Currency = strings.TrimSpace(req.Currency)
	}
	if req.AccountType != "" {
		row.AccountType = strings.TrimSpace(req.AccountType)
	}

	if err := h.acc.UpdateAccount(ctx, accountID, row); err != nil {
		h.log.Error().Err(err).Msg("Failed to update account")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to update account")
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"account_id": accountID,
		"status":     "updated",
	})
}

// DeleteAccount handles DELETE /api/accounts/:accountId
func (h *AccountsHandler) DeleteAccount(w http.ResponseWriter, r *http.Request, accountID string) {
	ctx := r.Context()

	existing, err := h.acc.GetAccountByID(ctx, accountID)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to retrieve account")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to delete account")
		return
	}
	if existing == nil {
		middleware.WriteError(w, http.StatusNotFound, "Account not found")
		return
	}

	if err := h.acc.DeleteAccount(ctx, accountID); err != nil {
		h.log.Error().Err(err).Msg("Failed to delete account")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	middleware.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"account_id": accountID,
		"status":     "deleted",
	})
}
