import { useMutation, useQueryClient } from '@tanstack/react-query';
import { updateMerchantCategory } from '@/features/merchants/services/merchantsApi';

export function useUpdateMerchantCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ merchantId, categoryId }: { merchantId: string; categoryId: string }) =>
      updateMerchantCategory(merchantId, categoryId),
    onSuccess: () => {
      // Invalidate merchants so the list reflects the new category name immediately.
      // Invalidate transactions so analytics and dashboards re-compute with the new category.
      queryClient.invalidateQueries({ queryKey: ['merchants'] });
      queryClient.invalidateQueries({ queryKey: ['transactions'] });
    },
  });
}
