import client from './client'
import type { APIResponse, BlockedPhone } from '@/types/api'

export async function getBlockedPhones(): Promise<BlockedPhone[]> {
  const { data } = await client.get<APIResponse<BlockedPhone[]>>('/v1/api/admin/blocked-phones')
  return data.data
}

export async function blockPhone(payload: {
  phone: string
  reason: string
}): Promise<BlockedPhone> {
  const { data } = await client.post<APIResponse<BlockedPhone>>('/v1/api/admin/blocked-phones', payload)
  return data.data
}

export async function unblockPhone(phone: string): Promise<void> {
  await client.delete(`/v1/api/admin/blocked-phones/${encodeURIComponent(phone)}`)
}