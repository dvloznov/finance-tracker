package notionsync

import (
	"context"
	"fmt"

	"github.com/jomei/notionapi"
)

// NotionClient is the concrete implementation of NotionService using the official Notion SDK.
type NotionClient struct {
	client *notionapi.Client
}

// NewNotionClient creates a new NotionClient with the provided API token.
func NewNotionClient(token string) *NotionClient {
	return &NotionClient{
		client: notionapi.NewClient(notionapi.Token(token)),
	}
}

// CreatePage creates a new page in a Notion database with the given properties.
func (n *NotionClient) CreatePage(ctx context.Context, databaseID string, properties notionapi.Properties) (*notionapi.Page, error) {
	req := &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			Type:       notionapi.ParentTypeDatabaseID,
			DatabaseID: notionapi.DatabaseID(databaseID),
		},
		Properties: properties,
	}

	page, err := n.client.Page.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("CreatePage: %w", err)
	}

	return page, nil
}

// UpdatePage updates an existing Notion page with the given properties.
func (n *NotionClient) UpdatePage(ctx context.Context, pageID string, properties notionapi.Properties) (*notionapi.Page, error) {
	req := &notionapi.PageUpdateRequest{
		Properties: properties,
	}

	page, err := n.client.Page.Update(ctx, notionapi.PageID(pageID), req)
	if err != nil {
		return nil, fmt.Errorf("UpdatePage: %w", err)
	}

	return page, nil
}

// QueryDatabase queries a Notion database with the given filter.
func (n *NotionClient) QueryDatabase(ctx context.Context, databaseID string, filter *notionapi.DatabaseQueryRequest) (*notionapi.DatabaseQueryResponse, error) {
	resp, err := n.client.Database.Query(ctx, notionapi.DatabaseID(databaseID), filter)
	if err != nil {
		return nil, fmt.Errorf("QueryDatabase: %w", err)
	}

	return resp, nil
}

// DeletePage archives a Notion page by setting its archived property to true.
func (n *NotionClient) DeletePage(ctx context.Context, pageID string) error {
	req := &notionapi.PageUpdateRequest{
		Archived: true,
	}

	_, err := n.client.Page.Update(ctx, notionapi.PageID(pageID), req)
	if err != nil {
		return fmt.Errorf("DeletePage: %w", err)
	}

	return nil
}
