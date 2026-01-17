import { apiClient } from '@/lib/api-client';
import type { Account } from '@/shared/types/api';

export async function listAccounts(): Promise<Account[]> {
  return apiClient.listAccounts();
}
