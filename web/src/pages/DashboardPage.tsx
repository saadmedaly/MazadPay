import {
  Gavel, CreditCard, Users, Flag,
  TrendingUp, RefreshCw, AlertTriangle, AlertCircle, ShieldCheck,
  Plus, Settings, Edit2, Trash2, Grid3x3
} from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import {
  AreaChart, Area,
  XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer
} from 'recharts'
import { PageHeader } from '@/components/shared/PageHeader'
 import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { useDashboardStats, useRevenueChart, useActivityFeed } from '@/hooks/useDashboard'
import { 
  useDashboardWidgets, 
  useCreateDashboardWidget, 
  useUpdateDashboardWidget, 
  useDeleteDashboardWidget,
  useRepositionWidgets 
} from '@/hooks/useDashboardWidgets'
import { formatPrice, formatRelative, formatDateShort } from '@/lib/formatters'
import { MetricCard } from '@/components/shared/MetricCard'
import { ConfirmDialog } from '@/components/shared/ConfirmDialog'
import { Input } from '@/components/ui/input'
import { useState } from 'react'

const ChartTooltip = ({ active, payload, label }: any) => {
  if (!active || !payload?.length) return null
  return (
    <div className="bg-surface-card border border-surface-border rounded-lg p-3 text-xs shadow-2xl">
      <p className="text-surface-muted mb-1 font-medium">{label}</p>
      <p className="text-white font-mono font-bold text-sm">{formatPrice(payload[0].value)}</p>
    </div>
  )
}

