import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Loader2, User, Calendar, DollarSign, Image as ImageIcon, FileText, CheckCircle, XCircle, MapPin } from 'lucide-react'
import { useAuctionRequestByID, useBannerRequestByID } from '@/hooks/useRequests'
import type { AuctionRequest, BannerRequest } from '@/hooks/useRequests'
import { format } from 'date-fns'
import { ar } from 'date-fns/locale'

interface RequestDetailModalProps {
  isOpen: boolean
  onClose: () => void
  type: 'auction' | 'banner'
  requestId: string | null
  onApprove?: (id: string) => void
  onReject?: (id: string) => void
  onDelete?: (id: string) => void
}

export function RequestDetailModal({
  isOpen,
  onClose,
  type,
  requestId,
  onApprove,
  onReject,
  onDelete
}: RequestDetailModalProps) {
  const { data: auctionRequest, isLoading: isLoadingAuction } = useAuctionRequestByID(
    type === 'auction' ? requestId : null
  )
  const { data: bannerRequest, isLoading: isLoadingBanner } = useBannerRequestByID(
    type === 'banner' ? requestId : null
  )

  const isLoading = isLoadingAuction || isLoadingBanner
  const request = type === 'auction' ? auctionRequest : bannerRequest

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'approved':
        return <Badge className="bg-green-100 text-green-800">مقبول</Badge>
      case 'rejected':
        return <Badge className="bg-red-100 text-red-800">مرفوض</Badge>
      default:
        return <Badge className="bg-yellow-100 text-yellow-800">قيد الانتظار</Badge>
    }
  }

  const renderAuctionDetails = (req: AuctionRequest) => (
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

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-blue-50 p-4 rounded-lg">
          <div className="flex items-center gap-2 mb-2">
            <DollarSign className="w-4 h-4 text-blue-600" />
            <p className="text-sm text-gray-600">السعر الابتدائي</p>
          </div>
          <p className="text-lg font-bold text-blue-700">{req.start_price} د.ج</p>
        </div>

        {req.reserve_price && (
          <div className="bg-purple-50 p-4 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <DollarSign className="w-4 h-4 text-purple-600" />
              <p className="text-sm text-gray-600">سعر الاحتياط</p>
            </div>
            <p className="text-lg font-bold text-purple-700">{req.reserve_price} د.ج</p>
          </div>
        )}

        {req.buy_now_price && (
          <div className="bg-green-50 p-4 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <DollarSign className="w-4 h-4 text-green-600" />
              <p className="text-sm text-gray-600">سعر الشراء الفوري</p>
            </div>
            <p className="text-lg font-bold text-green-700">{req.buy_now_price} د.ج</p>
          </div>
        )}

        <div className="bg-orange-50 p-4 rounded-lg">
          <div className="flex items-center gap-2 mb-2">
            <DollarSign className="w-4 h-4 text-orange-600" />
            <p className="text-sm text-gray-600">الحد الأدنى للزيادة</p>
          </div>
          <p className="text-lg font-bold text-orange-700">{req.min_increment} د.ج</p>
        </div>
      </div>

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

      <div className="text-sm text-gray-500 border-t pt-4">
        <p>تاريخ الإنشاء: {format(new Date(req.created_at), 'PPP p', { locale: ar })}</p>
        {req.reviewed_at && (
          <p>تاريخ المراجعة: {format(new Date(req.reviewed_at), 'PPP p', { locale: ar })}</p>
        )}
      </div>
    </div>
  )

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
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="details">التفاصيل</TabsTrigger>
                {request.status === 'pending' && (
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
