package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
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
		FROM %s.%s.accounts
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
