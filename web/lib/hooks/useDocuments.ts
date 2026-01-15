import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/lib/api-client';

export function useDocuments() {
  return useQuery({
    queryKey: ['documents'],
    queryFn: () => apiClient.listDocuments(),
  });
}
