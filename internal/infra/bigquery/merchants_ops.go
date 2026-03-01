package bigquery

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	bq "github.com/dvloznov/finance-tracker/internal/bigquery"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

const merchantsTable = "merchants"

// FindMerchantByNormalizedName finds a merchant by normalized name.
func FindMerchantByNormalizedName(ctx context.Context, normalizedName string) (*bq.MerchantRow, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("FindMerchantByNormalizedName: bigquery client: %w", err)
	}
	defer client.Close()

	return FindMerchantByNormalizedNameWithClient(ctx, client, normalizedName)
}

// FindMerchantByNormalizedNameWithClient finds a merchant by normalized name using a shared client.
func FindMerchantByNormalizedNameWithClient(ctx context.Context, client *bigquery.Client, normalizedName string) (*bq.MerchantRow, error) {
	q := client.Query(fmt.Sprintf(`
		SELECT
			merchant_id,
			merchant_name,
			normalized_name,
			category_id,
			created_ts
		FROM %s.%s
		WHERE normalized_name = @normalized_name
		LIMIT 1
	`, datasetID, merchantsTable))
	q.Parameters = []bigquery.QueryParameter{{Name: "normalized_name", Value: normalizedName}}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("FindMerchantByNormalizedName: query read: %w", err)
	}

	var row bq.MerchantRow
	if err := it.Next(&row); err != nil {
		if err == iterator.Done {
			return nil, nil
		}
		return nil, fmt.Errorf("FindMerchantByNormalizedName: iter next: %w", err)
	}

	return &row, nil
}

// InsertMerchant inserts a new merchant row and returns its ID.
func InsertMerchant(ctx context.Context, row *bq.MerchantRow) (string, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("InsertMerchant: bigquery client: %w", err)
	}
	defer client.Close()

	return InsertMerchantWithClient(ctx, client, row)
}

// ListMerchantsWithClient returns all merchants joined with their transaction counts,
// ordered by transaction_count descending so the most-used merchants appear first.
func ListMerchantsWithClient(ctx context.Context, client *bigquery.Client) ([]*bq.MerchantWithCount, error) {
	q := client.Query(fmt.Sprintf(`
		SELECT
			m.merchant_id,
			m.merchant_name,
			m.normalized_name,
			m.category_id,
			m.created_ts,
			COUNT(t.transaction_id) AS transaction_count
		FROM %s.%s AS m
		LEFT JOIN %s.%s AS t ON t.merchant_id = m.merchant_id
		GROUP BY
			m.merchant_id,
			m.merchant_name,
			m.normalized_name,
			m.category_id,
			m.created_ts
		ORDER BY transaction_count DESC, m.merchant_name ASC
	`, datasetID, merchantsTable, datasetID, "transactions"))

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListMerchants: query read: %w", err)
	}

	var merchants []*bq.MerchantWithCount
	for {
		var row bq.MerchantWithCount
		if err := it.Next(&row); err != nil {
			if err == iterator.Done {
				break
			}
			return nil, fmt.Errorf("ListMerchants: iter next: %w", err)
		}
		r := row
		merchants = append(merchants, &r)
	}

	return merchants, nil
}

// ListMerchants returns all merchants ordered by transaction count descending.
func ListMerchants(ctx context.Context) ([]*bq.MerchantWithCount, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ListMerchants: bigquery client: %w", err)
	}
	defer client.Close()

	return ListMerchantsWithClient(ctx, client)
}

// UpdateMerchantCategoryWithClient updates the category_id for a merchant using a shared client.
func UpdateMerchantCategoryWithClient(ctx context.Context, client *bigquery.Client, merchantID, categoryID string) error {
	q := client.Query(fmt.Sprintf(`
		UPDATE %s.%s
		SET category_id = @category_id
		WHERE merchant_id = @merchant_id
	`, datasetID, merchantsTable))
	q.Parameters = []bigquery.QueryParameter{
		{Name: "category_id", Value: categoryID},
		{Name: "merchant_id", Value: merchantID},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("UpdateMerchantCategory: running update query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("UpdateMerchantCategory: waiting for job: %w", err)
	}
	if err := status.Err(); err != nil {
		return fmt.Errorf("UpdateMerchantCategory: job error: %w", err)
	}

	return nil
}

// UpdateMerchantCategory updates the category_id for a merchant.
func UpdateMerchantCategory(ctx context.Context, merchantID, categoryID string) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("UpdateMerchantCategory: bigquery client: %w", err)
	}
	defer client.Close()

	return UpdateMerchantCategoryWithClient(ctx, client, merchantID, categoryID)
}

// InsertMerchantWithClient inserts a new merchant row using a shared client and returns its ID.
func InsertMerchantWithClient(ctx context.Context, client *bigquery.Client, row *bq.MerchantRow) (string, error) {
	merchantID := row.MerchantID
	if merchantID == "" {
		merchantID = uuid.NewString()
	}

	createdTS := row.CreatedTS
	if createdTS.IsZero() {
		createdTS = time.Now()
	}

	q := client.Query(fmt.Sprintf(`
		INSERT %s.%s (
			merchant_id,
			merchant_name,
			normalized_name,
			category_id,
			created_ts
		)
		VALUES (
			@merchant_id,
			@merchant_name,
			@normalized_name,
			@category_id,
			@created_ts
		)
	`, datasetID, merchantsTable))
	q.Parameters = []bigquery.QueryParameter{
		{Name: "merchant_id", Value: merchantID},
		{Name: "merchant_name", Value: row.MerchantName},
		{Name: "normalized_name", Value: row.NormalizedName},
		{Name: "category_id", Value: row.CategoryID},
		{Name: "created_ts", Value: createdTS},
	}

	job, err := q.Run(ctx)
	if err != nil {
		return "", fmt.Errorf("InsertMerchant: running insert query: %w", err)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("InsertMerchant: waiting for job: %w", err)
	}
	if err := status.Err(); err != nil {
		return "", fmt.Errorf("InsertMerchant: job error: %w", err)
	}

	return merchantID, nil
}
