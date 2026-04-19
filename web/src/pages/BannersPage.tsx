import { useState } from 'react'
import { Plus, Trash2, ToggleLeft, ToggleRight, Image as ImageIcon, Loader2, AlertCircle, Edit2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { PageHeader } from '@/components/shared/PageHeader'
import { EmptyState } from '@/components/shared/EmptyState'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { Input } from '@/components/ui/input'
import { formatDate } from '@/lib/formatters'
import client from '@/api/client'

interface Banner {
  id: number
  title_fr: string | null
  title_ar: string | null
  title_en: string | null
  image_url: string
  target_url: string | null
  is_active: boolean
  display_order: number
  starts_at: string | null
  ends_at: string | null
}

function useBanners() {
  return useQuery({
    queryKey: ['banners'],
    queryFn: async () => {
      const { data } = await client.get('/v1/api/banners')
      return data.data as Banner[]
    },
  })
}

export function BannersPage() {
  const qc = useQueryClient()
  const { data: banners = [], isLoading, isError } = useBanners()
  const [showForm, setShowForm] = useState(false)
  const [editingBanner, setEditingBanner] = useState<Banner | null>(null)
  const [newBanner, setNewBanner] = useState({
    title_fr: '',
    title_ar: '',
    title_en: '',
    image_url: '',
    target_url: '',
    display_order: 1
  })
  const [deleteId, setDeleteId] = useState<number | null>(null)

  const createMutation = useMutation({
    mutationFn: (b: typeof newBanner) => client.post('/v1/api/banners', b),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['banners'] })
      qc.invalidateQueries({ queryKey: ['dashboard'] })
      setShowForm(false)
      setNewBanner({ title_fr: '', title_ar: '', title_en: '', image_url: '', target_url: '', display_order: 1 })
      toast.success('تم إنشاء الإعلان بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })

  const toggleActive = useMutation({
    mutationFn: ({ id, active }: { id: number; active: boolean }) =>
      client.put(`/v1/api/admin/banners/${id}/toggle`, { is_active: active }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['banners'] })
      toast.success('تم تحديث حالة الإعلان')
    },
    onError: (err: Error) => toast.error(err.message),
  })

  const updateMutation = useMutation({
    mutationFn: (b: Banner) => client.put(`/v1/api/admin/banners/${b.id}`, b),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['banners'] })
      toast.success('تم تحديث الإعلان')
      setEditingBanner(null)
    },
    onError: (err: Error) => toast.error(err.message),
  })

  const deleteBanner = useMutation({
    mutationFn: (id: number) => client.delete(`/v1/api/admin/banners/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['banners'] })
      toast.success('تم حذف الإعلان بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })

  if (isError) return (
    <div className="admin-card p-20 text-center flex flex-col items-center gap-4">
      <AlertCircle className="w-12 h-12 text-red-500/20" />
      <p className="text-red-400 font-bold">فشل تحميل الإعلانات</p>
      <button onClick={() => window.location.reload()} className="bg-surface-border text-white px-6 py-2 rounded-xl text-sm font-bold">إعادة المحاولة</button>
    </div>
  )

  return (
    <div className="animate-fade-in" dir="rtl">
      <PageHeader title="إدارة الإعلانات " subtitle="إدارة الصور الترويجية التي تظهر في الصفحة الرئيسية للتطبيق">
        <button
           onClick={() => setShowForm(!showForm)}
           className="flex items-center gap-2 px-6 py-2.5 rounded-xl bg-mazad-primary text-white text-sm font-bold shadow-lg shadow-mazad-primary/20 hover:bg-mazad-primary-dk transition-all"
        >
          <Plus className="w-4 h-4" />
          {showForm ? 'إلغاء' : 'إضافة إعلان جديد'}
        </button>
      </PageHeader>

      {/* Banner Creation Form */}
      {showForm && (
        <div className="admin-card p-6 mb-8 border-mazad-primary/30 animate-slide-in">
          <h3 className="font-display font-bold text-white text-lg mb-6">إضافة إعلان ترويجي جديد</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
            <div>
              <label className="text-xs font-bold text-surface-muted uppercase tracking-widest block mb-2">العنوان (بالعربية)</label>
              <Input
                value={newBanner.title_ar}
                onChange={(e) => setNewBanner({ ...newBanner, title_ar: e.target.value })}
                placeholder="مثال: مزاد الساعات الفاخرة..."
              />
            </div>
            <div>
              <label className="text-xs font-bold text-surface-muted uppercase tracking-widest block mb-2">العنوان (بالفرنسية)</label>
              <Input
                value={newBanner.title_fr}
                onChange={(e) => setNewBanner({ ...newBanner, title_fr: e.target.value })}
                placeholder="Ex: Vente de voitures..."
                dir="ltr"
                className="text-left"
              />
            </div>
            <div>
              <label className="text-xs font-bold text-surface-muted uppercase tracking-widest block mb-2">Title (English)</label>
              <Input
                value={newBanner.title_en}
                onChange={(e) => setNewBanner({ ...newBanner, title_en: e.target.value })}
                placeholder="Ex: Luxury Auction"
                dir="ltr"
                className="text-left"
              />
            </div>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
            <div>
              <label className="text-xs font-bold text-surface-muted uppercase tracking-widest block mb-2">رابط الصورة (URL) <span className="text-red-400">*</span></label>
              <Input
                value={newBanner.image_url}
                onChange={(e) => setNewBanner({ ...newBanner, image_url: e.target.value })}
                placeholder="https://example.com/image.jpg"
                dir="ltr"
                className="text-left"
              />
            </div>
            <div>
              <label className="text-xs font-bold text-surface-muted uppercase tracking-widest block mb-2">رابط الويب (اختياري)</label>
              <Input
                value={newBanner.target_url}
                onChange={(e) => setNewBanner({ ...newBanner, target_url: e.target.value })}
                placeholder="https://example.com/..."
                dir="ltr"
                className="text-left"
              />
            </div>
          </div>
          <div className="mb-6">
            <label className="text-xs font-bold text-surface-muted uppercase tracking-widest block mb-2">ترتيب العرض</label>
            <Input
              type="number"
              value={newBanner.display_order}
              onChange={(e) => setNewBanner({ ...newBanner, display_order: parseInt(e.target.value) || 1 })}
              placeholder="1"
              className="w-32"
            />
          </div>
          <div className="flex gap-3 justify-end">
            <button
              onClick={() => setShowForm(false)}
              className="px-6 py-2.5 rounded-xl text-sm font-bold text-surface-muted border border-surface-border hover:bg-surface-border/50 transition-all font-bold"
            >
              إلغاء
            </button>
            <button
              disabled={!newBanner.image_url || createMutation.isPending}
              onClick={() => createMutation.mutate(newBanner)}
              className="px-8 py-2.5 rounded-xl text-sm font-bold bg-mazad-primary hover:bg-mazad-primary-dk
                         text-white disabled:opacity-40 transition-all shadow-lg shadow-mazad-primary/20 flex items-center gap-2"
            >
              {createMutation.isPending && <Loader2 className="w-4 h-4 animate-spin" />}
              {createMutation.isPending ? 'جاري الحفظ...' : 'إنشاء الإعلان'}
            </button>
          </div>
        </div>
      )}

      {isLoading ? (
        <div className="admin-card p-20 text-center">
           <LoadingSpinner label="جاري تحميل الإعلانات..." />
        </div>
      ) : !banners || banners.length === 0 ? (
        <div className="admin-card">
          <EmptyState
            icon={ImageIcon}
            title="لا توجد إعلانات"
            description="ابدأ بإضافة أول إعلان ترويجي للمستخدمين."
          />
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {banners.map((banner) => (
            <div key={banner.id} className="admin-card overflow-hidden group">
              <div className="relative aspect-[21/9] bg-surface-base">
                <img
                  src={banner.image_url}
                  alt={banner.title_ar ?? 'إعلان'}
                  className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
                />
                <div className="absolute top-4 right-4">
                  <span className={`px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider backdrop-blur-md border shadow-2xl ${
                      banner.is_active 
                        ? 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30' 
                        : 'bg-red-500/20 text-red-400 border-red-500/30'
                  }`}>
                      {banner.is_active ? 'نشط' : 'متوقف'}
                  </span>
                </div>
              </div>
              
              <div className="p-5">
                <div className="flex items-start justify-between gap-4 mb-4">
                  <div>
                    <h3 className="font-bold text-white text-base mb-1">{banner.title_ar ?? 'بدون عنوان عربي'}</h3>
                    <p className="text-xs text-surface-muted font-medium italic">{banner.title_fr ?? 'Pas de titre FR'}</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => setEditingBanner(banner)}
                      className="p-2 rounded-xl text-mazad-primary border border-transparent hover:border-mazad-primary/20 hover:bg-mazad-primary/10 transition-all"
                      title="تعديل"
                    >
                      <Edit2 className="w-5 h-5" />
                    </button>
                    <button
                      onClick={() => toggleActive.mutate({ id: banner.id, active: !banner.is_active })}
                      className={`p-2 rounded-xl transition-all border ${
                        banner.is_active 
                          ? 'text-emerald-400 border-emerald-500/10 hover:bg-emerald-500/10' 
                          : 'text-surface-muted border-surface-border hover:bg-surface-border'
                      }`}
                      title={banner.is_active ? 'تعطيل' : 'تفعيل'}
                    >
                      {banner.is_active ? <ToggleRight className="w-6 h-6" /> : <ToggleLeft className="w-6 h-6" />}
                    </button>
                    <button
                      onClick={() => setDeleteId(banner.id)}
                      className="p-2 rounded-xl text-red-400 border border-transparent hover:border-red-500/20 hover:bg-red-500/10 transition-all font-bold"
                      title="حذف"
                    >
                      <Trash2 className="w-5 h-5" />
                    </button>
                  </div>
                </div>
                
                <div className="flex items-center justify-between py-3 border-t border-surface-border mt-2">
                  <div className="flex items-center gap-2">
                      <span className="text-[10px] font-bold text-surface-muted uppercase">ترتيب العرض:</span>
                      <span className="text-sm font-bold text-white bg-surface-border/40 px-2 py-0.5 rounded-lg">{banner.display_order}</span>
                  </div>
                  <div className="text-[10px] text-surface-muted font-medium">
                      {banner.starts_at && `منذ: ${formatDate(banner.starts_at)}`}
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      <ConfirmDialog
        open={!!deleteId}
        onOpenChange={(v) => !v && setDeleteId(null)}
        title="هل أنت متأكد من حذف هذا الإعلان؟"
        description="هذا الإجراء سيؤدي لإزالة الإعلان نهائياً من التطبيق للمستخدمين."
        confirmLabel="حذف نهائي"
        variant="danger"
        loading={deleteBanner.isPending}
        onConfirm={() => {
          if (deleteId) deleteBanner.mutate(deleteId, { onSuccess: () => setDeleteId(null) })
        }}
      />

      <ConfirmDialog
        open={!!editingBanner}
        onOpenChange={(v) => !v && setEditingBanner(null)}
        title="تعديل الإعلان"
        description={
          editingBanner && (
            <div className="space-y-4 pt-4 text-right" dir="rtl">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                  <label className="text-xs text-surface-muted font-bold block mb-2">العنوان (عربي)</label>
                  <Input
                    value={editingBanner.title_ar ?? ''}
                    onChange={(e) => setEditingBanner({ ...editingBanner, title_ar: e.target.value })}
                  />
                </div>
                <div>
                  <label className="text-xs text-surface-muted font-bold block mb-2">Titre (Français)</label>
                  <Input
                    value={editingBanner.title_fr ?? ''}
                    onChange={(e) => setEditingBanner({ ...editingBanner, title_fr: e.target.value })}
                    dir="ltr"
                    className="text-left"
                  />
                </div>
                <div>
                  <label className="text-xs text-surface-muted font-bold block mb-2">Title (English)</label>
                  <Input
                    value={editingBanner.title_en ?? ''}
                    onChange={(e) => setEditingBanner({ ...editingBanner, title_en: e.target.value })}
                    dir="ltr"
                    className="text-left"
                  />
                </div>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="text-xs text-surface-muted font-bold block mb-2">رابط الصورة</label>
                  <Input
                    value={editingBanner.image_url}
                    onChange={(e) => setEditingBanner({ ...editingBanner, image_url: e.target.value })}
                    dir="ltr"
                    className="text-left"
                  />
                </div>
                <div>
                  <label className="text-xs text-surface-muted font-bold block mb-2">رابط الويب</label>
                  <Input
                    value={editingBanner.target_url ?? ''}
                    onChange={(e) => setEditingBanner({ ...editingBanner, target_url: e.target.value })}
                    dir="ltr"
                    className="text-left"
                  />
                </div>
              </div>
              <div>
                <label className="text-xs text-surface-muted font-bold block mb-2">ترتيب العرض</label>
                <Input
                  type="number"
                  value={editingBanner.display_order}
                  onChange={(e) => setEditingBanner({ ...editingBanner, display_order: parseInt(e.target.value) || 1 })}
                  className="w-32"
                />
              </div>
              <div className="flex items-center gap-3">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={editingBanner.is_active}
                    onChange={(e) => setEditingBanner({ ...editingBanner, is_active: e.target.checked })}
                    className="w-4 h-4 rounded border-surface-border bg-surface-base text-mazad-primary"
                  />
                  <span className="text-sm text-white font-bold">نشط</span>
                </label>
              </div>
            </div>
          )
        }
        confirmLabel="حفظ"
        loading={updateMutation.isPending}
        onConfirm={() => editingBanner && updateMutation.mutate(editingBanner)}
      />
    </div>
  )
}
