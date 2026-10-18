import { useNavigate, useSearchParams } from 'react-router-dom'
import { AlertCircle, AlertTriangle, RefreshCw, Zap } from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { DataTable } from '@/components/shared/DataTable'
import { useBankilyPayments, useRecheckBankilyPayment } from '@/hooks/useBankilyPayments'
import { formatPrice, formatDate, shortID } from '@/lib/formatters'
import type { BankilyPayment, BankilyPaymentStatus } from '@/api/bankilyPayments'
import type { ColumnDef } from '@tanstack/react-table'

const STATUS_LABELS: Record<BankilyPaymentStatus, { label: string; cls: string }> = {
  INITIATED:            { label: 'قيد الإنشاء',  cls: 'bg-slate-500/15 text-slate-300' },
  SUBMITTED:            { label: 'مُرسلة',       cls: 'bg-sky-500/15 text-sky-300' },
  PENDING_VERIFICATION: { label: 'قيد التحقق',   cls: 'bg-amber-500/15 text-amber-300' },
  SUCCESS:              { label: 'ناجحة',        cls: 'bg-emerald-500/15 text-emerald-300' },
  FAILED:               { label: 'فاشلة',        cls: 'bg-red-500/15 text-red-300' },
}

const TABS = [
  { label: 'الكل',       value: '' },
  { label: 'قيد التحقق', value: 'PENDING_VERIFICATION' },
  { label: 'ناجحة',      value: 'SUCCESS' },
  { label: 'فاشلة',      value: 'FAILED' },
]

