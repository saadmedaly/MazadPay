import { useSearchParams } from 'react-router-dom'
import { Flag, CheckCircle2, XCircle, AlertCircle } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { PageHeader } from '@/components/shared/PageHeader'
import { StatusBadge } from '@/components/shared/StatusBadge'
import { EmptyState } from '@/components/shared/EmptyState'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { formatDate, shortID } from '@/lib/formatters'
import client from '@/api/client'

interface Report {
  id: string
  auction_id: string
  reporter_id: string
  reason: string
  status: 'pending' | 'reviewed' | 'dismissed'
  created_at: string
  admin_notes: string | null
}

function useReports(status: string) {
  return useQuery({
    queryKey: ['reports', status],
    queryFn: async () => {
      const { data } = await client.get('/v1/api/admin/reports', { params: { status: status || undefined } })
      return data.data as Report[]
    },
    refetchInterval: 60_000,
  })
}

export function ReportsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const status = searchParams.get('status') ?? 'pending'
  const qc = useQueryClient()

  const { data: reports = [], isLoading, isError } = useReports(status)

  const reviewMutation = useMutation({
    mutationFn: ({ id, action, notes }: { id: string; action: 'reviewed' | 'dismissed'; notes?: string }) =>
      client.put(`/v1/api/admin/reports/${id}/review`, { status: action, notes }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['reports'] })
      qc.invalidateQueries({ queryKey: ['dashboard'] })
      toast.success('تم تحديث حالة البلاغ بنجاح')
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
      <p className="text-red-400 font-bold">فشل تحميل البلاغات</p>
      <button onClick={() => window.location.reload()} className="bg-surface-border text-white px-6 py-2 rounded-xl text-sm font-bold">إعادة المحاولة</button>
    </div>
  )

  return (
    <div className="animate-fade-in" dir="rtl">
      <PageHeader title="بلاغات المخالفات" subtitle="مراجعة البلاغات المقدمة من المستخدمين على المزادات" />

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
          <div className="admin-card p-20 text-center">
             <LoadingSpinner label="جاري تحميل البلاغات..." />
          </div>
        ) : reports.length === 0 ? (
          <div className="admin-card">
            <EmptyState
              icon={Flag}
              title="لا توجد بلاغات"
              description="لا توجد أي بلاغات في هذا القسم حالياً."
            />
          </div>
        ) : (
          reports.map((report) => (
            <div key={report.id} className="admin-card p-6 hover:border-surface-border/60 transition-colors group">
              <div className="flex items-start justify-between gap-6">
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-3">
                    <StatusBadge status={report.status} />
                    <span className="text-xs text-surface-muted font-bold font-mono">{formatDate(report.created_at)}</span>
                  </div>
                  <p className="text-sm text-white font-bold mb-2 flex items-center gap-2">
                     المزاد المبلغ عنه: 
                     <span className="font-mono text-mazad-accent bg-mazad-accent/5 px-2 py-0.5 rounded border border-mazad-accent/10">
                        {shortID(report.auction_id)}
                     </span>
                  </p>
                  <div className="bg-surface-base/50 rounded-xl p-4 border border-surface-border mb-3 font-medium text-sm text-white/90">
                     {report.reason}
                  </div>
                  {report.admin_notes && (
                    <p className="text-xs text-surface-muted mt-3 py-2 px-3 bg-surface-base/30 rounded-lg flex items-center gap-2 font-medium italic border-r-2 border-surface-muted">
                      ملاحظة المسؤول: {report.admin_notes}
                    </p>
                  )}
                </div>
                
                {report.status === 'pending' && (
                  <div className="flex flex-col gap-2 shrink-0">
                    <button
                      onClick={() => reviewMutation.mutate({ id: report.id, action: 'reviewed' })}
                      className="flex items-center justify-center gap-2 px-6 py-2.5 rounded-xl text-xs font-bold
                                 bg-mazad-primary text-white hover:bg-mazad-primary-dk transition-all shadow-lg shadow-mazad-primary/20"
                    >
                      <CheckCircle2 className="w-4 h-4" />
                      معالجة
                    </button>
                    <button
                      onClick={() => reviewMutation.mutate({ id: report.id, action: 'dismissed' })}
                      className="flex items-center justify-center gap-2 px-6 py-2.5 rounded-xl text-xs font-bold border border-surface-border
                                 text-surface-muted hover:text-white hover:bg-red-500/10 hover:border-red-500/20 transition-all font-bold"
                    >
                      <XCircle className="w-4 h-4" />
                      تجاهل
                    </button>
                  </div>
                )}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
