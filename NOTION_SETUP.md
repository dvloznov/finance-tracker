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

**Purpose**: Hierarchical categorization system for transactions.

### Properties to Create:

| Property Name | Type | Description | Configuration |
|--------------|------|-------------|---------------|
| Category | Title | Category name | **REQUIRED - Must be Title property** |
| Slug | Text | URL-friendly identifier | - |
| Depth | Number | Hierarchy level (0 = top level) | - |
| Description | Text | Category description | - |
| Is Active | Checkbox | Whether category is in use | - |
| Parent Category | Relation | Link to parent category | Relation to same database (optional) |
| Created | Created time | Auto-populated | - |

### Suggested Category Structure:
Create these as starting categories:
- **Income** (depth: 0)
  - Salary (depth: 1)
  - Freelance (depth: 1)
  - Investment Income (depth: 1)
- **Housing** (depth: 0)
  - Rent/Mortgage (depth: 1)
  - Utilities (depth: 1)
  - Maintenance (depth: 1)
- **Transportation** (depth: 0)
  - Public Transit (depth: 1)
  - Fuel (depth: 1)
  - Parking (depth: 1)
- **Food & Dining** (depth: 0)
  - Groceries (depth: 1)
  - Restaurants (depth: 1)
  - Coffee Shops (depth: 1)
- **Shopping** (depth: 0)
  - Clothing (depth: 1)
  - Electronics (depth: 1)
  - Home Goods (depth: 1)
- **Healthcare** (depth: 0)
- **Entertainment** (depth: 0)
- **Travel** (depth: 0)
- **Subscriptions** (depth: 0)
- **Transfers** (depth: 0)
- **Uncategorized** (depth: 0)

### Views to Create:
- **All Categories** - Tree view showing hierarchy
- **Active Categories** - Filter: Is Active is checked
- **Top Level** - Filter: Depth equals 0
- **By Parent** - Group by: Parent Category

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
| Date | Date | When transaction occurred | **REQUIRED - Primary date field** |
| Amount | Number | Transaction amount | **REQUIRED - Format: Number with 2 decimals** |
| Currency | Select | Transaction currency | **REQUIRED - Options: GBP, USD, EUR, etc.** |
| Balance After | Number | Account balance after transaction | Format: Number with 2 decimals |
| Account | Text | Account identifier | - |
| Category | Select | Transaction category | Options: Add categories as needed |
| Subcategory | Select | Transaction subcategory | **REQUIRED - Must be Select type, Options: Add as needed** |
| Parsing Run ID | Text | Internal processing ID | - |
| Document ID | Text | Source document ID | **REQUIRED** |
| Imported At | Date | When synced to Notion | **REQUIRED** |
| Notes | Text | Additional notes/normalized description | - |
| Is Corrected | Checkbox | Manually corrected | **REQUIRED - Default: false** |

### Notes on Property Types:
- **Description**: Must be Title (primary property)
- **Date**: Must be Date type (not "Transaction Date")  
- **Subcategory**: Must be Select type (not Text)
- **Document ID, Parsing Run ID, Account, Notes**: Must be Text/Rich Text
- **Is Corrected**: Must be Checkbox
- **Imported At**: Must be Date type

### Views to Create:
- **All Transactions** - Default table view
- **This Month** - Filter: Date is within this month
- **By Subcategory** - Group by: Subcategory
- **Recent** - Sort by: Date (descending)
- **Needs Review** - Filter: Is Corrected is unchecked

---

## 5. Database Relations Setup

### Critical Relations:

1. **Transactions → Accounts**
   - Type: Relation
   - Shows which account each transaction belongs to
   - Enables rollups for account balance tracking

2. **Transactions → Categories**
   - Type: Relation
   - Allows categorization of spending
   - Enables spending analysis by category

3. **Transactions → Documents**
   - Type: Relation
   - Links transactions to source documents
   - Useful for audit trail

4. **Documents → Accounts**
   - Type: Relation
   - Shows which account a statement belongs to

5. **Categories → Parent Category** (Self-relation)
   - Type: Relation (to same database)
   - Creates hierarchy
   - Allows multi-level categorization

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

**Transactions sync:**
```
transaction_id       → Transaction ID (hidden or formula)
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
