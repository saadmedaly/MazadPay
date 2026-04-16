import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

interface Props {
  size?: 'sm' | 'md' | 'lg'
  fullPage?: boolean
  className?: string
  label?: string
}

export function LoadingSpinner({ size = 'md', fullPage, className, label }: Props) {
  const sizeCls = {
    sm: 'w-4 h-4',
    md: 'w-8 h-8',
    lg: 'w-12 h-12',
  }[size]

  const content = (
    <div className={cn('flex flex-col items-center justify-center gap-3', className)}>
      <Loader2 className={cn(sizeCls, 'text-mazad-primary animate-spin')} />
      {label && <span className="text-xs font-bold text-surface-muted">{label}</span>}
    </div>
  )

  if (fullPage) {
    return (
      <div className="fixed inset-0 bg-surface-base/80 backdrop-blur-sm z-[100] flex items-center justify-center">
        {content}
      </div>
    )
  }

  return content
}
