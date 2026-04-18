import client from './client'
import type { APIResponse, SystemSetting } from '@/types/api'

export async function getSettings(): Promise<SystemSetting[]> {
  const { data } = await client.get<APIResponse<SystemSetting[]>>('/v1/api/admin/settings')
  return data.data
}

export async function updateSetting(payload: {
  key: string
  value: string
  type: string
}): Promise<void> {
  await client.put(`/v1/api/admin/settings/${payload.key}`, payload)
}