import React, { useState } from 'react'
import {
  Plus, Search, Pencil, Trash2, MapPin,
  AlertCircle, Loader2
} from 'lucide-react'
import { useLocations, useCreateLocation, useUpdateLocation, useDeleteLocation } from '@/hooks/useMetadata'
 import { type ColumnDef } from '@tanstack/react-table'
import { type Location } from '@/types/api'
import { PageHeader } from '@/components/shared/PageHeader'
import { DataTable } from '@/components/shared/DataTable'
import { Input } from '@/components/ui/input'

export function LocationsPage() {
  const { data: locations, isLoading, isError } = useLocations()
  const createMut = useCreateLocation()
  const updateMut = useUpdateLocation()
  const deleteMut = useDeleteLocation()

  const [search, setSearch] = useState('')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingLocation, setEditingLocation] = useState<Location | null>(null)
  const [form, setForm] = useState({
    city_name_ar: '',
    city_name_fr: '',
    area_name_ar: '',
    area_name_fr: ''
  })
  const [deleteId, setDeleteId] = useState<number | null>(null)

  // Filter
  const filtered = locations?.filter(l => 
    l.id.toString().includes(search) ||
    l.city_name_ar.toLowerCase().includes(search.toLowerCase()) ||
    l.city_name_fr.toLowerCase().includes(search.toLowerCase()) ||
    (l.area_name_ar && l.area_name_ar.toLowerCase().includes(search.toLowerCase())) ||
    (l.area_name_fr && l.area_name_fr.toLowerCase().includes(search.toLowerCase()))
  ) || []

  const openAdd = () => {
    setEditingLocation(null)
    setForm({ city_name_ar: '', city_name_fr: '', area_name_ar: '', area_name_fr: '' })
    setIsModalOpen(true)
  }

  const openEdit = (loc: Location) => {
    setEditingLocation(loc)
    setForm({
      city_name_ar: loc.city_name_ar,
      city_name_fr: loc.city_name_fr,
      area_name_ar: loc.area_name_ar || '',
      area_name_fr: loc.area_name_fr || ''
    })
    setIsModalOpen(true)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const payload = {
      city_name_ar: form.city_name_ar.trim(),
      city_name_fr: form.city_name_fr.trim(),
      area_name_ar: form.area_name_ar.trim() || '',
      area_name_fr: form.area_name_fr.trim() || ''
    }

    if (!payload.city_name_ar || !payload.city_name_fr) {
      return
    }

    if (editingLocation) {
      updateMut.mutate({ id: editingLocation.id, payload }, {
        onSuccess: () => setIsModalOpen(false)
      })
    } else {
      createMut.mutate(payload, {
        onSuccess: () => setIsModalOpen(false)
      })
    }
  }

  const columns: ColumnDef<Location>[] = [
    {
      header: 'ID',
      accessorKey: 'id',
      cell: ({ getValue }) => <span className="font-mono text-xs text-surface-muted">#{getValue<number>()}</span>
    },
    {
      header: 'المدينة / Nom',
      accessorKey: 'city_name_ar',
      cell: ({ row }) => (
        <div className="flex flex-col">
          <span className="font-bold text-white">{row.original.city_name_ar}</span>
          <span className="text-xs text-surface-muted" dir="ltr">{row.original.city_name_fr}</span>
        </div>
      )
    },
    {
      header: 'المنطقة / Zone',
      accessorKey: 'area_name_ar',
      cell: ({ row }) => (
        <div className="flex flex-col">
          <span className="text-sm text-white">{row.original.area_name_ar}</span>
          <span className="text-[10px] text-surface-muted uppercase tracking-wider" dir="ltr">{row.original.area_name_fr}</span>
        </div>
      )
    },
    {
      header: 'الإجراءات',
      id: 'actions',
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          <button
            onClick={() => openEdit(row.original)}
            className="p-1.5 rounded-lg text-blue-400 hover:bg-blue-500/10 transition-all"
          >
            <Pencil className="w-4 h-4" />
          </button>
          <button
            onClick={() => setDeleteId(row.original.id)}
            className="p-1.5 rounded-lg text-red-400 hover:bg-red-500/10 transition-all"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      )
    }
  ]

  return (
    <div className="animate-fade-in" dir="rtl">
      <PageHeader
        title="المواقع"
        subtitle="إدارة المدن والمناطق المتاحة في النظام"
        action={{ label: 'إضافة موقع جديد', icon: Plus, onClick: openAdd }}
      />

      {/* Filters */}
      <div className="mb-6 flex gap-4">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute right-4 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted" />
          <Input
            placeholder="البحث عن موقع..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            className="pr-11"
          />
        </div>
      </div>

      {isLoading ? (
        <div className="admin-card p-20 flex flex-col items-center justify-center gap-4">
          <Loader2 className="w-12 h-12 text-mazad-primary animate-spin" />
          <p className="text-surface-muted animate-pulse">جاري تحميل المواقع...</p>
        </div>
      ) : isError ? (
        <div className="admin-card p-20 text-center flex flex-col items-center gap-4">
          <AlertCircle className="w-12 h-12 text-red-500/20" />
          <p className="text-red-400 font-bold">فشل تحميل المواقع</p>
        </div>
      ) : (
        <div className="admin-card overflow-hidden">
          <DataTable columns={columns} data={filtered} />
        </div>
      )}

      {/* Add/Edit Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/80 backdrop-blur-sm" onClick={() => setIsModalOpen(false)} />
          <div className="relative bg-surface-card border border-surface-border w-full max-w-lg rounded-2xl shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
            <div className="px-6 py-4 border-b border-surface-border flex items-center justify-between">
              <h3 className="text-lg font-bold text-white">{editingLocation ? 'تعديل الموقع' : 'إضافة موقع جديد'}</h3>
              <button onClick={() => setIsModalOpen(false)} className="p-2 text-surface-muted hover:text-white transition-colors">
                <Plus className="w-5 h-5 rotate-45" />
              </button>
            </div>

            <form onSubmit={handleSubmit} className="p-6 space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Arabic Side */}
                <div className="space-y-4">
                  <div className="space-y-2">
                    <label className="text-xs text-mazad-primary font-bold block">اسم المدينة (عربي)</label>
                    <Input
                      required
                      value={form.city_name_ar}
                      onChange={e => setForm(f => ({ ...f, city_name_ar: e.target.value }))}
                      placeholder="مثال: نواكشوط"
                    />
                  </div>
                  <div className="space-y-2">
                    <label className="text-xs text-surface-muted font-bold block">المنطقة / الحي (عربي)</label>
                    <Input
                      value={form.area_name_ar}
                      onChange={e => setForm(f => ({ ...f, area_name_ar: e.target.value }))}
                      placeholder="مثال: تفرغ زينة"
                    />
                  </div>
                </div>

                {/* French Side */}
                <div className="space-y-4">
                  <div className="space-y-2">
                    <label className="text-xs text-mazad-primary font-bold block">المدينة (Français)</label>
                    <Input
                      required
                      dir="ltr"
                      value={form.city_name_fr}
                      onChange={e => setForm(f => ({ ...f, city_name_fr: e.target.value }))}
                      placeholder="Ex: Nouakchott"
                      className="text-left"
                    />
                  </div>
                  <div className="space-y-2">
                    <label className="text-xs text-surface-muted font-bold block">Zone / Quartier (Français)</label>
                    <Input
                      dir="ltr"
                      value={form.area_name_fr}
                      onChange={e => setForm(f => ({ ...f, area_name_fr: e.target.value }))}
                      placeholder="Ex: Tevragh Zeina"
                      className="text-left"
                    />
                  </div>
                </div>
              </div>

              <div className="pt-6 border-t border-surface-border flex justify-end gap-3">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-6 py-2.5 rounded-xl text-sm font-bold text-surface-muted border border-surface-border hover:bg-surface-border/50 transition-all"
                >
                  إلغاء
                </button>
                <button
                  type="submit"
                  disabled={createMut.isPending || updateMut.isPending}
                  className="px-8 py-2.5 rounded-xl text-sm font-bold bg-mazad-primary hover:bg-mazad-primary-dark text-white shadow-lg shadow-mazad-primary/20 disabled:opacity-50 flex items-center gap-2"
                >
                  {(createMut.isPending || updateMut.isPending) ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    'حفظ'
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {deleteId && (
        <div className="fixed inset-0 z-[110] flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/80 backdrop-blur-sm" onClick={() => setDeleteId(null)} />
          <div className="relative bg-surface-card border border-surface-border w-full max-w-sm rounded-2xl p-6 shadow-2xl animate-in zoom-in duration-200 text-center">
            <div className="w-16 h-16 bg-red-500/10 text-red-500 rounded-full flex items-center justify-center mx-auto mb-4">
              <Trash2 className="w-8 h-8" />
            </div>
            <h3 className="text-xl font-bold text-white mb-2">تأكيد الحذف</h3>
            <p className="text-surface-muted mb-6">
              هل أنت متأكد من حذف هذا الموقع؟ هذا الإجراء لا يمكن التراجع عنه.
            </p>
            <div className="flex gap-3">
              <button
                onClick={() => setDeleteId(null)}
                className="flex-1 px-4 py-2.5 rounded-xl text-sm font-bold text-surface-muted border border-surface-border hover:bg-surface-border/50 transition-all"
              >
                إلغاء
              </button>
              <button
                onClick={() => {
                  deleteMut.mutate(deleteId, {
                    onSuccess: () => setDeleteId(null)
                  })
                }}
                disabled={deleteMut.isPending}
                className="flex-1 px-4 py-2.5 rounded-xl text-sm font-bold bg-red-500 hover:bg-red-600 text-white shadow-lg shadow-red-500/20 transition-all disabled:opacity-50"
              >
                {deleteMut.isPending ? <Loader2 className="w-4 h-4 animate-spin mx-auto" /> : 'حذف الموقع'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
