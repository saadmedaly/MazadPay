import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import * as api from '@/api/bankilyPayments'

export const bankilyKeys = {
  all:  ['bankily-payments'] as const,
  list: (f: api.BankilyFilters) => [...bankilyKeys.all, 'list', f] as const,
}

export function useBankilyPayments(filters: api.BankilyFilters) {
  return useQuery({
    queryKey: bankilyKeys.list(filters),
    queryFn:  () => api.fetchBankilyPayments(filters),
    refetchInterval: 30_000,
    placeholderData: (prev) => prev,
  })
}

export function useRecheckBankilyPayment() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.recheckBankilyPayment(id),
    onSuccess: (p) => {
      qc.invalidateQueries({ queryKey: bankilyKeys.all })
      qc.invalidateQueries({ queryKey: ['transactions'] })
      toast.success(`تم التحقق من بنكيلي: ${p.status}`)
    },
    onError: () => toast.error('تعذر التحقق من بنكيلي، حاول لاحقاً'),
  })
}
