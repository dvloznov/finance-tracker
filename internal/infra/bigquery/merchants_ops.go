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
