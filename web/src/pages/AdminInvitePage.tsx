import { useState } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useCreateAdmin } from '@/hooks/useUsers'
import { ShieldCheck, User, Phone, Mail, Lock, Loader2 } from 'lucide-react'

export function AdminInvitePage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const token = searchParams.get('token')
  const createAdmin = useCreateAdmin()

  const [formData, setFormData] = useState({
    full_name: '',
    phone: '',
    email: '',
    pin: '',
    confirm_pin: '',
  })

  if (!token) {
    return (
      <div className="min-h-screen bg-surface-base flex items-center justify-center p-4">
        <Card className="max-w-md w-full text-center">
          <CardHeader>
            <div className="w-16 h-16 bg-red-500/10 rounded-full flex items-center justify-center mx-auto mb-4">
               <ShieldCheck className="w-8 h-8 text-red-500" />
            </div>
            <CardTitle className="text-red-500">رابط دعوة غير صالح</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-surface-muted">عذراً، هذا الرابط غير صالح أو انتهت صلاحيته.</p>
            <Button className="mt-6 w-full" onClick={() => navigate('/')}>
              العودة للرئيسية
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (formData.pin !== formData.confirm_pin) return
    
    createAdmin.mutate({
      token,
      phone: formData.phone,
      full_name: formData.full_name,
      email: formData.email,
      pin: formData.pin
    }, {
      onSuccess: () => {
        setTimeout(() => navigate('/'), 2000)
      }
    })
  }

  return (
    <div className="min-h-screen bg-surface-base flex items-center justify-center p-4" dir="rtl">
      <div className="w-full max-w-lg space-y-8 animate-in fade-in zoom-in-95 duration-500">
        <div className="text-center">
          <div className="w-20 h-20 bg-mazad-primary rounded-2xl flex items-center justify-center mx-auto mb-6 shadow-xl shadow-mazad-primary/20 rotate-3">
            <ShieldCheck className="w-10 h-10 text-white" />
          </div>
          <h1 className="text-3xl font-display font-bold text-white tracking-tight">دعوة للانضمام للمشرفين</h1>
          <p className="text-surface-muted mt-2">يرجى إكمال معلوماتك لتفعيل حسابك كمسؤول في MazadPay</p>
        </div>

        <Card className="border-mazad-primary/20 bg-surface-card/80 backdrop-blur-xl">
          <CardContent className="pt-8">
            <form onSubmit={handleSubmit} className="space-y-5">
              <div className="space-y-2">
                <Label className="flex items-center gap-2">
                  <User className="w-3.5 h-3.5" /> الاسم الكامل
                </Label>
                <Input 
                  required
                  value={formData.full_name}
                  onChange={e => setFormData({...formData, full_name: e.target.value})}
                  placeholder="محمد الأمين"
                  className="bg-surface-base/50"
                />
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                <div className="space-y-2">
                  <Label className="flex items-center gap-2">
                    <Phone className="w-3.5 h-3.5" /> رقم الهاتف
                  </Label>
                  <Input 
                    required
                    value={formData.phone}
                    onChange={e => setFormData({...formData, phone: e.target.value})}
                    placeholder="222XXXXXXXX"
                    className="bg-surface-base/50"
                  />
                </div>
                <div className="space-y-2">
                  <Label className="flex items-center gap-2">
                    <Mail className="w-3.5 h-3.5" /> البريد الإلكتروني
                  </Label>
                  <Input 
                    required
                    type="email"
                    value={formData.email}
                    onChange={e => setFormData({...formData, email: e.target.value})}
                    placeholder="admin@mazadpay.com"
                    className="bg-surface-base/50"
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                <div className="space-y-2">
                  <Label className="flex items-center gap-2">
                    <Lock className="w-3.5 h-3.5" /> رمز PIN الجديد
                  </Label>
                  <Input 
                    required
                    type="password"
                    maxLength={4}
                    value={formData.pin}
                    onChange={e => setFormData({...formData, pin: e.target.value})}
                    placeholder="****"
                    className="bg-surface-base/50"
                  />
                </div>
                <div className="space-y-2">
                  <Label className="flex items-center gap-2">
                    <Lock className="w-3.5 h-3.5" /> تأكيد رمز PIN
                  </Label>
                  <Input 
                    required
                    type="password"
                    maxLength={4}
                    value={formData.confirm_pin}
                    onChange={e => setFormData({...formData, confirm_pin: e.target.value})}
                    placeholder="****"
                    className="bg-surface-base/50 shadow-sm"
                  />
                </div>
              </div>

              <Button 
                type="submit" 
                className="w-full h-12 text-base font-bold bg-mazad-primary hover:bg-mazad-primary/90 mt-4"
                disabled={createAdmin.isPending || (formData.pin !== formData.confirm_pin)}
              >
                {createAdmin.isPending ? (
                  <>
                    <Loader2 className="w-5 h-5 ml-2 animate-spin" />
                    جاري تفعيل الحساب...
                  </>
                ) : 'تفعيل حساب المسؤول'}
              </Button>
            </form>
          </CardContent>
        </Card>
        
        <p className="text-center text-[10px] text-surface-muted uppercase tracking-widest font-bold opacity-50">
          MazadPay Administration Panel Security Protocol
        </p>
      </div>
    </div>
  )
}
