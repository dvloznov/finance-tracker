import { apiClient, type Transaction } from '@/lib/api-client';

type ListTransactionsParams = {
  start_date?: string;
  end_date?: string;
};

export async function listTransactions(params?: ListTransactionsParams): Promise<Transaction[]> {
  return apiClient.listTransactions(params);
}
