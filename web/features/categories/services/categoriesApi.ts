import { apiClient, type Category } from '@/lib/api-client';

export async function listCategories(): Promise<Category[]> {
  return apiClient.listCategories();
}
