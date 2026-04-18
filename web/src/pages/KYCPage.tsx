import { useState } from 'react'
import { CheckCircle2, XCircle, Eye, ShieldCheck } from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { StatusBadge } from '@/components/shared/StatusBadge'
import { DataTable } from '@/components/shared/DataTable'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { useKYCs, useReviewKYC } from '@/hooks/useKYC'
import { formatDate, shortID } from '@/lib/formatters'
import type { KYCVerification } from '@/types/api'
import type { ColumnDef } from '@tanstack/react-table'

export function KYCPage() {
  const [status, setStatus] = useState('pending')
  const [reviewTarget, setReviewTarget] = useState<{ userId: string; name: string; status: 'approved' | 'rejected' } | null>(null)
  const [notes, setNotes] = useState('')

  const { data: kycs, isLoading } = useKYCs(status)
  const reviewKYC = useReviewKYC()

  const columns: ColumnDef<KYCVerification>[] = [
    {
      header: 'المستخدم',
      accessorKey: 'user.full_name',
      cell: ({ row }) => (
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-surface-border flex items-center justify-center text-[10px] font-bold">
            {(row.original.user?.full_name ?? 'U')[0]}
          </div>
          <div>
            <p className="text-white text-xs font-bold">{row.original.user?.full_name ?? 'مستخدم غير معروف'}</p>
            <p className="text-[10px] text-surface-muted font-mono">{shortID(row.original.user_id)}</p>
          </div>
        </div>
      )
    },
    {
      header: 'رقم التعريف NNI',
      accessorKey: 'nni_number',
      cell: ({ getValue }) => <span className="font-mono text-xs">{getValue<string>() ?? '---'}</span>
    },
    {
      header: 'تاريخ الطلب',
      accessorKey: 'created_at',
      cell: ({ getValue }) => <span className="text-xs text-surface-muted">{formatDate(getValue<string>())}</span>
    },
    {
      header: 'الحالة',
      accessorKey: 'status',
      cell: ({ getValue }) => <StatusBadge status={getValue<string>() as any} />
    },
    {
      header: 'الوثائق',
      id: 'documents',
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          {row.original.id_card_front_url && (
            <a href={row.original.id_card_front_url} target="_blank" rel="noreferrer" 
               className="p-1.5 rounded bg-surface-border/50 text-surface-muted hover:text-white transition-colors">
              <Eye className="w-3.5 h-3.5" />
            </a>
          )}
          <span className="text-[10px] text-surface-muted">البطاقة الشخصية</span>
        </div>
      )
    },
    {
      header: 'الإجراءات',
      id: 'actions',
      cell: ({ row }) => {
        if (row.original.status !== 'pending') return null
        return (
          <div className="flex items-center gap-2">
            <button
              onClick={() => setReviewTarget({ userId: row.original.user_id, name: row.original.user?.full_name ?? 'المستخدم', status: 'approved' })}
              className="p-2 rounded-lg text-emerald-400 hover:bg-emerald-500/10 transition-colors"
              title="قبول التوثيق"
            >
              <CheckCircle2 className="w-4 h-4" />
            </button>
            <button
              onClick={() => setReviewTarget({ userId: row.original.user_id, name: row.original.user?.full_name ?? 'المستخدم', status: 'rejected' })}
              className="p-2 rounded-lg text-red-400 hover:bg-red-500/10 transition-colors"
              title="رفض التوثيق"
            >
              <XCircle className="w-4 h-4" />
            </button>
          </div>
        )
      }
    }
  ]

  return (
    <div className="animate-fade-in" dir="rtl">
      <PageHeader 
        title="توثيق الحسابات  " 
        subtitle="مراجعة وثائق الهوية وتفعيل حسابات المستخدمين"
        icon={ShieldCheck}
      />

      <div className="flex items-center gap-2 mb-6">
        {['pending', 'approved', 'rejected'].map((s) => (
          <button
            key={s}
            onClick={() => setStatus(s)}
            className={`px-4 py-2 rounded-xl text-xs font-bold transition-all border ${
              status === s 
                ? 'bg-mazad-primary text-white border-mazad-primary shadow-lg shadow-mazad-primary/20' 
                : 'bg-surface-card text-surface-muted border-surface-border hover:text-white'
            }`}
          >
            {s === 'pending' ? 'بانتظار المراجعة' : s === 'approved' ? 'موثقة' : 'مرفوضة'}
          </button>
        ))}
      </div>

      <DataTable
        columns={columns}
        data={kycs ?? []}
        isLoading={isLoading}
        total={kycs?.length}
        emptyTitle="لا توجد طلبات"
        emptyDescription="ليس هناك أي طلبات توثيق في هذا القسم حالياً."
      />

      <ConfirmDialog
        open={!!reviewTarget}
        onOpenChange={(v) => {
          if (!v) {
            setReviewTarget(null)
            setNotes('')
          }
        }}
        title={reviewTarget?.status === 'approved' ? 'قبول توثيق الحساب' : 'رفض طلب التوثيق'}
        description={
          <div className="space-y-4 pt-2 text-right" dir="rtl">
            <p>هل أنت متأكد من {reviewTarget?.status === 'approved' ? 'قبول' : 'رفض'} طلب التوثيق للمستخدم <span className="text-white font-bold">{reviewTarget?.name}</span>؟</p>
            <div className="space-y-2">
              <label className="text-xs text-surface-muted block font-bold">ملاحظات إضافية (اختياري)</label>
              <textarea
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                className="w-full bg-surface-bg border border-surface-border rounded-xl p-3 text-xs text-white focus:outline-none focus:border-mazad-primary min-h-[100px]"
                placeholder={reviewTarget?.status === 'approved' ? 'مثال: تم التحقق من البيانات' : 'مثال: الصورة غير واضحة...'}
              />
            </div>
          </div>
        }
        confirmLabel={reviewTarget?.status === 'approved' ? 'تأكيد القبول' : 'تأكيد الرفض'}
        variant={reviewTarget?.status === 'approved' ? 'success' : 'danger'}
        loading={reviewKYC.isPending}
        onConfirm={() => {
          if (reviewTarget) {
            reviewKYC.mutate(
              { userId: reviewTarget.userId, status: reviewTarget.status, notes },
              { onSuccess: () => { setReviewTarget(null); setNotes('') } }
            )
          }
        }}
      />
    </div>
  )
}
