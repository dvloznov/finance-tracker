# Notion Sync Usage Guide

This document explains how to use the `sync-notion` command to synchronize data from BigQuery to Notion.

## Overview

The sync-notion tool supports syncing the following data types from BigQuery to Notion:
- **Accounts** - Bank accounts and financial institutions
- **Categories** - Transaction categories (hierarchical)
- **Documents** - Uploaded bank statements and documents
- **Transactions** - Individual financial transactions

## Prerequisites

Before running the sync, ensure:

1. You have set up Notion databases following the [NOTION_SETUP.md](../NOTION_SETUP.md) guide
2. You have a Notion integration token with access to your databases
3. You have the database IDs for each Notion database you want to sync
4. BigQuery tables are populated with data
5. You have appropriate GCP permissions to query BigQuery

## Command Line Usage

### Basic Syntax

```bash
go run cmd/sync-notion/main.go \
  --type=<sync_type> \
  --notion-token=<your_notion_token> \
  [additional flags based on sync type]
```

### Available Sync Types

- `transactions` - Sync transactions within a date range
- `accounts` - Sync all accounts
- `categories` - Sync all active categories
- `documents` - Sync all documents
- `all` - Sync all data types

## Examples

### 1. Sync Transactions

Sync transactions for a specific date range:

```bash
go run cmd/sync-notion/main.go \
  --type=transactions \
  --notion-token=secret_abc123... \
  --notion-transactions-db-id=abc123def456... \
  --start-date=2024-01-01 \
  --end-date=2024-01-31
```

### 2. Sync Accounts

Sync all accounts from BigQuery:

```bash
go run cmd/sync-notion/main.go \
  --type=accounts \
  --notion-token=secret_abc123... \
  --notion-accounts-db-id=abc123def456...
```

### 3. Sync Categories

Sync all active categories:

```bash
go run cmd/sync-notion/main.go \
  --type=categories \
  --notion-token=secret_abc123... \
  --notion-categories-db-id=abc123def456...
```

### 4. Sync Documents

Sync all documents:

```bash
go run cmd/sync-notion/main.go \
  --type=documents \
  --notion-token=secret_abc123... \
  --notion-documents-db-id=abc123def456...
```

### 5. Sync All Data

Sync all data types at once:

```bash
go run cmd/sync-notion/main.go \
  --type=all \
  --notion-token=secret_abc123... \
  --notion-accounts-db-id=abc123def456... \
  --notion-categories-db-id=def456ghi789... \
  --notion-documents-db-id=ghi789jkl012... \
  --notion-transactions-db-id=jkl012mno345... \
  --start-date=2024-01-01 \
  --end-date=2024-01-31
```

## Command Line Flags

### Required Flags

| Flag | Description | Required For |
|------|-------------|--------------|
| `--type` | Type of data to sync | All |
| `--notion-token` | Notion API integration token | All |
| `--start-date` | Start date (YYYY-MM-DD) | transactions, all |
| `--end-date` | End date (YYYY-MM-DD) | transactions, all |

### Database ID Flags

| Flag | Description | Required For |
|------|-------------|--------------|
| `--notion-transactions-db-id` | Transactions database ID | transactions, all |
| `--notion-accounts-db-id` | Accounts database ID | accounts, all (optional) |
| `--notion-categories-db-id` | Categories database ID | categories, all (optional) |
| `--notion-documents-db-id` | Documents database ID | documents, all (optional) |

### Optional Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--dry-run` | Preview changes without syncing | false |

## Dry Run Mode

Use `--dry-run` to preview what would be synced without making any changes:

```bash
go run cmd/sync-notion/main.go \
  --type=transactions \
  --notion-token=secret_abc123... \
  --notion-transactions-db-id=abc123def456... \
  --start-date=2024-01-01 \
  --end-date=2024-01-31 \
  --dry-run
```

This is useful for:
- Testing configuration
- Previewing sync results
- Estimating sync time for large datasets

## Getting Notion Credentials

### 1. Create a Notion Integration

1. Go to https://www.notion.so/my-integrations
2. Click "New integration"
3. Give it a name (e.g., "Finance Tracker Sync")
4. Select the workspace
5. Click "Submit"
6. Copy the "Internal Integration Token" (starts with `secret_`)

