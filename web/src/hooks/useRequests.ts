import client from '@/api/client'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import type { AxiosError } from 'axios'

type ApiErrorShape = { error?: { code?: string; message?: string } }

function getApiErrorMessage(error: unknown, fallback: string): string {
  const axiosErr = error as AxiosError<ApiErrorShape>
  return axiosErr?.response?.data?.error?.message || fallback
}

function getApiErrorCode(error: unknown): string | undefined {
  const axiosErr = error as AxiosError<ApiErrorShape>
  return axiosErr?.response?.data?.error?.code
}

export interface AuctionRequest {
  id: string
  user_id: string
  category_id: number
  location_id?: number
  title_ar: string
  title_fr?: string
  title_en?: string
  description_ar?: string
  description_fr?: string
  description_en?: string
  start_price: string
  min_increment: string
  insurance_amount: string
  reserve_price?: string
  buy_now_price?: string
  start_date: string
  end_date: string
  images?: string[] | unknown
  quantity?: number
  status: 'draft' | 'pending' | 'approved' | 'rejected'
  admin_notes?: string
  reviewed_by?: string
  reviewed_at?: string
  created_at: string
  updated_at: string
  user?: {
    id: string
    phone: string
    full_name?: string
    role: string
  }
}

export interface BannerRequest {
  id: string
  user_id: string
  title_ar: string
  title_fr?: string
  title_en?: string
  description_ar?: string
  description_fr?: string
  description_en?: string
  image_url: string
  target_url?: string
  starts_at: string
  ends_at: string
  status: 'pending' | 'approved' | 'rejected'
  admin_notes?: string
  reviewed_by?: string
  reviewed_at?: string
  created_at: string
  updated_at: string
  user?: {
    id: string
    phone: string
    full_name?: string
    role: string
  }
}

// Auction Requests Hooks
export interface AuctionRequestFilters {
  status?: string
  category_id?: number
  location_id?: number
  min_price?: number
  max_price?: number
  date_from?: string
  date_to?: string
  sort_by?: string
  sort_order?: 'ASC' | 'DESC'
}

export const useAuctionRequests = (
  filters: AuctionRequestFilters = {},
  page: number = 1,
  perPage: number = 20
) => {
  return useQuery({
    queryKey: ['auction-requests', filters, page, perPage],
    queryFn: async () => {
      const response = await client.get<{ data: AuctionRequest[]; total: number; page: number; per_page: number }>('/v1/api/admin/requests/auctions', {
        params: {
          ...filters,
          page,
          per_page: perPage
        }
      })

      return {
        data: response.data.data || [],
        total: response.data.total,
        page: response.data.page,
        perPage: response.data.per_page
      }
    }
  })
}

export const useReviewAuctionRequest = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, status, notes }: { id: string; status: 'approved' | 'rejected'; notes?: string }) =>
      client.put(`/v1/api/admin/requests/auctions/${id}/review`, { status, notes: notes ?? '' }),
    onSuccess: (_, variables) => {
      toast.success('تمت مراجعة طلب المزاد بنجاح')
      queryClient.invalidateQueries({ queryKey: ['auction-requests'] })
      queryClient.invalidateQueries({ queryKey: ['auction-request', variables.id] })
    },
    onError: (error: unknown) => {
      const code = getApiErrorCode(error)
      if (code === 'rejection_notes_required') {
        toast.error('سبب الرفض مطلوب')
      } else {
        toast.error(getApiErrorMessage(error, 'فشل مراجعة طلب المزاد'))
      }
    }
  })
}

export const useUpdateAuctionRequest = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: Partial<AuctionRequest> }) =>
      client.put(`/v1/api/admin/requests/auctions/${id}`, payload),
    onSuccess: (_, variables) => {
      toast.success('تم تحديث طلب المزاد بنجاح')
      queryClient.invalidateQueries({ queryKey: ['auction-requests'] })
      queryClient.invalidateQueries({ queryKey: ['auction-request', variables.id] })
    },
    onError: (error: unknown) => {
      toast.error(getApiErrorMessage(error, 'فشل تحديث طلب المزاد'))
    }
  })
}

export const useCreateAuctionRequestAdmin = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: Partial<AuctionRequest>) =>
      client.post('/v1/api/admin/requests/auctions', payload),
    onSuccess: () => {
      toast.success('تم إنشاء طلب المزاد بنجاح')
      queryClient.invalidateQueries({ queryKey: ['auction-requests'] })
    },
    onError: (error: unknown) => {
      toast.error(getApiErrorMessage(error, 'فشل إنشاء طلب المزاد'))
    }
  })
}

