import type { Category, Transaction } from '@/shared/types/api';
import type { TransactionVM } from '@/features/transactions/types';

type CategoryLookup = Map<string, Category>;

export function toTransactionVM(
  transaction: Transaction,
  categoryLookup?: CategoryLookup
): TransactionVM {
  if (!transaction.category_id || !categoryLookup) {
    return transaction;
  }

  const category = categoryLookup.get(transaction.category_id);
  if (!category) {
    return transaction;
  }

  return {
    ...transaction,
    category_name: transaction.category_name ?? category.category_name,
    subcategory_name: transaction.subcategory_name ?? category.subcategory_name,
  };
}
