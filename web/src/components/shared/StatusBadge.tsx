import { cn } from '@/lib/utils'

export type StatusType = 
  | 'pending' | 'active' | 'ended' | 'rejected' | 'canceled'
  | 'pending_review' | 'completed' | 'failed' | 'refunded'
  | 'verified' | 'unverified' | 'blocked' | 'reviewed' | 'dismissed'

const STATUS_CONFIG: Record<StatusType, { label: string, className: string }> = {
  // Auctions
  pending:        { label: 'قيد المراجعة',  className: 'bg-amber-500/15 text-amber-400 border-amber-500/30' },
  active:         { label: 'نشط',        className: 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30' },
  ended:          { label: 'منتهي',       className: 'bg-slate-500/15 text-slate-400 border-slate-500/30' },
  rejected:       { label: 'مرفوض',       className: 'bg-red-500/15 text-red-400 border-red-500/30' },
  canceled:       { label: 'ملغي',        className: 'bg-red-500/15 text-red-400 border-red-500/30' },

  // Transactions
  pending_review: { label: 'في انتظار المراجعة', className: 'bg-orange-500/15 text-orange-400 border-orange-500/30' },
  completed:      { label: 'مكتمل',       className: 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30' },
  failed:         { label: 'فاشل',        className: 'bg-red-500/15 text-red-400 border-red-500/30' },
  refunded:       { label: 'مسترجع',       className: 'bg-blue-500/15 text-blue-400 border-blue-500/30' },

  // Users
  verified:       { label: 'موثق',        className: 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30' },
  unverified:     { label: 'غير موثق',      className: 'bg-amber-500/15 text-amber-400 border-amber-500/30' },
  blocked:        { label: 'محظور',        className: 'bg-red-500/15 text-red-400 border-red-500/30' },

  // Reports
  reviewed:       { label: 'تمت المعالجة',   className: 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30' },
  dismissed:      { label: 'تم التجاهل',    className: 'bg-slate-500/15 text-slate-400 border-slate-500/30' },
}

interface StatusBadgeProps {
  status: string | undefined
  className?: string
}

export function StatusBadge({ status, className }: StatusBadgeProps) {
  const config = STATUS_CONFIG[status as StatusType] || { label: status, className: 'bg-surface-border text-surface-muted' }

  return (
    <span className={cn(
      'px-2.5 py-1 rounded-lg text-[10px] font-bold border uppercase tracking-wider whitespace-nowrap inline-flex items-center justify-center',
      config.className,
      className
    )}>
      {config.label}
    </span>
  )
}