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

export async function toggleCategory(id: number, isActive: boolean): Promise<void> {
  await client.put(`/v1/api/admin/categories/${id}/toggle`, { is_active: isActive })
}

// Admin listing (includes hidden categories) -- distinct from the public
// fetchCategories() below, which excludes is_active=false rows.
export async function fetchAdminCategories(): Promise<Category[]> {
  const { data } = await client.get<APIResponse<Category[]>>('/v1/api/admin/categories')
  return data.data
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

// MAZADPAY locations-list UI staleness bug: this never accepted/forwarded
// React Query's AbortSignal to axios. useLocations() is observed by
// multiple pages (LocationsPage, AuctionsPage, AuctionRequestFormPage), so
// a background refetch (mount/window-focus) can already be in flight when
// a create/update/delete's invalidateQueries triggers its own refetch for
// the same ['locations'] key. React Query's invalidateQueries always calls
// refetchQueries with cancelRefetch:true (queryClient.js), which calls
// query.cancel() on the stale in-flight fetch -- but with no signal wired
// through, that cancellation only updates React Query's own bookkeeping;
// the actual HTTP GET keeps running on the wire. Both requests then
// resolve independently, and whichever response lands LAST (not
// necessarily the fresh, invalidation-triggered one) is what the query
// cache ends up holding -- observed live as a create/update/delete
// succeeding while the table kept rendering pre-mutation data. Passing
// signal through lets axios genuinely abort the superseded request instead
// of letting it race the real one to the cache.
export async function fetchLocations(signal?: AbortSignal): Promise<Location[]> {
  const { data } = await client.get<APIResponse<Location[]>>('/v1/api/locations', { signal })
  return data.data
}
