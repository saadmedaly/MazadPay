import client from './client'
import type { APIResponse, Country } from '@/types/api'

export async function getCountries(): Promise<Country[]> {
  const { data } = await client.get<APIResponse<Country[]>>('/v1/api/countries')
  return data.data
}

export async function createCountry(payload: {
  code: string
  name_ar: string
  name_fr: string
  name_en: string
  flag_emoji: string
}): Promise<Country> {
  const { data } = await client.post<APIResponse<Country>>('/v1/api/admin/countries', payload)
  return data.data
}

export async function updateCountry(id: number, payload: Partial<{
  code: string
  name_ar: string
  name_fr: string
  name_en: string
  flag_emoji: string
  is_active: boolean
}>): Promise<Country> {
  const { data } = await client.put<APIResponse<Country>>(`/v1/api/admin/countries/${id}`, payload)
  return data.data
}

export async function deleteCountry(id: number): Promise<void> {
  await client.delete(`/v1/api/admin/countries/${id}`)
}
