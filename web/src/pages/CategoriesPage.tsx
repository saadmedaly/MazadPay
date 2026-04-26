import React, { useState, useMemo } from 'react'
import {
  Plus, Search, Pencil, Trash2,
  AlertCircle, Loader2, ChevronRight, Folder,
  Image as ImageIcon, Link, X, Eye
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
    parent_id: null as number | null,
    image_url: '' as string
  })
  const [previewImage, setPreviewImage] = useState<string | null>(null)

  // Build category hierarchy
  const categoryTree = useMemo(() => {
    if (!categories) return []
    
    const parentCategories = categories.filter(c => !c.parent_id).sort((a, b) => a.display_order - b.display_order)
    const childCategories = categories.filter(c => c.parent_id).sort((a, b) => a.display_order - b.display_order)
    
    return parentCategories.map(parent => ({
      ...parent,
      children: childCategories.filter(child => child.parent_id === parent.id)
    }))
  }, [categories])

  // Flatten tree for display with level indicator
  const flattenedCategories = useMemo(() => {
    const result: (Category & { level: number; parentName?: string })[] = []
    
    categoryTree.forEach(parent => {
      result.push({ ...parent, level: 0 })
      parent.children?.forEach(child => {
        result.push({ 
          ...child, 
          level: 1, 
          parentName: parent.name_ar 
        })
      })
    })
    
    // Filter
    if (!search) return result
    return result.filter(c => 
      c.id.toString().includes(search) ||
      c.name_ar.toLowerCase().includes(search.toLowerCase()) ||
      c.name_fr.toLowerCase().includes(search.toLowerCase()) ||
      c.parentName?.toLowerCase().includes(search.toLowerCase())
    )
  }, [categoryTree, search])

  // Get only parent categories for dropdown
  const parentCategories = useMemo(() => {
    return categories?.filter(c => !c.parent_id).sort((a, b) => a.display_order - b.display_order) || []
  }, [categories])

  const openAdd = () => {
    setEditingCategory(null)
    setForm({ name_ar: '', name_fr: '', name_en: '', icon_name: '', display_order: 0, parent_id: null, image_url: '' })
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
      parent_id: cat.parent_id,
      image_url: cat.image_url || ''
    })
    setIsModalOpen(true)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const payload = {
      ...form,
      name_ar: form.name_ar.trim(),
      name_fr: form.name_fr.trim(),
      name_en: form.name_en.trim(),
      icon_name: form.icon_name.trim(),
      image_url: form.image_url.trim() || null,
      is_active: true,
    }

    if (!payload.name_ar || !payload.name_fr) {
      return // browser check 'required' handles this usually, but safe to trim
    }

    if (editingCategory) {
      updateMut.mutate({ id: editingCategory.id, payload }, {
        onSuccess: () => setIsModalOpen(false)
      })
    } else {
      createMut.mutate(payload, {
        onSuccess: () => setIsModalOpen(false)
      })
    }
  }

  const columns: ColumnDef<Category & { level: number; parentName?: string }>[] = [
    {
      header: 'ID',
      accessorKey: 'id',
      cell: ({ row }) => {
        const level = row.original.level
        return (
          <div className="flex items-center gap-2">
            {level === 0 ? (
              <Folder className="w-4 h-4 text-mazad-primary" />
            ) : (
              <ChevronRight className="w-4 h-4 text-surface-muted mr-4" />
            )}
            <span className="font-mono text-xs text-surface-muted">#{row.original.id}</span>
          </div>
        )
      }
    },
    {
      header: 'الاسم (عربي)',
      accessorKey: 'name_ar',
      cell: ({ row }) => {
        const level = row.original.level
        const hasChildren = level === 0 && (row.original as any).children?.length > 0
        return (
          <div className="flex items-center gap-2">
            <span className={`font-bold ${level === 0 ? 'text-white text-base' : 'text-surface-muted text-sm'}`}>
              {row.original.name_ar}
            </span>
            {hasChildren && (
              <span className="text-[10px] bg-mazad-primary/20 text-mazad-primary px-2 py-0.5 rounded-full">
                {(row.original as any).children?.length} فرعية
              </span>
            )}
            {level === 1 && row.original.parentName && (
              <span className="text-[10px] text-surface-muted">
                ← {row.original.parentName}
              </span>
            )}
          </div>
        )
      }
    },
    {
      header: 'Nom (Français)',
      accessorKey: 'name_fr',
      cell: ({ row }) => (
        <span className={`${row.original.level === 0 ? 'text-white' : 'text-surface-muted text-sm'}`} dir="ltr">
          {row.original.name_fr}
        </span>
      )
    },
    {
      header: 'النوع',
      accessorKey: 'level',
      cell: ({ row }) => (
        <span className={`text-xs px-2 py-1 rounded-full ${
          row.original.level === 0 
            ? 'bg-blue-500/20 text-blue-400' 
            : 'bg-emerald-500/20 text-emerald-400'
        }`}>
          {row.original.level === 0 ? 'فئة رئيسية' : 'فئة فرعية'}
        </span>
      )
    },
    {
      header: 'الصورة',
      id: 'image',
      cell: ({ row }) => {
        const imgUrl = row.original.image_url
        return imgUrl ? (
          <button
            onClick={() => setPreviewImage(imgUrl)}
            className="relative group"
          >
            <img
              src={imgUrl}
              alt={row.original.name_ar}
              className="w-10 h-10 rounded-lg object-cover border border-surface-border"
            />
            <div className="absolute inset-0 bg-black/50 rounded-lg flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
              <Eye className="w-4 h-4 text-white" />
            </div>
          </button>
        ) : (
          <div className="w-10 h-10 rounded-lg bg-surface-base border border-surface-border flex items-center justify-center">
            <ImageIcon className="w-4 h-4 text-surface-muted" />
          </div>
        )
      }
    },
    {
      header: 'الترتيب',
      accessorKey: 'display_order',
      cell: ({ row }) => (
        <span className="text-surface-muted">{row.original.display_order}</span>
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
            onClick={() => {
              const hasChildren = (row.original as any).children?.length > 0
              const msg = hasChildren 
                ? 'هذه الفئة تحتوي على فئات فرعية. حذفها سيحذف جميع الفئات الفرعية أيضاً. هل أنت متأكد؟'
                : 'هل أنت متأكد من حذف هذه الفئة؟'
              if (confirm(msg)) {
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
          <DataTable columns={columns} data={flattenedCategories} />
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
                  <label className="text-xs text-surface-muted font-bold block">الفئة الأب (اختياري)</label>
                  <select
                    value={form.parent_id || ''}
                    onChange={e => setForm(f => ({ ...f, parent_id: e.target.value ? parseInt(e.target.value) : null }))}
                    className="w-full bg-surface-base border border-surface-border rounded-xl p-3 text-sm text-white focus:border-mazad-primary/60 outline-none"
                  >
                    <option value="">فئة رئيسية (بدون أب)</option>
                    {parentCategories.map(parent => (
                      <option key={parent.id} value={parent.id}>
                        {parent.name_ar} (ID: {parent.id})
                      </option>
                    ))}
                  </select>
                  <p className="text-[10px] text-surface-muted">
                    اتركه فارغاً لإنشاء فئة رئيسية، أو اختر فئة أب لإنشاء فئة فرعية
                  </p>
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

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label className="text-xs text-surface-muted font-bold block">أيقونة (اختياري)</label>
                  <Input
                    value={form.icon_name}
                    onChange={e => setForm(f => ({ ...f, icon_name: e.target.value }))}
                    placeholder="car, home, tech..."
                  />
                </div>
              </div>

              {/* Image URL */}
              <div className="space-y-2">
                <label className="text-xs text-surface-muted font-bold flex items-center gap-1.5">
                  <ImageIcon className="w-3 h-3" />
                  رابط الصورة (اختياري)
                </label>
                <div className="flex items-center gap-2">
                  <div className="relative flex-1">
                    <Link className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted" />
                    <Input
                      dir="ltr"
                      value={form.image_url}
                      onChange={e => setForm(f => ({ ...f, image_url: e.target.value }))}
                      placeholder="https://example.com/image.png"
                      className="pl-10 text-left"
                    />
                  </div>
                  {form.image_url && (
                    <button
                      type="button"
                      onClick={() => setForm(f => ({ ...f, image_url: '' }))}
                      className="p-2 rounded-lg text-red-400 hover:bg-red-500/10 transition-all"
                      title="حذف الصورة"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  )}
                </div>
                {/* Image Preview */}
                {form.image_url && (
                  <div className="mt-2 relative group inline-block">
                    <img
                      src={form.image_url}
                      alt="معاينة"
                      className="w-24 h-24 rounded-xl object-cover border border-surface-border"
                      onError={e => { (e.target as HTMLImageElement).style.display = 'none' }}
                    />
                    <div className="absolute inset-0 bg-black/50 rounded-xl flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity gap-1">
                      <button
                        type="button"
                        onClick={() => setPreviewImage(form.image_url)}
                        className="p-1 rounded text-white hover:bg-white/20"
                        title="عرض الصورة"
                      >
                        <Eye className="w-4 h-4" />
                      </button>
                      <button
                        type="button"
                        onClick={() => setForm(f => ({ ...f, image_url: '' }))}
                        className="p-1 rounded text-white hover:bg-white/20"
                        title="حذف الصورة"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                )}
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

      {/* Image Preview Overlay */}
      {previewImage && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/90 backdrop-blur-sm" onClick={() => setPreviewImage(null)} />
          <div className="relative max-w-2xl max-h-[80vh]">
            <img
              src={previewImage}
              alt="معاينة الصورة"
              className="max-w-full max-h-[80vh] rounded-2xl object-contain shadow-2xl"
            />
            <button
              onClick={() => setPreviewImage(null)}
              className="absolute -top-3 -right-3 p-2 bg-surface-card border border-surface-border rounded-full text-white hover:bg-red-500/20 hover:text-red-400 transition-all"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
