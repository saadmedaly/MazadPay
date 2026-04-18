import type { LucideIcon } from 'lucide-react'

interface Props {
  icon?: LucideIcon
  title?: string
  description?: string
  action?: React.ReactNode
}

export function EmptyState({ icon: Icon, title, description, action }: Props) {
  return (
    <div className="flex flex-col items-center justify-center py-20 text-center animate-fade-in" dir="rtl">
      {Icon && (
        <div className="p-5 rounded-full bg-surface-card border border-surface-border mb-6 shadow-xl shadow-black/20">
          <Icon className="w-10 h-10 text-surface-muted" />
        </div>
      )}
      <h3 className="font-display font-bold text-white text-lg mb-2 tracking-tight">{title || 'لا يوجد بيانات'}</h3>
      {description && <p className="text-sm text-surface-muted max-w-sm font-medium leading-relaxed">{description}</p>}
      {action && <div className="mt-8">{action}</div>}
    </div>
  )
}
