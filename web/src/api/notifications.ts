import client from './client'
import type { APIResponse, Notification } from '@/types/api'

export interface NotificationsParams {
  page?: number
  per_page?: number
  type?: string
  is_read?: boolean
}

export async function getNotifications(params?: NotificationsParams): Promise<{ data: Notification[]; total: number }> {
  const { data } = await client.get<APIResponse<Notification[]>>('/v1/api/admin/notifications', { params })
  return { data: data.data, total: data.data.length }
}

export async function markNotificationAsRead(id: string): Promise<void> {
  await client.put(`/v1/api/admin/notifications/${id}/read`)
}

export async function markAllNotificationsAsRead(): Promise<void> {
  await client.put('/v1/api/admin/notifications/read-all')
}