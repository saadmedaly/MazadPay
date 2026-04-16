import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import * as api from '@/api/transactions'

export const txnKeys = {
  all:   ['transactions'] as const,
  list:  (f: api.TxnFilters) => [...txnKeys.all, 'list', f] as const,
  byId:  (id: string) => [...txnKeys.all, id] as const,
}

export function useTransactions(filters: api.TxnFilters) {
  return useQuery({
    queryKey: txnKeys.list(filters),
    queryFn:  () => api.fetchTransactions(filters),
    refetchInterval: 30_000,
    placeholderData: (prev) => prev,
  })
}

export function useTransaction(id: string) {
  return useQuery({
    queryKey: txnKeys.byId(id),
    queryFn:  () => api.fetchTransaction(id),
    enabled:  !!id,
  })
}

export function useValidateTransaction() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: api.validateTransaction,
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: txnKeys.all })
      qc.invalidateQueries({ queryKey: ['dashboard'] })
      toast.success(vars.approve
        ? 'تم اعتماد الإيداع بنجاح'
        : 'تم رفض الإيداع'
      )
    },
    onError: (err: Error) => toast.error(err.message),
  })
}
