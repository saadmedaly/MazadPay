import client from './client'
import type { APIResponse, Location } from '@/types/api'

export async function getLocations(): Promise<Location[]> {
  const { data } = await client.get<APIResponse<Location[]>>('/v1/api/locations')
  return data.data
}

export async function createLocation(payload: {
  city_name_ar: string
  city_name_fr: string
  area_name_ar: string
  area_name_fr: string
}): Promise<Location> {
  const { data } = await client.post<APIResponse<Location>>('/v1/api/admin/locations', payload)
  return data.data
}

export async function updateLocation(id: number, payload: {
  city_name_ar: string
  city_name_fr: string
  area_name_ar: string
  area_name_fr: string
}): Promise<Location> {
  const { data } = await client.put<APIResponse<Location>>(`/v1/api/admin/locations/${id}`, payload)
  return data.data
}

export async function deleteLocation(id: number): Promise<void> {
  await client.delete(`/v1/api/admin/locations/${id}`)
}