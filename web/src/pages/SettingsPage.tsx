import { useState } from 'react'
import { Save, Settings as SettingsIcon, Shield, Clock, Phone, AlertTriangle, Check, Plus, Edit2, Trash2 } from 'lucide-react'
import { useSettings, useUpdateSetting, useCreateSetting, useDeleteSetting, useBulkUpdateSettings } from '@/hooks/useSettingsCRUD'
import { PageHeader } from '@/components/shared/PageHeader'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { Input } from '@/components/ui/input'
import { toast } from 'sonner'

interface Setting {
  id: number
  key: string
  value: string
  type: string
}

interface SettingGroup {
  title: string
  icon: React.ElementType
  settings: {
    key: string
    labelAr: string
    labelFr: string
    description: string
    type: 'boolean' | 'number' | 'text'
    options?: { value: string; label: string }[]
  }[]
}

const SETTING_GROUPS: SettingGroup[] = [
  {
    title: 'إعدادات عامة',
    icon: Shield,
    settings: [
      {
        key: 'maintenance_mode',
        labelAr: 'وضع الصيانة',
        labelFr: 'Maintenance Mode',
        description: 'تفعيل وضع الصيانة لإيقاف الموقع مؤقتاً',
        type: 'boolean'
      },
      {
        key: 'registration_open',
        labelAr: 'فتح التسجيل',
        labelFr: 'Registration Open',
        description: 'السماح بتسجيل مستخدمين جدد',
        type: 'boolean'
      }
    ]
  },
  {
    title: 'إعدادات المزادات',
    icon: Clock,
    settings: [
      {
        key: 'max_auction_duration_hours',
        labelAr: 'مدة المزاد القصوى',
        labelFr: 'Max Auction Duration',
        description: 'أقصى وقت للمزاد بالساعات',
        type: 'number'
      },
      {
        key: 'default_insurance_amount',
        labelAr: 'مبلغ التأمين الافتراضي',
        labelFr: 'Default Insurance',
        description: 'مبلغ التأمين المطلوب للمشاركة',
        type: 'number'
      },
      {
        key: 'min_bid_increment',
        labelAr: 'الحد الأدنى للزيادة',
        labelFr: 'Min Bid Increment',
        description: 'الحد الأدنى لزيادة المزايدة',
        type: 'number'
      }
    ]
  },
  {
    title: 'معلومات الاتصال',
    icon: Phone,
    settings: [
      {
        key: 'contact_whatsapp',
        labelAr: 'رقم واتساب',
        labelFr: 'WhatsApp Number',
        description: 'رقم الواتساب للتواصل',
        type: 'text'
      },
      {
        key: 'contact_email',
        labelAr: 'البريد الإلكتروني',
        labelFr: 'Contact Email',
        description: 'بريد إلكتروني للتواصل',
        type: 'text'
      }
    ]
  }
]

function getSettingValue(settings: Setting[], key: string): string {
  return settings.find(s => s.key === key)?.value || ''
}

