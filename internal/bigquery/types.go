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

	// InsertTransactions inserts a batch of TransactionRow into the database.
	InsertTransactions(ctx context.Context, rows []*TransactionRow) error

	// InsertModelOutput inserts a single ModelOutputRow into the database.
	InsertModelOutput(ctx context.Context, row *ModelOutputRow) error

	// StartParsingRun inserts a new parsing run with status=RUNNING and returns the parsing_run_id.
	StartParsingRun(ctx context.Context, documentID string) (string, error)

	// MarkParsingRunFailed sets status=FAILED, finished_ts and error_message for a parsing run.
	MarkParsingRunFailed(ctx context.Context, parsingRunID string, parseErr error)

	// MarkParsingRunSucceeded sets status=SUCCESS and finished_ts for a parsing run.
	MarkParsingRunSucceeded(ctx context.Context, parsingRunID string) error

	// ListActiveCategories retrieves all active categories from the database.
	ListActiveCategories(ctx context.Context) ([]CategoryRow, error)

	// QueryTransactionsByDateRange queries transactions within the specified date range.
	QueryTransactionsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*TransactionRow, error)

	// ListAllAccounts retrieves all accounts from the database.
	ListAllAccounts(ctx context.Context) ([]*AccountRow, error)

	// ListAllDocuments retrieves all documents from the database.
	ListAllDocuments(ctx context.Context) ([]*DocumentRow, error)

	// FindDocumentByChecksum retrieves a document by its SHA-256 checksum.
	FindDocumentByChecksum(ctx context.Context, checksum string) (*DocumentRow, error)

	// MarkParsingRunsAsSuperseded marks all non-running parsing runs for a document as SUPERSEDED.
	MarkParsingRunsAsSuperseded(ctx context.Context, documentID string) error
}

// AccountRepository provides an interface for account-related database operations.
type AccountRepository interface {
	// UpsertAccount finds an existing account by (account_number, currency) or creates a new one.
	UpsertAccount(ctx context.Context, row *AccountRow) (string, error)

	// FindAccountByNumberAndCurrency finds an account by normalized account_number and currency.
	FindAccountByNumberAndCurrency(ctx context.Context, accountNumber, currency string) (*AccountRow, error)

	// ListAllAccounts retrieves all accounts from the database.
	ListAllAccounts(ctx context.Context) ([]*AccountRow, error)
}

// CategoryRepository provides an interface for category-related database operations.
type CategoryRepository interface {
	// ListActiveCategories retrieves all active categories from the database.
	ListActiveCategories(ctx context.Context) ([]CategoryRow, error)
}

// DocumentRow represents a document record in BigQuery.
type DocumentRow struct {
	DocumentID string `bigquery:"document_id"`
	UserID     string `bigquery:"user_id"`
	GCSURI     string `bigquery:"gcs_uri"`

	DocumentType string `bigquery:"document_type"`
	SourceSystem string `bigquery:"source_system"`

	InstitutionID string `bigquery:"institution_id"`
	AccountID     string `bigquery:"account_id"`

	StatementStartDate bigquery.NullDate `bigquery:"statement_start_date"`
	StatementEndDate   bigquery.NullDate `bigquery:"statement_end_date"`

	UploadTS    time.Time              `bigquery:"upload_ts"`
	ProcessedTS bigquery.NullTimestamp `bigquery:"processed_ts"`

	ParsingStatus string `bigquery:"parsing_status"`

	OriginalFilename string `bigquery:"original_filename"`
	FileMimeType     string `bigquery:"file_mime_type"`

	TextGCSURI string `bigquery:"text_gcs_uri"`

	ChecksumSHA256 string `bigquery:"checksum_sha256"`

	Metadata bigquery.NullJSON `bigquery:"metadata"`
}

