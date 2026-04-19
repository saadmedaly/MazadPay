import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getLocations, createLocation, updateLocation, deleteLocation } from '@/api/locations'
import { toast } from 'sonner'

export function useLocations() {
  return useQuery({
    queryKey: ['locations'],
    queryFn: getLocations,
  })
}

export function useCreateLocation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: createLocation,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['locations'] })
      toast.success('تمت إضافة الموقع بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}

export function useUpdateLocation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: Parameters<typeof updateLocation>[1] }) =>
      updateLocation(id, payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['locations'] })
      toast.success('تم تحديث الموقع بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}

export function useDeleteLocation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: deleteLocation,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['locations'] })
      toast.success('تم حذف الموقع بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}