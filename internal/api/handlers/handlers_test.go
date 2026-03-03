package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dvloznov/finance-tracker/internal/api/handlers"
	"github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/rs/zerolog"
)

// ── Stub repository ──────────────────────────────────────────────────────────

type stubRepo struct {
	documents    []*bigquery.DocumentRow
	transactions []*bigquery.TransactionRow
	categories   []bigquery.CategoryRow
	accounts     []*bigquery.AccountRow

	deleteErr       error
	getDocumentErr  error
	getDocumentNil  bool // return (nil, nil) to simulate not found
}

func (r *stubRepo) InsertDocument(_ context.Context, _ *bigquery.DocumentRow) error { return nil }
func (r *stubRepo) GetDocumentByID(_ context.Context, _ string) (*bigquery.DocumentRow, error) {
	if r.getDocumentNil {
		return nil, nil
	}
	if r.getDocumentErr != nil {
		return nil, r.getDocumentErr
	}
	if len(r.documents) > 0 {
		return r.documents[0], nil
	}
	return nil, nil
}
func (r *stubRepo) ListAllDocuments(_ context.Context) ([]*bigquery.DocumentRow, error) {
	return r.documents, nil
}
func (r *stubRepo) DeleteDocument(_ context.Context, _ string) error         { return r.deleteErr }
func (r *stubRepo) UpdateDocumentParsingStatus(_ context.Context, _, _ string) error { return nil }
func (r *stubRepo) UpdateDocumentAccountAndInstitution(_ context.Context, _, _, _ string) error {
	return nil
}
func (r *stubRepo) UpdateDocumentStatementDates(_ context.Context, _, _, _ string) error {
	return nil
}
func (r *stubRepo) InsertTransactions(_ context.Context, _ []*bigquery.TransactionRow) error {
	return nil
}
func (r *stubRepo) InsertModelOutput(_ context.Context, _ *bigquery.ModelOutputRow) error {
	return nil
}
func (r *stubRepo) StartParsingRun(_ context.Context, _ string) (string, error) { return "run-1", nil }
func (r *stubRepo) MarkParsingRunFailed(_ context.Context, _ string, _ error) error { return nil }
func (r *stubRepo) MarkParsingRunSucceeded(_ context.Context, _ string) error       { return nil }
func (r *stubRepo) MarkParsingRunsAsSuperseded(_ context.Context, _ string) error   { return nil }
func (r *stubRepo) ListActiveCategories(_ context.Context) ([]bigquery.CategoryRow, error) {
	return r.categories, nil
}
func (r *stubRepo) QueryTransactions(_ context.Context, _ bigquery.TransactionQuery) ([]*bigquery.TransactionRow, error) {
	return r.transactions, nil
}
func (r *stubRepo) ListAllAccounts(_ context.Context) ([]*bigquery.AccountRow, error) {
	return r.accounts, nil
}
func (r *stubRepo) ListMerchants(_ context.Context, _ bigquery.MerchantQuery) ([]*bigquery.MerchantWithCount, error) {
	return nil, nil
}
func (r *stubRepo) UpdateMerchantCategory(_ context.Context, _, _ string) error { return nil }
func (r *stubRepo) UpdateMerchantMergeInto(_ context.Context, _, _ string) error { return nil }
func (r *stubRepo) ClearMerchantMergeInto(_ context.Context, _ string) error     { return nil }

// stubAccountRepo implements AccountRepository for tests.
type stubAccountRepo struct {
	account *bigquery.AccountRow
}

