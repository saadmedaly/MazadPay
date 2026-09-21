import { useState } from 'react'
import type { LucideIcon } from 'lucide-react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  MapPin,
  Tag,
  Check,
  X,
  Gavel,
  Loader2,
  AlertCircle,
  ImageIcon,
  Calendar,
  User,
  Pencil,
  AlertTriangle,
  FileText,
  ShieldCheck,
} from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { StatusBadge } from '@/components/shared/StatusBadge'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { ImagePreview } from '@/components/shared/ImagePreview'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { useAuctionRequestByID, useReviewAuctionRequest, useDeleteAuctionRequest } from '@/hooks/useRequests'
import { formatPrice, formatDate, shortID } from '@/lib/formatters'
import { requestIsApprovableForInsurance } from '@/lib/insurancePolicy'
import { cn } from '@/lib/utils'

// Client feedback (Note #8): the user-submitted ad review was a crowded,
// form-like modal (RequestDetailModal.tsx) with raw data-entry styling and
// no visual hierarchy for the image/core info. This page reuses the exact
// visual structure already established by AuctionDetailPage.tsx (gallery +
// details card on the left, status/actions card on the right) so a
// pending/rejected AUCTION request now gets the same professional details
// view as a live auction, without inventing a parallel design. All fields
// previously shown in RequestDetailModal's renderAuctionDetails are
// preserved here -- none were dropped, only regrouped visually. Banner
// requests are untouched (still use RequestDetailModal), matching the
// client's screenshot which is specifically about auction ad review.
function DetailItem({ label, value, icon: Icon, color = "text-white" }: { label: string, value: string | number, icon?: LucideIcon, color?: string }) {
  return (
    <div className="flex flex-col">
      <span className="text-[10px] font-bold text-surface-muted uppercase mb-1.5 flex items-center gap-1.5 tracking-wider">
        {Icon && <Icon className="w-3 h-3 text-mazad-primary" />}
        {label}
      </span>
      <span className={cn("text-sm font-bold truncate", color)}>{value}</span>
    </div>
  )
}

