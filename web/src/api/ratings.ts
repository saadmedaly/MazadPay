import client from './client'
import type { APIResponse, AppRating } from '@/types/api'

export interface RatingsParams {
  page?: number
  per_page?: number
  rating?: number
}

export async function getRatings(params?: RatingsParams): Promise<{ data: AppRating[]; total: number }> {
  const { data } = await client.get<APIResponse<AppRating[]>>('/v1/api/admin/ratings', { params })
  return { data: data.data, total: data.data.length }
}

export async function deleteRating(id: string): Promise<void> {
  await client.delete(`/v1/api/admin/ratings/${id}`)
}