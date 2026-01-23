package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	bq "github.com/dvloznov/finance-tracker/internal/bigquery"
	"google.golang.org/api/iterator"
)

// ListActiveCategories returns all active categories ordered by depth, parent, name.
func ListActiveCategories(ctx context.Context) ([]bq.CategoryRow, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ListActiveCategories: bigquery client: %w", err)
	}
	defer client.Close()

	return ListActiveCategoriesWithClient(ctx, client)
}

// ListActiveCategoriesWithClient returns all active categories ordered by depth, parent, name
// using the provided BigQuery client.
func ListActiveCategoriesWithClient(ctx context.Context, client *bigquery.Client) ([]bq.CategoryRow, error) {
	q := client.Query(`
		SELECT
		  category_id,
		  category_name,
		  subcategory_name,
		  slug
		FROM finance.categories
		ORDER BY category_name, subcategory_name
	`)

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListActiveCategories: query read: %w", err)
	}

	var rows []bq.CategoryRow
	for {
		var r bq.CategoryRow
		err := it.Next(&r)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ListActiveCategories: iter next: %w", err)
		}
		rows = append(rows, r)
	}

	return rows, nil
}
