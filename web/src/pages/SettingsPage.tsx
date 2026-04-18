import { useState } from 'react'
import { Save, Settings as SettingsIcon } from 'lucide-react'
import { useSettings, useUpdateSetting } from '@/hooks/useSettings'
import { PageHeader } from '@/components/shared/PageHeader'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'

interface Setting {
  id: number
  key: string
  value: string
  type: string
}

export function SettingsPage() {
  const { data, isLoading, refetch } = useSettings()
  const updateSettings = useUpdateSetting()
  const [editingKey, setEditingKey] = useState<string | null>(null)
  const [editValue, setEditValue] = useState('')

  const settings = (data as Setting[]) || []

  const handleSave = async (key: string, type: string) => {
    await updateSettings.mutateAsync({ key, value: editValue, type })
    setEditingKey(null)
    refetch()
  }

  if (isLoading) return <LoadingSpinner />

  return (
    <div>
      <PageHeader
        title="الإعدادات"
        subtitle="إعدادات النظام العامة"
      />

      <div className="bg-surface-card rounded-lg border border-surface-border overflow-hidden">
        <table className="w-full">
          <thead className="bg-surface-border/50">
            <tr>
              <th className="text-right px-4 py-3 text-sm font-medium text-surface-muted">المفتاح</th>
              <th className="text-right px-4 py-3 text-sm font-medium text-surface-muted">القيمة</th>
              <th className="text-right px-4 py-3 text-sm font-medium text-surface-muted">النوع</th>
              <th className="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody className="divide-y divide-surface-border">
            {settings.map((setting) => (
              <tr key={setting.key} className="hover:bg-surface-border/30">
                <td className="px-4 py-3 font-mono text-sm">{setting.key}</td>
                <td className="px-4 py-3">
                  {editingKey === setting.key ? (
                    <input
                      type={setting.type === 'number' ? 'number' : 'text'}
                      value={editValue}
                      onChange={(e) => setEditValue(e.target.value)}
                      className="bg-surface-input border border-surface-border rounded px-2 py-1 text-sm w-full"
                    />
                  ) : (
                    <span className="text-sm">{setting.value}</span>
                  )}
                </td>
                <td className="px-4 py-3">
                  <span className="text-xs bg-surface-border/50 px-2 py-1 rounded">
                    {setting.type}
                  </span>
                </td>
                <td className="px-4 py-3">
                  {editingKey === setting.key ? (
                    <button
                      onClick={() => handleSave(setting.key, setting.type)}
                      disabled={updateSettings.isPending}
                      className="p-1 text-green-500 hover:text-green-400"
                    >
                      <Save className="w-4 h-4" />
                    </button>
                  ) : (
                    <button
                      onClick={() => {
                        setEditingKey(setting.key)
                        setEditValue(setting.value)
                      }}
                      className="p-1 text-surface-muted hover:text-white"
                    >
                      <SettingsIcon className="w-4 h-4" />
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}