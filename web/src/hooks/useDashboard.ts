import { useQuery } from '@tanstack/react-query'
import * as api from '@/api/dashboard'
import { useConversations } from './useMessages'

export function useDashboardStats() {
  return useQuery({
    queryKey: ['dashboard', 'stats'],
    queryFn: api.fetchStats,
    refetchInterval: 30_000,
  })
}

export function useRevenueChart() {
  return useQuery({
    queryKey: ['dashboard', 'revenue'],
    queryFn: api.fetchRevenueChart,
    staleTime: 5 * 60_000,
  })
}

export function useActivityFeed() {
  return useQuery({
    queryKey: ['dashboard', 'activity'],
    queryFn: api.fetchActivityFeed,
    refetchInterval: 15_000,
  })
}

// Hook used by Layout for sidebar badges
export function useAdminBadges() {
  const { data: stats } = useDashboardStats()
  const { data: convsData } = useConversations()
  const convs = convsData ?? []

  const unreadMessages = convs.reduce((acc, c) => acc + (c.unread_count || 0), 0)

  return {
    badges: {
      pendingAuctions: stats?.pending_auctions ?? 0,
      pendingTxns:     stats?.pending_transactions ?? 0,
      pendingReports:  stats?.pending_reports ?? 0,
      pendingKYCs:     stats?.pending_kycs ?? 0,
      unreadMessages:  unreadMessages,
    } as Record<string, number>
  }
}
