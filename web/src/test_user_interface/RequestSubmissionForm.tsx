import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import client from '@/api/client'
import { toast } from 'sonner'
import { Loader2, Upload, X, Image as ImageIcon } from 'lucide-react'

const auctionRequestSchema = z.object({
  title_ar: z.string().min(3, 'العنوان بالعربية مطلوب'),
  title_fr: z.string().optional(),
  title_en: z.string().optional(),
  description_ar: z.string().optional(),
  description_fr: z.string().optional(),
  description_en: z.string().optional(),
  category_id: z.number().min(1, 'التصنيف مطلوب'),
  location_id: z.number().optional(),
  start_price: z.number().min(0, 'السعر الابتدائي يجب أن يكون أكبر من 0'),
  reserve_price: z.number().optional(),
  min_increment: z.number().min(0, 'الحد الأدنى للزيادة مطلوب'),
  insurance_amount: z.number().min(0, 'مبلغ التأمين مطلوب'),
  buy_now_price: z.number().optional(),
  start_date: z.string().min(1, 'تاريخ البدء مطلوب'),
  end_date: z.string().min(1, 'تاريخ الانتهاء مطلوب'),
}).refine((data) => {
  if (data.start_date && data.end_date) {
    return new Date(data.end_date) > new Date(data.start_date)
  }
  return true
}, {
  message: 'تاريخ الانتهاء يجب أن يكون بعد تاريخ البدء',
  path: ['end_date']
}).refine((data) => {
  if (data.reserve_price && data.start_price) {
    return data.reserve_price >= data.start_price
  }
  return true
}, {
  message: 'سعر الاحتياط يجب أن يكون أكبر أو يساوي السعر الابتدائي',
  path: ['reserve_price']
})

const bannerRequestSchema = z.object({
  title_ar: z.string().min(3, 'العنوان بالعربية مطلوب'),
  title_fr: z.string().optional(),
  title_en: z.string().optional(),
  image_url: z.string().url('رابط الصورة مطلوب ويجب أن يكون URL صالح'),
  target_url: z.string().url('يجب أن يكون URL صالح').optional(),
  starts_at: z.string().min(1, 'تاريخ البدء مطلوب'),
  ends_at: z.string().min(1, 'تاريخ الانتهاء مطلوب'),
}).refine((data) => {
  if (data.starts_at && data.ends_at) {
    return new Date(data.ends_at) > new Date(data.starts_at)
  }
  return true
}, {
  message: 'تاريخ الانتهاء يجب أن يكون بعد تاريخ البدء',
  path: ['ends_at']
})

type AuctionRequestForm = z.infer<typeof auctionRequestSchema>
type BannerRequestForm = z.infer<typeof bannerRequestSchema>

interface RequestSubmissionFormProps {
  type: 'auction' | 'banner'
  onSuccess?: () => void
}

