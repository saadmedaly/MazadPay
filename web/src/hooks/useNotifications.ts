import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getNotifications, markNotificationAsRead, markAllNotificationsAsRead, type NotificationsParams } from '@/api/notifications'

export function useNotifications(params?: NotificationsParams) {
  return useQuery({
    queryKey: ['notifications', params],
    queryFn: () => getNotifications(params),
  })
}

export function useMarkNotificationAsRead() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: markNotificationAsRead,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['notifications'] }),
  })
}

export function useMarkAllNotificationsAsRead() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: markAllNotificationsAsRead,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['notifications'] }),
  })
}