import { useQuery } from '@tanstack/react-query';
import { toDocumentVM } from '@/features/documents/adapters/documentVm';
import { listDocuments } from '@/features/documents/services/documentsApi';

export function useDocuments() {
  return useQuery({
    queryKey: ['documents'],
    queryFn: async () => {
      const documents = await listDocuments();
      return documents.map(toDocumentVM);
    },
  });
}
