import React, { useState } from 'react'
import toast from 'react-hot-toast'
import { 
  Tag, ImageIcon, MinusCircle, Plus, Search,
  Calendar, Eye, Check, X, Loader2, AlertCircle, Save, Pencil, Trash2,
  MapPin, List
} from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { Input } from '@/components/ui/input'
import {
  useAuctions, useValidateAuction, useCreateAuction,
  useUpdateAuction, useDeleteAuction
} from '@/hooks/useAuctions'
import { useCategories, useLocations } from '@/hooks/useMetadata'
import { formatPrice, formatDate, shortID } from '@/lib/formatters'
import type { Auction } from '@/types/api'
import type { AuctionPayload } from '@/api/auctions'
import { fetchAuction } from '@/api/auctions'
import type { ColumnDef } from '@tanstack/react-table'
import { StatusBadge } from '@/components/shared/StatusBadge'
import { DataTable } from '@/components/shared/DataTable'
import { useSearchParams, useNavigate } from 'react-router-dom'

const STATUS_TABS = [
  { label: 'الكل',              value: '' },
  { label: 'في انتظار المراجعة', value: 'pending' },
  { label: 'المزادات النشطة',    value: 'active' },
  { label: 'المزادات المنتهية',   value: 'ended' },
  { label: 'المزادات المرفوضة',   value: 'rejected' },
]

const EMPTY_FORM = {
  title_ar: '', title_fr: '', title_en: '',
  description_ar: '', description_fr: '', description_en: '',
  start_price: 0, buy_now_price: 0, min_increment: 0, insurance_amount: 0,
  category_id: 1, location_id: 1,
  start_time: '', 
  end_time: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().slice(0, 16),
  phone_contact: '',
  images: [''] as string[],
  item_details: {} as Record<string, any>,
}

type FormState = typeof EMPTY_FORM