// TransactionRow represents a transaction record in BigQuery.
type TransactionRow struct {
	TransactionID string `bigquery:"transaction_id" json:"transaction_id"`

	UserID    string `bigquery:"user_id" json:"user_id"`
	AccountID string `bigquery:"account_id" json:"account_id"`

	DocumentID   string `bigquery:"document_id" json:"document_id"`
	ParsingRunID string `bigquery:"parsing_run_id" json:"parsing_run_id"`

	TransactionDate civil.Date            `bigquery:"transaction_date" json:"transaction_date"`
	PostingDate     bigquery.NullDate     `bigquery:"posting_date" json:"posting_date,omitempty"`
	BookingDatetime bigquery.NullDateTime `bigquery:"booking_datetime" json:"booking_datetime,omitempty"`

	Amount   *big.Rat `bigquery:"amount" json:"amount"`
	Currency string   `bigquery:"currency" json:"currency"`

	BalanceAfter *big.Rat `bigquery:"balance_after" json:"balance_after,omitempty"`

	Direction bigquery.NullString `bigquery:"direction" json:"direction,omitempty"`

	RawDescription        string              `bigquery:"raw_description" json:"raw_description"`
	NormalizedDescription bigquery.NullString `bigquery:"normalized_description" json:"normalized_description,omitempty"`

	CategoryID      bigquery.NullString `bigquery:"category_id" json:"category_id,omitempty"`
	CategoryName    bigquery.NullString `bigquery:"category_name" json:"category_name,omitempty"`
	SubcategoryName bigquery.NullString `bigquery:"subcategory_name" json:"subcategory_name,omitempty"`

	StatementLineNo bigquery.NullInt64 `bigquery:"statement_line_no" json:"statement_line_no,omitempty"`
	StatementPageNo bigquery.NullInt64 `bigquery:"statement_page_no" json:"statement_page_no,omitempty"`

	IsPending          bigquery.NullBool `bigquery:"is_pending" json:"is_pending,omitempty"`
	IsRefund           bigquery.NullBool `bigquery:"is_refund" json:"is_refund,omitempty"`
	IsInternalTransfer bigquery.NullBool `bigquery:"is_internal_transfer" json:"is_internal_transfer,omitempty"`
	IsSplitParent      bigquery.NullBool `bigquery:"is_split_parent" json:"is_split_parent,omitempty"`
	IsSplitChild       bigquery.NullBool `bigquery:"is_split_child" json:"is_split_child,omitempty"`

	ExternalReference bigquery.NullString `bigquery:"external_reference" json:"external_reference,omitempty"`

	Tags []string `bigquery:"tags" json:"tags,omitempty"`

	MerchantID bigquery.NullString `bigquery:"merchant_id" json:"merchant_id,omitempty"`

	Notes bigquery.NullString `bigquery:"notes" json:"notes,omitempty"`

	ModelConfidenceScore bigquery.NullFloat64 `bigquery:"model_confidence_score" json:"model_confidence_score,omitempty"`

	CreatedTS time.Time              `bigquery:"created_ts" json:"created_ts"`
	UpdatedTS bigquery.NullTimestamp `bigquery:"updated_ts" json:"updated_ts,omitempty"`
}

// MarshalJSON customizes JSON serialization for TransactionRow.
func (t TransactionRow) MarshalJSON() ([]byte, error) {
	type Alias TransactionRow
	return json.Marshal(&struct {
		Amount       string  `json:"amount"`
		BalanceAfter *string `json:"balance_after,omitempty"`
		*Alias
	}{
		Amount: func() string {
			if t.Amount == nil {
				return "0"
			}
			f, _ := t.Amount.Float64()
			return fmt.Sprintf("%.2f", f)
		}(),
		BalanceAfter: func() *string {
			if t.BalanceAfter == nil {
				return nil
			}
			f, _ := t.BalanceAfter.Float64()
			s := fmt.Sprintf("%.2f", f)
			return &s
		}(),
		Alias: (*Alias)(&t),
	})
}

// AccountRow represents an account record in BigQuery.
type AccountRow struct {
	AccountID string `bigquery:"account_id"`

	UserID        string `bigquery:"user_id"`
	InstitutionID string `bigquery:"institution_id"`
	AccountName   string `bigquery:"account_name"`
	AccountNumber string `bigquery:"account_number"`
	SortCode      string `bigquery:"sort_code"`
	IBAN          string `bigquery:"iban"`
	Currency      string `bigquery:"currency"`
	AccountType   string `bigquery:"account_type"`

	OpenedDate bigquery.NullDate      `bigquery:"opened_date"`
	ClosedDate bigquery.NullDate      `bigquery:"closed_date"`
	IsPrimary  bigquery.NullBool      `bigquery:"is_primary"`
	Metadata   bigquery.NullJSON      `bigquery:"metadata"`
	CreatedTS  bigquery.NullTimestamp `bigquery:"created_ts"`
	UpdatedTS  bigquery.NullTimestamp `bigquery:"updated_ts"`
}

// CategoryRow represents a denormalized category-subcategory pair.
type CategoryRow struct {
	CategoryID      string              `bigquery:"category_id"`
	CategoryName    string              `bigquery:"category_name"`
	SubcategoryName bigquery.NullString `bigquery:"subcategory_name"`

	Slug string `bigquery:"slug"`

	Description bigquery.NullString `bigquery:"description"`
	IsActive    bigquery.NullBool   `bigquery:"is_active"`

	CreatedTS bigquery.NullTimestamp `bigquery:"created_ts"`
	RetiredTS bigquery.NullTimestamp `bigquery:"retired_ts"`

	Metadata bigquery.NullJSON `bigquery:"metadata"`
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

	TokensInput  bigquery.NullInt64 `bigquery:"tokens_input"`
	TokensOutput bigquery.NullInt64 `bigquery:"tokens_output"`

	Metadata bigquery.NullJSON `bigquery:"metadata"`
}

// ModelOutputRow represents a model output record in BigQuery.
type ModelOutputRow struct {
	OutputID     string `bigquery:"output_id"`
	ParsingRunID string `bigquery:"parsing_run_id"`
	DocumentID   string `bigquery:"document_id"`

	ModelName    string              `bigquery:"model_name"`
	ModelVersion bigquery.NullString `bigquery:"model_version"`

	RawJSON       bigquery.NullJSON   `bigquery:"raw_json"`
	ExtractedText bigquery.NullString `bigquery:"extracted_text"`

	CreatedTS bigquery.NullTimestamp `bigquery:"created_ts"`
	Notes     bigquery.NullString    `bigquery:"notes"`

	Metadata bigquery.NullJSON `bigquery:"metadata"`
}
