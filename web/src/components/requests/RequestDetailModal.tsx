import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import { Loader2, User, Calendar, DollarSign, Image as ImageIcon, FileText, CheckCircle, XCircle, MapPin, Pencil, AlertTriangle } from 'lucide-react'
import { useAuctionRequestByID, useBannerRequestByID } from '@/hooks/useRequests'
import type { AuctionRequest, BannerRequest } from '@/hooks/useRequests'
import { StatusBadge } from '@/components/shared/StatusBadge'
import { format } from 'date-fns'
import { ar } from 'date-fns/locale'
import { formatPrice } from '@/lib/formatters'

interface RequestDetailModalProps {
  isOpen: boolean
  onClose: () => void
  type: 'auction' | 'banner'
  requestId: string | null
  onApprove?: (id: string) => void
  onReject?: (id: string) => void
  onDelete?: (id: string) => void
  onEdit?: (id: string) => void
}

export function RequestDetailModal({
  isOpen,
  onClose,
  type,
  requestId,
  onApprove,
  onReject,
  onDelete,
  onEdit
}: RequestDetailModalProps) {
  const { data: auctionRequest, isLoading: isLoadingAuction } = useAuctionRequestByID(
    type === 'auction' ? requestId : null
  )
  const { data: bannerRequest, isLoading: isLoadingBanner } = useBannerRequestByID(
    type === 'banner' ? requestId : null
  )

  const isLoading = isLoadingAuction || isLoadingBanner
  const request = type === 'auction' ? auctionRequest : bannerRequest

  const getStatusBadge = (status: string) => <StatusBadge status={status} />

  const renderAuctionDetails = (req: AuctionRequest) => {
    // Client feedback A7 follow-up: mirrors the backend's approval gate
    // (ReviewAuctionRequest rejects with request_insurance_not_set when
    // insurance_amount <= 0) so the admin sees the same rule here before
    // clicking Approve, not just as a failed request afterward.
    const hasInsurance = Number(req.insurance_amount) > 0
    return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <User className="w-5 h-5 text-gray-500" />
          <div>
            <p className="text-sm text-gray-500">المستخدم</p>
            <p className="font-medium">{req.user?.full_name || req.user?.phone || 'غير معروف'}</p>
          </div>
        </div>
        {getStatusBadge(req.status)}
      </div>

      {/* Market/currency (migration 000046, Phase 2): server-derived from the
          requester's account market, shown here for admin visibility only --
          never editable, and Admin stays global (no FX conversion). */}
      {(req.market_country_iso || req.currency_code) && (
        <div className="flex items-center gap-3 bg-indigo-50 p-3 rounded-lg w-fit">
          <MapPin className="w-4 h-4 text-indigo-600" />
          <p className="text-sm text-indigo-700 font-medium">
            السوق: {req.market_country_iso ?? '—'} · العملة: {req.currency_code ?? 'MRU'}
          </p>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-gray-50 p-4 rounded-lg">
          <p className="text-sm text-gray-500 mb-1">العنوان (عربي)</p>
          <p className="font-medium">{req.title_ar}</p>
        </div>
        {req.title_fr && (
          <div className="bg-gray-50 p-4 rounded-lg">
            <p className="text-sm text-gray-500 mb-1">العنوان (فرنسي)</p>
            <p className="font-medium">{req.title_fr}</p>
          </div>
        )}
        {req.title_en && (
          <div className="bg-gray-50 p-4 rounded-lg">
            <p className="text-sm text-gray-500 mb-1">العنوان (إنجليزي)</p>
            <p className="font-medium">{req.title_en}</p>
          </div>
        )}
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-blue-50 p-4 rounded-lg">
          <div className="flex items-center gap-2 mb-2">
            <DollarSign className="w-4 h-4 text-blue-600" />
            <p className="text-sm text-gray-600">السعر الابتدائي</p>
          </div>
          <p className="text-lg font-bold text-blue-700">{formatPrice(req.start_price, req.currency_code)}</p>
        </div>

        {req.reserve_price && (
          <div className="bg-purple-50 p-4 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <DollarSign className="w-4 h-4 text-purple-600" />
              <p className="text-sm text-gray-600">سعر الاحتياط</p>
            </div>
            <p className="text-lg font-bold text-purple-700">{formatPrice(req.reserve_price, req.currency_code)}</p>
          </div>
        )}

        {req.buy_now_price && (
          <div className="bg-green-50 p-4 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <DollarSign className="w-4 h-4 text-green-600" />
              <p className="text-sm text-gray-600">سعر الشراء الفوري</p>
            </div>
            <p className="text-lg font-bold text-green-700">{formatPrice(req.buy_now_price, req.currency_code)}</p>
          </div>
        )}

        <div className="bg-orange-50 p-4 rounded-lg">
          <div className="flex items-center gap-2 mb-2">
            <DollarSign className="w-4 h-4 text-orange-600" />
            <p className="text-sm text-gray-600">الحد الأدنى للزيادة</p>
          </div>
          <p className="text-lg font-bold text-orange-700">{formatPrice(req.min_increment, req.currency_code)}</p>
        </div>

        {/* Client feedback A7 follow-up: insurance is staff/admin-only -- the
            user-side form never collects it, so it must be visible here so the
            admin can see/fix it (via "تعديل الطلب", which already carries the
            field) before approving. Backend independently rejects approval
            with insurance_amount <= 0 regardless of what's shown here. */}
        <div className={`p-4 rounded-lg ${hasInsurance ? 'bg-teal-50' : 'bg-red-50 border border-red-200'}`}>
          <div className="flex items-center gap-2 mb-2">
            <DollarSign className={`w-4 h-4 ${hasInsurance ? 'text-teal-600' : 'text-red-600'}`} />
            <p className="text-sm text-gray-600">مبلغ التأمين</p>
          </div>
          <p className={`text-lg font-bold ${hasInsurance ? 'text-teal-700' : 'text-red-700'}`}>
            {hasInsurance ? formatPrice(req.insurance_amount, req.currency_code) : 'غير محدد'}
          </p>
        </div>
      </div>

      {!hasInsurance && req.status === 'pending' && (
        <div className="bg-red-100 border border-red-200 p-4 rounded-lg flex items-start gap-2">
          <AlertTriangle className="w-4 h-4 text-red-800 mt-0.5 shrink-0" />
          <p className="text-sm text-red-800">
            لم يتم تحديد مبلغ التأمين لهذا الطلب. يجب على الإدارة تحديد مبلغ تأمين صالح (أكبر من صفر) قبل الموافقة،
            وإلا سيتم رفض الموافقة تلقائياً. استخدم زر "تعديل الطلب" لتحديد المبلغ.
          </p>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="flex items-center gap-3 bg-gray-50 p-4 rounded-lg">
          <Calendar className="w-5 h-5 text-gray-500" />
          <div>
            <p className="text-sm text-gray-500">تاريخ البدء</p>
            <p className="font-medium">
              {format(new Date(req.start_date), 'PPP p', { locale: ar })}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-3 bg-gray-50 p-4 rounded-lg">
          <Calendar className="w-5 h-5 text-gray-500" />
          <div>
            <p className="text-sm text-gray-500">تاريخ الانتهاء</p>
            <p className="font-medium">
              {format(new Date(req.end_date), 'PPP p', { locale: ar })}
            </p>
          </div>
        </div>
      </div>

      {req.description_ar && (
        <div className="bg-gray-50 p-4 rounded-lg">
          <div className="flex items-center gap-2 mb-2">
            <FileText className="w-4 h-4 text-gray-500" />
            <p className="text-sm text-gray-500">الوصف (عربي)</p>
          </div>
          <p className="text-gray-700 whitespace-pre-wrap">{req.description_ar}</p>
        </div>
      )}

      {req.description_fr && (
        <div className="bg-gray-50 p-4 rounded-lg">
          <div className="flex items-center gap-2 mb-2">
            <FileText className="w-4 h-4 text-gray-500" />
            <p className="text-sm text-gray-500">الوصف (فرنسي)</p>
          </div>
          <p className="text-gray-700 whitespace-pre-wrap">{req.description_fr}</p>
        </div>
      )}

      {req.description_en && (
        <div className="bg-gray-50 p-4 rounded-lg">
          <div className="flex items-center gap-2 mb-2">
            <FileText className="w-4 h-4 text-gray-500" />
            <p className="text-sm text-gray-500">الوصف (إنجليزي)</p>
          </div>
          <p className="text-gray-700 whitespace-pre-wrap">{req.description_en}</p>
        </div>
      )}

      {req.images && req.images.length > 0 && (
        <div>
          <div className="flex items-center gap-2 mb-3">
            <ImageIcon className="w-4 h-4 text-gray-500" />
            <p className="text-sm text-gray-500">الصور ({req.images.length})</p>
          </div>
          <div className="grid grid-cols-4 gap-2">
            {req.images.map((img: string, index: number) => (
              <a
                key={index}
                href={img}
                target="_blank"
                rel="noopener noreferrer"
                className="block"
              >
                <img
                  src={img}
                  alt={`صورة ${index + 1}`}
                  className="w-full h-24 object-cover rounded-lg hover:opacity-80 transition-opacity"
                />
              </a>
            ))}
          </div>
        </div>
      )}

      {req.status === 'rejected' && req.admin_notes && (
        <div className="bg-red-100 border border-red-200 p-4 rounded-lg">
          <div className="flex items-center gap-2 mb-2">
            <AlertTriangle className="w-4 h-4 text-red-800" />
            <p className="text-sm font-bold text-red-800">سبب الرفض</p>
          </div>
          <p className="text-red-800 whitespace-pre-wrap">{req.admin_notes}</p>
        </div>
      )}

      <div className="text-sm text-gray-500 border-t pt-4">
        <p>تاريخ الإنشاء: {format(new Date(req.created_at), 'PPP p', { locale: ar })}</p>
        {req.reviewed_at && (
          <p>تاريخ المراجعة: {format(new Date(req.reviewed_at), 'PPP p', { locale: ar })}</p>
        )}
      </div>
    </div>
    )
  }

  const renderBannerDetails = (req: BannerRequest) => (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <User className="w-5 h-5 text-gray-500" />
          <div>
            <p className="text-sm text-gray-500">المستخدم</p>
            <p className="font-medium">{req.user?.full_name || req.user?.phone || 'غير معروف'}</p>
          </div>
        </div>
        {getStatusBadge(req.status)}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-gray-50 p-4 rounded-lg">
          <p className="text-sm text-gray-500 mb-1">العنوان (عربي)</p>
          <p className="font-medium">{req.title_ar}</p>
        </div>
        {req.title_fr && (
          <div className="bg-gray-50 p-4 rounded-lg">
            <p className="text-sm text-gray-500 mb-1">العنوان (فرنسي)</p>
            <p className="font-medium">{req.title_fr}</p>
          </div>
        )}
        {req.title_en && (
          <div className="bg-gray-50 p-4 rounded-lg">
            <p className="text-sm text-gray-500 mb-1">العنوان (إنجليزي)</p>
            <p className="font-medium">{req.title_en}</p>
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="bg-blue-50 p-4 rounded-lg">
          <div className="flex items-center gap-2 mb-2">
            <Calendar className="w-4 h-4 text-blue-600" />
            <p className="text-sm text-gray-600">تاريخ البدء</p>
          </div>
          <p className="text-lg font-medium text-blue-700">
            {format(new Date(req.starts_at), 'PPP p', { locale: ar })}
          </p>
        </div>

        <div className="bg-purple-50 p-4 rounded-lg">
          <div className="flex items-center gap-2 mb-2">
            <Calendar className="w-4 h-4 text-purple-600" />
            <p className="text-sm text-gray-600">تاريخ الانتهاء</p>
          </div>
          <p className="text-lg font-medium text-purple-700">
            {format(new Date(req.ends_at), 'PPP p', { locale: ar })}
          </p>
        </div>
      </div>
      
      {req.description_ar && (
        <div className="bg-gray-50 p-4 rounded-lg">
          <div className="flex items-center gap-2 mb-2">
            <FileText className="w-4 h-4 text-gray-500" />
            <p className="text-sm text-gray-500">الوصف (عربي)</p>
          </div>
          <p className="text-gray-700 whitespace-pre-wrap">{req.description_ar}</p>
        </div>
      )}

      {req.description_fr && (
        <div className="bg-gray-50 p-4 rounded-lg">
          <div className="flex items-center gap-2 mb-2">
            <FileText className="w-4 h-4 text-gray-500" />
            <p className="text-sm text-gray-500">الوصف (فرنسي)</p>
          </div>
          <p className="text-gray-700 whitespace-pre-wrap">{req.description_fr}</p>
        </div>
      )}

      {req.description_en && (
        <div className="bg-gray-50 p-4 rounded-lg">
          <div className="flex items-center gap-2 mb-2">
            <FileText className="w-4 h-4 text-gray-500" />
            <p className="text-sm text-gray-500">الوصف (إنجليزي)</p>
          </div>
          <p className="text-gray-700 whitespace-pre-wrap">{req.description_en}</p>
        </div>
      )}

      <div className="bg-gray-50 p-4 rounded-lg">
        <div className="flex items-center gap-2 mb-2">
          <ImageIcon className="w-4 h-4 text-gray-500" />
          <p className="text-sm text-gray-500">صورة البانر</p>
        </div>
        <a href={req.image_url} target="_blank" rel="noopener noreferrer" className="block">
          <img
            src={req.image_url}
            alt={req.title_ar}
            className="w-full max-w-md h-48 object-cover rounded-lg hover:opacity-80 transition-opacity"
          />
        </a>
      </div>

      {req.target_url && (
        <div className="bg-green-50 p-4 rounded-lg">
          <div className="flex items-center gap-2 mb-2">
            <MapPin className="w-4 h-4 text-green-600" />
            <p className="text-sm text-gray-600">رابط الوجهة</p>
          </div>
          <a
            href={req.target_url}
            target="_blank"
            rel="noopener noreferrer"
            className="text-green-700 hover:underline break-all"
          >
            {req.target_url}
          </a>
        </div>
      )}

      <div className="text-sm text-gray-500 border-t pt-4">
        <p>تاريخ الإنشاء: {format(new Date(req.created_at), 'PPP p', { locale: ar })}</p>
        {req.reviewed_at && (
          <p>تاريخ المراجعة: {format(new Date(req.reviewed_at), 'PPP p', { locale: ar })}</p>
        )}
      </div>
    </div>
  )

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold">
            تفاصيل طلب {type === 'auction' ? 'المزاد' : 'البانر'}
          </DialogTitle>
        </DialogHeader>

        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
          </div>
        ) : !request ? (
          <div className="text-center py-12 text-gray-500">
            لم يتم العثور على الطلب
          </div>
        ) : (
          <>
            <Tabs defaultValue="details" className="mt-4">
              <TabsList className={`grid w-full ${(request.status === 'pending' || request.status === 'draft' || request.status === 'rejected') ? 'grid-cols-2' : 'grid-cols-1'}`}>
                <TabsTrigger value="details">التفاصيل</TabsTrigger>
                {(request.status === 'pending' || request.status === 'draft' || request.status === 'rejected') && (
                  <TabsTrigger value="actions">الإجراءات</TabsTrigger>
                )}
              </TabsList>

              <TabsContent value="details" className="mt-4">
                {type === 'auction'
                  ? renderAuctionDetails(request as AuctionRequest)
                  : renderBannerDetails(request as BannerRequest)
                }
              </TabsContent>

              {request.status === 'pending' && (
                <TabsContent value="actions" className="mt-4">
                  <div className="space-y-4">
                    <p className="text-gray-600">اختر الإجراء المناسب لهذا الطلب:</p>
                    {/* Client feedback A7 follow-up: block Approve client-side when
                        insurance is missing on an auction request, mirroring the
                        backend's ReviewAuctionRequest gate -- offer Edit instead so
                        the admin can set it (AdminUpdateAuctionRequest) without a
                        failed-approval round trip. Backend still enforces this
                        independently either way. */}
                    {type === 'auction' && !(Number((request as AuctionRequest).insurance_amount) > 0) ? (
                      <div className="flex gap-3">
                        <Button
                          onClick={() => onEdit?.(request.id)}
                          className="bg-blue-600 hover:bg-blue-700 flex-1"
                        >
                          <Pencil className="w-4 h-4 mr-2" />
                          تحديد مبلغ التأمين وتعديل الطلب
                        </Button>
                        <Button
                          onClick={() => onReject?.(request.id)}
                          variant="destructive"
                          className="flex-1"
                        >
                          <XCircle className="w-4 h-4 mr-2" />
                          رفض الطلب
                        </Button>
                      </div>
                    ) : (
                      <div className="flex gap-3">
                        <Button
                          onClick={() => onApprove?.(request.id)}
                          className="bg-green-600 hover:bg-green-700 flex-1"
                        >
                          <CheckCircle className="w-4 h-4 mr-2" />
                          قبول الطلب
                        </Button>
                        <Button
                          onClick={() => onReject?.(request.id)}
                          variant="destructive"
                          className="flex-1"
                        >
                          <XCircle className="w-4 h-4 mr-2" />
                          رفض الطلب
                        </Button>
                      </div>
                    )}
                  </div>
                </TabsContent>
              )}

              {(request.status === 'draft' || request.status === 'rejected') && type === 'auction' && (
                <TabsContent value="actions" className="mt-4">
                  <div className="space-y-4">
                    <p className="text-gray-600">
                      {request.status === 'draft'
                        ? 'يمكنك تعديل هذه المسودة قبل إرسالها للمراجعة.'
                        : 'يمكنك تعديل هذا الطلب المرفوض وإعادة إرساله للمراجعة.'}
                    </p>
                    <Button
                      onClick={() => onEdit?.(request.id)}
                      className="bg-blue-600 hover:bg-blue-700 w-full"
                    >
                      <Pencil className="w-4 h-4 mr-2" />
                      تعديل الطلب
                    </Button>
                  </div>
                </TabsContent>
              )}
            </Tabs>

            <div className="flex justify-end gap-2 mt-6 pt-4 border-t">
              <Button variant="outline" onClick={onClose}>
                إغلاق
              </Button>
              {onDelete && (
                <Button
                  variant="destructive"
                  onClick={() => onDelete(request.id)}
                >
                  حذف الطلب
                </Button>
              )}
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
