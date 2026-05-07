import { useState, useRef } from 'react'
import { Plus, Video, Trash2, PlayCircle, ExternalLink, Edit2, Upload, X, Loader2, Image as ImageIcon } from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { Input } from '@/components/ui/input'
import { useTutorials, useCreateTutorial, useUpdateTutorial, useDeleteTutorial } from '@/hooks/useContent'
import type { Tutorial } from '@/types/api'
import client from '@/api/client'
import { toast } from 'sonner'

export function TutorialsPage() {
  const [editingTutorial, setEditingTutorial] = useState<Tutorial | null>(null)
  const [isAdding, setIsAdding] = useState(false)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [uploadingVideo, setUploadingVideo] = useState(false)
  const [uploadingThumbnail, setUploadingThumbnail] = useState(false)
  const videoInputRef = useRef<HTMLInputElement>(null)
  const thumbnailInputRef = useRef<HTMLInputElement>(null)

  const [newTutorial, setNewTutorial] = useState({
    title_ar: '',
    video_url: '',
    thumbnail_url: '',
    category: 'عام',
    display_order: 1
  })

  const { data: tutorials, isLoading } = useTutorials()
  const createTutorial = useCreateTutorial()
  const updateTutorial = useUpdateTutorial()
  const deleteTutorial = useDeleteTutorial()

  // Video upload handler
  const handleVideoUpload = async (e: React.ChangeEvent<HTMLInputElement>, isEdit = false) => {
    const file = e.target.files?.[0]
    if (!file) return

    // Validate file type
    if (!file.type.startsWith('video/')) {
      toast.error('يرجى اختيار ملف فيديو فقط (mp4, webm, mov)')
      return
    }

    // Validate file size (max 100MB)
    if (file.size > 100 * 1024 * 1024) {
      toast.error('حجم الفيديو كبير جداً (الحد الأقصى 100 ميجابايت)')
      return
    }

    setUploadingVideo(true)
    try {
      const formData = new FormData()
      formData.append('file', file)

      const res = await client.post('/v1/api/admin/tutorials/upload-video', formData, {
        timeout: 120_000, // 2 min for videos
        headers: { 'Content-Type': undefined },
      })

      const uploadedUrl = res.data.data.url
      
      if (isEdit && editingTutorial) {
        setEditingTutorial({ ...editingTutorial, video_url: uploadedUrl })
      } else {
        setNewTutorial({ ...newTutorial, video_url: uploadedUrl })
      }
      
      toast.success('تم رفع الفيديو بنجاح')
    } catch (err: any) {
      const message = err.response?.data?.message || err.message || 'فشل رفع الفيديو'
      toast.error(message)
    } finally {
      setUploadingVideo(false)
      if (videoInputRef.current) videoInputRef.current.value = ''
    }
  }

  // Thumbnail upload handler
  const handleThumbnailUpload = async (e: React.ChangeEvent<HTMLInputElement>, isEdit = false) => {
    const file = e.target.files?.[0]
    if (!file) return

    // Validate file type
    if (!file.type.startsWith('image/')) {
      toast.error('يرجى اختيار ملف صورة فقط (jpg, png, webp)')
      return
    }

    // Validate file size (max 10MB)
    if (file.size > 10 * 1024 * 1024) {
      toast.error('حجم الصورة كبير جداً (الحد الأقصى 10 ميجابايت)')
      return
    }

    setUploadingThumbnail(true)
    try {
      const formData = new FormData()
      formData.append('file', file)

      const res = await client.post('/v1/api/admin/tutorials/upload-thumbnail', formData, {
        timeout: 30_000,
        headers: { 'Content-Type': undefined },
      })

      const uploadedUrl = res.data.data.url
      
      if (isEdit && editingTutorial) {
        setEditingTutorial({ ...editingTutorial, thumbnail_url: uploadedUrl })
      } else {
        setNewTutorial({ ...newTutorial, thumbnail_url: uploadedUrl })
      }
      
      toast.success('تم رفع الصورة المصغرة بنجاح')
    } catch (err: any) {
      const message = err.response?.data?.message || err.message || 'فشل رفع الصورة المصغرة'
      toast.error(message)
    } finally {
      setUploadingThumbnail(false)
      if (thumbnailInputRef.current) thumbnailInputRef.current.value = ''
    }
  }

  const handleCreate = () => {
    if (!newTutorial.title_ar || !newTutorial.video_url) return
    createTutorial.mutate(newTutorial, {
      onSuccess: () => {
        setIsAdding(false)
        setNewTutorial({ title_ar: '', video_url: '', thumbnail_url: '', category: 'عام', display_order: (tutorials?.length ?? 0) + 1 })
      }
    })
  }

  const handleUpdate = () => {
    if (!editingTutorial?.title_ar || !editingTutorial?.video_url) return
    updateTutorial.mutate(editingTutorial, {
      onSuccess: () => setEditingTutorial(null)
    })
  }

  return (
    <div className="animate-fade-in" dir="rtl">
      <PageHeader 
        title="شروحات الفيديو" 
        subtitle="إدارة مقاطع الفيديو التعليمية للمستخدمين"
        icon={Video}
        action={{
          label: 'إضافة فيديو جديد',
          icon: Plus,
          onClick: () => setIsAdding(true)
        }}
      />

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {isLoading ? (
          [1,2,3].map(i => (
            <div key={i} className="admin-card aspect-video animate-pulse bg-surface-border/20" />
          ))
        ) : tutorials?.length === 0 ? (
          <div className="admin-card p-20 text-center col-span-full">
            <Video className="w-12 h-12 text-surface-muted mx-auto mb-4" />
            <h3 className="text-white font-bold mb-2">لا توجد فيديوهات حالياً</h3>
            <p className="text-surface-muted text-sm">ابدأ بإضافة مقاطع فيديو توضح كيفية استخدام التطبيق.</p>
          </div>
        ) : (
          tutorials?.map((video) => (
            <div key={video.id} className="admin-card group overflow-hidden flex flex-col">
              <div className="relative aspect-video bg-black/40 flex items-center justify-center overflow-hidden">
                {video.thumbnail_url ? (
                  <img src={video.thumbnail_url} alt="" className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" />
                ) : (
                  <PlayCircle className="w-12 h-12 text-mazad-primary/40" />
                )}
                <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-3">
                  <a href={video.video_url} target="_blank" rel="noreferrer" 
                     className="w-10 h-10 rounded-full bg-mazad-primary text-white flex items-center justify-center hover:scale-110 transition-transform">
                    <ExternalLink className="w-5 h-5" />
                  </a>
                  <button 
                    onClick={() => setEditingTutorial(video)}
                    className="w-10 h-10 rounded-full bg-white text-mazad-primary flex items-center justify-center hover:scale-110 transition-transform">
                    <Edit2 className="w-5 h-5" />
                  </button>
                </div>
                <div className="absolute top-3 right-3">
                  <span className="px-2 py-1 rounded bg-black/60 text-[10px] text-white font-bold backdrop-blur-sm border border-white/10 uppercase tracking-widest">
                    {video.category ?? 'عام'}
                  </span>
                </div>
              </div>
              <div className="p-4 flex-1 flex flex-col">
                <div className="flex items-start justify-between gap-2 mb-2">
                  <h4 className="text-white font-bold text-sm leading-tight">{video.title_ar}</h4>
                  <button 
                    onClick={() => setDeleteId(video.id)}
                    className="p-1.5 rounded-lg text-red-500/60 hover:text-red-400 hover:bg-red-500/10 transition-all opacity-0 group-hover:opacity-100"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
                <div className="mt-auto flex items-center justify-between text-surface-muted">
                  <span className="text-[10px] font-bold font-mono">#{video.display_order}</span>
                  <p className="text-[10px] font-medium truncate max-w-[150px]">{video.video_url}</p>
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Add Modal */}
      <ConfirmDialog
        open={isAdding}
        onOpenChange={setIsAdding}
        title="إضافة فيديو تعليمي"
        description={
          <div className="space-y-4 pt-4 text-right" dir="rtl">
            <div className="space-y-2">
              <label className="text-xs text-surface-muted font-bold block">عنوان الفيديو (بالعربية) <span className="text-red-400">*</span></label>
              <Input 
                value={newTutorial.title_ar}
                onChange={(e) => setNewTutorial({...newTutorial, title_ar: e.target.value})}
                placeholder="مثال: كيف تزايد في المزاد..."
              />
            </div>
            
            {/* Video Upload Section */}
            <div className="space-y-2">
              <label className="text-xs text-surface-muted font-bold block">ملف الفيديو <span className="text-red-400">*</span></label>
              
              {/* Video Preview */}
              {newTutorial.video_url && (
                <div className="relative mb-3 rounded-xl overflow-hidden border border-surface-border bg-black">
                  <video src={newTutorial.video_url} className="w-full h-32 object-contain" controls />
                  <button
                    onClick={() => setNewTutorial({ ...newTutorial, video_url: '' })}
                    className="absolute top-2 left-2 p-1.5 bg-red-500/80 hover:bg-red-500 text-white rounded-lg transition-colors"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              )}
              
              {/* Upload Button */}
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => videoInputRef.current?.click()}
                  disabled={uploadingVideo}
                  className="flex-1 flex items-center justify-center gap-2 px-4 py-3 rounded-xl bg-surface-border/50 hover:bg-surface-border border border-surface-border text-white transition-all disabled:opacity-50"
                >
                  {uploadingVideo ? (
                    <><Loader2 className="w-5 h-5 animate-spin" /> جاري رفع الفيديو...</>
                  ) : (
                    <><Upload className="w-5 h-5" /> رفع فيديو من الجهاز</>
                  )}
                </button>
                <span className="text-surface-muted text-sm self-center">أو</span>
                <Input
                  value={newTutorial.video_url}
                  onChange={(e) => setNewTutorial({...newTutorial, video_url: e.target.value})}
                  placeholder="https://example.com/video.mp4"
                  dir="ltr"
                  className="flex-1 text-left"
                />
              </div>
              <input
                ref={videoInputRef}
                type="file"
                accept="video/mp4,video/webm,video/quicktime"
                onChange={(e) => handleVideoUpload(e)}
                className="hidden"
              />
              <p className="text-xs text-surface-muted">الفيديوهات المسموح بها: MP4, WebM, MOV (الحد الأقصى 100MB)</p>
            </div>
            
            {/* Thumbnail Upload Section */}
            <div className="space-y-2">
              <label className="text-xs text-surface-muted font-bold block">صورة مصغرة (Thumbnail)</label>
              
              {/* Thumbnail Preview */}
              {newTutorial.thumbnail_url && (
                <div className="relative mb-3 rounded-xl overflow-hidden border border-surface-border">
                  <img src={newTutorial.thumbnail_url} alt="Thumbnail" className="w-full h-24 object-cover" />
                  <button
                    onClick={() => setNewTutorial({ ...newTutorial, thumbnail_url: '' })}
                    className="absolute top-2 left-2 p-1.5 bg-red-500/80 hover:bg-red-500 text-white rounded-lg transition-colors"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              )}
              
              {/* Upload Button */}
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => thumbnailInputRef.current?.click()}
                  disabled={uploadingThumbnail}
                  className="flex-1 flex items-center justify-center gap-2 px-4 py-2 rounded-xl bg-surface-border/50 hover:bg-surface-border border border-surface-border text-white transition-all disabled:opacity-50"
                >
                  {uploadingThumbnail ? (
                    <><Loader2 className="w-4 h-4 animate-spin" /> جاري الرفع...</>
                  ) : (
                    <><ImageIcon className="w-4 h-4" /> رفع صورة مصغرة</>
                  )}
                </button>
                <span className="text-surface-muted text-sm self-center">أو</span>
                <Input
                  value={newTutorial.thumbnail_url}
                  onChange={(e) => setNewTutorial({...newTutorial, thumbnail_url: e.target.value})}
                  placeholder="https://example.com/thumb.jpg"
                  dir="ltr"
                  className="flex-1 text-left"
                />
              </div>
              <input
                ref={thumbnailInputRef}
                type="file"
                accept="image/jpeg,image/png,image/webp"
                onChange={(e) => handleThumbnailUpload(e)}
                className="hidden"
              />
              <p className="text-xs text-surface-muted">الصور المسموح بها: JPG, PNG, WebP (الحد الأقصى 10MB)</p>
            </div>
            
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-xs text-surface-muted font-bold block">الفئة</label>
                <Input 
                  value={newTutorial.category}
                  onChange={(e) => setNewTutorial({...newTutorial, category: e.target.value})}
                />
              </div>
              <div className="space-y-2">
                <label className="text-xs text-surface-muted font-bold block">ترتيب العرض</label>
                <Input 
                  type="number"
                  value={newTutorial.display_order}
                  onChange={(e) => setNewTutorial({...newTutorial, display_order: parseInt(e.target.value) || 1})}
                />
              </div>
            </div>
          </div>
        }
        confirmLabel="حفظ الفيديو"
        loading={createTutorial.isPending}
        onConfirm={handleCreate}
      />

      {/* Edit Modal */}
      <ConfirmDialog
        open={!!editingTutorial}
        onOpenChange={(v) => !v && setEditingTutorial(null)}
        title="تعديل الفيديو التعليمي"
        description={
          editingTutorial && (
            <div className="space-y-4 pt-4 text-right" dir="rtl">
              <div className="space-y-2">
                <label className="text-xs text-surface-muted font-bold block">عنوان الفيديو (بالعربية)</label>
                <Input 
                  value={editingTutorial.title_ar}
                  onChange={(e) => setEditingTutorial({...editingTutorial, title_ar: e.target.value})}
                />
              </div>
              
              {/* Video Upload Section for Edit */}
              <div className="space-y-2">
                <label className="text-xs text-surface-muted font-bold block">ملف الفيديو</label>
                
                {/* Video Preview */}
                {editingTutorial.video_url && (
                  <div className="relative mb-3 rounded-xl overflow-hidden border border-surface-border bg-black">
                    <video src={editingTutorial.video_url} className="w-full h-24 object-contain" controls />
                    <button
                      onClick={() => setEditingTutorial({ ...editingTutorial, video_url: '' })}
                      className="absolute top-2 left-2 p-1.5 bg-red-500/80 hover:bg-red-500 text-white rounded-lg transition-colors"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  </div>
                )}
                
                {/* Upload Button */}
                <div className="flex gap-3">
                  <button
                    type="button"
                    onClick={() => videoInputRef.current?.click()}
                    disabled={uploadingVideo}
                    className="flex-1 flex items-center justify-center gap-2 px-4 py-2 rounded-xl bg-surface-border/50 hover:bg-surface-border border border-surface-border text-white transition-all disabled:opacity-50"
                  >
                    {uploadingVideo ? (
                      <><Loader2 className="w-4 h-4 animate-spin" /> جاري رفع الفيديو...</>
                    ) : (
                      <><Upload className="w-4 h-4" /> تغيير الفيديو</>
                    )}
                  </button>
                  <span className="text-surface-muted text-sm self-center">أو</span>
                  <Input
                    value={editingTutorial.video_url}
                    onChange={(e) => setEditingTutorial({...editingTutorial, video_url: e.target.value})}
                    placeholder="https://example.com/video.mp4"
                    dir="ltr"
                    className="flex-1 text-left"
                  />
                </div>
                <input
                  ref={videoInputRef}
                  type="file"
                  accept="video/mp4,video/webm,video/quicktime"
                  onChange={(e) => handleVideoUpload(e, true)}
                  className="hidden"
                />
              </div>
              
              {/* Thumbnail Upload Section for Edit */}
              <div className="space-y-2">
                <label className="text-xs text-surface-muted font-bold block">صورة مصغرة</label>
                
                {/* Thumbnail Preview */}
                {editingTutorial.thumbnail_url && (
                  <div className="relative mb-3 rounded-xl overflow-hidden border border-surface-border">
                    <img src={editingTutorial.thumbnail_url} alt="Thumbnail" className="w-full h-20 object-cover" />
                    <button
                      onClick={() => setEditingTutorial({ ...editingTutorial, thumbnail_url: '' })}
                      className="absolute top-2 left-2 p-1.5 bg-red-500/80 hover:bg-red-500 text-white rounded-lg transition-colors"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  </div>
                )}
                
                {/* Upload Button */}
                <div className="flex gap-3">
                  <button
                    type="button"
                    onClick={() => thumbnailInputRef.current?.click()}
                    disabled={uploadingThumbnail}
                    className="flex-1 flex items-center justify-center gap-2 px-4 py-2 rounded-xl bg-surface-border/50 hover:bg-surface-border border border-surface-border text-white transition-all disabled:opacity-50"
                  >
                    {uploadingThumbnail ? (
                      <><Loader2 className="w-4 h-4 animate-spin" /> جاري الرفع...</>
                    ) : (
                      <><ImageIcon className="w-4 h-4" /> تغيير الصورة المصغرة</>
                    )}
                  </button>
                  <span className="text-surface-muted text-sm self-center">أو</span>
                  <Input
                    value={editingTutorial.thumbnail_url || ''}
                    onChange={(e) => setEditingTutorial({...editingTutorial, thumbnail_url: e.target.value})}
                    placeholder="https://example.com/thumb.jpg"
                    dir="ltr"
                    className="flex-1 text-left"
                  />
                </div>
                <input
                  ref={thumbnailInputRef}
                  type="file"
                  accept="image/jpeg,image/png,image/webp"
                  onChange={(e) => handleThumbnailUpload(e, true)}
                  className="hidden"
                />
              </div>
              
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label className="text-xs text-surface-muted font-bold block">الفئة</label>
                  <Input 
                    value={editingTutorial.category ?? ''}
                    onChange={(e) => setEditingTutorial({...editingTutorial, category: e.target.value})}
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-xs text-surface-muted font-bold block">ترتيب العرض</label>
                  <Input 
                    type="number"
                    value={editingTutorial.display_order}
                    onChange={(e) => setEditingTutorial({...editingTutorial, display_order: parseInt(e.target.value) || 1})}
                  />
                </div>
              </div>
            </div>
          )
        }
        confirmLabel="تحديث"
        loading={updateTutorial.isPending}
        onConfirm={handleUpdate}
      />

      <ConfirmDialog
        open={!!deleteId}
        onOpenChange={(v) => !v && setDeleteId(null)}
        title="حذف الفيديو"
        description="هل أنت متأكد من رغبتك في حذف هذا الفيديو؟ سيتم اختفاؤه من قسم الشروحات لدى المستخدمين."
        variant="danger"
        confirmLabel="حذف الآن"
        loading={deleteTutorial.isPending}
        onConfirm={() => {
          if (deleteId) deleteTutorial.mutate(deleteId, { onSuccess: () => setDeleteId(null) })
        }}
      />
    </div>
  )
}
