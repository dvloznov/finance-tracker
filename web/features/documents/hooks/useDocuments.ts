import { useQuery } from '@tanstack/react-query';
import { toDocumentVM } from '@/features/documents/adapters/documentVm';
import { listDocuments } from '@/features/documents/services/documentsApi';
import type { AccountScope } from '@/shared/account-scope/types';

type UseDocumentsParams = {
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

export function useDocuments({ scope }: UseDocumentsParams = {}) {
  const scopeParams = getScopeParams(scope);

  return useQuery({
    queryKey: ['documents', scope?.mode ?? 'all', scope?.institutionId ?? null, scope?.accountId ?? null],
    queryFn: async () => {
      const documents = await listDocuments(scopeParams);
      return documents.map(toDocumentVM);
    },
  });
}
