import { useMutation, useQueryClient } from '@tanstack/react-query';
import { unmergeMerchant } from '@/features/merchants/services/merchantsApi';

export function useUnmergeMerchant() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (merchantId: string) => unmergeMerchant(merchantId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['merchants'] });
      queryClient.invalidateQueries({ queryKey: ['transactions'] });
    },
  });
}
