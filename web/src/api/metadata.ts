import client from './client'
import type { APIResponse, Category, Location } from '@/types/api'

// Categories
export async function createCategory(payload: Partial<Category>): Promise<Category> {
  const { data } = await client.post<APIResponse<Category>>('/v1/api/admin/categories', payload)
  return data.data
}

export async function updateCategory(id: number, payload: Partial<Category>): Promise<Category> {
  const { data } = await client.put<APIResponse<Category>>(`/v1/api/admin/categories/${id}`, payload)
  return data.data
}

export async function deleteCategory(id: number): Promise<void> {
  await client.delete(`/v1/api/admin/categories/${id}`)
}

// Locations
export async function createLocation(payload: Partial<Location>): Promise<Location> {
  const { data } = await client.post<APIResponse<Location>>('/v1/api/admin/locations', payload)
  return data.data
}

export async function updateLocation(id: number, payload: Partial<Location>): Promise<Location> {
  const { data } = await client.put<APIResponse<Location>>(`/v1/api/admin/locations/${id}`, payload)
  return data.data
}

export async function deleteLocation(id: number): Promise<void> {
  await client.delete(`/v1/api/admin/locations/${id}`)
}

// Fetch lists (Publicly available)
export async function fetchCategories(): Promise<Category[]> {
  const { data } = await client.get<APIResponse<Category[]>>('/v1/api/categories')
  return data.data
}

export async function fetchLocations(): Promise<Location[]> {
  const { data } = await client.get<APIResponse<Location[]>>('/v1/api/locations')
  return data.data
}
