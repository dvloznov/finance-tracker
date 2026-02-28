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
  const balanceHistory: Array<{ x: string; y: number }> = [];
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

      balanceHistory.unshift({ x: format(date, 'MMM dd'), y: workingBalance });
    }
  } else {
    // No balance_after data — reconstruct from transaction flow.
    // Credit cards start with 0 owed; spend makes it more negative on the net-worth scale,
    // but for the single-account chart we show amount owed (positive).
    let running = 0;
    for (const txn of sorted) {
      const date = new Date(txn.transaction_date);
      if (isNaN(date.getTime())) continue;
      running += parseFloat(txn.amount);
      // For credit cards: amount owed = -running (spend is negative after prompt fix).
      balanceHistory.push({ x: format(date, 'MMM dd'), y: isCreditCard ? -running : running });
    }
  }

  const decimated = decimate(balanceHistory);
  return [{ id: isCreditCard ? 'amount owed' : 'balance', data: decimated }];
}

function buildMultiAccountSeries(sorted: TransactionVM[]): BalanceSeries {
  // Track the latest known signed balance per account.
  const latestBalance = new Map<string, number>();

  // Collect all unique dates (by day string) in order.
  const datePoints: Array<{ x: string; netWorth: number }> = [];

  for (const txn of sorted) {
    const date = new Date(txn.transaction_date);
    if (isNaN(date.getTime())) continue;

    const accountId = txn.account_id ?? '__unknown__';
    const accountType = txn.account_type;

    if (txn.balance_after) {
      const raw = parseFloat(txn.balance_after);
      latestBalance.set(accountId, signedBalance(raw, accountType));
    } else {
      // Adjust the known balance for this account by the transaction amount.
      // For credit cards, a negative amount (spend) increases the amount owed,
      // which means it *decreases* net worth — already handled by signedBalance
      // being negative for CC balances, and spend being negative after the prompt fix.
      const prev = latestBalance.get(accountId) ?? 0;
      latestBalance.set(accountId, prev + parseFloat(txn.amount));
    }

    // Sum all known account balances for this point in time.
    let netWorth = 0;
    for (const b of latestBalance.values()) netWorth += b;

    datePoints.push({ x: format(date, 'MMM dd'), netWorth });
  }

  const data = decimate(datePoints.map((p) => ({ x: p.x, y: p.netWorth })));
  return [{ id: 'net worth', data }];
}

function decimate(points: Array<{ x: string; y: number }>): Array<{ x: string; y: number }> {
  if (points.length <= 30) return points;
  const step = Math.ceil(points.length / 30);
  return points.filter((_, i) => i % step === 0);
}
