package bigquery

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
)

// DocumentRepository provides an interface for document-related database operations.
type DocumentRepository interface {
	// InsertDocument inserts a single DocumentRow into the database.
	InsertDocument(ctx context.Context, row *DocumentRow) error

	// GetDocumentByID retrieves a single document by its ID.
	GetDocumentByID(ctx context.Context, documentID string) (*DocumentRow, error)

	// ListAllDocuments retrieves all documents from the database.
	ListAllDocuments(ctx context.Context) ([]*DocumentRow, error)

	// DeleteDocument deletes a document and all its related data.
	DeleteDocument(ctx context.Context, documentID string) error

	// UpdateDocumentParsingStatus updates the parsing_status field for a document.
	UpdateDocumentParsingStatus(ctx context.Context, documentID, status string) error

	// UpdateDocumentAccountAndInstitution updates the account_id and institution_id for a document.
	UpdateDocumentAccountAndInstitution(ctx context.Context, documentID, accountID, institutionID string) error

	// InsertTransactions inserts a batch of TransactionRow into the database.
	InsertTransactions(ctx context.Context, rows []*TransactionRow) error

	// InsertModelOutput inserts a single ModelOutputRow into the database.
	InsertModelOutput(ctx context.Context, row *ModelOutputRow) error

	// StartParsingRun inserts a new parsing run with status=RUNNING and returns the parsing_run_id.
	StartParsingRun(ctx context.Context, documentID string) (string, error)

	// MarkParsingRunFailed sets status=FAILED, finished_ts and error_message for a parsing run.
	MarkParsingRunFailed(ctx context.Context, parsingRunID string, parseErr error) error

	// MarkParsingRunSucceeded sets status=SUCCESS and finished_ts for a parsing run.
	MarkParsingRunSucceeded(ctx context.Context, parsingRunID string) error

	// MarkParsingRunsAsSuperseded marks all non-running parsing runs for a document as SUPERSEDED.
	MarkParsingRunsAsSuperseded(ctx context.Context, documentID string) error

	// ListActiveCategories retrieves all active categories from the database.
	ListActiveCategories(ctx context.Context) ([]CategoryRow, error)

	// QueryTransactions queries transactions with optional filters.
	// Pass empty strings for institutionID/accountID to skip those filters.
	QueryTransactions(ctx context.Context, opts TransactionQuery) ([]*TransactionRow, error)


	// ListAllAccounts retrieves all accounts from the database.
	ListAllAccounts(ctx context.Context) ([]*AccountRow, error)
}

// AccountRepository provides an interface for account-related database operations.
type AccountRepository interface {
	// UpsertAccount finds an existing account by (account_number, currency) or creates a new one.
	UpsertAccount(ctx context.Context, row *AccountRow) (string, error)

	// FindAccountByNumberAndCurrency finds an account by normalized account_number, currency, and institution_id.
	FindAccountByNumberAndCurrency(ctx context.Context, accountNumber, currency, institutionID string) (*AccountRow, error)

	// ListAllAccounts retrieves all accounts from the database.
	ListAllAccounts(ctx context.Context) ([]*AccountRow, error)
}

// InstitutionRepository provides an interface for institution-related database operations.
type InstitutionRepository interface {
	// UpsertInstitution creates or updates an institution and returns its ID.
	UpsertInstitution(ctx context.Context, row *InstitutionRow) (string, error)

	// ListAllInstitutions retrieves all institutions from the database.
	ListAllInstitutions(ctx context.Context) ([]*InstitutionRow, error)
}

// CategoryRepository provides an interface for category-related database operations.
type CategoryRepository interface {
	// ListActiveCategories retrieves all active categories from the database.
	ListActiveCategories(ctx context.Context) ([]CategoryRow, error)
}

// MerchantRepository provides an interface for merchant-related database operations.
type MerchantRepository interface {
	// FindMerchantByNormalizedName finds a merchant by normalized name.
	FindMerchantByNormalizedName(ctx context.Context, normalizedName string) (*MerchantRow, error)

	// InsertMerchant inserts a new merchant and returns its ID.
	InsertMerchant(ctx context.Context, row *MerchantRow) (string, error)
}