### 2. Get Database IDs

For each Notion database:

1. Open the database in Notion
2. Click the "..." menu in the top right
3. Select "Copy link"
4. The database ID is the part between the workspace name and the "?" in the URL:
   ```
   https://www.notion.so/workspace/abc123def456?v=...
                                  ^^^^^^^^^^^^^
                                  Database ID
   ```

### 3. Share Databases with Integration

For each database you want to sync:

1. Open the database page
2. Click "..." menu in top right
3. Scroll to "Connections"
4. Click "Connect to" and select your integration

## Best Practices

### Initial Setup

1. **Start with dry-run**: Always test with `--dry-run` first
2. **Sync master data first**: Sync accounts and categories before transactions
3. **Test with small date range**: Start with a week or month of transactions
4. **Verify data quality**: Check the synced data in Notion before proceeding

### Regular Syncing

1. **Accounts & Categories**: Sync weekly or when changes occur
2. **Documents**: Sync after each upload batch
3. **Transactions**: Sync daily or after processing new statements
4. **Use incremental syncs**: Only sync new date ranges for transactions

### Performance Tips

1. **Batch size**: The tool processes in batches of 100 items
2. **Rate limits**: Notion API has rate limits (~3 requests/second)
3. **Large datasets**: For > 1000 transactions, consider splitting by month
4. **Monitoring**: Check logs for errors or warnings during sync

## Troubleshooting

### Common Issues

**"Error: --notion-token is required"**
- Make sure you provide a valid Notion integration token
- Check that the token starts with `secret_`

**"Error: invalid start-date format"**
- Dates must be in YYYY-MM-DD format
- Example: 2024-01-01

**"Failed to create Notion page"**
- Verify the database ID is correct
- Ensure the integration has access to the database
- Check that required properties exist in the Notion database

**"Failed to query transactions/accounts/etc"**
- Verify BigQuery credentials are configured
- Check that tables exist and have data
- Ensure GCP project ID and dataset ID are set correctly

### Debugging

Enable detailed logging by checking the console output. The tool logs:
- Number of items retrieved from BigQuery
- Number of items created/updated
- Errors for individual items (with context)
- Summary statistics

## Environment Variables

The tool expects these environment variables (configured in BigQuery client):

- `GOOGLE_APPLICATION_CREDENTIALS` - Path to GCP service account key
- Project ID and Dataset ID are typically configured in the BigQuery client code

## Advanced Usage

### Automated Syncing

Create a cron job or scheduled task:

```bash
#!/bin/bash
# sync-daily.sh

# Sync yesterday's transactions
YESTERDAY=$(date -d "yesterday" +%Y-%m-%d)

go run cmd/sync-notion/main.go \
  --type=transactions \
  --notion-token=$NOTION_TOKEN \
  --notion-transactions-db-id=$NOTION_TRANSACTIONS_DB \
  --start-date=$YESTERDAY \
  --end-date=$YESTERDAY
```

### Using Environment Variables

Store credentials securely:

```bash
export NOTION_TOKEN="secret_abc123..."
export NOTION_TRANSACTIONS_DB="abc123def456..."
export NOTION_ACCOUNTS_DB="def456ghi789..."

go run cmd/sync-notion/main.go \
  --type=all \
  --notion-token=$NOTION_TOKEN \
  --notion-transactions-db-id=$NOTION_TRANSACTIONS_DB \
  --notion-accounts-db-id=$NOTION_ACCOUNTS_DB \
  --start-date=2024-01-01 \
  --end-date=2024-01-31
```

## Migration from Old Version

If you were using the old sync command (with `--notion-db-id`), update to:

**Old:**
```bash
--notion-db-id=abc123...
```

**New:**
```bash
--type=transactions \
--notion-transactions-db-id=abc123...
```

## Support

For issues or questions:
1. Check [NOTION_SETUP.md](../NOTION_SETUP.md) for database setup
2. Review logs for specific error messages
3. Verify credentials and database IDs
4. Ensure BigQuery data exists and is accessible
