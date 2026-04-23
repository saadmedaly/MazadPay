import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { useMe, useUpdateProfile, useChangePin } from '@/hooks/useUsers'
import { 
  User, Shield, Phone, Mail, MapPin, Loader2, Eye, EyeOff, 
  Lock, Globe, Calendar, CheckCircle, Save, Pencil 
} from 'lucide-react'
import { formatDate } from '@/lib/formatters'
import { toast } from 'sonner'

export function ProfilePage() {
  const { data: user, isLoading } = useMe()
  const { mutate: updateProfile, isPending: isUpdating } = useUpdateProfile()
  const { mutate: changePin, isPending: isChangingPin } = useChangePin()

  const [formData, setFormData] = useState({
    full_name: '',
    email: '',
    city: '',
    country_code: '',
    address: '',
    postal_code: '',
    date_of_birth: '',
    gender: '',
  })

  const [pinData, setPinData] = useState({
    old_pin: '',
    new_pin: '',
    confirm_pin: '',
  })
  const [showOldPin, setShowOldPin] = useState(false)
  const [showNewPin, setShowNewPin] = useState(false)
  const [showConfirmPin, setShowConfirmPin] = useState(false)

  // Set initial data when user is loaded
  useEffect(() => {
    if (user) {
      setFormData({
        full_name: user.full_name || '',
        email: user.email || '',
        city: user.city || '',
        country_code: user.country_code || '',
        address: user.address || '',
        postal_code: user.postal_code || '',
        date_of_birth: user.date_of_birth || '',
        gender: user.gender || '',
      })
    }
  }, [user])

  if (isLoading) return <div className="flex justify-center p-20"><Loader2 className="animate-spin text-mazad-primary" /></div>

  const handleUpdateProfile = (e: React.FormEvent) => {
    e.preventDefault()
    updateProfile(formData)
  }

  const handleUpdatePin = (e: React.FormEvent) => {
    e.preventDefault()
    if (pinData.new_pin !== pinData.confirm_pin) {
      toast.error('PIN غير متطابق')
      return
    }
    if (pinData.new_pin.length < 4) {
      toast.error('يجب أن يكون PIN مكون من 4 أرقام على الأقل')
      return
    }
    changePin({ old_pin: pinData.old_pin, new_pin: pinData.new_pin }, {
      onSuccess: () => {
        toast.success('تم تغيير PIN بنجاح')
        setPinData({ old_pin: '', new_pin: '', confirm_pin: '' })
      },
      onError: (err: any) => {
        toast.error(err?.response?.data?.message || 'فشل تغيير PIN')
      }
    })
  }

  return (
    <div className="space-y-6 animate-in fade-in duration-500" dir="rtl">
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-bold">الملف الشخصي</h1>
        <p className="text-surface-muted text-sm">إدارة معلومات حسابك والأمان</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* User Info Card */}
        <Card className="lg:col-span-1 h-fit bg-surface-card border-surface-border">
          <CardHeader className="text-center relative overflow-hidden">
            <div className="absolute top-0 right-0 w-32 h-32 bg-mazad-primary/5 rounded-full -mr-16 -mt-16 blur-2xl" />
            <div className="relative">
              <div className="w-24 h-24 rounded-3xl bg-mazad-primary/10 border-2 border-mazad-primary/20 flex items-center justify-center text-3xl text-mazad-primary font-bold mx-auto mb-4">
                {(user?.full_name ?? 'U')[0]}
              </div>
              <CardTitle className="text-xl font-display">{user?.full_name || 'مشرف النظام'}</CardTitle>
              <div className="flex justify-center gap-2 mt-2">
                <Badge variant="secondary" className="capitalize">{user?.role === 'admin' ? 'مدير' : user?.role}</Badge>
                <Badge variant={user?.is_verified ? "default" : "outline"}>
                  {user?.is_verified ? 'حساب موثق' : 'غير موثق'}
                </Badge>
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Full Name */}
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-surface-base flex items-center justify-center shrink-0">
                <User className="w-4 h-4 text-surface-muted" />
              </div>
              <div className="min-w-0">
                <p className="text-[10px] font-bold text-surface-muted uppercase">الاسم الكامل</p>
                <p className="text-sm font-medium text-white truncate">{user?.full_name || 'غير متوفر'}</p>
              </div>
            </div>

            {/* Phone */}
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-surface-base flex items-center justify-center shrink-0">
                <Phone className="w-4 h-4 text-surface-muted" />
              </div>
              <div>
                <p className="text-[10px] font-bold text-surface-muted uppercase">رقم الهاتف</p>
                <p className="text-sm font-mono font-bold text-white">{user?.phone}</p>
              </div>
            </div>

            {/* Email */}
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-surface-base flex items-center justify-center shrink-0">
                <Mail className="w-4 h-4 text-surface-muted" />
              </div>
              <div className="min-w-0">
                <p className="text-[10px] font-bold text-surface-muted uppercase">البريد الإلكتروني</p>
                <p className="text-sm font-medium text-white truncate">{user?.email || 'غير متوفر'}</p>
              </div>
            </div>

            {/* City */}
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-surface-base flex items-center justify-center shrink-0">
                <MapPin className="w-4 h-4 text-surface-muted" />
              </div>
              <div>
                <p className="text-[10px] font-bold text-surface-muted uppercase">المدينة</p>
                <p className="text-sm font-medium text-white">{user?.city || 'غير محدد'}</p>
              </div>
            </div>

            {/* Address */}
            {user?.address && (
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-lg bg-surface-base flex items-center justify-center shrink-0">
                  <Mail className="w-4 h-4 text-surface-muted" />
                </div>
                <div className="min-w-0">
                  <p className="text-[10px] font-bold text-surface-muted uppercase">العنوان</p>
                  <p className="text-sm font-medium text-white truncate">{user.address}</p>
                </div>
              </div>
            )}

            {/* Date of Birth */}
            {user?.date_of_birth && (
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-lg bg-surface-base flex items-center justify-center shrink-0">
                  <Calendar className="w-4 h-4 text-surface-muted" />
                </div>
                <div>
                  <p className="text-[10px] font-bold text-surface-muted uppercase">تاريخ الميلاد</p>
                  <p className="text-sm font-medium text-white">{formatDate(user.date_of_birth)}</p>
                </div>
              </div>
            )}

            {/* Gender */}
            {user?.gender && (
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-lg bg-surface-base flex items-center justify-center shrink-0">
                  <User className="w-4 h-4 text-surface-muted" />
                </div>
                <div>
                  <p className="text-[10px] font-bold text-surface-muted uppercase">الجنس</p>
                  <p className="text-sm font-medium text-white">
                    {user.gender === 'male' ? 'ذكر' : user.gender === 'female' ? 'أنثى' : user.gender}
                  </p>
                </div>
              </div>
            )}

            {/* Language */}
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-surface-base flex items-center justify-center shrink-0">
                <Globe className="w-4 h-4 text-surface-muted" />
              </div>
              <div>
                <p className="text-[10px] font-bold text-surface-muted uppercase">اللغة المفضلة</p>
                <p className="text-sm font-medium text-white">
                  {user?.language_pref === 'ar' ? 'العربية' : 
                   user?.language_pref === 'fr' ? 'Français' : 
                   user?.language_pref === 'en' ? 'English' : user?.language_pref}
                </p>
              </div>
            </div>

            {/* Join Date */}
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-surface-base flex items-center justify-center shrink-0">
                <Calendar className="w-4 h-4 text-surface-muted" />
              </div>
              <div>
                <p className="text-[10px] font-bold text-surface-muted uppercase">تاريخ الانضمام</p>
                <p className="text-sm font-medium text-white">{user?.created_at ? formatDate(user.created_at) : '—'}</p>
              </div>
            </div>

            {/* Last Login */}
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-surface-base flex items-center justify-center shrink-0">
                <CheckCircle className="w-4 h-4 text-surface-muted" />
              </div>
              <div>
                <p className="text-[10px] font-bold text-surface-muted uppercase">آخر تسجيل دخول</p>
                <p className="text-sm font-medium text-white">
                  {user?.last_login_at ? formatDate(user.last_login_at) : 'غير متوفر'}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Editing Tabs */}
        <div className="lg:col-span-2 space-y-6">
          <Card className="bg-surface-card border-surface-border">
            <CardHeader className="border-b border-surface-border">
              <CardTitle className="flex items-center gap-2 text-white">
                <Pencil className="w-5 h-5 text-mazad-primary" />
                تعديل المعلومات الشخصية
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-6">
              <form onSubmit={handleUpdateProfile} className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="full_name" className="text-surface-muted">الاسم الكامل</Label>
                    <Input 
                      id="full_name" 
                      value={formData.full_name} 
                      onChange={e => setFormData({...formData, full_name: e.target.value})}
                      placeholder="أدخل اسمك الكامل"
                      className="text-right bg-surface-base border-surface-border"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="email" className="text-surface-muted">البريد الإلكتروني</Label>
                    <Input 
                      id="email" 
                      type="email"
                      value={formData.email} 
                      onChange={e => setFormData({...formData, email: e.target.value})}
                      placeholder="admin@mazadpay.com"
                      className="text-right bg-surface-base border-surface-border"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="city" className="text-surface-muted">المدينة</Label>
                    <Input 
                      id="city" 
                      value={formData.city} 
                      onChange={e => setFormData({...formData, city: e.target.value})}
                      placeholder="نواكشوط"
                      className="text-right bg-surface-base border-surface-border"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="country_code" className="text-surface-muted">رمز الدولة</Label>
                    <Input 
                      id="country_code" 
                      value={formData.country_code} 
                      onChange={e => setFormData({...formData, country_code: e.target.value})}
                      placeholder="MR"
                      className="text-right bg-surface-base border-surface-border"
                    />
                  </div>
                  <div className="space-y-2 md:col-span-2">
                    <Label htmlFor="address" className="text-surface-muted">العنوان</Label>
                    <Input 
                      id="address" 
                      value={formData.address} 
                      onChange={e => setFormData({...formData, address: e.target.value})}
                      placeholder="عنوانك الكامل"
                      className="text-right bg-surface-base border-surface-border"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="postal_code" className="text-surface-muted">الرمز البريدي</Label>
                    <Input 
                      id="postal_code" 
                      value={formData.postal_code} 
                      onChange={e => setFormData({...formData, postal_code: e.target.value})}
                      placeholder="00000"
                      className="text-right bg-surface-base border-surface-border"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="date_of_birth" className="text-surface-muted">تاريخ الميلاد</Label>
                    <Input 
                      id="date_of_birth" 
                      type="date"
                      value={formData.date_of_birth} 
                      onChange={e => setFormData({...formData, date_of_birth: e.target.value})}
                      className="text-right bg-surface-base border-surface-border"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="gender" className="text-surface-muted">الجنس</Label>
                    <select
                      id="gender"
                      value={formData.gender}
                      onChange={e => setFormData({...formData, gender: e.target.value})}
                      className="w-full h-10 px-3 rounded-md bg-surface-base border border-surface-border text-white text-right"
                    >
                      <option value="">اختر</option>
                      <option value="male">ذكر</option>
                      <option value="female">أنثى</option>
                    </select>
                  </div>
                </div>
                <div className="flex justify-start pt-2">
                  <Button 
                    type="submit" 
                    disabled={isUpdating}
                    className="bg-mazad-primary hover:bg-mazad-primary/90"
                  >
                    {isUpdating && <Loader2 className="w-4 h-4 ml-2 animate-spin" />}
                    <Save className="w-4 h-4 ml-2" />
                    حفظ التغييرات
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>

          <Card className="bg-surface-card border-surface-border">
            <CardHeader className="border-b border-surface-border">
              <CardTitle className="flex items-center gap-2 text-red-500">
                <Lock className="w-5 h-5" />
                الأمان (تغيير رمز PIN)
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-6">
              <form onSubmit={handleUpdatePin} className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  {/* Old PIN */}
                  <div className="space-y-2">
                    <Label htmlFor="old_pin" className="text-surface-muted">رمز PIN الحالي</Label>
                    <div className="relative">
                      <Input 
                        id="old_pin" 
                        type={showOldPin ? "text" : "password"}
                        maxLength={6}
                        value={pinData.old_pin} 
                        onChange={e => setPinData({...pinData, old_pin: e.target.value})}
                        className="text-center tracking-widest bg-surface-base border-surface-border pr-10"
                      />
                      <button
                        type="button"
                        onClick={() => setShowOldPin(!showOldPin)}
                        className="absolute left-2 top-1/2 -translate-y-1/2 text-surface-muted hover:text-white transition-colors"
                      >
                        {showOldPin ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                      </button>
                    </div>
                  </div>
                  
                  {/* New PIN */}
                  <div className="space-y-2">
                    <Label htmlFor="new_pin" className="text-surface-muted">الرمز الجديد</Label>
                    <div className="relative">
                      <Input 
                        id="new_pin" 
                        type={showNewPin ? "text" : "password"}
                        maxLength={6}
                        value={pinData.new_pin} 
                        onChange={e => setPinData({...pinData, new_pin: e.target.value})}
                        className="text-center tracking-widest bg-surface-base border-surface-border pr-10"
                      />
                      <button
                        type="button"
                        onClick={() => setShowNewPin(!showNewPin)}
                        className="absolute left-2 top-1/2 -translate-y-1/2 text-surface-muted hover:text-white transition-colors"
                      >
                        {showNewPin ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                      </button>
                    </div>
                  </div>
                  
                  {/* Confirm PIN */}
                  <div className="space-y-2">
                    <Label htmlFor="confirm_pin" className="text-surface-muted">تأكيد الرمز</Label>
                    <div className="relative">
                      <Input 
                        id="confirm_pin" 
                        type={showConfirmPin ? "text" : "password"}
                        maxLength={6}
                        value={pinData.confirm_pin} 
                        onChange={e => setPinData({...pinData, confirm_pin: e.target.value})}
                        className="text-center tracking-widest bg-surface-base border-surface-border pr-10"
                      />
                      <button
                        type="button"
                        onClick={() => setShowConfirmPin(!showConfirmPin)}
                        className="absolute left-2 top-1/2 -translate-y-1/2 text-surface-muted hover:text-white transition-colors"
                      >
                        {showConfirmPin ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                      </button>
                    </div>
                  </div>
                </div>
                <div className="flex justify-start pt-2">
                  <Button 
                    type="submit" 
                    disabled={isChangingPin || !pinData.new_pin || pinData.new_pin !== pinData.confirm_pin}
                    className="bg-red-500 hover:bg-red-600 text-white"
                  >
                    {isChangingPin && <Loader2 className="w-4 h-4 ml-2 animate-spin" />}
                    <Shield className="w-4 h-4 ml-2" />
                    تحديث رمز PIN
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
