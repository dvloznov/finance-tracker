# Notion Database Setup Guide

This guide provides step-by-step instructions for setting up Notion databases to visualize financial data synced from BigQuery.

## Overview

The following tables should be synced from BigQuery to Notion:
- **Accounts** - Bank accounts and financial institutions
- **Categories** - Transaction categories (hierarchical)
- **Documents** - Uploaded bank statements and documents
- **Transactions** - Individual financial transactions

Internal tables (schema_migrations, parsing_runs, model_outputs) remain in BigQuery only as they contain technical metadata.

---

## 1. Accounts Database

**Purpose**: Track bank accounts across different institutions.

### Properties to Create:

| Property Name | Type | Description | Configuration |
|--------------|------|-------------|---------------|
| Account ID | Title | Unique identifier | Primary key from BigQuery |
| Account Name | Text | Display name (e.g., "Main Checking") | - |
| Institution | Select | Bank/institution name | Options: Add as needed |
| Account Type | Select | Type of account | Options: Checking, Savings, Credit Card, Investment, Loan |
| Currency | Select | Account currency | Options: USD, EUR, GBP, etc. |
| Account Number | Text | Last 4 digits or masked number | - |
| IBAN | Text | International account number | - |
| Is Primary | Checkbox | Mark primary account | - |
| Opened Date | Date | When account was opened | - |
| Closed Date | Date | When account was closed (if applicable) | - |
| Created | Created time | Auto-populated | - |
| Last Updated | Last edited time | Auto-populated | - |

### Views to Create:
- **All Accounts** - Default view showing all accounts
- **Active Accounts** - Filter: Closed Date is empty
- **By Institution** - Group by: Institution
- **By Type** - Group by: Account Type

---

## 2. Categories Database

**Purpose**: Denormalized category-subcategory combinations for transaction categorization. Each row represents one unique category-subcategory pair (or a parent-only category).

### Properties to Create:

| Property Name | Type | Description | Configuration |
|--------------|------|-------------|---------------|
| Category | Title | Full display name | **REQUIRED - Format: "Category → Subcategory" or just "Category"** |
| Category Name | Text | Parent category name | **REQUIRED - e.g., "Food & Dining"** |
| Subcategory Name | Text | Subcategory name | Optional - empty for parent-only categories like "Healthcare" |
| Slug | Text | URL-friendly identifier | - |
| Description | Text | Category description | - |
| Is Active | Checkbox | Whether category is in use | - |
| Created | Created time | Auto-populated | - |

### Category Structure (Denormalized):
Each row is a complete category-subcategory pair that can be selected in transactions:

**With Subcategories:**
- Income → Salary
- Income → Freelance
- Income → Investment Income
- Housing → Rent/Mortgage
- Housing → Utilities
- Housing → Maintenance
- Transportation → Public Transit
- Transportation → Fuel
- Transportation → Parking
- Food & Dining → Groceries
- Food & Dining → Restaurants
- Food & Dining → Coffee Shops
- Shopping → Clothing
- Shopping → Electronics
- Shopping → Home Goods

**Parent-Only (No Subcategory):**
- Healthcare
- Entertainment
- Travel
- Subscriptions
- Transfers
- Uncategorized

Total: 21 category rows

### Views to Create:
- **All Categories** - Default table view
- **Active Categories** - Filter: Is Active is checked
- **By Category** - Group by: Category Name
- **With Subcategories** - Filter: Subcategory Name is not empty

---

## 3. Documents Database

**Purpose**: Track uploaded bank statements and financial documents.

### Properties to Create:

| Property Name | Type | Description | Configuration |
|--------------|------|-------------|---------------|
| Document ID | Title | Unique identifier | **REQUIRED - Must be Title property** |
| Original Filename | Text | File name when uploaded | - |
| Document Type | Select | Type of document | **REQUIRED - Options: Bank Statement, Credit Card Statement, Invoice, Receipt** |
| Statement Period | Text | Date range (e.g., "Jan 2024") | Formatted from start/end dates |
| Statement Start | Date | Period start date | - |
| Statement End | Date | Period end date | - |
| Upload Date | Date | When document was uploaded | **REQUIRED** |
| Processing Status | Select | Current status | Options: Uploaded, Processing, Processed, Failed |
| Processed Date | Date | When processing completed | - |
| File Type | Select | MIME type/format | Options: PDF, CSV, Excel, Image |
| GCS Link | URL | Link to Google Cloud Storage | - |

