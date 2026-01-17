'use client';

import type { Account, Institution } from '@/shared/types/api';
import { useAccountScope } from '@/shared/account-scope/context';

export type AccountScopeSelectProps = {
  institutions?: Institution[];
  accounts?: Account[];
};

export function AccountScopeSelect({ institutions = [], accounts = [] }: AccountScopeSelectProps) {
  const { scope, setMode, setInstitutionId, setAccountId } = useAccountScope();

  const hasInstitutions = institutions.length > 0;
  const hasAccounts = accounts.length > 0;

  return (
    <div className="flex flex-wrap items-center gap-2">
      <select
        value={scope.mode}
        onChange={(e) => setMode(e.target.value as typeof scope.mode)}
        className="px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-slate-900/10"
      >
        <option value="all">All accounts</option>
        <option value="institution">Institution</option>
        <option value="account">Account</option>
      </select>

      {scope.mode === 'institution' && (
        <select
          value={scope.institutionId ?? ''}
          onChange={(e) => setInstitutionId(e.target.value || null)}
          className="px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-slate-900/10"
          disabled={!hasInstitutions}
        >
          <option value="">Select institution</option>
          {institutions.map((institution) => (
            <option key={institution.institution_id} value={institution.institution_id}>
              {institution.name}
            </option>
          ))}
        </select>
      )}

      {scope.mode === 'account' && (
        <select
          value={scope.accountId ?? ''}
          onChange={(e) => setAccountId(e.target.value || null)}
          className="px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-slate-900/10"
          disabled={!hasAccounts}
        >
          <option value="">Select account</option>
          {accounts.map((account) => (
            <option key={account.account_id} value={account.account_id}>
              {account.account_name || account.account_number || account.account_id}
            </option>
          ))}
        </select>
      )}
    </div>
  );
}
