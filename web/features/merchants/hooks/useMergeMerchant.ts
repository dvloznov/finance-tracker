import { useMutation, useQueryClient } from '@tanstack/react-query';
import { mergeMerchant } from '@/features/merchants/services/merchantsApi';

export function useMergeMerchant() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ merchantId, canonicalMerchantId }: { merchantId: string; canonicalMerchantId: string }) =>
      mergeMerchant(merchantId, canonicalMerchantId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['merchants'] });
      queryClient.invalidateQueries({ queryKey: ['transactions'] });
    },
  });
}
