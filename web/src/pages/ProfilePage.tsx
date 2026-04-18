import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { useMe, useUpdateProfile, useChangePin } from '@/hooks/useUsers'
import { User, Shield, Phone, Mail, MapPin, Loader2 } from 'lucide-react'

export function ProfilePage() {
  const { data: user, isLoading } = useMe()
  const { mutate: updateProfile, isPending: isUpdating } = useUpdateProfile()
  const { mutate: changePin, isPending: isChangingPin } = useChangePin()

  const [formData, setFormData] = useState({
    full_name: '',
    email: '',
    city: '',
  })

  const [pinData, setPinData] = useState({
    old_pin: '',
    new_pin: '',
    confirm_pin: '',
  })

  // Set initial data when user is loaded
  useEffect(() => {
    if (user) {
      setFormData({
        full_name: user.full_name || '',
        email: user.email || '',
        city: user.city || '',
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
    if (pinData.new_pin !== pinData.confirm_pin) return
    changePin({ old_pin: pinData.old_pin, new_pin: pinData.new_pin })
    setPinData({ old_pin: '', new_pin: '', confirm_pin: '' })
  }

  return (
    <div className="space-y-6 animate-in fade-in duration-500" dir="rtl">
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-bold">الملف الشخصي</h1>
        <p className="text-surface-muted text-sm">إدارة معلومات حسابك والأمان</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* User Card */}
        <Card className="lg:col-span-1 h-fit">
          <CardHeader className="text-center">
             <div className="w-24 h-24 bg-mazad-primary/10 rounded-full flex items-center justify-center mx-auto mb-4 border border-mazad-primary/20">
                <User className="w-12 h-12 text-mazad-primary" />
             </div>
             <CardTitle>{user?.full_name || 'مشرف النظام'}</CardTitle>
             <div className="flex justify-center gap-2 mt-2">
                <Badge variant="secondary" className="capitalize">{user?.role === 'admin' ? 'مدير' : user?.role}</Badge>
                <Badge variant={user?.is_verified ? "default" : "outline"}>
                   {user?.is_verified ? 'حساب موثق' : 'غير موثق'}
                </Badge>
             </div>
          </CardHeader>
          <CardContent className="space-y-4">
             <div className="flex items-center gap-3 text-sm">
                <Phone className="w-4 h-4 text-surface-muted" />
                <span className="font-mono">{user?.phone}</span>
             </div>
             <div className="flex items-center gap-3 text-sm">
                <Mail className="w-4 h-4 text-surface-muted" />
                <span>{user?.email || 'لا يوجد بريد'}</span>
             </div>
             <div className="flex items-center gap-3 text-sm">
                <MapPin className="w-4 h-4 text-surface-muted" />
                <span>{user?.city || 'غير محدد'}</span>
             </div>
             <div className="pt-4 mt-4 border-t border-surface-border text-xs text-surface-muted text-center">
                تاريخ الإنضمام: {user?.created_at ? new Date(user.created_at).toLocaleDateString('ar-EG') : '—'}
             </div>
          </CardContent>
        </Card>

        {/* Editing Tabs */}
        <div className="lg:col-span-2 space-y-6">
          <Card>
            <CardHeader className="border-b border-surface-border">
              <CardTitle className="flex items-center gap-2">
                <User className="w-5 h-5 text-mazad-primary" />
                المعلومات الشخصية
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-6">
              <form onSubmit={handleUpdateProfile} className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="full_name">الاسم الكامل</Label>
                    <Input 
                      id="full_name" 
                      value={formData.full_name} 
                      onChange={e => setFormData({...formData, full_name: e.target.value})}
                      placeholder="أدخل اسمك الكامل"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="email">البريد الإلكتروني</Label>
                    <Input 
                      id="email" 
                      type="email"
                      value={formData.email} 
                      onChange={e => setFormData({...formData, email: e.target.value})}
                      placeholder="admin@mazadpay.com"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="city">المدينة</Label>
                    <Input 
                      id="city" 
                      value={formData.city} 
                      onChange={e => setFormData({...formData, city: e.target.value})}
                      placeholder="نواكشوط"
                    />
                  </div>
                </div>
                <div className="flex justify-start">
                  <Button type="submit" disabled={isUpdating}>
                    {isUpdating && <Loader2 className="w-4 h-4 ml-2 animate-spin" />}
                    حفظ التغييرات
                  </Button>
                </div>
              </form>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="border-b border-surface-border">
              <CardTitle className="flex items-center gap-2 text-red-500">
                <Shield className="w-5 h-5" />
                الأمان (تغيير رمز PIN)
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-6">
              <form onSubmit={handleUpdatePin} className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="old_pin">رمز PIN الحالي</Label>
                    <Input 
                      id="old_pin" 
                      type="password"
                      maxLength={4}
                      value={pinData.old_pin} 
                      onChange={e => setPinData({...pinData, old_pin: e.target.value})}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="new_pin">الرمز الجديد</Label>
                    <Input 
                      id="new_pin" 
                      type="password"
                      maxLength={4}
                      value={pinData.new_pin} 
                      onChange={e => setPinData({...pinData, new_pin: e.target.value})}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="confirm_pin">تأكيد الرمز</Label>
                    <Input 
                      id="confirm_pin" 
                      type="password"
                      maxLength={4}
                      value={pinData.confirm_pin} 
                      onChange={e => setPinData({...pinData, confirm_pin: e.target.value})}
                    />
                  </div>
                </div>
                <div className="flex justify-start">
                  <Button variant="destructive" type="submit" disabled={isChangingPin || !pinData.new_pin || pinData.new_pin !== pinData.confirm_pin}>
                    {isChangingPin && <Loader2 className="w-4 h-4 ml-2 animate-spin" />}
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
