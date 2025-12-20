# Notion Views & Dashboards Guide

This guide provides detailed instructions for creating useful views, charts, and dashboards in your Notion finance tracker to gain insights into your spending, income, and account activity.

---

## Overview

With your financial data synced to Notion, you can create powerful views and dashboards to:
- Track spending by category and time period
- Monitor account balances
- Identify spending patterns and trends
- Analyze income sources
- Review large transactions
- Compare spending across accounts

---

## 1. Transactions Database Views

### Essential Views

#### 1.1 All Transactions (Default)
- **Type**: Table view
- **Sort**: Date (newest first)
- **Purpose**: Complete transaction history
- **Visible Properties**: Description, Date, Amount, Currency, Category, Account, Balance After

#### 1.2 This Month
- **Type**: Table view
- **Filter**: Date is within → This month
- **Sort**: Date (newest first)
- **Purpose**: Track current month spending
- **Calculate**: Add sum of Amount at bottom to see monthly total

#### 1.3 Last Month
- **Type**: Table view
- **Filter**: Date is within → Last month
- **Sort**: Date (newest first)
- **Purpose**: Review previous month for comparison

#### 1.4 By Category
- **Type**: Table view
- **Group By**: Category
- **Sort**: Date (newest first)
- **Purpose**: See all spending grouped by category
- **Sub-totals**: Shows spending per category automatically
- **Tip**: Collapse groups to see category totals at a glance

#### 1.5 By Account
- **Type**: Table view
- **Group By**: Account
- **Sort**: Date (newest first)
- **Purpose**: Track spending per account
- **Use Case**: See which accounts you're using most

#### 1.6 Income Only
- **Type**: Table view
- **Filter**: Amount > 0
- **Sort**: Date (newest first)
- **Purpose**: Track all income sources
- **Calculate**: Sum of Amount to see total income

#### 1.7 Expenses Only
- **Type**: Table view
- **Filter**: Amount < 0
- **Sort**: Amount (smallest first, i.e., largest expenses)
- **Purpose**: Focus on outgoing money
- **Calculate**: Sum of Amount to see total expenses

#### 1.8 Large Expenses
- **Type**: Table view
- **Filter**: Amount < -100 (adjust threshold as needed)
- **Sort**: Amount (smallest first)
- **Purpose**: Identify significant purchases for review
- **Color Code**: Add a red tag or background color

#### 1.9 Recent Transactions (Last 7 Days)
- **Type**: Table view
- **Filter**: Date is within → Past week
- **Sort**: Date (newest first)
- **Purpose**: Quick review of recent activity

#### 1.10 Uncategorized Transactions
- **Type**: Table view
- **Filter**: Category is empty
- **Sort**: Date (newest first)
- **Purpose**: Find transactions that need categorization
- **Action**: Review and assign categories

#### 1.11 Pending Review
- **Type**: Table view
- **Filter**: Is Corrected is unchecked
- **Sort**: Date (newest first)
- **Purpose**: Track transactions that may need manual correction
- **Action**: Verify amounts, descriptions, categories

---

## 2. Categories Database Views

### Essential Views

#### 2.1 All Categories (Default)
- **Type**: Table view
- **Filter**: Is Active is checked
- **Purpose**: See all available categories

#### 2.2 By Parent Category
- **Type**: Table view
- **Group By**: Category Name
- **Purpose**: See subcategories grouped under parents
- **Example**: All "Food & Dining" subcategories together

#### 2.3 Top Spending Categories (This Month)
- **Type**: Table view (linked database from Transactions)
- **Filter**: Date is within → This month
- **Group By**: Category
- **Calculate**: Sum of Amount per group
- **Sort Groups**: By sum (ascending, largest expenses first)
- **Purpose**: Identify where most money goes

---

## 3. Accounts Database Views

### Essential Views

#### 3.1 All Accounts (Default)
- **Type**: Table view
- **Purpose**: Overview of all accounts

#### 3.2 Active Accounts Only
- **Type**: Table view
- **Filter**: Closed Date is empty
- **Purpose**: Focus on current accounts

#### 3.3 By Institution
- **Type**: Table view
- **Group By**: Institution
- **Purpose**: See accounts grouped by bank

