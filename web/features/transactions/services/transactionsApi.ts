import { apiClient } from '@/lib/api-client';
import type { Transaction } from '@/shared/types/api';

type ListTransactionsParams = {
  start_date?: string;
  end_date?: string;
};

export async function listTransactions(params?: ListTransactionsParams): Promise<Transaction[]> {
  return apiClient.listTransactions(params);
}
