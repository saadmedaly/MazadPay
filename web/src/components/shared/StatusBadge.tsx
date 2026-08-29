import { cn } from '@/lib/utils'

export type StatusType =
  | 'draft' | 'pending' | 'approved' | 'active' | 'ended' | 'rejected' | 'canceled'
  | 'pending_review' | 'completed' | 'failed' | 'refunded'
  | 'verified' | 'unverified' | 'blocked' | 'reviewed' | 'dismissed'

const STATUS_CONFIG: Record<StatusType, { label: string; bg: string; color: string; border: string; dot: string }> = {
  draft:          { label: 'مسودة',              bg: 'rgba(100,116,139,0.12)',color: '#94A3B8', border: 'rgba(100,116,139,0.2)', dot: '#64748B' },
  pending:        { label: 'قيد المراجعة',       bg: 'rgba(245,158,11,0.12)', color: '#FCD34D', border: 'rgba(245,158,11,0.25)', dot: '#F59E0B' },
  approved:       { label: 'مقبول',              bg: 'rgba(16,185,129,0.12)', color: '#34D399', border: 'rgba(16,185,129,0.25)', dot: '#10B981' },
  active:         { label: 'نشط',               bg: 'rgba(16,185,129,0.12)', color: '#34D399', border: 'rgba(16,185,129,0.25)', dot: '#10B981' },
  ended:          { label: 'منتهي',              bg: 'rgba(100,116,139,0.12)',color: '#94A3B8', border: 'rgba(100,116,139,0.2)', dot: '#64748B' },
  rejected:       { label: 'مرفوض',              bg: 'rgba(239,68,68,0.1)',   color: '#F87171', border: 'rgba(239,68,68,0.25)', dot: '#EF4444' },
  canceled:       { label: 'ملغي',               bg: 'rgba(239,68,68,0.1)',   color: '#F87171', border: 'rgba(239,68,68,0.25)', dot: '#EF4444' },
  pending_review: { label: 'في انتظار المراجعة', bg: 'rgba(249,115,22,0.12)', color: '#FB923C', border: 'rgba(249,115,22,0.25)', dot: '#F97316' },
  completed:      { label: 'مكتمل',              bg: 'rgba(16,185,129,0.12)', color: '#34D399', border: 'rgba(16,185,129,0.25)', dot: '#10B981' },
  failed:         { label: 'فاشل',               bg: 'rgba(239,68,68,0.1)',   color: '#F87171', border: 'rgba(239,68,68,0.25)', dot: '#EF4444' },
  refunded:       { label: 'مسترجع',              bg: 'rgba(59,130,246,0.12)', color: '#60A5FA', border: 'rgba(59,130,246,0.25)', dot: '#3B82F6' },
  verified:       { label: 'موثق',               bg: 'rgba(16,185,129,0.12)', color: '#34D399', border: 'rgba(16,185,129,0.25)', dot: '#10B981' },
  unverified:     { label: 'غير موثق',           bg: 'rgba(245,158,11,0.12)', color: '#FCD34D', border: 'rgba(245,158,11,0.25)', dot: '#F59E0B' },
  blocked:        { label: 'محظور',              bg: 'rgba(239,68,68,0.1)',   color: '#F87171', border: 'rgba(239,68,68,0.25)', dot: '#EF4444' },
  reviewed:       { label: 'تمت المعالجة',        bg: 'rgba(16,185,129,0.12)', color: '#34D399', border: 'rgba(16,185,129,0.25)', dot: '#10B981' },
  dismissed:      { label: 'تم التجاهل',         bg: 'rgba(100,116,139,0.12)',color: '#94A3B8', border: 'rgba(100,116,139,0.2)', dot: '#64748B' },
}

interface StatusBadgeProps {
  status: string | undefined
  className?: string
}

export function StatusBadge({ status, className }: StatusBadgeProps) {
  const cfg = STATUS_CONFIG[status as StatusType]

  if (!cfg) {
    return (
      <span
        className={cn('inline-flex items-center gap-1.5 px-2.5 py-[3px] rounded-full text-[10.5px] font-bold', className)}
        style={{ background: 'rgba(255,255,255,0.06)', color: '#4A6080', border: '1px solid rgba(255,255,255,0.08)' }}
      >
        {status}
      </span>
    )
  }

  return (
    <span
      className={cn('inline-flex items-center gap-[5px] px-2.5 py-[3px] rounded-full text-[10.5px] font-bold whitespace-nowrap', className)}
      style={{ background: cfg.bg, color: cfg.color, border: `1px solid ${cfg.border}` }}
    >
      <span
        className="w-[5px] h-[5px] rounded-full shrink-0"
        style={{ background: cfg.dot, boxShadow: `0 0 4px ${cfg.dot}` }}
      />
      {cfg.label}
    </span>
  )
}
