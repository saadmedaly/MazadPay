import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getAuctionPayments, markAuctionPaymentAsPaid, type AuctionPaymentsParams } from '@/api/auctionPayments'

export function useAuctionPayments(params?: AuctionPaymentsParams) {
  return useQuery({
    queryKey: ['auctionPayments', params],
    queryFn: () => getAuctionPayments(params),
  })
}

export function useMarkAuctionPaymentAsPaid() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: markAuctionPaymentAsPaid,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['auctionPayments'] }),
  })
}