#### 3.4 By Account Type
- **Type**: Gallery or Table view
- **Group By**: Account Type
- **Purpose**: Separate checking, savings, credit cards, etc.

---

## 4. Dashboard Pages

Create dedicated Notion pages with multiple linked database views for comprehensive dashboards.

### Dashboard 1: Monthly Financial Overview

**Create a new page called "Monthly Finance Dashboard"**

**Section 1: Key Metrics**
Create callout blocks or databases with formulas:
- **Total Income This Month**: Linked Transactions DB → Filter: Amount > 0, Date is this month → Calculate: Sum
- **Total Expenses This Month**: Linked Transactions DB → Filter: Amount < 0, Date is this month → Calculate: Sum
- **Net Savings This Month**: (Create a formula property or manual calculation)
- **Transaction Count**: Linked Transactions DB → Filter: Date is this month → Calculate: Count

**Section 2: This Month's Transactions**
- Linked Transactions database
- Filter: Date is within → This month
- View: Table, sorted by Date (newest first)

**Section 3: Spending by Category**
- Linked Transactions database
- Filter: Date is this month AND Amount < 0
- Group By: Category
- Shows spending breakdown by category with subtotals

**Section 4: Top 5 Expenses**
- Linked Transactions database
- Filter: Date is this month AND Amount < 0
- Sort: Amount (ascending)
- Limit: Show first 5 rows only
- Purpose: Highlight biggest purchases

---

### Dashboard 2: Spending Analysis

**Create a new page called "Spending Analysis"**

**Section 1: Year-to-Date Overview**
- **Total Spent This Year**: Linked Transactions DB → Filter: Amount < 0, Date is this year → Sum
- **Average Monthly Spending**: Total / number of months
- **Largest Single Expense**: Linked Transactions DB → Filter: Amount < 0 → Sort by Amount (ascending) → Show top 1

**Section 2: Category Breakdown (This Year)**
- Linked Transactions database
- Filter: Date is this year AND Amount < 0
- Group By: Category
- Sort Groups: By sum of Amount (ascending)
- View: Shows which categories consume most budget

**Section 3: Monthly Trend**
- Create multiple linked databases, one per month
- Or use a Timeline view with Date property
- Shows spending patterns over time
- Helps identify seasonal trends

**Section 4: Spending by Day of Week**
*Note: Notion doesn't have native day-of-week grouping, but you can:*
- Create a Formula property in Transactions: `formatDate(prop("Date"), "dddd")` to extract day name
- Group By: This new formula property
- Purpose: See if you spend more on weekends, etc.

---

### Dashboard 3: Account Summary

**Create a new page called "Account Overview"**

**Section 1: All Active Accounts**
- Linked Accounts database
- Filter: Closed Date is empty
- Show: Account Name, Institution, Account Type, Currency

**Section 2: Recent Activity by Account**
For each major account, create a sub-section:
- **Account Name - Recent Transactions**
  - Linked Transactions database
  - Filter: Account is [Specific Account] AND Date is within past 30 days
  - Sort: Date (newest first)
  - Purpose: See recent activity per account

**Section 3: Account Balance Tracking**
*Note: Balance After shows balance after each transaction*
- Create a view showing latest transaction per account
- Filter: Date is within past 7 days
- Group By: Account
- Shows most recent balance for each account

---

### Dashboard 4: Income Tracker

**Create a new page called "Income Analysis"**

**Section 1: Total Income This Year**
- Linked Transactions database
- Filter: Amount > 0 AND Date is this year
- Calculate: Sum of Amount

**Section 2: Income by Category**
- Linked Transactions database
- Filter: Amount > 0 AND Date is this year
- Group By: Category
- Purpose: Break down income sources (Salary, Freelance, Investments, etc.)

**Section 3: Monthly Income Comparison**
- Create 12 linked database views (one per month)
- Or create a Timeline view
- Shows income fluctuation across months

**Section 4: Income Transactions**
- Linked Transactions database
- Filter: Amount > 0
- Sort: Date (newest first)
- Purpose: Full list of income transactions

---

## 5. Advanced Views & Formulas

### Formula Properties to Add

