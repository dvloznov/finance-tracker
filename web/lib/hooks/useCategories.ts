import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/lib/api-client';

export function useCategories() {
  return useQuery({
    queryKey: ['categories'],
    queryFn: () => apiClient.listCategories(),
  });
}
