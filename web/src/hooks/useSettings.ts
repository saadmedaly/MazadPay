import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getSettings, updateSetting } from '@/api/settings'

export function useSettings() {
  return useQuery({
    queryKey: ['settings'],
    queryFn: getSettings,
  })
}

export function useUpdateSetting() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: updateSetting,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] })
    },
  })
}