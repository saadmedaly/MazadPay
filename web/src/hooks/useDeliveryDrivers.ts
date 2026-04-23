import client from '@/api/client'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

export interface DeliveryDriver {
  id: string
  user_id?: string
  vehicle_type?: string
  vehicle_plate?: string
  vehicle_color?: string
  license_number?: string
  rating?: number
  total_deliveries: number
  is_available: boolean
  current_location_lat?: number
  current_location_lng?: number
  created_at: string
}

export const useDeliveryDrivers = () => {
  return useQuery({
    queryKey: ['delivery-drivers'],
    queryFn: async () => {
      const response = await client.get<{ drivers: DeliveryDriver[] }>('/v1/api/admin/drivers')
      return response.data.drivers || []
    }
  })
}

export const useAvailableDrivers = () => {
  return useQuery({
    queryKey: ['delivery-drivers', 'available'],
    queryFn: async () => {
      const response = await client.get<{ drivers: DeliveryDriver[] }>('/v1/api/admin/drivers/available')
      return response.data.drivers || []
    }
  })
}

export const useRegisterDriver = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: {
      user_id: string
      vehicle_type: string
      vehicle_plate: string
      vehicle_color?: string
      license_number: string
    }) =>
      client.post('/v1/api/admin/drivers/register', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['delivery-drivers'] })
      toast.success('Chauffeur enregistré avec succès')
    },
    onError: () => {
      toast.error('Échec d\'enregistrement')
    }
  })
}

export const useUpdateDriver = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<DeliveryDriver> }) =>
      client.put(`/v1/api/admin/drivers/${id}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['delivery-drivers'] })
      toast.success('Chauffeur modifié avec succès')
    },
    onError: () => {
      toast.error('Échec de modification')
    }
  })
}

export const useDeleteDriver = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) =>
      client.delete(`/v1/api/admin/drivers/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['delivery-drivers'] })
      toast.success('Chauffeur supprimé avec succès')
    },
    onError: () => {
      toast.error('Échec de suppression')
    }
  })
}

export const useToggleDriverAvailability = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (_id: string) =>
      client.patch(`/v1/api/drivers/availability`, { available: true }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['delivery-drivers'] })
      toast.success('Disponibilité modifiée')
    },
    onError: () => {
      toast.error('Échec de modification')
    }
  })
}
