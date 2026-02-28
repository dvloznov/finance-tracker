package bigquery

import (
	"context"
	"fmt"
	"math/big"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	bq "github.com/dvloznov/finance-tracker/internal/bigquery"
	"google.golang.org/api/iterator"
)

const (
	transactionsTable = "transactions"
	dateFormat        = "2006-01-02"
)

// InsertTransactions inserts a batch of TransactionRow into finance.transactions.
func InsertTransactions(ctx context.Context, rows []*bq.TransactionRow) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("InsertTransactions: bigquery client: %w", err)
	}
	defer client.Close()

	return InsertTransactionsWithClient(ctx, client, rows)
}

// InsertTransactionsWithClient inserts a batch of TransactionRow into finance.transactions
// using the provided BigQuery client. Uses DML INSERT to avoid streaming buffer issues.
func InsertTransactionsWithClient(ctx context.Context, client *bigquery.Client, rows []*bq.TransactionRow) error {
	if len(rows) == 0 {
		return nil
	}

	paramNumericString := func(v *big.Rat) bigquery.NullString {
		if v == nil {
			return bigquery.NullString{Valid: false}
		}
		return bigquery.NullString{StringVal: bigquery.NumericString(v), Valid: true}
	}

	// Build INSERT statement with multiple rows
	queryStr := `
		INSERT INTO ` + "`" + projectID + "." + datasetID + ".transactions" + "`" + ` (
			transaction_id, user_id, account_id, institution_id, document_id, parsing_run_id,
			transaction_date, statement_date, transaction_type,
			amount, currency, balance_after, direction,
			raw_description, merchant_id,
			created_ts
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
			(@transaction_id_%d, @user_id_%d, @account_id_%d, @institution_id_%d, @document_id_%d, @parsing_run_id_%d,
			 @transaction_date_%d, @statement_date_%d, @transaction_type_%d,
			 CAST(@amount_%d AS NUMERIC), @currency_%d, CAST(@balance_after_%d AS NUMERIC), @direction_%d,
			 @raw_description_%d, @merchant_id_%d,
			 @created_ts_%d)`,
			i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i)

		params = append(params,
			bigquery.QueryParameter{Name: fmt.Sprintf("transaction_id_%d", i), Value: row.TransactionID},
			bigquery.QueryParameter{Name: fmt.Sprintf("user_id_%d", i), Value: row.UserID},
			bigquery.QueryParameter{Name: fmt.Sprintf("account_id_%d", i), Value: row.AccountID},
			bigquery.QueryParameter{Name: fmt.Sprintf("institution_id_%d", i), Value: row.InstitutionID},
			bigquery.QueryParameter{Name: fmt.Sprintf("document_id_%d", i), Value: row.DocumentID},
			bigquery.QueryParameter{Name: fmt.Sprintf("parsing_run_id_%d", i), Value: row.ParsingRunID},
			bigquery.QueryParameter{Name: fmt.Sprintf("transaction_date_%d", i), Value: row.TransactionDate},
			bigquery.QueryParameter{Name: fmt.Sprintf("statement_date_%d", i), Value: row.StatementDate},
			bigquery.QueryParameter{Name: fmt.Sprintf("transaction_type_%d", i), Value: row.TransactionType},
			bigquery.QueryParameter{Name: fmt.Sprintf("amount_%d", i), Value: paramNumericString(row.Amount)},
			bigquery.QueryParameter{Name: fmt.Sprintf("currency_%d", i), Value: row.Currency},
			bigquery.QueryParameter{Name: fmt.Sprintf("balance_after_%d", i), Value: paramNumericString(row.BalanceAfter)},
			bigquery.QueryParameter{Name: fmt.Sprintf("direction_%d", i), Value: row.Direction},
			bigquery.QueryParameter{Name: fmt.Sprintf("raw_description_%d", i), Value: row.RawDescription},
			bigquery.QueryParameter{Name: fmt.Sprintf("merchant_id_%d", i), Value: row.MerchantID},
			bigquery.QueryParameter{Name: fmt.Sprintf("created_ts_%d", i), Value: row.CreatedTS},
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

// QueryTransactionsWithClient queries transactions using the provided client, with optional filters.
func QueryTransactionsWithClient(ctx context.Context, client *bigquery.Client, opts bq.TransactionQuery) ([]*bq.TransactionRow, error) {
	queryStr := `
		SELECT
			t.transaction_id,
			t.user_id,
			t.account_id,
			t.institution_id,
			t.document_id,
			t.parsing_run_id,
			t.transaction_date,
			t.statement_date,
			t.transaction_type,
			t.amount,
			t.currency,
			t.balance_after,
			t.direction,
			t.raw_description,
			t.merchant_id,
			m.merchant_name,
			m.category_id,
			t.created_ts
		FROM finance.transactions t
		INNER JOIN finance.parsing_runs pr
		  ON t.parsing_run_id = pr.parsing_run_id
		LEFT JOIN finance.merchants m
		  ON t.merchant_id = m.merchant_id
		WHERE t.transaction_date >= @start_date
		  AND t.transaction_date <= @end_date
		  AND pr.status = 'SUCCESS'
	`

	params := []bigquery.QueryParameter{
		{Name: "start_date", Value: civil.DateOf(opts.StartDate)},
		{Name: "end_date", Value: civil.DateOf(opts.EndDate)},
	}

	if opts.InstitutionID != "" {
		queryStr += "  AND t.institution_id = @institution_id\n"
		params = append(params, bigquery.QueryParameter{Name: "institution_id", Value: opts.InstitutionID})
	}
	if opts.AccountID != "" {
		queryStr += "  AND t.account_id = @account_id\n"
		params = append(params, bigquery.QueryParameter{Name: "account_id", Value: opts.AccountID})
	}

	queryStr += "ORDER BY t.transaction_date, t.created_ts"

	q := client.Query(queryStr)
	q.Parameters = params

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("QueryTransactions: query read: %w", err)
	}

	var rows []*bq.TransactionRow
	for {
		var r bq.TransactionRow
		err := it.Next(&r)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("QueryTransactions: iter next: %w", err)
		}
		rows = append(rows, &r)
	}

	return rows, nil
}
