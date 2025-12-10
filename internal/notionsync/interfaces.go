package notionsync

import (
	"context"

	"github.com/jomei/notionapi"
)

// NotionService defines the interface for interacting with Notion API.
// This interface enables mocking and testing of Notion operations.
type NotionService interface {
	// CreatePage creates a new page in a Notion database with the given properties.
	CreatePage(ctx context.Context, databaseID string, properties notionapi.Properties) (*notionapi.Page, error)

	// UpdatePage updates an existing Notion page with the given properties.
	UpdatePage(ctx context.Context, pageID string, properties notionapi.Properties) (*notionapi.Page, error)

	// QueryDatabase queries a Notion database with the given filter.
	QueryDatabase(ctx context.Context, databaseID string, filter *notionapi.DatabaseQueryRequest) (*notionapi.DatabaseQueryResponse, error)
}