export function AuctionRequestDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [rejectDialog, setRejectDialog] = useState(false)
  const [rejectionReason, setRejectionReason] = useState('')
  const [approveConfirm, setApproveConfirm] = useState(false)
  const [activeImage, setActiveImage] = useState<string | null>(null)
  const [activeLang, setActiveLang] = useState<'ar' | 'fr' | 'en'>('ar')
  const [deleteConfirm, setDeleteConfirm] = useState(false)

  const { data: request, isLoading, isError } = useAuctionRequestByID(id ?? null)
  const review = useReviewAuctionRequest()
  const deleteRequest = useDeleteAuctionRequest()

  const images: string[] = (() => {
    if (!request?.images) return []
    if (Array.isArray(request.images)) return request.images.filter((i): i is string => typeof i === 'string')
    return []
  })()

  const handleReview = (approve: boolean) => {
    if (!id) return
    review.mutate(
      { id, status: approve ? 'approved' : 'rejected', notes: rejectionReason },
      {
        onSuccess: () => {
          setRejectDialog(false)
          setApproveConfirm(false)
          navigate('/requests')
        }
      }
    )
  }

  if (isLoading) return <LoadingSpinner fullPage label="جاري تحميل تفاصيل الطلب..." />

  if (isError || !request) return (
    <div className="flex flex-col items-center justify-center h-64 text-surface-muted gap-4">
      <AlertCircle className="w-12 h-12 opacity-20" />
      <p className="font-bold">فشل في العثور على الطلب</p>
      <button onClick={() => navigate(-1)} className="text-mazad-primary text-sm font-bold">رجوع للوراء</button>
    </div>
  )

  const isPending = request.status === 'pending'
  const isDraftOrRejected = request.status === 'draft' || request.status === 'rejected'
  const hasInsurance = requestIsApprovableForInsurance(request)

  return (
    <div className="animate-fade-in max-w-6xl" dir="rtl">
      <PageHeader title="تفاصيل طلب المزاد">
        <button
          onClick={() => navigate(-1)}
          className="flex items-center gap-2 text-sm font-bold text-surface-muted hover:text-white transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          رجوع
        </button>
      </PageHeader>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Left Column: Media & Details */}
        <div className="lg:col-span-2 space-y-6">
          {/* Main Gallery */}
          <div className="admin-card overflow-hidden">
            <div className="aspect-video bg-surface-base relative group">
              {images.length > 0 ? (
                <ImagePreview
                  src={images[0]}
                  className="w-full h-full object-contain cursor-zoom-in transition-transform group-hover:scale-[1.02]"
                  onClick={() => setActiveImage(images[0])}
                />
              ) : (
                <div className="w-full h-full flex flex-col items-center justify-center text-surface-muted gap-3">
                  <ImageIcon className="w-12 h-12 opacity-20" />
                  <span className="text-sm font-bold">لا يوجد صور لهذا الطلب</span>
                </div>
              )}
            </div>
            {images.length > 1 && (
              <div className="p-4 flex gap-3 overflow-x-auto border-t border-surface-border bg-surface-base/30 custom-scrollbar">
                {images.map((img, i) => (
                  <button
                    key={i}
                    onClick={() => setActiveImage(img)}
                    className="w-20 h-20 rounded-lg overflow-hidden border border-surface-border hover:border-mazad-primary/40 transition-all shrink-0"
                  >
                    <img src={img} className="w-full h-full object-cover" alt="" />
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Description & Specs */}
          <div className="admin-card p-6 overflow-hidden">
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6 border-b border-surface-border pb-4">
              <h2 className={cn(
                "font-display font-bold text-white text-lg",
                activeLang !== 'ar' && "font-sans"
              )}>
                {activeLang === 'ar' ? 'وصف الطلب' :
                 activeLang === 'fr' ? 'Description' : 'Description'}
              </h2>

              <div className="flex bg-surface-base p-1 rounded-lg border border-surface-border w-fit">
                {(['ar', 'fr', 'en'] as const).map(lang => (
                  <button
                    key={lang}
                    onClick={() => setActiveLang(lang)}
                    className={cn(
                      "px-3 py-1 text-[10px] font-bold rounded flex items-center justify-center transition-all",
                      activeLang === lang
                        ? "bg-mazad-primary text-white shadow-sm"
                        : "text-surface-muted hover:text-white"
                    )}
                  >
                    {lang.toUpperCase()}
                  </button>
                ))}
              </div>
            </div>

            <p className={cn(
              "text-sm leading-relaxed font-medium",
              activeLang === 'ar' ? "text-right text-surface-muted" : "text-left text-surface-muted font-sans",
              !request[`description_${activeLang}` as keyof typeof request] && "italic opacity-50"
            )} dir={activeLang === 'ar' ? 'rtl' : 'ltr'}>
              {(request[`description_${activeLang}` as keyof typeof request] as string) ||
                (activeLang === 'ar' ? 'لا يوجد وصف متاح.' : 'No description available in this language.')}
            </p>

            <div className="mt-8 grid grid-cols-2 gap-4">
              {(['ar', 'fr', 'en'] as const)
                .filter(lang => lang !== 'ar' && request[`title_${lang}` as keyof typeof request])
                .map(lang => (
                  <div key={lang} className="p-4 bg-surface-base/50 rounded-xl border border-surface-border">
                    <span className="text-[10px] font-bold text-surface-muted uppercase mb-1 block">
                      العنوان ({lang === 'fr' ? 'فرنسي' : 'إنجليزي'})
                    </span>
                    <span className="text-sm font-bold text-white">{request[`title_${lang}` as keyof typeof request] as string}</span>
                  </div>
              ))}
            </div>
          </div>

          {/* Financial & Time Details Card */}
          <div className="admin-card p-6 mt-6">
            <div className="flex items-center gap-3 mb-6 border-b border-surface-border pb-4">
              <div className="p-2 bg-mazad-primary/20 text-mazad-primary rounded-lg">
                <Gavel className="w-5 h-5" />
              </div>
              <h2 className="font-display font-bold text-white text-lg">التفاصيل المالية والزمنية</h2>
            </div>

            <div className="grid grid-cols-2 lg:grid-cols-3 gap-y-8 gap-x-6">
              {request.market_country_iso && (
                <DetailItem
                  label="السوق"
                  value={`${request.market_country_iso} · ${request.currency_code ?? 'MRU'}`}
                  icon={MapPin}
                />
              )}
              <DetailItem
                label="السعر الابتدائي"
                value={formatPrice(request.start_price, request.currency_code)}
              />
              <DetailItem
                label="الحد الأدنى للزيادة"
                value={formatPrice(request.min_increment, request.currency_code)}
              />
              {request.reserve_price && (
                <DetailItem
                  label="سعر الاحتياط"
                  value={formatPrice(request.reserve_price, request.currency_code)}
                />
              )}
              {request.buy_now_price && (
                <DetailItem
                  label="سعر الشراء الفوري"
                  value={formatPrice(request.buy_now_price, request.currency_code)}
                  color="text-mazad-accent"
                />
              )}
              <DetailItem
                label="مبلغ التأمين"
                value={request.insurance_policy === 'not_required'
                  ? 'بدون تأمين'
                  : hasInsurance ? formatPrice(request.insurance_amount, request.currency_code) : 'غير محدد'}
                icon={ShieldCheck}
                color={request.insurance_policy === 'not_required' || hasInsurance ? 'text-white' : 'text-red-400'}
              />
              <DetailItem
                label="تاريخ البدء"
                value={formatDate(request.start_date)}
                icon={Calendar}
              />
              <DetailItem
                label="تاريخ الانتهاء"
                value={formatDate(request.end_date)}
                icon={Calendar}
              />
              <DetailItem
                label="رقم الطلب"
                value={shortID(request.id)}
                icon={Tag}
              />
            </div>

            {!hasInsurance && isPending && (
              <div className="mt-8 pt-6 border-t border-surface-border/50 bg-red-500/10 border border-red-500/20 rounded-xl p-4 flex items-start gap-2">
                <AlertTriangle className="w-4 h-4 text-red-400 mt-0.5 shrink-0" />
                <p className="text-sm text-red-400">
                  لم يتم تحديد مبلغ التأمين لهذا الطلب. يجب على الإدارة تحديد مبلغ تأمين صالح (أكبر من صفر) قبل الموافقة،
                  وإلا سيتم رفض الموافقة تلقائياً. استخدم زر "تعديل الطلب" لتحديد المبلغ.
                </p>
              </div>
            )}
          </div>

          {/* Rejection reason, if rejected */}
          {request.status === 'rejected' && request.admin_notes && (
            <div className="admin-card p-6 mt-6 border-red-500/20">
              <div className="flex items-center gap-2 mb-3">
                <AlertTriangle className="w-4 h-4 text-red-400" />
                <h3 className="font-bold text-red-400">سبب الرفض</h3>
              </div>
              <p className="text-sm text-surface-muted whitespace-pre-wrap">{request.admin_notes}</p>
            </div>
          )}
        </div>

        {/* Right Column: Status & Actions */}
        <div className="lg:col-span-1 space-y-6">
          <div className="admin-card p-6 border-mazad-accent/20 relative overflow-hidden">
            <div className="absolute top-0 right-0 w-32 h-32 bg-mazad-accent/5 rounded-full -mr-16 -mt-16 blur-2xl" />

            <div className="relative">
              <StatusBadge status={request.status} className="mb-4" />
              <h1 className="text-2xl font-display font-bold text-white mb-4 leading-tight">{request.title_ar}</h1>

              <div className="grid gap-3 mb-6">
                <button
                  onClick={() => navigate(`/users/${request.user_id}`)}
                  className="flex items-center justify-between gap-3 bg-surface-base/60 rounded-2xl p-4 border border-surface-border hover:border-mazad-primary/40 transition-all"
                >
                  <div className="min-w-0 text-left">
                    <p className="text-[10px] font-bold text-surface-muted uppercase mb-1 tracking-widest flex items-center gap-1">
                      <User className="w-3.5 h-3.5 text-mazad-primary" /> المستخدم
                    </p>
                    <p className="text-sm font-bold text-white truncate">
                      {request.user?.full_name || request.user?.phone || shortID(request.user_id)}
                    </p>
                    <p className="text-[10px] text-surface-muted truncate">
                      {request.user?.phone ?? request.user_id}
                    </p>
                  </div>
                </button>

                <div className="bg-surface-base/60 rounded-2xl p-4 border border-surface-border">
                  <p className="text-[10px] font-bold text-surface-muted uppercase mb-1 tracking-widest flex items-center gap-1">
                    <Calendar className="w-3.5 h-3.5 text-mazad-primary" /> تاريخ الإنشاء
                  </p>
                  <p className="text-sm font-bold text-white">{formatDate(request.created_at)}</p>
                </div>

                {request.reviewed_at && (
                  <div className="bg-surface-base/60 rounded-2xl p-4 border border-surface-border">
                    <p className="text-[10px] font-bold text-surface-muted uppercase mb-1 tracking-widest flex items-center gap-1">
                      <FileText className="w-3.5 h-3.5 text-mazad-primary" /> تاريخ المراجعة
                    </p>
                    <p className="text-sm font-bold text-white">{formatDate(request.reviewed_at)}</p>
                  </div>
                )}
              </div>
            </div>

            {isPending && (
              <div className="mt-8 pt-8 border-t border-surface-border space-y-3">
                {!hasInsurance ? (
                  <>
                    <button
                      onClick={() => navigate(`/requests/auctions/${request.id}/edit`)}
                      className="w-full py-3.5 bg-mazad-primary hover:bg-mazad-primary/90 text-white font-bold rounded-xl shadow-lg shadow-mazad-primary/10 transition-all flex items-center justify-center gap-2"
                    >
                      <Pencil className="w-5 h-5" />
                      تحديد مبلغ التأمين وتعديل الطلب
                    </button>
                    <button
                      onClick={() => setRejectDialog(true)}
                      className="w-full py-3.5 bg-red-500/10 hover:bg-red-500 text-red-500 hover:text-white border border-red-500/20 font-bold rounded-xl transition-all flex items-center justify-center gap-2"
                    >
                      <X className="w-5 h-5" />
                      رفض الطلب
                    </button>
                  </>
                ) : (
                  <>
                    <button
                      onClick={() => setApproveConfirm(true)}
                      className="w-full py-3.5 bg-emerald-500 hover:bg-emerald-600 text-white font-bold rounded-xl shadow-lg shadow-emerald-500/10 transition-all flex items-center justify-center gap-2"
                    >
                      <Check className="w-5 h-5" />
                      قبول الطلب
                    </button>
                    <button
                      onClick={() => setRejectDialog(true)}
                      className="w-full py-3.5 bg-red-500/10 hover:bg-red-500 text-red-500 hover:text-white border border-red-500/20 font-bold rounded-xl transition-all flex items-center justify-center gap-2"
                    >
                      <X className="w-5 h-5" />
                      رفض الطلب
                    </button>
                  </>
                )}
              </div>
            )}

            {isDraftOrRejected && (
              <div className="mt-8 pt-8 border-t border-surface-border space-y-3">
                <p className="text-sm text-surface-muted">
                  {request.status === 'draft'
                    ? 'يمكنك تعديل هذه المسودة قبل إرسالها للمراجعة.'
                    : 'يمكنك تعديل هذا الطلب المرفوض وإعادة إرساله للمراجعة.'}
                </p>
                <button
                  onClick={() => navigate(`/requests/auctions/${request.id}/edit`)}
                  className="w-full py-3.5 bg-mazad-primary hover:bg-mazad-primary/90 text-white font-bold rounded-xl shadow-lg shadow-mazad-primary/10 transition-all flex items-center justify-center gap-2"
                >
                  <Pencil className="w-5 h-5" />
                  تعديل الطلب
                </button>
              </div>
            )}

            <div className="mt-6 pt-6 border-t border-surface-border">
              <button
                onClick={() => setDeleteConfirm(true)}
                className="w-full py-3 text-sm font-bold text-red-400 hover:text-red-300 transition-colors"
              >
                حذف الطلب
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Media Overlay */}
      {activeImage && (
        <ImagePreview
          fullScreen
          src={activeImage}
          onClose={() => setActiveImage(null)}
        />
      )}

      {/* Confirmation Dialogs */}
      <ConfirmDialog
        open={approveConfirm}
        onOpenChange={setApproveConfirm}
        title="هل تود الموافقة على هذا الطلب؟"
        description="سيتم إرسال إشعار للمستخدم وسيصبح المزاد متاحاً."
        confirmLabel="تأكيد القبول"
        variant="success"
        loading={review.isPending}
        onConfirm={() => handleReview(true)}
      />

      {/* Reject Modal */}
      {rejectDialog && (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4 backdrop-blur-sm">
          <div className="admin-card p-8 w-full max-w-md animate-slide-in relative overflow-hidden">
            <h3 className="font-display font-bold text-white text-xl mb-1">رفض الطلب</h3>
            <p className="text-sm text-surface-muted mb-6">يرجى كتابة سبب الرفض ليتم إخطار المستخدم.</p>
            <textarea
              value={rejectionReason}
              onChange={(e) => setRejectionReason(e.target.value)}
              placeholder="اكتب السبب هنا..."
              rows={4}
              className="w-full bg-surface-base border border-surface-border rounded-xl p-4 text-sm text-white resize-none focus:outline-none focus:border-red-500 mb-6"
            />
            <div className="flex gap-3 justify-end leading-none">
              <button
                onClick={() => setRejectDialog(false)}
                className="px-6 py-2.5 rounded-xl text-sm font-bold text-surface-muted border border-surface-border hover:bg-surface-border/50 transition-all"
              >
                إلغاء
              </button>
              <button
                disabled={!rejectionReason.trim() || review.isPending}
                onClick={() => handleReview(false)}
                className="px-6 py-2.5 rounded-xl text-sm font-bold bg-red-500 text-white disabled:opacity-40 transition-all shadow-lg shadow-red-500/20 flex items-center gap-2"
              >
                {review.isPending && <Loader2 className="w-4 h-4 animate-spin" />}
                تأكيد الرفض
              </button>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={deleteConfirm}
        onOpenChange={setDeleteConfirm}
        title="هل تود حذف هذا الطلب؟"
        description="هذا الإجراء لا يمكن التراجع عنه."
        confirmLabel="تأكيد الحذف"
        variant="danger"
        loading={deleteRequest.isPending}
        onConfirm={() => {
          if (!id) return
          deleteRequest.mutate(id, {
            onSuccess: () => {
              setDeleteConfirm(false)
              navigate('/requests')
            }
          })
        }}
      />
    </div>
  )
}
