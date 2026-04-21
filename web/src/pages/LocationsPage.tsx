import React, { useState } from 'react'
import {
  Plus, Search, Pencil, Trash2, Globe, MapPin,
  AlertCircle, Loader2
} from 'lucide-react'
import { useLocations, useCreateLocation, useUpdateLocation, useDeleteLocation } from '@/hooks/useMetadata'
import { useCountries, useCreateCountry, useUpdateCountry, useDeleteCountry } from '@/hooks/useCountries'
 import { type ColumnDef } from '@tanstack/react-table'
import { type Location, type Country } from '@/types/api'
import { PageHeader } from '@/components/shared/PageHeader'
import { DataTable } from '@/components/shared/DataTable'
import { Input } from '@/components/ui/input'
import { toast } from 'sonner'

export function LocationsPage() {
  const { data: locations, isLoading, isError } = useLocations()
  const { data: countries, isLoading: isLoadingCountries } = useCountries()
  const createMut = useCreateLocation()
  const updateMut = useUpdateLocation()
  const deleteMut = useDeleteLocation()
  const createCountryMut = useCreateCountry()
  const updateCountryMut = useUpdateCountry()
  const deleteCountryMut = useDeleteCountry()

  const [activeTab, setActiveTab] = useState<'locations' | 'countries'>('locations')
  const [search, setSearch] = useState('')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingLocation, setEditingLocation] = useState<Location | null>(null)
  const [editingCountry, setEditingCountry] = useState<Country | null>(null)
  const [form, setForm] = useState({
    city_name_ar: '',
    city_name_fr: '',
    area_name_ar: '',
    area_name_fr: '',
    country_id: ''
  })
  const [countryForm, setCountryForm] = useState({
    code: '',
    name_ar: '',
    name_fr: '',
    name_en: '',
    flag_emoji: ''
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

  // Filter countries
  const filteredCountries = countries?.filter(c => 
    c.id.toString().includes(search) ||
    c.code.toLowerCase().includes(search.toLowerCase()) ||
    c.name_ar.toLowerCase().includes(search.toLowerCase()) ||
    c.name_fr.toLowerCase().includes(search.toLowerCase()) ||
    c.name_en.toLowerCase().includes(search.toLowerCase())
  ) || []

  // Country handlers
  const openAddCountry = () => {
    setEditingCountry(null)
    setCountryForm({ code: '', name_ar: '', name_fr: '', name_en: '', flag_emoji: '' })
    setIsModalOpen(true)
  }

  const openEditCountry = (country: Country) => {
    setEditingCountry(country)
    setCountryForm({
      code: country.code,
      name_ar: country.name_ar,
      name_fr: country.name_fr,
      name_en: country.name_en,
      flag_emoji: country.flag_emoji
    })
    setIsModalOpen(true)
  }

  const handleCountrySubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const payload = {
      code: countryForm.code.trim(),
      name_ar: countryForm.name_ar.trim(),
      name_fr: countryForm.name_fr.trim(),
      name_en: countryForm.name_en.trim(),
      flag_emoji: countryForm.flag_emoji.trim()
    }

    if (!payload.code || !payload.name_ar || !payload.name_fr || !payload.name_en) {
      return
    }

    if (editingCountry) {
      updateCountryMut.mutate({ id: editingCountry.id, payload }, {
        onSuccess: () => {
          setIsModalOpen(false)
          toast.success('Country updated successfully')
        }
      })
    } else {
      createCountryMut.mutate(payload, {
        onSuccess: () => {
          setIsModalOpen(false)
          toast.success('Country created successfully')
        }
      })
    }
  }

  const handleDeleteCountry = (id: number) => {
    deleteCountryMut.mutate(id, {
      onSuccess: () => {
        toast.success('Country deleted successfully')
      }
    })
  }

  const openAdd = () => {
    setEditingLocation(null)
    setForm({ city_name_ar: '', city_name_fr: '', area_name_ar: '', area_name_fr: '', country_id: '' })
    setIsModalOpen(true)
  }

  const openEdit = (loc: Location) => {
    setEditingLocation(loc)
    setForm({
      city_name_ar: loc.city_name_ar,
      city_name_fr: loc.city_name_fr,
      area_name_ar: loc.area_name_ar || '',
      area_name_fr: loc.area_name_fr || '',
      country_id: loc.country_id?.toString() || ''
    })
    setIsModalOpen(true)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const payload = {
      city_name_ar: form.city_name_ar.trim(),
      city_name_fr: form.city_name_fr.trim(),
      area_name_ar: form.area_name_ar.trim() || '',
      area_name_fr: form.area_name_fr.trim() || '',
      country_id: form.country_id ? parseInt(form.country_id) : null
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
      header: 'الدولة',
      accessorKey: 'country_id',
      cell: ({ row }) => {
        const country = countries?.find(c => c.id === row.original.country_id)
        return (
          <div className="flex items-center gap-2">
            <span className="text-lg">{country?.flag_emoji || '🌍'}</span>
            <span className="text-sm text-white">{country?.name_ar || 'غير محدد'}</span>
          </div>
        )
      }
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

  const countryColumns: ColumnDef<Country>[] = [
    {
      header: 'ID',
      accessorKey: 'id',
      cell: ({ getValue }) => <span className="font-mono text-xs text-surface-muted">#{getValue<number>()}</span>
    },
    {
      header: 'Code / Flag',
      accessorKey: 'code',
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          <span className="text-lg">{row.original.flag_emoji}</span>
          <span className="font-mono font-bold text-white">{row.original.code}</span>
        </div>
      )
    },
    {
      header: 'Country Name',
      accessorKey: 'name_ar',
      cell: ({ row }) => (
        <div className="flex flex-col">
          <span className="font-bold text-white">{row.original.name_ar}</span>
          <span className="text-xs text-surface-muted" dir="ltr">{row.original.name_fr}</span>
          <span className="text-xs text-surface-muted" dir="ltr">{row.original.name_en}</span>
        </div>
      )
    },
    {
      id: 'actions',
      header: '',
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          <button
            onClick={() => openEditCountry(row.original)}
            className="p-2 rounded-xl text-mazad-primary border border-transparent hover:border-mazad-primary/20 hover:bg-mazad-primary/10 transition-all"
            title="Edit"
          >
            <Pencil className="w-4 h-4" />
          </button>
          <button
            onClick={() => handleDeleteCountry(row.original.id)}
            className="p-2 rounded-xl text-red-400 border border-transparent hover:border-red-500/20 hover:bg-red-500/10 transition-all"
            title="Delete"
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
        title="الدول والمناطق"
        subtitle="إدارة الدول والمناطق"
      />

      {/* Tab Navigation */}
      <div className="mb-6 flex gap-2 border-b border-surface-border">
        <button
          onClick={() => setActiveTab('locations')}
          className={`px-4 py-2 font-bold text-sm transition-all border-b-2 ${
            activeTab === 'locations'
              ? 'text-mazad-primary border-mazad-primary'
              : 'text-surface-muted border-transparent hover:text-white'
          }`}
        >
          <MapPin className="w-4 h-4 inline mr-2" />
          المواقع
        </button>
        <button
          onClick={() => setActiveTab('countries')}
          className={`px-4 py-2 font-bold text-sm transition-all border-b-2 ${
            activeTab === 'countries'
              ? 'text-mazad-primary border-mazad-primary'
              : 'text-surface-muted border-transparent hover:text-white'
          }`}
        >
          <Globe className="w-4 h-4 inline mr-2" />
          الدول
        </button>
      </div>

      {/* Locations Tab */}
      {activeTab === 'locations' && (
        <div>
          <div className="mb-6 flex items-center justify-between">
            <div className="relative max-w-md">
              <Search className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted" />
              <Input
                placeholder="Search locations..."
                value={search}
                onChange={e => setSearch(e.target.value)}
                className="pr-10"
              />
            </div>
            <button
              onClick={openAdd}
              className="flex items-center gap-2 bg-mazad-primary hover:bg-mazad-primary/90 px-4 py-2.5 rounded-xl text-sm font-bold text-white shadow-lg shadow-mazad-primary/20 active:scale-95 transition-all"
            >
              <Plus className="w-4 h-4" />
              إضافة موقع
            </button>
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
        </div>
      )}

      {/* Countries Tab */}
      {activeTab === 'countries' && (
        <div>
          <div className="mb-6 flex items-center justify-between">
            <div className="relative max-w-md">
              <Search className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted" />
              <Input
                placeholder="Search countries..."
                value={search}
                onChange={e => setSearch(e.target.value)}
                className="pr-10"
              />
            </div>
            <button
              onClick={openAddCountry}
              className="flex items-center gap-2 bg-mazad-primary hover:bg-mazad-primary/90 px-4 py-2.5 rounded-xl text-sm font-bold text-white shadow-lg shadow-mazad-primary/20 active:scale-95 transition-all"
            >
              <Plus className="w-4 h-4" />
              إضافة دولة
            </button>
          </div>

          {isLoadingCountries ? (
            <div className="admin-card p-20 flex flex-col items-center justify-center gap-4">
              <Loader2 className="w-12 h-12 text-mazad-primary animate-spin" />
              <p className="text-surface-muted animate-pulse">جاري تحميل الدول...</p>
            </div>
          ) : (
            <div className="admin-card overflow-hidden">
              <DataTable columns={countryColumns} data={filteredCountries} />
            </div>
          )}
        </div>
      )}

      {/* Add/Edit Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/80 backdrop-blur-sm" onClick={() => setIsModalOpen(false)} />
          <div className="relative bg-surface-card border border-surface-border w-full max-w-lg rounded-2xl shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
            <div className="px-6 py-4 border-b border-surface-border flex items-center justify-between">
              <h3 className="text-lg font-bold text-white">
                {editingCountry ? 'تعديل الدولة' : editingLocation ? 'تعديل الموقع' : 
                 activeTab === 'countries' ? 'إضافة دولة جديدة' : 'إضافة موقع جديد'}
              </h3>
              <button onClick={() => setIsModalOpen(false)} className="p-2 text-surface-muted hover:text-white transition-colors">
                <Plus className="w-5 h-5 rotate-45" />
              </button>
            </div>

            {/* Location Form */}
            {editingLocation || activeTab === 'locations' ? (
              <form onSubmit={handleSubmit} className="p-6 space-y-6">
                <div className="mb-6">
                  <label className="text-xs text-mazad-primary font-bold block">اختر الدولة</label>
                  <select
                    value={form.country_id}
                    onChange={e => setForm(f => ({ ...f, country_id: e.target.value }))}
                    className="w-full px-4 py-3 bg-surface-input border border-surface-border rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-mazad-primary/50 focus:border-mazad-primary"
                  >
                    <option value="">اختر الدولة...</option>
                    {countries?.map(country => (
                      <option key={country.id} value={country.id}>
                        {country.flag_emoji} {country.name_ar}
                      </option>
                    ))}
                  </select>
                </div>
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
            ) : (
              /* Country Form */
              <form onSubmit={handleCountrySubmit} className="p-6 space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  {/* Left Side */}
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <label className="text-xs text-mazad-primary font-bold block">رمز الدولة</label>
                      <Input
                        required
                        maxLength={2}
                        value={countryForm.code}
                        onChange={e => setCountryForm(f => ({ ...f, code: e.target.value.toUpperCase() }))}
                        placeholder="MR"
                        className="font-mono uppercase"
                      />
                    </div>
                    <div className="space-y-2">
                      <label className="text-xs text-mazad-primary font-bold block">اسم الدولة (عربي)</label>
                      <Input
                        required
                        value={countryForm.name_ar}
                        onChange={e => setCountryForm(f => ({ ...f, name_ar: e.target.value }))}
                        placeholder="موريتانيا"
                      />
                    </div>
                  </div>

                  {/* Right Side */}
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <label className="text-xs text-mazad-primary font-bold block">Country Name (Français)</label>
                      <Input
                        required
                        dir="ltr"
                        value={countryForm.name_fr}
                        onChange={e => setCountryForm(f => ({ ...f, name_fr: e.target.value }))}
                        placeholder="Mauritanie"
                        className="text-left"
                      />
                    </div>
                    <div className="space-y-2">
                      <label className="text-xs text-mazad-primary font-bold block">Country Name (English)</label>
                      <Input
                        required
                        dir="ltr"
                        value={countryForm.name_en}
                        onChange={e => setCountryForm(f => ({ ...f, name_en: e.target.value }))}
                        placeholder="Mauritania"
                        className="text-left"
                      />
                    </div>
                    <div className="space-y-2">
                      <label className="text-xs text-mazad-primary font-bold block">علم الدولة</label>
                      <Input
                        value={countryForm.flag_emoji}
                        onChange={e => setCountryForm(f => ({ ...f, flag_emoji: e.target.value }))}
                        placeholder="🇲🇷"
                        className="text-2xl text-center"
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
                    disabled={createCountryMut.isPending || updateCountryMut.isPending}
                    className="px-8 py-2.5 rounded-xl text-sm font-bold bg-mazad-primary hover:bg-mazad-primary-dark text-white shadow-lg shadow-mazad-primary/20 disabled:opacity-50 flex items-center gap-2"
                  >
                    {(createCountryMut.isPending || updateCountryMut.isPending) ? (
                      <Loader2 className="w-4 h-4 animate-spin" />
                    ) : (
                      'حفظ'
                    )}
                  </button>
                </div>
              </form>
            )}
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
