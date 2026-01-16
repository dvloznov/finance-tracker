import { useQuery } from '@tanstack/react-query';
import { listCategories } from '@/features/categories/services/categoriesApi';

export function useCategories() {
  return useQuery({
    queryKey: ['categories'],
    queryFn: () => listCategories(),
  });
}
