import client from '@/api/client'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

export interface AuctionBoost {
  id: string
  auction_id: string
  boost_type: string
  start_at: string
  end_at: string
  amount?: number
  status: string
  created_at: string
}

export const useAuctionBoosts = () => {
  return useQuery({
    queryKey: ['auction-boosts'],
    queryFn: async () => {
      const response = await client.get<{ boosts: AuctionBoost[] }>('/v1/api/admin/boosts/active')
      return response.data.boosts || []
    }
  })
}

export const useAuctionBoostsByAuction = (auctionId: string | null) => {
  return useQuery({
    queryKey: ['auction-boosts', auctionId],
    queryFn: async () => {
      if (!auctionId) return []
      const response = await client.get<{ boosts: AuctionBoost[] }>(`/v1/api/auctions/${auctionId}/boosts`)
      return response.data.boosts || []
    },
    enabled: !!auctionId
  })
}

export const useCreateBoost = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ auctionId, data }: { auctionId: string; data: {
      boost_type: string
      start_at: string
      end_at: string
      amount?: number
    }}) =>
      client.post(`/v1/api/auctions/${auctionId}/boost`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['auction-boosts'] })
      toast.success('Boost créé avec succès')
    },
    onError: () => {
      toast.error('Échec de création')
    }
  })
}

export const useCancelBoost = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ auctionId, boostId }: { auctionId: string; boostId: string }) =>
      client.delete(`/v1/api/auctions/${auctionId}/boosts/${boostId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['auction-boosts'] })
      toast.success('Boost annulé avec succès')
    },
    onError: () => {
      toast.error('Échec d\'annulation')
    }
  })
}

export const useUpdateBoostStatus = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      client.put(`/v1/api/admin/boosts/${id}/status`, { status }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['auction-boosts'] })
      toast.success('Statut modifié avec succès')
    },
    onError: () => {
      toast.error('Échec de modification')
    }
  })
}
