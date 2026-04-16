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
