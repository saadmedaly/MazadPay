import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import client from '@/api/client'
import type { KYCVerification } from '@/types/api'

export const kycKeys = {
  all: ['kyc'] as const,
  list: (status: string) => [...kycKeys.all, 'list', { status }] as const,
}

export function useKYCs(status: string) {
  return useQuery({
    queryKey: kycKeys.list(status),
    queryFn: async () => {
      const { data } = await client.get<{ data: KYCVerification[] }>(
        '/v1/api/admin/kyc', { params: { status } }
      )
      return data.data
    },
  })
}

export function useReviewKYC() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ userId, status, notes }: { userId: string; status: string; notes: string }) =>
      client.put(`/v1/api/admin/kyc/${userId}`, { status, notes }),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: kycKeys.all })
      qc.invalidateQueries({ queryKey: ['users'] })
      qc.invalidateQueries({ queryKey: ['dashboard'] })
      toast.success(vars.status === 'approved' ? 'تم قبول توثيق الحساب' : 'تم رفض طلب التوثيق')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}
