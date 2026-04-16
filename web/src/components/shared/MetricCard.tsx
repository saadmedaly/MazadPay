import type { LucideIcon } from 'lucide-react'
import { cn } from '@/lib/utils'

interface Props {
  label: string
  value: string | number
  icon: LucideIcon
  trend?: { value: number; label: string }
  accent?: 'blue' | 'amber' | 'green' | 'red' | 'orange'
  urgent?: boolean
}

const ACCENT_MAP = {
  blue:   'text-blue-400 bg-blue-500/10',
  amber:  'text-amber-400 bg-amber-500/10',
  green:  'text-emerald-400 bg-emerald-500/10',
  red:    'text-red-400 bg-red-500/10',
  orange: 'text-orange-400 bg-orange-500/10',
}

export function MetricCard({ label, value, icon: Icon, trend, accent = 'blue', urgent }: Props) {
  const accentCls = ACCENT_MAP[accent]

  return (
    <div className={cn(
      'admin-card p-6 animate-fade-in relative overflow-hidden',
      urgent && 'border-orange-500/40 shadow-orange-500/10 shadow-lg'
    )}>
      {urgent && <div className="absolute top-0 right-0 w-1 h-full bg-orange-500" />}
      
      <div className="flex items-start justify-between mb-4">
        <span className="text-sm text-surface-muted font-bold tracking-wide">{label}</span>
        <div className={cn('p-2.5 rounded-lg shrink-0', accentCls)}>
          <Icon className="w-4 h-4" />
        </div>
      </div>
      
      <div className="font-display text-3xl font-bold text-white mb-2 leading-none">{value}</div>
      
      {trend && (
        <p className={cn('text-[11px] font-medium flex items-center gap-1', trend.value >= 0 ? 'text-emerald-400' : 'text-red-400')}>
          <span className="text-lg leading-none">{trend.value >= 0 ? '↑' : '↓'}</span>
          <span>{Math.abs(trend.value)}%</span>
          <span className="text-surface-muted">{trend.label}</span>
        </p>
      )}
    </div>
  )
}