// Bankily B-PAY automatic top-ups. Confirmed payments are credited by the
// backend exactly once and never need manual approval; the only action here
// is "re-check now", which asks Bankily (checkTransaction) and never credits
// a wallet by itself.
export function BankilyPaymentsPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const status = searchParams.get('status') ?? ''
  const page = parseInt(searchParams.get('page') ?? '1')
  const { data, isLoading, isError } = useBankilyPayments({ status: status || undefined, page, per_page: 25 })
  const recheck = useRecheckBankilyPayment()

  const columns: ColumnDef<BankilyPayment>[] = [
    {
      header: 'رقم العملية',
      accessorKey: 'operation_id',
      cell: ({ row }) => (
        <div className="flex flex-col gap-1">
          <span className="font-mono text-xs text-white font-bold">{row.original.operation_id}</span>
          <span className="inline-flex w-fit items-center gap-1 text-[10px] font-bold text-mazad-primary bg-mazad-primary/10 px-2 py-0.5 rounded">
            <Zap className="w-3 h-3" /> تلقائي
          </span>
        </div>
      ),
    },
    {
      header: 'معاملة بنكيلي',
      accessorKey: 'bankily_transaction_id',
      cell: ({ getValue }) => <span className="font-mono text-xs text-surface-muted">{getValue<string>() || '—'}</span>,
    },
    {
      header: 'المستخدم',
      accessorKey: 'user_id',
      cell: ({ row }) => (
        <button onClick={() => navigate(`/users/${row.original.user_id}`)} className="text-xs font-bold text-white hover:text-mazad-primary">
          {row.original.user_full_name || row.original.user_phone || shortID(row.original.user_id)}
          <span className="block font-mono text-[10px] text-surface-muted">بنكيلي: {row.original.client_phone}</span>
        </button>
      ),
    },
    {
      header: 'المبلغ',
      accessorKey: 'amount',
      cell: ({ row }) => <span className="font-mono font-bold text-mazad-accent">{formatPrice(parseFloat(row.original.amount), row.original.currency_code)}</span>,
    },
    {
      header: 'الحالة',
      accessorKey: 'status',
      cell: ({ row }) => {
        const s = STATUS_LABELS[row.original.status] ?? { label: row.original.status, cls: 'bg-surface-border/40 text-white' }
        return (
          <div className="flex flex-col gap-1">
            <span className={`w-fit text-xs font-bold px-2 py-1 rounded ${s.cls}`}>{s.label}</span>
            {row.original.bankily_status && <span className="text-[10px] text-surface-muted font-mono">Bankily: {row.original.bankily_status}</span>}
            {row.original.failure_reason && <span className="text-[10px] text-red-300/80">{row.original.failure_reason}</span>}
            {row.original.needs_review && (
              <span className="inline-flex w-fit items-center gap-1 text-[10px] font-bold text-orange-300">
                <AlertTriangle className="w-3 h-3" /> يحتاج مراجعة
              </span>
            )}
          </div>
        )
      },
    },
    {
      header: 'محاولات التحقق',
      accessorKey: 'check_attempts',
      cell: ({ getValue }) => <span className="font-mono text-xs text-surface-muted">{getValue<number>()}</span>,
    },
    {
      header: 'التوقيت',
      accessorKey: 'created_at',
      cell: ({ row }) => (
        <div className="text-[11px] text-surface-muted leading-5">
          <div>أُنشئت: {formatDate(row.original.created_at)}</div>
          {row.original.completed_at && <div>اكتملت: {formatDate(row.original.completed_at)}</div>}
          {row.original.last_checked_at && <div>آخر تحقق: {formatDate(row.original.last_checked_at)}</div>}
        </div>
      ),
    },
    {
      header: 'الإجراءات',
      id: 'actions',
      cell: ({ row }) => {
        const p = row.original
        if (p.status === 'SUCCESS' && p.wallet_transaction_id) {
          return (
            <button onClick={() => navigate(`/transactions/${p.wallet_transaction_id}`)} className="text-xs font-bold text-mazad-primary">
              قيد المحفظة
            </button>
          )
        }
        if (p.status === 'PENDING_VERIFICATION' || p.status === 'SUBMITTED') {
          return (
            <button
              disabled={recheck.isPending}
              onClick={() => recheck.mutate(p.id)}
              className="flex items-center gap-2 px-3 py-2 rounded-lg text-xs font-bold bg-mazad-primary/15 text-mazad-primary
                         hover:bg-mazad-primary hover:text-white transition-all disabled:opacity-40"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${recheck.isPending && recheck.variables === p.id ? 'animate-spin' : ''}`} />
              إعادة التحقق الآن
            </button>
          )
        }
        return <span className="text-xs text-surface-muted">—</span>
      },
    },
  ]

  if (isError) return (
    <div className="admin-card p-20 text-center flex flex-col items-center gap-4">
      <AlertCircle className="w-12 h-12 text-red-500/20" />
      <p className="text-red-400 font-bold">فشل تحميل مدفوعات بنكيلي</p>
    </div>
  )

  return (
    <div className="animate-fade-in" dir="rtl">
      <PageHeader title="مدفوعات بنكيلي B-PAY" subtitle="شحن تلقائي للمحفظة — لا يتطلب موافقة يدوية" />

      <div className="flex gap-1 mb-6 bg-surface-card border border-surface-border rounded-xl p-1 w-fit">
        {TABS.map((tab) => (
          <button
            key={tab.value}
            onClick={() => setSearchParams({ status: tab.value, page: '1' })}
            className={`px-4 py-2 rounded-lg text-xs font-bold transition-all ${
              status === tab.value ? 'bg-mazad-primary text-white' : 'text-surface-muted hover:text-white hover:bg-surface-border/50'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <div className="flex items-center gap-3 bg-sky-500/10 border border-sky-500/20 text-sky-300 text-xs font-bold rounded-xl px-5 py-4 mb-6">
        <Zap className="w-5 h-5 shrink-0" />
        <span>
          الدفعات المؤكدة من بنكيلي (TS) تُضاف إلى المحفظة تلقائياً مرة واحدة فقط. "إعادة التحقق" يستعلم بنكيلي فقط ولا يضيف أي رصيد يدوياً.
        </span>
      </div>

      <DataTable
        columns={columns}
        data={data?.data ?? []}
        isLoading={isLoading}
        total={data?.total}
        page={page}
        onPageChange={(p) => setSearchParams((prev) => { prev.set('page', p.toString()); return prev })}
        emptyTitle="لا توجد مدفوعات"
        emptyDescription="لم يتم العثور على مدفوعات بنكيلي تطابق هذا الفلتر."
      />
    </div>
  )
}
