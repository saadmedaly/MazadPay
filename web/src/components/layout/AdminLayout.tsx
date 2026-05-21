import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { useAdminBadges } from '@/hooks/useDashboard'
import { useNotifications } from '@/hooks/useNotifications'
import { NotificationBell } from '../shared/NotificationBell'

export function AdminLayout() {
  const { badges } = useAdminBadges()
  useNotifications()

  return (
    <div
      className="flex h-screen overflow-hidden"
      dir="rtl"
      style={{ background: '#07111F' }}
    >
      <Sidebar badges={badges} />

      <div className="flex-1 flex flex-col min-w-0">
        {/* Topbar */}
        <header
          className="shrink-0 flex items-center justify-end px-6"
          style={{
            height: '52px',
            background: 'rgba(9,22,42,0.95)',
            borderBottom: '1px solid rgba(255,255,255,0.05)',
            backdropFilter: 'blur(8px)',
          }}
        >
          <div className="flex items-center gap-2">
            <NotificationBell />
          </div>
        </header>

        {/* Content */}
        <main className="flex-1 overflow-y-auto scrollbar-thin">
          <div className="max-w-[1400px] mx-auto px-6 py-7">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}
