import { apiClient } from '@/lib/api-client';
import type { Merchant } from '@/shared/types/api';

export async function listMerchants(params?: { start_date?: string; end_date?: string }): Promise<Merchant[]> {
  return apiClient.listMerchants(params);
}

export async function updateMerchantCategory(merchantId: string, categoryId: string): Promise<void> {
  return apiClient.updateMerchantCategory(merchantId, categoryId);
}

export async function mergeMerchant(merchantId: string, canonicalMerchantId: string): Promise<void> {
  return apiClient.mergeMerchant(merchantId, canonicalMerchantId);
}

export async function unmergeMerchant(merchantId: string): Promise<void> {
  return apiClient.unmergeMerchant(merchantId);
}
