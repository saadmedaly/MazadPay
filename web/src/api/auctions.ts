import client from './client'
import type { APIResponse, PaginatedResponse, Auction } from '@/types/api'

export interface AuctionFilters {
  status?: string
  category_id?: number
  q?: string
  page?: number
  per_page?: number
}

export interface AuctionPayload {
  category_id: number
  location_id?: number
  title_ar: string
  title_fr?: string
  title_en?: string
  description_ar?: string
  description_fr?: string
  description_en?: string
  start_price: number
  min_increment: number
  insurance_amount?: number
  buy_now_price?: number
  start_time?: string
  end_time: string
  phone_contact?: string
  images: string[]
  item_details?: Record<string, unknown>
}

export async function fetchAuctions(filters: AuctionFilters): Promise<{ data: Auction[]; total: number }> {
  const { data } = await client.get<PaginatedResponse<Auction>>(
    '/v1/api/admin/auctions', { params: filters }
  )
  return { data: data.data, total: data.meta.total }
}

export async function fetchAuction(id: string): Promise<Auction> {
  const { data } = await client.get<APIResponse<{ auction: Auction; images: string[] }>>(`/v1/api/auctions/${id}`)
  return { ...data.data.auction, images: data.data.images }
}

export async function validateAuction(payload: {
  id: string
  approve: boolean
  reason?: string
}): Promise<void> {
  await client.put(`/v1/api/admin/auctions/${payload.id}/validate`, payload)
}

export async function createAuction(payload: AuctionPayload): Promise<Auction> {
  console.log('%c[createAuction] Sending payload:', 'color: #fbbf24; font-weight: bold', JSON.parse(JSON.stringify(payload)))
  const { data } = await client.post<APIResponse<Auction>>('/v1/api/auctions', payload)
  console.log('%c[createAuction] Response:', 'color: #34d399; font-weight: bold', data)
  return data.data
}

export async function updateAuction(id: string, payload: AuctionPayload): Promise<void> {
  console.log('%c[updateAuction] Sending payload for', 'color: #fbbf24; font-weight: bold', id, JSON.parse(JSON.stringify(payload)))
  await client.put(`/v1/api/admin/auctions/${id}`, payload)
}

export async function deleteAuction(id: string): Promise<void> {
  await client.delete(`/v1/api/admin/auctions/${id}`)
}
