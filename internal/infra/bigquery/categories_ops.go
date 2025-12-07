package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

// ListActiveCategories returns all active categories ordered by depth, parent, name.
func ListActiveCategories(ctx context.Context) ([]CategoryRow, error) {
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ListActiveCategories: bigquery client: %w", err)
	}
	defer client.Close()

	q := client.Query(`
		SELECT
		  category_id,
		  parent_category_id,
		  depth,
		  name,
		  is_active
		FROM finance.categories
		WHERE is_active = TRUE
		ORDER BY depth, parent_category_id, name
	`)

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListActiveCategories: query read: %w", err)
	}

	var rows []CategoryRow
	for {
		var r CategoryRow
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
