import { format } from 'date-fns';
import type { TransactionVM } from '@/features/transactions/types';

export type BalanceSeries = Array<{ id: string; data: Array<{ x: string; y: number }> }>;

/**
 * Compute the signed balance contribution for a balance_after value.
 * Credit cards are liabilities: their outstanding balance is subtracted from net worth.
 */
function signedBalance(rawBalance: number, accountType: string | undefined): number {
  return accountType === 'CREDIT_CARD' ? -rawBalance : rawBalance;
}

/**
 * Build a balance-over-time series for the chart.
 *
 * Single-account view:
 *   - For current/savings: y = running account balance (positive = asset).
 *   - For credit card: y = amount owed (positive on chart represents debt).
 *     The sign is kept as stored so the chart label reads naturally ("you owe £X").
 *
 * All-accounts / mixed view:
 *   - Each account contributes its signed balance to a combined net worth figure.
 *   - Credit card balances are negated before summing (liability reduces net worth).
 */
export function getBalanceSeries(transactions?: TransactionVM[] | null): BalanceSeries {
  if (!transactions || !Array.isArray(transactions)) return [];

  const sorted = [...transactions]
    .filter((txn) => txn.transaction_date)
    .sort((a, b) => new Date(a.transaction_date).getTime() - new Date(b.transaction_date).getTime());

  // Determine whether all transactions are from a single account type.
  const accountTypes = new Set(transactions.map((t) => t.account_type ?? 'CURRENT'));
  const allCreditCard = accountTypes.size === 1 && accountTypes.has('CREDIT_CARD');
  const isMixed = accountTypes.size > 1 || (accountTypes.size === 1 && !allCreditCard && accountTypes.has('CURRENT'));

  // Group into per-account buckets so we can track running balances separately
  // before combining them into a net worth figure for mixed views.
  const accountIds = [...new Set(transactions.map((t) => t.account_id ?? '__unknown__'))];
  const isSingleAccount = accountIds.length === 1;

  if (isSingleAccount) {
    // Simple single-account path (same as before, but label-aware for credit cards).
    return buildSingleAccountSeries(sorted, allCreditCard);
  }

  // Multi-account: combine per-account balances into a net worth series keyed by date.
  return buildMultiAccountSeries(sorted);
}

function buildSingleAccountSeries(sorted: TransactionVM[], isCreditCard: boolean): BalanceSeries {
  // Use ISO date string as the key so multiple transactions on the same day
  // collapse into one point (the last balance of that day wins).
  const dayMap = new Map<string, number>();
  const txnsWithBalance = sorted.filter((txn) => txn.balance_after);

  if (txnsWithBalance.length > 0) {
    const lastKnownBalance = parseFloat(txnsWithBalance[txnsWithBalance.length - 1].balance_after!);
    let workingBalance = lastKnownBalance;

    for (let i = sorted.length - 1; i >= 0; i--) {
      const txn = sorted[i];
      const date = new Date(txn.transaction_date);
      if (isNaN(date.getTime())) continue;

      if (txn.balance_after) {
        workingBalance = parseFloat(txn.balance_after);
      } else {
        workingBalance -= parseFloat(txn.amount);
      }

      dayMap.set(txn.transaction_date.slice(0, 10), workingBalance);
    }
  } else {
    let running = 0;
    for (const txn of sorted) {
      const date = new Date(txn.transaction_date);
      if (isNaN(date.getTime())) continue;
      running += parseFloat(txn.amount);
      dayMap.set(txn.transaction_date.slice(0, 10), isCreditCard ? -running : running);
    }
  }

  const points = [...dayMap.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([iso, y]) => ({ x: format(new Date(iso), 'MMM d'), y }));

  return [{ id: isCreditCard ? 'amount owed' : 'balance', data: decimate(points) }];
}

function buildMultiAccountSeries(sorted: TransactionVM[]): BalanceSeries {
  const latestBalance = new Map<string, number>();
  // Use ISO date as key so same-day transactions collapse to one net-worth point.
  const dayMap = new Map<string, number>();

  for (const txn of sorted) {
    const date = new Date(txn.transaction_date);
    if (isNaN(date.getTime())) continue;

    const accountId = txn.account_id ?? '__unknown__';
    const accountType = txn.account_type;

    if (txn.balance_after) {
      const raw = parseFloat(txn.balance_after);
      latestBalance.set(accountId, signedBalance(raw, accountType));
    } else {
      const prev = latestBalance.get(accountId) ?? 0;
      latestBalance.set(accountId, prev + parseFloat(txn.amount));
    }

    let netWorth = 0;
    for (const b of latestBalance.values()) netWorth += b;

    dayMap.set(txn.transaction_date.slice(0, 10), netWorth);
  }

  const points = [...dayMap.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([iso, y]) => ({ x: format(new Date(iso), 'MMM d'), y }));

  return [{ id: 'net worth', data: decimate(points) }];
}

/**
 * Thin the series to at most maxPoints, always keeping the first and last points
 * so the chart always shows the true earliest and latest dates.
 */
function decimate(
  points: Array<{ x: string; y: number }>,
  maxPoints = 24
): Array<{ x: string; y: number }> {
  if (points.length <= maxPoints) return points;
  const step = Math.ceil((points.length - 2) / (maxPoints - 2));
  const result: Array<{ x: string; y: number }> = [points[0]];
  for (let i = 1; i < points.length - 1; i++) {
    if ((i - 1) % step === 0) result.push(points[i]);
  }
  result.push(points[points.length - 1]);
  return result;
}
