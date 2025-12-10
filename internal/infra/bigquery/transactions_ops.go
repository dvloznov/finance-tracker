package bigquery

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

const (
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

	inserter := client.Dataset(datasetID).Table(transactionsTable).Inserter()
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
// using the provided BigQuery client.
func QueryTransactionsByDateRangeWithClient(ctx context.Context, client *bigquery.Client, startDate, endDate time.Time) ([]*TransactionRow, error) {
	q := client.Query(`
		SELECT
			transaction_id,
			user_id,
			account_id,
			document_id,
			parsing_run_id,
			transaction_date,
			posting_date,
			booking_datetime,
			amount,
			currency,
			balance_after,
			direction,
			raw_description,
			normalized_description,
			category_name,
			subcategory_name,
			statement_line_no,
			statement_page_no,
			is_pending,
			is_refund,
			is_internal_transfer,
			is_split_parent,
			is_split_child,
			external_reference,
			tags,
			created_ts,
			updated_ts,
			extra
		FROM finance.transactions
		WHERE transaction_date >= @start_date
		  AND transaction_date <= @end_date
		ORDER BY transaction_date, created_ts
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
