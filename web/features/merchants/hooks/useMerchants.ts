import { useQuery } from '@tanstack/react-query';
import { listMerchants } from '@/features/merchants/services/merchantsApi';

export function useMerchants() {
  return useQuery({
    queryKey: ['merchants'],
    queryFn: () => listMerchants(),
  });
}
