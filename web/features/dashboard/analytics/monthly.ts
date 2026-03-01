import { format, parse } from 'date-fns';
import type { TransactionVM } from '@/features/transactions/types';

export type MonthlyTotals = {
  month: string;
  income: number;
  expenses: number;
};

export type BarSelection = { month: string; type: 'income' | 'expenses' };

/**
 * Returns true if a credit card transaction is a repayment/payment (money used to
 * pay off the card balance). These are transfers — not income — and should be
 * excluded from the income bucket when showing combined account totals.
 */
function isCreditCardRepayment(txn: TransactionVM): boolean {
  if (txn.account_type !== 'CREDIT_CARD') return false;
  const type = (txn.transaction_type ?? '').toUpperCase();
  return (
    type.includes('PAYMENT') ||
    type.includes('PAYMENT RECEIVED') ||
    txn.raw_description?.toUpperCase().includes('PAYMENT RECEIVED')
  );
}

export function getMonthlyTotals(
  transactions?: TransactionVM[] | null,
  transferIds?: Set<string>
): MonthlyTotals[] {
  if (!transactions || !Array.isArray(transactions)) return [];

  // Exclude both legs of detected transfer pairs from monthly flow metrics.
  const flowTxns = transferIds && transferIds.size > 0
    ? transactions.filter((t) => !transferIds.has(t.transaction_id))
    : transactions;

  const monthlyMap = new Map<string, { income: number; expenses: number }>();

  flowTxns.forEach((txn) => {
    const dateStr = typeof txn.transaction_date === 'string'
      ? txn.transaction_date
      : String(txn.transaction_date || '');

    if (!dateStr) return;

    const date = new Date(dateStr);
    if (isNaN(date.getTime())) return;

    const monthKey = format(date, 'MMM yyyy');
    const amount = parseFloat(txn.amount);

    if (!monthlyMap.has(monthKey)) {
      monthlyMap.set(monthKey, { income: 0, expenses: 0 });
    }

    const data = monthlyMap.get(monthKey)!;

    if (amount > 0) {
      // Positive amount on a credit card = payment received toward the balance owed.
      // This is a liability reduction (transfer), not real income — skip it.
      if (!isCreditCardRepayment(txn)) {
        data.income += amount;
      }
    } else {
      data.expenses += Math.abs(amount);
    }
  });

  return Array.from(monthlyMap.entries())
    .map(([month, data]) => ({ month, ...data }))
    .slice(-6);
}

/**
 * Filters transactions to those that contribute to a specific bar (month + type).
 * Uses the same logic as getMonthlyTotals: excludes transfers, and for income
 * excludes credit card repayments.
 */
export function filterTransactionsByBar(
  transactions: TransactionVM[],
  selection: BarSelection,
  transferIds?: Set<string>
): TransactionVM[] {
  if (!transactions?.length) return [];

  const flowTxns = transferIds?.size
    ? transactions.filter((t) => !transferIds.has(t.transaction_id))
    : transactions;

  let targetMonth: Date;
  try {
    targetMonth = parse(selection.month, 'MMM yyyy', new Date());
  } catch {
    return [];
  }
  const targetMonthNum = targetMonth.getMonth();
  const targetYear = targetMonth.getFullYear();

  return flowTxns.filter((txn) => {
    const dateStr = typeof txn.transaction_date === 'string'
      ? txn.transaction_date
      : String(txn.transaction_date || '');
    if (!dateStr) return false;

    const date = new Date(dateStr);
    if (isNaN(date.getTime())) return false;
    if (date.getMonth() !== targetMonthNum || date.getFullYear() !== targetYear) return false;

    const amount = parseFloat(txn.amount);
    if (selection.type === 'income') {
      return amount > 0 && !isCreditCardRepayment(txn);
    }
    return amount < 0;
  });
}
