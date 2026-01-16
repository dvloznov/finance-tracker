import { format } from 'date-fns';
import type { TransactionVM } from '@/features/transactions/types';

export type MonthlyTotals = {
  month: string;
  income: number;
  expenses: number;
};

export function getMonthlyTotals(transactions?: TransactionVM[] | null): MonthlyTotals[] {
  if (!transactions || !Array.isArray(transactions)) return [];

  const monthlyMap = new Map<string, { income: number; expenses: number }>();

  transactions.forEach((txn) => {
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
      data.income += amount;
    } else {
      data.expenses += Math.abs(amount);
    }
  });

  return Array.from(monthlyMap.entries())
    .map(([month, data]) => ({ month, ...data }))
    .slice(-6);
}
