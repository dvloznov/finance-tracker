import { useQuery } from '@tanstack/react-query';
import { toTransactionVM } from '@/features/transactions/adapters/transactionVm';
import { listTransactions } from '@/features/transactions/services/transactionsApi';
import type { AccountScope } from '@/shared/account-scope/types';

type UseTransactionsParams = {
  scope?: AccountScope;
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

export function useTransactions({ scope }: UseTransactionsParams = {}) {
  const scopeParams = getScopeParams(scope);

  return useQuery({
    queryKey: ['transactions', scope?.mode ?? 'all', scope?.institutionId ?? null, scope?.accountId ?? null],
    queryFn: async () => {
      const transactions = await listTransactions(scopeParams);
      return transactions.map(toTransactionVM);
    },
  });
}
