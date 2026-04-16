import { useNavigate, useSearchParams } from 'react-router-dom'
import { Eye, CreditCard, Clock, AlertCircle } from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { StatusBadge } from '@/components/shared/StatusBadge'
import { DataTable } from '@/components/shared/DataTable'
import { useTransactions } from '@/hooks/useTransactions'
import { formatPrice, formatDate, shortID } from '@/lib/formatters'
import { GATEWAY_LABELS } from '@/lib/constants'
import type { Transaction } from '@/types/api'
import type { ColumnDef } from '@tanstack/react-table'

const TYPE_LABELS: Record<string, string> = {
  deposit:    'إيداع',
  withdraw:   'سحب',
  bid_hold:   'تأمين مزايدة',
  bid_refund: 'استرداد تأمين',
  payment:    'دفع',
}

const STATUS_TABS = [
  { label: 'بانتظار المراجعة', value: 'pending_review', icon: Clock },
  { label: 'السحوبات',       value: 'withdraw',       icon: CreditCard },
  { label: 'الكل',           value: '',               icon: null },
]

export function TransactionsPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const status = searchParams.get('status') ?? 'pending_review'
  const page = parseInt(searchParams.get('page') ?? '1')

  const { data, isLoading, isError } = useTransactions({
    status: status || undefined,
    page,
    per_page: 25,
  })

  const columns: ColumnDef<Transaction>[] = [
    {
      header: 'المعرف',
      accessorKey: 'id',
      cell: ({ getValue }) => <span className="font-mono text-xs text-surface-muted font-bold">{shortID(getValue<string>())}</span>
    },
    {
      header: 'المستخدم',
      accessorKey: 'user_id',
      cell: ({ getValue }) => <span className="font-mono text-xs text-surface-muted font-bold">{shortID(getValue<string>())}</span>
    },
    {
      header: 'النوع',
      accessorKey: 'type',
      cell: ({ getValue }) => (
        <span className="text-xs font-bold text-white bg-surface-border/40 px-2 py-1 rounded">
          {TYPE_LABELS[getValue<string>()] ?? getValue<string>()}
        </span>
      )
    },
    {
      header: 'المبلغ',
      accessorKey: 'amount',
      cell: ({ getValue }) => <span className="font-mono font-bold text-mazad-accent text-base">{formatPrice(parseFloat(getValue<string>()))}</span>
    },
    {
      header: 'البوابة',
      accessorKey: 'gateway',
      cell: ({ getValue }) => <span className="text-xs text-surface-muted font-bold">{getValue<string>() ? GATEWAY_LABELS[getValue<string>()] ?? getValue<string>() : '—'}</span>
    },
    {
      header: 'التاريخ',
      accessorKey: 'created_at',
      cell: ({ getValue }) => <span className="text-xs text-surface-muted font-medium">{formatDate(getValue<string>())}</span>
    },
    {
      header: 'الحالة',
      accessorKey: 'status',
      cell: ({ getValue }) => <StatusBadge status={getValue<string>()} />
    },
    {
      header: 'الإجراءات',
      id: 'actions',
      cell: ({ row }) => (
        <button
          onClick={() => navigate(`/transactions/${row.original.id}`)}
          className="flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-bold
                     bg-mazad-primary/15 text-mazad-primary hover:bg-mazad-primary hover:text-white transition-all shadow-sm"
        >
          <Eye className="w-3.5 h-3.5" />
          {row.original.status === 'pending_review' ? 'مراجعة وتأكيد' : 'عرض التفاصيل'}
        </button>
      )
    }
  ]

  if (isError) return (
    <div className="admin-card p-20 text-center flex flex-col items-center gap-4">
      <AlertCircle className="w-12 h-12 text-red-500/20" />
      <p className="text-red-400 font-bold">فشل تحميل المعاملات</p>
      <button onClick={() => window.location.reload()} className="bg-surface-border text-white px-6 py-2 rounded-xl text-sm font-bold">إعادة المحاولة</button>
    </div>
  )

  return (
    <div className="animate-fade-in" dir="rtl">
      <PageHeader
        title="المعاملات المالية"
        subtitle="مراجعة وتأكيد الإيداعات والسحوبات"
      />

      {/* Tabs */}
      <div className="flex gap-1 mb-6 bg-surface-card border border-surface-border rounded-xl p-1 w-fit">
        {STATUS_TABS.map((tab) => (
          <button
            key={tab.value}
            onClick={() => setSearchParams({ status: tab.value, page: '1' })}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-bold transition-all ${
              status === tab.value
                ? 'bg-mazad-primary text-white shadow-lg shadow-mazad-primary/20'
                : 'text-surface-muted hover:text-white hover:bg-surface-border/50'
            }`}
          >
            {tab.icon && <tab.icon className="w-3.5 h-3.5" />}
            {tab.label}
          </button>
        ))}
      </div>

      {/* Warning for pending deposits */}
      {status === 'pending_review' && (data?.total ?? 0) > 0 && (
        <div className="flex items-center gap-3 bg-orange-500/10 border border-orange-500/20
                        text-orange-400 text-xs font-bold rounded-xl px-5 py-4 mb-6 animate-pulse-subtle">
          <Clock className="w-5 h-5 shrink-0" />
          <span>
             يوجد <strong>{data?.total}</strong> إيداع في انتظار المراجعة. يرجى التأكد من صورة الإيصال قبل الموافقة.
          </span>
        </div>
      )}

      <DataTable
        columns={columns}
        data={data?.data ?? []}
        isLoading={isLoading}
        total={data?.total}
        page={page}
        onPageChange={(p) => setSearchParams((prev) => { prev.set('page', p.toString()); return prev })}
        emptyTitle="لا توجد معاملات"
        emptyDescription="لم يتم العثور على أي معاملات تطابق هذا الفلتر."
      />
    </div>
  )
}
