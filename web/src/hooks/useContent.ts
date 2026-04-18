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
