import client from '@/api/client'
import { useMutation, useQuery } from '@tanstack/react-query'
import { toast } from 'sonner'

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
  images?: any
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

export interface BannerRequest {
  id: string
  user_id: string
  title_ar: string
  title_fr?: string
  title_en?: string
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
      // Récupérer les enchères avec statut pending (en attente d'approbation)
      const response = await client.get<{ data: any[]; total: number; page: number; per_page: number }>('/v1/api/auctions', {
        params: {
          status: 'pending',
          ...filters,
          page,
          per_page: perPage
        }
      })

      // Transformer les enchères en format AuctionRequest
      const auctions = response.data.data || []
      const requests: AuctionRequest[] = auctions.map(auction => ({
        id: auction.id,
        user_id: auction.user_id || auction.userId,
        title: auction.title_ar || auction.titleAr || auction.title,
        description: auction.description_ar || auction.descriptionAr || auction.description,
        category: auction.category,
        status: auction.status === 'active' ? 'approved' : auction.status === 'pending' ? 'pending' : 'rejected',
        created_at: auction.created_at || auction.createdAt,
        updated_at: auction.updated_at || auction.updatedAt,
        // Champs additionnels
        start_price: auction.start_price || auction.startPrice,
        location: auction.location,
        images: auction.images || [],
        end_time: auction.end_time || auction.endTime
      }))

      return {
        data: requests,
        total: response.data.total,
        page: response.data.page,
        perPage: response.data.per_page
      }
    }
  })
}

export const useReviewAuctionRequest = () => {
  return useMutation({
    mutationFn: ({ id, status, notes }: { id: string; status: 'approved' | 'rejected'; notes?: string }) => {
      // Convertir 'approved'/'rejected' en statut backend
      const auctionStatus = status === 'approved' ? 'active' : 'rejected'
      return client.put(`/v1/api/auctions/${id}/status`, { status: auctionStatus, notes })
    },
    onSuccess: () => {
      toast.success('تمت مراجعة طلب المزاد بنجاح')
    },
    onError: () => {
      toast.error('فشل مراجعة طلب المزاد')
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
