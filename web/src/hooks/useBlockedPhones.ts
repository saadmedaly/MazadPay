import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getBlockedPhones, blockPhone, unblockPhone } from '@/api/blockedPhones'

export function useBlockedPhones() {
  return useQuery({
    queryKey: ['blockedPhones'],
    queryFn: getBlockedPhones,
  })
}

export function useBlockPhone() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: blockPhone,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['blockedPhones'] }),
  })
}

export function useUnblockPhone() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: unblockPhone,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['blockedPhones'] }),
  })
}