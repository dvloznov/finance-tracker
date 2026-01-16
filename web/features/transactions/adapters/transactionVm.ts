import type { Transaction } from '@/lib/api-client';
import type { TransactionVM } from '@/features/transactions/types';

export function toTransactionVM(transaction: Transaction): TransactionVM {
  return transaction;
}