export function SettingsPage() {
  const { data: settings, isLoading } = useSettings()
  const { mutate: updateSetting } = useUpdateSetting()
  const createSetting = useCreateSetting()
  const deleteSetting = useDeleteSetting()
  
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)
  const [newSetting, setNewSetting] = useState({
    key: '',
    value: '',
    type: 'text' as const,
    description: '',
    group: 'general'
  })
  const [editingKey, setEditingKey] = useState<string | null>(null)
  const [editValue, setEditValue] = useState('')
  const [saving, setSaving] = useState(false)

  const settingsData = (settings as Setting[]) || []

  const handleSave = async (key: string, type: string) => {
    setSaving(true)
    try {
      await updateSettings.mutateAsync({ key, value: editValue, type })
      setEditingKey(null)
      refetch()
      toast.success('تم حفظ الإعداد بنجاح')
    } catch (err) {
      toast.error('فشل حفظ الإعداد')
    } finally {
      setSaving(false)
    }
  }

  const handleToggle = async (key: string, currentValue: string) => {
    const newValue = currentValue === 'true' ? 'false' : 'true'
    setSaving(true)
    try {
      await updateSettings.mutateAsync({ key, value: newValue, type: 'boolean' })
      refetch()
      toast.success('تم حفظ الإعداد بنجاح')
    } catch (err) {
      toast.error('فشل حفظ الإعداد')
    } finally {
      setSaving(false)
    }
  }

  if (isLoading) return <LoadingSpinner />

  return (
    <div>
      <PageHeader
        title="الإعدادات"
        subtitle="إعدادات النظام العامة"
      />

      <div className="space-y-6">
        {SETTING_GROUPS.map((group) => {
          const Icon = group.icon
          return (
            <div key={group.title} className="bg-surface-card rounded-xl border border-surface-border overflow-hidden">
              <div className="flex items-center gap-3 px-6 py-4 bg-surface-border/30 border-b border-surface-border">
                <Icon className="w-5 h-5 text-mazad-primary" />
                <h2 className="text-lg font-bold text-white">{group.title}</h2>
              </div>

              <div className="divide-y divide-surface-border">
                {group.settings.map((settingDef) => {
                  const currentValue = getSettingValue(settings, settingDef.key)
                  const isEditing = editingKey === settingDef.key

                  return (
                    <div key={settingDef.key} className="p-6">
                      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                        <div className="flex-1">
                          <div className="flex items-center gap-2 mb-1">
                            <h3 className="text-base font-bold text-white">{settingDef.labelAr}</h3>
                            {settingDef.type === 'boolean' && (
                              <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${
                                currentValue === 'true' 
                                  ? 'bg-green-500/20 text-green-400' 
                                  : 'bg-red-500/20 text-red-400'
                              }`}>
                                {currentValue === 'true' ? 'مفعل' : 'معطل'}
                              </span>
                            )}
                          </div>
                          <p className="text-sm text-surface-muted">{settingDef.description}</p>
                        </div>

                        <div className="w-full md:w-64">
                          {settingDef.type === 'boolean' ? (
                            <button
                              onClick={() => handleToggle(settingDef.key, currentValue)}
                              disabled={saving}
                              className={`w-full h-12 rounded-xl border-2 flex items-center justify-center gap-2 font-bold transition-all ${
                                currentValue === 'true'
                                  ? 'bg-green-500/20 border-green-500/50 text-green-400 hover:bg-green-500/30'
                                  : 'bg-red-500/20 border-red-500/50 text-red-400 hover:bg-red-500/30'
                              }`}
                            >
                              {currentValue === 'true' ? (
                                <>
                                  <Check className="w-5 h-5" />
                                  مفعل
                                </>
                              ) : (
                                <>
                                  <AlertTriangle className="w-5 h-5" />
                                  معطل
                                </>
                              )}
                            </button>
                          ) : isEditing ? (
                            <div className="flex gap-2">
                              <input
                                type={settingDef.type === 'number' ? 'number' : 'text'}
                                value={editValue}
                                onChange={(e) => setEditValue(e.target.value)}
                                className="flex-1 bg-surface-input border border-surface-border rounded-xl px-4 py-2.5 text-white font-medium"
                              />
                              <button
                                onClick={() => handleSave(settingDef.key, settingDef.type)}
                                disabled={saving}
                                className="px-4 py-2.5 bg-mazad-primary hover:bg-mazad-primary/90 text-white font-bold rounded-xl flex items-center gap-2"
                              >
                                <Save className="w-4 h-4" />
                              </button>
                            </div>
                          ) : (
                            <button
                              onClick={() => {
                                setEditingKey(settingDef.key)
                                setEditValue(currentValue)
                              }}
                              className="w-full h-12 rounded-xl border border-surface-border px-4 py-2.5 text-white font-medium flex items-center justify-between hover:border-mazad-primary/50 transition-all"
                            >
                              <span>{currentValue || '—'}</span>
                              <SettingsIcon className="w-4 h-4 text-surface-muted" />
                            </button>
                          )}
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}