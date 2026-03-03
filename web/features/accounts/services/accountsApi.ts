import { apiClient } from '@/lib/api-client';
import type { Account } from '@/shared/types/api';

export async function listAccounts(): Promise<Account[]> {
  return apiClient.listAccounts();
}

export async function createAccount(payload: {
  institution_id: string;
  account_name?: string;
  account_number?: string;
  sort_code?: string;
  iban?: string;
  currency?: string;
  account_type?: string;
}): Promise<{ account_id: string; account: Account }> {
  return apiClient.createAccount(payload);
}

export async function updateAccount(
  accountId: string,
  payload: Partial<{
    institution_id: string;
    account_name: string;
    account_number: string;
    sort_code: string;
    iban: string;
    currency: string;
    account_type: string;
  }>
): Promise<{ account_id: string; status: string }> {
  return apiClient.updateAccount(accountId, payload);
}

export async function deleteAccount(accountId: string): Promise<{ account_id: string; status: string }> {
  return apiClient.deleteAccount(accountId);
}
