import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as api from '@/api/auctions'
import type { AuctionFilters, AuctionPayload } from '@/api/auctions'
import toast from 'react-hot-toast'

export const auctionKeys = {
  all: ['auctions'] as const,
  list: (f: AuctionFilters) => [...auctionKeys.all, f] as const,
}

export function useAuctions(filters: AuctionFilters) {
  return useQuery({
    queryKey: auctionKeys.list(filters),
    queryFn: () => api.fetchAuctions(filters),
  })
}

export function useAuction(id: string) {
  return useQuery({
    queryKey: [...auctionKeys.all, id],
    queryFn: () => api.fetchAuction(id),
    enabled: !!id,
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

export function useCreateAuction() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: api.createAuction,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: auctionKeys.all })
      toast.success('تم إنشاء المزاد بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}

export function useUpdateAuction() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: AuctionPayload }) =>
      api.updateAuction(id, payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: auctionKeys.all })
      toast.success('تم تعديل المزاد بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}

export function useDeleteAuction() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.deleteAuction(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: auctionKeys.all })
      toast.success('تم حذف المزاد بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}
