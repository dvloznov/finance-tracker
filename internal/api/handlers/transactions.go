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

type transactionResponse struct {
	TransactionID   string  `json:"transaction_id"`
	DocumentID      string  `json:"document_id"`
	AccountID       string  `json:"account_id,omitempty"`
	InstitutionID   string  `json:"institution_id,omitempty"`
	TransactionDate string  `json:"transaction_date"`
	Amount          string  `json:"amount"`
	Currency        string  `json:"currency"`
	RawDescription  string  `json:"raw_description"`
	CategoryID      string  `json:"category_id,omitempty"`
	CategoryName    string  `json:"category_name,omitempty"`
	SubcategoryName string  `json:"subcategory_name,omitempty"`
	BalanceAfter    *string `json:"balance_after,omitempty"`
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

	categories, err := h.repo.ListActiveCategories(ctx)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to list categories")
		categories = []bigquery.CategoryRow{}
	}
	categoryLookup := make(map[string]bigquery.CategoryRow, len(categories))
	for _, category := range categories {
		categoryLookup[category.CategoryID] = category
	}

	responses := make([]transactionResponse, 0, len(transactions))
	for _, txn := range transactions {
		amount := "0"
		if txn.Amount != nil {
			amount = txn.Amount.FloatString(2)
		}

		var balanceAfter *string
		if txn.BalanceAfter != nil {
			b := txn.BalanceAfter.FloatString(2)
			balanceAfter = &b
		}

		categoryID := ""
		if txn.CategoryID.Valid {
			categoryID = txn.CategoryID.StringVal
		}

		categoryName := ""
		subcategoryName := ""
		if categoryID != "" {
			if category, ok := categoryLookup[categoryID]; ok {
				categoryName = category.CategoryName
				if category.SubcategoryName.Valid {
					subcategoryName = category.SubcategoryName.StringVal
				}
			}
		}

		responses = append(responses, transactionResponse{
			TransactionID:   txn.TransactionID,
			DocumentID:      txn.DocumentID,
			AccountID:       txn.AccountID,
			InstitutionID:   txn.InstitutionID,
			TransactionDate: txn.TransactionDate.String(),
			Amount:          amount,
			Currency:        txn.Currency,
			RawDescription:  txn.RawDescription,
			CategoryID:      categoryID,
			CategoryName:    categoryName,
			SubcategoryName: subcategoryName,
			BalanceAfter:    balanceAfter,
		})
	}

	if responses == nil {
		responses = []transactionResponse{}
	}
	middleware.WriteJSON(w, http.StatusOK, responses)
}
