import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import * as api from '@/api/auctions'

export const auctionKeys = {
  all:   ['auctions'] as const,
  list:  (f: api.AuctionFilters) => [...auctionKeys.all, 'list', f] as const,
  byId:  (id: string) => [...auctionKeys.all, id] as const,
}

export function useAuctions(filters: api.AuctionFilters) {
  return useQuery({
    queryKey: auctionKeys.list(filters),
    queryFn:  () => api.fetchAuctions(filters),
    placeholderData: (prev) => prev,
  })
}

export function useAuction(id: string) {
  return useQuery({
    queryKey: auctionKeys.byId(id),
    queryFn:  () => api.fetchAuction(id),
    enabled:  !!id,
  })
}

export function useValidateAuction() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: api.validateAuction,
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: auctionKeys.all })
      qc.invalidateQueries({ queryKey: ['dashboard'] })
      toast.success(vars.approve ? 'تمت الموافقة على المزاد' : 'تم رفض المزاد')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}
