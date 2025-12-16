package bigquery

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

const (
	// Note: projectID and datasetID are also defined in parsing_runs_ops.go
	// but redefined here for clarity and to avoid any import order issues
	txProjectID       = "studious-union-470122-v7"
	txDatasetID       = "finance"
	transactionsTable = "transactions"
	dateFormat        = "2006-01-02"
)

// InsertTransactions inserts a batch of TransactionRow into finance.transactions.
func InsertTransactions(ctx context.Context, rows []*TransactionRow) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("InsertTransactions: bigquery client: %w", err)
	}
	defer client.Close()

	return InsertTransactionsWithClient(ctx, client, rows)
}

// InsertTransactionsWithClient inserts a batch of TransactionRow into finance.transactions
// using the provided BigQuery client.
func InsertTransactionsWithClient(ctx context.Context, client *bigquery.Client, rows []*TransactionRow) error {
	if len(rows) == 0 {
		return nil
	}

	// Use fully qualified table name to avoid project ID issues
	table := client.DatasetInProject(txProjectID, txDatasetID).Table(transactionsTable)
	inserter := table.Inserter()
	if err := inserter.Put(ctx, rows); err != nil {
		return fmt.Errorf("InsertTransactions: inserting rows: %w", err)
	}

	return nil
}

// QueryTransactionsByDateRange queries transactions within the specified date range.
func QueryTransactionsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*TransactionRow, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("QueryTransactionsByDateRange: bigquery client: %w", err)
	}
	defer client.Close()

	return QueryTransactionsByDateRangeWithClient(ctx, client, startDate, endDate)
}

// QueryTransactionsByDateRangeWithClient queries transactions within the specified date range
// using the provided BigQuery client. Only includes transactions from successful parsing runs,
// excluding transactions from superseded runs.
func QueryTransactionsByDateRangeWithClient(ctx context.Context, client *bigquery.Client, startDate, endDate time.Time) ([]*TransactionRow, error) {
	q := client.Query(`
		SELECT
			t.transaction_id,
			t.user_id,
			t.account_id,
			t.document_id,
			t.parsing_run_id,
			t.transaction_date,
			t.posting_date,
			t.booking_datetime,
			t.amount,
			t.currency,
			t.balance_after,
			t.direction,
			t.raw_description,
			t.normalized_description,
			t.category_id,
			t.category_name,
			t.subcategory_name,
			t.statement_line_no,
			t.statement_page_no,
			t.is_pending,
			t.is_refund,
			t.is_internal_transfer,
			t.is_split_parent,
			t.is_split_child,
			t.external_reference,
			t.tags,
			t.created_ts,
			t.updated_ts,
			t.extra
		FROM finance.transactions t
		INNER JOIN finance.parsing_runs pr
		  ON t.parsing_run_id = pr.parsing_run_id
		WHERE t.transaction_date >= @start_date
		  AND t.transaction_date <= @end_date
		  AND pr.status = 'SUCCESS'
		ORDER BY t.transaction_date, t.created_ts
	`)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "start_date", Value: startDate.Format(dateFormat)},
		{Name: "end_date", Value: endDate.Format(dateFormat)},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("QueryTransactionsByDateRange: query read: %w", err)
	}

	var rows []*TransactionRow
	for {
		var r TransactionRow
		err := it.Next(&r)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("QueryTransactionsByDateRange: iter next: %w", err)
		}
		rows = append(rows, &r)
	}

	return rows, nil
}
