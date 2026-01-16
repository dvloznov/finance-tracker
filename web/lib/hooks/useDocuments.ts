import { useQuery } from '@tanstack/react-query';
import { listDocuments } from '@/features/documents/services/documentsApi';

export function useDocuments() {
  return useQuery({
    queryKey: ['documents'],
    queryFn: () => listDocuments(),
  });
}
