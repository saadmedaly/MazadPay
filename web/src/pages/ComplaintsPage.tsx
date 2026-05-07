import { useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { MessageSquare, CheckCircle2, XCircle, AlertCircle, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { PageHeader } from '@/components/shared/PageHeader'
import { StatusBadge } from '@/components/shared/StatusBadge'
import { EmptyState } from '@/components/shared/EmptyState'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { formatDate } from '@/lib/formatters'
import client from '@/api/client'

interface Report {
  id: string
  auction_id: string | null
  reporter_id: string
  reason: string
  status: 'pending' | 'reviewed' | 'dismissed'
  created_at: string
  admin_notes: string | null
}

function useComplaints(status: string) {
  return useQuery({
    queryKey: ['complaints', status],
    queryFn: async () => {
      const { data } = await client.get('/v1/api/admin/reports', { params: { status: status || undefined, type: 'app' } })
      return data.data as Report[]
    },
    refetchInterval: 60_000,
  })
}

export function ComplaintsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const status = searchParams.get('status') ?? 'pending'
  const qc = useQueryClient()
  const [deleteId, setDeleteId] = useState<string | null>(null)

  const { data: complaints = [], isLoading, isError } = useComplaints(status)

  const reviewMutation = useMutation({
    mutationFn: ({ id, action, notes }: { id: string; action: 'reviewed' | 'dismissed'; notes?: string }) =>
      client.put(`/v1/api/admin/reports/${id}/review`, { status: action, notes }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['complaints'] })
      toast.success('تم تحديث حالة الشكوى بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => client.delete(`/v1/api/admin/reports/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['complaints'] })
      toast.success('تم حذف الشكوى بنجاح')
      setDeleteId(null)
    },
    onError: (err: Error) => toast.error(err.message),
  })

  const STATUS_TABS = [
    { label: 'بانتظار المراجعة', value: 'pending'  },
    { label: 'تمت معالجتها',    value: 'reviewed' },
    { label: 'تجاهلها',         value: 'dismissed' },
  ]

  if (isError) return (
    <div className="admin-card p-20 text-center flex flex-col items-center gap-4">
      <AlertCircle className="w-12 h-12 text-red-500/20" />
      <p className="text-red-400 font-bold">فشل تحميل الشكاوى</p>
      <button onClick={() => window.location.reload()} className="bg-surface-border text-white px-6 py-2 rounded-xl text-sm font-bold">إعادة المحاولة</button>
    </div>
  )

  return (
    <div className="animate-fade-in" dir="rtl">
      <PageHeader title="شكاوى المستخدمين" subtitle="مراجعة الشكاوى والمقترحات المرسلة من قبل المستخدمين" />

      <div className="flex gap-1 mb-8 bg-surface-card border border-surface-border rounded-xl p-1 w-fit">
        {STATUS_TABS.map((tab) => (
          <button
            key={tab.value}
            onClick={() => setSearchParams({ status: tab.value })}
            className={`px-4 py-2 rounded-lg text-xs font-bold transition-all ${
              status === tab.value
                ? 'bg-mazad-primary text-white shadow-lg shadow-mazad-primary/20'
                : 'text-surface-muted hover:text-white hover:bg-surface-border/50'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <div className="space-y-4">
        {isLoading ? (
          <div className="admin-card p-20 text-center"><LoadingSpinner label="جاري تحميل الشكاوى..." /></div>
        ) : complaints.length === 0 ? (
          <div className="admin-card">
            <EmptyState icon={MessageSquare} title="لا توجد شكاوى" description="لا توجد أي شكاوى في هذا القسم حالياً." />
          </div>
        ) : (
          complaints.map((report) => (
            <div key={report.id} className="admin-card p-6 hover:border-surface-border/60 transition-colors group">
              <div className="flex items-start justify-between gap-6">
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-3">
                    <StatusBadge status={report.status} />
                    <span className="text-xs text-surface-muted font-bold font-mono">{formatDate(report.created_at)}</span>
                  </div>
                  
                  <div className="bg-surface-base/50 rounded-xl p-4 border border-surface-border mb-3 font-medium text-sm text-white/90">
                     {report.reason}
                  </div>
                  
                  <div className="flex items-center gap-2 text-[10px] text-surface-muted uppercase font-bold">
                     صاحب الشكوى: <span className="text-white font-mono">{report.reporter_id}</span>
                  </div>

                  {report.admin_notes && (
                    <p className="text-xs text-surface-muted mt-3 py-2 px-3 bg-surface-base/30 rounded-lg flex items-center gap-2 font-medium italic border-r-2 border-surface-muted">
                      ملاحظة المسؤول: {report.admin_notes}
                    </p>
                  )}
                </div>
                
                  <div className="flex flex-col gap-2 shrink-0">
                    {report.status === 'pending' && (
                      <>
                        <button
                          onClick={() => reviewMutation.mutate({ id: report.id, action: 'reviewed' })}
                          className="flex items-center justify-center gap-2 px-6 py-2.5 rounded-xl text-xs font-bold bg-mazad-primary text-white hover:bg-mazad-primary-dk transition-all"
                        >
                          <CheckCircle2 className="w-4 h-4" /> معالجة
                        </button>
                        <button
                          onClick={() => reviewMutation.mutate({ id: report.id, action: 'dismissed' })}
                          className="flex items-center justify-center gap-2 px-6 py-2.5 rounded-xl text-xs font-bold border border-surface-border text-surface-muted hover:text-white"
                        >
                          <XCircle className="w-4 h-4" /> تجاهل
                        </button>
                      </>
                    )}
                    <button
                      onClick={() => setDeleteId(report.id)}
                      className="flex items-center justify-center gap-2 px-6 py-2.5 rounded-xl text-xs font-bold border border-red-500/20 text-red-400 hover:bg-red-500/10 transition-all"
                    >
                      <Trash2 className="w-4 h-4" /> حذف
                    </button>
                  </div>
              </div>
            </div>
          ))
        )}
      </div>
      <ConfirmDialog
        open={!!deleteId}
        onOpenChange={(v) => !v && setDeleteId(null)}
        title="حذف الشكوى؟"
        description="هذا الإجراء سيؤدي لحذف الشكوى نهائياً."
        confirmLabel="حذف"
        variant="danger"
        onConfirm={() => deleteId && deleteMutation.mutate(deleteId)}
      />
    </div>
  )
}
