import { useRef, useState } from 'react';
import {
  Bell,
  AlertTriangle,
  Package,
  CreditCard,
  Send,
  Check,
  SearchX,
  Clock,
  CheckCircle,
  Trash2,
  Upload,
  Loader2,
  X
} from 'lucide-react';
import { toast } from 'sonner';
import { PageHeader } from '@/components/shared/PageHeader';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog';
import { Input } from '@/components/ui/input';
import { format } from 'date-fns';
import { fr } from 'date-fns/locale';
import { cn } from '@/lib/utils';
import client from '@/api/client';
import { useSendNotification, useFetchAdminNotifications, useDeleteNotification, useMarkNotificationAsRead, useMarkAllAsReadAdmin } from '@/hooks/useNotifications';

export const NotificationsPage = () => {
  const sendNotification = useSendNotification();
  const { data: notifications = [], isLoading: loading, refetch } = useFetchAdminNotifications();
  const deleteNotification = useDeleteNotification();
  const markAsRead = useMarkNotificationAsRead();
  const markAllAsRead = useMarkAllAsReadAdmin();

  const [filter, setFilter] = useState<'all' | 'unread'>('all');
  const [showSendModal, setShowSendModal] = useState(false)
  const [deleteId, setDeleteId] = useState<string | null>(null)
  const [newNotif, setNewNotif] = useState({ title: '', body: '', type: 'general', user_id: '', broadcast: true, image_url: '' })
  // Customer #22: optional broadcast image -- same upload-first-then-
  // reference pattern already used by BannersPage.tsx (handleFileUpload
  // below mirrors it), not a redesign of this modal.
  const [uploadingImage, setUploadingImage] = useState(false)
  const imageInputRef = useRef<HTMLInputElement>(null)

  const resetNewNotif = () => {
    setNewNotif({ title: '', body: '', type: 'general', user_id: '', broadcast: true, image_url: '' })
  }

  const handleSendNotification = async () => {
    try {
      await sendNotification.mutateAsync(newNotif);
      setShowSendModal(false);
      resetNewNotif();
      refetch();
    } catch (err) {
      // Error handling is done in the hook
      console.log(err, 'send notification error');
    }
  };

  // Customer #22: one optional image, uploaded via POST
  // /admin/notifications/upload (reuses MediaService/R2 exactly like
  // banners). Client-side validation mirrors BannersPage.tsx's
  // handleFileUpload, adjusted to this feature's stricter jpg/jpeg/png/webp
  // (no gif) + 10MB limits.
  const handleImageUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    const allowedTypes = ['image/jpeg', 'image/png', 'image/webp']
    if (!allowedTypes.includes(file.type)) {
      toast.error('يرجى اختيار ملف صورة فقط (jpg, png, webp)')
      return
    }

    if (file.size > 10 * 1024 * 1024) {
      toast.error('حجم الصورة كبير جداً (الحد الأقصى 10 ميجابايت)')
      return
    }

    setUploadingImage(true)
    try {
      const formData = new FormData()
      formData.append('file', file)

      const res = await client.post('/v1/api/admin/notifications/upload', formData, {
        timeout: 30_000,
        headers: { 'Content-Type': undefined },
      })

      const uploadedUrl = res.data.data.url
      setNewNotif((prev) => ({ ...prev, image_url: uploadedUrl }))
      toast.success('تم رفع الصورة بنجاح')
    } catch (err: any) {
      // Client feedback #22, item 11: a failed upload must never submit a
      // broken image_url -- image_url in state is only ever set on a
      // successful upload above, so a failure here simply leaves it as-is
      // (empty, or the previously-uploaded URL if the admin is replacing).
      const message = err.response?.data?.message || err.message || 'فشل رفع الصورة'
      toast.error(message)
    } finally {
      setUploadingImage(false)
      if (imageInputRef.current) imageInputRef.current.value = ''
    }
  }

  const handleDeleteNotification = async (id: string) => {
    try {
      await deleteNotification.mutateAsync(id);
      setDeleteId(null);
      refetch();
 
    } catch (err) {
      console.log(err, 'delete notification error');
    }
  };

  const handleMarkAsRead = async (id: string) => {
    try {
      await markAsRead.mutateAsync(id);
      refetch();
    } catch (err) {
      console.log(err, 'mark as read error');
      // Error handling is done in the hook
    }
  };

  const handleMarkAllAsRead = async () => {
    try {
      await markAllAsRead.mutateAsync();
      refetch();
    } catch (err) {
      console.log(err, 'mark all as read error');
      // Error handling is done in the hook
    }
  };

  const getIcon = (type: string, isRead: boolean) => {
    const iconClass = cn(
      "w-5 h-5",
      !isRead ? "text-mazad-primary" : "text-surface-muted"
    );
    
    switch (type) {
      case 'new_auction': return <Package className={iconClass} />;
      case 'report': return <AlertTriangle className={cn(iconClass, !isRead && "text-red-500")} />;
      case 'transaction': return <CreditCard className={cn(iconClass, !isRead && "text-green-500")} />;
      default: return <Bell className={iconClass} />;
    }
  };

  interface NotificationItem {
    id: string;
    type: string;
    is_read: boolean;
    title: string;
    body: string;
    created_at: string;
    reference_id?: string;
  }

  const uniqueNotifications = (notifications as NotificationItem[]).filter((notif, index, self) => 
    index === self.findIndex((n) => n.id === notif.id)
  );
  
  const filtered = filter === 'unread' 
    ? uniqueNotifications.filter(n => !n.is_read) 
    : uniqueNotifications;

  return (
    <div className="animate-fade-in space-y-6" dir="rtl">
      <PageHeader
        title="الإشعارات"
        subtitle="إدارة الإشعارات والأنشطة الأخيرة في المنصة"
        icon={Bell}
        action={{
          label: 'إرسال إشعار',
          icon: Send,
          onClick: () => setShowSendModal(true)
        }}
      >
        <div className="flex items-center gap-3">
          <div className="flex bg-surface-card border border-surface-border p-1 rounded-xl">
            <button
              onClick={() => setFilter('all')}
              className={cn(
                "px-4 py-1.5 rounded-lg text-xs font-bold transition-all whitespace-nowrap",
                filter === 'all' 
                  ? "bg-mazad-primary/10 text-mazad-primary border border-mazad-primary/20 shadow-sm" 
                  : "text-surface-muted hover:text-white"
              )}
            >
              الكل
            </button>
            <button
              onClick={() => setFilter('unread')}
              className={cn(
                "px-4 py-1.5 rounded-lg text-xs font-bold transition-all whitespace-nowrap",
                filter === 'unread' 
                  ? "bg-mazad-primary/10 text-mazad-primary border border-mazad-primary/20 shadow-sm" 
                  : "text-surface-muted hover:text-white"
              )}
            >
              غير مقروء
            </button>
          </div>
          
          <button 
            onClick={handleMarkAllAsRead}
            disabled={!notifications.some(n => !n.is_read)}
            className="flex items-center gap-2 px-4 py-2 bg-surface-card border border-surface-border rounded-xl text-xs font-bold text-white hover:bg-surface-border transition-all disabled:opacity-30"
          >
            <Check className="w-4 h-4" />
            تحديد الكل كمقروء
          </button>
        </div>
      </PageHeader>

      <div className="admin-card overflow-hidden shadow-2xl">
        {loading ? (
          <div className="p-20 text-center text-surface-muted animate-pulse font-medium">
            جاري تحميل الإشعارات...
          </div>
        ) : filtered.length === 0 ? (
          <div className="p-20 text-center text-surface-muted">
            <SearchX className="w-16 h-16 mx-auto text-surface-border mb-6" />
            <p className="text-xl font-display font-bold text-white mb-2">لا توجد تنبيهات</p>
            <p className="text-sm">سوف تظهر الإشعارات الجديدة هنا فور حدوثها</p>
          </div>
        ) : (
          <div className="divide-y divide-surface-border">
            {filtered.map((notif) => (
              <div 
                key={notif.id} 
                className={cn(
                  "p-5 hover:bg-surface-border/30 transition-all flex gap-5 group",
                  !notif.is_read && "bg-mazad-primary/5"
                )}
              >
                <div className={cn(
                  "p-3 rounded-2xl h-fit border transition-all",
                  !notif.is_read 
                    ? "bg-mazad-primary/10 border-mazad-primary/20 shadow-lg shadow-mazad-primary/10" 
                    : "bg-surface-base border-surface-border"
                )}>
                  {getIcon(notif.type, notif.is_read)}
                </div>
                
                <div className="flex-1 min-w-0">
                  <div className="flex justify-between items-start mb-2">
                    <h3 className={cn(
                      "text-base font-bold underline-offset-4 decoration-mazad-primary/30",
                      !notif.is_read ? "text-white" : "text-surface-muted"
                    )}>
                      {notif.title}
                    </h3>
                    <span className="text-[10px] font-bold text-surface-muted whitespace-nowrap flex items-center gap-1.5 bg-surface-base px-2.5 py-1 rounded-full border border-surface-border">
                      <Clock className="w-3 h-3" />
                      {format(new Date(notif.created_at), 'HH:mm - d MMM yyyy', { locale: fr })}
                    </span>
                  </div>
                  
                  <p className={cn(
                    "text-sm leading-relaxed max-w-2xl font-medium",
                    !notif.is_read ? "text-surface-muted/90" : "text-surface-muted/60"
                  )}>
                    {notif.body}
                  </p>
                  
                  {notif.reference_id && (
                    <button className="mt-4 text-xs font-bold text-mazad-primary hover:text-white flex items-center gap-1.5 group/btn transition-colors">
                      <span className="border-b border-mazad-primary/30 group-hover/btn:border-white">
                        عرض التفاصيل
                      </span>
                      <CheckCircle className="w-3.5 h-3.5 transition-transform group-hover/btn:translate-x-1" />
                    </button>
                  )}
                  
                  <div className="flex gap-2 mt-3">
                    {!notif.is_read && (
                      <button 
                        onClick={() => handleMarkAsRead(notif.id)}
                        className="text-xs font-bold text-mazad-primary hover:text-white flex items-center gap-1 transition-colors"
                      >
                        <Check className="w-3.5 h-3.5" />
                        تحديد كمقروء
                      </button>
                    )}
                    <button 
                      onClick={() => setDeleteId(notif.id)}
                      className="text-xs font-bold text-red-400 hover:text-red-300 flex items-center gap-1 transition-colors"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                      حذف
                    </button>
                  </div>
                </div>

                {!notif.is_read && (
                  <div className="w-2.5 h-2.5 rounded-full bg-mazad-primary mt-3 shrink-0 shadow-lg shadow-mazad-primary/50 animate-pulse" />
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      <ConfirmDialog
        open={showSendModal}
        onOpenChange={(v) => {
          setShowSendModal(v)
          // Client feedback #22, item 11: clear local image URL/preview
          // state on modal close/cancel -- an uploaded-then-cancelled image
          // stays orphaned in R2 (same accepted pattern as banners/FAQ), but
          // the form itself must not resurrect a stale preview/URL next
          // time the modal opens.
          if (!v) resetNewNotif()
        }}
        title="إرسال إشعار"
        description="Remplissez le formulaire ci-dessous pour envoyer une notification"
        confirmLabel="إرسال"
        loading={sendNotification.isPending}
        // Customer #22, hardening item 9: the send button must not be
        // usable while an image upload is still in flight -- otherwise a
        // send could fire before image_url is populated, silently sending
        // a text-only notification despite the admin having picked an
        // image.
        confirmDisabled={uploadingImage}
        onConfirm={handleSendNotification}
      >
        <div className="space-y-4 pt-4 text-right" dir="rtl">
          <div className="space-y-2">
            <label className="text-xs text-surface-muted font-bold block">عنوان الإشعار</label>
            <Input
              value={newNotif.title}
              onChange={(e) => setNewNotif({ ...newNotif, title: e.target.value })}
              placeholder="مثال: مزاد جديد متاح"
            />
          </div>
          <div className="space-y-2">
            <label className="text-xs text-surface-muted font-bold block">محتوى الإشعار</label>
            <textarea
              value={newNotif.body}
              onChange={(e) => setNewNotif({ ...newNotif, body: e.target.value })}
              className="w-full bg-surface-base border border-surface-border rounded-xl p-3 text-sm text-white min-h-[100px]"
              placeholder="نص الإشعار..."
            />
          </div>
          <div className="space-y-2">
            <label className="text-xs text-surface-muted font-bold block">نوع الإشعار</label>
            <select
              value={newNotif.type}
              onChange={(e) => setNewNotif({ ...newNotif, type: e.target.value })}
              className="w-full bg-surface-base border border-surface-border rounded-xl p-3 text-sm text-white"
            >
              <option value="general">عام</option>
              <option value="new_auction">مزاد جديد</option>
              <option value="transaction">معاملة</option>
            </select>
          </div>
          <div className="space-y-2">
            <label className="text-xs text-surface-muted font-bold block">صورة الإشعار (اختياري)</label>

            {newNotif.image_url && (
              <div className="relative mb-3 rounded-xl overflow-hidden border border-surface-border">
                <img src={newNotif.image_url} alt="Preview" className="w-full h-40 object-cover" />
                <button
                  type="button"
                  onClick={() => setNewNotif({ ...newNotif, image_url: '' })}
                  className="absolute top-2 left-2 p-1.5 bg-red-500/80 hover:bg-red-500 text-white rounded-lg transition-colors"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            )}

            <button
              type="button"
              onClick={() => imageInputRef.current?.click()}
              disabled={uploadingImage}
              className="w-full flex items-center justify-center gap-2 px-4 py-3 rounded-xl bg-surface-border/50 hover:bg-surface-border border border-surface-border text-white transition-all disabled:opacity-50"
            >
              {uploadingImage ? (
                <><Loader2 className="w-5 h-5 animate-spin" /> جاري الرفع...</>
              ) : newNotif.image_url ? (
                <><Upload className="w-5 h-5" /> تغيير الصورة</>
              ) : (
                <><Upload className="w-5 h-5" /> إضافة صورة</>
              )}
            </button>
            <input
              ref={imageInputRef}
              type="file"
              accept="image/jpeg,image/png,image/webp"
              onChange={handleImageUpload}
              className="hidden"
            />
            <p className="text-xs text-surface-muted mt-1">JPG, PNG, WebP (الحد الأقصى 10MB) — اختياري</p>
          </div>
        </div>
      </ConfirmDialog>

      <ConfirmDialog
        open={!!deleteId}
        onOpenChange={(v) => !v && setDeleteId(null)}
        title="حذف الإشعار"
        description="هل أنت متأكد من حذف هذا الإشعار؟"
        variant="danger"
        confirmLabel="حذف"
        loading={deleteNotification.isPending}
        onConfirm={() => deleteId && handleDeleteNotification(deleteId)}
      />
    </div>
  );
};
