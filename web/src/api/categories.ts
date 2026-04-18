import client from './client'
import type { APIResponse, Category } from '@/types/api'

export async function getCategories(): Promise<Category[]> {
  const { data } = await client.get<APIResponse<Category[]>>('/v1/api/categories')
  return data.data
}

export async function createCategory(payload: {
  name_ar: string
  name_fr: string
  parent_id?: number
  display_order?: number
}): Promise<Category> {
  const { data } = await client.post<APIResponse<Category>>('/v1/api/admin/categories', payload)
  return data.data
}

export async function updateCategory(id: number, payload: {
  name_ar: string
  name_fr: string
  parent_id?: number
  display_order?: number
}): Promise<Category> {
  const { data } = await client.put<APIResponse<Category>>(`/v1/api/admin/categories/${id}`, payload)
  return data.data
}

export async function deleteCategory(id: number): Promise<void> {
  await client.delete(`/v1/api/admin/categories/${id}`)
}