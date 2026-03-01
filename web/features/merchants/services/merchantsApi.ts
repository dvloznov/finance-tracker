import { apiClient } from '@/lib/api-client';
import type { Merchant } from '@/shared/types/api';

export async function listMerchants(): Promise<Merchant[]> {
  return apiClient.listMerchants();
}

export async function updateMerchantCategory(merchantId: string, categoryId: string): Promise<void> {
  return apiClient.updateMerchantCategory(merchantId, categoryId);
}
