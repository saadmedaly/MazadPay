import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Search, Eye, Shield, ShieldOff, Users, UserPlus, X, Copy, Link as LinkIcon, Check } from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { StatusBadge } from '@/components/shared/StatusBadge'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { DataTable } from '@/components/shared/DataTable'
import { Input } from '@/components/ui/input'
import { formatDate, maskPhone, shortID } from '@/lib/formatters'
import type { AdminUser } from '@/types/api'
import type { ColumnDef } from '@tanstack/react-table'
import { useUsers, useBlockUser, useGenerateInvitation } from '@/hooks/useUsers'
import { toast } from 'sonner'

export function UsersPage() {
  const navigate = useNavigate()
  const [q, setQ] = useState('')
  const [page, setPage] = useState(1)
  const [blockTarget, setBlockTarget] = useState<{ id: string; block: boolean; name: string } | null>(null)
  const [showInviteModal, setShowInviteModal] = useState(false)
  const [token, setToken] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)

  const { data, isLoading, isError } = useUsers(q, page)
  const blockUser = useBlockUser()
  const generateInvitation = useGenerateInvitation()

  const handleGenerateLink = () => {
    generateInvitation.mutate(undefined, {
      onSuccess: (newToken) => {
        setToken(newToken)
      }
    })
  }

  const copyLink = () => {
    if (!token) return
    const link = `${window.location.origin}/admin/register-admin?token=${token}`
    navigator.clipboard.writeText(link)
    setCopied(true)
    toast.success('تم نسخ الرابط')
    setTimeout(() => setCopied(false), 2000)
  }

  const columns: ColumnDef<AdminUser>[] = [
    {
      header: 'المستخدم',
      accessorKey: 'full_name',
      cell: ({ row }) => {
        const user = row.original
        return (
          <div className="flex items-center gap-4">
            <div className="w-10 h-10 rounded-xl bg-mazad-primary/10 border border-mazad-primary/20 flex items-center justify-center
                            text-mazad-primary text-sm font-bold uppercase shrink-0">
              {(user.full_name ?? 'U')[0]}
            </div>
            <div className="min-w-0">
              <p className="text-white font-bold text-sm truncate">{user.full_name ?? 'بدون اسم'}</p>
              <p className="text-[10px] text-surface-muted font-mono font-bold tracking-tight">{shortID(user.id)}</p>
            </div>
          </div>
        )
      }
    },
    {
      header: 'الهاتف',
      accessorKey: 'phone',
      cell: ({ getValue }) => <span className="font-mono text-xs text-surface-muted font-bold">{maskPhone(getValue<string>())}</span>
    },
    {
      header: 'الدور',
      accessorKey: 'role',
      cell: ({ getValue }) => {
        const role = getValue<string>()
        return (
          <span className={`text-[10px] font-bold px-2.5 py-1 rounded-lg uppercase tracking-wider ${
            role === 'admin'
              ? 'bg-mazad-primary/20 text-mazad-primary border border-mazad-primary/20'
              : role === 'driver'
              ? 'bg-purple-500/20 text-purple-400 border border-purple-500/20'
              : 'bg-surface-border/40 text-surface-muted'
          }`}>
            {role === 'admin' ? 'مدير' : role === 'driver' ? 'سائق' : 'مزايد'}
          </span>
        )
      }
    },
    {
      header: 'الحالة',
      accessorKey: 'is_verified',
      cell: ({ row }) => <StatusBadge status={row.original.is_verified ? 'verified' : 'unverified'} />
    },
    {
      header: 'تاريخ التسجيل',
      accessorKey: 'created_at',
      cell: ({ getValue }) => <span className="text-xs text-surface-muted font-medium">{formatDate(getValue<string>())}</span>
    },
    {
      header: 'الإجراءات',
      id: 'actions',
      cell: ({ row }) => {
        const user = row.original
        return (
          <div className="flex items-center gap-2">
            <button
              onClick={() => navigate(`/users/${user.id}`)}
              className="p-2 rounded-lg text-surface-muted hover:text-white hover:bg-surface-border transition-all"
              title="عرض الملف الشخصي"
            >
              <Eye className="w-4 h-4" />
            </button>
            {user.role !== 'admin' && (
              <button
                onClick={() => setBlockTarget({ id: user.id, block: user.is_active, name: user.full_name ?? shortID(user.id) })}
                className={`p-2 rounded-lg transition-all ${
                  user.is_active
                    ? 'text-red-400 hover:bg-red-500/10 hover:border-red-500/20'
                    : 'text-emerald-400 hover:bg-emerald-500/10 hover:border-emerald-500/20'
                } border border-transparent`}
                title={user.is_active ? 'حظر المستخدم' : 'إلغاء الحظر'}
              >
                {user.is_active ? <ShieldOff className="w-4 h-4" /> : <Shield className="w-4 h-4" />}
              </button>
            )}
          </div>
        )
      }
    }
  ]

  if (isError) return (
    <div className="admin-card p-20 text-center">
      <p className="text-red-400 font-bold mb-4">فشل تحميل المستخدمين</p>
      <button onClick={() => window.location.reload()} className="bg-mazad-primary text-white px-6 py-2 rounded-xl text-sm font-bold">إعادة المحاولة</button>
    </div>
  )

  return (
    <div className="animate-fade-in" dir="rtl">
      <PageHeader title="إدارة المستخدمين" subtitle={`${data?.total ?? 0} حساب مسجل في النظام`} />

      {/* Header & Controls */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6">
        <div className="relative flex-1 max-w-md group">
          <Search className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted group-focus-within:text-mazad-primary transition-colors" />
          <Input
            value={q}
            onChange={(e) => {
              setQ(e.target.value)
              setPage(1)
            }}
            placeholder="ابحث بالاسم، بريد إلكتروني أو رقم الهاتف..."
            className="pr-10"
          />
        </div>
        
        <button 
          onClick={() => {
            setToken(null)
            setShowInviteModal(true)
          }}
          className="bg-mazad-primary hover:bg-mazad-primary/90 text-white px-4 py-2 rounded-xl text-sm font-bold flex items-center gap-2 transition-all shadow-lg shadow-mazad-primary/20"
        >
          <UserPlus className="w-4 h-4" />
          دعوة مشرف جديد
        </button>
      </div>

      <DataTable
        columns={columns}
        data={data?.data ?? []}
        isLoading={isLoading}
        total={data?.total}
        page={page}
        onPageChange={setPage}
        emptyTitle="لا يوجد مستخدمين"
        emptyDescription="لم يتم العثور على أي نتائج تطابق بحثك."
      />

    <ConfirmDialog
        open={!!blockTarget}
        onOpenChange={(v) => !v && setBlockTarget(null)}
        title={blockTarget?.block ? `هل تود حظر ${blockTarget?.name}؟` : `هل تود إلغاء حظر ${blockTarget?.name}؟`}
        description={blockTarget?.block
          ? "في حالة الحظر، لن يتمكن المستخدم من تسجيل الدخول أو المشاركة في المزادات."
          : "سيتمكن المستخدم من العودة لاستخدام التطبيق مرة أخرى بشكل طبيعي."}
        confirmLabel={blockTarget?.block ? 'حظر المستخدم' : 'إلغاء الحظر'}
        variant={blockTarget?.block ? 'danger' : 'success'}
        loading={blockUser.isPending}
        onConfirm={() => {
          if (blockTarget) blockUser.mutate(
            { id: blockTarget.id, block: blockTarget.block },
            { onSuccess: () => setBlockTarget(null) }
          )
        }} 
      />


      {/* Invite Admin Modal */}
      {showInviteModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={() => setShowInviteModal(false)} />
          <div className="admin-card w-full max-w-md relative z-10 animate-in zoom-in-95 duration-200">
            <div className="p-6 border-b border-surface-border flex items-center justify-between">
              <h3 className="text-lg font-bold">دعوة مشرف جديد</h3>
              <button onClick={() => setShowInviteModal(false)} className="text-surface-muted hover:text-white">
                <X className="w-5 h-5" />
              </button>
            </div>
            
            <div className="p-6 space-y-6">
              {!token ? (
                <div className="text-center space-y-4">
                  <div className="w-16 h-16 bg-mazad-primary/10 rounded-full flex items-center justify-center mx-auto">
                    <LinkIcon className="w-8 h-8 text-mazad-primary" />
                  </div>
                  <div>
                    <p className="text-sm text-surface-muted">
                      قم بتوليد رابط دعوة خاص لإرساله إلى المشرف الجديد.
                      الرابط صالح لمرة واحدة ولمدة 24 ساعة.
                    </p>
                  </div>
                  <button
                    onClick={handleGenerateLink}
                    disabled={generateInvitation.isPending}
                    className="w-full bg-mazad-primary text-white py-3 rounded-xl font-bold hover:bg-mazad-primary/90 transition-all disabled:opacity-50"
                  >
                    {generateInvitation.isPending ? 'جاري التوليد...' : 'توليد رابط الدعوة'}
                  </button>
                </div>
              ) : (
                <div className="space-y-4">
                  <div className="p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-xl text-center">
                    <p className="text-emerald-400 text-sm font-bold">تم توليد الرابط بنجاح!</p>
                  </div>
                  
                  <div className="space-y-2">
                    <label className="text-xs font-bold text-surface-muted uppercase tracking-wider">رابط الدعوة</label>
                    <div className="relative group">
                      <Input 
                        readOnly
                        value={`${window.location.origin}/admin/register-admin?token=${token}`}
                        className="pr-10 bg-surface-base/50"
                      />
                      <button 
                        onClick={copyLink}
                        className="absolute left-2 top-1/2 -translate-y-1/2 p-2 hover:bg-surface-border rounded-lg transition-colors"
                      >
                        {copied ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4 text-surface-muted" />}
                      </button>
                    </div>
                  </div>
                  
                  <p className="text-[10px] text-surface-muted text-center">
                    شارك هذا الرابط مع الشخص الذي تود دعوته.
                  </p>
                  
                  <button
                    onClick={() => setShowInviteModal(false)}
                    className="w-full bg-surface-border text-white py-2.5 rounded-xl font-bold hover:bg-surface-border/80 transition-all"
                  >
                    إغلاق
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
