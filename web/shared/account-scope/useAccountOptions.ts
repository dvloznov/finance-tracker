'use client';

import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { listDocuments } from '@/features/documents/services/documentsApi';
import type { Account, Institution } from '@/shared/types/api';

function buildInstitutions(documents: Array<{ institution_id?: string | null }>): Institution[] {
  const map = new Map<string, Institution>();
  documents.forEach((doc) => {
    if (!doc.institution_id) return;
    if (!map.has(doc.institution_id)) {
      map.set(doc.institution_id, {
        institution_id: doc.institution_id,
        name: doc.institution_id,
      });
    }
  });
  return Array.from(map.values());
}

function buildAccounts(documents: Array<{ account_id?: string | null; institution_id?: string | null }>): Account[] {
  const map = new Map<string, Account>();
  documents.forEach((doc) => {
    if (!doc.account_id) return;
    if (!map.has(doc.account_id)) {
      map.set(doc.account_id, {
        account_id: doc.account_id,
        institution_id: doc.institution_id ?? undefined,
        account_name: doc.account_id,
      });
    }
  });
  return Array.from(map.values());
}

export function useAccountOptions() {
  const { data: documents = [], isLoading } = useQuery({
    queryKey: ['documents', 'options'],
    queryFn: () => listDocuments(),
  });

  const institutions = useMemo(() => buildInstitutions(documents), [documents]);
  const accounts = useMemo(() => buildAccounts(documents), [documents]);

  return {
    institutions,
    accounts,
    isLoading,
  };
}
