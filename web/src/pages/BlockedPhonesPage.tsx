import { useState, useEffect } from 'react'
import { Plus, Trash2, Search, User, X } from 'lucide-react'
import { useBlockedPhones, useBlockPhone, useUnblockPhone } from '@/hooks/useBlockedPhones'
import { PageHeader } from '@/components/shared/PageHeader'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { DataTable } from '@/components/shared/DataTable'
import { Input } from '@/components/ui/input'
import { toast } from 'sonner'
import type { ColumnDef } from '@tanstack/react-table'
import client from '@/api/client'

interface BlockedPhone {
  phone: string
  reason: string | null
  blocked_at: string
  expires_at: string | null
}

interface UserOption {
  id: string
  phone: string
  full_name: string | null
  is_active: boolean
}

export function BlockedPhonesPage() {
  const { data, isLoading, refetch } = useBlockedPhones()
  const blockPhone = useBlockPhone()
  const unblockPhone = useUnblockPhone()
  
  const [showAddManual, setShowAddManual] = useState(false)
  const [showUserSelector, setShowUserSelector] = useState(false)
  
  const [unblockTarget, setUnblockTarget] = useState<{ phone: string; name?: string } | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<UserOption[]>([])
  const [searching, setSearching] = useState(false)
  const [userToBlock, setUserToBlock] = useState<UserOption | null>(null)

  const [manualPhone, setManualPhone] = useState('')
  const [manualReason, setManualReason] = useState('')

  const phones = (data as BlockedPhone[]) || []
  const blockedList = phones.map(p => p.phone)

  const handleBlockUser = async (user: UserOption) => {
    try {
      await blockPhone.mutateAsync({ 
        phone: user.phone, 
        reason: user.full_name ? `حظر المستخدم: ${user.full_name}` : 'حظر من لوحة الإدارة' 
      })
      refetch()
      toast.success(`تم حظر ${user.full_name || user.phone} بنجاح`)
    } catch {
      toast.error('فشل حظر المستخدم')
    }
  }

  const handleUnblock = async () => {
    if (!unblockTarget) return
    try {
      await unblockPhone.mutateAsync(unblockTarget.phone)
      refetch()
      toast.success('تم إلغاء الحظر بنجاح')
    } catch {
      toast.error('فشل إلغاء الحظر')
    }
  }

  const handleSearchUsers = async (query: string) => {
    if (query.length < 2) { setSearchResults([]); return }
    setSearching(true)
    try {
      const { data: res } = await client.get(`/v1/api/admin/users?q=${encodeURIComponent(query)}&per_page=15`)
      if (res.success) {
        setSearchResults((res.data || []).filter((u: UserOption) => !blockedList.includes(u.phone)))
      }
    } catch (e) {
      console.error(e)
    } finally {
      setSearching(false)
    }
  }

  useEffect(() => {
    const timer = setTimeout(() => handleSearchUsers(searchQuery), 500)
    return () => clearTimeout(timer)
  }, [searchQuery])

  const columns: ColumnDef<BlockedPhone>[] = [
    {
      header: 'الرقم',
      accessorKey: 'phone',
      cell: ({ getValue }) => {
        const phone = getValue<string>()
        const formatted = phone.startsWith('222') ? `+222 ${phone.slice(3)}` : phone
        return <span className="font-mono font-bold text-red-400 text-sm">{formatted}</span>
      }
    },
    {
      header: 'السبب',
      accessorKey: 'reason',
      cell: ({ getValue }) => (
        <span className="text-sm text-surface-muted">{getValue<string>() || '—'}</span>
      )
    },
    {
      header: 'تاريخ الحظر',
      accessorKey: 'blocked_at',
      cell: ({ getValue }) => (
        <span className="text-sm text-surface-muted">{new Date(getValue<string>()).toLocaleDateString('ar')}</span>
      )
    },
    {
      id: 'actions',
      header: '',
      cell: ({ row }) => {
        const phone = row.original.phone
        return (
          <button
            onClick={() => setUnblockTarget({ phone })}
            className="p-2 text-surface-muted hover:text-green-400 hover:bg-green-500/10 rounded-lg transition-all"
            title="إلغاء الحظر"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        )
      }
    }
  ]

  if (isLoading) return <LoadingSpinner />

  return (
    <div>
      <PageHeader
        title="أرقام محظورة"
        subtitle="إدارة الأرقام والمستخدمين المحظورين"
      />

      <div className="flex flex-wrap gap-3 mb-6">
        <button
          onClick={() => setShowUserSelector(true)}
          className="flex items-center gap-2 bg-mazad-primary hover:bg-mazad-primary/90 px-4 py-2.5 rounded-xl text-sm font-bold text-white shadow-lg shadow-mazad-primary/20 active:scale-95 transition-all"
        >
          <User className="w-4 h-4" />
          حظر مستخدم
        </button>
        <button
          onClick={() => setShowAddManual(true)}
          className="flex items-center gap-2 bg-surface-card border border-surface-border hover:border-red-500/50 px-4 py-2.5 rounded-xl text-sm text-surface-muted hover:text-white transition-all"
        >
          <Plus className="w-4 h-4" />
          حظر رقم يدوي
        </button>
      </div>

      <div className="admin-card overflow-hidden">
        <DataTable columns={columns} data={phones} emptyMessage="لا توجد أرقام محظورة" />
      </div>

      {showUserSelector && (
        <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="bg-surface-card w-full max-w-lg rounded-2xl border border-surface-border overflow-hidden flex flex-col max-h-[80vh]">
            <div className="flex items-center justify-between p-4 border-b border-surface-border">
              <h3 className="text-lg font-bold text-white">اختر المستخدم للحظر</h3>
              <button
                onClick={() => { setShowUserSelector(false); setSearchQuery(''); setSearchResults([]) }}
                className="p-2 hover:bg-surface-border rounded-lg transition-all"
              >
                <X className="w-5 h-5 text-surface-muted" />
              </button>
            </div>

            <div className="p-4 border-b border-surface-border">
              <div className="relative">
                <Search className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted" />
                <Input
                  placeholder="بحث بالاسم أو الرقم..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pr-10"
                />
              </div>
            </div>

            <div className="flex-1 overflow-y-auto p-4">
              {searching ? (
                <div className="p-8 text-center text-surface-muted">جاري البحث...</div>
              ) : searchResults.length === 0 && searchQuery.length >= 2 ? (
                <div className="p-8 text-center text-surface-muted">لم يتم العثور على مستخدمين</div>
              ) : searchQuery.length < 2 ? (
                <div className="p-8 text-center text-surface-muted">ابحث بالاسم أو الرقم للمتابعة</div>
              ) : (
                <div className="space-y-2">
                  {searchResults.map((user) => (
                    <div
                      key={user.id}
                      className="flex items-center justify-between p-3 rounded-xl bg-surface-base border border-surface-border hover:border-mazad-primary/30 transition-all"
                    >
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-xl bg-mazad-primary/10 border border-mazad-primary/20 flex items-center justify-center text-mazad-primary text-sm font-bold uppercase">
                          {(user.full_name ?? 'U')[0]}
                        </div>
                        <div>
                          <p className="font-bold text-white text-sm">{user.full_name || 'بدون اسم'}</p>
                          <p className="text-xs text-surface-muted font-mono">{user.phone}</p>
                        </div>
                      </div>
                      <button
                        onClick={() => handleBlockUser(user)}
                        className="px-3 py-1.5 rounded-lg text-xs font-bold bg-red-600 hover:bg-red-500 text-white"
                      >
                        حظر
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {showAddManual && (
        <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="bg-surface-card w-full max-w-md rounded-2xl border border-surface-border p-6">
            <h3 className="text-lg font-bold text-white mb-4">حظر رقم جديد</h3>
            
            <div className="space-y-4">
              <div>
                <label className="text-sm text-surface-muted mb-1.5 block">رقم الهاتف</label>
                <Input
                  placeholder="+22212345678"
                  value={manualPhone}
                  onChange={(e) => setManualPhone(e.target.value.replace(/[^\d+]/g, ''))}
                  className="font-mono text-lg"
                  dir="ltr"
                />
              </div>
              
              <div>
                <label className="text-sm text-surface-muted mb-1.5 block">السبب (اختياري)</label>
                <textarea
                  value={manualReason}
                  onChange={(e) => setManualReason(e.target.value)}
                  className="w-full bg-surface-input border border-surface-border rounded-xl p-3 text-sm"
                  rows={2}
                />
              </div>
            </div>

            <div className="flex gap-3 justify-end mt-6">
              <button
                onClick={() => { setShowAddManual(false); setManualPhone(''); setManualReason('') }}
                className="px-4 py-2.5 text-surface-muted hover:text-white"
              >
                إلغاء
              </button>
              <button
                onClick={async () => {
                  if (!manualPhone || manualPhone.length < 8) return
                  try {
                    // Format phone number with +222 prefix if not present
                    const formattedPhone = manualPhone.startsWith('+') ? manualPhone : `+222${manualPhone}`
                    await blockPhone.mutateAsync({ phone: formattedPhone, reason: manualReason })
                    refetch()
                    setShowAddManual(false)
                    setManualPhone('')
                    setManualReason('')
                    toast.success('تم حظر الرقم بنجاح')
                  } catch {
                    toast.error('فشل حظر الرقم')
                  }
                }}
                disabled={!manualPhone || manualPhone.length < 8 || blockPhone.isPending}
                className="px-4 py-2.5 bg-red-600 hover:bg-red-500 rounded-xl font-bold text-white disabled:opacity-50"
              >
                {blockPhone.isPending ? 'جاري...' : 'حظر'}
              </button>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={!!unblockTarget}
        onOpenChange={(v) => !v && setUnblockTarget(null)}
        title="إلغاء حظر"
        message={`هل أنت متأكد من إلغاء حظر الرقم ${unblockTarget?.phone}؟`}
        confirmLabel="نعم, إلغاء الحظر"
        variant="danger"
        onConfirm={handleUnblock}
        onCancel={() => setUnblockTarget(null)}
      />

      <ConfirmDialog
        open={!!userToBlock}
        onOpenChange={(v) => !v && setUserToBlock(null)}
        title="تأكيد حظر"
        message={`هل أنت متأكد من حظر ${userToBlock?.full_name || userToBlock?.phone}؟`}
        confirmLabel="نعم, حظر"
        variant="danger"
        onConfirm={() => {
          if (userToBlock) handleBlockUser(userToBlock)
          setUserToBlock(null)
        }}
        onCancel={() => setUserToBlock(null)}
      />
    </div>
  )
}