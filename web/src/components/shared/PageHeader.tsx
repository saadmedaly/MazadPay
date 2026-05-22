import type { LucideIcon } from 'lucide-react'
import { cn } from '@/lib/utils'

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
    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-7 animate-fade-in">
      <div className="flex items-center gap-3.5">
        {Icon && (
          <div
            className="w-10 h-10 rounded-xl flex items-center justify-center shrink-0"
            style={{
              background: 'rgba(59,130,246,0.12)',
              color: '#60A5FA',
              border: '1px solid rgba(59,130,246,0.2)',
              boxShadow: '0 4px 12px rgba(59,130,246,0.1)',
            }}
          >
            <Icon className="w-5 h-5" strokeWidth={2} />
          </div>
        )}
        <div>
          <h1
            className="text-[22px] font-black leading-tight"
            style={{ color: '#E2E8F0', fontFamily: 'Cairo' }}
          >
            {title}
          </h1>
          {subtitle && (
            <p className="text-[12.5px] font-medium mt-0.5" style={{ color: '#3D5A78' }}>
              {subtitle}
            </p>
          )}
        </div>
      </div>

      {(allActions.length > 0 || children) && (
        <div className="flex items-center gap-2 shrink-0">
          {allActions.map((act, idx) => (
            <button
              key={idx}
              onClick={act.onClick}
              className={cn(
                'btn',
                act.variant === 'outline' ? 'btn-outline'
                  : act.variant === 'ghost' ? 'btn-ghost'
                  : 'btn-primary'
              )}
            >
              <act.icon className="w-4 h-4" />
              {act.label}
            </button>
          ))}
          {children}
        </div>
      )}
    </div>
  )
}
