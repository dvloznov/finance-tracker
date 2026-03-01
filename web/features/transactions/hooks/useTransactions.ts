import { useQuery } from '@tanstack/react-query';
import { toTransactionVM } from '@/features/transactions/adapters/transactionVm';
import { listTransactions } from '@/features/transactions/services/transactionsApi';
import { listCategories } from '@/features/categories/services/categoriesApi';
import type { AccountScope } from '@/shared/account-scope/types';

type UseTransactionsParams = {
  scope?: AccountScope;
  start_date?: string;
  end_date?: string;
};

function getScopeParams(scope?: AccountScope) {
  if (!scope || scope.mode === 'all') return {};
  if (scope.mode === 'institution' && scope.institutionId) {
    return { institution_id: scope.institutionId };
  }
  if (scope.mode === 'account' && scope.accountId) {
    return { account_id: scope.accountId };
  }
  return {};
}

export function useTransactions({ scope, start_date, end_date }: UseTransactionsParams = {}) {
  const scopeParams = getScopeParams(scope);
  const params = {
    ...scopeParams,
    ...(start_date != null && { start_date }),
    ...(end_date != null && { end_date }),
  };

  return useQuery({
    queryKey: ['transactions', scope?.mode ?? 'all', scope?.institutionId ?? null, scope?.accountId ?? null, start_date ?? null, end_date ?? null],
    queryFn: async () => {
      const [transactions, categories] = await Promise.all([
        listTransactions(params),
        listCategories(),
      ]);
      const categoryLookup = new Map(categories.map((category) => [category.category_id, category]));
      return transactions.map((transaction) => toTransactionVM(transaction, categoryLookup));
    },
  });
}
