import { useState } from 'react'
import { Star, Trash2, MessageSquare, User, Calendar, AlertCircle } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { PageHeader } from '@/components/shared/PageHeader'
import { EmptyState } from '@/components/shared/EmptyState'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { formatDate } from '@/lib/formatters'
import client from '@/api/client'

interface Rating {
  id: string
  user_id: string
  title: string | null
  rating: number
  comment: string | null
  created_at: string
  user_name?: string | null
}

function useRatings(page: number) {
  return useQuery({
    queryKey: ['app-ratings', page],
    queryFn: async () => {
      const { data } = await client.get('/v1/api/admin/app/ratings', { params: { page, limit: 20 } })
      return data.data
    },
  })
}

function useStats() {
  return useQuery({
    queryKey: ['app-ratings-stats'],
    queryFn: async () => {
      const { data } = await client.get('/v1/api/admin/app/ratings/stats')
      return data.data
    },
  })
}

export function RatingsPage() {
  const [page, setPage] = useState(1)
  const qc = useQueryClient()
  const { data, isLoading, isError } = useRatings(page)
  const { data: stats } = useStats()
  const [deleteId, setDeleteId] = useState<string | null>(null)

  const deleteMutation = useMutation({
    mutationFn: (id: string) => client.delete(`/v1/api/admin/app/ratings/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['app-ratings'] })
      qc.invalidateQueries({ queryKey: ['app-ratings-stats'] })
      toast.success('تم حذف التقييم بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })

  if (isError) return (
    <div className="admin-card p-20 text-center flex flex-col items-center gap-4">
      <AlertCircle className="w-12 h-12 text-red-500/20" />
      <p className="text-red-400 font-bold">فشل تحميل التقييمات</p>
      <button onClick={() => window.location.reload()} className="bg-surface-border text-white px-6 py-2 rounded-xl text-sm font-bold">إعادة المحاولة</button>
    </div>
  )

  const ratings = data?.ratings || []
  const total = data?.total || 0

  return (
    <div className="animate-fade-in" dir="rtl">
      <PageHeader title="تقييمات التطبيق" subtitle="مراجعة آراء وتقييمات المستخدمين للتطبيق">
        {stats && (
           <div className="flex gap-4">
              <div className="bg-surface-card border border-surface-border px-4 py-2 rounded-xl flex items-center gap-3">
                 <div className="bg-yellow-500/10 p-2 rounded-lg"><Star className="w-4 h-4 text-yellow-500 fill-yellow-500" /></div>
                 <div>
                    <p className="text-[10px] text-surface-muted font-bold uppercase">متوسط التقييم</p>
                    <p className="text-lg font-display font-bold text-white">{stats.average_rating?.toFixed(1) || 0}</p>
                 </div>
              </div>
              <div className="bg-surface-card border border-surface-border px-4 py-2 rounded-xl flex items-center gap-3">
                 <div className="bg-mazad-primary/10 p-2 rounded-lg"><MessageSquare className="w-4 h-4 text-mazad-primary" /></div>
                 <div>
                    <p className="text-[10px] text-surface-muted font-bold uppercase">إجمالي التقييمات</p>
                    <p className="text-lg font-display font-bold text-white">{stats.total_ratings || 0}</p>
                 </div>
              </div>
           </div>
        )}
      </PageHeader>

      <div className="space-y-4">
        {isLoading ? (
          <div className="admin-card p-20 text-center"><LoadingSpinner /></div>
        ) : ratings.length === 0 ? (
          <div className="admin-card">
            <EmptyState icon={Star} title="لا توجد تقييمات" description="لم يقم أي مستخدم بتقييم التطبيق بعد." />
          </div>
        ) : (
          ratings.map((r: Rating) => (
            <div key={r.id} className="admin-card p-6 hover:border-surface-border/60 transition-all group">
              <div className="flex items-start justify-between gap-6">
                <div className="flex-1">
                  <div className="flex items-center gap-4 mb-4">
                    <div className="flex gap-0.5">
                      {[1, 2, 3, 4, 5].map((s) => (
                        <Star key={s} className={`w-4 h-4 ${s <= r.rating ? 'text-yellow-500 fill-yellow-500' : 'text-surface-border'}`} />
                      ))}
                    </div>
                    <div className="h-4 w-px bg-surface-border" />
                    <div className="flex items-center gap-2 text-surface-muted text-xs font-bold">
                       <Calendar className="w-3 h-3" />
                       {formatDate(r.created_at)}
                    </div>
                  </div>
                  
                  <h4 className="text-white font-bold text-lg mb-2">{r.title || 'بدون عنوان'}</h4>
                  <p className="text-surface-muted text-sm leading-relaxed mb-4">{r.comment || 'لا يوجد تعليق'}</p>
                  
                  <div className="flex items-center gap-2 text-[10px] text-surface-muted uppercase font-bold tracking-wider">
                     <User className="w-3 h-3" />
                     المستخدم: <span className="text-white">{r.user_name || r.user_id}</span>
                  </div>
                </div>

                <button
                  onClick={() => setDeleteId(r.id)}
                  className="p-3 rounded-xl bg-red-500/10 text-red-400 hover:bg-red-500/20 transition-all opacity-0 group-hover:opacity-100"
                  title="حذف التقييم"
                >
                  <Trash2 className="w-5 h-5" />
                </button>
              </div>
            </div>
          ))
        )}
      </div>

      {total > 20 && (
        <div className="flex justify-center gap-2 mt-8">
           {/* Simple pagination could be added here */}
        </div>
      )}

      <ConfirmDialog
        open={!!deleteId}
        onOpenChange={(v) => !v && setDeleteId(null)}
        title="حذف التقييم؟"
        description="هذا الإجراء سيؤدي لحذف التقييم نهائياً من قاعدة البيانات."
        confirmLabel="حذف"
        variant="danger"
        onConfirm={() => deleteId && deleteMutation.mutate(deleteId, { onSuccess: () => setDeleteId(null) })}
      />
    </div>
  )
}
