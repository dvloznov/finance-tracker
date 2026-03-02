'use client';

import { useMemo } from 'react';
import { AppNav } from '@/shared/ui/AppNav';
import { cardClass } from '@/lib/ui';
import { useAccountOptions } from '@/shared/account-scope/useAccountOptions';
import type { Account, Institution } from '@/shared/types/api';

function formatAccountType(type?: string): string {
  if (!type) return '—';
  const map: Record<string, string> = {
    CURRENT: 'Current',
    SAVINGS: 'Savings',
    CREDIT_CARD: 'Credit Card',
  };
  return map[type] ?? type;
}

export default function AccountsPage() {
  const { institutions = [], accounts = [], isLoading } = useAccountOptions();

  const accountsByInstitution = useMemo(() => {
    const byInst = new Map<string, Account[]>();
    const unassigned: Account[] = [];

    for (const acc of accounts) {
      const instId = acc.institution_id ?? '__unassigned__';
      if (instId === '__unassigned__') {
        unassigned.push(acc);
      } else {
        if (!byInst.has(instId)) byInst.set(instId, []);
        byInst.get(instId)!.push(acc);
      }
    }

    return { byInst, unassigned };
  }, [accounts]);

  const institutionOrder = useMemo(() => {
    const order: Institution[] = [...institutions].sort((a, b) =>
      (a.name ?? '').localeCompare(b.name ?? '')
    );
    return order;
  }, [institutions]);

  return (
    <div className="min-h-screen bg-slate-50">
      <AppNav active="accounts" />

      <main className="container mx-auto px-6 py-8">
        <div className="space-y-6">
          <div className="space-y-1">
            <h1 className="text-3xl font-semibold tracking-tight text-slate-900">Accounts</h1>
            <p className="text-sm text-slate-600">
              Institutions and accounts linked to your statements
            </p>
          </div>

          {isLoading ? (
            <p className="text-sm text-slate-600">Loading...</p>
          ) : (
            <div className="space-y-6">
              {institutionOrder.map((inst) => {
                const instAccounts = accountsByInstitution.byInst.get(inst.institution_id) ?? [];

                return (
                  <div key={inst.institution_id} className={cardClass}>
                    <h2 className="text-base font-semibold text-slate-900 mb-4">{inst.name}</h2>
                    <div className="overflow-x-auto">
                      <table className="w-full">
                        {instAccounts.length > 0 && (
                        <thead>
                          <tr className="border-b border-slate-100">
                            <th className="text-left py-2 text-[11px] font-medium text-slate-500 uppercase tracking-wider">
                              Account
                            </th>
                            <th className="text-left py-2 text-[11px] font-medium text-slate-500 uppercase tracking-wider">
                              Type
                            </th>
                            <th className="text-left py-2 text-[11px] font-medium text-slate-500 uppercase tracking-wider">
                              Account number
                            </th>
                            <th className="text-left py-2 text-[11px] font-medium text-slate-500 uppercase tracking-wider">
                              Currency
                            </th>
                          </tr>
                        </thead>
                        )}
                        <tbody>
                          {instAccounts.length === 0 ? (
                            <tr>
                              <td colSpan={4} className="py-4 text-sm text-slate-500 text-center">
                                No accounts
                              </td>
                            </tr>
                          ) : (
                          instAccounts.map((acc) => (
                            <tr key={acc.account_id} className="border-b border-slate-50 last:border-0">
                              <td className="py-3 text-sm font-medium text-slate-900">
                                {acc.account_name || '—'}
                              </td>
                              <td className="py-3 text-sm text-slate-600">
                                {formatAccountType(acc.account_type)}
                              </td>
                              <td className="py-3 text-sm text-slate-600 tabular-nums">
                                {acc.account_number || '—'}
                              </td>
                              <td className="py-3 text-sm text-slate-600">
                                {acc.currency || '—'}
                              </td>
                            </tr>
                          )))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                );
              })}

              {accountsByInstitution.unassigned.length > 0 && (
                <div className={cardClass}>
                  <h2 className="text-base font-semibold text-slate-900 mb-4">Unassigned</h2>
                  <div className="overflow-x-auto">
                    <table className="w-full">
                      <thead>
                        <tr className="border-b border-slate-100">
                          <th className="text-left py-2 text-[11px] font-medium text-slate-500 uppercase tracking-wider">
                            Account
                          </th>
                          <th className="text-left py-2 text-[11px] font-medium text-slate-500 uppercase tracking-wider">
                            Type
                          </th>
                          <th className="text-left py-2 text-[11px] font-medium text-slate-500 uppercase tracking-wider">
                            Account number
                          </th>
                          <th className="text-left py-2 text-[11px] font-medium text-slate-500 uppercase tracking-wider">
                            Currency
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        {accountsByInstitution.unassigned.map((acc) => (
                          <tr key={acc.account_id} className="border-b border-slate-50 last:border-0">
                            <td className="py-3 text-sm font-medium text-slate-900">
                              {acc.account_name || '—'}
                            </td>
                            <td className="py-3 text-sm text-slate-600">
                              {formatAccountType(acc.account_type)}
                            </td>
                            <td className="py-3 text-sm text-slate-600 tabular-nums">
                              {acc.account_number || '—'}
                            </td>
                            <td className="py-3 text-sm text-slate-600">
                              {acc.currency || '—'}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

              {institutions.length === 0 && accounts.length === 0 && (
                <p className="text-sm text-slate-500 py-8 text-center">
                  No institutions or accounts yet. Upload bank statements to get started.
                </p>
              )}
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
