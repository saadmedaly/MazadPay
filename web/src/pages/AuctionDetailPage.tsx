import { useState, useEffect, useCallback, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { 
  ArrowLeft, 
  MapPin, 
  Tag, 
  Clock, 
  Check, 
  X, 
  Gavel,
  Loader2,
  AlertCircle,
  ImageIcon,
  ChevronRight,
  ChevronLeft,
  Eye,
  Calendar,
  User,
  Send,
  TrendingUp,
} from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { StatusBadge } from '@/components/shared/StatusBadge'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { ImagePreview } from '@/components/shared/ImagePreview'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { useAuction, useValidateAuction } from '@/hooks/useAuctions'
import { useBidHistory } from '@/hooks/useBids'
import { formatPrice, formatDate, shortID } from '@/lib/formatters'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/authStore'

function DetailItem({ label, value, icon: Icon, color = "text-white" }: { label: string, value: string | number, icon?: any, color?: string }) {
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

export function AuctionDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [rejectDialog, setRejectDialog] = useState(false)
  const [rejectionReason, setRejectionReason] = useState('')
  const [approveConfirm, setApproveConfirm] = useState(false)
  const [activeImage, setActiveImage] = useState<string | null>(null)
  const [now, setNow] = useState(Date.now())
  const [bidAmount, setBidAmount] = useState('')
  const [wsConnected, setWsConnected] = useState(false)
  const [currentPrice, setCurrentPrice] = useState(0)
  const [bidHistory, setBidHistory] = useState<any[]>([])

  const { data: auction, isLoading, isError } = useAuction(id!)
  const { data: bidHistoryData = [] } = useBidHistory(id!)
  const validate = useValidateAuction()
  const [selectedImg, setSelectedImg] = useState<string | null>(null)
  const thumbScrollRef = useRef<HTMLDivElement>(null)
  const [activeLang, setActiveLang] = useState<'ar' | 'fr' | 'en'>('ar')

  const { token } = useAuthStore()

  // WebSocket connection for real-time bidding
  useEffect(() => {
    if (!auction) return

    const wsUrl = `ws://localhost:8082/ws/auction/${id}?token=${token || ''}`
    const ws = new WebSocket(wsUrl)


    ws.onopen = () => {
      console.log('WebSocket connected')
      setWsConnected(true)
    }

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data)
      
      if (data.type === 'bid_placed') {
        setCurrentPrice(data.payload.new_price)
        setBidHistory(prev => [data.payload, ...prev.slice(0, 9)])
      }
      
      if (data.type === 'auction_ended') {
        // Handle auction end
        console.log('Auction ended:', data.payload)
      }
    }

    ws.onclose = () => {
      console.log('WebSocket disconnected')
      setWsConnected(false)
    }

    return () => {
      ws.close()
    }
  }, [id, auction])

  // Update current price when auction data loads
  useEffect(() => {
    if (auction) {
      setCurrentPrice(parseFloat(auction.current_price))
    }
  }, [auction])

   const images = (() => {
    if (!auction) return []
    const base = auction.images || []
    
    let extra: string[] = []
    const details = auction.item_details || {}
    const possibleKeys = ['images', 'id_images', 'photos', 'gallery', 'item_images']
    
    possibleKeys.forEach(key => {
      const val = details[key]
      if (Array.isArray(val)) {
        extra = [...extra, ...val.map(v => String(v))]
      } else if (typeof val === 'string' && val.includes(',')) {
        extra = [...extra, ...val.split(',').map(s => s.trim())]
      } else if (typeof val === 'string' && val.length > 0) {
        extra = [...extra, val]
      }
    })
    
    const combined = [...base, ...extra].filter(Boolean)
    return Array.from(new Set(combined)) // Unique only
  })()

  // Navigation helpers
  const currentIndex = selectedImg ? images.indexOf(selectedImg) : -1
  
  const handleNext = useCallback(() => {
    if (images.length <= 1) return
    const nextIdx = (currentIndex + 1) % images.length
    setSelectedImg(images[nextIdx])
  }, [currentIndex, images])

  const handlePrev = useCallback(() => {
    if (images.length <= 1) return
    const prevIdx = (currentIndex - 1 + images.length) % images.length
    setSelectedImg(images[prevIdx])
  }, [currentIndex, images])

  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'ArrowRight') handlePrev() // RTL direction
      if (e.key === 'ArrowLeft') handleNext()
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handleNext, handlePrev])

  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 1000)
    return () => window.clearInterval(timer)
  }, [])

  // Scroll active thumbnail into view
  useEffect(() => {
    if (currentIndex >= 0 && thumbScrollRef.current) {
      const activeThumb = thumbScrollRef.current.children[currentIndex] as HTMLElement
      if (activeThumb) {
        const container = thumbScrollRef.current
        const scrollLeft = activeThumb.offsetLeft - container.offsetWidth / 2 + activeThumb.offsetWidth / 2
        container.scrollTo({ left: scrollLeft, behavior: 'smooth' })
      }
    }
  }, [currentIndex])

  // Initialize selected image when data finishes loading
  if (images.length > 0 && !selectedImg && !isLoading) {
    setSelectedImg(images[0])
  }

  const handleValidate = (approve: boolean) => {
    validate.mutate(
      { id: id!, approve, reason: rejectionReason },
      { onSuccess: () => navigate('/auctions') }
    )
  }

  const formatImgUrl = (url: any) => {
    if (!url) return ''
    const realUrl = typeof url === 'string' ? url : (url.url || '')
    if (!realUrl || typeof realUrl !== 'string') return ''

    if (realUrl.startsWith('http') || realUrl.startsWith('data:')) return realUrl
    const baseUrl = import.meta.env.VITE_API_URL ?? 'http://localhost:8082'
    return `${baseUrl}${realUrl.startsWith('/') ? '' : '/'}${realUrl}`
  }

  const formatRemainingTime = (endTime: string) => {
    const diff = new Date(endTime).getTime() - now
    if (isNaN(diff)) return 'غير متوفر'
    if (diff <= 0) return 'انتهى'

    const days = Math.floor(diff / 86400000)
    const hours = Math.floor((diff % 86400000) / 3600000)
    const minutes = Math.floor((diff % 3600000) / 60000)
    const seconds = Math.floor((diff % 60000) / 1000)

    return [
      days > 0 ? `${days}ي` : null,
      hours > 0 ? `${hours}س` : null,
      minutes > 0 ? `${minutes}د` : null,
      `${seconds}ث`,
    ].filter(Boolean).join(' ')
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
              {selectedImg ? (
                <>
                  <ImagePreview 
                     src={formatImgUrl(selectedImg)} 
                     className="w-full h-full object-contain cursor-zoom-in transition-transform group-hover:scale-[1.02]"
                     onClick={() => setActiveImage(selectedImg)}
                  />
                  
                  {/* Prev/Next Buttons */}
                  {images.length > 1 && (
                    <>
                      <button 
                        onClick={(e) => { e.stopPropagation(); handlePrev(); }}
                        className="absolute left-4 top-1/2 -translate-y-1/2 w-10 h-10 bg-black/40 hover:bg-mazad-primary backdrop-blur-md rounded-full flex items-center justify-center text-white transition-all opacity-0 group-hover:opacity-100 border border-white/10"
                      >
                        <ChevronLeft className="w-6 h-6" />
                      </button>
                      <button 
                        onClick={(e) => { e.stopPropagation(); handleNext(); }}
                        className="absolute right-4 top-1/2 -translate-y-1/2 w-10 h-10 bg-black/40 hover:bg-mazad-primary backdrop-blur-md rounded-full flex items-center justify-center text-white transition-all opacity-0 group-hover:opacity-100 border border-white/10"
                      >
                        <ChevronRight className="w-6 h-6" />
                      </button>
                    </>
                  )}

                  {/* Image Counter */}
                  {images.length > 0 && (
                    <div className="absolute bottom-4 left-4 bg-black/60 backdrop-blur-md px-3 py-1.5 rounded-full text-[10px] font-bold text-white border border-white/10 z-10 transition-opacity">
                      {currentIndex + 1} / {images.length}
                    </div>
                  )}
                </>
              ) : (
                <div className="w-full h-full flex flex-col items-center justify-center text-surface-muted gap-3">
                  <ImageIcon className="w-12 h-12 opacity-20" />
                  <span className="text-sm font-bold">لا يوجد صور لهذا المزاد</span>
                </div>
              )}
            </div>
            {images.length > 1 && (
              <div 
                ref={thumbScrollRef}
                className="p-4 flex gap-3 overflow-x-auto border-t border-surface-border bg-surface-base/30 custom-scrollbar scroll-smooth"
              >
                {images.map((img, i) => (
                  <button 
                    key={i} 
                    onClick={() => setSelectedImg(img)}
                    className={cn(
                      "w-20 h-20 rounded-lg overflow-hidden border transition-all shrink-0",
                      selectedImg === img ? "border-mazad-primary ring-2 ring-mazad-primary/20" : "border-surface-border hover:border-surface-muted"
                    )}
                  >
                    <img src={formatImgUrl(img)} className="w-full h-full object-cover" alt="" />
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
                {activeLang === 'ar' ? 'وصف المزاد والمواصفات' : 
                 activeLang === 'fr' ? 'Description & Spécifications' : 'Description & Specifications'}
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
              "text-sm leading-relaxed font-medium transition-all duration-300",
              activeLang === 'ar' ? "text-right text-surface-muted" : "text-left text-surface-muted font-sans",
              !auction?.[`description_${activeLang}` as keyof typeof auction] && "italic opacity-50"
            )} dir={activeLang === 'ar' ? 'rtl' : 'ltr'}>
              {auction ? (
                (auction[`description_${activeLang}` as keyof typeof auction] as string) || 
                (activeLang === 'ar' ? 'لا يوجد وصف متاح.' : 'No description available in this language.')
              ) : ''}
            </p>

            <div className="mt-8 grid grid-cols-2 gap-4">
              <div className="p-4 bg-surface-base/50 rounded-xl border border-surface-border">
                <span className="text-[10px] font-bold text-surface-muted uppercase mb-1 block">رقم القطعة</span>
                <span className="text-sm font-bold text-white uppercase tracking-wider">{auction.lot_number || shortID(auction.id)}</span>
              </div>
              <div className="p-4 bg-surface-base/50 rounded-xl border border-surface-border">
                <span className="text-[10px] font-bold text-surface-muted uppercase mb-1 block">تاريخ الانتهاء</span>
                <span className="text-sm font-bold text-white">{formatDate(auction.end_time)}</span>
              </div>
            </div>
          </div>

          {/* Auction Details Card */}
          <div className="admin-card p-6 mt-6">
            <div className="flex items-center gap-3 mb-6 border-b border-surface-border pb-4">
              <div className="p-2 bg-mazad-primary/20 text-mazad-primary rounded-lg">
                <Gavel className="w-5 h-5" />
              </div>
              <h2 className="font-display font-bold text-white text-lg">تفاصيل المزاد المالية والزمنية</h2>
            </div>

            <div className="grid grid-cols-2 lg:grid-cols-3 gap-y-8 gap-x-6">
              <DetailItem 
                label="الفئة" 
                value={auction.category || 'غير محدد'} 
                icon={Tag}
              />
              <DetailItem 
                label="الموقع" 
                value={auction.city || 'نواكشوط'} 
                icon={MapPin}
              />
              <DetailItem 
                label="السعر الافتتاحي" 
                value={formatPrice(auction.start_price)} 
                color="text-white"
              />
              <DetailItem 
                label="الحد الأدنى للمزايدة" 
                value={formatPrice(auction.min_increment)} 
                color="text-white"
              />
              <DetailItem 
                label="مبلغ التأمين" 
                value={formatPrice(auction.insurance_amount)} 
                color="text-white"
              />

              {auction.buy_now_price && (
                <DetailItem 
                  label="سعر الشراء المباشر" 
                  value={formatPrice(auction.buy_now_price)} 
                  color="text-mazad-accent"
                />
              )}

              <DetailItem 
                label="تاريخ البدء" 
                value={formatDate(auction.start_time)} 
                icon={Calendar}
              />
              
              <div className="flex items-center gap-6 pt-2">
                <div className="flex flex-col">
                  <span className="text-[10px] font-bold text-surface-muted uppercase mb-1 flex items-center gap-1.5 tracking-wider">
                    <Eye className="w-3 h-3 text-mazad-primary" /> المشاهدات
                  </span>
                  <span className="text-sm font-bold text-white">{auction.views}</span>
                </div>
                <div className="flex flex-col">
                  <span className="text-[10px] font-bold text-surface-muted uppercase mb-1 flex items-center gap-1.5 tracking-wider">
                    <Gavel className="w-3 h-3 text-mazad-primary" /> المزايدات
                  </span>
                  <span className="text-sm font-bold text-white">{auction.bidder_count}</span>
                </div>
              </div>
            </div>

            {/* Technical Specifications (Dynamic from item_details) */}
            {auction.item_details && Object.keys(auction.item_details).length > 0 && (
              <div className="mt-10 pt-8 border-t border-surface-border/50">
                <h3 className="text-[10px] font-bold text-mazad-primary uppercase tracking-[0.2em] mb-6 flex items-center gap-2">
                  <div className="w-1.5 h-1.5 rounded-full bg-mazad-primary" />
                  المواصفات التقنية والخصائص
                </h3>
                <div className="grid grid-cols-2 lg:grid-cols-3 gap-4">
                  {Object.entries(auction.item_details).map(([key, value]) => (
                    <div key={key} className="p-4 bg-surface-base/30 rounded-xl border border-surface-border/50 hover:border-surface-muted/30 transition-colors">
                      <span className="text-[10px] font-bold text-surface-muted block mb-1 opacity-70">{key}</span>
                      <span className="text-sm font-bold text-white break-words">{String(value)}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Right Column: Bidding & Actions */}
        <div className="lg:col-span-1 space-y-6">
          <div className="admin-card p-6 border-mazad-accent/20 relative overflow-hidden">
            <div className="absolute top-0 right-0 w-32 h-32 bg-mazad-accent/5 rounded-full -mr-16 -mt-16 blur-2xl" />
            
            <div className="relative">
              <StatusBadge status={auction.status} className="mb-4" />
              <h1 className="text-2xl font-display font-bold text-white mb-4 leading-tight">{auction.title_ar}</h1>

              <div className="grid gap-3 mb-6">
                <div className="flex items-center justify-between gap-4 bg-surface-base/60 rounded-2xl p-4 border border-surface-border">
                  <div>
                    <p className="text-[10px] font-bold text-surface-muted uppercase mb-1 tracking-widest">الوقت المتبقي</p>
                    <p className="text-sm font-bold text-emerald-300">{formatRemainingTime(auction.end_time)}</p>
                  </div>
                  <div className="flex items-center gap-2 text-xs text-surface-muted">
                    <Clock className="w-4 h-4" />
                    <span>وقت</span>
                  </div>
                </div>
                <button
                  onClick={() => navigate(`/users/${auction.seller_id}`)}
                  className="flex items-center justify-between gap-3 bg-surface-base/60 rounded-2xl p-4 border border-surface-border hover:border-mazad-primary/40 transition-all"
                >
                  <div className="min-w-0 text-left">
                    <p className="text-[10px] font-bold text-surface-muted uppercase mb-1 tracking-widest flex items-center gap-1">
                      <User className="w-3.5 h-3.5 text-mazad-primary" /> البائع
                    </p>
                    <p className="text-sm font-bold text-white truncate">
                      {auction.seller?.full_name || auction.seller?.phone || shortID(auction.seller_id)}
                    </p>
                    <p className="text-[10px] text-surface-muted truncate">
                      {auction.seller?.phone ? auction.seller.phone : auction.seller_id}
                    </p>
                  </div>
                  <Eye className="w-4 h-4 text-mazad-primary" />
                </button>
              </div>

              <div className="space-y-6">
                 {/* Real-time Price Display */}
                 <div className="bg-surface-base/60 rounded-2xl p-6 border border-surface-border">
                    <div className="flex items-center justify-between mb-4">
                      <span className="text-xs font-bold text-surface-muted uppercase tracking-widest flex items-center gap-2">
                        <TrendingUp className="w-3 h-3 text-mazad-primary" />
                        السعر الحالي
                      </span>
                      <div className="flex items-center gap-2">
                        <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
                        <span className="text-xs text-emerald-400">مباشر</span>
                      </div>
                    </div>
                    <div className="text-3xl font-display font-bold text-mazad-accent">{formatPrice(currentPrice)}</div>
                    {wsConnected && (
                      <div className="text-xs text-emerald-400 mt-2 animate-pulse">
                        تحديث مباشر...
                      </div>
                    )}
                 </div>

                 {/* Bidding Interface */}
                 {auction.status === 'active' && (
                   <div className="bg-surface-base/60 rounded-2xl p-6 border border-surface-border">
                     <div className="space-y-4">
                       <div>
                         <label className="text-xs font-bold text-surface-muted uppercase tracking-widest block mb-2">
                           قيمة المزايدة
                         </label>
                         <div className="relative">
                           <input
                             type="number"
                             value={bidAmount}
                             onChange={(e) => setBidAmount(e.target.value)}
                             placeholder="أدخل المبلغ..."
                             className="w-full px-4 py-3 bg-surface-input border border-surface-border rounded-xl text-white text-lg font-bold focus:outline-none focus:border-mazad-primary focus:ring-2 focus:ring-mazad-primary/20"
                             min={parseFloat(auction.current_price) + parseFloat(auction.min_increment)}
                             step={parseFloat(auction.min_increment)}
                           />
                           <span className="absolute left-4 top-1/2 -translate-y-1/2 text-xs text-surface-muted">
                             {formatPrice(parseFloat(auction.current_price) + parseFloat(auction.min_increment))}
                           </span>
                         </div>
                       </div>
                       
                       <button
                         onClick={() => {
                            // Send bid via WebSocket
                            const ws = new WebSocket(`ws://localhost:8082/ws/auction/${id}?token=${token || ''}`)
                            ws.onopen = () => {
                              ws.send(JSON.stringify({
                                type: 'place_bid',
                                amount: parseFloat(bidAmount)
                              }))
                            }
                         }}
                         disabled={!bidAmount || parseFloat(bidAmount) <= currentPrice}
                         className="w-full py-4 bg-mazad-primary hover:bg-mazad-primary/90 text-white font-bold rounded-xl shadow-lg shadow-mazad-primary/20 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center justify-center gap-2"
                       >
                         <Send className="w-5 h-5" />
                         إرسال المزايدة
                       </button>
                     </div>
                   </div>
                 )}

                 {/* Bid History */}
                 {bidHistory.length > 0 && (
                   <div className="bg-surface-base/60 rounded-2xl p-6 border border-surface-border">
                     <h3 className="text-xs font-bold text-surface-muted uppercase tracking-widest mb-4 flex items-center gap-2">
                       <Clock className="w-3 h-3 text-mazad-primary" />
                       آخر المزايدات
                     </h3>
                     <div className="space-y-3 max-h-40 overflow-y-auto">
                       {bidHistory.map((bid, index) => (
                         <div key={index} className="flex items-center justify-between p-3 bg-surface-base/30 rounded-lg border border-surface-border/50">
                           <div className="flex items-center gap-3">
                             <span className="text-xs text-surface-muted">#{bidHistory.length - index}</span>
                             <span className="text-sm font-bold text-white">{bid.bidder_masked}</span>
                           </div>
                           <span className="text-sm font-bold text-mazad-accent">{formatPrice(bid.new_price)}</span>
                         </div>
                       ))}
                     </div>
                   </div>
                 )}
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

            {auction.status === 'rejected' && (
              <div className="mt-8 pt-8 border-t border-surface-border space-y-3">
                <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-4 mb-4">
                  <p className="text-sm text-red-400 font-bold flex items-center gap-2">
                    <X className="w-4 h-4" />
                    تم رفض هذا المزاد
                  </p>
                </div>
                <button
                  onClick={() => setApproveConfirm(true)}
                  className="w-full py-3.5 bg-emerald-500 hover:bg-emerald-600 text-white font-bold rounded-xl shadow-lg shadow-emerald-500/10 transition-all flex items-center justify-center gap-2"
                >
                  <Check className="w-5 h-5" />
                   الموافقة وتفعيل المزاد
                </button>
              </div>
            )}
          </div>

          <div className="admin-card p-6">
            <h3 className="text-xs font-bold text-surface-muted uppercase tracking-widest mb-4 flex items-center gap-2">
               <Gavel className="w-3.5 h-3.5 text-mazad-primary" />
               آخر المزايدات (أعلى 5)
            </h3>
            <div className="space-y-4">
               {bidHistoryData.length > 0 ? (
                 bidHistoryData.slice(0, 5).map((bid, index) => (
                   <div key={bid.id} className="flex items-center justify-between p-3 bg-surface-base/50 rounded-xl border border-surface-border/50">
                     <div className="flex items-center gap-3">
                       <div className="w-8 h-8 rounded-full bg-mazad-primary/20 flex items-center justify-center text-mazad-primary font-bold text-sm">
                         {index + 1}
                       </div>
                       <div>
                         <p className="text-sm font-bold text-white">
                           {bid.bidder_name || 'مستخدم مجهول'}
                         </p>
                         <p className="text-[10px] text-surface-muted">
                           {new Date(bid.created_at).toLocaleDateString('ar-SA')}
                         </p>
                       </div>
                     </div>
                     <div className="text-right">
                       <p className="text-sm font-bold text-mazad-accent">
                         {formatPrice(bid.amount)}
                       </p>
                       {bid.is_winning && (
                         <span className="text-[10px] text-emerald-400">الأعلى</span>
                       )}
                     </div>
                   </div>
                 ))
               ) : (
                 <div className="text-center py-8 opacity-20">
                    <Clock className="w-8 h-8 mx-auto mb-2" />
                    <p className="text-xs font-bold">لا توجد مزايدات بعد</p>
                 </div>
               )}
            </div>
          </div>
        </div>
      </div>

      {/* Media Overlay */}
      {activeImage && (
        <ImagePreview 
          fullScreen 
          src={formatImgUrl(activeImage)} 
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
                className="px-6 py-2.5 rounded-xl text-sm font-bold text-surface-muted border border-surface-border hover:bg-surface-border/50 transition-all"
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
