package bigquery

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

// ListAllAccounts retrieves all accounts from the database.
func ListAllAccounts(ctx context.Context) ([]*AccountRow, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ListAllAccounts: creating client: %w", err)
	}
	defer client.Close()

	return ListAllAccountsWithClient(ctx, client)
}

// ListAllAccountsWithClient retrieves all accounts using the provided BigQuery client.
func ListAllAccountsWithClient(ctx context.Context, client *bigquery.Client) ([]*AccountRow, error) {
	query := fmt.Sprintf(`
		SELECT
			account_id,
			user_id,
			institution_id,
			account_name,
			account_number,
			sort_code,
			iban,
			currency,
			account_type,
			opened_date,
			closed_date,
			is_primary,
			metadata,
			created_ts,
			updated_ts
	FROM `+"`%s.%s.accounts`"+`
	ORDER BY created_ts DESC
	`, projectID, datasetID)

	q := client.Query(query)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListAllAccountsWithClient: reading query: %w", err)
	}

	var accounts []*AccountRow
	for {
		var row AccountRow
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ListAllAccountsWithClient: iterating: %w", err)
		}
		accounts = append(accounts, &row)
	}

	return accounts, nil
}

// FindAccountByNumberAndCurrency finds an account by normalized account_number and currency.
// Returns nil if no matching account is found.
// Normalization: trims whitespace and converts to uppercase for comparison.
func FindAccountByNumberAndCurrency(ctx context.Context, accountNumber, currency string) (*AccountRow, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("FindAccountByNumberAndCurrency: creating client: %w", err)
	}
	defer client.Close()

	return FindAccountByNumberAndCurrencyWithClient(ctx, client, accountNumber, currency)
}

// FindAccountByNumberAndCurrencyWithClient finds an account using the provided BigQuery client.
func FindAccountByNumberAndCurrencyWithClient(ctx context.Context, client *bigquery.Client, accountNumber, currency string) (*AccountRow, error) {
	// Normalize inputs
	normNumber := strings.ToUpper(strings.TrimSpace(accountNumber))
	normCurrency := strings.ToUpper(strings.TrimSpace(currency))

	if normNumber == "" {
		return nil, fmt.Errorf("FindAccountByNumberAndCurrencyWithClient: account_number cannot be empty")
	}
	if normCurrency == "" {
		return nil, fmt.Errorf("FindAccountByNumberAndCurrencyWithClient: currency cannot be empty")
	}

	query := fmt.Sprintf(`
		SELECT
			account_id,
			user_id,
			institution_id,
			account_name,
			account_number,
			sort_code,
			iban,
			currency,
			account_type,
			opened_date,
			closed_date,
			is_primary,
			metadata,
			created_ts,
			updated_ts
		FROM `+"`%s.%s.accounts`"+`
		WHERE UPPER(TRIM(account_number)) = @accountNumber
		  AND UPPER(TRIM(currency)) = @currency
		ORDER BY created_ts DESC
		LIMIT 1
	`, projectID, datasetID)

	q := client.Query(query)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "accountNumber", Value: normNumber},
		{Name: "currency", Value: normCurrency},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("FindAccountByNumberAndCurrencyWithClient: reading query: %w", err)
	}

	var row AccountRow
	err = it.Next(&row)
	if err == iterator.Done {
		// No matching account found
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("FindAccountByNumberAndCurrencyWithClient: iterating: %w", err)
	}

	return &row, nil
}

// UpsertAccount finds an existing account by (account_number, currency) or creates a new one.
// Returns the account_id of the found or created account.
// If account_number is empty/null, always creates a new account (for document-scoped defaults).
func UpsertAccount(ctx context.Context, row *AccountRow) (string, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("UpsertAccount: creating client: %w", err)
	}
	defer client.Close()

	return UpsertAccountWithClient(ctx, client, row)
}

// UpsertAccountWithClient finds or creates an account using the provided BigQuery client.
func UpsertAccountWithClient(ctx context.Context, client *bigquery.Client, row *AccountRow) (string, error) {
	// If account_number is provided, try to find existing account
	if row.AccountNumber != "" && row.Currency != "" {
		existing, err := FindAccountByNumberAndCurrencyWithClient(ctx, client, row.AccountNumber, row.Currency)
		if err != nil {
			return "", fmt.Errorf("UpsertAccountWithClient: finding existing account: %w", err)
		}
		if existing != nil {
			// Account already exists - return its ID
			return existing.AccountID, nil
		}
	}

	// No existing account found or account_number is empty - create new account
	if row.AccountID == "" {
		row.AccountID = uuid.NewString()
	}

	inserter := client.Dataset(datasetID).Table("accounts").Inserter()
	if err := inserter.Put(ctx, row); err != nil {
		return "", fmt.Errorf("UpsertAccountWithClient: inserting account: %w", err)
	}

	return row.AccountID, nil
}
