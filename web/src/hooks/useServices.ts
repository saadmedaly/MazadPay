import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getServiceRequests, getServiceRequestById, updateServiceRequest, type ServicesParams } from '@/api/services'
import { toast } from 'sonner'

export function useServiceRequests(params?: ServicesParams) {
  return useQuery({
    queryKey: ['serviceRequests', params],
    queryFn: () => getServiceRequests(params),
  })
}

export function useServiceRequest(id: string) {
  return useQuery({
    queryKey: ['serviceRequest', id],
    queryFn: () => getServiceRequestById(id),
    enabled: !!id,
  })
}

export function useUpdateServiceRequest() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: Parameters<typeof updateServiceRequest>[1] }) =>
      updateServiceRequest(id, payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['serviceRequests'] })
      qc.invalidateQueries({ queryKey: ['serviceRequest'] })
      toast.success('تم تحديث طلب الخدمة بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}