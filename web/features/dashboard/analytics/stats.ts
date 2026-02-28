import type { TransactionVM } from '@/features/transactions/types';

export type StatsSummary = {
  totalIncome: number;
  totalExpenses: number;
  netBalance: number;
};

/**
 * Returns the latest transaction (by date) that has a balance_after value,
 * within a group of transactions for a single account.
 */
function latestWithBalance(txns: TransactionVM[]): TransactionVM | undefined {
  return [...txns]
    .filter((t) => t.transaction_date && t.balance_after)
    .sort((a, b) => new Date(b.transaction_date).getTime() - new Date(a.transaction_date).getTime())[0];
}

/**
 * Compute net worth across all accounts, accounting for account type:
 * - Current/savings balance_after is an asset  → added to net worth
 * - Credit card balance_after is a liability   → subtracted from net worth
 *
 * Falls back to cumulative income - expenses when no balance_after data is available.
 */
function getNetWorth(transactions: TransactionVM[]): number {
  // Group by account_id
  const byAccount = new Map<string, TransactionVM[]>();
  for (const txn of transactions) {
    const key = txn.account_id ?? '__unknown__';
    if (!byAccount.has(key)) byAccount.set(key, []);
    byAccount.get(key)!.push(txn);
  }

  let hasAnyBalance = false;
  let total = 0;

  for (const txns of byAccount.values()) {
    const latest = latestWithBalance(txns);
    if (!latest) continue;

    hasAnyBalance = true;
    const balance = parseFloat(latest.balance_after!);
    const isCreditCard = latest.account_type === 'CREDIT_CARD';
    // Credit card balance is money owed — subtract from net worth.
    total += isCreditCard ? -balance : balance;
  }

  if (hasAnyBalance) return total;

  // No balance_after data at all — fall back to cumulative sum.
  const income = transactions.filter((t) => parseFloat(t.amount) > 0).reduce((s, t) => s + parseFloat(t.amount), 0);
  const expenses = transactions.filter((t) => parseFloat(t.amount) < 0).reduce((s, t) => s + parseFloat(t.amount), 0);
  return income + expenses;
}

export function getStatsSummary(
  transactions?: TransactionVM[] | null,
  transferIds?: Set<string>
): StatsSummary | null {
  if (!transactions || !Array.isArray(transactions)) return null;

  // Exclude both legs of detected transfer pairs from flow metrics (income/expenses).
  // Net worth (balance_after snapshots) is not affected — it is already correct.
  const flowTxns = transferIds && transferIds.size > 0
    ? transactions.filter((t) => !transferIds.has(t.transaction_id))
    : transactions;

  const totalIncome = flowTxns
    .filter((t) => parseFloat(t.amount) > 0)
    .reduce((sum, t) => sum + parseFloat(t.amount), 0);

  const totalExpenses = Math.abs(
    flowTxns
      .filter((t) => parseFloat(t.amount) < 0)
      .reduce((sum, t) => sum + parseFloat(t.amount), 0)
  );

  const netBalance = getNetWorth(transactions);

  return {
    totalIncome,
    totalExpenses,
    netBalance,
  };
}
