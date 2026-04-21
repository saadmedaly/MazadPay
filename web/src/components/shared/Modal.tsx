import React from 'react'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ModalProps {
  isOpen: boolean
  onOpenChange: (open: boolean) => void
  title: string
  children: React.ReactNode
  size?: 'sm' | 'md' | 'lg' | 'xl'
  showCloseButton?: boolean
}

export function Modal({
  isOpen,
  onOpenChange,
  title,
  children,
  size = 'md',
  showCloseButton = true
}: ModalProps) {
  if (!isOpen) return null

  const sizeClasses = {
    sm: 'max-w-md',
    md: 'max-w-lg',
    lg: 'max-w-2xl',
    xl: 'max-w-4xl'
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-black/80 backdrop-blur-sm"
        onClick={() => onOpenChange(false)}
      />
      
      {/* Modal */}
      <div className={cn(
        "relative bg-surface-card border border-surface-border rounded-2xl shadow-2xl w-full",
        sizeClasses[size],
        "animate-in fade-in zoom-in duration-200"
      )}>
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-surface-border">
          <h2 className="text-lg font-bold text-white">{title}</h2>
          {showCloseButton && (
            <button
              onClick={() => onOpenChange(false)}
              className="p-2 text-surface-muted hover:text-white transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          )}
        </div>
        
        {/* Content */}
        <div className="p-6">
          {children}
        </div>
      </div>
    </div>
  )
}
