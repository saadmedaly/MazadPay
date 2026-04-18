import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import client from '@/api/client'
import type { AdminUser, PaginatedResponse } from '@/types/api'

export const userKeys = {
  all:     ['users'] as const,
  list:    (q: string, p: number) => [...userKeys.all, 'list', { q, p }] as const,
  byId:    (id: string) => [...userKeys.all, id] as const,
  history: (id: string, type: 'auctions' | 'transactions') => [...userKeys.byId(id), 'history', type] as const,
}

export function useUsers(q: string, page: number) {
  return useQuery({
    queryKey: userKeys.list(q, page),
    queryFn: async () => {
      const { data } = await client.get<PaginatedResponse<AdminUser>>(
        '/v1/api/admin/users', { params: { q, page, per_page: 25 } }
      )
      return { data: data.data, total: data.meta.total }
    },
  })
}

export function useUser(id: string) {
  return useQuery({
    queryKey: userKeys.byId(id),
    queryFn: async () => {
      const { data } = await client.get<{ data: AdminUser }>(`/v1/api/admin/users/${id}`)
      return data.data
    },
    enabled: !!id,
  })
}

export function useMe() {
  return useQuery({
    queryKey: ['users', 'me'],
    queryFn: async () => {
      const { data } = await client.get<{ data: AdminUser }>('/v1/api/users/me')
      return data.data
    },
  })
}

export function useUpdateProfile() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (updates: { full_name: string; email: string; city: string }) =>
      client.put('/v1/api/users/me', updates),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['users', 'me'] })
      toast.success('تم تحديث الملف الشخصي بنجاح')
    },
    onError: (err: any) => toast.error(err.response?.data?.message || err.message),
  })
}

export function useChangePin() {
  return useMutation({
    mutationFn: (data: { old_pin: string; new_pin: string }) =>
      client.put('/v1/api/auth/change-password', data),
    onSuccess: () => {
      toast.success('تم تغيير كلمة المرور بنجاح')
    },
    onError: (err: any) => toast.error(err.response?.data?.message || err.message),
  })
}

export function useCreateAdmin() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { token: string; phone: string; pin: string; full_name: string; email: string }) =>
      client.post('/v1/api/auth/register-admin', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: userKeys.all })
      toast.success('تم تفعيل حساب المشرف بنجاح')
    },
    onError: (err: any) => toast.error(err.response?.data?.message || err.message),
  })
}

export function useGenerateInvitation() {
  return useMutation({
    mutationFn: async () => {
      const { data } = await client.post<{ data: { token: string } }>('/v1/api/admin/invitations')
      return data.data.token
    },
    onError: (err: any) => toast.error(err.response?.data?.message || err.message),
  })
}

export function useUserHistory(id: string, type: 'auctions' | 'transactions') {
  return useQuery({
    queryKey: userKeys.history(id, type),
    queryFn: async () => {
      const endpoint = type === 'auctions' ? 'auctions' : 'transactions'
      const { data } = await client.get<{ data: any[] }>(`/v1/api/admin/users/${id}/${endpoint}`)
      return data.data
    },
    enabled: !!id,
  })
}

export function useBlockUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, block }: { id: string; block: boolean }) =>
      client.put(`/v1/api/admin/users/${id}/block`, { block }),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: userKeys.all })
      qc.invalidateQueries({ queryKey: ['dashboard'] })
      toast.success(vars.block ? 'تم حظر المستخدم بنجاح' : 'تم إلغاء حظر المستخدم')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}
