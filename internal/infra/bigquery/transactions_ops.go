package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
)

const transactionsTable = "transactions"

// InsertTransactions inserts a batch of TransactionRow into finance.transactions.
func InsertTransactions(ctx context.Context, rows []*TransactionRow) error {
	if len(rows) == 0 {
		return nil
	}

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("InsertTransactions: bigquery client: %w", err)
	}
	defer client.Close()

	inserter := client.Dataset(datasetID).Table(transactionsTable).Inserter()
	if err := inserter.Put(ctx, rows); err != nil {
		return fmt.Errorf("InsertTransactions: inserting rows: %w", err)
	}

	return nil
}
