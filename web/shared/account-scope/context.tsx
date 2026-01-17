'use client';

import { createContext, useContext, useMemo, useState, type ReactNode } from 'react';
import type { AccountScope, AccountScopeMode } from '@/shared/account-scope/types';

export type AccountScopeContextValue = {
  scope: AccountScope;
  setMode: (mode: AccountScopeMode) => void;
  setInstitutionId: (institutionId: string | null) => void;
  setAccountId: (accountId: string | null) => void;
};

const AccountScopeContext = createContext<AccountScopeContextValue | null>(null);

export function AccountScopeProvider({ children }: { children: ReactNode }) {
  const [scope, setScope] = useState<AccountScope>({
    mode: 'all',
    institutionId: null,
    accountId: null,
  });

  const setMode = (mode: AccountScopeMode) => {
    setScope((prev) => ({
      mode,
      institutionId: mode === 'institution' ? prev.institutionId : null,
      accountId: mode === 'account' ? prev.accountId : null,
    }));
  };

  const setInstitutionId = (institutionId: string | null) => {
    setScope((prev) => ({
      ...prev,
      institutionId,
      accountId: prev.mode === 'account' ? prev.accountId : null,
    }));
  };

  const setAccountId = (accountId: string | null) => {
    setScope((prev) => ({
      ...prev,
      accountId,
    }));
  };

  const value = useMemo(
    () => ({
      scope,
      setMode,
      setInstitutionId,
      setAccountId,
    }),
    [scope]
  );

  return (
    <AccountScopeContext.Provider value={value}>
      {children}
    </AccountScopeContext.Provider>
  );
}

export function useAccountScope() {
  const context = useContext(AccountScopeContext);
  if (!context) {
    throw new Error('useAccountScope must be used within an AccountScopeProvider');
  }
  return context;
}
