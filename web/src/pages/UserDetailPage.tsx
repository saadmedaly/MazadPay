import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  Mail,
  Phone,
  MapPin,
  Calendar,
  Shield,
  ShieldOff,
  Gavel,
  CreditCard,
  AlertCircle,
  Pencil,
  Eye,
  EyeOff,
  Lock,
  User,
  Globe,
  CheckCircle,
  Save,
  X,
  Wallet,
  WalletCards,
} from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { StatusBadge } from '@/components/shared/StatusBadge'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { DataTable } from '@/components/shared/DataTable'
import { useUser, useUserHistory, useBlockUser, useUpdateProfile, useUserWallet, useSetWalletDisabled } from '@/hooks/useUsers'
import { formatDate, formatPrice, formatFullPhone, shortID } from '@/lib/formatters'
import { cn } from '@/lib/utils'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'

export function UserDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState<'auctions' | 'transactions'>('auctions')
  const [blockConfirm, setBlockConfirm] = useState(false)
  const [walletDisableConfirm, setWalletDisableConfirm] = useState(false)
  const [walletDisableReason, setWalletDisableReason] = useState('')
  const [isEditing, setIsEditing] = useState(false)
  const [showPin, setShowPin] = useState(false)
  const [showNewPin, setShowNewPin] = useState(false)
  const [editForm, setEditForm] = useState({
    full_name: '',
    email: '',
    city: '',
    phone: '',
  })
  const [pinForm, setPinForm] = useState({
    new_pin: '',
    confirm_pin: '',
  })
  const [showPinModal, setShowPinModal] = useState(false)

  const { data: user, isLoading, isError, refetch } = useUser(id!)
  const { data: history, isLoading: historyLoading } = useUserHistory(id!, activeTab)
  const { data: wallet } = useUserWallet(id!)
  const blockUser = useBlockUser()
  const updateProfile = useUpdateProfile()
  const setWalletDisabled = useSetWalletDisabled()

  // Initialize edit form when user data loads
  const startEditing = () => {
    if (user) {
      setEditForm({
        full_name: user.full_name || '',
        email: user.email || '',
        city: user.city || '',
        phone: user.phone || '',
      })
      setIsEditing(true)
    }
  }

  const cancelEditing = () => {
    setIsEditing(false)
    setEditForm({ full_name: '', email: '', city: '', phone: '' })
  }

  const saveProfile = () => {
    if (!user) return
    
    updateProfile.mutate(
      { 
        full_name: editForm.full_name,
        email: editForm.email,
        city: editForm.city 
      },
      {
        onSuccess: () => {
          toast.success('تم تحديث الملف الشخصي بنجاح')
          setIsEditing(false)
          refetch()
        },
        onError: (err: any) => {
          toast.error(err?.response?.data?.message || 'فشل تحديث الملف الشخصي')
        }
      }
    )
  }

  const handleResetPin = () => {
    if (pinForm.new_pin.length < 4) {
      toast.error('يجب أن يكون PIN مكون من 4 أرقام على الأقل')
      return
    }
    if (pinForm.new_pin !== pinForm.confirm_pin) {
      toast.error('PIN غير متطابق')
      return
    }
    
    // TODO: Call API to reset PIN
    toast.success('تم إعادة تعيين PIN بنجاح')
    setShowPinModal(false)
    setPinForm({ new_pin: '', confirm_pin: '' })
  }

  if (isLoading) return <LoadingSpinner fullPage label="جاري تحميل بيانات المستخدم..." />

  if (isError || !user) return (
    <div className="flex flex-col items-center justify-center h-64 text-surface-muted gap-4">
      <AlertCircle className="w-12 h-12 opacity-20" />
      <p className="font-bold">فشل في العثور على المستخدم</p>
      <button onClick={() => navigate(-1)} className="text-mazad-primary text-sm font-bold">رجوع للوراء</button>
    </div>
  )

  const handleBlockToggle = () => {
    blockUser.mutate(
      { id: user.id, block: user.is_active },
      { onSuccess: () => setBlockConfirm(false) }
    )
  }

  // MAZADPAY -- admin wallet controls: re-enable needs no reason/confirm
  // dialog (a low-risk, easily-reversible action); disable does, since it
  // blocks the user's spending and should carry an explicit reason.
  const handleWalletDisableToggle = () => {
    if (!wallet) return
    if (wallet.is_disabled) {
      setWalletDisabled.mutate({ id: user.id, disabled: false })
      return
    }
    setWalletDisableConfirm(true)
  }

  const confirmWalletDisable = () => {
    setWalletDisabled.mutate(
      { id: user.id, disabled: true, reason: walletDisableReason },
      { onSuccess: () => { setWalletDisableConfirm(false); setWalletDisableReason('') } }
    )
  }

  return (
    <div className="animate-fade-in max-w-6xl" dir="rtl">
      <PageHeader title="ملف المستخدم">
        <div className="flex items-center gap-2">
          {!isEditing && user && (
            <button
              onClick={startEditing}
              className="flex items-center gap-2 text-sm font-bold text-mazad-primary hover:text-white transition-colors bg-mazad-primary/10 px-3 py-1.5 rounded-lg"
            >
              <Pencil className="w-4 h-4" />
              تعديل
            </button>
          )}
          <button
            onClick={() => navigate(-1)}
            className="flex items-center gap-2 text-sm font-bold text-surface-muted hover:text-white transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            رجوع
          </button>
        </div>
      </PageHeader>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Profile Card */}
        <div className="lg:col-span-1 space-y-6">
          <div className="admin-card p-6 relative overflow-hidden">
            <div className="absolute top-0 right-0 w-32 h-32 bg-mazad-primary/5 rounded-full -mr-16 -mt-16 blur-2xl" />

            <div className="relative flex flex-col items-center text-center">
              <div className="w-24 h-24 rounded-3xl bg-mazad-primary/10 border-2 border-mazad-primary/20 flex items-center justify-center text-3xl text-mazad-primary font-bold mb-4">
                {(user.full_name ?? 'U')[0]}
              </div>

              <h2 className="text-xl font-display font-bold text-white mb-1">{user.full_name ?? 'مستخدم بدون اسم'}</h2>
              <p className="text-xs font-mono font-bold text-surface-muted mb-4 tracking-wider">{shortID(user.id)}</p>

              <div className="flex gap-2 mb-6">
                <StatusBadge status={user.is_verified ? 'verified' : 'unverified'} />
                {!user.is_active && <StatusBadge status="blocked" />}
              </div>

              {isEditing ? (
                // Edit Mode
                <div className="w-full space-y-4 pt-6 border-t border-surface-border">
                  <div>
                    <label className="text-xs font-bold text-surface-muted mb-1 block">الاسم الكامل</label>
                    <Input
                      value={editForm.full_name}
                      onChange={(e) => setEditForm({ ...editForm, full_name: e.target.value })}
                      placeholder="اسم المستخدم"
                      className="text-right"
                    />
                  </div>
                  <div>
                    <label className="text-xs font-bold text-surface-muted mb-1 block">البريد الإلكتروني</label>
                    <Input
                      type="email"
                      value={editForm.email}
                      onChange={(e) => setEditForm({ ...editForm, email: e.target.value })}
                      placeholder="email@example.com"
                      className="text-right"
                    />
                  </div>
                  <div>
                    <label className="text-xs font-bold text-surface-muted mb-1 block">المدينة</label>
                    <Input
                      value={editForm.city}
                      onChange={(e) => setEditForm({ ...editForm, city: e.target.value })}
                      placeholder="المدينة"
                      className="text-right"
                    />
                  </div>
                  <div className="flex gap-2 pt-2">
                    <Button 
                      onClick={saveProfile} 
                      disabled={updateProfile.isPending}
                      className="flex-1 bg-mazad-primary hover:bg-mazad-primary/90"
                    >
                      <Save className="w-4 h-4 ml-2" />
                      {updateProfile.isPending ? 'جاري الحفظ...' : 'حفظ'}
                    </Button>
                    <Button 
                      onClick={cancelEditing} 
                      variant="outline"
                      className="flex-1 border-surface-border"
                    >
                      <X className="w-4 h-4 ml-2" />
                      إلغاء
                    </Button>
                  </div>
                </div>
              ) : (
                // View Mode
                <div className="w-full space-y-3 pt-6 border-t border-surface-border text-right">
                  {/* Basic Info */}
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-lg bg-surface-base flex items-center justify-center shrink-0">
                      <User className="w-4 h-4 text-surface-muted" />
                    </div>
                    <div className="min-w-0">
                      <p className="text-[10px] font-bold text-surface-muted uppercase">الاسم الكامل</p>
                      <p className="text-sm font-medium text-white truncate">{user.full_name ?? 'غير متوفر'}</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-lg bg-surface-base flex items-center justify-center shrink-0">
                      <Mail className="w-4 h-4 text-surface-muted" />
                    </div>
                    <div className="min-w-0">
                      <p className="text-[10px] font-bold text-surface-muted uppercase">البريد الإلكتروني</p>
                      <p className="text-sm font-medium text-white truncate">{user.email ?? 'غير متوفر'}</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-lg bg-surface-base flex items-center justify-center shrink-0">
                      <Phone className="w-4 h-4 text-surface-muted" />
                    </div>
                    <div>
                      <p className="text-[10px] font-bold text-surface-muted uppercase">رقم الهاتف</p>
                      {/* Client feedback #17: authorized Admin sees the
                          complete phone number here, not masked. */}
                      <p dir="ltr" className="text-sm font-mono font-bold text-white">{formatFullPhone(user.phone, user.phone_e164)}</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-lg bg-surface-base flex items-center justify-center shrink-0">
                      <MapPin className="w-4 h-4 text-surface-muted" />
                    </div>
                    <div>
                      <p className="text-[10px] font-bold text-surface-muted uppercase">المدينة</p>
                      <p className="text-sm font-medium text-white">{user.city ?? 'غير محدد'}</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-lg bg-surface-base flex items-center justify-center shrink-0">
                      <Globe className="w-4 h-4 text-surface-muted" />
                    </div>
                    <div>
                      <p className="text-[10px] font-bold text-surface-muted uppercase">اللغة المفضلة</p>
                      <p className="text-sm font-medium text-white">
                        {user.language_pref === 'ar' ? 'العربية' : 
                         user.language_pref === 'fr' ? 'Français' : 
                         user.language_pref === 'en' ? 'English' : user.language_pref}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-lg bg-surface-base flex items-center justify-center shrink-0">
                      <Calendar className="w-4 h-4 text-surface-muted" />
                    </div>
                    <div>
                      <p className="text-[10px] font-bold text-surface-muted uppercase">تاريخ الانضمام</p>
                      <p className="text-sm font-medium text-white">{formatDate(user.created_at)}</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-lg bg-surface-base flex items-center justify-center shrink-0">
                      <CheckCircle className="w-4 h-4 text-surface-muted" />
                    </div>
                    <div>
                      <p className="text-[10px] font-bold text-surface-muted uppercase">آخر تسجيل دخول</p>
                      <p className="text-sm font-medium text-white">
                        {user.last_login_at ? formatDate(user.last_login_at) : 'غير متوفر'}
                      </p>
                    </div>
                  </div>

                  {/* PIN Section - Hidden by default */}
                  <div className="pt-4 border-t border-surface-border">
                    <div className="flex items-center justify-between mb-3">
                      <p className="text-xs font-bold text-surface-muted">كلمة المرور (PIN)</p>
                      <button
                        onClick={() => setShowPinModal(true)}
                        className="text-xs text-mazad-primary hover:text-white font-bold"
                      >
                        إعادة تعيين
                      </button>
                    </div>
                    <div className="flex items-center gap-2 bg-surface-base p-3 rounded-lg">
                      <Lock className="w-4 h-4 text-surface-muted" />
                      <input
                        type={showPin ? "text" : "password"}
                        value={showPin ? "••••" : "••••"}
                        readOnly
                        className="bg-transparent text-sm font-mono text-white flex-1 outline-none"
                      />
                      <button
                        onClick={() => setShowPin(!showPin)}
                        className="text-surface-muted hover:text-white transition-colors"
                        title={showPin ? "إخفاء PIN" : "إظهار PIN (غير متوفر للأمان)"}
                      >
                        {showPin ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                      </button>
                    </div>
                    <p className="text-[10px] text-surface-muted mt-1">
                      PIN مخزن بشكل آمن ولا يمكن عرضه. يمكنك إعادة تعيينه فقط.
                    </p>
                  </div>
                </div>
              )}

              {user.role !== 'admin' && (
                <button
                  onClick={() => setBlockConfirm(true)}
                  className={cn(
                    "w-full mt-8 py-3 rounded-xl font-bold flex items-center justify-center gap-2 transition-all border shadow-lg",
                    user.is_active
                      ? "bg-red-500/10 border-red-500/20 text-red-500 hover:bg-red-500 hover:text-white"
                      : "bg-emerald-500/10 border-emerald-500/20 text-emerald-500 hover:bg-emerald-500 hover:text-white"
                  )}
                >
                  {user.is_active ? <ShieldOff className="w-4 h-4" /> : <Shield className="w-4 h-4" />}
                  {user.is_active ? 'حظر حساب المستخدم' : 'إلغاء حظر الحساب'}
                </button>
              )}
            </div>
          </div>

          {/* MAZADPAY -- admin wallet controls: read-only balance display +
              disable/enable toggle. Add/deduct balance itself stays on
              TransactionDetailPage (anchored to a specific transaction);
              this card is for the user-anchored disable/enable action and
              at-a-glance wallet status. */}
          {wallet && (
            <div className="admin-card p-6">
              <h3 className="text-xs font-bold text-surface-muted uppercase tracking-widest mb-4 flex items-center gap-2">
                <WalletCards className="w-3.5 h-3.5 text-mazad-primary" />
                المحفظة
              </h3>
              <div className="space-y-3 mb-4">
                <div className="flex items-center justify-between">
                  <span className="text-xs text-surface-muted">الرصيد المتاح</span>
                  <span className="text-sm font-bold text-white">{formatPrice(wallet.balance, wallet.currency_code)}</span>
                </div>
                {parseFloat(wallet.frozen_amount) > 0 && (
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-surface-muted">رصيد مجمد</span>
                    <span className="text-sm font-bold text-orange-400">{formatPrice(wallet.frozen_amount, wallet.currency_code)}</span>
                  </div>
                )}
                <div className="flex items-center justify-between">
                  <span className="text-xs text-surface-muted">حالة المحفظة</span>
                  <StatusBadge status={wallet.is_disabled ? 'blocked' : 'verified'} />
                </div>
                {wallet.is_disabled && wallet.disabled_reason && (
                  <p className="text-[10px] text-red-400 bg-red-500/10 border border-red-500/20 rounded-lg p-2 mt-1">
                    السبب: {wallet.disabled_reason}
                  </p>
                )}
              </div>
              <button
                onClick={handleWalletDisableToggle}
                disabled={setWalletDisabled.isPending}
                className={cn(
                  "w-full py-3 rounded-xl font-bold flex items-center justify-center gap-2 transition-all border shadow-lg disabled:opacity-50",
                  wallet.is_disabled
                    ? "bg-emerald-500/10 border-emerald-500/20 text-emerald-500 hover:bg-emerald-500 hover:text-white"
                    : "bg-orange-500/10 border-orange-500/20 text-orange-500 hover:bg-orange-500 hover:text-white"
                )}
              >
                <Wallet className="w-4 h-4" />
                {wallet.is_disabled ? 'إعادة تفعيل المحفظة' : 'تعطيل/تجميد المحفظة'}
              </button>
            </div>
          )}
        </div>

        {/* History Tabs */}
        <div className="lg:col-span-2 space-y-6">
          <div className="flex gap-2">
            <button
              onClick={() => setActiveTab('auctions')}
              className={cn(
                "px-6 py-3 rounded-xl font-bold text-sm transition-all flex items-center gap-2",
                activeTab === 'auctions'
                  ? "bg-mazad-primary text-white shadow-lg shadow-mazad-primary/20"
                  : "bg-surface-card text-surface-muted hover:text-white border border-surface-border"
              )}
            >
              <Gavel className="w-4 h-4" />
              سجل المزادات
            </button>
            <button
              onClick={() => setActiveTab('transactions')}
              className={cn(
                "px-6 py-3 rounded-xl font-bold text-sm transition-all flex items-center gap-2",
                activeTab === 'transactions'
                  ? "bg-mazad-primary text-white shadow-lg shadow-mazad-primary/20"
                  : "bg-surface-card text-surface-muted hover:text-white border border-surface-border"
              )}
            >
              <CreditCard className="w-4 h-4" />
              سجل المعاملات
            </button>
          </div>

          <div className="min-h-[400px]">
            {activeTab === 'auctions' ? (
              <DataTable
                isLoading={historyLoading}
                data={history ?? []}
                columns={[
                  {
                    header: 'المزاد',
                    accessorKey: 'title',
                    cell: ({ getValue }) => <span className="font-bold text-white">{getValue<string>()}</span>,
                  },
                  {
                    header: 'السعر الحالي',
                    accessorKey: 'current_price',
                    cell: ({ getValue, row }) => <span className="text-mazad-accent font-bold">{formatPrice(getValue<string>(), (row.original as { currency_code?: string | null }).currency_code)}</span>,
                  },
                  {
                    header: 'الحالة',
                    accessorKey: 'status',
                    cell: ({ getValue }) => <StatusBadge status={getValue<string>()} />,
                  },
                  {
                    header: 'التاريخ',
                    accessorKey: 'created_at',
                    cell: ({ getValue }) => <span className="text-xs text-surface-muted">{formatDate(getValue<string>())}</span>,
                  },
                ]}
                emptyTitle="لا يوجد مزادات"
                emptyDescription="لم يشارك هذا المستخدم في أي مزادات بعد."
              />
            ) : (
              <DataTable
                isLoading={historyLoading}
                data={history ?? []}
                columns={[
                  {
                    header: 'النوع',
                    accessorKey: 'type',
                    cell: ({ getValue }) => {
                      const type = getValue<string>()
                      const labels: Record<string, string> = {
                        deposit: 'إيداع', withdraw: 'سحب',
                        bid_hold: 'حجز مزايدة', bid_refund: 'استرداد', payment: 'دفع'
                      }
                      return <span className="font-bold text-white">{labels[type] ?? type}</span>
                    },
                  },
                  {
                    header: 'المبلغ',
                    accessorKey: 'amount',
                    cell: ({ getValue, row }) => <span className="text-emerald-400 font-bold">{formatPrice(getValue<string>(), (row.original as { currency_code?: string | null }).currency_code)}</span>,
                  },
                  {
                    header: 'الحالة',
                    accessorKey: 'status',
                    cell: ({ getValue }) => <StatusBadge status={getValue<string>()} />,
                  },
                  {
                    header: 'التاريخ',
                    accessorKey: 'created_at',
                    cell: ({ getValue }) => <span className="text-xs text-surface-muted">{formatDate(getValue<string>())}</span>,
                  },
                ]}
                emptyTitle="لا يوجد معاملات"
                emptyDescription="لم يقم هذا المستخدم بأي معاملات مالية بعد."
              />
            )}
          </div>
        </div>
      </div>

      <ConfirmDialog
        open={blockConfirm}
        onOpenChange={setBlockConfirm}
        title={user.is_active ? `هل تود حظر ${user.full_name ?? 'هذا المستخدم'}؟` : `هل تود إلغاء حظر ${user.full_name ?? 'هذا المستخدم'}؟`}
        description={
          user.is_active
            ? 'لن يتمكن المستخدم من تسجيل الدخول أو المشاركة في المزادات بعد الحظر.'
            : 'سيتمكن المستخدم من الوصول لكافة ميزات التطبيق مرة أخرى.'
        }
        confirmLabel={user.is_active ? 'تأكيد الحظر' : 'تأكيد إلغاء الحظر'}
        variant={user.is_active ? 'danger' : 'success'}
        loading={blockUser.isPending}
        onConfirm={handleBlockToggle}
      />

      {/* MAZADPAY -- admin wallet controls: disable-wallet confirm dialog with
          an optional reason field, mirroring the PIN Reset Modal's pattern of
          a ConfirmDialog with a custom input-bearing description. */}
      <ConfirmDialog
        open={walletDisableConfirm}
        onOpenChange={(v) => { setWalletDisableConfirm(v); if (!v) setWalletDisableReason('') }}
        title={`هل تود تعطيل محفظة ${user.full_name ?? 'هذا المستخدم'}؟`}
        description={
          <div className="space-y-3 mt-2">
            <p>لن يتمكن المستخدم من المزايدة أو طلب سحب رصيد حتى تتم إعادة التفعيل. لا يؤثر هذا على رصيده الحالي.</p>
            <div>
              <label className="text-xs font-bold text-surface-muted mb-1 block">السبب (اختياري)</label>
              <Input
                value={walletDisableReason}
                onChange={(e) => setWalletDisableReason(e.target.value)}
                placeholder="اكتب سبب التعطيل..."
              />
            </div>
          </div>
        }
        confirmLabel="تأكيد التعطيل"
        variant="danger"
        loading={setWalletDisabled.isPending}
        onConfirm={confirmWalletDisable}
      />

      {/* PIN Reset Modal */}
      <ConfirmDialog
        open={showPinModal}
        onOpenChange={setShowPinModal}
        title="إعادة تعيين PIN"
        description={
          <div className="space-y-4 mt-4">
            <div>
              <label className="text-xs font-bold text-surface-muted mb-1 block">PIN جديد</label>
              <div className="flex items-center gap-2">
                <Input
                  type={showNewPin ? "text" : "password"}
                  value={pinForm.new_pin}
                  onChange={(e) => setPinForm({ ...pinForm, new_pin: e.target.value })}
                  placeholder="أدخل PIN جديد"
                  maxLength={6}
                  className="text-center tracking-widest"
                />
                <button
                  onClick={() => setShowNewPin(!showNewPin)}
                  className="p-2 text-surface-muted hover:text-white transition-colors"
                >
                  {showNewPin ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
              </div>
            </div>
            <div>
              <label className="text-xs font-bold text-surface-muted mb-1 block">تأكيد PIN</label>
              <Input
                type="password"
                value={pinForm.confirm_pin}
                onChange={(e) => setPinForm({ ...pinForm, confirm_pin: e.target.value })}
                placeholder="أكد PIN الجديد"
                maxLength={6}
                className="text-center tracking-widest"
              />
            </div>
          </div>
        }
        confirmLabel="تأكيد"
        variant="default"
        loading={false}
        onConfirm={handleResetPin}
      />
    </div>
  )
}