export function AuctionsPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const [rejectDialog, setRejectDialog] = useState<{ id: string; reason: string } | null>(null)
  const [approveId, setApproveId]       = useState<string | null>(null)
  const [deleteId, setDeleteId]         = useState<string | null>(null)
  const [q, setQ]                       = useState('')

  // Mode: null = list, 'create' = new form, 'edit' = edit form
  const [mode, setMode]           = useState<null | 'create' | 'edit'>(null)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [activeLang, setActiveLang] = useState<'ar'|'fr'|'en'>('ar')
  const [form, setForm]           = useState<FormState>(EMPTY_FORM)
  const [editLoading, setEditLoading] = useState(false)

  const status = searchParams.get('status') ?? ''
  const page   = parseInt(searchParams.get('page') ?? '1')

  const { data, isLoading, isError, refetch } = useAuctions({ status: status || undefined, q, page, per_page: 25 })
  const validate    = useValidateAuction()
  const createMut   = useCreateAuction()
  const updateMut   = useUpdateAuction()
  const deleteMut   = useDeleteAuction()

  const { data: categories } = useCategories()
  const { data: locations }  = useLocations()

  // Set default location to Nouakchott
  React.useEffect(() => {
    if (mode === 'create' && locations && locations.length > 0) {
      const nktt = locations.find(l => 
        l.city_name_ar.includes('نواكشوط') || 
        l.city_name_fr.toLowerCase().includes('nouakchott')
      )
      if (nktt) {
        setForm(f => ({ ...f, location_id: nktt.id }))
      } else if (!form.location_id) {
        setForm(f => ({ ...f, location_id: locations[0].id }))
      }
    }
  }, [mode, locations])

  /* ─── helpers ─── */
  const parseISO = (s: string) => {
    const d = new Date(s)
    if (isNaN(d.getTime())) throw new Error()
    return d.toISOString()
  }

  const buildPayload = (): AuctionPayload | null => {
    if (!form.title_ar.trim()) { toast.error('العنوان بالعربية مطلوب'); return null }
    if (!form.start_price)     { toast.error('السعر الافتتاحي مطلوب'); return null }
    if (!form.end_time)        { toast.error('تاريخ الإغلاق مطلوب'); return null }

    let end_time: string
    try { 
      const d = new Date(form.end_time)
      if (d <= new Date()) {
        toast.error('تاريخ الإغلاق يجب أن يكون في المستقبل')
        return null
      }
      end_time = d.toISOString() 
    }
    catch { toast.error('تاريخ الإغلاق غير صالح'); return null }

    let start_time: string | undefined
    if (form.start_time) {
      try { start_time = parseISO(form.start_time) }
      catch { toast.error('تاريخ البدء غير صالح'); return null }
    }

    if (form.buy_now_price > 0 && form.buy_now_price <= form.start_price) {
      toast.error('سعر البيع المباشر يجب أن يكون أكبر من السعر الافتتاحي')
      return null
    }

    const phone = form.phone_contact?.trim()
    if (phone && !/^[234]\d{7}$/.test(phone)) {
      toast.error('رقم الهاتف غير صحيح (يجب أن يكون 8 أرقام تبدأ بـ 2 أو 3 أو 4)')
      return null
    }

    return {
      category_id:      form.category_id,
      location_id:      form.location_id || undefined,
      title_ar:         form.title_ar.trim(),
      title_fr:         form.title_fr?.trim()  || undefined,
      title_en:         form.title_en?.trim()  || undefined,
      description_ar:   form.description_ar?.trim()  || undefined,
      description_fr:   form.description_fr?.trim()  || undefined,
      description_en:   form.description_en?.trim()  || undefined,
      start_price:      form.start_price,
      min_increment:    form.min_increment || Math.max(form.start_price * 0.05, 100),
      insurance_amount: form.insurance_amount || undefined,
      buy_now_price:    form.buy_now_price   || undefined,
      phone_contact:    phone || undefined,
      start_time,
      end_time,
      images: form.images.filter(img => img && img.trim()),
    }
  }

  const handleCreate = () => {
    const payload = buildPayload()
    if (!payload) return
    createMut.mutate(payload, {
      onSuccess: () => { setMode(null); setForm(EMPTY_FORM); refetch(); }
    })
  }

  const handleUpdate = () => {
    if (!editingId) return
    const payload = buildPayload()
    if (!payload) return
    updateMut.mutate({ id: editingId!, payload }, {
      onSuccess: () => { setMode(null); setEditingId(null); setForm(EMPTY_FORM); refetch(); }
    })
  }

  const openEdit = async (auction: Auction) => {
    setEditingId(auction.id)
    setEditLoading(true)
    try {
      const full = await fetchAuction(auction.id)
      const toDatetimeLocal = (iso: string) => {
        if (!iso) return ''
        return iso.slice(0, 16)
      }
      const imageUrls = (full.images as unknown as string[] | {url: string}[]) ?? []
      const imgs: string[] = imageUrls.map((img: string | {url: string}) =>
        typeof img === 'string' ? img : img.url
      ).filter(Boolean)

      setForm({
        title_ar:       full.title_ar ?? '',
        title_fr:       full.title_fr ?? '',
        title_en:       full.title_en ?? '',
        description_ar: full.description_ar ?? '',
        description_fr: full.description_fr ?? '',
        description_en: full.description_en ?? '',
        start_price:    parseFloat(String(full.start_price))  || 0,
        buy_now_price:  parseFloat(String(full?.buy_now_price ?? 0)) || 0,
        min_increment:  parseFloat(String(full.min_increment)) || 0,
        insurance_amount: parseFloat(String(full.insurance_amount)) || 0,
        category_id:    full.category_id  || 1,
        location_id:    full.location_id  || 1,
        start_time:     toDatetimeLocal(full.start_time as unknown as string),
        end_time:       toDatetimeLocal(full.end_time   as unknown as string),
        phone_contact:  '',
        images:         imgs.length > 0 ? imgs : [''],
        item_details:   full.item_details ?? {},
      })
      setMode('edit')
      setActiveLang('ar')
    } catch (e) {
      console.error('[openEdit] Failed to fetch auction details:', e)
      toast.error('فشل تحميل بيانات المزاد')
      setEditingId(null)
    } finally {
      setEditLoading(false)
    }
  }

  const addImage    = ()          => setForm(f => ({ ...f, images: [...f.images, ''] }))
  const removeImage = (i: number) => setForm(f => {
    const imgs = [...f.images]; imgs.splice(i, 1)
    return { ...f, images: imgs.length ? imgs : [''] }
  })

  /* ─── columns ─── */
  const columns: ColumnDef<Auction>[] = [
    {
      header: 'رقم القطعة',
      accessorKey: 'id',
      cell: ({ row }) => <span className="font-mono text-xs text-surface-muted font-bold">{row.original.lot_number ?? shortID(row.original.id)}</span>
    },
    {
      header: 'العنوان',
      cell: ({ row }) => {
        const a = row.original
        const title = activeLang === 'ar' ? a.title_ar : activeLang === 'fr' ? (a.title_fr || a.title_ar) : (a.title_en || a.title_fr || a.title_ar)
        return <p className="text-white font-bold truncate max-w-[300px]" title={title}>{title}</p>
      }
    },
    {
      header: 'الفئة',
      cell: ({ row }) => {
        const cat = categories?.find(c => c.id === row.original.category_id)
        if (!cat) return <span className="text-surface-muted text-xs">-</span>
        const name = activeLang === 'ar' ? cat.name_ar : activeLang === 'fr' ? cat.name_fr : (cat.name_en || cat.name_fr)
        return <span className="text-xs text-blue-400 font-medium">{name}</span>
      }
    },
    {
      header: 'الموقع',
      cell: ({ row }) => {
        const loc = locations?.find(l => l.id === row.original.location_id)
        if (!loc) return <span className="text-surface-muted text-xs">عام</span>
        const city = activeLang === 'ar' ? loc.city_name_ar : loc.city_name_fr
        const area = activeLang === 'ar' ? loc.area_name_ar : loc.area_name_fr
        return (
          <div className="flex flex-col">
            <span className="text-xs text-white font-medium">{city}</span>
            {area && <span className="text-[10px] text-surface-muted">{area}</span>}
          </div>
        )
      }
    },
    {
      header: 'السعر الحالي',
      accessorKey: 'current_price',
      cell: ({ getValue }) => <span className="font-mono font-bold text-mazad-accent">{formatPrice(parseFloat(getValue<string>()))}</span>
    },
    {
      header: 'تاريخ الانتهاء',
      accessorKey: 'end_time',
      cell: ({ getValue }) => <span className="text-xs text-surface-muted">{formatDate(getValue<string>())}</span>
    },
    {
      header: 'المزايدات',
      accessorKey: 'bidder_count',
      cell: ({ getValue }) => <span className="font-bold text-surface-muted">{getValue<number>()}</span>
    },
    {
      header: 'الحالة',
      accessorKey: 'status',
      cell: ({ getValue }) => <StatusBadge status={getValue<string>()} />
    },
    {
      header: 'الإجراءات',
      id: 'actions',
      cell: ({ row }) => {
        const auction = row.original
        return (
          <div className="flex items-center gap-1.5">
            <button
              onClick={() => navigate(`/auctions/${auction.id}`)}
              className="p-1.5 rounded-lg text-surface-muted hover:text-white hover:bg-surface-border transition-all"
              title="عرض"
            ><Eye className="w-4 h-4" /></button>

            <button
              onClick={() => openEdit(auction)}
              disabled={editLoading && editingId === auction.id}
              className="p-1.5 rounded-lg text-blue-400 hover:bg-blue-500/10 transition-all disabled:opacity-50"
              title="تعديل"
            >
              {editLoading && editingId === auction.id ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Pencil className="w-4 h-4" />
              )}
            </button>

            {auction.status === 'pending' && (<>
              <button
                onClick={() => setApproveId(auction.id)}
                className="p-1.5 rounded-lg text-emerald-400 hover:bg-emerald-500/10 transition-all"
                title="الموافقة"
              ><Check className="w-4 h-4" /></button>
              <button
                onClick={() => setRejectDialog({ id: auction.id, reason: '' })}
                className="p-1.5 rounded-lg text-orange-400 hover:bg-orange-500/10 transition-all"
                title="رفض"
              ><X className="w-4 h-4" /></button>
            </>)}

            <button
              onClick={() => setDeleteId(auction.id)}
              className="p-1.5 rounded-lg text-red-400 hover:bg-red-500/10 transition-all"
              title="حذف"
            ><Trash2 className="w-4 h-4" /></button>
          </div>
        )
      }
    }
  ]

  /* ─── Form (Create / Edit) ─── */
  if (mode === 'create' || mode === 'edit') {
    const isEdit    = mode === 'edit'
    const isPending = isEdit ? updateMut.isPending : createMut.isPending

    return (
      <div className="animate-fade-in" dir="rtl">
        <PageHeader
          title={isEdit ? 'تعديل المزاد' : 'إضافة مزاد جديد'}
          subtitle={isEdit ? 'عدّل تفاصيل المزاد ثم احفظ' : 'أدخل تفاصيل ومواصفات السلعة'}
          action={{ label: 'إلغاء والعودة', icon: X, onClick: () => { setMode(null); setForm(EMPTY_FORM) } }}
        />

        <div className="admin-card p-6 md:p-8 mt-6 space-y-8 animate-slide-up">

          {/* Header & Lang Tabs */}
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-6 border-b border-surface-border">
            <div className="flex items-center gap-3">
              <div className="p-3 bg-mazad-primary/20 text-mazad-primary rounded-xl">
                <Tag className="w-6 h-6" />
              </div>
              <div>
                <h3 className="font-display text-lg font-bold text-white">المعلومات الأساسية</h3>
                <p className="text-sm text-surface-muted">تفاصيل الموصفات والأسعار للسلعة</p>
              </div>
            </div>
            <div className="flex bg-surface-base p-1 rounded-xl border border-surface-border w-fit">
              {(['ar', 'fr', 'en'] as const).map(lang => (
                <button key={lang} onClick={() => setActiveLang(lang)}
                  className={`px-4 py-2 text-sm font-bold rounded-lg transition-all ${activeLang === lang ? 'bg-mazad-primary text-white shadow-md' : 'text-surface-muted hover:text-white'}`}>
                  {lang === 'ar' ? 'العربية' : lang === 'fr' ? 'Français' : 'English'}
                </button>
              ))}
            </div>
          </div>

          {/* Category & Location selection */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-xs text-surface-muted font-bold flex items-center gap-2">
                <List className="w-3 h-3" /> الفئة <span className="text-red-500">*</span>
              </label>
              <select
                value={form.category_id}
                onChange={e => setForm(f => ({ ...f, category_id: parseInt(e.target.value) }))}
                className="w-full bg-surface-base border border-surface-border rounded-xl p-3 text-sm text-white focus:border-mazad-primary/60 outline-none appearance-none"
              >
                <option value="">اختر الفئة...</option>
                {categories?.map(c => (
                  <option key={c.id} value={c.id}>
                    {activeLang === 'ar' ? c.name_ar : activeLang === 'fr' ? c.name_fr : (c.name_en || c.name_fr)}
                  </option>
                ))}
              </select>
            </div>
            <div className="space-y-2">
              <label className="text-xs text-surface-muted font-bold flex items-center gap-2">
                <MapPin className="w-3 h-3" /> الموقع <span className="text-red-500">*</span>
              </label>
              <select
                value={form.location_id}
                onChange={e => setForm(f => ({ ...f, location_id: parseInt(e.target.value) }))}
                className="w-full bg-surface-base border border-surface-border rounded-xl p-3 text-sm text-white focus:border-mazad-primary/60 outline-none appearance-none"
              >
                <option value="">اختر الموقع...</option>
                {locations?.map(l => (
                  <option key={l.id} value={l.id}>
                    {activeLang === 'ar' ? l.city_name_ar : (l.city_name_fr || l.city_name_ar)}
                    {activeLang === 'ar' 
                      ? (l.area_name_ar ? ` - ${l.area_name_ar}` : '')
                      : (l.area_name_fr ? ` - ${l.area_name_fr}` : (l.area_name_ar ? ` - ${l.area_name_ar}` : ''))
                    }
                  </option>
                ))}
              </select>
            </div>
          </div>

          {/* Title & Description (per lang) */}
          <div className="col-span-3 space-y-4">
            {activeLang === 'ar' && (
              <>
                <div className="space-y-2">
                  <label className="text-xs text-surface-muted font-bold block">العنوان (بالعربية) <span className="text-red-500">*</span></label>
                  <Input value={form.title_ar} onChange={e => setForm(f => ({...f, title_ar: e.target.value}))} placeholder="مثال: سيارة تويوتا كورولا 2022" />
                </div>
                <div className="space-y-2">
                  <label className="text-xs text-surface-muted font-bold block">الوصف (بالعربية)</label>
                  <textarea rows={4} value={form.description_ar} onChange={e => setForm(f => ({...f, description_ar: e.target.value}))} placeholder="اكتب تفاصيل السلعة..." className="w-full bg-surface-base border border-surface-border rounded-xl p-4 text-sm text-white focus:border-mazad-primary/60 outline-none resize-none" />
                </div>
              </>
            )}
            {activeLang === 'fr' && (
              <div dir="ltr" className="space-y-4 text-left">
                <div className="space-y-2">
                  <label className="text-xs text-surface-muted font-bold block">Titre (en Français)</label>
                  <Input value={form.title_fr} onChange={e => setForm(f => ({...f, title_fr: e.target.value}))} placeholder="Ex: Toyota Corolla 2022" />
                </div>
                <div className="space-y-2">
                  <label className="text-xs text-surface-muted font-bold block">Description (en Français)</label>
                  <textarea rows={4} value={form.description_fr} onChange={e => setForm(f => ({...f, description_fr: e.target.value}))} placeholder="Détails de l'article..." className="w-full bg-surface-base border border-surface-border rounded-xl p-4 text-sm text-white focus:border-mazad-primary/60 outline-none resize-none" />
                </div>
              </div>
            )}
            {activeLang === 'en' && (
              <div dir="ltr" className="space-y-4 text-left">
                <div className="space-y-2">
                  <label className="text-xs text-surface-muted font-bold block">Title (in English)</label>
                  <Input value={form.title_en} onChange={e => setForm(f => ({...f, title_en: e.target.value}))} placeholder="Ex: Toyota Corolla 2022" />
                </div>
                <div className="space-y-2">
                  <label className="text-xs text-surface-muted font-bold block">Description (in English)</label>
                  <textarea rows={4} value={form.description_en} onChange={e => setForm(f => ({...f, description_en: e.target.value}))} placeholder="Item details..." className="w-full bg-surface-base border border-surface-border rounded-xl p-4 text-sm text-white focus:border-mazad-primary/60 outline-none resize-none" />
                </div>
              </div>
            )}
          </div>

          {/* Prices Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 pt-4 border-t border-surface-border">
            <div className="space-y-2">
              <label className="text-xs text-surface-muted font-bold block">السعر الافتتاحي (أوقية) <span className="text-red-500">*</span></label>
              <Input type="number" value={form.start_price || ''} onChange={e => setForm(f => ({...f, start_price: parseFloat(e.target.value) || 0}))} placeholder="0.00" />
            </div>
            <div className="space-y-2">
              <label className="text-xs text-surface-muted font-bold block">سعر الشراء المباشر (اختياري)</label>
              <Input type="number" value={form.buy_now_price || ''} onChange={e => setForm(f => ({...f, buy_now_price: parseFloat(e.target.value) || 0}))} placeholder="0.00" />
            </div>
            <div className="space-y-2">
              <label className="text-xs text-surface-muted font-bold block">الحد الأدنى للمزايدة (تلقائي إن تُرِك 0)</label>
              <Input type="number" value={form.min_increment || ''} onChange={e => setForm(f => ({...f, min_increment: parseFloat(e.target.value) || 0}))} placeholder="0.00" />
            </div>
            <div className="space-y-2">
              <label className="text-xs text-surface-muted font-bold block">مبلغ التأمين (اختياري)</label>
              <Input type="number" value={form.insurance_amount || ''} onChange={e => setForm(f => ({...f, insurance_amount: parseFloat(e.target.value) || 0}))} placeholder="0.00" />
            </div>
          </div>

          {/* Dates */}
          <div className="pt-4 border-t border-surface-border">
            <div className="flex items-center gap-3 mb-4">
              <div className="p-2 bg-purple-500/20 text-purple-400 rounded-lg"><Calendar className="w-5 h-5" /></div>
              <h4 className="font-bold text-white">التوقيت</h4>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-2">
                <label className="text-xs text-surface-muted font-bold block">وقت وتاريخ البدء</label>
                <Input type="datetime-local" value={form.start_time} onChange={e => setForm(f => ({...f, start_time: e.target.value}))} />
              </div>
              <div className="space-y-2">
                <label className="text-xs text-surface-muted font-bold block">وقت وتاريخ الإغلاق <span className="text-red-500">*</span></label>
                <Input type="datetime-local" value={form.end_time} onChange={e => setForm(f => ({...f, end_time: e.target.value}))} />
              </div>
            </div>
          </div>

          {/* Images */}
          <div className="pt-4 border-t border-surface-border">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-blue-500/20 text-blue-400 rounded-lg"><ImageIcon className="w-5 h-5" /></div>
                <h4 className="font-bold text-white">صور المزاد</h4>
              </div>
              <button onClick={addImage} className="flex items-center gap-2 text-xs font-bold text-blue-400 border border-blue-500/30 px-3 py-1.5 rounded-lg hover:bg-blue-500/10 transition-colors">
                <Plus className="w-4 h-4" /> إضافة رابط صورة
              </button>
            </div>
            <div className="grid gap-3 max-w-2xl">
              {form.images.map((url, i) => (
                <div key={i} className="flex items-center gap-2">
                  <Input placeholder="https://..." value={url} dir="ltr" className="ltr-input"
                    onChange={e => {
                      const imgs = [...form.images]; imgs[i] = e.target.value
                      setForm(f => ({ ...f, images: imgs }))
                    }} />
                  <button onClick={() => removeImage(i)} className="p-2 text-red-400 hover:bg-red-500/20 rounded-lg transition-colors">
                    <MinusCircle className="w-5 h-5" />
                  </button>
                </div>
              ))}
            </div>
          </div>

          {/* Actions */}
          <div className="pt-6 border-t border-surface-border flex justify-end gap-3">
            <button onClick={() => { setMode(null); setForm(EMPTY_FORM) }} className="px-6 py-3 rounded-xl text-sm font-bold text-surface-muted border border-surface-border hover:bg-surface-border/50 transition-all">إلغاء</button>
            <button
              disabled={isPending || !form.title_ar || !form.start_price || !form.end_time}
              onClick={isEdit ? handleUpdate : handleCreate}
              className="px-8 py-3 rounded-xl text-sm font-bold bg-mazad-primary hover:bg-mazad-primary-dark text-white shadow-lg shadow-mazad-primary/20 disabled:opacity-50 flex items-center gap-2 transition-all"
            >
              {isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
              {isEdit ? 'حفظ التعديلات' : 'حفظ وبث المزاد'}
            </button>
          </div>
        </div>
      </div>
    )
  }

  /* ─── Error State ─── */
  if (isError) return (
    <div className="admin-card p-20 text-center flex flex-col items-center gap-4">
      <AlertCircle className="w-12 h-12 text-red-500/20" />
      <p className="text-red-400 font-bold">فشل تحميل المزادات</p>
      <button onClick={() => window.location.reload()} className="bg-surface-border text-white px-6 py-2 rounded-xl text-sm font-bold hover:bg-surface-border/80 transition-all">إعادة المحاولة</button>
    </div>
  )

  /* ─── List ─── */
  return (
    <div className="animate-fade-in" dir="rtl">
      <PageHeader
        title="المزادات"
        subtitle={`${data?.total ?? 0} مزاد في المجمل`}
        action={{ label: 'إضافة مزاد جديد', icon: Plus, onClick: () => { setForm(EMPTY_FORM); setMode('create') } }}
      />

      {/* Tabs */}
      <div className="flex gap-1 mb-6 bg-surface-card border border-surface-border rounded-xl p-1 w-fit flex-wrap">
        {STATUS_TABS.map(tab => (
          <button key={tab.value} onClick={() => setSearchParams({ status: tab.value, page: '1' })}
            className={`px-4 py-2 rounded-lg text-xs font-bold transition-all ${status === tab.value ? 'bg-mazad-primary text-white shadow-lg' : 'text-surface-muted hover:text-white hover:bg-surface-border/50'}`}>
            {tab.label}
          </button>
        ))}
      </div>

      {/* Search */}
      <div className="relative mb-6 max-w-md group">
        <Search className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted group-focus-within:text-mazad-primary transition-colors" />
        <Input value={q} onChange={e => { setQ(e.target.value); setSearchParams(p => { p.set('page','1'); return p }) }}
          placeholder="ابحث عن مزاد..." className="pr-10" />
      </div>

      <DataTable
        columns={columns}
        data={data?.data ?? []}
        isLoading={isLoading}
        total={data?.total}
        page={page}
        onPageChange={p => setSearchParams(prev => { prev.set('page', p.toString()); return prev })}
        emptyTitle="لا توجد مزادات"
        emptyDescription="لم يتم العثور على أي مزادات تطابق الفلتر الحالي."
      />

      {/* Approve Dialog */}
      <ConfirmDialog
        open={!!approveId}
        onOpenChange={v => !v && setApproveId(null)}
        title="الموافقة على بث هذا المزاد؟"
        description="سيتم نشر المزاد فوراً وسيتمكن الجمهور من المزايدة عليه."
        confirmLabel="موافقة"
        variant="success"
        loading={validate.isPending}
        onConfirm={() => {
          if (approveId) validate.mutate({ id: approveId, approve: true }, { onSuccess: () => setApproveId(null) })
        }}
      />

      {/* Delete Dialog */}
      <ConfirmDialog
        open={!!deleteId}
        onOpenChange={v => !v && setDeleteId(null)}
        title="حذف هذا المزاد؟"
        description="سيتم حذف المزاد نهائياً ولا يمكن التراجع عن هذا الإجراء."
        confirmLabel="حذف"
        variant="danger"
        loading={deleteMut.isPending}
        onConfirm={() => {
          if (deleteId) deleteMut.mutate(deleteId, { onSuccess: () => setDeleteId(null) })
        }}
      />

      {/* Reject Dialog */}
      {rejectDialog && (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4 backdrop-blur-sm">
          <div className="admin-card p-8 w-full max-w-md animate-slide-in relative overflow-hidden">
            <div className="absolute top-0 right-0 w-24 h-24 bg-red-500/5 rounded-full -mr-12 -mt-12 blur-3xl" />
            <h3 className="font-display font-bold text-white text-xl mb-1 relative">رفض هذا المزاد</h3>
            <p className="text-sm text-surface-muted mb-6 font-medium relative italic">سيتم إرسال سبب الرفض إلى البائع.</p>
            <textarea
              value={rejectDialog.reason}
              onChange={e => setRejectDialog({ ...rejectDialog, reason: e.target.value })}
              placeholder="اكتب سبب الرفض هنا (إلزامي)..."
              rows={4}
              className="w-full bg-surface-base border border-surface-border rounded-xl p-4 text-sm text-white placeholder:text-surface-muted/30 focus:outline-none focus:border-red-500/60 transition-all resize-none mb-6 shadow-inner"
            />
            <div className="flex gap-3 justify-end items-center">
              <button onClick={() => setRejectDialog(null)} className="px-5 py-2.5 rounded-xl text-sm font-bold text-surface-muted border border-surface-border hover:bg-surface-border/50 transition-all">إلغاء</button>
              <button
                disabled={!rejectDialog.reason.trim() || validate.isPending}
                onClick={() => validate.mutate(
                  { id: rejectDialog.id, approve: false, reason: rejectDialog.reason },
                  { onSuccess: () => setRejectDialog(null) }
                )}
                className="px-6 py-2.5 rounded-xl text-sm font-bold bg-red-500 hover:bg-red-600 text-white disabled:opacity-40 transition-all flex items-center gap-2"
              >
                {validate.isPending && <Loader2 className="w-4 h-4 animate-spin" />}
                {validate.isPending ? 'جاري التنفيذ...' : 'تأكيد الرفض'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
