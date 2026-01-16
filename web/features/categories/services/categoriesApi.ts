import { apiClient } from '@/lib/api-client';
import type { Category } from '@/shared/types/api';

export async function listCategories(): Promise<Category[]> {
  return apiClient.listCategories();
}
