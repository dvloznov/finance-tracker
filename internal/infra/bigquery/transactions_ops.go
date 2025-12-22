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
// using the provided BigQuery client. Uses DML INSERT to avoid streaming buffer issues.
func InsertTransactionsWithClient(ctx context.Context, client *bigquery.Client, rows []*TransactionRow) error {
	if len(rows) == 0 {
		return nil
	}

	// Build INSERT statement with multiple rows
	queryStr := `
		INSERT INTO ` + "`" + txProjectID + "." + txDatasetID + ".transactions" + "`" + ` (
			transaction_id, user_id, account_id, document_id, parsing_run_id,
			transaction_date, posting_date, booking_datetime,
			amount, currency, balance_after, direction,
			raw_description, normalized_description,
			category_id, category_name, subcategory_name,
			statement_line_no, statement_page_no,
			is_pending, is_refund, is_internal_transfer, is_split_parent, is_split_child,
			external_reference, tags, created_ts, updated_ts
		)
		VALUES
	`

	// Build parameters for each row
	var params []bigquery.QueryParameter
	for i, row := range rows {
		if i > 0 {
			queryStr += ","
		}
		queryStr += fmt.Sprintf(`
			(@transaction_id_%d, @user_id_%d, @account_id_%d, @document_id_%d, @parsing_run_id_%d,
			 @transaction_date_%d, @posting_date_%d, @booking_datetime_%d,
			 @amount_%d, @currency_%d, @balance_after_%d, @direction_%d,
			 @raw_description_%d, @normalized_description_%d,
			 @category_id_%d, @category_name_%d, @subcategory_name_%d,
			 @statement_line_no_%d, @statement_page_no_%d,
			 @is_pending_%d, @is_refund_%d, @is_internal_transfer_%d, @is_split_parent_%d, @is_split_child_%d,
			 @external_reference_%d, @tags_%d, @created_ts_%d, @updated_ts_%d)`, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i)

		params = append(params,
			bigquery.QueryParameter{Name: fmt.Sprintf("transaction_id_%d", i), Value: row.TransactionID},
			bigquery.QueryParameter{Name: fmt.Sprintf("user_id_%d", i), Value: row.UserID},
			bigquery.QueryParameter{Name: fmt.Sprintf("account_id_%d", i), Value: row.AccountID},
			bigquery.QueryParameter{Name: fmt.Sprintf("document_id_%d", i), Value: row.DocumentID},
			bigquery.QueryParameter{Name: fmt.Sprintf("parsing_run_id_%d", i), Value: row.ParsingRunID},
			bigquery.QueryParameter{Name: fmt.Sprintf("transaction_date_%d", i), Value: row.TransactionDate},
			bigquery.QueryParameter{Name: fmt.Sprintf("posting_date_%d", i), Value: row.PostingDate},
			bigquery.QueryParameter{Name: fmt.Sprintf("booking_datetime_%d", i), Value: row.BookingDatetime},
			bigquery.QueryParameter{Name: fmt.Sprintf("amount_%d", i), Value: row.Amount},
			bigquery.QueryParameter{Name: fmt.Sprintf("currency_%d", i), Value: row.Currency},
			bigquery.QueryParameter{Name: fmt.Sprintf("balance_after_%d", i), Value: row.BalanceAfter},
			bigquery.QueryParameter{Name: fmt.Sprintf("direction_%d", i), Value: row.Direction},
			bigquery.QueryParameter{Name: fmt.Sprintf("raw_description_%d", i), Value: row.RawDescription},
			bigquery.QueryParameter{Name: fmt.Sprintf("normalized_description_%d", i), Value: row.NormalizedDescription},
			bigquery.QueryParameter{Name: fmt.Sprintf("category_id_%d", i), Value: row.CategoryID},
			bigquery.QueryParameter{Name: fmt.Sprintf("category_name_%d", i), Value: row.CategoryName},
			bigquery.QueryParameter{Name: fmt.Sprintf("subcategory_name_%d", i), Value: row.SubcategoryName},
			bigquery.QueryParameter{Name: fmt.Sprintf("statement_line_no_%d", i), Value: row.StatementLineNo},
			bigquery.QueryParameter{Name: fmt.Sprintf("statement_page_no_%d", i), Value: row.StatementPageNo},
			bigquery.QueryParameter{Name: fmt.Sprintf("is_pending_%d", i), Value: row.IsPending},
			bigquery.QueryParameter{Name: fmt.Sprintf("is_refund_%d", i), Value: row.IsRefund},
			bigquery.QueryParameter{Name: fmt.Sprintf("is_internal_transfer_%d", i), Value: row.IsInternalTransfer},
			bigquery.QueryParameter{Name: fmt.Sprintf("is_split_parent_%d", i), Value: row.IsSplitParent},
			bigquery.QueryParameter{Name: fmt.Sprintf("is_split_child_%d", i), Value: row.IsSplitChild},
			bigquery.QueryParameter{Name: fmt.Sprintf("external_reference_%d", i), Value: row.ExternalReference},
			bigquery.QueryParameter{Name: fmt.Sprintf("tags_%d", i), Value: row.Tags},
			bigquery.QueryParameter{Name: fmt.Sprintf("created_ts_%d", i), Value: row.CreatedTS},
			bigquery.QueryParameter{Name: fmt.Sprintf("updated_ts_%d", i), Value: row.UpdatedTS},
		)
	}

	q := client.Query(queryStr)
	q.Parameters = params

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("InsertTransactions: running insert query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("InsertTransactions: waiting for job: %w", err)
	}
	if err := status.Err(); err != nil {
		return fmt.Errorf("InsertTransactions: job error: %w", err)
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
