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
    <div className="p-6 space-y-6 flex flex-col items-center justify-center min-h-[60vh] text-center" dir="rtl">
      <div className="w-24 h-24 rounded-full bg-mazad-primary/10 flex items-center justify-center mb-6">
        <Truck className="w-12 h-12 text-mazad-primary opacity-50" />
      </div>
      <PageHeader
        title="خدمة التوصيل"
        subtitle="إدارة سائقي التوصيل وعمليات النقل"
        icon={Truck}
      />
      <div className="mt-8 p-10 bg-surface-card border border-surface-border rounded-3xl shadow-xl max-w-lg">
        <h2 className="text-2xl font-bold text-white mb-4">الخدمة غير متوفرة حالياً</h2>
        <p className="text-surface-muted leading-relaxed mb-8">
          نحن نعمل على إطلاق خدمة التوصيل MazadDelivery في الإصدارات القادمة. 
          ترقبوا التحديثات الجديدة قريباً!
        </p>
        <div className="py-2 px-6 bg-mazad-primary/20 text-mazad-primary rounded-full inline-block font-bold">
          قريباً في V2.5
        </div>
      </div>
    </div>
  )
}