export function RequestSubmissionForm({ type, onSuccess }: RequestSubmissionFormProps) {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [images, setImages] = useState<File[]>([])
  const [imagePreviews, setImagePreviews] = useState<string[]>([])

  const auctionForm = useForm<AuctionRequestForm>({
    resolver: zodResolver(auctionRequestSchema),
    defaultValues: {
      start_price: 0,
      min_increment: 100,
      insurance_amount: 0,
    }
  })

  const bannerForm = useForm<BannerRequestForm>({
    resolver: zodResolver(bannerRequestSchema),
  })

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || [])
    if (files.length + images.length > 10) {
      toast.error('يمكنك رفع 10 صور كحد أقصى')
      return
    }

    const newImages = [...images, ...files]
    setImages(newImages)

    const newPreviews = files.map(file => URL.createObjectURL(file))
    setImagePreviews([...imagePreviews, ...newPreviews])
  }

  const removeImage = (index: number) => {
    const newImages = images.filter((_, i) => i !== index)
    const newPreviews = imagePreviews.filter((_, i) => i !== index)
    setImages(newImages)
    setImagePreviews(newPreviews)
  }

  const uploadImages = async (): Promise<string[]> => {
    if (images.length === 0) return []

    const formData = new FormData()
    images.forEach(image => {
      formData.append('images', image)
    })

    const response = await client.post('/v1/api/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })

    return response.data.urls || []
  }

  const onSubmitAuction = async (data: AuctionRequestForm) => {
    setIsSubmitting(true)
    try {
      const imageUrls = await uploadImages()

      await client.post('/v1/api/requests/auctions', {
        ...data,
        images: imageUrls,
        start_price: data.start_price.toString(),
        reserve_price: data.reserve_price?.toString(),
        min_increment: data.min_increment.toString(),
        insurance_amount: data.insurance_amount.toString(),
        buy_now_price: data.buy_now_price?.toString(),
      })

      toast.success('تم إرسال طلب المزاد بنجاح')
      auctionForm.reset()
      setImages([])
      setImagePreviews([])
      onSuccess?.()
    } catch (error) {
      toast.error('فشل إرسال طلب المزاد')
    } finally {
      setIsSubmitting(false)
    }
  }

  const onSubmitBanner = async (data: BannerRequestForm) => {
    setIsSubmitting(true)
    try {
      await client.post('/v1/api/requests/banners', data)
      toast.success('تم إرسال طلب البانر بنجاح')
      bannerForm.reset()
      onSuccess?.()
    } catch (error) {
      toast.error('فشل إرسال طلب البانر')
    } finally {
      setIsSubmitting(false)
    }
  }

  if (type === 'auction') {
    return (
      <form onSubmit={auctionForm.handleSubmit(onSubmitAuction)} className="space-y-6 max-w-4xl mx-auto p-6">
        <h2 className="text-2xl font-bold mb-6">طلب إضافة مزاد</h2>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium mb-2">العنوان (عربي) *</label>
            <input
              {...auctionForm.register('title_ar')}
              className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
              placeholder="عنوان المزاد بالعربية"
            />
            {auctionForm.formState.errors.title_ar && (
              <p className="text-red-500 text-sm mt-1">{auctionForm.formState.errors.title_ar.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">العنوان (فرنسي)</label>
            <input
              {...auctionForm.register('title_fr')}
              className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
              placeholder="Titre en français"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">العنوان (إنجليزي)</label>
            <input
              {...auctionForm.register('title_en')}
              className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
              placeholder="Title in English"
            />
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium mb-2">التصنيف *</label>
            <input
              type="number"
              {...auctionForm.register('category_id', { valueAsNumber: true })}
              className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
              placeholder="رقم التصنيف"
            />
            {auctionForm.formState.errors.category_id && (
              <p className="text-red-500 text-sm mt-1">{auctionForm.formState.errors.category_id.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">الموقع</label>
            <input
              type="number"
              {...auctionForm.register('location_id', { valueAsNumber: true })}
              className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
              placeholder="رقم الموقع"
            />
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium mb-2">السعر الابتدائي *</label>
            <input
              type="number"
              step="0.01"
              {...auctionForm.register('start_price', { valueAsNumber: true })}
              className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
            />
            {auctionForm.formState.errors.start_price && (
              <p className="text-red-500 text-sm mt-1">{auctionForm.formState.errors.start_price.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">سعر الاحتياط</label>
            <input
              type="number"
              step="0.01"
              {...auctionForm.register('reserve_price', { valueAsNumber: true })}
              className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
            />
            {auctionForm.formState.errors.reserve_price && (
              <p className="text-red-500 text-sm mt-1">{auctionForm.formState.errors.reserve_price.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">سعر الشراء الفوري</label>
            <input
              type="number"
              step="0.01"
              {...auctionForm.register('buy_now_price', { valueAsNumber: true })}
              className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium mb-2">الحد الأدنى للزيادة *</label>
            <input
              type="number"
              step="0.01"
              {...auctionForm.register('min_increment', { valueAsNumber: true })}
              className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">مبلغ التأمين *</label>
            <input
              type="number"
              step="0.01"
              {...auctionForm.register('insurance_amount', { valueAsNumber: true })}
              className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium mb-2">تاريخ البدء *</label>
            <input
              type="datetime-local"
              {...auctionForm.register('start_date')}
              className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
            />
            {auctionForm.formState.errors.start_date && (
              <p className="text-red-500 text-sm mt-1">{auctionForm.formState.errors.start_date.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">تاريخ الانتهاء *</label>
            <input
              type="datetime-local"
              {...auctionForm.register('end_date')}
              className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
            />
            {auctionForm.formState.errors.end_date && (
              <p className="text-red-500 text-sm mt-1">{auctionForm.formState.errors.end_date.message}</p>
            )}
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium mb-2">الوصف (عربي)</label>
          <textarea
            {...auctionForm.register('description_ar')}
            rows={3}
            className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
            placeholder="وصف المزاد بالعربية"
          />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium mb-2">الوصف (فرنسي)</label>
            <textarea
              {...auctionForm.register('description_fr')}
              rows={3}
              className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
              placeholder="Description en français"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">الوصف (إنجليزي)</label>
            <textarea
              {...auctionForm.register('description_en')}
              rows={3}
              className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
              placeholder="Description in English"
            />
          </div>
        </div>

        {/* Images Upload */}
        <div>
          <label className="block text-sm font-medium mb-2">صور المزاد (حتى 10 صور)</label>
          <div className="border-2 border-dashed border-gray-300 rounded-lg p-6 text-center hover:border-blue-500 transition-colors">
            <input
              type="file"
              multiple
              accept="image/*"
              onChange={handleImageChange}
              className="hidden"
              id="auction-images"
            />
            <label htmlFor="auction-images" className="cursor-pointer flex flex-col items-center">
              <ImageIcon className="w-12 h-12 text-gray-400 mb-2" />
              <span className="text-gray-600">اضغط لرفع الصور أو اسحبها هنا</span>
            </label>
          </div>

          {imagePreviews.length > 0 && (
            <div className="grid grid-cols-5 gap-2 mt-4">
              {imagePreviews.map((preview, index) => (
                <div key={index} className="relative">
                  <img
                    src={preview}
                    alt={`Preview ${index + 1}`}
                    className="w-full h-24 object-cover rounded-lg"
                  />
                  <button
                    type="button"
                    onClick={() => removeImage(index)}
                    className="absolute -top-2 -right-2 bg-red-500 text-white rounded-full p-1 w-6 h-6 flex items-center justify-center hover:bg-red-600"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        <button
          type="submit"
          disabled={isSubmitting}
          className="w-full bg-blue-600 text-white py-3 rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
        >
          {isSubmitting ? (
            <>
              <Loader2 className="w-5 h-5 animate-spin" />
              جاري الإرسال...
            </>
          ) : (
            'إرسال طلب المزاد'
          )}
        </button>
      </form>
    )
  }

  // Banner Form
  return (
    <form onSubmit={bannerForm.handleSubmit(onSubmitBanner)} className="space-y-6 max-w-2xl mx-auto p-6">
      <h2 className="text-2xl font-bold mb-6">طلب إضافة بانر إعلاني</h2>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div>
          <label className="block text-sm font-medium mb-2">العنوان (عربي) *</label>
          <input
            {...bannerForm.register('title_ar')}
            className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
            placeholder="عنوان البانر بالعربية"
          />
          {bannerForm.formState.errors.title_ar && (
            <p className="text-red-500 text-sm mt-1">{bannerForm.formState.errors.title_ar.message}</p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium mb-2">العنوان (فرنسي)</label>
          <input
            {...bannerForm.register('title_fr')}
            className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
            placeholder="Titre en français"
          />
        </div>

        <div>
          <label className="block text-sm font-medium mb-2">العنوان (إنجليزي)</label>
          <input
            {...bannerForm.register('title_en')}
            className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
            placeholder="Title in English"
          />
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium mb-2">رابط الصورة *</label>
        <input
          {...bannerForm.register('image_url')}
          className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
          placeholder="https://example.com/image.jpg"
        />
        {bannerForm.formState.errors.image_url && (
          <p className="text-red-500 text-sm mt-1">{bannerForm.formState.errors.image_url.message}</p>
        )}
      </div>

      <div>
        <label className="block text-sm font-medium mb-2">رابط الوجهة</label>
        <input
          {...bannerForm.register('target_url')}
          className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
          placeholder="https://example.com"
        />
        {bannerForm.formState.errors.target_url && (
          <p className="text-red-500 text-sm mt-1">{bannerForm.formState.errors.target_url.message}</p>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium mb-2">تاريخ البدء *</label>
          <input
            type="datetime-local"
            {...bannerForm.register('starts_at')}
            className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
          />
          {bannerForm.formState.errors.starts_at && (
            <p className="text-red-500 text-sm mt-1">{bannerForm.formState.errors.starts_at.message}</p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium mb-2">تاريخ الانتهاء *</label>
          <input
            type="datetime-local"
            {...bannerForm.register('ends_at')}
            className="w-full p-3 border rounded-lg focus:ring-2 focus:ring-blue-500"
          />
          {bannerForm.formState.errors.ends_at && (
            <p className="text-red-500 text-sm mt-1">{bannerForm.formState.errors.ends_at.message}</p>
          )}
        </div>
      </div>

      <button
        type="submit"
        disabled={isSubmitting}
        className="w-full bg-green-600 text-white py-3 rounded-lg font-medium hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
      >
        {isSubmitting ? (
          <>
            <Loader2 className="w-5 h-5 animate-spin" />
            جاري الإرسال...
          </>
        ) : (
          'إرسال طلب البانر'
        )}
      </button>
    </form>
  )
}
