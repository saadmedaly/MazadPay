import client from './client'
import type { APIResponse, DashboardStats } from '@/types/api'

export async function fetchStats(): Promise<DashboardStats> {
  const { data } = await client.get<APIResponse<DashboardStats>>(
    '/v1/api/admin/dashboard/stats'
  )
  return data.data
}

export async function fetchRevenueChart(): Promise<{ date: string; amount: number }[]> {
  const { data } = await client.get('/v1/api/admin/dashboard/revenue-chart')
  return data.data
}

export const fetchActivityFeed = async (): Promise<ActivityItem[]> => {
  const { data } = await client.get('/v1/api/admin/dashboard/activity')
  return data.data
}

export const exportRevenue = async (): Promise<void> => {
  const { data } = await client.get('/v1/api/admin/reports/revenue/export', {
    responseType: 'blob'
  })
  
  const url = window.URL.createObjectURL(new Blob([data]))
  const link = document.createElement('a')
  link.href = url
  link.setAttribute('download', `revenue_${new Date().toISOString().split('T')[0]}.csv`)
  document.body.appendChild(link)
  link.click()
  link.remove()
}

export interface ActivityItem {
  id: string
  type: 'bid' | 'deposit' | 'registration' | 'auction_ended'
  description: string
  amount?: number
  created_at: string
}
