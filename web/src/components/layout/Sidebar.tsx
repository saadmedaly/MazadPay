import { NavLink } from 'react-router-dom'
import {
  LayoutDashboard, Gavel, CreditCard, Users,
  Flag, Image, LogOut, Hammer, ShieldCheck, HelpCircle, Video, Bell, User, Settings, PhoneOff,
  Wallet, MessageSquare, Handshake, Star, LifeBuoy, Grid2x2, MapPin
} from 'lucide-react'
import { useAuthStore } from '@/stores/authStore'
import { cn } from '@/lib/utils'

interface NavBadgeProps { count?: number }
function NavBadge({ count }: NavBadgeProps) {
  if (!count) return null
  return (
    <span
      className="mr-auto text-[9px] font-black rounded-full min-w-[16px] h-4 flex items-center justify-center px-1"
      style={{
        background: 'linear-gradient(135deg,#3B82F6,#2563EB)',
        color: '#fff',
        boxShadow: '0 2px 8px rgba(59,130,246,0.45)',
      }}
    >
      {count > 99 ? '99+' : count}
    </span>
  )
}

const NAV_SECTIONS = [
  {
    label: 'عام',
    items: [
      { label: 'لوحة التحكم',    icon: LayoutDashboard, to: '/',              badgeKey: null },
      { label: 'الملف الشخصي',  icon: User,             to: '/profile',       badgeKey: null },
      { label: 'الإشعارات',     icon: Bell,             to: '/notifications', badgeKey: null },
      { label: 'الرسائل',       icon: MessageSquare,    to: '/messages',      badgeKey: 'unreadMessages' },
    ]
  },
  {
    label: 'الإدارة',
    items: [
      { label: 'المزادات',        icon: Gavel,       to: '/auctions',        badgeKey: 'pendingAuctions' },
      { label: 'المعاملات',       icon: CreditCard,  to: '/transactions',    badgeKey: 'pendingTxns' },
      { label: 'المستخدمين',      icon: Users,       to: '/users',           badgeKey: null },
      { label: 'طرق الدفع',      icon: Wallet,      to: '/payment-methods', badgeKey: null },
      { label: 'البلاغات',        icon: Flag,        to: '/reports',         badgeKey: 'pendingReports' },
      { label: 'شكاوى التطبيق',   icon: LifeBuoy,   to: '/complaints',      badgeKey: null },
      { label: 'تقييمات التطبيق', icon: Star,        to: '/ratings',         badgeKey: null },
      { label: 'الفئات',          icon: Grid2x2,     to: '/categories',      badgeKey: null },
      { label: 'المواقع',         icon: MapPin,      to: '/locations',       badgeKey: null },
      { label: 'الطلبات',         icon: ShieldCheck, to: '/requests',        badgeKey: 'pendingAuctions' },
      { label: 'أرقام محظورة',    icon: PhoneOff,    to: '/blocked-phones',  badgeKey: null },
      { label: 'الإعدادات',       icon: Settings,    to: '/settings',        badgeKey: null },
    ]
  },
  {
    label: 'المحتوى',
    items: [
      { label: 'الإعلانات',       icon: Image,      to: '/banners',   badgeKey: null },
      { label: 'الرعاة',          icon: Handshake,  to: '/sponsors',  badgeKey: null },
      { label: 'الأسئلة الشائعة', icon: HelpCircle, to: '/faq',       badgeKey: null },
      { label: 'شروحات الفيديو',  icon: Video,      to: '/tutorials', badgeKey: null },
    ]
  },
]

interface SidebarProps { badges?: Record<string, number> }

