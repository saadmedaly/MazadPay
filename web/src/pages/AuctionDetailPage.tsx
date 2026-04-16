import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { 
  ArrowLeft, 
  MapPin, 
  Tag, 
  Clock, 
  Check, 
  X, 
  AlertTriangle,
  Gavel,
  Loader2,
  AlertCircle,
  ImageIcon
} from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { StatusBadge } from '@/components/shared/StatusBadge'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { ImagePreview } from '@/components/shared/ImagePreview'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { useAuction, useValidateAuction } from '@/hooks/useAuctions'
import { formatPrice, formatDate, shortID } from '@/lib/formatters'
import { cn } from '@/lib/utils'

export function AuctionDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [rejectDialog, setRejectDialog] = useState(false)
  const [rejectionReason, setRejectionReason] = useState('')
  const [approveConfirm, setApproveConfirm] = useState(false)
  const [activeImage, setActiveImage] = useState<string | null>(null)

  const { data: auction, isLoading, isError } = useAuction(id!)
  const validate = useValidateAuction()

  const handleValidate = (approve: boolean) => {
    validate.mutate(
      { id: id!, approve, reason: rejectionReason },
      { onSuccess: () => navigate('/auctions') }
    )
  }

  if (isLoading) return <LoadingSpinner fullPage label="جاري تحميل تفاصيل المزاد..." />

  if (isError || !auction) return (
    <div className="flex flex-col items-center justify-center h-64 text-surface-muted gap-4">
      <AlertCircle className="w-12 h-12 opacity-20" />
      <p className="font-bold">فشل في العثور على المزاد</p>
      <button onClick={() => navigate(-1)} className="text-mazad-primary text-sm font-bold">رجوع للوراء</button>
    </div>
  )

  const isPending = auction.status === 'pending'
  const images = auction.images?.length ? auction.images : []

  return (
    <div className="animate-fade-in max-w-6xl" dir="rtl">
      <PageHeader title="تفاصيل المزاد">
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
                   className="w-full h-full object-contain"
                />
              ) : (
                <div className="w-full h-full flex flex-col items-center justify-center text-surface-muted gap-3">
                  <ImageIcon className="w-12 h-12 opacity-20" />
                  <span className="text-sm font-bold">لا يوجد صور لهذا المزاد</span>
                </div>
              )}
            </div>
            {images.length > 1 && (
              <div className="p-4 flex gap-3 overflow-x-auto border-t border-surface-border bg-surface-base/30">
                {images.map((img, i) => (
                  <button 
                    key={i} 
                    onClick={() => setActiveImage(img)}
                    className="w-20 h-20 rounded-lg overflow-hidden border border-surface-border hover:border-mazad-primary transition-all shrink-0"
                  >
                    <img src={img} className="w-full h-full object-cover" alt="" />
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Details Tabs placeholder */}
          <div className="admin-card p-6">
            <h2 className="font-display font-bold text-white text-lg mb-4 border-b border-surface-border pb-4">وصف المزاد والمواصفات</h2>
            <p className="text-sm text-surface-muted leading-relaxed font-medium">
              {auction.description || 'لا يوجد وصف متاح لهذا المزاد.'}
            </p>

            <div className="mt-8 grid grid-cols-2 gap-4">
              <div className="p-4 bg-surface-base/50 rounded-xl border border-surface-border">
                <span className="text-[10px] font-bold text-surface-muted uppercase mb-1 block">الفئة</span>
                <span className="text-sm font-bold text-white">{auction.category || 'غير محدد'}</span>
              </div>
              <div className="p-4 bg-surface-base/50 rounded-xl border border-surface-border">
                <span className="text-[10px] font-bold text-surface-muted uppercase mb-1 block">الموقع</span>
                <span className="text-sm font-bold text-white flex items-center gap-2">
                  <MapPin className="w-3.5 h-3.5 text-mazad-primary" />
                  {auction.city || 'نواكشوط'}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Right Column: Bidding & Actions */}
        <div className="lg:col-span-1 space-y-6">
          <div className="admin-card p-6 border-mazad-accent/20 relative overflow-hidden">
            <div className="absolute top-0 right-0 w-32 h-32 bg-mazad-accent/5 rounded-full -mr-16 -mt-16 blur-2xl" />
            
            <div className="relative">
              <StatusBadge status={auction.status} className="mb-4" />
              <h1 className="text-2xl font-display font-bold text-white mb-6 leading-tight">{auction.title}</h1>
              
              <div className="space-y-6">
                 <div>
                    <span className="text-xs font-bold text-surface-muted uppercase tracking-widest block mb-1">السعر الحالي</span>
                    <div className="text-3xl font-display font-bold text-mazad-accent">{formatPrice(parseFloat(auction.current_price))}</div>
                 </div>
                 
                 <div className="grid grid-cols-2 gap-4 border-t border-surface-border pt-6">
                    <div>
                       <span className="text-[10px] font-bold text-surface-muted uppercase block">المزايدين</span>
                       <div className="text-lg font-bold text-white">{auction.bidder_count}</div>
                    </div>
                    <div>
                       <span className="text-[10px] font-bold text-surface-muted uppercase block">وقت البدء</span>
                       <div className="text-sm font-medium text-white">{formatDate(auction.start_time )}</div>
                    </div>
                 </div>
              </div>
            </div>

            {isPending && (
              <div className="mt-8 pt-8 border-t border-surface-border space-y-3">
                <button 
                  onClick={() => setApproveConfirm(true)}
                  className="w-full py-3.5 bg-emerald-500 hover:bg-emerald-600 text-white font-bold rounded-xl shadow-lg shadow-emerald-500/10 transition-all flex items-center justify-center gap-2"
                >
                  <Check className="w-5 h-5" />
                   الموافقة وبدء المزاد
                </button>
                <button 
                  onClick={() => setRejectDialog(true)}
                  className="w-full py-3.5 bg-red-500/10 hover:bg-red-500 text-red-500 hover:text-white border border-red-500/20 font-bold rounded-xl transition-all flex items-center justify-center gap-2"
                >
                  <X className="w-5 h-5" />
                   رفض الطلب
                </button>
              </div>
            )}
          </div>

          <div className="admin-card p-6">
            <h3 className="text-xs font-bold text-surface-muted uppercase tracking-widest mb-4 flex items-center gap-2">
               <Gavel className="w-3.5 h-3.5 text-mazad-primary" />
               آخر المزايدات
            </h3>
            <div className="space-y-4">
               {/* Simplified list check */}
               <div className="text-center py-8 opacity-20">
                  <Clock className="w-8 h-8 mx-auto mb-2" />
                  <p className="text-xs font-bold">لا توجد مزايدات بعد</p>
               </div>
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
        title="هل تود الموافقة على هذا المزاد؟"
        description="سيتم إرسال إشعار للبائع وسيصبح المزاد نشطاً للجمهور فوراً."
        confirmLabel="موافقة بـث"
        variant="success"
        loading={validate.isPending}
        onConfirm={() => handleValidate(true)}
      />

      {/* Reject Modal */}
      {rejectDialog && (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4 backdrop-blur-sm">
          <div className="admin-card p-8 w-full max-w-md animate-slide-in relative overflow-hidden">
            <h3 className="font-display font-bold text-white text-xl mb-1">رفض المزاد</h3>
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
                className="px-6 py-2.5 rounded-xl text-sm font-bold text-surface-muted border border-surface-border hover:bg-surface-border/50 transition-all font-bold"
              >
                إلغاء
              </button>
              <button 
                disabled={!rejectionReason.trim() || validate.isPending}
                onClick={() => handleValidate(false)}
                className="px-6 py-2.5 rounded-xl text-sm font-bold bg-red-500 text-white disabled:opacity-40 transition-all shadow-lg shadow-red-500/20 flex items-center gap-2"
              >
                {validate.isPending && <Loader2 className="w-4 h-4 animate-spin" />}
                تأكيد الرفض
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
