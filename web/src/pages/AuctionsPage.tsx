import { useState } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { Eye, Check, X, Search, Gavel, Loader2, AlertCircle } from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
 
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { useAuctions, useValidateAuction } from '@/hooks/useAuctions'
import { formatPrice, formatDate, shortID } from '@/lib/formatters'
import { Input } from '@/components/ui/input'
import type { Auction } from '@/types/api'
import type { ColumnDef } from '@tanstack/react-table'
import { StatusBadge } from '@/components/shared/StatusBadge'
import { DataTable } from '@/components/shared/DataTable'

const STATUS_TABS = [
  { label: 'الكل',             value: '' },
  { label: 'في انتظار المراجعة', value: 'pending' },
  { label: 'المزادات النشطة',   value: 'active' },
  { label: 'المزادات المنتهية',  value: 'ended' },
  { label: 'المزادات المرفوضة',  value: 'rejected' },
]

export function AuctionsPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const [q, setQ] = useState('')
  const [rejectDialog, setRejectDialog] = useState<{ id: string; reason: string } | null>(null)
  const [approveId, setApproveId] = useState<string | null>(null)

  const status = searchParams.get('status') ?? ''
  const page = parseInt(searchParams.get('page') ?? '1')

  const { data, isLoading, isError } = useAuctions({ status: status || undefined, q, page, per_page: 25 })
  const validate = useValidateAuction()

  const columns: ColumnDef<Auction>[] = [
    {
      header: 'رقم القطعة',
      accessorKey: 'id',
      cell: ({ row }) => <span className="font-mono text-xs text-surface-muted font-bold">{row.original.lot_number ?? shortID(row.original.id)}</span>
    },
    {
      header: 'العنوان',
      accessorKey: 'title',
      cell: ({ getValue }) => <p className="text-white font-bold truncate max-w-[200px]">{getValue<string>()}</p>
    },
    {
      header: 'السعر الحالي',
      accessorKey: 'current_price',
      cell: ({ getValue }) => <span className="font-mono font-bold text-mazad-accent text-base">{formatPrice(parseFloat(getValue<string>()))}</span>
    },
    {
      header: 'تاريخ الانتهاء',
      accessorKey: 'end_time',
      cell: ({ getValue }) => <span className="text-xs text-surface-muted font-medium">{formatDate(getValue<string>())}</span>
    },
    {
      header: 'المزايدات',
      accessorKey: 'bidder_count',
      cell: ({ getValue }) => <span className="font-bold text-surface-muted">{getValue<number>()}</span>
    },
    {
      header: 'الحالة',
      accessorKey: 'status',
      cell: ({ getValue }) => <StatusBadge status={getValue<string>()} />
    },
    {
      header: 'الإجراءات',
      id: 'actions',
      cell: ({ row }) => {
        const auction = row.original
        return (
          <div className="flex items-center gap-2">
            <button
              onClick={() => navigate(`/auctions/${auction.id}`)}
              className="p-2 rounded-lg text-surface-muted hover:text-white hover:bg-surface-border border border-transparent hover:border-surface-border transition-all"
              title="عرض التفاصيل"
            >
              <Eye className="w-4 h-4" />
            </button>
            {auction.status === 'pending' && (
              <>
                <button
                  onClick={() => setApproveId(auction.id)}
                  className="p-2 rounded-lg text-emerald-400 hover:bg-emerald-500/10 border border-transparent hover:border-emerald-500/20 transition-all"
                  title="الموافقة وبث المزاد"
                >
                  <Check className="w-4 h-4" />
                </button>
                <button
                  onClick={() => setRejectDialog({ id: auction.id, reason: '' })}
                  className="p-2 rounded-lg text-red-400 hover:bg-red-500/10 border border-transparent hover:border-red-500/20 transition-all"
                  title="رفض المزاد"
                >
                  <X className="w-4 h-4" />
                </button>
              </>
            )}
          </div>
        )
      }
    }
  ]

  if (isError) return (
    <div className="admin-card p-20 text-center flex flex-col items-center gap-4">
      <AlertCircle className="w-12 h-12 text-red-500/20" />
      <p className="text-red-400 font-bold">فشل تحميل المزادات</p>
      <button onClick={() => window.location.reload()} className="bg-surface-border text-white px-6 py-2 rounded-xl text-sm font-bold hover:bg-surface-border/80 transition-all">إعادة المحاولة</button>
    </div>
  )

  return (
    <div className="animate-fade-in" dir="rtl">
      <PageHeader title="المزادات" subtitle={`${data?.total ?? 0} مزاد في المجمل`} />

      {/* Tabs */}
      <div className="flex gap-1 mb-6 bg-surface-card border border-surface-border rounded-xl p-1 w-fit">
        {STATUS_TABS.map((tab) => (
          <button
            key={tab.value}
            onClick={() => setSearchParams({ status: tab.value, page: '1' })}
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

      {/* Search */}
      <div className="relative mb-6 max-w-md group">
        <Search className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted group-focus-within:text-mazad-primary transition-colors" />
        <Input
          value={q}
          onChange={(e) => {
            setQ(e.target.value)
            setSearchParams((p) => { p.set('page', '1'); return p })
          }}
          placeholder="ابحث عن مزاد..."
          className="pr-10"
        />
      </div>

      <DataTable
        columns={columns}
        data={data?.data ?? []}
        isLoading={isLoading}
        total={data?.total}
        page={page}
        onPageChange={(p) => setSearchParams((prev) => { prev.set('page', p.toString()); return prev })}
        emptyTitle="لا توجد مزادات"
        emptyDescription="لم يتم العثور على أي مزادات تطابق الفلتر الحالي."
      />

      {/* Confirm Approve */}
      <ConfirmDialog
        open={!!approveId}
        onOpenChange={(v) => !v && setApproveId(null)}
        title="الموافقة على بث هذا المزاد؟"
        description="سيتم نشر المزاد فوراً وسيتمكن الجمهور من المزايدة عليه."
        confirmLabel="موافقة"
        variant="success"
        loading={validate.isPending}
        onConfirm={() => {
          if (approveId) validate.mutate(
            { id: approveId, approve: true },
            { onSuccess: () => setApproveId(null) }
          )
        }}
      />

      {/* Reject Dialog */}
      {rejectDialog && (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4 backdrop-blur-sm">
          <div className="admin-card p-8 w-full max-w-md animate-slide-in relative overflow-hidden">
            <div className="absolute top-0 right-0 w-24 h-24 bg-red-500/5 rounded-full -mr-12 -mt-12 blur-3xl" />
            
            <h3 className="font-display font-bold text-white text-xl mb-1 relative">رفض هذا المزاد</h3>
            <p className="text-sm text-surface-muted mb-6 font-medium relative italic">سيتم إرسال سبب الرفض إلى البائع.</p>
            
            <textarea
              value={rejectDialog.reason}
              onChange={(e) => setRejectDialog({ ...rejectDialog, reason: e.target.value })}
              placeholder="اكتب سبب الرفض هنا (إلزامي)..."
              rows={4}
              className="w-full bg-surface-base border border-surface-border rounded-xl
                         p-4 text-sm text-white placeholder:text-surface-muted/30
                         focus:outline-none focus:border-red-500/60 transition-all font-medium resize-none mb-6 shadow-inner"
            />
            
            <div className="flex gap-3 justify-end items-center">
              <button
                onClick={() => setRejectDialog(null)}
                className="px-5 py-2.5 rounded-xl text-sm font-bold text-surface-muted border border-surface-border hover:bg-surface-border/50 transition-all"
              >
                إلغاء
              </button>
              <button
                disabled={!rejectDialog.reason.trim() || validate.isPending}
                onClick={() => validate.mutate(
                  { id: rejectDialog.id, approve: false, reason: rejectDialog.reason },
                  { onSuccess: () => setRejectDialog(null) }
                )}
                className="px-6 py-2.5 rounded-xl text-sm font-bold bg-red-500 hover:bg-red-600
                           text-white disabled:opacity-40 transition-all shadow-lg shadow-red-500/20 flex items-center gap-2"
              >
                {validate.isPending && <Loader2 className="w-4 h-4 animate-spin" />}
                {validate.isPending ? 'جاري التنفيذ...' : 'تأكيد الرفض'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
