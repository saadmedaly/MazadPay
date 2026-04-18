import {
  AlertDialog, AlertDialogAction, AlertDialogCancel,
  AlertDialogContent, AlertDialogDescription,
  AlertDialogFooter, AlertDialogHeader, AlertDialogTitle
} from '@/components/ui/alert-dialog'
import { cn } from '@/lib/utils'

interface Props {
  open: boolean
  onOpenChange: (v: boolean) => void
  title: string
  description: React.ReactNode
  confirmLabel?: string
  cancelLabel?: string
  variant?: 'danger' | 'success' | 'default'
  onConfirm: () => void
  loading?: boolean
}

export function ConfirmDialog({
  open, onOpenChange, title, description,
  confirmLabel = 'تأكيد', cancelLabel = 'إلغاء',
  variant = 'default', onConfirm, loading
}: Props) {
  const btnCls = {
    danger:  'bg-red-500 hover:bg-red-600 text-white',
    success: 'bg-emerald-500 hover:bg-emerald-600 text-white',
    default: 'bg-mazad-primary hover:bg-mazad-primary-dk text-white',
  }[variant]

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent className="bg-surface-card border-surface-border" dir="rtl">
        <AlertDialogHeader>
          <AlertDialogTitle className="text-white font-display text-right">{title}</AlertDialogTitle>
          <AlertDialogDescription className="text-surface-muted text-right">{description}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter className="flex-row-reverse gap-3">
          <AlertDialogCancel className="border-surface-border text-surface-muted hover:bg-surface-border hover:text-white mt-0">
            {cancelLabel}
          </AlertDialogCancel>
          <AlertDialogAction
            onClick={(e) => {
              e.preventDefault()
              onConfirm()
            }}
            disabled={loading}
            className={cn(btnCls, 'border-0', loading && 'opacity-50 cursor-not-allowed')}
          >
            {loading ? 'جاري التنفيذ...' : confirmLabel}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
