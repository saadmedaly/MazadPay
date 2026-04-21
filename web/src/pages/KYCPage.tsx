import { useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { CheckCircle2, XCircle, Eye, ShieldCheck, Image as ImageIcon, Tag, Trash2, Search, Calendar, CheckSquare, Square } from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { StatusBadge } from '@/components/shared/StatusBadge'
import { DataTable } from '@/components/shared/DataTable'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { RequestDetailModal } from '@/components/requests/RequestDetailModal'
import {
  useAuctionRequests, useReviewAuctionRequest, useDeleteAuctionRequest,
  useBannerRequests, useReviewBannerRequest, useDeleteBannerRequest,
  useBulkReviewAuctionRequests, useBulkDeleteAuctionRequests,
  useBulkReviewBannerRequests, useBulkDeleteBannerRequests
} from '@/hooks/useRequests'
import { formatDate, shortID } from '@/lib/formatters'
import type { AuctionRequest, BannerRequest } from '@/hooks/useRequests'
import type { ColumnDef } from '@tanstack/react-table'

export function KYCPage() {
  const queryClient = useQueryClient()
  const [requestType, setRequestType] = useState<'auction' | 'banner'>('auction')
  const [status, setStatus] = useState('pending')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState(20)
  const [search, setSearch] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [reviewTarget, setReviewTarget] = useState<{ id: string; name: string; type: 'auction' | 'banner'; status: 'approved' | 'rejected' } | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<{ id: string; name: string; type: 'auction' | 'banner' } | null>(null)
  const [notes, setNotes] = useState('')
  const [selectedIds, setSelectedIds] = useState<string[]>([])
  const [detailRequestId, setDetailRequestId] = useState<string | null>(null)
  const [isDetailModalOpen, setIsDetailModalOpen] = useState(false)
  const [bulkAction, setBulkAction] = useState<'approve' | 'reject' | 'delete' | null>(null)

  const [categoryId, setCategoryId] = useState<number | undefined>()
  const [locationId, setLocationId] = useState<number | undefined>()
  const [minPrice, setMinPrice] = useState<number | undefined>()
  const [maxPrice, setMaxPrice] = useState<number | undefined>()

  const auctionFilters = {
    status,
    category_id: categoryId,
    location_id: locationId,
    min_price: minPrice,
    max_price: maxPrice,
    date_from: dateFrom || undefined,
    date_to: dateTo || undefined
  }

  const { data: auctionData, isLoading: loadingAuctions } = useAuctionRequests(auctionFilters, page, perPage)
  const { data: bannerData, isLoading: loadingBanners } = useBannerRequests(status, page, perPage)
  const reviewAuctionRequest = useReviewAuctionRequest()
  const deleteAuctionRequest = useDeleteAuctionRequest()
  const reviewBannerRequest = useReviewBannerRequest()
  const deleteBannerRequest = useDeleteBannerRequest()
  const bulkReviewAuction = useBulkReviewAuctionRequests()
  const bulkDeleteAuction = useBulkDeleteAuctionRequests()
  const bulkReviewBanner = useBulkReviewBannerRequests()
  const bulkDeleteBanner = useBulkDeleteBannerRequests()

  const currentData = requestType === 'auction' ? auctionData?.data : bannerData?.data
  const currentTotal = requestType === 'auction' ? auctionData?.total : bannerData?.total
  const currentLoading = requestType === 'auction' ? loadingAuctions : loadingBanners

  // Filter data based on search and date
  const filteredData = currentData?.filter(item => {
    const matchesSearch = !search || item.user?.full_name?.toLowerCase().includes(search.toLowerCase()) || item.user?.phone.includes(search)
    const matchesDateFrom = !dateFrom || new Date(item.created_at) >= new Date(dateFrom)
    const matchesDateTo = !dateTo || new Date(item.created_at) <= new Date(dateTo)
    return matchesSearch && matchesDateFrom && matchesDateTo
  }) || []

  const totalPages = Math.ceil((currentTotal || 0) / perPage)

  const auctionColumns: ColumnDef<AuctionRequest>[] = [
    {
      header: ({ table }) => (
        <input
          type="checkbox"
          checked={table.getIsAllPageRowsSelected()}
          onChange={(e) => handleSelectAll(e.target.checked)}
          className="w-4 h-4 rounded border-gray-300"
        />
      ),
      id: 'select',
      cell: ({ row }) => (
        <input
          type="checkbox"
          checked={selectedIds.includes(row.original.id)}
          onChange={(e) => handleSelectOne(row.original.id, e.target.checked)}
          className="w-4 h-4 rounded border-gray-300"
        />
      )
    },
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
      header: 'عنوان المزاد',
      accessorKey: 'title_ar',
      cell: ({ getValue }) => <span className="text-xs font-bold">{getValue<string>()}</span>
    },
    {
      header: 'السعر',
      accessorKey: 'start_price',
      cell: ({ getValue }) => <span className="text-xs text-surface-muted font-mono">{getValue<string>()} MRU</span>
    },
    {
      header: 'تاريخ البدء',
      accessorKey: 'start_date',
      cell: ({ getValue }) => <span className="text-xs text-surface-muted">{formatDate(getValue<string>())}</span>
    },
    {
      header: 'الحالة',
      accessorKey: 'status',
      cell: ({ getValue }) => <StatusBadge status={getValue<string>() as any} />
    },
    {
      header: 'الإجراءات',
      id: 'actions',
      cell: ({ row }) => {
        return (
          <div className="flex items-center gap-2">
            <button
              onClick={() => openDetailModal(row.original.id)}
              className="p-2 rounded-lg text-blue-400 hover:bg-blue-500/10 transition-colors"
              title="عرض التفاصيل"
            >
              <Eye className="w-4 h-4" />
            </button>
            {row.original.status === 'pending' && (
              <>
                <button
                  onClick={() => setReviewTarget({ id: row.original.id, name: row.original.title_ar, type: 'auction', status: 'approved' })}
                  className="p-2 rounded-lg text-emerald-400 hover:bg-emerald-500/10 transition-colors"
                  title="قبول الطلب"
                >
                  <CheckCircle2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => setReviewTarget({ id: row.original.id, name: row.original.title_ar, type: 'auction', status: 'rejected' })}
                  className="p-2 rounded-lg text-red-400 hover:bg-red-500/10 transition-colors"
                  title="رفض الطلب"
                >
                  <XCircle className="w-4 h-4" />
                </button>
              </>
            )}
            <button
              onClick={() => setDeleteTarget({ id: row.original.id, name: row.original.title_ar, type: 'auction' })}
              className="p-2 rounded-lg text-surface-muted hover:text-red-400 hover:bg-red-500/10 transition-colors"
              title="حذف الطلب"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
        )
      }
    }
  ]

  const bannerColumns: ColumnDef<BannerRequest>[] = [
    {
      header: ({ table }) => (
        <input
          type="checkbox"
          checked={table.getIsAllPageRowsSelected()}
          onChange={(e) => handleSelectAll(e.target.checked)}
          className="w-4 h-4 rounded border-gray-300"
        />
      ),
      id: 'select',
      cell: ({ row }) => (
        <input
          type="checkbox"
          checked={selectedIds.includes(row.original.id)}
          onChange={(e) => handleSelectOne(row.original.id, e.target.checked)}
          className="w-4 h-4 rounded border-gray-300"
        />
      )
    },
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
      header: 'عنوان البانر',
      accessorKey: 'title_ar',
      cell: ({ getValue }) => <span className="text-xs font-bold">{getValue<string>()}</span>
    },
    {
      header: 'الصورة',
      id: 'image',
      cell: ({ row }) => (
        <a href={row.original.image_url} target="_blank" rel="noreferrer"
           className="p-1.5 rounded bg-surface-border/50 text-surface-muted hover:text-white transition-colors">
          <ImageIcon className="w-3.5 h-3.5" />
        </a>
      )
    },
    {
      header: 'تاريخ البدء',
      accessorKey: 'starts_at',
      cell: ({ getValue }) => <span className="text-xs text-surface-muted">{formatDate(getValue<string>())}</span>
    },
    {
      header: 'الحالة',
      accessorKey: 'status',
      cell: ({ getValue }) => <StatusBadge status={getValue<string>() as any} />
    },
    {
      header: 'الإجراءات',
      id: 'actions',
      cell: ({ row }) => {
        return (
          <div className="flex items-center gap-2">
            <button
              onClick={() => openDetailModal(row.original.id)}
              className="p-2 rounded-lg text-blue-400 hover:bg-blue-500/10 transition-colors"
              title="عرض التفاصيل"
            >
              <Eye className="w-4 h-4" />
            </button>
            {row.original.status === 'pending' && (
              <>
                <button
                  onClick={() => setReviewTarget({ id: row.original.id, name: row.original.title_ar, type: 'banner', status: 'approved' })}
                  className="p-2 rounded-lg text-emerald-400 hover:bg-emerald-500/10 transition-colors"
                  title="قبول الطلب"
                >
                  <CheckCircle2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => setReviewTarget({ id: row.original.id, name: row.original.title_ar, type: 'banner', status: 'rejected' })}
                  className="p-2 rounded-lg text-red-400 hover:bg-red-500/10 transition-colors"
                  title="رفض الطلب"
                >
                  <XCircle className="w-4 h-4" />
                </button>
              </>
            )}
            <button
              onClick={() => setDeleteTarget({ id: row.original.id, name: row.original.title_ar, type: 'banner' })}
              className="p-2 rounded-lg text-surface-muted hover:text-red-400 hover:bg-red-500/10 transition-colors"
              title="حذف الطلب"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
        )
      }
    }
  ]

  const columns = requestType === 'auction' ? auctionColumns : bannerColumns

  const handleReview = () => {
    if (!reviewTarget) return
    
    const mutation = reviewTarget.type === 'auction' ? reviewAuctionRequest : reviewBannerRequest
    mutation.mutate(
      { id: reviewTarget.id, status: reviewTarget.status, notes },
      { 
        onSuccess: () => { 
          queryClient.invalidateQueries({ queryKey: ['auction-requests'] })
          queryClient.invalidateQueries({ queryKey: ['banner-requests'] })
          setReviewTarget(null)
          setNotes('')
        } 
      }
    )
  }

  const handleDelete = () => {
    if (!deleteTarget) return
    
    const mutation = deleteTarget.type === 'auction' ? deleteAuctionRequest : deleteBannerRequest
    mutation.mutate(deleteTarget.id, {
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['auction-requests'] })
        queryClient.invalidateQueries({ queryKey: ['banner-requests'] })
        setDeleteTarget(null)
      }
    })
  }

  // Bulk action handlers
  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedIds(currentData?.map(item => item.id) || [])
    } else {
      setSelectedIds([])
    }
  }

  const handleSelectOne = (id: string, checked: boolean) => {
    if (checked) {
      setSelectedIds(prev => [...prev, id])
    } else {
      setSelectedIds(prev => prev.filter(i => i !== id))
    }
  }

  const handleBulkAction = () => {
    if (!bulkAction || selectedIds.length === 0) return

    if (bulkAction === 'delete') {
      const mutation = requestType === 'auction' ? bulkDeleteAuction : bulkDeleteBanner
      mutation.mutate(selectedIds, {
        onSuccess: () => {
          queryClient.invalidateQueries({ queryKey: ['auction-requests'] })
          queryClient.invalidateQueries({ queryKey: ['banner-requests'] })
          setSelectedIds([])
          setBulkAction(null)
        }
      })
    } else {
      const mutation = requestType === 'auction' ? bulkReviewAuction : bulkReviewBanner
      mutation.mutate({ ids: selectedIds, status: bulkAction, notes: notes || undefined }, {
        onSuccess: () => {
          queryClient.invalidateQueries({ queryKey: ['auction-requests'] })
          queryClient.invalidateQueries({ queryKey: ['banner-requests'] })
          setSelectedIds([])
          setNotes('')
          setBulkAction(null)
        }
      })
    }
  }

  const openDetailModal = (id: string) => {
    setDetailRequestId(id)
    setIsDetailModalOpen(true)
  }

  const closeDetailModal = () => {
    setIsDetailModalOpen(false)
    setDetailRequestId(null)
  }

  return (
    <div className="animate-fade-in" dir="rtl">
      <PageHeader 
        title="إدارة الطلبات" 
        subtitle="مراجعة طلبات إضافة المزادات   الاعلانات"
        icon={ShieldCheck}
      />

      <div className="flex items-center gap-2 mb-6">
        <button
          onClick={() => setRequestType('auction')}
          className={`px-4 py-2 rounded-xl text-xs font-bold transition-all border flex items-center gap-2 ${
            requestType === 'auction' 
              ? 'bg-mazad-primary text-white border-mazad-primary shadow-lg shadow-mazad-primary/20' 
              : 'bg-surface-card text-surface-muted border-surface-border hover:text-white'
          }`}
        >
          <Tag className="w-3.5 h-3.5" />
          طلبات المزادات
        </button>
        <button
          onClick={() => setRequestType('banner')}
          className={`px-4 py-2 rounded-xl text-xs font-bold transition-all border flex items-center gap-2 ${
            requestType === 'banner' 
              ? 'bg-mazad-primary text-white border-mazad-primary shadow-lg shadow-mazad-primary/20' 
              : 'bg-surface-card text-surface-muted border-surface-border hover:text-white'
          }`}
        >
          <ImageIcon className="w-3.5 h-3.5" />
          طلبات الإعلانات
        </button>
      </div>

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
            {s === 'pending' ? 'بانتظار المراجعة' : s === 'approved' ? 'مقبولة' : 'مرفوضة'}
          </button>
        ))}
      </div>

      {/* Search and Filter Controls */}
      <div className="flex items-center gap-4 mb-4 px-4">
        <div className="flex-1 relative">
          <Search className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted" />
          <input
            type="text"
            placeholder="بحث بالاسم أو رقم الهاتف..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pr-10 pl-4 py-2 bg-surface-border border border-surface-border rounded-lg text-sm text-white placeholder:text-surface-muted focus:outline-none focus:border-surface-accent"
          />
        </div>
        <div className="flex items-center gap-2">
          <div className="relative">
            <Calendar className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted" />
            <input
              type="date"
              value={dateFrom}
              onChange={(e) => setDateFrom(e.target.value)}
              className="pr-10 pl-4 py-2 bg-surface-border border border-surface-border rounded-lg text-sm text-white focus:outline-none focus:border-surface-accent"
            />
          </div>
          <span className="text-surface-muted">-</span>
          <div className="relative">
            <Calendar className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted" />
            <input
              type="date"
              value={dateTo}
              onChange={(e) => setDateTo(e.target.value)}
              className="pr-10 pl-4 py-2 bg-surface-border border border-surface-border rounded-lg text-sm text-white focus:outline-none focus:border-surface-accent"
            />
          </div>
        </div>
      </div>

      {/* Bulk Actions Bar */}
      {selectedIds.length > 0 && (
        <div className="flex items-center justify-between bg-surface-card border border-surface-border rounded-lg p-3 mb-4">
          <div className="flex items-center gap-2">
            <span className="text-sm text-surface-muted">
              تم اختيار {selectedIds.length} عنصر
            </span>
            <button
              onClick={() => setSelectedIds([])}
              className="text-xs text-surface-muted hover:text-white underline"
            >
              إلغاء التحديد
            </button>
          </div>
          <div className="flex items-center gap-2">
            {status === 'pending' && (
              <>
                <button
                  onClick={() => setBulkAction('approve')}
                  className="px-3 py-1.5 text-xs bg-emerald-500/20 text-emerald-400 rounded hover:bg-emerald-500/30 transition-colors"
                >
                  قبول الكل
                </button>
                <button
                  onClick={() => setBulkAction('reject')}
                  className="px-3 py-1.5 text-xs bg-red-500/20 text-red-400 rounded hover:bg-red-500/30 transition-colors"
                >
                  رفض الكل
                </button>
              </>
            )}
            <button
              onClick={() => setBulkAction('delete')}
              className="px-3 py-1.5 text-xs bg-red-500/20 text-red-400 rounded hover:bg-red-500/30 transition-colors"
            >
              حذف الكل
            </button>
          </div>
        </div>
      )}

      <DataTable
        columns={columns}
        data={filteredData}
        isLoading={currentLoading}
        total={filteredData.length}
        emptyTitle="لا توجد طلبات"
        emptyDescription="ليس هناك أي طلبات في هذا القسم حالياً."
      />

      {/* Pagination Controls */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between mt-4 px-4">
          <div className="flex items-center gap-4">
            <div className="text-xs text-surface-muted">
              صفحة {page} من {totalPages} ({currentTotal || 0} إجمالي)
            </div>
            <div className="flex items-center gap-2">
              <span className="text-xs text-surface-muted">عرض:</span>
              <select
                value={perPage}
                onChange={(e) => {
                  setPerPage(Number(e.target.value))
                  setPage(1)
                }}
                className="px-2 py-1 text-xs bg-surface-border border border-surface-border rounded focus:outline-none focus:border-surface-accent"
              >
                <option value={10}>10</option>
                <option value={20}>20</option>
                <option value={50}>50</option>
                <option value={100}>100</option>
              </select>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="px-3 py-1 text-xs bg-surface-border rounded disabled:opacity-50 disabled:cursor-not-allowed"
            >
              السابق
            </button>
            <span className="text-xs">{page}</span>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="px-3 py-1 text-xs bg-surface-border rounded disabled:opacity-50 disabled:cursor-not-allowed"
            >
              التالي
            </button>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={!!reviewTarget}
        onOpenChange={(v) => {
          if (!v) {
            setReviewTarget(null)
            setNotes('')
          }
        }}
        title={reviewTarget?.status === 'approved' ? `قبول طلب ${reviewTarget?.type === 'auction' ? 'المزاد' : 'البانر'}` : `رفض طلب ${reviewTarget?.type === 'auction' ? 'المزاد' : 'البانر'}`}
        description={
          <div className="space-y-4 pt-2 text-right" dir="rtl">
            <p>هل أنت متأكد من {reviewTarget?.status === 'approved' ? 'قبول' : 'رفض'} طلب <span className="text-white font-bold">{reviewTarget?.name}</span>؟</p>
            <div className="space-y-2">
              <label className="text-xs text-surface-muted block font-bold">ملاحظات إضافية (اختياري)</label>
              <textarea
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                className="w-full bg-surface-bg border border-surface-border rounded-xl p-3 text-xs text-white focus:outline-none focus:border-mazad-primary min-h-[100px]"
                placeholder={reviewTarget?.status === 'approved' ? 'مثال: تم التحقق من البيانات' : 'مثال: البيانات غير كافية...'}
              />
            </div>
          </div>
        }
        confirmLabel={reviewTarget?.status === 'approved' ? 'تأكيد القبول' : 'تأكيد الرفض'}
        variant={reviewTarget?.status === 'approved' ? 'success' : 'danger'}
        loading={reviewAuctionRequest.isPending || reviewBannerRequest.isPending}
        onConfirm={handleReview}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(v) => {
          if (!v) {
            setDeleteTarget(null)
          }
        }}
        title={`حذف طلب ${deleteTarget?.type === 'auction' ? 'المزاد' : 'البانر'}`}
        description={
          <div className="space-y-4 pt-2 text-right" dir="rtl">
            <p>هل أنت متأكد من حذف طلب <span className="text-white font-bold">{deleteTarget?.name}</span>؟</p>
            <p className="text-xs text-red-400">⚠️ هذا الإجراء لا يمكن التراجع عنه</p>
          </div>
        }
        confirmLabel="تأكيد الحذف"
        variant="danger"
        loading={deleteAuctionRequest.isPending || deleteBannerRequest.isPending}
        onConfirm={handleDelete}
      />

      {/* Bulk Action ConfirmDialog */}
      <ConfirmDialog
        open={!!bulkAction}
        onOpenChange={(v) => {
          if (!v) {
            setBulkAction(null)
            setNotes('')
          }
        }}
        title={
          bulkAction === 'approve' ? 'قبول الطلبات المحددة' :
          bulkAction === 'reject' ? 'رفض الطلبات المحددة' :
          'حذف الطلبات المحددة'
        }
        description={
          <div className="space-y-4 pt-2 text-right" dir="rtl">
            <p>هل أنت متأكد من {bulkAction === 'approve' ? 'قبول' : bulkAction === 'reject' ? 'رفض' : 'حذف'} {selectedIds.length} طلب؟</p>
            {bulkAction !== 'delete' && (
              <div className="space-y-2">
                <label className="text-xs text-surface-muted block font-bold">ملاحظات إضافية (اختياري)</label>
                <textarea
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  className="w-full bg-surface-bg border border-surface-border rounded-xl p-3 text-xs text-white focus:outline-none focus:border-mazad-primary min-h-[100px]"
                  placeholder={bulkAction === 'approve' ? 'مثال: تم التحقق من البيانات' : 'مثال: البيانات غير كافية...'}
                />
              </div>
            )}
            {bulkAction === 'delete' && (
              <p className="text-xs text-red-400">⚠️ هذا الإجراء لا يمكن التراجع عنه</p>
            )}
          </div>
        }
        confirmLabel={
          bulkAction === 'approve' ? 'تأكيد القبول' :
          bulkAction === 'reject' ? 'تأكيد الرفض' :
          'تأكيد الحذف'
        }
        variant={bulkAction === 'approve' ? 'success' : 'danger'}
        loading={bulkReviewAuction.isPending || bulkDeleteAuction.isPending || bulkReviewBanner.isPending || bulkDeleteBanner.isPending}
        onConfirm={handleBulkAction}
      />

      {/* Request Detail Modal */}
      <RequestDetailModal
        isOpen={isDetailModalOpen}
        onClose={closeDetailModal}
        type={requestType}
        requestId={detailRequestId}
        onApprove={(id) => {
          setReviewTarget({ id, name: '', type: requestType, status: 'approved' })
          closeDetailModal()
        }}
        onReject={(id) => {
          setReviewTarget({ id, name: '', type: requestType, status: 'rejected' })
          closeDetailModal()
        }}
        onDelete={(id) => {
          const item = currentData?.find(i => i.id === id)
          setDeleteTarget({ id, name: item?.title_ar || '', type: requestType })
          closeDetailModal()
        }}
      />
    </div>
  )
}
