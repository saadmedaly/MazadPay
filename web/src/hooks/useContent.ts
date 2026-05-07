import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import client from '@/api/client'
import type { FAQItem, Tutorial } from '@/types/api'

export const contentKeys = {
  faq: ['faq'] as const,
  tutorials: ['tutorials'] as const,
}

export function useFAQs() {
  return useQuery({
    queryKey: contentKeys.faq,
    queryFn: async () => {
      const { data } = await client.get<{ data: FAQItem[] }>('/v1/api/faq')
      return data.data
    },
  })
}

export function useCreateFAQ() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (item: Partial<FAQItem>) => client.post('/v1/api/admin/faq', item), // Assuming endpoint exists or admin logic
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: contentKeys.faq })
      toast.success('تمت إضافة السؤال بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}

export function useDeleteFAQ() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => client.delete(`/v1/api/admin/faq/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: contentKeys.faq })
      toast.success('تم حذف السؤال')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}

export function useUpdateFAQ() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (item: FAQItem) => client.put(`/v1/api/admin/faq/${item.id}`, item),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: contentKeys.faq })
      toast.success('تم تحديث السؤال بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}

export function useTutorials() {
  return useQuery({
    queryKey: contentKeys.tutorials,
    queryFn: async () => {
      const { data } = await client.get<{ data: Tutorial[] }>('/v1/api/tutorials')
      return data.data
    },
  })
}

export function useCreateTutorial() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: Partial<Tutorial>) => client.post('/v1/api/admin/tutorials', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: contentKeys.tutorials })
      toast.success('تمت إضافة الفيديو بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}

export function useDeleteTutorial() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => client.delete(`/v1/api/admin/tutorials/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: contentKeys.tutorials })
      toast.success('تم حذف الفيديو')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}

export function useUpdateTutorial() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: Tutorial) => client.put(`/v1/api/admin/tutorials/${data.id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: contentKeys.tutorials })
      toast.success('تم تحديث الفيديو بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}

// Upload hooks for R2
export function useUploadFAQImage() {
  return useMutation({
    mutationFn: (file: File) => {
      const formData = new FormData()
      formData.append('file', file)
      return client.post('/v1/api/admin/faq/upload', formData, {
        timeout: 30_000,
        headers: { 'Content-Type': undefined },
      }).then(res => res.data.data)
    },
    onError: (err: any) => {
      const message = err.response?.data?.message || err.message || 'فشل رفع الصورة'
      toast.error(message)
    },
  })
}

export function useUploadTutorialVideo() {
  return useMutation({
    mutationFn: (file: File) => {
      const formData = new FormData()
      formData.append('file', file)
      return client.post('/v1/api/admin/tutorials/upload-video', formData, {
        timeout: 120_000, // 2 min for videos
        headers: { 'Content-Type': undefined },
      }).then(res => res.data.data)
    },
    onError: (err: any) => {
      const message = err.response?.data?.message || err.message || 'فشل رفع الفيديو'
      toast.error(message)
    },
  })
}

export function useUploadTutorialThumbnail() {
  return useMutation({
    mutationFn: (file: File) => {
      const formData = new FormData()
      formData.append('file', file)
      return client.post('/v1/api/admin/tutorials/upload-thumbnail', formData, {
        timeout: 30_000,
        headers: { 'Content-Type': undefined },
      }).then(res => res.data.data)
    },
    onError: (err: any) => {
      const message = err.response?.data?.message || err.message || 'فشل رفع الصورة المصغرة'
      toast.error(message)
    },
  })
}

// Banner hooks
export function useBanners() {
  return useQuery({
    queryKey: ['banners'] as const,
    queryFn: async () => {
      const { data } = await client.get<{ data: { id: number; title_ar: string; image_url: string; is_active: boolean }[] }>('/v1/api/banners')
      return data.data
    },
  })
}

export function useAdminBanners() {
  return useQuery({
    queryKey: ['admin-banners'] as const,
    queryFn: async () => {
      const { data } = await client.get<{ data: { id: number; title_ar: string; image_url: string; is_active: boolean }[] }>('/v1/api/admin/banners')
      return data.data
    },
  })
}

export function useCreateBanner() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { title_ar: string; title_fr: string; title_en: string; image_url: string; target_url: string }) =>
      client.post('/v1/api/admin/banners', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['banners'] })
      qc.invalidateQueries({ queryKey: ['admin-banners'] })
      toast.success('تمت إضافة الإعلان بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}

export function useUpdateBanner() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<{ title_ar: string; image_url: string; is_active: boolean }> }) =>
      client.put(`/v1/api/admin/banners/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['banners'] })
      qc.invalidateQueries({ queryKey: ['admin-banners'] })
      toast.success('تم تحديث الإعلان بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}

export function useDeleteBanner() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => client.delete(`/v1/api/admin/banners/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['banners'] })
      qc.invalidateQueries({ queryKey: ['admin-banners'] })
      toast.success('تم حذف الإعلان')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}

export function useToggleBanner() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, isActive }: { id: number; isActive: boolean }) =>
      client.put(`/v1/api/admin/banners/${id}/toggle`, { is_active: isActive }),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: ['banners'] })
      qc.invalidateQueries({ queryKey: ['admin-banners'] })
      toast.success(vars.isActive ? 'تم تفعيل الإعلان' : 'تم تعطيل الإعلان')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}

export function useUploadBannerImage() {
  return useMutation({
    mutationFn: (file: File) => {
      const formData = new FormData()
      formData.append('file', file)
      return client.post('/v1/api/admin/banners/upload', formData, {
        timeout: 30_000,
        headers: { 'Content-Type': undefined },
      }).then(res => res.data.data)
    },
    onError: (err: any) => {
      const message = err.response?.data?.message || err.message || 'فشل رفع صورة الإعلان'
      toast.error(message)
    },
  })
}
