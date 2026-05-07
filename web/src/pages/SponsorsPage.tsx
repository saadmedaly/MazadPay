import { useState } from 'react'
import { Plus, Trash2, ToggleLeft, ToggleRight, Building2, Loader2, AlertCircle, Edit2, Upload, X, Phone, Globe } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { PageHeader } from '@/components/shared/PageHeader'
import { EmptyState } from '@/components/shared/EmptyState'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { Input } from '@/components/ui/input'
import { formatDate } from '@/lib/formatters'
import client from '@/api/client'

interface Sponsor {
  id: string
  name: string
  phone: string
  image_url: string
  link_url: string | null
  is_active: boolean
  created_at: string
}

function useSponsors() {
  return useQuery({
    queryKey: ['sponsors'],
    queryFn: async () => {
      const { data } = await client.get('/v1/api/admin/sponsors')
      return data.data.sponsors as Sponsor[]
    },
  })
}

export function SponsorsPage() {
  const qc = useQueryClient()
  const { data: sponsors = [], isLoading, isError } = useSponsors()
  const [showForm, setShowForm] = useState(false)
  const [editingSponsor, setEditingSponsor] = useState<Sponsor | null>(null)
  const [newSponsor, setNewSponsor] = useState({
    name: '',
    phone: '',
    image_url: '',
    link_url: '',
    is_active: true
  })
  const [deleteId, setDeleteId] = useState<string | null>(null)
  const [uploadingFile, setUploadingFile] = useState(false)

  const createMutation = useMutation({
    mutationFn: (s: typeof newSponsor) => client.post('/v1/api/admin/sponsors', s),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['sponsors'] })
      setShowForm(false)
      setNewSponsor({ name: '', phone: '', image_url: '', link_url: '', is_active: true })
      toast.success('تم إضافة الراعي بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })

  const toggleActive = useMutation({
    mutationFn: ({ id }: { id: string }) => client.patch(`/v1/api/admin/sponsors/${id}/toggle`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['sponsors'] })
      toast.success('تم تحديث حالة الراعي')
    },
    onError: (err: Error) => toast.error(err.message),
  })

  const updateMutation = useMutation({
    mutationFn: (s: Sponsor) => client.put(`/v1/api/admin/sponsors/${s.id}`, s),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['sponsors'] })
      toast.success('تم تحديث بيانات الراعي')
      setEditingSponsor(null)
    },
    onError: (err: Error) => toast.error(err.message),
  })

  const deleteSponsor = useMutation({
    mutationFn: (id: string) => client.delete(`/v1/api/admin/sponsors/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['sponsors'] })
      toast.success('تم حذف الراعي بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>, isEdit = false) => {
    const file = e.target.files?.[0]
    if (!file) return
    
    setUploadingFile(true)
    try {
      const formData = new FormData()
      formData.append('file', file)
      // Reuse banners upload endpoint or generic one if exists
      const res = await client.post('/v1/api/admin/banners/upload', formData, {
        headers: { 'Content-Type': undefined },
      })
      const url = res.data.data.url
      if (isEdit && editingSponsor) {
        setEditingSponsor({ ...editingSponsor, image_url: url })
      } else {
        setNewSponsor({ ...newSponsor, image_url: url })
      }
      toast.success('تم رفع الصورة')
    } catch (err: any) {
      toast.error('فشل رفع الصورة')
    } finally {
      setUploadingFile(false)
    }
  }

  if (isError) return (
    <div className="admin-card p-20 text-center flex flex-col items-center gap-4">
      <AlertCircle className="w-12 h-12 text-red-500/20" />
      <p className="text-red-400 font-bold">فشل تحميل الرعاة</p>
      <button onClick={() => window.location.reload()} className="bg-surface-border text-white px-6 py-2 rounded-xl text-sm font-bold">إعادة المحاولة</button>
    </div>
  )

  return (
    <div className="animate-fade-in" dir="rtl">
      <PageHeader title="إدارة الرعاة" subtitle="إدارة شركاء ورعاة التطبيق">
        <button
           onClick={() => setShowForm(!showForm)}
           className="flex items-center gap-2 px-6 py-2.5 rounded-xl bg-mazad-primary text-white text-sm font-bold shadow-lg shadow-mazad-primary/20 hover:bg-mazad-primary-dk transition-all"
        >
          <Plus className="w-4 h-4" />
          {showForm ? 'إلغاء' : 'إضافة راعي جديد'}
        </button>
      </PageHeader>

      {showForm && (
        <div className="admin-card p-6 mb-8 border-mazad-primary/30 animate-slide-in">
          <h3 className="font-display font-bold text-white text-lg mb-6">إضافة راعي جديد</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
            <div>
              <label className="text-xs font-bold text-surface-muted uppercase tracking-widest block mb-2">اسم الراعي</label>
              <Input
                value={newSponsor.name}
                onChange={(e) => setNewSponsor({ ...newSponsor, name: e.target.value })}
                placeholder="مثال: بنكيلي، ماتل..."
              />
            </div>
            <div>
              <label className="text-xs font-bold text-surface-muted uppercase tracking-widest block mb-2">رقم الهاتف</label>
              <Input
                value={newSponsor.phone}
                onChange={(e) => setNewSponsor({ ...newSponsor, phone: e.target.value })}
                placeholder="44XXXXXX"
                dir="ltr"
              />
            </div>
          </div>
          
          <div className="mb-4">
            <label className="text-xs font-bold text-surface-muted uppercase tracking-widest block mb-2">صورة الشعار</label>
            <div className="flex gap-3">
              <button
                type="button"
                onClick={() => document.getElementById('sponsor-upload')?.click()}
                disabled={uploadingFile}
                className="flex-1 flex items-center justify-center gap-2 px-4 py-3 rounded-xl bg-surface-border/50 hover:bg-surface-border border border-surface-border text-white transition-all"
              >
                {uploadingFile ? <Loader2 className="w-5 h-5 animate-spin" /> : <Upload className="w-5 h-5" />}
                {uploadingFile ? 'جاري الرفع...' : 'رفع شعار'}
              </button>
              <Input
                value={newSponsor.image_url}
                onChange={(e) => setNewSponsor({ ...newSponsor, image_url: e.target.value })}
                placeholder="URL الصورة"
                className="flex-1"
                dir="ltr"
              />
            </div>
            <input id="sponsor-upload" type="file" className="hidden" onChange={handleFileUpload} />
          </div>

          <div className="mb-4">
            <label className="text-xs font-bold text-surface-muted uppercase tracking-widest block mb-2">رابط الموقع</label>
            <Input
              value={newSponsor.link_url}
              onChange={(e) => setNewSponsor({ ...newSponsor, link_url: e.target.value })}
              placeholder="https://..."
              dir="ltr"
            />
          </div>

          <div className="flex gap-3 justify-end mt-6">
            <button onClick={() => setShowForm(false)} className="px-6 py-2.5 rounded-xl text-sm font-bold text-surface-muted border border-surface-border">إلغاء</button>
            <button
              disabled={!newSponsor.name || !newSponsor.image_url || createMutation.isPending}
              onClick={() => createMutation.mutate(newSponsor)}
              className="px-8 py-2.5 rounded-xl text-sm font-bold bg-mazad-primary text-white shadow-lg"
            >
              حفظ الراعي
            </button>
          </div>
        </div>
      )}

      {isLoading ? (
        <div className="admin-card p-20 text-center"><LoadingSpinner /></div>
      ) : (sponsors || []).length === 0 ? (
        <div className="admin-card">
          <EmptyState icon={Building2} title="لا يوجد رعاة" description="ابدأ بإضافة أول راعي للتطبيق." />
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {(sponsors || []).map((sponsor) => (
            <div key={sponsor.id} className="admin-card p-6 flex flex-col items-center text-center group">
              <div className="relative w-24 h-24 mb-4 rounded-2xl overflow-hidden bg-white p-2 border border-surface-border group-hover:border-mazad-primary/30 transition-colors">
                <img src={sponsor.image_url} alt={sponsor.name} className="w-full h-full object-contain" />
              </div>
              <h3 className="font-bold text-white text-lg mb-1">{sponsor.name}</h3>
              <p className="text-xs text-surface-muted flex items-center gap-1 mb-4"><Phone className="w-3 h-3" /> {sponsor.phone}</p>
              
              <div className="flex items-center gap-2 mt-auto w-full pt-4 border-t border-surface-border">
                <button
                  onClick={() => setEditingSponsor(sponsor)}
                  className="flex-1 flex items-center justify-center p-2 rounded-lg bg-surface-border/30 text-white hover:bg-surface-border transition-all"
                >
                  <Edit2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => toggleActive.mutate({ id: sponsor.id })}
                  className={`flex-1 flex items-center justify-center p-2 rounded-lg border transition-all ${
                    sponsor.is_active ? 'text-emerald-400 border-emerald-500/20 bg-emerald-500/5' : 'text-surface-muted border-surface-border'
                  }`}
                >
                  {sponsor.is_active ? <ToggleRight className="w-6 h-6" /> : <ToggleLeft className="w-6 h-6" />}
                </button>
                <button
                  onClick={() => setDeleteId(sponsor.id)}
                  className="flex-1 flex items-center justify-center p-2 rounded-lg bg-red-500/10 text-red-400 hover:bg-red-500/20 transition-all"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
              
              {sponsor.link_url && (
                <a href={sponsor.link_url} target="_blank" rel="noreferrer" className="mt-3 text-[10px] text-mazad-primary flex items-center gap-1 hover:underline">
                  <Globe className="w-3 h-3" /> زيارة الموقع
                </a>
              )}
            </div>
          ))}
        </div>
      )}

      <ConfirmDialog
        open={!!deleteId}
        onOpenChange={(v) => !v && setDeleteId(null)}
        title="حذف الراعي؟"
        description="هل أنت متأكد من حذف هذا الراعي نهائياً؟"
        confirmLabel="حذف"
        variant="danger"
        onConfirm={() => deleteId && deleteSponsor.mutate(deleteId, { onSuccess: () => setDeleteId(null) })}
      />

      <ConfirmDialog
        open={!!editingSponsor}
        onOpenChange={(v) => !v && setEditingSponsor(null)}
        title="تعديل الراعي"
        description={
          editingSponsor && (
            <div className="space-y-4 pt-4 text-right" dir="rtl">
              <Input value={editingSponsor.name} onChange={(e) => setEditingSponsor({...editingSponsor, name: e.target.value})} placeholder="الاسم" />
              <Input value={editingSponsor.phone} onChange={(e) => setEditingSponsor({...editingSponsor, phone: e.target.value})} placeholder="الهاتف" dir="ltr" />
              <Input value={editingSponsor.link_url ?? ''} onChange={(e) => setEditingSponsor({...editingSponsor, link_url: e.target.value})} placeholder="الموقع" dir="ltr" />
            </div>
          )
        }
        onConfirm={() => editingSponsor && updateMutation.mutate(editingSponsor)}
      />
    </div>
  )
}
