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

export function useCreateCategory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: api.createCategory,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['categories'] })
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
      toast.success('تم حذف الفئة بنجاح')
    },
    onError: () => toast.error('فشل حذف الفئة')
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
    onError: () => toast.error('فشل إضافة الموقع')
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
    onError: () => toast.error('فشل تحديث الموقع')
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
    onError: () => toast.error('فشل حذف الموقع')
  })
}
