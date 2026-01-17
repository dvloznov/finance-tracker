import { apiClient } from '@/lib/api-client';
import type { Institution } from '@/shared/types/api';

export async function listInstitutions(): Promise<Institution[]> {
  return apiClient.listInstitutions();
}
