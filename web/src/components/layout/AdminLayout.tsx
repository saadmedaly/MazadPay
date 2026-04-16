import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { useAdminBadges } from '@/hooks/useDashboard'

export function AdminLayout() {
  const { badges } = useAdminBadges()

  return (
    <div className="flex h-screen bg-surface-base overflow-hidden" dir="rtl">
      <Sidebar badges={badges} />
      <main className="flex-1 overflow-y-auto scrollbar-thin">
        <div className="max-w-7xl mx-auto px-6 py-8">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
