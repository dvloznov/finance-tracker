import { useQuery } from '@tanstack/react-query';
import { listMerchants } from '@/features/merchants/services/merchantsApi';

type UseMerchantsParams = {
  start_date?: string;
  end_date?: string;
};

export function useMerchants(params?: UseMerchantsParams) {
  return useQuery({
    queryKey: ['merchants', params?.start_date ?? null, params?.end_date ?? null],
    queryFn: () => listMerchants(params),
  });
}
