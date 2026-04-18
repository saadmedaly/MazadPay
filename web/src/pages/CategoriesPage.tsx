import React, { useState } from 'react'
import {
  Plus, Search, Pencil, Trash2,
  AlertCircle, Loader2,  
} from 'lucide-react'
 import { useCategories, useCreateCategory, useUpdateCategory, useDeleteCategory } from '@/hooks/useMetadata'
 import { type ColumnDef } from '@tanstack/react-table'
import { type Category } from '@/types/api'
 import { PageHeader } from '@/components/shared/PageHeader'
 import { DataTable } from '@/components/shared/DataTable'
import { Input } from '@/components/ui/input'

export function CategoriesPage() {
  const { data: categories, isLoading, isError } = useCategories()
  const createMut = useCreateCategory()
  const updateMut = useUpdateCategory()
  const deleteMut = useDeleteCategory()

  const [search, setSearch] = useState('')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingCategory, setEditingCategory] = useState<Category | null>(null)
  const [form, setForm] = useState({
    name_ar: '',
    name_fr: '',
    name_en: '',
    icon_name: '',
    display_order: 0,
    parent_id: null as number | null
  })

  // Filter
  const filtered = categories?.filter(c => 
    c.id.toString().includes(search) ||
    c.name_ar.toLowerCase().includes(search.toLowerCase()) ||
    c.name_fr.toLowerCase().includes(search.toLowerCase())
  ) || []

  const openAdd = () => {
    setEditingCategory(null)
    setForm({ name_ar: '', name_fr: '', name_en: '', icon_name: '', display_order: 0, parent_id: null })
    setIsModalOpen(true)
  }

  const openEdit = (cat: Category) => {
    setEditingCategory(cat)
    setForm({
      name_ar: cat.name_ar,
      name_fr: cat.name_fr,
      name_en: cat.name_en || '',
      icon_name: cat.icon_name || '',
      display_order: cat.display_order,
      parent_id: cat.parent_id
    })
    setIsModalOpen(true)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (editingCategory) {
      updateMut.mutate({ id: editingCategory.id, payload: form }, {
        onSuccess: () => setIsModalOpen(false)
      })
    } else {
      createMut.mutate(form, {
        onSuccess: () => setIsModalOpen(false)
      })
    }
  }

  const columns: ColumnDef<Category>[] = [
    {
      header: 'ID',
      accessorKey: 'id',
      cell: ({ getValue }) => <span className="font-mono text-xs text-surface-muted">#{getValue<number>()}</span>
    },
    {
      header: 'الاسم (عربي)',
      accessorKey: 'name_ar',
      cell: ({ getValue }) => <span className="font-bold text-white">{getValue<string>()}</span>
    },
    {
      header: 'Nom (Français)',
      accessorKey: 'name_fr',
      cell: ({ getValue }) => <span className="text-surface-muted" dir="ltr">{getValue<string>()}</span>
    },
    {
      header: 'Name (English)',
      accessorKey: 'name_en',
      cell: ({ getValue }) => <span className="text-surface-muted" dir="ltr">{getValue<string>()}</span>
    },
    {
      header: 'الترتيب',
      accessorKey: 'display_order',
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
            onClick={() => {
              if (confirm('هل أنت متأكد من حذف هذه الفئة؟')) {
                deleteMut.mutate(row.original.id)
              }
            }}
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
        title="الفئات"
        subtitle="إدارة فئات المزادات المتاحة في النظام"
        action={{ label: 'إضافة فئة جديدة', icon: Plus, onClick: openAdd }}
      />

      {/* Filters */}
      <div className="mb-6 flex gap-4">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute right-4 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted" />
          <Input
            placeholder="البحث عن فئة..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            className="pr-11"
          />
        </div>
      </div>

      {isLoading ? (
        <div className="admin-card p-20 flex flex-col items-center justify-center gap-4">
          <Loader2 className="w-12 h-12 text-mazad-primary animate-spin" />
          <p className="text-surface-muted animate-pulse">جاري تحميل الفئات...</p>
        </div>
      ) : isError ? (
        <div className="admin-card p-20 text-center flex flex-col items-center gap-4">
          <AlertCircle className="w-12 h-12 text-red-500/20" />
          <p className="text-red-400 font-bold">فشل تحميل الفئات</p>
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
              <h3 className="text-lg font-bold text-white">{editingCategory ? 'تعديل الفئة' : 'إضافة فئة جديدة'}</h3>
              <button onClick={() => setIsModalOpen(false)} className="p-2 text-surface-muted hover:text-white transition-colors">
                <Plus className="w-5 h-5 rotate-45" />
              </button>
            </div>

            <form onSubmit={handleSubmit} className="p-6 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label className="text-xs text-surface-muted font-bold block">الاسم (بالعربية)</label>
                  <Input
                    required
                    value={form.name_ar}
                    onChange={e => setForm(f => ({ ...f, name_ar: e.target.value }))}
                    placeholder="مثال: سيارات"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-xs text-surface-muted font-bold block">Nom (Français)</label>
                  <Input
                    required
                    dir="ltr"
                    value={form.name_fr}
                    onChange={e => setForm(f => ({ ...f, name_fr: e.target.value }))}
                    placeholder="Ex: Voitures"
                    className="text-left"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-xs text-surface-muted font-bold block">Name (English)</label>
                  <Input
                    required
                    dir="ltr"
                    value={form.name_en}
                    onChange={e => setForm(f => ({ ...f, name_en: e.target.value }))}
                    placeholder="Ex: Cars"
                    className="text-left"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label className="text-xs text-surface-muted font-bold block">أيقونة (اختياري)</label>
                  <Input
                    value={form.icon_name}
                    onChange={e => setForm(f => ({ ...f, icon_name: e.target.value }))}
                    placeholder="car, home, tech..."
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-xs text-surface-muted font-bold block">الترتيب</label>
                  <Input
                    type="number"
                    value={form.display_order}
                    onChange={e => setForm(f => ({ ...f, display_order: parseInt(e.target.value) || 0 }))}
                  />
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
    </div>
  )
}
