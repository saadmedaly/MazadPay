import { NavLink } from 'react-router-dom'
import {
  LayoutDashboard, Gavel, CreditCard, Users,
  Flag, Image, LogOut, Hammer, ShieldCheck, HelpCircle, Video, Bell, User, Settings, PhoneOff
} from 'lucide-react'
import { useAuthStore } from '@/stores/authStore'
import { cn } from '@/lib/utils'

interface NavBadgeProps { count?: number }
function NavBadge({ count }: NavBadgeProps) {
  if (!count) return null
  return (
    <span className="mr-auto text-[10px] font-bold bg-red-500 text-white rounded-full
                     min-w-[18px] h-[18px] flex items-center justify-center px-1">
      {count > 99 ? '99+' : count}
    </span>
  )
}

const NAV_SECTIONS = [
  {
    label: 'نظرة عامة',
    items: [
      { label: 'لوحة التحكم',    icon: LayoutDashboard, to: '/',             badgeKey: null },
      { label: 'الملف الشخصي',  icon: User,           to: '/profile',      badgeKey: null },
      { label: 'الإشعارات',     icon: Bell,            to: '/notifications', badgeKey: null },
    ]
  },
  {
    label: 'الإدارة',
    items: [
      { label: 'المزادات',     icon: Gavel,           to: '/auctions',     badgeKey: 'pendingAuctions' },
      { label: 'المعاملات',    icon: CreditCard,      to: '/transactions', badgeKey: 'pendingTxns' },
      { label: 'المستخدمين',    icon: Users,           to: '/users',        badgeKey: null },
      { label: 'البلاغات',     icon: Flag,            to: '/reports',      badgeKey: 'pendingReports' },
      { label: 'توثيق الحسابات', icon: ShieldCheck,     to: '/kyc',          badgeKey: 'pendingKYCs' },
      { label: 'الفئات',      icon: LayoutDashboard, to: '/categories',   badgeKey: null },
      { label: 'المواقع',      icon: Flag,            to: '/locations',    badgeKey: null },
      { label: 'الإعدادات',    icon: Settings,        to: '/settings',     badgeKey: null },
      { label: 'أرقام محظورة', icon: PhoneOff,        to: '/blocked-phones', badgeKey: null },
    ]
  },
  {
    label: 'المحتوى',
    items: [
      { label: 'الإعلانات ',     icon: Image,           to: '/banners',      badgeKey: null },
      { label: 'الأسئلة الشائعة', icon: HelpCircle,      to: '/faq',          badgeKey: null },
      { label: 'شروحات الفيديو', icon: Video,           to: '/tutorials',    badgeKey: null },
    ]
  },
]

interface SidebarProps {
  badges?: Record<string, number>
}

export function Sidebar({ badges = {} }: SidebarProps) {
  const logout = useAuthStore((s) => s.logout)

  return (
    <aside className="w-64 h-screen bg-surface-card border-l border-surface-border flex flex-col shrink-0">
      {/* Logo */}
      <div className="px-5 py-6 border-b border-surface-border">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-lg bg-mazad-primary flex items-center justify-center shadow-lg shadow-mazad-primary/20">
            <Hammer className="w-5 h-5 text-white" />
          </div>
          <div>
            <p className="font-display font-bold text-white text-base">MazadPay</p>
            <p className="text-[10px] text-surface-muted uppercase tracking-widest font-medium">لوحة الإدارة</p>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto scrollbar-thin py-6 px-3">
        {NAV_SECTIONS.map((section) => (
          <div key={section.label} className="mb-6">
            <p className="text-[10px] font-bold text-surface-muted uppercase tracking-widest
                          px-3 mb-3">{section.label}</p>
            <ul className="space-y-1">
              {section.items.map((item) => (
                <li key={item.to}>
                  <NavLink
                    to={item.to}
                    end={item.to === '/'}
                    className={({ isActive }) => cn(
                      'flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200',
                      isActive
                        ? 'bg-mazad-primary/15 text-mazad-primary border border-mazad-primary/20 shadow-sm'
                        : 'text-surface-muted hover:text-white hover:bg-surface-border/50'
                    )}
                  >
                    <item.icon className="w-4 h-4 shrink-0" />
                    <span className="flex-1">{item.label}</span>
                    {item.badgeKey && <NavBadge count={badges[item.badgeKey]} />}
                  </NavLink>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </nav>

      {/* User + Logout */}
      <div className="border-t border-surface-border p-4 bg-surface-base/30">
  
        <button
          onClick={logout}
          className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm text-surface-muted
                     hover:text-red-400 hover:bg-red-500/10 transition-all duration-200"
        >
          <LogOut className="w-4 h-4 scale-x-[-1]" />
          <span>تسجيل الخروج</span>
        </button>
      </div>
    </aside>
  )
}
