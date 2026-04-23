import React, { useState } from 'react'
import {
  Plus, Search, Pencil, Trash2, CreditCard,
  ToggleLeft, ToggleRight,
} from 'lucide-react'
import { type ColumnDef } from '@tanstack/react-table'
import { PageHeader } from '@/components/shared/PageHeader'
import { DataTable } from '@/components/shared/DataTable'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Modal } from '@/components/shared/Modal'
import { Label } from '@/components/ui/label'
import { StatusBadge } from '@/components/shared/StatusBadge'
import {
  usePaymentMethods,
  useCreatePaymentMethod,
  useUpdatePaymentMethod,
  useDeletePaymentMethod,
  useTogglePaymentMethodStatus,
  type PaymentMethod
} from '@/hooks/usePaymentMethods'

interface PaymentMethodForm {
  code: string
  name_ar: string
  name_fr: string
  name_en: string
  logo_url: string
  is_active: boolean
  country_id: number | null
}

export function PaymentMethodsPage() {
  const [search, setSearch] = useState('')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingMethod, setEditingMethod] = useState<PaymentMethod | null>(null)
  const [form, setForm] = useState<PaymentMethodForm>({
    code: '',
    name_ar: '',
    name_fr: '',
    name_en: '',
    logo_url: '',
    country_id: null as number | null,
    is_active: true
  })

  const { data: paymentMethods = [], isLoading } = usePaymentMethods()
  const createMutation = useCreatePaymentMethod()
  const updateMutation = useUpdatePaymentMethod()
  const deleteMutation = useDeletePaymentMethod()
  const toggleMutation = useTogglePaymentMethodStatus()

  const filtered = paymentMethods.filter(m =>
    m.id.toString().includes(search) ||
    m.code.toLowerCase().includes(search.toLowerCase()) ||
    m.name_ar.toLowerCase().includes(search.toLowerCase()) ||
    m.name_fr.toLowerCase().includes(search.toLowerCase())
  )

  const openAdd = () => {
    setEditingMethod(null)
    setForm({ code: '', name_ar: '', name_fr: '', name_en: '', logo_url: '', country_id: null, is_active: true })
    setIsModalOpen(true)
  }

  const openEdit = (method: PaymentMethod) => {
    setEditingMethod(method)
    setForm({
      code: method.code,
      name_ar: method.name_ar,
      name_fr: method.name_fr,
      name_en: method.name_en || '',
      logo_url: method.logo_url || '',
      country_id: method.country_id || null,
      is_active: method.is_active
    })
    setIsModalOpen(true)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const data = {
      code: form.code,
      name_ar: form.name_ar,
      name_fr: form.name_fr,
      name_en: form.name_en || undefined,
      logo_url: form.logo_url || undefined,
      country_id: form.country_id || undefined,
      is_active: form.is_active
    }

    if (editingMethod) {
      updateMutation.mutate({ id: editingMethod.id, data })
    } else {
      createMutation.mutate(data as Omit<PaymentMethod, 'id' | 'created_at'>)
    }
    setIsModalOpen(false)
  }

  const toggleStatus = (method: PaymentMethod) => {
    toggleMutation.mutate(method.id)
  }

  const deleteMethod = (method: PaymentMethod) => {
    if (confirm(`Voulez-vous vraiment supprimer ${method.name_fr}?`)) {
      deleteMutation.mutate(method.id)
    }
  }

  const columns: ColumnDef<PaymentMethod>[] = [
    {
      header: 'ID',
      accessorKey: 'id',
      cell: ({ getValue }) => <span className="font-mono text-xs text-surface-muted">#{getValue<number>()}</span>
    },
    {
      header: 'Code',
      accessorKey: 'code',
      cell: ({ getValue }) => <span className="font-mono text-sm">{getValue<string>()}</span>
    },
    {
      header: 'الاسم (عربي)',
      accessorKey: 'name_ar',
      cell: ({ getValue }) => <span className="font-bold text-white">{getValue<string>()}</span>
    },
    {
      header: 'Nom (Français)',
      accessorKey: 'name_fr',
      cell: ({ getValue }) => <span className="text-surface-muted">{getValue<string>()}</span>
    },
    {
      header: 'Name (English)',
      accessorKey: 'name_en',
      cell: ({ getValue }) => <span className="text-surface-muted">{getValue<string>() || '-'}</span>
    },
    {
      header: 'Statut',
      accessorKey: 'is_active',
      cell: ({ getValue }) => (
        <StatusBadge status={getValue<boolean>() ? 'active' : 'inactive'} />
      )
    },
    {
      header: 'Actions',
      cell: ({ row }) => (
        <div className="flex gap-2">
          <Button
            size="sm"
            variant="ghost"
            onClick={() => toggleStatus(row.original)}
            title={row.original.is_active ? 'Désactiver' : 'Activer'}
          >
            {row.original.is_active ? <ToggleRight className="w-4 h-4 text-green-500" /> : <ToggleLeft className="w-4 h-4 text-surface-muted" />}
          </Button>
          <Button
            size="sm"
            variant="ghost"
            onClick={() => openEdit(row.original)}
          >
            <Pencil className="w-4 h-4" />
          </Button>
          <Button
            size="sm"
            variant="ghost"
            onClick={() => deleteMethod(row.original)}
          >
            <Trash2 className="w-4 h-4 text-red-500" />
          </Button>
        </div>
      )
    }
  ]

  return (
    <div className="p-6 space-y-6">
      <PageHeader
        title="طرق الدفع"
        subtitle="إدارة طرق الدفع المتاحة"
        icon={CreditCard}
      />

      <div className="flex gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted" />
          <Input
            placeholder="بحث..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-10"
          />
        </div>
        <Button onClick={openAdd}>
          <Plus className="w-4 h-4 mr-2" />
          إضافة
        </Button>
      </div>

      <DataTable
        columns={columns}
        data={filtered}
        isLoading={isLoading}
      />

      <Modal isOpen={isModalOpen} onOpenChange={setIsModalOpen} title={editingMethod ? 'تعديل طريقة الدفع' : 'طريقة دفع جديدة'}>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="code">الرمز *</Label>
            <Input
              id="code"
              value={form.code}
              onChange={(e) => setForm({ ...form, code: e.target.value })}
              placeholder="مثال: masrvi"
              required
            />
          </div>

          <div>
            <Label htmlFor="name_ar">الاسم بالعربية *</Label>
            <Input
              id="name_ar"
              value={form.name_ar}
              onChange={(e) => setForm({ ...form, name_ar: e.target.value })}
              placeholder="مصروفي"
              required
              dir="rtl"
            />
          </div>

          <div>
            <Label htmlFor="name_fr">الاسم بالفرنسية *</Label>
            <Input
              id="name_fr"
              value={form.name_fr}
              onChange={(e) => setForm({ ...form, name_fr: e.target.value })}
              placeholder="Masrivi"
              required
            />
          </div>

          <div>
            <Label htmlFor="name_en">الاسم بالإنجليزية</Label>
            <Input
              id="name_en"
              value={form.name_en}
              onChange={(e) => setForm({ ...form, name_en: e.target.value })}
              placeholder="Masrivi"
            />
          </div>

          <div>
            <Label htmlFor="logo_url">رابط الشعار</Label>
            <Input
              id="logo_url"
              value={form.logo_url}
              onChange={(e) => setForm({ ...form, logo_url: e.target.value })}
              placeholder="https://..."
            />
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="is_active"
              checked={form.is_active}
              onChange={(e) => setForm({ ...form, is_active: e.target.checked })}
              className="w-4 h-4"
            />
            <Label htmlFor="is_active">نشط</Label>
          </div>

          <div className="flex justify-end gap-2 pt-4">
            <Button type="button" variant="outline" onClick={() => setIsModalOpen(false)}>
              إلغاء
            </Button>
            <Button type="submit">
              {editingMethod ? 'تعديل' : 'إنشاء'}
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  )
}