### Views to Create:
- **All Documents** - Default view
- **Recent Uploads** - Sort by: Upload Date (descending)
- **By Status** - Group by: Processing Status

---

## 4. Transactions Database

**Purpose**: The main database for all financial transactions (most important for visualization).

### Properties to Create:

| Property Name | Type | Description | Configuration |
|--------------|------|-------------|---------------|
| Description | Title | Transaction description | **REQUIRED - Must be Title property** |
| Transaction ID | Text | Unique transaction identifier | **REQUIRED - Used for sync deduplication** |
| Date | Date | When transaction occurred | **REQUIRED - Primary date field** |
| Amount | Number | Transaction amount | **REQUIRED - Format: Number with 2 decimals** |
| Currency | Select | Transaction currency | **REQUIRED - Options: GBP, USD, EUR, etc.** |
| Balance After | Number | Account balance after transaction | Format: Number with 2 decimals |
| Account | Relation | Link to Accounts database | **REQUIRED - Relation to Accounts database. Enables filtering and rollups by account.** |
| Category | Relation | Link to Categories database | **REQUIRED - Relation to Categories database** |
| Parsing Run ID | Text | Internal processing ID | - |
| Document ID | Text | Source document ID | **REQUIRED** |
| Imported At | Date | When synced to Notion | **REQUIRED** |
| Notes | Text | Additional notes/normalized description | - |
| Is Corrected | Checkbox | Manually corrected | **REQUIRED - Default: false** |

### Critical Configuration Notes:
- **Description**: Must be Title (primary property)
- **Transaction ID**: Must be Text type - Contains the unique BigQuery transaction ID for sync deduplication
- **Date**: Must be Date type (not "Transaction Date")  
- **Category**: Must be Relation type pointing to Categories database
  - Each transaction links to ONE row in the Categories database
  - That row contains both the category and subcategory information
  - Example: Transaction links to "Food & Dining → Coffee Shops" row
- **Document ID, Parsing Run ID, Account, Notes**: Must be Text/Rich Text
- **Is Corrected**: Must be Checkbox
- **Imported At**: Must be Date type

### How Category Relations Work:
Since Categories are denormalized, each transaction has a single relation to one category row that already contains both the parent category and subcategory:
- Transaction: "Starbucks" → Links to Category: "Food & Dining → Coffee Shops"
- Transaction: "Salary" → Links to Category: "Income → Salary"
- Transaction: "Doctor visit" → Links to Category: "Healthcare" (no subcategory)

### Views to Create:
- **All Transactions** - Default table view
- **This Month** - Filter: Date is within this month
- **By Category** - Group by: Category
- **By Account** - Group by: Account (enables per-account spending analysis)
- **Large Expenses** - Filter: Amount < -100 (or your threshold), Sort by: Amount ascending
- **Income** - Filter: Amount > 0

---

## 5. Database Relations Setup

1. **Transactions → Accounts**
   - Type: Relation
   - Shows which account each transaction belongs to
   - Enables rollups for account balance tracking

2. **Transactions → Categories**
   - Type: Relation
   - **Required**: Points to Categories database
   - Each transaction links to ONE category row (which contains both category and subcategory)
   - Example: "Starbucks" transaction → links to "Food & Dining → Coffee Shops" category row
   - Allows categorization and spending analysis

3. **Transactions → Documents**
   - Type: Relation
   - Links transactions to source documents
   - Useful for audit trail

4. **Documents → Accounts**
   - Type: Relation
   - Shows which account a statement belongs to

---

## 6. Recommended Dashboards

Create separate Notion pages for these dashboard views:

### Monthly Overview Dashboard
- Current month transactions table
- Income vs Expenses (sum formulas)
- Top spending categories (linked database view)
- Account balances (from Accounts database)

### Spending Analysis Dashboard
- Transactions by category (pie chart via third-party or manual)
- Monthly spending trend
- Top merchants/descriptions
- Budget vs actual (if you add budget fields)

### Account Summary Dashboard
- All accounts list
- Recent transactions per account (filtered views)
- Balance tracking over time

