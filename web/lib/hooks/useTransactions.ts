import { useQuery } from '@tanstack/react-query';
import { listTransactions } from '@/features/transactions/services/transactionsApi';

export function useTransactions() {
  return useQuery({
    queryKey: ['transactions'],
    queryFn: () => listTransactions(),
  });
}