// TransactionQuery holds optional filter parameters for querying transactions.
type TransactionQuery struct {
	StartDate     time.Time
	EndDate       time.Time
	InstitutionID string // optional
	AccountID     string // optional
}

// DocumentRow represents a document record in BigQuery.
type DocumentRow struct {
	DocumentID string `bigquery:"document_id" json:"document_id"`
	UserID     string `bigquery:"user_id" json:"user_id"`
	GCSURI     string `bigquery:"gcs_uri" json:"gcs_uri"`

	InstitutionID string `bigquery:"institution_id" json:"institution_id,omitempty"`
	AccountID     string `bigquery:"account_id" json:"account_id,omitempty"`

	UploadTS time.Time `bigquery:"upload_ts" json:"upload_ts"`

	ParsingStatus string `bigquery:"parsing_status" json:"parsing_status"`

	OriginalFilename string `bigquery:"original_filename" json:"original_filename"`
	FileMimeType     string `bigquery:"file_mime_type" json:"file_mime_type,omitempty"`
}

// InstitutionRow represents an institution record in BigQuery.
type InstitutionRow struct {
	InstitutionID string                 `bigquery:"institution_id" json:"institution_id"`
	Name          string                 `bigquery:"name" json:"name"`
	CreatedTS     time.Time              `bigquery:"created_ts" json:"created_ts,omitempty"`
	UpdatedTS     bigquery.NullTimestamp `bigquery:"updated_ts" json:"updated_ts,omitempty"`
}

// TransactionRow represents a transaction record in BigQuery.
type TransactionRow struct {
	TransactionID string `bigquery:"transaction_id" json:"transaction_id"`

	UserID        string `bigquery:"user_id" json:"user_id"`
	AccountID     string `bigquery:"account_id" json:"account_id"`
	InstitutionID string `bigquery:"institution_id" json:"institution_id"`

	DocumentID   string `bigquery:"document_id" json:"document_id"`
	ParsingRunID string `bigquery:"parsing_run_id" json:"parsing_run_id"`

	TransactionDate civil.Date `bigquery:"transaction_date" json:"transaction_date"`
	StatementDate   civil.Date `bigquery:"statement_date" json:"statement_date"`

	Amount   *big.Rat `bigquery:"amount" json:"amount"`
	Currency string   `bigquery:"currency" json:"currency"`

	BalanceAfter *big.Rat `bigquery:"balance_after" json:"balance_after,omitempty"`

	Direction bigquery.NullString `bigquery:"direction" json:"direction,omitempty"`

	RawDescription  string              `bigquery:"raw_description" json:"raw_description"`
	MerchantID      string              `bigquery:"merchant_id" json:"merchant_id"`
	MerchantName    string              `bigquery:"merchant_name" json:"merchant_name"`
	TransactionType bigquery.NullString `bigquery:"transaction_type" json:"transaction_type,omitempty"`
	CategoryID      bigquery.NullString `bigquery:"category_id" json:"category_id,omitempty"`

	CreatedTS time.Time `bigquery:"created_ts" json:"created_ts"`
}

// MerchantRow represents a merchant record in BigQuery.
type MerchantRow struct {
	MerchantID     string    `bigquery:"merchant_id" json:"merchant_id"`
	MerchantName   string    `bigquery:"merchant_name" json:"merchant_name"`
	NormalizedName string    `bigquery:"normalized_name" json:"normalized_name"`
	CategoryID     string    `bigquery:"category_id" json:"category_id"`
	CreatedTS      time.Time `bigquery:"created_ts" json:"created_ts"`
}

