import type { LucideIcon } from 'lucide-react'
import { cn } from '@/lib/utils'

interface Props {
  label: string
  value: string | number
  icon: LucideIcon
  trend?: { value: number; label: string }
  accent?: 'blue' | 'amber' | 'green' | 'red' | 'orange' | 'purple'
  urgent?: boolean
}

const ACCENT: Record<string, {
  iconColor: string; iconBg: string; glow: string; border: string; topLine: string
}> = {
  blue:   { iconColor: '#60A5FA', iconBg: 'rgba(59,130,246,0.12)',  glow: 'rgba(59,130,246,0.15)',  border: 'rgba(59,130,246,0.2)',  topLine: '#3B82F6' },
  amber:  { iconColor: '#FCD34D', iconBg: 'rgba(245,158,11,0.12)',  glow: 'rgba(245,158,11,0.15)',  border: 'rgba(245,158,11,0.2)',  topLine: '#F59E0B' },
  green:  { iconColor: '#34D399', iconBg: 'rgba(16,185,129,0.12)',  glow: 'rgba(16,185,129,0.15)',  border: 'rgba(16,185,129,0.2)',  topLine: '#10B981' },
  red:    { iconColor: '#F87171', iconBg: 'rgba(239,68,68,0.12)',   glow: 'rgba(239,68,68,0.15)',   border: 'rgba(239,68,68,0.2)',   topLine: '#EF4444' },
  orange: { iconColor: '#FB923C', iconBg: 'rgba(249,115,22,0.12)',  glow: 'rgba(249,115,22,0.15)',  border: 'rgba(249,115,22,0.2)',  topLine: '#F97316' },
  purple: { iconColor: '#A78BFA', iconBg: 'rgba(124,58,237,0.12)', glow: 'rgba(124,58,237,0.15)', border: 'rgba(124,58,237,0.2)', topLine: '#8B5CF6' },
}

export function MetricCard({ label, value, icon: Icon, trend, accent = 'blue', urgent }: Props) {
  const a = ACCENT[accent]

  return (
    <div
      className={cn('rounded-xl p-5 relative overflow-hidden transition-all duration-250 animate-fade-in group cursor-default')}
      style={{
        background: 'linear-gradient(145deg, #0D1B2E 0%, #0A1828 100%)',
        border: `1px solid ${urgent ? 'rgba(249,115,22,0.3)' : a.border}`,
        boxShadow: urgent
          ? '0 4px 20px rgba(249,115,22,0.12), inset 0 1px 0 rgba(255,255,255,0.04)'
          : 'inset 0 1px 0 rgba(255,255,255,0.04), 0 2px 8px rgba(0,0,0,0.3)',
      }}
      onMouseEnter={e => {
        const el = e.currentTarget as HTMLElement
        el.style.transform = 'translateY(-2px)'
        el.style.boxShadow = `0 12px 32px ${a.glow}, 0 4px 12px rgba(0,0,0,0.4), inset 0 1px 0 rgba(255,255,255,0.06)`
        el.style.borderColor = a.topLine.replace('#', 'rgba(') + (a.topLine.length === 7 ? ',0.4)' : '')
      }}
      onMouseLeave={e => {
        const el = e.currentTarget as HTMLElement
        el.style.transform = 'translateY(0)'
        el.style.boxShadow = urgent
          ? '0 4px 20px rgba(249,115,22,0.12), inset 0 1px 0 rgba(255,255,255,0.04)'
          : 'inset 0 1px 0 rgba(255,255,255,0.04), 0 2px 8px rgba(0,0,0,0.3)'
        el.style.borderColor = urgent ? 'rgba(249,115,22,0.3)' : a.border
      }}
    >
      {/* Top color line */}
      <div
        className="absolute top-0 right-0 left-0 h-[2px] rounded-t-xl"
        style={{ background: `linear-gradient(90deg, transparent, ${a.topLine}, transparent)`, opacity: 0.6 }}
      />

      {/* Urgent side bar */}
      {urgent && (
        <div
          className="absolute top-0 right-0 w-[2px] h-full rounded-r-xl"
          style={{ background: 'linear-gradient(180deg, #F97316, transparent)' }}
        />
      )}

      {/* Subtle background glow */}
      <div
        className="absolute top-0 left-0 w-24 h-24 pointer-events-none rounded-full"
        style={{
          background: `radial-gradient(circle, ${a.glow} 0%, transparent 70%)`,
          filter: 'blur(20px)',
        }}
      />

      {/* Top row */}
      <div className="relative flex items-start justify-between mb-4">
        <p className="text-[12px] font-semibold leading-snug" style={{ color: '#4A6080' }}>
          {label}
        </p>
        <div
          className="w-9 h-9 rounded-xl flex items-center justify-center shrink-0 transition-transform duration-250 group-hover:scale-110 group-hover:rotate-3"
          style={{
            background: a.iconBg,
            color: a.iconColor,
            boxShadow: `0 4px 12px ${a.glow}`,
          }}
        >
          <Icon className="w-[17px] h-[17px]" strokeWidth={2} />
        </div>
      </div>

      {/* Value */}
      <p
        className="relative text-[32px] font-black leading-none mb-2.5"
        style={{ color: '#E2E8F0', fontFamily: 'Cairo' }}
      >
        {value}
      </p>

      {/* Trend */}
      {trend && (
        <div className="relative flex items-center gap-1.5">
          <span
            className="text-[10px] font-bold px-2 py-0.5 rounded-full"
            style={{
              background: trend.value >= 0 ? 'rgba(16,185,129,0.12)' : 'rgba(239,68,68,0.12)',
              color: trend.value >= 0 ? '#34D399' : '#F87171',
            }}
          >
            {trend.value >= 0 ? '↑' : '↓'} {Math.abs(trend.value)}%
          </span>
          <span className="text-[10px] font-medium" style={{ color: '#2D4560' }}>
            {trend.label}
          </span>
        </div>
      )}
    </div>
  )
}
