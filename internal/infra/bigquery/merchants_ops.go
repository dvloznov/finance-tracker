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
			merged_into_merchant_id,
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
// Transaction count includes direct transactions plus transitive (merchants merged into this one).
func ListMerchantsWithClient(ctx context.Context, client *bigquery.Client) ([]*bq.MerchantWithCount, error) {
	q := client.Query(fmt.Sprintf(`
		WITH merchant_resolved AS (
			SELECT
				merchant_id,
				COALESCE(merged_into_merchant_id, merchant_id) AS resolved_merchant_id
			FROM %s.%s
		),
		txn_counts AS (
			SELECT
				mr.resolved_merchant_id,
				COUNT(t.transaction_id) AS cnt
			FROM %s.%s AS t
			INNER JOIN merchant_resolved mr ON t.merchant_id = mr.merchant_id
			GROUP BY mr.resolved_merchant_id
		)
		SELECT
			m.merchant_id,
			m.merchant_name,
			m.normalized_name,
			m.category_id,
			m.merged_into_merchant_id,
			m.created_ts,
			COALESCE(tc.cnt, 0) AS transaction_count,
			m_canon.merchant_name AS canonical_merchant_name
		FROM %s.%s AS m
		LEFT JOIN txn_counts tc ON m.merchant_id = tc.resolved_merchant_id
		LEFT JOIN %s.%s AS m_canon ON m.merged_into_merchant_id = m_canon.merchant_id
		ORDER BY transaction_count DESC, m.merchant_name ASC
	`, datasetID, merchantsTable, datasetID, "transactions", datasetID, merchantsTable, datasetID, merchantsTable))

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

// GetMerchantByIDWithClient fetches a merchant by merchant_id.
func GetMerchantByIDWithClient(ctx context.Context, client *bigquery.Client, merchantID string) (*bq.MerchantRow, error) {
	q := client.Query(fmt.Sprintf(`
		SELECT
			merchant_id,
			merchant_name,
			normalized_name,
			category_id,
			merged_into_merchant_id,
			created_ts
		FROM %s.%s
		WHERE merchant_id = @merchant_id
		LIMIT 1
	`, datasetID, merchantsTable))
	q.Parameters = []bigquery.QueryParameter{{Name: "merchant_id", Value: merchantID}}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetMerchantByID: query read: %w", err)
	}

	var row bq.MerchantRow
	if err := it.Next(&row); err != nil {
		if err == iterator.Done {
			return nil, nil
		}
		return nil, fmt.Errorf("GetMerchantByID: iter next: %w", err)
	}

	return &row, nil
}

// UpdateMerchantMergeIntoWithClient merges variantMerchantID into canonicalMerchantID.
// Flattens any existing merges (e.g. C→A becomes C→B when merging A into B).
func UpdateMerchantMergeIntoWithClient(ctx context.Context, client *bigquery.Client, variantMerchantID, canonicalMerchantID string) error {
	if variantMerchantID == canonicalMerchantID {
		return fmt.Errorf("UpdateMerchantMergeInto: variant and canonical must differ")
	}

	canonical, err := GetMerchantByIDWithClient(ctx, client, canonicalMerchantID)
	if err != nil {
		return fmt.Errorf("UpdateMerchantMergeInto: get canonical: %w", err)
	}
	if canonical == nil {
		return fmt.Errorf("UpdateMerchantMergeInto: canonical merchant %q not found", canonicalMerchantID)
	}
	if canonical.MergedIntoMerchantID.Valid {
		return fmt.Errorf("UpdateMerchantMergeInto: canonical merchant %q is already merged into another", canonicalMerchantID)
	}

	variant, err := GetMerchantByIDWithClient(ctx, client, variantMerchantID)
	if err != nil {
		return fmt.Errorf("UpdateMerchantMergeInto: get variant: %w", err)
	}
	if variant == nil {
		return fmt.Errorf("UpdateMerchantMergeInto: variant merchant %q not found", variantMerchantID)
	}

	// 1. Flatten: any merchant merged into variant should be merged into canonical instead
	flattenQ := client.Query(fmt.Sprintf(`
		UPDATE %s.%s
		SET merged_into_merchant_id = @canonical
		WHERE merged_into_merchant_id = @variant
	`, datasetID, merchantsTable))
	flattenQ.Parameters = []bigquery.QueryParameter{
		{Name: "canonical", Value: canonicalMerchantID},
		{Name: "variant", Value: variantMerchantID},
	}
	if job, err := flattenQ.Run(ctx); err != nil {
		return fmt.Errorf("UpdateMerchantMergeInto: flatten: %w", err)
	} else if status, err := job.Wait(ctx); err != nil {
		return fmt.Errorf("UpdateMerchantMergeInto: flatten wait: %w", err)
	} else if err := status.Err(); err != nil {
		return fmt.Errorf("UpdateMerchantMergeInto: flatten job: %w", err)
	}

	// 2. Set variant's merged_into_merchant_id to canonical
	setQ := client.Query(fmt.Sprintf(`
		UPDATE %s.%s
		SET merged_into_merchant_id = @canonical
		WHERE merchant_id = @variant
	`, datasetID, merchantsTable))
	setQ.Parameters = []bigquery.QueryParameter{
		{Name: "canonical", Value: canonicalMerchantID},
		{Name: "variant", Value: variantMerchantID},
	}
	if job, err := setQ.Run(ctx); err != nil {
		return fmt.Errorf("UpdateMerchantMergeInto: set: %w", err)
	} else if status, err := job.Wait(ctx); err != nil {
		return fmt.Errorf("UpdateMerchantMergeInto: set wait: %w", err)
	} else if err := status.Err(); err != nil {
		return fmt.Errorf("UpdateMerchantMergeInto: set job: %w", err)
	}

	return nil
}

// UpdateMerchantMergeInto merges variantMerchantID into canonicalMerchantID.
func UpdateMerchantMergeInto(ctx context.Context, variantMerchantID, canonicalMerchantID string) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("UpdateMerchantMergeInto: bigquery client: %w", err)
	}
	defer client.Close()

	return UpdateMerchantMergeIntoWithClient(ctx, client, variantMerchantID, canonicalMerchantID)
}

// ClearMerchantMergeIntoWithClient removes the merge for a merchant (sets merged_into_merchant_id to NULL).
func ClearMerchantMergeIntoWithClient(ctx context.Context, client *bigquery.Client, merchantID string) error {
	q := client.Query(fmt.Sprintf(`
		UPDATE %s.%s
		SET merged_into_merchant_id = NULL
		WHERE merchant_id = @merchant_id AND merged_into_merchant_id IS NOT NULL
	`, datasetID, merchantsTable))
	q.Parameters = []bigquery.QueryParameter{{Name: "merchant_id", Value: merchantID}}

	job, err := q.Run(ctx)
	if err != nil {
		return fmt.Errorf("ClearMerchantMergeInto: running update query: %w", err)
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("ClearMerchantMergeInto: waiting for job: %w", err)
	}
	if err := status.Err(); err != nil {
		return fmt.Errorf("ClearMerchantMergeInto: job error: %w", err)
	}
	return nil
}

// ClearMerchantMergeInto removes the merge for a merchant.
func ClearMerchantMergeInto(ctx context.Context, merchantID string) error {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("ClearMerchantMergeInto: bigquery client: %w", err)
	}
	defer client.Close()

	return ClearMerchantMergeIntoWithClient(ctx, client, merchantID)
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
