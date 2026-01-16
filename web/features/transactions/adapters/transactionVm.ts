import type { Transaction } from '@/shared/types/api';
import type { TransactionVM } from '@/features/transactions/types';

export function toTransactionVM(transaction: Transaction): TransactionVM {
  return transaction;
}
