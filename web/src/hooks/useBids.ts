import { useQuery } from '@tanstack/react-query'
import { fetchBidHistory } from '@/api/bids'

export const bidKeys = {
  all: ['bids'] as const,
  history: (auctionId: string) => [...bidKeys.all, 'history', auctionId] as const,
}

export function useBidHistory(auctionId: string) {
  return useQuery({
    queryKey: bidKeys.history(auctionId),
    queryFn: () => fetchBidHistory(auctionId),
    enabled: !!auctionId,
  })
}
