import { useState } from 'react'
import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { useAdminBadges } from '@/hooks/useDashboard'
import { useNotifications } from '@/hooks/useNotifications'
import { NotificationBell } from '../shared/NotificationBell'
import { Menu, X } from 'lucide-react'

export function AdminLayout() {
  const { badges } = useAdminBadges()
  useNotifications()
  const [sidebarOpen, setSidebarOpen] = useState(false)

  return (
    <div
      className="flex h-screen overflow-hidden"
      dir="rtl"
      style={{ background: '#07111F' }}
    >
      {/* Mobile overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 z-30 bg-black/60 md:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar — hidden on mobile unless open */}
      <div
        className={`
          fixed inset-y-0 right-0 z-40 transition-transform duration-300
          md:relative md:translate-x-0 md:flex md:flex-col md:shrink-0
          ${sidebarOpen ? 'translate-x-0' : 'translate-x-full md:translate-x-0'}
        `}
        style={{ width: '220px' }}
      >
        <Sidebar badges={badges} onClose={() => setSidebarOpen(false)} />
      </div>

      <div className="flex-1 flex flex-col min-w-0">
        {/* Topbar */}
        <header
          className="shrink-0 flex items-center justify-between px-4 md:px-6"
          style={{
            height: '52px',
            background: 'rgba(9,22,42,0.95)',
            borderBottom: '1px solid rgba(255,255,255,0.05)',
            backdropFilter: 'blur(8px)',
          }}
        >
          {/* Hamburger — mobile only */}
          <button
            className="md:hidden flex items-center justify-center w-8 h-8 rounded-lg transition-colors"
            style={{ color: '#8BA3C0' }}
            onClick={() => setSidebarOpen(true)}
            aria-label="فتح القائمة"
          >
            <Menu className="w-5 h-5" />
          </button>

          <div className="flex items-center gap-2 mr-auto">
            <NotificationBell />
          </div>
        </header>

        {/* Content */}
        <main className="flex-1 overflow-y-auto scrollbar-thin">
          <div className="max-w-[1400px] mx-auto px-3 py-4 md:px-6 md:py-7">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}
