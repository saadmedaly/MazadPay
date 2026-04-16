import client from './client'
import type { APIResponse, PaginatedResponse, Auction } from '@/types/api'

export interface AuctionFilters {
  status?: string
  category_id?: number
  q?: string
  page?: number
  per_page?: number
}

export async function fetchAuctions(filters: AuctionFilters): Promise<{ data: Auction[]; total: number }> {
  const { data } = await client.get<PaginatedResponse<Auction>>(
    '/v1/api/admin/auctions', { params: filters }
  )
  return { data: data.data, total: data.meta.total }
}

export async function fetchAuction(id: string): Promise<Auction> {
  const { data } = await client.get<APIResponse<Auction>>(`/v1/api/auctions/${id}`)
  return data.data
}

export async function validateAuction(payload: {
  id: string
  approve: boolean
  reason?: string
}): Promise<void> {
  await client.put(`/v1/api/admin/auctions/${payload.id}/validate`, payload)
}