export function DashboardPage() {
  const navigate = useNavigate()
  const [showWidgetModal, setShowWidgetModal] = useState(false)
  const [editingWidget, setEditingWidget] = useState<any>(null)
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)
  const [isEditMode, setIsEditMode] = useState(false)

  const { data: stats, isLoading, isError, refetch, isFetching } = useDashboardStats()
  const { data: revenueData } = useRevenueChart()
  const revenue = Array.isArray(revenueData) ? revenueData : []
  const { data: activityData } = useActivityFeed()
  const activity = Array.isArray(activityData) ? activityData : []

  // Widget management hooks
  const { data: widgets = [] } = useDashboardWidgets()
  const createWidget = useCreateDashboardWidget()
  const updateWidget = useUpdateDashboardWidget()
  const deleteWidget = useDeleteDashboardWidget()
  const repositionWidgets = useRepositionWidgets()

  const [newWidget, setNewWidget] = useState({
    title: '',
    type: 'metric' as const,
    position: { x: 0, y: 0, w: 4, h: 2 },
    config: {}
  })

  // Widget handlers
  const handleCreateWidget = () => {
    createWidget.mutate(newWidget, {
      onSuccess: () => {
        setShowWidgetModal(false)
        setNewWidget({
          title: '',
          type: 'metric',
          position: { x: 0, y: 0, w: 4, h: 2 },
          config: {}
        })
      }
    })
  }

  const handleUpdateWidget = () => {
    if (!editingWidget) return
    updateWidget.mutate({
      id: editingWidget.id,
      data: newWidget
    }, {
      onSuccess: () => {
        setShowWidgetModal(false)
        setEditingWidget(null)
        setNewWidget({
          title: '',
          type: 'metric',
          position: { x: 0, y: 0, w: 4, h: 2 },
          config: {}
        })
      }
    })
  }

  const handleDeleteWidget = () => {
    if (!deleteConfirm) return
    deleteWidget.mutate(deleteConfirm, {
      onSuccess: () => {
        setDeleteConfirm(null)
      }
    })
  }

  const METRICS = [
    {
      label: 'المزادات النشطة',
      value: stats?.active_auctions ?? '—',
      icon: Gavel,
      accent: 'blue' as const,
    },
    {
      label: 'في انتظار الموافقة',
      value: stats?.pending_auctions ?? '—',
      icon: AlertTriangle,
      accent: 'amber' as const,
      urgent: (stats?.pending_auctions ?? 0) > 0,
    },
    {
      label: 'إيداعات للمراجعة',
      value: stats?.pending_transactions ?? '—',
      icon: CreditCard,
      accent: 'orange' as const,
      urgent: (stats?.pending_transactions ?? 0) > 0,
    },
    {
      label: 'المستخدمين المسجلين',
      value: stats?.total_users ?? '—',
      icon: Users,
      accent: 'green' as const,
    },
    {
      label: 'البلاغات المفتوحة',
      value: stats?.pending_reports ?? '—',
      icon: Flag,
      accent: 'red' as const,
      urgent: (stats?.pending_reports ?? 0) > 0,
    },
    {
      label: 'توثيق الحسابات  ',
      value: stats?.pending_kycs ?? '—',
      icon: ShieldCheck,
      accent: 'purple' as const,
      urgent: (stats?.pending_kycs ?? 0) > 0,
    },
    {
      label: 'الإيداعات (7 أيام)',
      value: stats ? formatPrice(stats.week_deposits) : '—',
      icon: TrendingUp,
      accent: 'blue' as const,
    },
  ]

  if (isLoading) return <LoadingSpinner fullPage label="جاري تحميل إحصائيات لوحة التحكم..." />

  if (isError) return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] text-surface-muted gap-4">
      <AlertCircle className="w-16 h-16 opacity-20 text-red-500" />
      <h2 className="text-white font-display font-bold text-xl">فشل في الاتصال بمزود البيانات</h2>
      <p className="text-sm max-w-md text-center">يرجى التحقق من اتصالك بالإنترنت أو تأكد من أن خادم الواجهة البرمجية يعمل بشكل صحيح.</p>
      <button 
        onClick={() => refetch()} 
        className="mt-4 flex items-center gap-2 bg-mazad-primary text-white px-8 py-3 rounded-xl font-bold shadow-lg shadow-mazad-primary/20"
      >
        <RefreshCw className="w-4 h-4" />
        إعادة المحاولة الآن
      </button>
    </div>
  )

  return (
    <div className="animate-fade-in" dir="rtl">
      <PageHeader 
        title="لوحة التحكم"
        subtitle="نظرة مباشرة على منصة MazadPay"
        actions={[
          {
            label: isEditMode ? "إيقاف تعديل" : "وضع التعديل",
            icon: isEditMode ? AlertCircle : Edit2,
            onClick: () => setIsEditMode(!isEditMode)
          },
          {
            label: "إضافة widget",
            icon: Plus,
            onClick: () => setShowWidgetModal(true)
          },
          {
            label: "إدارة widgets",
            icon: Settings,
            onClick: () => {/* Navigate to widget management */}}
        ]}
      />

      {/* Widget Management Modal */}
      {showWidgetModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-surface-card border border-surface-border rounded-xl p-6 max-w-md w-full mx-4">
            <h3 className="text-white font-bold text-lg mb-4">
              {editingWidget ? "تعديل widget" : "إضافة widget"}
            </h3>
            <div className="space-y-4">
              <div>
                <label className="text-surface-muted text-sm block mb-1">العنوان</label>
                <Input
                  value={newWidget.title}
                  onChange={(e) => setNewWidget({...newWidget, title: e.target.value})}
                  placeholder="اسم widget"
                />
              </div>
              <div>
                <label className="text-surface-muted text-sm block mb-1">النوع</label>
                <select 
                  value={newWidget.type}
                  onChange={(e) => setNewWidget({...newWidget, type: e.target.value as any})}
                  className="w-full bg-surface-base border border-surface-border rounded-lg px-3 py-2 text-white"
                >
                  <option value="metric">مقياس</option>
                  <option value="chart">رسم بياني</option>
                  <option value="table">جدول</option>
                </select>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={editingWidget ? handleUpdateWidget : handleCreateWidget}
                  className="flex-1 bg-mazad-primary text-white px-4 py-2 rounded-lg font-bold"
                >
                  {editingWidget ? "تحديث" : "إنشاء"}
                </button>
                <button
                  onClick={() => {
                    setShowWidgetModal(false)
                    setEditingWidget(null)
                    setNewWidget({
                      title: '',
                      type: 'metric',
                      position: { x: 0, y: 0, w: 4, h: 2 },
                      config: {}
                    })
                  }}
                  className="flex-1 bg-surface-border text-white px-4 py-2 rounded-lg font-bold"
                >
                  إلغاء
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation */}
      {deleteConfirm && (
        <ConfirmDialog
          title="حذف widget"
          description="هل أنت متأكد من حذف widget هذا؟"
          onConfirm={handleDeleteWidget}
          onCancel={() => setDeleteConfirm(null)}
        />
      )}

      {/* Métriques */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
        {METRICS.map((m) => (
          <MetricCard key={m.label} {...m} />
        ))}
      </div>

      {/* Graphiques */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
        {/* Revenue Chart — 2/3 */}
        <div className="lg:col-span-2 admin-card p-6">
          <h2 className="font-display font-bold text-white text-base mb-6 px-1">
            الإيداعات المعتمدة — آخر 30 يوماً
          </h2>
          <div className="h-[250px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={revenue} margin={{ top: 0, right: 0, left: -20, bottom: 0 }}>
                <defs>
                  <linearGradient id="revenueGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%"  stopColor="#1B4FD8" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#1B4FD8" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#2A2D3E" vertical={false} />
                <XAxis 
                   dataKey="date" 
                   tick={{ fill: '#6B7280', fontSize: 11, fontWeight: 500 }} 
                   tickFormatter={formatDateShort} 
                   axisLine={false}
                   tickLine={false}
                   dy={10}
                />
                <YAxis 
                   tick={{ fill: '#6B7280', fontSize: 11, fontWeight: 500 }}
                   axisLine={false}
                   tickLine={false} 
                />
                <Tooltip content={<ChartTooltip />} />
                <Area type="monotone" dataKey="amount" stroke="#1B4FD8" strokeWidth={3}
                      fill="url(#revenueGrad)" animationDuration={1000} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Activity Feed — 1/3 */}
        <div className="admin-card p-6">
          <h2 className="font-display font-bold text-white text-base mb-6 px-1">
            النشاط الأخير
          </h2>
          <div className="space-y-6">
            {activity.slice(0, 6).map((item) => (
              <div key={item.id} className="flex gap-4 items-start relative pb-5 last:pb-0">
                {/* Timeline connector */}
                <div className="absolute top-5 bottom-0 right-[7px] w-[2px] bg-surface-border last:hidden" />
                
                <div className="w-4 h-4 rounded-full bg-mazad-primary/20 border-2 border-mazad-primary mt-1 shrink-0 z-10" />
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-white font-bold leading-tight">{item.description}</p>
                  <p className="text-[10px] text-surface-muted mt-2 font-bold uppercase tracking-wider">{formatRelative(item.created_at)}</p>
                </div>
              </div>
            ))}
            {!isLoading && activity.length === 0 && (
              <div className="text-center py-10 opacity-50 flex flex-col items-center gap-2">
                 <RefreshCw className="w-8 h-8 text-surface-muted animate-pulse" />
                 <p className="text-xs text-surface-muted font-bold">لا يوجد نشاط متاح حالياً</p>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Alertes urgentes */}
      {((stats?.pending_auctions ?? 0) > 0 || (stats?.pending_transactions ?? 0) > 0) && (
        <div className="admin-card border-orange-500/30 p-8 shadow-2xl shadow-orange-500/5 relative overflow-hidden">
          <div className="absolute top-0 right-0 w-32 h-32 bg-orange-500/5 rounded-full -mr-16 -mt-16 blur-3xl" />
          
          <div className="flex items-center gap-3 mb-6 relative">
            <div className="p-2 rounded-lg bg-orange-500/10 text-orange-400">
              <AlertTriangle className="w-5 h-5" />
            </div>
            <h2 className="font-display font-bold text-orange-400 text-lg">إجراءات إدارية معلقة</h2>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 relative">
            {(stats?.pending_auctions ?? 0) > 0 && (
              <button
                onClick={() => navigate('/auctions?status=pending')}
                className="flex items-center justify-between bg-amber-500/5 border border-amber-500/20 
                           text-amber-400 text-sm font-bold py-5 px-6 rounded-2xl hover:bg-amber-500/15 
                           transition-all active:scale-[0.98] text-right shadow-sm"
              >
                <span>{stats?.pending_auctions} مزادات في انتظار المراجعة</span>
                <span className="text-xl">←</span>
              </button>
            )}
            {(stats?.pending_transactions ?? 0) > 0 && (
              <button
                onClick={() => navigate('/transactions?status=pending_review')}
                className="flex items-center justify-between bg-orange-500/5 border border-orange-500/20 
                           text-orange-400 text-sm font-bold py-5 px-6 rounded-2xl hover:bg-orange-500/15 
                           transition-all active:scale-[0.98] text-right shadow-sm"
              >
                <span>{stats?.pending_transactions} إيداعات مالية للموافقة عليها</span>
                <span className="text-xl">←</span>
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
