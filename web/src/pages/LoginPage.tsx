import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useMutation } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { Hammer, Phone, Lock, Loader2 } from 'lucide-react'
import { loginAdmin } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'

const schema = z.object({
  phone: z.string().min(8, 'رقم الهاتف غير صالح'),
  pin:   z.string().length(4, 'الرمز السري يجب أن يكون 4 أرقام').regex(/^\d+$/, 'أرقام فقط'),
})

type Form = z.infer<typeof schema>

export function LoginPage() {
  const navigate = useNavigate()
  const setAuth = useAuthStore((s) => s.setAuth)

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
    <div className="min-h-screen bg-surface-base flex items-center justify-center p-4" dir="rtl">
      {/* Background subtle grid */}
      <div className="absolute inset-0 bg-[linear-gradient(rgba(42,45,62,0.4)_1px,transparent_1px),linear-gradient(90deg,rgba(42,45,62,0.4)_1px,transparent_1px)]
                      bg-[size:48px_48px] pointer-events-none" />

      <div className="relative w-full max-w-md animate-slide-in">
        {/* Card */}
        <div className="admin-card p-10 overflow-hidden relative">
          <div className="absolute top-0 right-0 w-24 h-24 bg-mazad-primary/5 rounded-full -mr-12 -mt-12 blur-3xl" />
          
          {/* Logo */}
          <div className="flex flex-col items-center mb-10 relative">
            <div className="w-14 h-14 rounded-2xl bg-mazad-primary flex items-center justify-center mb-4 shadow-xl shadow-mazad-primary/30">
              <Hammer className="w-7 h-7 text-white" />
            </div>
            <h1 className="font-display text-3xl font-bold text-white tracking-tight">MazadPay</h1>
            <p className="text-sm text-surface-muted mt-2 font-medium">لوحة التحكم للمسؤولين</p>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit((data) => mutate(data))} className="space-y-6">
            <div>
              <label className="text-[11px] font-bold text-surface-muted uppercase tracking-widest block mb-2 px-1">
                رقم الهاتف
              </label>
              <div className="relative group">
                <Phone className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted group-focus-within:text-mazad-primary transition-colors" />
                <input
                  {...register('phone')}
                  placeholder="+22247601175"
                  className="w-full bg-surface-base border border-surface-border rounded-xl
                             pr-10 pl-3 py-3 text-sm text-white placeholder:text-surface-muted/50
                             focus:outline-none focus:border-mazad-primary/60 focus:ring-4 focus:ring-mazad-primary/10
                             transition-all"
                />
              </div>
              {errors.phone && <p className="text-xs text-red-400 mt-2 font-medium">{errors.phone.message}</p>}
            </div>

            <div>
              <label className="text-[11px] font-bold text-surface-muted uppercase tracking-widest block mb-2 px-1">
                الرمز السري (4 أرقام)
              </label>
              <div className="relative group">
                <Lock className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted group-focus-within:text-mazad-primary transition-colors" />
                <input
                  {...register('pin')}
                  type="password"
                  maxLength={4}
                  placeholder="••••"
                  className="w-full bg-surface-base border border-surface-border rounded-xl
                             pr-10 pl-3 py-3 text-sm text-white placeholder:text-surface-muted/50
                             focus:outline-none focus:border-mazad-primary/60 focus:ring-4 focus:ring-mazad-primary/10
                             tracking-widest transition-all"
                />
              </div>
              {errors.pin && <p className="text-xs text-red-400 mt-2 font-medium">{errors.pin.message}</p>}
            </div>

            <button
              type="submit"
              disabled={isPending}
              className="w-full bg-mazad-primary hover:bg-mazad-primary-dk disabled:opacity-50
                         text-white font-bold py-3.5 rounded-xl text-sm transition-all
                         shadow-lg shadow-mazad-primary/20 flex items-center justify-center gap-3 mt-4
                         hover:translate-y-[-1px] active:translate-y-[0px]"
            >
              {isPending && <Loader2 className="w-4 h-4 animate-spin" />}
              {isPending ? 'جاري الاتصال...' : 'تسجيل الدخول'}
            </button>
          </form>
        </div>

        <p className="text-center text-xs text-surface-muted mt-8 font-medium">
          الدخول مخصص لمسؤولي MazadPay فقط
        </p>
      </div>
    </div>
  )
}