// MarshalJSON customizes JSON serialization for TransactionRow.
func (t TransactionRow) MarshalJSON() ([]byte, error) {
	amount := "0"
	if t.Amount != nil {
		if f, ok := t.Amount.Float64(); ok {
			amount = fmt.Sprintf("%.2f", f)
		}
	}

	var balanceAfter *string
	if t.BalanceAfter != nil {
		if f, ok := t.BalanceAfter.Float64(); ok {
			s := fmt.Sprintf("%.2f", f)
			balanceAfter = &s
		}
	}

	var direction *string
	if t.Direction.Valid {
		d := t.Direction.StringVal
		direction = &d
	}

	var categoryID *string
	if t.CategoryID.Valid {
		c := t.CategoryID.StringVal
		categoryID = &c
	}

	var transactionType *string
	if t.TransactionType.Valid {
		typeVal := t.TransactionType.StringVal
		transactionType = &typeVal
	}

	return json.Marshal(&struct {
		TransactionID   string    `json:"transaction_id"`
		UserID          string    `json:"user_id"`
		AccountID       string    `json:"account_id"`
		InstitutionID   string    `json:"institution_id"`
		DocumentID      string    `json:"document_id"`
		ParsingRunID    string    `json:"parsing_run_id"`
		TransactionDate string    `json:"transaction_date"`
		StatementDate   string    `json:"statement_date"`
		Amount          string    `json:"amount"`
		Currency        string    `json:"currency"`
		BalanceAfter    *string   `json:"balance_after,omitempty"`
		Direction       *string   `json:"direction,omitempty"`
		RawDescription  string    `json:"raw_description"`
		MerchantID      string    `json:"merchant_id"`
		MerchantName    string    `json:"merchant_name"`
		TransactionType *string   `json:"transaction_type,omitempty"`
		CategoryID      *string   `json:"category_id,omitempty"`
		CreatedTS       time.Time `json:"created_ts"`
	}{
		TransactionID:   t.TransactionID,
		UserID:          t.UserID,
		AccountID:       t.AccountID,
		InstitutionID:   t.InstitutionID,
		DocumentID:      t.DocumentID,
		ParsingRunID:    t.ParsingRunID,
		TransactionDate: t.TransactionDate.String(),
		StatementDate:   t.StatementDate.String(),
		Amount:          amount,
		Currency:        t.Currency,
		BalanceAfter:    balanceAfter,
		Direction:       direction,
		RawDescription:  t.RawDescription,
		MerchantID:      t.MerchantID,
		MerchantName:    t.MerchantName,
		TransactionType: transactionType,
		CategoryID:      categoryID,
		CreatedTS:       t.CreatedTS,
	})
}

// AccountRow represents an account record in BigQuery.
type AccountRow struct {
	AccountID string `bigquery:"account_id" json:"account_id"`

	UserID        string `bigquery:"user_id" json:"user_id"`
	InstitutionID string `bigquery:"institution_id" json:"institution_id"`
	AccountName   string `bigquery:"account_name" json:"account_name"`
	AccountNumber string `bigquery:"account_number" json:"account_number"`
	SortCode      string `bigquery:"sort_code" json:"sort_code"`
	IBAN          string `bigquery:"iban" json:"iban"`
	Currency      string `bigquery:"currency" json:"currency"`
	AccountType   string `bigquery:"account_type" json:"account_type"`
}

// CategoryRow represents a denormalized category-subcategory pair.
type CategoryRow struct {
	CategoryID      string              `bigquery:"category_id"`
	CategoryName    string              `bigquery:"category_name"`
	SubcategoryName bigquery.NullString `bigquery:"subcategory_name"`

	Slug string `bigquery:"slug"`
}

// ParsingRunRow represents a parsing run record in BigQuery.
type ParsingRunRow struct {
	ParsingRunID string `bigquery:"parsing_run_id"`
	DocumentID   string `bigquery:"document_id"`

	StartedTS  time.Time              `bigquery:"started_ts"`
	FinishedTS bigquery.NullTimestamp `bigquery:"finished_ts"`

	ParserType    string `bigquery:"parser_type"`
	ParserVersion string `bigquery:"parser_version"`

	Status       string `bigquery:"status"`
	ErrorMessage string `bigquery:"error_message"`
}

// ModelOutputRow represents a model output record in BigQuery.
type ModelOutputRow struct {
	OutputID     string `bigquery:"output_id"`
	ParsingRunID string `bigquery:"parsing_run_id"`
	DocumentID   string `bigquery:"document_id"`

	ModelName string                 `bigquery:"model_name"`
	RawJSON   bigquery.NullJSON      `bigquery:"raw_json"`
	CreatedTS bigquery.NullTimestamp `bigquery:"created_ts"`
}
