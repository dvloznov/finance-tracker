'use client';

import { useQuery } from '@tanstack/react-query';
import { listAccounts } from '@/features/accounts/services/accountsApi';
import { listInstitutions } from '@/features/institutions/services/institutionsApi';

export function useAccountOptions() {
  const { data: institutions = [], isLoading: institutionsLoading } = useQuery({
    queryKey: ['institutions'],
    queryFn: () => listInstitutions(),
  });

  const { data: accounts = [], isLoading: accountsLoading } = useQuery({
    queryKey: ['accounts'],
    queryFn: () => listAccounts(),
  });

  return {
    institutions,
    accounts,
    isLoading: institutionsLoading || accountsLoading,
  };
}