export const useDeleteAuctionRequest = () => {
  return useMutation({
    mutationFn: (id: string) =>
      client.delete(`/v1/api/auctions/${id}`),
    onSuccess: () => {
      toast.success('تم حذف طلب المزاد بنجاح')
    },
    onError: () => {
      toast.error('فشل حذف طلب المزاد')
    }
  })
}

// Banner Requests Hooks
export const useBannerRequests = (status: string = '', page: number = 1, perPage: number = 20) => {
  return useQuery({
    queryKey: ['banner-requests', status, page, perPage],
    queryFn: async () => {
      const response = await client.get<{ data: BannerRequest[]; total: number; page: number; per_page: number }>('/v1/api/admin/requests/banners', {
        params: { status, page, per_page: perPage }
      })
      return {
        data: response.data.data || [],
        total: response.data.total,
        page: response.data.page,
        perPage: response.data.per_page
      }
    }
  })
}

export const useReviewBannerRequest = () => {
  return useMutation({
    mutationFn: ({ id, status, notes }: { id: string; status: 'approved' | 'rejected'; notes?: string }) =>
      client.put(`/v1/api/admin/requests/banners/${id}/review`, { status, notes }),
    onSuccess: () => {
      toast.success('تمت مراجعة طلب البانر بنجاح')
    },
    onError: () => {
      toast.error('فشل مراجعة طلب البانر')
    }
  })
}

export const useDeleteBannerRequest = () => {
  return useMutation({
    mutationFn: (id: string) =>
      client.delete(`/v1/api/admin/requests/banners/${id}`),
    onSuccess: () => {
      toast.success('تم حذف طلب البانر بنجاح')
    },
    onError: () => {
      toast.error('فشل حذف طلب البانر')
    }
  })
}

// Detail Hooks
export const useAuctionRequestByID = (id: string | null) => {
  return useQuery({
    queryKey: ['auction-request', id],
    queryFn: async () => {
      if (!id) return null
      const response = await client.get<AuctionRequest>(`/v1/api/admin/requests/auctions/${id}`)
      return response.data
    },
    enabled: !!id
  })
}

export const useBannerRequestByID = (id: string | null) => {
  return useQuery({
    queryKey: ['banner-request', id],
    queryFn: async () => {
      if (!id) return null
      const response = await client.get<BannerRequest>(`/v1/api/admin/requests/banners/${id}`)
      return response.data
    },
    enabled: !!id
  })
}

// Bulk Actions Hooks
export const useBulkReviewAuctionRequests = () => {
  return useMutation({
    mutationFn: ({ ids, status, notes }: { ids: string[]; status: 'approved' | 'rejected'; notes?: string }) =>
      client.post('/v1/api/admin/requests/auctions/bulk/review', { ids, status, notes }),
    onSuccess: (_, variables) => {
      toast.success(`تمت مراجعة ${variables.ids.length} طلب مزاد بنجاح`)
    },
    onError: () => {
      toast.error('فشل مراجعة طلبات المزاد')
    }
  })
}

export const useBulkDeleteAuctionRequests = () => {
  return useMutation({
    mutationFn: (ids: string[]) =>
      client.post('/v1/api/admin/requests/auctions/bulk/delete', { ids }),
    onSuccess: (_, variables) => {
      toast.success(`تم حذف ${variables.length} طلب مزاد بنجاح`)
    },
    onError: () => {
      toast.error('فشل حذف طلبات المزاد')
    }
  })
}

export const useBulkReviewBannerRequests = () => {
  return useMutation({
    mutationFn: ({ ids, status, notes }: { ids: string[]; status: 'approved' | 'rejected'; notes?: string }) =>
      client.post('/v1/api/admin/requests/banners/bulk/review', { ids, status, notes }),
    onSuccess: (_, variables) => {
      toast.success(`تمت مراجعة ${variables.ids.length} طلب بانر بنجاح`)
    },
    onError: () => {
      toast.error('فشل مراجعة طلبات البانر')
    }
  })
}

export const useBulkDeleteBannerRequests = () => {
  return useMutation({
    mutationFn: (ids: string[]) =>
      client.post('/v1/api/admin/requests/banners/bulk/delete', { ids }),
    onSuccess: (_, variables) => {
      toast.success(`تم حذف ${variables.length} طلب بانر بنجاح`)
    },
    onError: () => {
      toast.error('فشل حذف طلبات البانر')
    }
  })
}
