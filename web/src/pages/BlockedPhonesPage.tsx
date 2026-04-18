import { useState } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import { useBlockedPhones, useBlockPhone, useUnblockPhone } from '@/hooks/useBlockedPhones'
import { PageHeader } from '@/components/shared/PageHeader'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'

interface BlockedPhone {
  phone: string
  reason: string | null
  blocked_at: string
  expires_at: string | null
}

export function BlockedPhonesPage() {
  const { data, isLoading, refetch } = useBlockedPhones()
  const blockPhone = useBlockPhone()
  const unblockPhone = useUnblockPhone()
  const [showAdd, setShowAdd] = useState(false)
  const [phone, setPhone] = useState('')
  const [reason, setReason] = useState('')
  const [unblockPhoneNum, setUnblockPhoneNum] = useState<string | null>(null)

  const phones = (data as BlockedPhone[]) || []

  const handleBlock = async () => {
    await blockPhone.mutateAsync({ phone, reason })
    setShowAdd(false)
    setPhone('')
    setReason('')
    refetch()
  }

  const handleUnblock = async () => {
    if (!unblockPhoneNum) return
    await unblockPhone.mutateAsync(unblockPhoneNum)
    setUnblockPhoneNum(null)
    refetch()
  }

  if (isLoading) return <LoadingSpinner />

  return (
    <div>
      <PageHeader
        title="أرقام محظورة"
        subtitle="إدارة الأرقام الممنوعة من التسجيل"
        action={
          <button
            onClick={() => setShowAdd(true)}
            className="flex items-center gap-2 bg-primary hover:bg-primary/90 px-4 py-2 rounded-lg text-sm"
          >
            <Plus className="w-4 h-4" />
            حظر رقم
          </button>
        }
      />

      <div className="bg-surface-card rounded-lg border border-surface-border overflow-hidden">
        <table className="w-full">
          <thead className="bg-surface-border/50">
            <tr>
              <th className="text-right px-4 py-3 text-sm font-medium text-surface-muted">الرقم</th>
              <th className="text-right px-4 py-3 text-sm font-medium text-surface-muted">السبب</th>
              <th className="text-right px-4 py-3 text-sm font-medium text-surface-muted">تاريخ الحظر</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody className="divide-y divide-surface-border">
            {phones.map((p) => (
              <tr key={p.phone} className="hover:bg-surface-border/30">
                <td className="px-4 py-3 font-mono font-bold text-red-400">{p.phone}</td>
                <td className="px-4 py-3 text-sm text-surface-muted">{p.reason || '—'}</td>
                <td className="px-4 py-3 text-sm">{new Date(p.blocked_at).toLocaleDateString('ar')}</td>
                <td className="px-4 py-3">
                  <button
                    onClick={() => setUnblockPhoneNum(p.phone)}
                    className="p-1 text-surface-muted hover:text-green-400"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </td>
              </tr>
            ))}
            {phones.length === 0 && (
              <tr>
                <td colSpan={4} className="px-4 py-8 text-center text-surface-muted">
                  لا توجد أرقام محظورة
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {showAdd && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-surface-card p-6 rounded-lg w-96 border border-surface-border">
            <h3 className="text-lg font-bold mb-4">حظر رقم جديد</h3>
            <input
              placeholder="رقم الهاتف"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              className="w-full bg-surface-input border border-surface-border rounded-lg p-3 mb-3"
            />
            <textarea
              placeholder="السبب (اختياري)"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              className="w-full bg-surface-input border border-surface-border rounded-lg p-3 mb-4"
              rows={3}
            />
            <div className="flex gap-2 justify-end">
              <button
                onClick={() => setShowAdd(false)}
                className="px-4 py-2 text-surface-muted hover:text-white"
              >
                إلغاء
              </button>
              <button
                onClick={handleBlock}
                disabled={!phone || blockPhone.isPending}
                className="bg-red-600 hover:bg-red-500 px-4 py-2 rounded-lg"
              >
                حظر
              </button>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={!!unblockPhoneNum}
        title="إلغاء حظر"
        message={`هل أنت متأكد من إلغاء حظر الرقم ${unblockPhoneNum}؟`}
        confirmLabel="نعم, إلغاء الحظر"
        onConfirm={handleUnblock}
        onCancel={() => setUnblockPhoneNum(null)}
      />
    </div>
  )
}