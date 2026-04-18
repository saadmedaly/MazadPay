import client from './client'
import type { APIResponse, ServiceRequest } from '@/types/api'

export interface ServicesParams {
  page?: number
  per_page?: number
  status?: string
}

export async function getServiceRequests(params?: ServicesParams): Promise<{ data: ServiceRequest[]; total: number }> {
  const { data } = await client.get<APIResponse<ServiceRequest[]>>('/v1/api/admin/service-requests', { params })
  return { data: data.data, total: data.data.length }
}

export async function getServiceRequestById(id: string): Promise<ServiceRequest> {
  const { data } = await client.get<APIResponse<ServiceRequest>>(`/v1/api/admin/service-requests/${id}`)
  return data.data
}

export async function updateServiceRequest(id: string, payload: {
  status?: string
  driver_id?: string
}): Promise<ServiceRequest> {
  const { data } = await client.put<APIResponse<ServiceRequest>>(`/v1/api/admin/service-requests/${id}`, payload)
  return data.data
}