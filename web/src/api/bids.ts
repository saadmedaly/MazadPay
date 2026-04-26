import client from './client'
import type { APIResponse, Bid } from '@/types/api'

export async function fetchBidHistory(auctionId: string): Promise<Bid[]> {
  const { data } = await client.get<APIResponse<Bid[]>>(`/v1/api/auctions/${auctionId}/bids`)
  return data.data || []
}
