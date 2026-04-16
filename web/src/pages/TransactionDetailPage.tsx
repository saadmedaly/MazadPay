import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft, ZoomIn, Check, X, AlertTriangle, User, Calendar, CreditCard, AlertCircle } from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { StatusBadge } from '@/components/shared/StatusBadge'
 import { ImagePreview } from '@/components/shared/ImagePreview'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { useTransaction, useValidateTransaction } from '@/hooks/useTransactions'
import { formatPrice, formatDate, shortID } from '@/lib/formatters'
import { GATEWAY_LABELS } from '@/lib/constants'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'

export function TransactionDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [notes, setNotes] = useState('')
  const [confirmAction, setConfirmAction] = useState<'approve' | 'reject' | null>(null)
  const [imgZoomed, setImgZoomed] = useState(false)

  const { data: txn, isLoading, isError } = useTransaction(id!)
  const validate = useValidateTransaction()

  const handleValidate = (approve: boolean) => {
    validate.mutate(
      { id: id!, approve, notes },
      { onSuccess: () => navigate('/transactions') }
    )
  }

  if (isLoading) return <LoadingSpinner fullPage label="جاري تحميل المعاملة..." />

  if (isError || !txn) return (
    <div className="flex flex-col items-center justify-center h-64 text-surface-muted gap-4">
      <AlertCircle className="w-12 h-12 opacity-20" />
      <p className="font-bold">فشل في العثور على المعاملة</p>
      <button onClick={() => navigate(-1)} className="text-mazad-primary text-sm font-bold">رجوع للوراء</button>
    </div>
  )

  const isPendingReview = txn.status === 'pending_review'

  return (
    <div className="animate-fade-in max-w-5xl" dir="rtl">
      <PageHeader title="تفاصيل المعاملة">
        <button
          onClick={() => navigate(-1)}
          className="flex items-center gap-2 text-sm font-bold text-surface-muted hover:text-white transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          رجوع
        </button>
      </PageHeader>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Transaction Infos */}
        <div className="admin-card p-6 border-surface-border">
          <h2 className="font-display font-bold text-white text-base mb-6 pb-4 border-b border-surface-border flex items-center gap-2">
            <CreditCard className="w-4 h-4 text-mazad-primary" />
             المعلومات الأساسية
          </h2>
          <div className="space-y-4">
            {[
              { label: 'رقم المعاملة',  value: <span className="font-mono font-bold">{shortID(txn.id)}</span>,    icon: null },
              { label: 'المستخدم',    value: <span className="font-mono font-bold cursor-pointer hover:text-mazad-primary transition-colors" onClick={() => navigate(`/users/${txn.user_id}`)}>{shortID(txn.user_id)}</span>, icon: User },
              { label: 'المبلغ',      value: <span className="font-bold text-mazad-accent text-2xl">{formatPrice(parseFloat(txn.amount))}</span>, icon: CreditCard },
              { label: 'بوابة الدفع',   value: txn.gateway ? (GATEWAY_LABELS[txn.gateway] ?? txn.gateway) : '—', icon: null },
              { label: 'التاريخ',      value: <span className="font-medium">{formatDate(txn.created_at)}</span>, icon: Calendar },
              { label: 'الحالة',       value: <StatusBadge status={txn.status} />, icon: null },
            ].map(({ label, value }) => (
              <div key={label} className="flex items-center justify-between gap-4 py-2 border-b border-surface-border/30 last:border-0">
                <span className="text-xs text-surface-muted font-bold shrink-0">{label}</span>
                <span className="text-sm text-white text-left">{value}</span>
              </div>
            ))}
          </div>

          {/* Admin Notes (if already processed) */}
          {txn.admin_notes && (
            <div className="mt-6 pt-6 border-t border-surface-border">
              <p className="text-xs text-surface-muted font-bold mb-2">ملاحظات المسؤول:</p>
              <div className="text-sm text-white bg-surface-base/50 rounded-xl p-4 border border-surface-border font-medium italic">
                {txn.admin_notes}
              </div>
            </div>
          )}
        </div>

        {/* Payment Receipt */}
        <div className="admin-card p-6">
          <h2 className="font-display font-bold text-white text-base mb-6 pb-4 border-b border-surface-border flex items-center gap-2">
            <ZoomIn className="w-4 h-4 text-mazad-primary" />
             إيصال الدفع
          </h2>
          <ImagePreview 
            src={txn.receipt_url} 
            alt="إيصال الدفع" 
            className="aspect-square bg-black/20"
          />
        </div>
      </div>

      {/* Validation Panel */}
      {isPendingReview && (
        <div className="admin-card p-8 mt-6 border-orange-500/30 relative overflow-hidden shadow-2xl shadow-orange-500/5">
          <div className="absolute top-0 right-0 w-32 h-32 bg-orange-500/5 rounded-full -mr-16 -mt-16 blur-3xl" />
          
          <div className="flex items-center gap-3 mb-6 relative">
            <div className="p-2 bg-orange-500/10 rounded-lg">
              <AlertTriangle className="w-5 h-5 text-orange-400" />
            </div>
            <div>
              <h2 className="font-display font-bold text-orange-400 text-lg">مراجعة المعاملة مطلوبة</h2>
              <p className="text-xs text-orange-400/70 font-medium">يرجى التأكد من مطابقة الإيصال مع المبلغ المطلوب.</p>
            </div>
          </div>

          <div className="mb-8 relative">
            <label className="text-xs font-bold text-surface-muted uppercase tracking-widest block mb-3">
               ملاحظات المسؤول (إلزامية عند الرفض، اختيارية عند القبول)
            </label>
            <textarea
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="مثال: تم التأكد من الإيصال - المبلغ صحيح | أو | الإيصال غير واضح، يرجى إعادة الرفض..."
              rows={3}
              className="w-full bg-surface-base border border-surface-border rounded-xl p-4
                         text-sm text-white placeholder:text-surface-muted/30 focus:outline-none
                         focus:border-mazad-primary transition-all font-medium resize-none shadow-inner"
            />
          </div>

          <div className="flex gap-4 relative">
            <button
              onClick={() => setConfirmAction('approve')}
              className="flex-1 flex items-center justify-center gap-3 py-3 rounded-xl text-sm font-bold
                         bg-emerald-500 hover:bg-emerald-600 text-white transition-all shadow-lg shadow-emerald-500/20"
            >
              <Check className="w-5 h-5" />
               تأكيد الإيداع وشحن المحفظة
            </button>
            <button
              disabled={!notes.trim()}
              onClick={() => setConfirmAction('reject')}
              className="flex-1 flex items-center justify-center gap-3 py-3 rounded-xl text-sm font-bold
                         bg-red-500 hover:bg-red-600 text-white transition-all shadow-lg shadow-red-500/20 disabled:opacity-40"
            >
              <X className="w-5 h-5" />
               رفض المعاملة
            </button>
          </div>
          {!notes.trim() && (
            <p className="text-[10px] text-surface-muted mt-3 text-center font-bold">
               يجب كتابة ملاحظة لتتمكن من رفض المعاملة
            </p>
          )}
        </div>
      )}

      {/* Confirm Dialogs */}
      <ConfirmDialog
        open={confirmAction === 'approve'}
        onOpenChange={(v) => !v && setConfirmAction(null)}
        title={`هل أنت متأكد من الموافقة على مبلغ ${formatPrice(parseFloat(txn.amount))}؟`}
        description="هذا الإجراء سيقوم بشحن محفظة المستخدم فوراً ولا يمكن التراجع عنه."
        confirmLabel="تأكيد الموافقة"
        variant="success"
        loading={validate.isPending}
        onConfirm={() => handleValidate(true)}
      />
      <ConfirmDialog
        open={confirmAction === 'reject'}
        onOpenChange={(v) => !v && setConfirmAction(null)}
        title="هل أنت متأكد من رفض هذا الإيداع؟"
        description="لن يتم شحن المحفظة وسيتلقى المستخدم إشعاراً بملاحظاتك."
        confirmLabel="تأكيد الرفض"
        variant="danger"
        loading={validate.isPending}
        onConfirm={() => handleValidate(false)}
      />
    </div>
  )
}
