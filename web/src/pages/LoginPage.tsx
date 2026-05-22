import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useMutation } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { Hammer, Phone, Lock, Loader2, Gavel, TrendingUp, Users, ShieldCheck } from 'lucide-react'
import { loginAdmin } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'

const schema = z.object({
  phone: z.string().min(8, 'رقم الهاتف غير صالح'),
  pin:   z.string().length(4, 'الرمز السري يجب أن يكون 4 أرقام').regex(/^\d+$/, 'أرقام فقط'),
})
type Form = z.infer<typeof schema>

const FEATURES = [
  { icon: Gavel,       label: 'إدارة المزادات',   color: '#3B82F6' },
  { icon: TrendingUp,  label: 'تحليلات متقدمة',   color: '#10B981' },
  { icon: Users,       label: 'إدارة المستخدمين', color: '#F59E0B' },
  { icon: ShieldCheck, label: 'أمان ومصداقية',    color: '#A78BFA' },
]

export function LoginPage() {
  const navigate = useNavigate()
  const setAuth  = useAuthStore((s) => s.setAuth)

  const { register, handleSubmit, formState: { errors } } = useForm<Form>({
    resolver: zodResolver(schema),
  })

  const { mutate, isPending } = useMutation({
    mutationFn: loginAdmin,
    onSuccess: ({ token, user }) => {
      setAuth(token, user)
      toast.success('تم تسجيل الدخول بنجاح')
      navigate('/', { replace: true })
    },
    onError: (err: Error) => toast.error(err.message),
  })

  return (
    <div
      className="min-h-screen flex"
      dir="rtl"
      style={{ background: '#07111F' }}
    >
      {/* ── Left brand panel ── */}
      <div
        className="hidden lg:flex lg:w-[420px] xl:w-[460px] shrink-0 flex-col justify-between p-10 relative overflow-hidden"
        style={{
          background: 'linear-gradient(160deg, #060F1C 0%, #07111F 60%, #081420 100%)',
          borderLeft: '1px solid rgba(255,255,255,0.05)',
        }}
      >
        {/* Dot grid background */}
        <div
          className="absolute inset-0 pointer-events-none dot-grid opacity-60"
        />

        {/* Blue glow orbs */}
        <div
          className="glow-blue w-64 h-64"
          style={{
            top: '-60px', left: '-60px',
            background: 'radial-gradient(circle, rgba(59,130,246,0.15) 0%, transparent 70%)',
          }}
        />
        <div
          className="glow-blue w-48 h-48"
          style={{
            bottom: '60px', right: '-40px',
            background: 'radial-gradient(circle, rgba(16,185,129,0.1) 0%, transparent 70%)',
            animationDelay: '1.5s',
          }}
        />

        {/* Logo */}
        <div className="relative flex items-center gap-3">
          <div
            className="w-11 h-11 rounded-2xl flex items-center justify-center"
            style={{
              background: 'linear-gradient(145deg, #60A5FA 0%, #2563EB 60%, #1D4ED8 100%)',
              boxShadow: '0 6px 24px rgba(59,130,246,0.45), 0 1px 0 rgba(255,255,255,0.2) inset',
            }}
          >
            <Hammer className="w-5 h-5 text-white" strokeWidth={2.5} />
          </div>
          <div>
            <p className="font-black text-[18px] text-white leading-none" style={{ fontFamily: 'Cairo' }}>
              MazadPay
            </p>
            <p className="text-[9px] font-bold uppercase tracking-widest mt-0.5" style={{ color: '#1E3A5F' }}>
              Admin Panel
            </p>
          </div>
        </div>

        {/* Center hero text */}
        <div className="relative space-y-5">
          <div>
            <h2
              className="text-[32px] font-black leading-[1.15] mb-3"
              style={{ color: '#E0EEFF', fontFamily: 'Cairo' }}
            >
              منصة إدارة
              <br />
              <span style={{
                background: 'linear-gradient(90deg, #60A5FA, #3B82F6)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
              }}>
                المزادات الإلكترونية
              </span>
            </h2>
            <p className="text-sm leading-relaxed font-medium" style={{ color: '#2D4A68' }}>
              لوحة تحكم احترافية لإدارة كل جوانب منصة MazadPay
            </p>
          </div>

          {/* Feature chips */}
          <div className="grid grid-cols-2 gap-2">
            {FEATURES.map(({ icon: Icon, label, color }) => (
              <div
                key={label}
                className="flex items-center gap-2 px-3 py-2.5 rounded-xl transition-all duration-200"
                style={{
                  background: 'rgba(255,255,255,0.03)',
                  border: '1px solid rgba(255,255,255,0.06)',
                }}
              >
                <div
                  className="w-7 h-7 rounded-lg flex items-center justify-center shrink-0"
                  style={{ background: `${color}18`, color }}
                >
                  <Icon className="w-3.5 h-3.5" strokeWidth={2} />
                </div>
                <span className="text-[11.5px] font-bold leading-tight" style={{ color: '#5A7A9A' }}>
                  {label}
                </span>
              </div>
            ))}
          </div>
        </div>

        {/* Footer */}
        <p className="relative text-[10px] font-medium" style={{ color: '#142233' }}>
          © 2026 MazadPay — جميع الحقوق محفوظة
        </p>
      </div>

      {/* ── Right: form ── */}
      <div className="flex-1 flex items-center justify-center p-6">
        <div className="w-full max-w-[360px] animate-slide-in">

          {/* Mobile logo */}
          <div className="lg:hidden flex flex-col items-center mb-8">
            <div
              className="w-12 h-12 rounded-2xl flex items-center justify-center mb-3"
              style={{
                background: 'linear-gradient(145deg, #60A5FA 0%, #2563EB 60%, #1D4ED8 100%)',
                boxShadow: '0 6px 20px rgba(59,130,246,0.4)',
              }}
            >
              <Hammer className="w-6 h-6 text-white" strokeWidth={2.5} />
            </div>
            <h1 className="font-black text-2xl text-white" style={{ fontFamily: 'Cairo' }}>MazadPay</h1>
          </div>

          {/* Card */}
          <div
            className="rounded-2xl p-8"
            style={{
              background: 'linear-gradient(145deg, #0D1B2E 0%, #0A1525 100%)',
              border: '1px solid rgba(255,255,255,0.08)',
              boxShadow: '0 24px 64px rgba(0,0,0,0.5), inset 0 1px 0 rgba(255,255,255,0.06)',
            }}
          >
            {/* Top glow accent */}
            <div
              className="absolute top-0 left-1/2 -translate-x-1/2 w-32 h-px pointer-events-none"
              style={{ background: 'linear-gradient(90deg, transparent, rgba(59,130,246,0.5), transparent)' }}
            />

            <div className="mb-7">
              <h1
                className="text-xl font-black mb-1"
                style={{ color: '#E2E8F0', fontFamily: 'Cairo' }}
              >
                تسجيل الدخول
              </h1>
              <p className="text-sm font-medium" style={{ color: '#3D5A78' }}>
                لوحة التحكم للمسؤولين
              </p>
            </div>

            <form onSubmit={handleSubmit((d) => mutate(d))} className="space-y-4">
              {/* Phone */}
              <div>
                <label className="text-[11px] font-bold block mb-1.5" style={{ color: '#4A6080' }}>
                  رقم الهاتف
                </label>
                <div className="relative">
                  <Phone
                    className="absolute right-3 top-1/2 -translate-y-1/2 w-[14px] h-[14px] pointer-events-none"
                    style={{ color: '#2D4560' }}
                  />
                  <input
                    {...register('phone')}
                    placeholder="+22247601175"
                    className="input-base pr-9"
                  />
                </div>
                {errors.phone && (
                  <p className="text-[11px] font-semibold mt-1.5" style={{ color: '#FC8181' }}>
                    {errors.phone.message}
                  </p>
                )}
              </div>

              {/* PIN */}
              <div>
                <label className="text-[11px] font-bold block mb-1.5" style={{ color: '#4A6080' }}>
                  الرمز السري (4 أرقام)
                </label>
                <div className="relative">
                  <Lock
                    className="absolute right-3 top-1/2 -translate-y-1/2 w-[14px] h-[14px] pointer-events-none"
                    style={{ color: '#2D4560' }}
                  />
                  <input
                    {...register('pin')}
                    type="password"
                    maxLength={4}
                    placeholder="••••"
                    className="input-base pr-9 tracking-widest"
                  />
                </div>
                {errors.pin && (
                  <p className="text-[11px] font-semibold mt-1.5" style={{ color: '#FC8181' }}>
                    {errors.pin.message}
                  </p>
                )}
              </div>

              {/* Submit */}
              <button
                type="submit"
                disabled={isPending}
                className="btn btn-primary w-full justify-center py-3 mt-3"
                style={{ borderRadius: '10px', fontSize: '14px' }}
              >
                {isPending && <Loader2 className="w-4 h-4 animate-spin" />}
                {isPending ? 'جاري الاتصال...' : 'تسجيل الدخول'}
              </button>
            </form>
          </div>

          <p className="text-center text-[11px] mt-5 font-medium" style={{ color: '#1E3A5F' }}>
            الدخول مخصص لمسؤولي MazadPay فقط
          </p>
        </div>
      </div>
    </div>
  )
}
