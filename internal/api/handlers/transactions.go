package handlers

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dvloznov/finance-tracker/internal/api/middleware"
	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/rs/zerolog"
)

const categoryTTL = 60 * time.Second

// TransactionsHandler handles transaction-related endpoints.
type TransactionsHandler struct {
	repo bigquery.DocumentRepository
	log  zerolog.Logger

	categoryMu        sync.Mutex
	categoryCache     []bigquery.CategoryRow
	categoryCachedAt  time.Time
}

type transactionResponse struct {
	TransactionID   string  `json:"transaction_id"`
	DocumentID      string  `json:"document_id"`
	AccountID       string  `json:"account_id,omitempty"`
	InstitutionID   string  `json:"institution_id,omitempty"`
	AccountType     string  `json:"account_type,omitempty"`
	TransactionDate string  `json:"transaction_date"`
	StatementDate   string  `json:"statement_date"`
	Amount          string  `json:"amount"`
	Currency        string  `json:"currency"`
	RawDescription  string  `json:"raw_description"`
	MerchantID      string  `json:"merchant_id"`
	MerchantName    string  `json:"merchant_name"`
	TransactionType string  `json:"transaction_type,omitempty"`
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

// cachedCategories returns the active categories, refreshing from the DB at most once per TTL.
func (h *TransactionsHandler) cachedCategories(ctx context.Context) []bigquery.CategoryRow {
	h.categoryMu.Lock()
	defer h.categoryMu.Unlock()

	if time.Since(h.categoryCachedAt) < categoryTTL && h.categoryCache != nil {
		return h.categoryCache
	}

	cats, err := h.repo.ListActiveCategories(ctx)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to refresh category cache")
		if h.categoryCache != nil {
			return h.categoryCache // serve stale on error
		}
		return []bigquery.CategoryRow{}
	}

	h.categoryCache = cats
	h.categoryCachedAt = time.Now()
	return cats
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
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "Invalid end_date format")
			return
		}
	}

	if startDateStr == "" {
		startDate = time.Time{}
	}
	if endDateStr == "" {
		endDate = time.Now()
	}

	transactions, err := h.repo.QueryTransactions(ctx, bigquery.TransactionQuery{
		StartDate:     startDate,
		EndDate:       endDate,
		InstitutionID: institutionID,
		AccountID:     accountID,
	})
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to query transactions")
		middleware.WriteError(w, http.StatusInternalServerError, "Failed to query transactions")
		return
	}

	categories := h.cachedCategories(ctx)
	categoryLookup := make(map[string]bigquery.CategoryRow, len(categories))
	for _, category := range categories {
		categoryLookup[category.CategoryID] = category
	}

	// Build account type lookup so each transaction response carries account_type.
	accountTypeLookup := h.buildAccountTypeLookup(ctx)

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

		transactionType := ""
		if txn.TransactionType.Valid {
			transactionType = txn.TransactionType.StringVal
		}

		responses = append(responses, transactionResponse{
			TransactionID:   txn.TransactionID,
			DocumentID:      txn.DocumentID,
			AccountID:       txn.AccountID,
			InstitutionID:   txn.InstitutionID,
			AccountType:     accountTypeLookup[txn.AccountID],
			TransactionDate: txn.TransactionDate.String(),
			StatementDate:   txn.StatementDate.String(),
			Amount:          amount,
			Currency:        txn.Currency,
			RawDescription:  txn.RawDescription,
			MerchantID:      txn.MerchantID,
			MerchantName:    txn.MerchantName,
			TransactionType: transactionType,
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

// buildAccountTypeLookup fetches all accounts and returns a map of account_id → account_type.
// Errors are logged and an empty map is returned so the caller degrades gracefully.
func (h *TransactionsHandler) buildAccountTypeLookup(ctx context.Context) map[string]string {
	accounts, err := h.repo.ListAllAccounts(ctx)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to fetch accounts for account_type lookup")
		return map[string]string{}
	}
	lookup := make(map[string]string, len(accounts))
	for _, a := range accounts {
		lookup[a.AccountID] = a.AccountType
	}
	return lookup
}
