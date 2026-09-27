import React, { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { toast } from 'sonner'
import { List, MapPin, Save, Loader2, ArrowRight } from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { Input } from '@/components/ui/input'
import { useCategories, useLocations } from '@/hooks/useMetadata'
import {
  useAuctionRequestByID,
  useCreateAuctionRequestAdmin,
  useUpdateAuctionRequest,
} from '@/hooks/useRequests'

const DESCRIPTION_MIN = 10
const DESCRIPTION_MAX = 5000

interface FormState {
  title_ar: string
  title_fr: string
  title_en: string
  description_ar: string
  description_fr: string
  description_en: string
  category_id: number
  location_id: number
  start_price: string
  min_increment: string
  insurance_amount: string
  // insurance_policy (migration 000048): 'required' | 'not_required'.
  // Defaults to 'required' -- never pre-selects "no insurance" so an admin
  // must take a deliberate action to disable it, matching the backend's DB
  // DEFAULT 'required' and preventing an accidental no-deposit auction.
  insurance_policy: 'required' | 'not_required'
  reserve_price: string
  buy_now_price: string
  start_date: string
  end_date: string
  quantity: number
  status: 'draft' | 'pending'
}

const EMPTY_FORM: FormState = {
  title_ar: '', title_fr: '', title_en: '',
  description_ar: '', description_fr: '', description_en: '',
  category_id: 0,
  location_id: 0,
  start_price: '',
  min_increment: '',
  insurance_amount: '',
  insurance_policy: 'required',
  reserve_price: '',
  buy_now_price: '',
  start_date: '',
  end_date: '',
  quantity: 1,
  status: 'pending',
}

// Convert an ISO date string to the value expected by <input type="datetime-local">
function toLocalInputValue(iso?: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function AuctionRequestFormPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const isEdit = !!id

  const { data: existing, isLoading: isLoadingExisting } = useAuctionRequestByID(id ?? null)
  const { data: categories } = useCategories()
  const { data: locations } = useLocations()
  const createMutation = useCreateAuctionRequestAdmin()
  const updateMutation = useUpdateAuctionRequest()

  const [activeLang, setActiveLang] = useState<'ar' | 'fr' | 'en'>('ar')
  const [form, setForm] = useState<FormState>(EMPTY_FORM)
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [loadedRequestId, setLoadedRequestId] = useState<string | null>(null)

  // Populate the form from the fetched request the first time it becomes available
  // (React-recommended "adjust state during render" pattern instead of a setState-in-effect).
  if (isEdit && existing && loadedRequestId !== existing.id) {
    setLoadedRequestId(existing.id)
    setForm({
      title_ar: existing.title_ar || '',
      title_fr: existing.title_fr || '',
      title_en: existing.title_en || '',
      description_ar: existing.description_ar || '',
      description_fr: existing.description_fr || '',
      description_en: existing.description_en || '',
      category_id: existing.category_id || 0,
      location_id: existing.location_id || 0,
      start_price: existing.start_price?.toString() || '',
      min_increment: existing.min_increment?.toString() || '',
      insurance_amount: existing.insurance_amount?.toString() || '',
      // Missing/unexpected value -> 'required', never 'not_required' (see
      // AuctionRequest.insurance_policy doc comment).
      insurance_policy: existing.insurance_policy === 'not_required' ? 'not_required' : 'required',
      reserve_price: existing.reserve_price?.toString() || '',
      buy_now_price: existing.buy_now_price?.toString() || '',
      start_date: toLocalInputValue(existing.start_date),
      end_date: toLocalInputValue(existing.end_date),
      quantity: existing.quantity || 1,
      status: (existing.status === 'draft' ? 'draft' : 'pending'),
    })
  }

  const validate = (): boolean => {
    const errs: Record<string, string> = {}
    if (!form.title_ar.trim()) errs.title_ar = 'العنوان بالعربية مطلوب'
    if (!form.description_ar.trim()) {
      errs.description_ar = 'الوصف بالعربية مطلوب'
    } else if (form.description_ar.trim().length < DESCRIPTION_MIN) {
      errs.description_ar = `الوصف يجب أن يكون ${DESCRIPTION_MIN} أحرف على الأقل`
    } else if (form.description_ar.length > DESCRIPTION_MAX) {
      errs.description_ar = `الوصف يجب ألا يتجاوز ${DESCRIPTION_MAX} حرف`
    }
    if (!form.category_id) errs.category_id = 'التصنيف مطلوب'
    if (!form.start_price || Number(form.start_price) <= 0) errs.start_price = 'السعر الابتدائي مطلوب'
    if (!form.min_increment || Number(form.min_increment) <= 0) errs.min_increment = 'الحد الأدنى للزيادة مطلوب'
    // insurance_amount is only required (and must be > 0) when the admin has
    // chosen the 'required' policy -- an explicit 'not_required' auction may
    // have no amount at all (migration 000048).
    if (form.insurance_policy === 'required') {
      if (!form.insurance_amount || Number(form.insurance_amount) <= 0) errs.insurance_amount = 'مبلغ التأمين مطلوب ويجب أن يكون أكبر من صفر'
    } else if (form.insurance_amount && Number(form.insurance_amount) < 0) {
      errs.insurance_amount = 'مبلغ التأمين لا يمكن أن يكون سالبًا'
    }
    if (!form.start_date) errs.start_date = 'تاريخ البدء مطلوب'
    if (!form.end_date) errs.end_date = 'تاريخ الانتهاء مطلوب'
    if (form.start_date && form.end_date && new Date(form.end_date) <= new Date(form.start_date)) {
      errs.end_date = 'تاريخ الانتهاء يجب أن يكون بعد تاريخ البدء'
    }
    setErrors(errs)
    return Object.keys(errs).length === 0
  }

  const buildPayload = () => {
    const payload: Record<string, unknown> = {
      title_ar: form.title_ar,
      title_fr: form.title_fr || undefined,
      title_en: form.title_en || undefined,
      description_ar: form.description_ar,
      description_fr: form.description_fr || undefined,
      description_en: form.description_en || undefined,
      category_id: form.category_id,
      location_id: form.location_id || undefined,
      start_price: Number(form.start_price),
      min_increment: Number(form.min_increment),
      // Canonicalized client-side too (defense in depth -- the backend
      // enforces this independently in AdminUpdateAuctionRequest): a
      // 'not_required' request is never sent with a stale positive amount.
      insurance_amount: form.insurance_policy === 'not_required' ? 0 : Number(form.insurance_amount),
      insurance_policy: form.insurance_policy,
      reserve_price: form.reserve_price ? Number(form.reserve_price) : null,
      buy_now_price: form.buy_now_price ? Number(form.buy_now_price) : null,
      start_date: new Date(form.start_date).toISOString(),
      end_date: new Date(form.end_date).toISOString(),
      quantity: form.quantity,
      status: form.status,
    }
    return payload
  }

  const handleSubmit = (statusOverride?: 'draft' | 'pending') => {
    const status = statusOverride ?? form.status
    const nextForm = { ...form, status }
    setForm(nextForm)
    if (!validate()) {
      toast.error('يرجى تصحيح الأخطاء في النموذج')
      return
    }
    const payload = { ...buildPayload(), status }

    if (isEdit && id) {
      updateMutation.mutate({ id, payload }, {
        onSuccess: () => navigate('/requests'),
      })
    } else {
      createMutation.mutate(payload, {
        onSuccess: () => navigate('/requests'),
      })
    }
  }

  const isSaving = createMutation.isPending || updateMutation.isPending
  const descLength = form.description_ar.length

  if (isEdit && isLoadingExisting) {
    return (
      <div className="flex items-center justify-center py-24">
        <Loader2 className="w-8 h-8 animate-spin text-mazad-primary" />
      </div>
    )
  }

  return (
    <div className="animate-fade-in" dir="rtl">
      <PageHeader
        title={isEdit ? 'تعديل طلب المزاد' : 'طلب مزاد جديد'}
        subtitle={isEdit ? 'تعديل بيانات طلب المزاد الحالي' : 'إنشاء طلب مزاد جديد كمسودة أو للمراجعة مباشرة'}
        icon={Save}
        action={{
          label: 'رجوع',
          icon: ArrowRight,
          variant: 'outline',
          onClick: () => navigate('/requests'),
        }}
      />

      <div className="bg-surface-card border border-surface-border rounded-2xl p-6 space-y-6">
        {/* Language tabs */}
        <div className="flex items-center gap-2 bg-surface-base rounded-xl p-1 w-fit">
          {(['ar', 'fr', 'en'] as const).map(lang => (
            <button
              key={lang}
              type="button"
              onClick={() => setActiveLang(lang)}
              className={`px-4 py-2 text-sm font-bold rounded-lg transition-all ${activeLang === lang ? 'bg-mazad-primary text-white shadow-md' : 'text-surface-muted hover:text-white'}`}
            >
              {lang === 'ar' ? 'العربية' : lang === 'fr' ? 'Français' : 'English'}
            </button>
          ))}
        </div>

        {/* Category & Location */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="space-y-2">
            <label className="text-xs text-surface-muted font-bold flex items-center gap-2">
              <List className="w-3 h-3" /> الفئة <span className="text-red-500">*</span>
            </label>
            <select
              value={form.category_id || ''}
              onChange={e => setForm(f => ({ ...f, category_id: parseInt(e.target.value) || 0 }))}
              className="w-full bg-surface-base border border-surface-border rounded-xl p-3 text-sm text-white focus:border-mazad-primary/60 outline-none appearance-none"
            >
              <option value="">اختر الفئة...</option>
              {categories?.filter(c => !c.parent_id).sort((a, b) => a.display_order - b.display_order).map(parent => {
                const children = categories.filter(c => c.parent_id === parent.id)
                return (
                  <React.Fragment key={parent.id}>
                    <option value={parent.id} className="font-bold bg-surface-card">
                      📁 {activeLang === 'ar' ? parent.name_ar : activeLang === 'fr' ? parent.name_fr : (parent.name_en || parent.name_fr)}
                    </option>
                    {children.map(child => (
                      <option key={child.id} value={child.id}>
                        └─ {activeLang === 'ar' ? child.name_ar : activeLang === 'fr' ? child.name_fr : (child.name_en || child.name_fr)}
                      </option>
                    ))}
                  </React.Fragment>
                )
              })}
            </select>
            {errors.category_id && <p className="text-xs text-red-400">{errors.category_id}</p>}
          </div>
          <div className="space-y-2">
            <label className="text-xs text-surface-muted font-bold flex items-center gap-2">
              <MapPin className="w-3 h-3" /> الموقع
            </label>
            <select
              value={form.location_id || ''}
              onChange={e => setForm(f => ({ ...f, location_id: parseInt(e.target.value) || 0 }))}
              className="w-full bg-surface-base border border-surface-border rounded-xl p-3 text-sm text-white focus:border-mazad-primary/60 outline-none appearance-none"
            >
              <option value="">اختر الموقع...</option>
              {locations?.map(l => (
                <option key={l.id} value={l.id}>
                  {activeLang === 'ar' ? l.city_name_ar : (l.city_name_fr || l.city_name_ar)}
                  {activeLang === 'ar'
                    ? (l.area_name_ar ? ` - ${l.area_name_ar}` : '')
                    : (l.area_name_fr ? ` - ${l.area_name_fr}` : (l.area_name_ar ? ` - ${l.area_name_ar}` : ''))}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Title & description per active lang */}
        <div className="space-y-4">
          {activeLang === 'ar' && (
            <>
              <div className="space-y-2">
                <label className="text-xs text-surface-muted font-bold block">العنوان (بالعربية) <span className="text-red-500">*</span></label>
                <Input value={form.title_ar} onChange={e => setForm(f => ({ ...f, title_ar: e.target.value }))} placeholder="مثال: سيارة تويوتا كورولا 2022" />
                {errors.title_ar && <p className="text-xs text-red-400">{errors.title_ar}</p>}
              </div>
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <label className="text-xs text-surface-muted font-bold block">الوصف (بالعربية) <span className="text-red-500">*</span></label>
                  <span className={`text-[10px] ${descLength < DESCRIPTION_MIN || descLength > DESCRIPTION_MAX ? 'text-red-400' : 'text-surface-muted'}`}>
                    {descLength} / {DESCRIPTION_MAX} (حد أدنى {DESCRIPTION_MIN})
                  </span>
                </div>
                <textarea
                  value={form.description_ar}
                  onChange={e => setForm(f => ({ ...f, description_ar: e.target.value }))}
                  rows={5}
                  maxLength={DESCRIPTION_MAX}
                  className="w-full bg-surface-base border border-surface-border rounded-xl p-3 text-sm text-white focus:border-mazad-primary/60 outline-none"
                  placeholder="وصف تفصيلي للمزاد بالعربية..."
                />
                {errors.description_ar && <p className="text-xs text-red-400">{errors.description_ar}</p>}
              </div>
            </>
          )}
          {activeLang === 'fr' && (
            <>
              <div className="space-y-2">
                <label className="text-xs text-surface-muted font-bold block">العنوان (بالفرنسية)</label>
                <Input value={form.title_fr} onChange={e => setForm(f => ({ ...f, title_fr: e.target.value }))} placeholder="Titre en français" />
              </div>
              <div className="space-y-2">
                <label className="text-xs text-surface-muted font-bold block">الوصف (بالفرنسية)</label>
                <textarea
                  value={form.description_fr}
                  onChange={e => setForm(f => ({ ...f, description_fr: e.target.value }))}
                  rows={5}
                  className="w-full bg-surface-base border border-surface-border rounded-xl p-3 text-sm text-white focus:border-mazad-primary/60 outline-none"
                  placeholder="Description en français"
                />
              </div>
            </>
          )}
          {activeLang === 'en' && (
            <>
              <div className="space-y-2">
                <label className="text-xs text-surface-muted font-bold block">العنوان (بالإنجليزية)</label>
                <Input value={form.title_en} onChange={e => setForm(f => ({ ...f, title_en: e.target.value }))} placeholder="Title in English" />
              </div>
              <div className="space-y-2">
                <label className="text-xs text-surface-muted font-bold block">الوصف (بالإنجليزية)</label>
                <textarea
                  value={form.description_en}
                  onChange={e => setForm(f => ({ ...f, description_en: e.target.value }))}
                  rows={5}
                  className="w-full bg-surface-base border border-surface-border rounded-xl p-3 text-sm text-white focus:border-mazad-primary/60 outline-none"
                  placeholder="Description in English"
                />
              </div>
            </>
          )}
        </div>

        {/* Pricing */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="space-y-2">
            <label className="text-xs text-surface-muted font-bold block">السعر الابتدائي <span className="text-red-500">*</span></label>
            <Input type="number" step="0.01" value={form.start_price} onChange={e => setForm(f => ({ ...f, start_price: e.target.value }))} />
            {errors.start_price && <p className="text-xs text-red-400">{errors.start_price}</p>}
          </div>
          <div className="space-y-2">
            <label className="text-xs text-surface-muted font-bold block">الحد الأدنى للزيادة <span className="text-red-500">*</span></label>
            <Input type="number" step="0.01" value={form.min_increment} onChange={e => setForm(f => ({ ...f, min_increment: e.target.value }))} />
            {errors.min_increment && <p className="text-xs text-red-400">{errors.min_increment}</p>}
          </div>
          <div className="space-y-2 md:col-span-3">
            <label className="text-xs text-surface-muted font-bold block">سياسة التأمين</label>
            {/* insurance_policy (migration 000048): explicit toggle, defaults
                to 'required' -- an admin must deliberately choose "بدون تأمين"
                to disable it, never a silent/default state. */}
            <div className="flex items-center gap-2">
              {(['required', 'not_required'] as const).map(p => (
                <button
                  key={p}
                  type="button"
                  onClick={() => setForm(f => ({ ...f, insurance_policy: p, insurance_amount: p === 'not_required' ? '0' : f.insurance_amount }))}
                  className={`px-4 py-2 rounded-xl text-xs font-bold transition-all border ${
                    form.insurance_policy === p
                      ? 'bg-mazad-primary text-white border-mazad-primary shadow-lg shadow-mazad-primary/20'
                      : 'bg-surface-card text-surface-muted border-surface-border hover:text-white'
                  }`}
                >
                  {p === 'required' ? 'يتطلب تأمين' : 'بدون تأمين'}
                </button>
              ))}
            </div>
          </div>
          {form.insurance_policy === 'required' && (
            <div className="space-y-2">
              <label className="text-xs text-surface-muted font-bold block">مبلغ التأمين <span className="text-red-500">*</span></label>
              <Input type="number" step="0.01" value={form.insurance_amount} onChange={e => setForm(f => ({ ...f, insurance_amount: e.target.value }))} />
              {errors.insurance_amount && <p className="text-xs text-red-400">{errors.insurance_amount}</p>}
            </div>
          )}
          <div className="space-y-2">
            <label className="text-xs text-surface-muted font-bold block">سعر الاحتياط (اختياري)</label>
            <Input type="number" step="0.01" value={form.reserve_price} onChange={e => setForm(f => ({ ...f, reserve_price: e.target.value }))} />
          </div>
          <div className="space-y-2">
            <label className="text-xs text-surface-muted font-bold block">سعر الشراء الفوري (اختياري)</label>
            <Input type="number" step="0.01" value={form.buy_now_price} onChange={e => setForm(f => ({ ...f, buy_now_price: e.target.value }))} />
          </div>
          <div className="space-y-2">
            <label className="text-xs text-surface-muted font-bold block">الكمية</label>
            <Input type="number" min="1" value={form.quantity} onChange={e => setForm(f => ({ ...f, quantity: parseInt(e.target.value) || 1 }))} />
          </div>
        </div>

        {/* Dates */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <label className="text-xs text-surface-muted font-bold block">تاريخ البدء <span className="text-red-500">*</span></label>
            <input
              type="datetime-local"
              value={form.start_date}
              onChange={e => setForm(f => ({ ...f, start_date: e.target.value }))}
              className="w-full bg-surface-base border border-surface-border rounded-xl p-3 text-sm text-white focus:border-mazad-primary/60 outline-none"
            />
            {errors.start_date && <p className="text-xs text-red-400">{errors.start_date}</p>}
          </div>
          <div className="space-y-2">
            <label className="text-xs text-surface-muted font-bold block">تاريخ الانتهاء <span className="text-red-500">*</span></label>
            <input
              type="datetime-local"
              value={form.end_date}
              onChange={e => setForm(f => ({ ...f, end_date: e.target.value }))}
              className="w-full bg-surface-base border border-surface-border rounded-xl p-3 text-sm text-white focus:border-mazad-primary/60 outline-none"
            />
            {errors.end_date && <p className="text-xs text-red-400">{errors.end_date}</p>}
          </div>
        </div>

        {/* Status selector */}
        <div className="space-y-2">
          <label className="text-xs text-surface-muted font-bold block">حالة الحفظ</label>
          <div className="flex items-center gap-2">
            {(['draft', 'pending'] as const).map(s => (
              <button
                key={s}
                type="button"
                onClick={() => setForm(f => ({ ...f, status: s }))}
                className={`px-4 py-2 rounded-xl text-xs font-bold transition-all border ${
                  form.status === s
                    ? 'bg-mazad-primary text-white border-mazad-primary shadow-lg shadow-mazad-primary/20'
                    : 'bg-surface-card text-surface-muted border-surface-border hover:text-white'
                }`}
              >
                {s === 'draft' ? 'حفظ كمسودة' : 'إرسال للمراجعة'}
              </button>
            ))}
          </div>
        </div>

        {/* Submit actions */}
        <div className="flex items-center justify-end gap-3 pt-4 border-t border-surface-border">
          <button
            type="button"
            onClick={() => navigate('/requests')}
            className="px-4 py-2 rounded-xl text-xs font-bold border border-surface-border text-surface-muted hover:text-white transition-colors"
          >
            إلغاء
          </button>
          <button
            type="button"
            disabled={isSaving}
            onClick={() => handleSubmit()}
            className="px-5 py-2.5 rounded-xl text-xs font-bold bg-mazad-primary text-white hover:bg-mazad-primary-dk transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {isSaving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
            {isEdit ? 'حفظ التعديلات' : (form.status === 'draft' ? 'حفظ كمسودة' : 'إرسال للمراجعة')}
          </button>
        </div>
      </div>
    </div>
  )
}
