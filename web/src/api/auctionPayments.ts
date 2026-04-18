import client from './client'
import type { APIResponse, AuctionPayment } from '@/types/api'

export interface AuctionPaymentsParams {
  page?: number
  per_page?: number
  status?: string
}

export async function getAuctionPayments(params?: AuctionPaymentsParams): Promise<{ data: AuctionPayment[]; total: number }> {
  const { data } = await client.get<APIResponse<AuctionPayment[]>>('/v1/api/admin/auction-payments', { params })
  return { data: data.data, total: data.data.length }
}

export async function markAuctionPaymentAsPaid(id: string): Promise<void> {
  await client.put(`/v1/api/admin/auction-payments/${id}/pay`)
}