---

## 7. Sync Strategy

### What Gets Synced:

From BigQuery to Notion:
- ✅ Accounts (master data)
- ✅ Categories (master data)
- ✅ Documents (metadata only, not actual files)
- ✅ Transactions (all transaction records)

Stays in BigQuery only:
- ❌ schema_migrations (technical metadata)
- ❌ parsing_runs (processing metadata)
- ❌ model_outputs (AI/ML processing data)

### Sync Fields Mapping:

**Accounts sync:**
```
account_id           → Account ID (title field)
account_name         → Account Name
institution_id       → Institution
account_type         → Account Type
currency             → Currency
account_number       → Account Number
iban                 → IBAN
is_primary           → Is Primary
opened_date          → Opened Date
closed_date          → Closed Date
```

**Categories sync:**
```
category_id          → Slug
category_name        → Category Name  
subcategory_name     → Subcategory Name
full_display_name    → Category (title field)
```

**Documents sync:**
```
document_id          → Document ID (title field)
original_filename    → Original Filename
document_type        → Document Type
statement_start_date → Statement Start
statement_end_date   → Statement End
upload_datetime      → Upload Date
processing_status    → Processing Status
processed_datetime   → Processed Date
file_type            → File Type
gcs_uri              → GCS Link
```

**Transactions sync:**
```
transaction_id       → Transaction ID (text field for deduplication)
transaction_date     → Transaction Date
amount               → Amount
currency             → Currency
direction            → Direction
raw_description      → Description
category_name        → Category (lookup/relation)
account_id           → Account (relation)
document_id          → Document (relation)
balance_after        → Balance After
is_pending           → Is Pending
is_refund            → Is Refund
is_internal_transfer → Is Transfer
tags                 → Tags
created_ts           → Created
updated_ts           → Updated
```

### Sync Behavior:

The sync process ensures Notion mirrors BigQuery exactly:

1. **Delete Stale Data**: Before syncing, the system queries all existing Notion pages and deletes any that don't exist in BigQuery
   - Transactions: Deleted if Transaction ID not found in BigQuery
   - Accounts: Deleted if Account ID not found in BigQuery  
   - Categories: Deleted if Category Slug not found in BigQuery
   - Documents: Deleted if Document ID not found in BigQuery

2. **Create New Data**: After cleanup, creates new Notion pages for all current BigQuery records

3. **Deduplication**: The Transaction ID, Account ID, Document ID, and Category Slug fields enable the sync to identify and remove stale records

**Important**: This means Notion is treated as a read-only mirror of BigQuery. Any manual edits in Notion will be overwritten on the next sync.

---

## 8. Data Preparation Checklist

Before syncing data:

- [ ] Create all four Notion databases (Accounts, Categories, Documents, Transactions)
- [ ] Set up all properties as specified above
- [ ] Populate Categories with initial category structure
- [ ] Configure database relations
- [ ] Create basic views for each database
- [ ] Test with small sample data first
- [ ] Set up proper permissions (who can view/edit)
- [ ] Configure integration API access if using automation

---

## 9. Best Practices

1. **IDs as Hidden Properties**: Keep BigQuery IDs in Notion but hide them from default views
2. **Use Relations**: Always link via relations (not text) for proper data integrity
3. **Consistent Currency**: If multi-currency, always display currency alongside amounts
4. **Archive Old Data**: For performance, consider archiving transactions older than 2 years
5. **Regular Syncs**: Set up automated syncs (daily or weekly) rather than manual imports
6. **Validation**: Add validation rules where possible (e.g., amount must be positive)
7. **Templates**: Create transaction templates for common recurring transactions

---

## 10. Performance Optimization

For large transaction volumes (10,000+ records):

- Use database filters instead of loading all data
- Create date-range views (current year, last quarter, etc.)
- Archive old transactions to separate "Archive" database
- Use rollups sparingly as they can slow down views
- Consider splitting by year if you have multi-year history

---

## Notes

- This setup assumes the sync service (`cmd/sync-notion`) will handle the data synchronization
- Notion API has rate limits - implement appropriate throttling in sync code
- Consider adding a "Sync Status" or "Last Synced" property to track data freshness
- Backup Notion data regularly as it's derived from BigQuery source of truth