#### 5.1 Month Name (in Transactions)
- **Property Name**: Month
- **Type**: Formula
- **Formula**: `formatDate(prop("Date"), "MMMM YYYY")`
- **Purpose**: Group transactions by month
- **Use**: Create view grouped by Month property

#### 5.2 Week Number (in Transactions)
- **Property Name**: Week
- **Type**: Formula
- **Formula**: `formatDate(prop("Date"), "YYYY-[W]WW")`
- **Purpose**: Group by week for detailed tracking

#### 5.3 Absolute Amount (in Transactions)
- **Property Name**: Amount (Abs)
- **Type**: Formula
- **Formula**: `abs(prop("Amount"))`
- **Purpose**: Sort expenses by magnitude regardless of sign

#### 5.4 Transaction Type (in Transactions)
- **Property Name**: Type
- **Type**: Formula
- **Formula**: `if(prop("Amount") > 0, "Income", "Expense")`
- **Purpose**: Quick filtering and grouping

#### 5.5 Quarter (in Transactions)
- **Property Name**: Quarter
- **Type**: Formula
- **Formula**: `formatDate(prop("Date"), "YYYY-[Q]Q")`
- **Purpose**: Quarterly analysis

---

## 6. Charts & Visualizations

Notion has limited native charting. Here are approaches:

### Option 1: Native Notion (Simple)
- Use **grouped views with subtotals** as visual indicators
- Collapse groups to see just totals
- Use **progress bars** (manual, via emoji or text)

### Option 2: Linked Databases + Aggregations
- Create multiple views showing different time periods side-by-side
- Example: Create 3 columns showing "Last Month", "This Month", "Next Month" expenses
- Use callouts with Sum calculations as pseudo-charts

### Option 3: Third-Party Tools (Advanced)
- Export data to Google Sheets or Excel for charts
- Use Notion integrations like NotionCharts or ChartBase
- Embed charts as images or iframes back into Notion

### Suggested Visualizations

#### 6.1 Monthly Spending Trend (Line Chart)
- **Data**: Sum of expenses per month over past 12 months
- **How**: Export to sheets or use third-party tool
- **Insight**: See if spending is increasing/decreasing

#### 6.2 Category Distribution (Pie Chart)
- **Data**: Total spending per category for a time period
- **How**: Group by Category, show sums
- **Insight**: Which categories dominate your budget

#### 6.3 Income vs Expenses (Bar Chart)
- **Data**: Monthly income and expenses side-by-side
- **How**: Create two linked databases showing income/expenses per month
- **Insight**: Are you saving or overspending?

#### 6.4 Account Balance Over Time (Line Chart)
- **Data**: Balance After from each transaction, plotted over time
- **How**: Export Balance After + Date, chart in sheets
- **Insight**: Track net worth growth

---

## 7. Rollup Properties for Advanced Analysis

### In Accounts Database

#### Total Transactions
- **Property Name**: Transaction Count
- **Type**: Rollup
- **Relation**: (Create relation from Accounts → Transactions)
- **Property**: Transaction ID
- **Calculate**: Count all
- **Purpose**: See how active each account is

#### Current Balance (Approximate)
- **Property Name**: Latest Balance
- **Type**: Rollup
- **Relation**: Accounts → Transactions
- **Property**: Balance After
- **Calculate**: Show original
- **Filter**: Latest transaction only
- **Purpose**: Quick balance check

#### Total Spent This Month (per Account)
- **Property Name**: This Month Spending
- **Type**: Rollup
- **Relation**: Accounts → Transactions
- **Property**: Amount
- **Calculate**: Sum
- **Filter**: Date is within this month AND Amount < 0

---

## 8. Workflow Tips

### Weekly Review Workflow
1. Open "Recent Transactions (Last 7 Days)" view
2. Check for any unusual transactions
3. Verify categories are correct
4. Mark reviewed transactions as "Is Corrected"

### Monthly Close Workflow
1. Open "This Month" view
2. Review total income and expenses
3. Check "Large Expenses" view for accuracy
4. Verify all transactions are categorized (check "Uncategorized" view)
5. Archive or export data if needed

### Category Analysis Workflow
1. Open "By Category" view for desired time period
2. Review category totals
3. Identify categories over budget
4. Drill into specific category to see transaction details
5. Look for optimization opportunities

