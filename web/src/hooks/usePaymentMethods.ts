import client from '@/api/client'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

export interface PaymentMethod {
  id: number
  code: string
  name_ar: string
  name_fr: string
  name_en?: string
  logo_url?: string
  is_active: boolean
  country_id?: number
  created_at: string
}

export const usePaymentMethods = () => {
  return useQuery({
    queryKey: ['payment-methods'],
    queryFn: async () => {
      const response = await client.get<{ payment_methods: PaymentMethod[] }>('/v1/api/admin/payment-methods')
      return response.data.payment_methods || []
    }
  })
}

export const useCreatePaymentMethod = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: Omit<PaymentMethod, 'id' | 'created_at'>) =>
      client.post('/v1/api/admin/payment-methods', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['payment-methods'] })
      toast.success('Méthode de paiement créée avec succès')
    },
    onError: () => {
      toast.error('Échec de création')
    }
  })
}

export const useUpdatePaymentMethod = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<PaymentMethod> }) =>
      client.put(`/v1/api/admin/payment-methods/${id}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['payment-methods'] })
      toast.success('Méthode de paiement modifiée avec succès')
    },
    onError: () => {
      toast.error('Échec de modification')
    }
  })
}

export const useDeletePaymentMethod = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) =>
      client.delete(`/v1/api/admin/payment-methods/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['payment-methods'] })
      toast.success('Méthode de paiement supprimée avec succès')
    },
    onError: () => {
      toast.error('Échec de suppression')
    }
  })
}

export const useTogglePaymentMethodStatus = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) =>
      client.patch(`/v1/api/admin/payment-methods/${id}/toggle`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['payment-methods'] })
      toast.success('Statut modifié avec succès')
    },
    onError: () => {
      toast.error('Échec de modification du statut')
    }
  })
}
