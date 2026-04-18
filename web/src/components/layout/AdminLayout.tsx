import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { useAdminBadges } from '@/hooks/useDashboard'
import { useNotifications } from '@/hooks/useNotifications'
import { NotificationBell } from '../shared/NotificationBell'
import { Search } from 'lucide-react'

export function AdminLayout() {
  const { badges } = useAdminBadges()
  useNotifications() // Activer l'écoute des notifications

  return (
    <div className="flex h-screen bg-surface-base overflow-hidden" dir="rtl">
      <Sidebar badges={badges} />
      
      <div className="flex-1 flex flex-col min-w-0">
        {/* TopBar Premium */}
        <header className=" bg-transparent border-b flex items-end justify-end px-8 shrink-0 shadow-sm z-10">
    

          <div className="flex items-center gap-4 m-3  ">
            <NotificationBell />
         
          </div>
        </header>

        <main className="flex-1 overflow-y-auto scrollbar-thin">
          <div className="max-w-7xl mx-auto px-6 py-8">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}