---

## 9. Mobile-Friendly Views

For Notion mobile app:

### Quick Check View
- Simple table with: Description, Date, Amount, Category
- Filter: Last 7 days
- Purpose: Fast review on mobile

### Add Transaction (Manual)
- Create a simple form view with just essential fields
- Though sync handles most transactions, useful for cash purchases

---

## 10. Automation Ideas

### Notion Automations (if available)
- **When**: New transaction with Amount < -500
- **Then**: Notify me via email
- **Purpose**: Alert for large expenses

### Recurring Transaction Tracking
- Create a "Recurring" checkbox property
- Filter view for recurring transactions
- Purpose: Track subscriptions and regular bills

---

## 11. Example Dashboard Layout

```
┌─────────────────────────────────────────┐
│     MONTHLY FINANCE DASHBOARD           │
├─────────────────────────────────────────┤
│  💰 Income: $5,000  |  💸 Expenses: -$3,200  |  💚 Net: +$1,800  │
├─────────────────────────────────────────┤
│                                         │
│  📊 Spending by Category (This Month)   │
│  ┌────────────────────────────────┐    │
│  │ Food & Dining    -$800         │    │
│  │ Housing          -$1,200       │    │
│  │ Transportation   -$400         │    │
│  │ Entertainment    -$300         │    │
│  └────────────────────────────────┘    │
│                                         │
├─────────────────────────────────────────┤
│  📅 Recent Transactions                 │
│  ┌────────────────────────────────┐    │
│  │ Dec 20 | Coffee Shop | -$8.50  │    │
│  │ Dec 19 | Grocery     | -$125   │    │
│  │ Dec 18 | Salary      | +$2,500 │    │
│  └────────────────────────────────┘    │
│                                         │
├─────────────────────────────────────────┤
│  ⚠️  Large Expenses This Month          │
│  ┌────────────────────────────────┐    │
│  │ Dec 15 | Rent       | -$1,200  │    │
│  │ Dec 10 | Insurance  | -$350    │    │
│  └────────────────────────────────┘    │
└─────────────────────────────────────────┘
```

---

## 12. Getting Started Checklist

- [ ] Create all 4 databases (Accounts, Categories, Documents, Transactions)
- [ ] Set up database relations (Transactions → Accounts, Transactions → Categories)
- [ ] Run initial sync to populate data
- [ ] Create "This Month" view in Transactions
- [ ] Create "By Category" grouped view
- [ ] Create "Income Only" and "Expenses Only" views
- [ ] Set up Monthly Financial Overview dashboard page
- [ ] Add formula properties (Month, Type, etc.)
- [ ] Create "Large Expenses" view with amount threshold
- [ ] Create "Uncategorized Transactions" view for cleanup
- [ ] Bookmark your main dashboard page
- [ ] Set up weekly review routine

---

## Tips for Success

1. **Start Simple**: Begin with basic views, add complexity as needed
2. **Use Filters Wisely**: Narrow down to relevant data for each view
3. **Leverage Grouping**: Group by Category or Account to see subtotals automatically
4. **Add Descriptions**: Use view descriptions to remind yourself what each view is for
5. **Color Code**: Use property colors or tags for visual organization
6. **Regular Reviews**: Schedule weekly/monthly reviews to keep data clean
7. **Experiment**: Try different view types (Table, Board, Timeline, Gallery)
8. **Mobile Access**: Create simplified views for mobile checking

---

## Common Questions

**Q: Can I create charts in Notion?**
A: Notion has limited native charting. Use grouped views with subtotals for simple visualization, or export to Google Sheets for advanced charts.

**Q: How do I see spending over multiple months?**
A: Create a Formula property for "Month" using formatDate, then group by that property. Or create separate filtered views for each month.

**Q: Can I track budgets?**
A: Add a "Budget" Number property to Categories, then use Rollups to compare actual spending vs budget.

**Q: How do I handle split transactions?**
A: Currently, each transaction is atomic. For splits, consider creating multiple transactions or adding a "Split From" relation property.

**Q: What's the best view to review daily?**
A: "Recent Transactions (Last 7 Days)" view is ideal for daily/weekly reviews.
