export type AccountScopeMode = 'all' | 'institution' | 'account';

export type AccountScope = {
  mode: AccountScopeMode;
  institutionId?: string | null;
  accountId?: string | null;
};
