import { useQuery } from '@tanstack/react-query';
import { toTransactionVM } from '@/features/transactions/adapters/transactionVm';
import { listTransactions } from '@/features/transactions/services/transactionsApi';

export function useTransactions() {
  return useQuery({
    queryKey: ['transactions'],
    queryFn: async () => {
      const transactions = await listTransactions();
      return transactions.map(toTransactionVM);
    },
  });
}
