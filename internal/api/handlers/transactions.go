package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/dvloznov/finance-tracker/internal/api/middleware"
	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/rs/zerolog"
)

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
	institutionID := strings.TrimSpace(query.Get("institution_id"))
	accountID := strings.TrimSpace(query.Get("account_id"))

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

	if institutionID != "" || accountID != "" {
		filtered := make([]*bigquery.TransactionRow, 0, len(transactions))
		for _, tx := range transactions {
			if institutionID != "" && tx.InstitutionID != institutionID {
				continue
			}
			if accountID != "" && tx.AccountID != accountID {
				continue
			}
			filtered = append(filtered, tx)
		}
		transactions = filtered
	}

	// Return array directly for frontend compatibility
	if transactions == nil {
		transactions = []*bigquery.TransactionRow{}
	}
	middleware.WriteJSON(w, http.StatusOK, transactions)
}