export function Sidebar({ badges = {} }: SidebarProps) {
  const logout = useAuthStore((s) => s.logout)

  return (
    <aside
      className="flex flex-col shrink-0 h-screen relative overflow-hidden"
      style={{
        width: '220px',
        background: 'linear-gradient(180deg, #060F1C 0%, #08131F 100%)',
        borderLeft: '1px solid rgba(255,255,255,0.05)',
      }}
    >
      {/* Subtle blue glow top */}
      <div
        className="absolute top-0 left-0 w-40 h-40 pointer-events-none"
        style={{
          background: 'radial-gradient(ellipse at 0% 0%, rgba(59,130,246,0.08) 0%, transparent 70%)',
        }}
      />

      {/* ── Logo ── */}
      <div
        className="relative flex items-center gap-3 px-4 py-4 shrink-0"
        style={{ borderBottom: '1px solid rgba(255,255,255,0.05)' }}
      >
        {/* 3D-style icon */}
        <div
          className="w-9 h-9 rounded-xl flex items-center justify-center shrink-0 relative"
          style={{
            background: 'linear-gradient(145deg, #60A5FA 0%, #2563EB 60%, #1D4ED8 100%)',
            boxShadow: '0 4px 16px rgba(59,130,246,0.45), 0 1px 0 rgba(255,255,255,0.2) inset, 0 -1px 0 rgba(0,0,0,0.2) inset',
          }}
        >
          <Hammer className="w-[17px] h-[17px] text-white" strokeWidth={2.5} />
        </div>
        <div>
          <p
            className="font-black text-[15px] leading-none tracking-tight"
            style={{ color: '#F0F7FF', fontFamily: 'Cairo' }}
          >
            MazadPay
          </p>
          <p
            className="text-[8.5px] mt-0.5 font-bold uppercase tracking-[0.14em]"
            style={{ color: '#1E3A5F' }}
          >
            Admin Panel
          </p>
        </div>
      </div>

      {/* ── Nav ── */}
      <nav className="flex-1 overflow-y-auto scrollbar-thin py-3 px-2 space-y-0.5">
        {NAV_SECTIONS.map((section, si) => (
          <div key={section.label} className={si > 0 ? 'mt-4' : ''}>
            <p
              className="text-[8.5px] font-black uppercase tracking-[0.16em] px-3 mb-1.5"
              style={{ color: '#1A2E45' }}
            >
              {section.label}
            </p>
            {section.items.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.to === '/'}
                className={({ isActive }) => cn(
                  'flex items-center gap-2.5 px-3 py-[6.5px] rounded-[8px] text-[12.5px] font-semibold transition-all duration-200 mb-0.5 group relative',
                  isActive ? 'text-white' : 'text-[#3D5A78]'
                )}
                style={({ isActive }) => isActive ? {
                  background: 'linear-gradient(90deg, rgba(59,130,246,0.18) 0%, rgba(59,130,246,0.06) 100%)',
                  borderRight: '2px solid #3B82F6',
                  boxShadow: '-4px 0 12px rgba(59,130,246,0.1)',
                  color: '#A8C8FF',
                } : {
                  borderRight: '2px solid transparent',
                }}
                onMouseEnter={e => {
                  const el = e.currentTarget as HTMLElement
                  if (el.style.borderRightColor !== 'rgb(59, 130, 246)') {
                    el.style.background = 'rgba(255,255,255,0.04)'
                    el.style.color = '#8BA3C0'
                  }
                }}
                onMouseLeave={e => {
                  const el = e.currentTarget as HTMLElement
                  if (el.style.borderRightColor !== 'rgb(59, 130, 246)') {
                    el.style.background = 'transparent'
                    el.style.color = '#3D5A78'
                  }
                }}
              >
                <item.icon
                  className="w-[13px] h-[13px] shrink-0 transition-transform duration-200 group-hover:scale-110"
                  strokeWidth={2.2}
                />
                <span className="flex-1 truncate">{item.label}</span>
                {item.badgeKey && <NavBadge count={badges[item.badgeKey]} />}
              </NavLink>
            ))}
          </div>
        ))}
      </nav>

      {/* ── Logout ── */}
      <div
        className="p-2.5 shrink-0"
        style={{ borderTop: '1px solid rgba(255,255,255,0.04)' }}
      >
        <button
          onClick={logout}
          className="w-full flex items-center gap-2.5 px-3 py-2 rounded-[8px] text-[12.5px] font-semibold transition-all duration-200 group"
          style={{ color: '#2D4560' }}
          onMouseEnter={e => {
            (e.currentTarget as HTMLElement).style.background = 'rgba(239,68,68,0.08)'
            ;(e.currentTarget as HTMLElement).style.color = '#FC8181'
          }}
          onMouseLeave={e => {
            (e.currentTarget as HTMLElement).style.background = 'transparent'
            ;(e.currentTarget as HTMLElement).style.color = '#2D4560'
          }}
        >
          <LogOut className="w-[13px] h-[13px] scale-x-[-1] transition-transform duration-200 group-hover:-translate-x-0.5" strokeWidth={2.2} />
          <span>تسجيل الخروج</span>
        </button>
      </div>
    </aside>
  )
}
