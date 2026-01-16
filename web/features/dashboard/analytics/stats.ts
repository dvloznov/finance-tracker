import type { TransactionVM } from '@/features/transactions/types';

export type StatsSummary = {
  totalIncome: number;
  totalExpenses: number;
  netBalance: number;
};

export function getStatsSummary(transactions?: TransactionVM[] | null): StatsSummary | null {
  if (!transactions || !Array.isArray(transactions)) return null;

  const totalIncome = transactions
    .filter((t) => parseFloat(t.amount) > 0)
    .reduce((sum, t) => sum + parseFloat(t.amount), 0);

  const totalExpenses = Math.abs(
    transactions
      .filter((t) => parseFloat(t.amount) < 0)
      .reduce((sum, t) => sum + parseFloat(t.amount), 0)
  );

  const latestBalance = [...transactions]
    .filter((t) => t.transaction_date && t.balance_after)
    .sort((a, b) => {
      const dateA = new Date(a.transaction_date).getTime();
      const dateB = new Date(b.transaction_date).getTime();
      return dateB - dateA;
    })[0]?.balance_after;

  const netBalance = latestBalance ? parseFloat(latestBalance) : totalIncome - totalExpenses;

  return {
    totalIncome,
    totalExpenses,
    netBalance,
  };
}
