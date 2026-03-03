import { apiClient } from '@/lib/api-client';
import type { Institution } from '@/shared/types/api';

export async function listInstitutions(): Promise<Institution[]> {
  return apiClient.listInstitutions();
}

export async function createInstitution(name: string): Promise<{ institution_id: string; name: string }> {
  return apiClient.createInstitution(name);
}

export async function updateInstitution(
  institutionId: string,
  name: string
): Promise<{ institution_id: string; status: string }> {
  return apiClient.updateInstitution(institutionId, name);
}

export async function deleteInstitution(
  institutionId: string
): Promise<{ institution_id: string; status: string }> {
  return apiClient.deleteInstitution(institutionId);
}
