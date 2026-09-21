import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as api from '@/api/metadata'
import { toast } from 'sonner'
import type { Category, Location } from '@/types/api'

// Categories
export function useCategories() {
  return useQuery({
    queryKey: ['categories'],
    queryFn: api.fetchCategories
  })
}

// Admin listing (Customer #27): includes hidden (is_active=false)
// categories/subcategories, unlike useCategories() above which uses the
// public endpoint and excludes them.
export function useAdminCategories() {
  return useQuery({
    queryKey: ['admin-categories'],
    queryFn: api.fetchAdminCategories
  })
}

export function useCreateCategory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: api.createCategory,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['categories'] })
      qc.invalidateQueries({ queryKey: ['admin-categories'] })
      toast.success('تمت إضافة الفئة بنجاح')
    },
    onError: () => toast.error('فشل إضافة الفئة')
  })
}

export function useUpdateCategory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: Partial<Category> }) => api.updateCategory(id, payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['categories'] })
      qc.invalidateQueries({ queryKey: ['admin-categories'] })
      toast.success('تم تحديث الفئة بنجاح')
    },
    onError: () => toast.error('فشل تحديث الفئة')
  })
}

export function useDeleteCategory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: api.deleteCategory,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['categories'] })
      qc.invalidateQueries({ queryKey: ['admin-categories'] })
      toast.success('تم حذف الفئة بنجاح')
    },
    onError: () => toast.error('فشل حذف الفئة')
  })
}

export function useToggleCategory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, isActive }: { id: number; isActive: boolean }) => api.toggleCategory(id, isActive),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['categories'] })
      qc.invalidateQueries({ queryKey: ['admin-categories'] })
      toast.success(variables.isActive ? 'تم إظهار الفئة' : 'تم إخفاء الفئة')
    },
    onError: () => toast.error('فشل تحديث حالة الفئة')
  })
}

// Locations
export function useLocations() {
  return useQuery({
    queryKey: ['locations'],
    queryFn: api.fetchLocations
  })
}

export function useCreateLocation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: api.createLocation,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['locations'] })
      toast.success('تمت إضافة الموقع بنجاح')
    },
    onError: (err: Error) => toast.error(err.message || 'فشل إضافة الموقع')
  })
}

export function useUpdateLocation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: Partial<Location> }) => api.updateLocation(id, payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['locations'] })
      toast.success('تم تحديث الموقع بنجاح')
    },
    onError: (err: Error) => toast.error(err.message || 'فشل تحديث الموقع')
  })
}

export function useDeleteLocation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: api.deleteLocation,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['locations'] })
      toast.success('تم حذف الموقع بنجاح')
    },
    // MAZADPAY location-delete UI bug: this previously discarded the real
    // error entirely and always showed the same generic "فشل حذف الموقع"
    // toast, regardless of whether the backend rejected with 403
    // (permission), 404 (stale/already-deleted ID), 409 (conflict), or a
    // genuine 500/network failure -- indistinguishable to the admin. The
    // axios interceptor (api/client.ts) already extracts the backend's
    // real error.message into err.message before this handler runs, so
    // surfacing it here is enough; no backend/delete-semantics change.
    //
    // A 404 specifically means the row is already gone server-side (proven
    // live: DELETE on a valid ID always returns 200, only a stale/
    // already-deleted ID returns 404) -- this is a client cache being out of
    // sync with the server, not a destructive failure, so it self-heals by
    // refetching the list (which will correctly drop the row) instead of
    // showing an alarming error.
    onError: (err: Error & { status?: number }) => {
      if (err.status === 404) {
        qc.invalidateQueries({ queryKey: ['locations'] })
        toast.info('الموقع غير موجود أو تم حذفه مسبقًا')
        return
      }
      toast.error(err.message || 'فشل حذف الموقع')
    }
  })
}
