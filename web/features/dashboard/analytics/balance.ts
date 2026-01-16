import { format } from 'date-fns';
import type { TransactionVM } from '@/features/transactions/types';

export type BalanceSeries = Array<{ id: string; data: Array<{ x: string; y: number }> }>;

export function getBalanceSeries(transactions?: TransactionVM[] | null): BalanceSeries {
  if (!transactions || !Array.isArray(transactions)) return [];

  const sorted = [...transactions]
    .filter((txn) => txn.transaction_date)
    .sort((a, b) => {
      const dateA = new Date(a.transaction_date);
      const dateB = new Date(b.transaction_date);
      return dateA.getTime() - dateB.getTime();
    });

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

      balanceHistory.unshift({
        x: format(date, 'MMM dd'),
        y: workingBalance,
      });
    }
  } else {
    let runningBalance = 0;
    for (const txn of sorted) {
      const date = new Date(txn.transaction_date);
      if (isNaN(date.getTime())) continue;

      runningBalance += parseFloat(txn.amount);
      balanceHistory.push({
        x: format(date, 'MMM dd'),
        y: runningBalance,
      });
    }
  }

  if (balanceHistory.length > 30) {
    const step = Math.ceil(balanceHistory.length / 30);
    return [{
      id: 'balance',
      data: balanceHistory.filter((_, i) => i % step === 0),
    }];
  }

  return [{
    id: 'balance',
    data: balanceHistory,
  }];
}
