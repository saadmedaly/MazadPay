import type { LucideIcon } from 'lucide-react'

interface ActionItem {
  label: string
  icon: LucideIcon
  onClick: () => void
  variant?: 'primary' | 'outline' | 'ghost'
}

interface Props {
  title: string
  subtitle?: string
  icon?: LucideIcon
  actions?: ActionItem[]
  action?: ActionItem
  children?: React.ReactNode
}

export function PageHeader({ title, subtitle, icon: Icon, actions, action, children }: Props) {
  const allActions = [...(actions || []), ...(action ? [action] : [])]

  return (
    <div className="flex flex-col md:flex-row md:items-start justify-between gap-4 mb-8">
      <div className="flex items-start gap-4">
        {Icon && (
          <div className="w-12 h-12 rounded-2xl bg-surface-card border border-surface-border flex items-center justify-center text-mazad-primary shadow-xl shrink-0">
            <Icon className="w-6 h-6" />
          </div>
        )}
        <div>
          <h1 className="text-3xl font-display font-bold text-white tracking-tight">{title}</h1>
          {subtitle && <p className="text-sm text-surface-muted mt-1.5 font-medium">{subtitle}</p>}
        </div>
      </div>
      
      <div className="flex items-center gap-2">
        {allActions.map((act, idx) => (
          <button
            key={idx}
            onClick={act.onClick}
            className={`flex items-center gap-2 px-5 py-2.5 rounded-xl text-sm font-bold shadow-lg transition-all active:scale-95 ${
              act.variant === 'outline' 
                ? 'bg-transparent border border-surface-border text-surface-muted hover:text-white hover:border-white'
                : 'bg-mazad-primary text-white shadow-mazad-primary/20 hover:bg-mazad-primary-dk'
            }`}
          >
            <act.icon className="w-4 h-4" />
            {act.label}
          </button>
        ))}
        {children}
      </div>
    </div>
  )
}
