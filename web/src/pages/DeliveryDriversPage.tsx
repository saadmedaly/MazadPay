import React, { useState } from 'react'
import {
  Plus, Search, Pencil, Trash2, Truck,
  MapPin, ToggleLeft, ToggleRight, Star,
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
  useDeliveryDrivers,
  useRegisterDriver,
  useUpdateDriver,
  useDeleteDriver,
  type DeliveryDriver
} from '@/hooks/useDeliveryDrivers'

interface DriverForm {
  user_id: string
  vehicle_type: string
  vehicle_plate: string
  vehicle_color: string
  license_number: string
  is_available: boolean
}

export function DeliveryDriversPage() {
  const [search, setSearch] = useState('')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingDriver, setEditingDriver] = useState<DeliveryDriver | null>(null)
  const [form, setForm] = useState<DriverForm>({
    user_id: '',
    vehicle_type: '',
    vehicle_plate: '',
    vehicle_color: '',
    license_number: '',
    is_available: true
  })

  const { data: drivers = [], isLoading } = useDeliveryDrivers()
  const registerMutation = useRegisterDriver()
  const updateMutation = useUpdateDriver()
  const deleteMutation = useDeleteDriver()

  const filtered = drivers.filter(d =>
    d.id.includes(search) ||
    d.vehicle_plate?.toLowerCase().includes(search.toLowerCase()) ||
    d.vehicle_type?.toLowerCase().includes(search.toLowerCase()) ||
    d.license_number?.toLowerCase().includes(search.toLowerCase())
  )

  const openAdd = () => {
    setEditingDriver(null)
    setForm({ user_id: '', vehicle_type: '', vehicle_plate: '', vehicle_color: '', license_number: '', is_available: true })
    setIsModalOpen(true)
  }

  const openEdit = (driver: DeliveryDriver) => {
    setEditingDriver(driver)
    setForm({
      user_id: driver.user_id || '',
      vehicle_type: driver.vehicle_type || '',
      vehicle_plate: driver.vehicle_plate || '',
      vehicle_color: driver.vehicle_color || '',
      license_number: driver.license_number || '',
      is_available: driver.is_available
    })
    setIsModalOpen(true)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (editingDriver) {
      updateMutation.mutate({
        id: editingDriver.id,
        data: {
          vehicle_type: form.vehicle_type,
          vehicle_plate: form.vehicle_plate,
          vehicle_color: form.vehicle_color,
          license_number: form.license_number,
          is_available: form.is_available
        }
      })
    } else {
      registerMutation.mutate({
        user_id: form.user_id,
        vehicle_type: form.vehicle_type,
        vehicle_plate: form.vehicle_plate,
        vehicle_color: form.vehicle_color,
        license_number: form.license_number
      })
    }
    setIsModalOpen(false)
  }

  const toggleAvailability = (driver: DeliveryDriver) => {
    updateMutation.mutate({
      id: driver.id,
      data: { is_available: !driver.is_available }
    })
  }

  const deleteDriver = (driver: DeliveryDriver) => {
    if (confirm(`Voulez-vous vraiment supprimer ce chauffeur?`)) {
      deleteMutation.mutate(driver.id)
    }
  }

  const columns: ColumnDef<DeliveryDriver>[] = [
    {
      header: 'ID',
      accessorKey: 'id',
      cell: ({ getValue }) => <span className="font-mono text-xs text-surface-muted">{getValue<string>().slice(0, 8)}...</span>
    },
    {
      header: 'Type Véhicule',
      accessorKey: 'vehicle_type',
      cell: ({ getValue }) => <span className="font-medium">{getValue<string>() || '-'}</span>
    },
    {
      header: 'Plaque',
      accessorKey: 'vehicle_plate',
      cell: ({ getValue }) => <span className="font-mono">{getValue<string>() || '-'}</span>
    },
    {
      header: 'Couleur',
      accessorKey: 'vehicle_color',
      cell: ({ getValue }) => <span>{getValue<string>() || '-'}</span>
    },
    {
      header: 'Note',
      accessorKey: 'rating',
      cell: ({ getValue }) => (
        <div className="flex items-center gap-1">
          <Star className="w-4 h-4 fill-yellow-500 text-yellow-500" />
          <span className="font-medium">{getValue<number>()?.toFixed(1) || '-'}</span>
        </div>
      )
    },
    {
      header: 'Livraisons',
      accessorKey: 'total_deliveries',
      cell: ({ getValue }) => <span className="font-medium">{getValue<number>()}</span>
    },
    {
      header: 'Disponible',
      accessorKey: 'is_available',
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
            onClick={() => toggleAvailability(row.original)}
            title={row.original.is_available ? 'Rendre indisponible' : 'Rendre disponible'}
          >
            {row.original.is_available ? <ToggleRight className="w-4 h-4 text-green-500" /> : <ToggleLeft className="w-4 h-4 text-surface-muted" />}
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
            onClick={() => deleteDriver(row.original)}
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
        title="سائقي التوصيل"
        subtitle="إدارة سائقي التوصيل"
        icon={Truck}
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

      <Modal isOpen={isModalOpen} onOpenChange={setIsModalOpen} title={editingDriver ? 'تعديل السائق' : 'سائق جديد'}>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="user_id">معرف المستخدم *</Label>
            <Input
              id="user_id"
              value={form.user_id}
              onChange={(e) => setForm({ ...form, user_id: e.target.value })}
              placeholder="معرف المستخدم"
              required
            />
          </div>

          <div>
            <Label htmlFor="vehicle_type">نوع المركبة *</Label>
            <Input
              id="vehicle_type"
              value={form.vehicle_type}
              onChange={(e) => setForm({ ...form, vehicle_type: e.target.value })}
              placeholder="مثال: Van, Motorcycle, Car"
              required
            />
          </div>

          <div>
            <Label htmlFor="vehicle_plate">رقم اللوحة *</Label>
            <Input
              id="vehicle_plate"
              value={form.vehicle_plate}
              onChange={(e) => setForm({ ...form, vehicle_plate: e.target.value })}
              placeholder="مثال: MRU-1234"
              required
            />
          </div>

          <div>
            <Label htmlFor="vehicle_color">اللون</Label>
            <Input
              id="vehicle_color"
              value={form.vehicle_color}
              onChange={(e) => setForm({ ...form, vehicle_color: e.target.value })}
              placeholder="مثال: White, Black, Red"
            />
          </div>

          <div>
            <Label htmlFor="license_number">رقم الرخصة *</Label>
            <Input
              id="license_number"
              value={form.license_number}
              onChange={(e) => setForm({ ...form, license_number: e.target.value })}
              placeholder="مثال: LIC-001"
              required
            />
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="is_available"
              checked={form.is_available}
              onChange={(e) => setForm({ ...form, is_available: e.target.checked })}
              className="w-4 h-4"
            />
            <Label htmlFor="is_available">متاح</Label>
          </div>

          <div className="flex justify-end gap-2 pt-4">
            <Button type="button" variant="outline" onClick={() => setIsModalOpen(false)} disabled={registerMutation.isPending || updateMutation.isPending}>
              إلغاء
            </Button>
            <Button type="submit" disabled={registerMutation.isPending || updateMutation.isPending}>
              {(registerMutation.isPending || updateMutation.isPending) ? 'جاري التحميل...' : (editingDriver ? 'تعديل' : 'إنشاء')}
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  )
}