func (r *stubAccountRepo) UpsertAccount(_ context.Context, _ *bigquery.AccountRow) (string, error) {
	return "stub-account-id", nil
}
func (r *stubAccountRepo) CreateAccount(_ context.Context, _ *bigquery.AccountRow) (string, error) {
	return "stub-account-id", nil
}
func (r *stubAccountRepo) GetAccountByID(_ context.Context, accountID string) (*bigquery.AccountRow, error) {
	if r.account != nil && r.account.AccountID == accountID {
		return r.account, nil
	}
	return nil, nil
}
func (r *stubAccountRepo) UpdateAccount(_ context.Context, _ string, _ *bigquery.AccountRow) error {
	return nil
}
func (r *stubAccountRepo) DeleteAccount(_ context.Context, _ string) error {
	return nil
}
func (r *stubAccountRepo) FindAccountByNumberAndCurrency(_ context.Context, _, _, _ string) (*bigquery.AccountRow, error) {
	return nil, nil
}
func (r *stubAccountRepo) ListAllAccounts(_ context.Context) ([]*bigquery.AccountRow, error) {
	return nil, nil
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func newLogger() zerolog.Logger { return zerolog.Nop() }

func decodeJSON(t *testing.T, body *strings.Reader) map[string]interface{} {
	t.Helper()
	var out map[string]interface{}
	if err := json.NewDecoder(body).Decode(&out); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	return out
}

// ── Documents handler tests ──────────────────────────────────────────────────

func TestListDocuments_Empty(t *testing.T) {
	repo := &stubRepo{}
	h := handlers.NewDocumentsHandler(repo, &stubAccountRepo{}, nil, "test-bucket", newLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/documents", nil)
	rr := httptest.NewRecorder()
	h.ListDocuments(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestListDocuments_ReturnsDocuments(t *testing.T) {
	repo := &stubRepo{
		documents: []*bigquery.DocumentRow{
			{DocumentID: "doc-1", UserID: "denis", ParsingStatus: "PENDING"},
		},
	}
	h := handlers.NewDocumentsHandler(repo, &stubAccountRepo{}, nil, "test-bucket", newLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/documents", nil)
	rr := httptest.NewRecorder()
	h.ListDocuments(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp struct {
		Documents []map[string]interface{} `json:"documents"`
		Count     float64                  `json:"count"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("expected count=1, got %.0f", resp.Count)
	}
	if len(resp.Documents) != 1 {
		t.Errorf("expected 1 document, got %d", len(resp.Documents))
	}
}

func TestDeleteDocument_NotFound(t *testing.T) {
	repo := &stubRepo{getDocumentNil: true}
	h := handlers.NewDocumentsHandler(repo, &stubAccountRepo{}, nil, "test-bucket", newLogger())

	req := httptest.NewRequest(http.MethodDelete, "/api/documents/missing-id", nil)
	rr := httptest.NewRecorder()
	h.DeleteDocument(rr, req, "missing-id")

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestDeleteDocument_Success(t *testing.T) {
	repo := &stubRepo{
		documents: []*bigquery.DocumentRow{
			{DocumentID: "doc-1", GCSURI: "gs://bucket/doc.pdf"},
		},
	}
	h := handlers.NewDocumentsHandler(repo, &stubAccountRepo{}, nil, "test-bucket", newLogger())

	req := httptest.NewRequest(http.MethodDelete, "/api/documents/doc-1", nil)
	rr := httptest.NewRecorder()
	h.DeleteDocument(rr, req, "doc-1")

	// GCS delete will fail (no real credentials), but the DB delete succeeds.
	// We expect 200 since GCS errors are non-fatal (only a warning).
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ── Transactions handler tests ────────────────────────────────────────────────

func TestListTransactions_Empty(t *testing.T) {
	repo := &stubRepo{}
	h := handlers.NewTransactionsHandler(repo, newLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/transactions", nil)
	rr := httptest.NewRecorder()
	h.ListTransactions(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty response, got %d items", len(resp))
	}
}

func TestListTransactions_InvalidStartDate(t *testing.T) {
	repo := &stubRepo{}
	h := handlers.NewTransactionsHandler(repo, newLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/transactions?start_date=not-a-date", nil)
	rr := httptest.NewRecorder()
	h.ListTransactions(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestListTransactions_InvalidEndDate(t *testing.T) {
	repo := &stubRepo{}
	h := handlers.NewTransactionsHandler(repo, newLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/transactions?end_date=bad", nil)
	rr := httptest.NewRecorder()
	h.ListTransactions(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
