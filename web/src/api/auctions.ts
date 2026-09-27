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

export interface UploadImagesResponse {
  message: string
  urls: string[]
  count: number
}

export async function fetchAuctions(filters: AuctionFilters): Promise<{ data: Auction[]; total: number }> {
  const { data } = await client.get<PaginatedResponse<Auction>>(
    '/v1/api/admin/auctions', { params: filters }
  )
  return { data: data.data, total: data.meta.total }
}

export async function fetchAuction(id: string): Promise<Auction> {
  const { data } = await client.get<APIResponse<{ auction: Auction; images: any[] }>>(`/v1/api/auctions/${id}`)
  const auction = data.data.auction
  // Merge images from dedicated array and handle potential strings or objects
  const images = (data.data.images || []).map((img: any) => img?.url || img)
  return { ...auction, images }
}

export async function validateAuction(payload: {
  id: string
  approve: boolean
  reason?: string
}): Promise<void> {
  await client.put(`/v1/api/admin/auctions/${payload.id}/validate`, payload)
}

// Customer #31: admin manual release of the auction WINNER's own insurance
// hold -- the one case the automatic non-winner refund (run at auction
// finalization) never covers. Takes only the auction id; the backend
// derives the winner and the refunded amount authoritatively, never from
// this client.
export async function refundWinnerInsurance(auctionId: string): Promise<{ transaction_id: string; amount: string }> {
  const { data } = await client.post<APIResponse<{ transaction_id: string; amount: string }>>(
    `/v1/api/admin/auctions/${auctionId}/refund-winner-insurance`
  )
  return data.data
}

// Customer #37: admin relist of an ended auction, reusing the Bug J relist
// primitive server-side. Takes only the auction id -- the new end_time is
// derived authoritatively on the backend from the auction's own original
// (start_time, end_time), never computed or supplied by this client.
export async function relistAuction(auctionId: string): Promise<{ id: string; status: string; end_time: string }> {
  const { data } = await client.post<APIResponse<{ id: string; status: string; end_time: string }>>(
    `/v1/api/admin/auctions/${auctionId}/relist`
  )
  return data.data
}

export async function createAuction(payload: AuctionPayload): Promise<Auction> {
  const { data } = await client.post<APIResponse<Auction>>('/v1/api/auctions', payload)
  return data.data
}

export async function updateAuction(id: string, payload: AuctionPayload): Promise<void> {
  await client.put(`/v1/api/admin/auctions/${id}`, payload)
}

export async function deleteAuction(id: string): Promise<void> {
  await client.delete(`/v1/api/admin/auctions/${id}`)
}

export async function uploadAuctionImages(auctionId: string, images: File[]): Promise<UploadImagesResponse> {
  const formData = new FormData()
  images.forEach((image) => {
    formData.append('images', image)
  })

  const { data } = await client.post<APIResponse<UploadImagesResponse>>(
    `/v1/api/admin/auctions/${auctionId}/images`,
    formData,
    {
      timeout: 60_000,
      headers: {
        'Content-Type': undefined, // Let browser set multipart/form-data with boundary
      },
    }
  )
  return data.data
}
