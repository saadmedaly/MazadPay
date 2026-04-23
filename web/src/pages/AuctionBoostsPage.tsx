import React, { useState } from 'react'
import {
  Plus, Search, Pencil, Zap,
  XCircle, CheckCircle, Clock,
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
  useAuctionBoosts,
  useCreateBoost,
  useCancelBoost,
  type AuctionBoost
} from '@/hooks/useAuctionBoosts'

interface BoostForm {
  auction_id: string
  boost_type: string
  start_at: string
  end_at: string
  amount: string
}

export function AuctionBoostsPage() {
  const [search, setSearch] = useState('')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingBoost, setEditingBoost] = useState<AuctionBoost | null>(null)
  const [form, setForm] = useState<BoostForm>({
    auction_id: '',
    boost_type: 'featured',
    start_at: '',
    end_at: '',
    amount: ''
  })

  const { data: boosts = [], isLoading } = useAuctionBoosts()
  const createMutation = useCreateBoost()
  const cancelMutation = useCancelBoost()

  const filtered = boosts.filter(b =>
    b.id.includes(search) ||
    b.auction_id.toLowerCase().includes(search.toLowerCase()) ||
    b.boost_type.toLowerCase().includes(search.toLowerCase())
  )

  const openAdd = () => {
    setEditingBoost(null)
    setForm({ auction_id: '', boost_type: 'featured', start_at: '', end_at: '', amount: '' })
    setIsModalOpen(true)
  }

  const openEdit = (boost: AuctionBoost) => {
    setEditingBoost(boost)
    setForm({
      auction_id: boost.auction_id,
      boost_type: boost.boost_type,
      start_at: boost.start_at.split('T')[0],
      end_at: boost.end_at.split('T')[0],
      amount: boost.amount?.toString() || ''
    })
    setIsModalOpen(true)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const data = {
      boost_type: form.boost_type,
      start_at: new Date(form.start_at).toISOString(),
      end_at: new Date(form.end_at).toISOString(),
      amount: form.amount ? parseFloat(form.amount) : undefined
    }
    createMutation.mutate({ auctionId: form.auction_id, data })
    setIsModalOpen(false)
  }

  const cancelBoost = (boost: AuctionBoost) => {
    if (confirm(`Voulez-vous vraiment annuler ce boost?`)) {
      cancelMutation.mutate({ auctionId: boost.auction_id, boostId: boost.id })
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active':
        return <CheckCircle className="w-4 h-4 text-green-500" />
      case 'completed':
        return <CheckCircle className="w-4 h-4 text-blue-500" />
      case 'cancelled':
        return <XCircle className="w-4 h-4 text-red-500" />
      default:
        return <Clock className="w-4 h-4 text-surface-muted" />
    }
  }

  const columns: ColumnDef<AuctionBoost>[] = [
    {
      header: 'ID',
      accessorKey: 'id',
      cell: ({ getValue }) => <span className="font-mono text-xs text-surface-muted">{getValue<string>().slice(0, 8)}...</span>
    },
    {
      header: 'Enchère ID',
      accessorKey: 'auction_id',
      cell: ({ getValue }) => <span className="font-mono text-sm">{getValue<string>().slice(0, 8)}...</span>
    },
    {
      header: 'Type de Boost',
      accessorKey: 'boost_type',
      cell: ({ getValue }) => {
        const type = getValue<string>()
        const typeLabels: Record<string, string> = {
          featured: 'À la une',
          urgent: 'Urgent',
          top: 'Top'
        }
        return <span className="font-medium capitalize">{typeLabels[type] || type}</span>
      }
    },
    {
      header: 'Date Début',
      accessorKey: 'start_at',
      cell: ({ getValue }) => <span className="text-sm">{getValue<string>().split('T')[0]}</span>
    },
    {
      header: 'Date Fin',
      accessorKey: 'end_at',
      cell: ({ getValue }) => <span className="text-sm">{getValue<string>().split('T')[0]}</span>
    },
    {
      header: 'Montant',
      accessorKey: 'amount',
      cell: ({ getValue }) => <span className="font-medium">{getValue<number>() ? `${getValue<number>()} MRU` : '-'}</span>
    },
    {
      header: 'Statut',
      accessorKey: 'status',
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          {getStatusIcon(row.original.status)}
          <StatusBadge status={row.original.status} />
        </div>
      )
    },
    {
      header: 'Actions',
      cell: ({ row }) => (
        <div className="flex gap-2">
          {row.original.status === 'active' && (
            <Button
              size="sm"
              variant="ghost"
              onClick={() => cancelBoost(row.original)}
              title="Annuler"
            >
              <XCircle className="w-4 h-4 text-red-500" />
            </Button>
          )}
          <Button
            size="sm"
            variant="ghost"
            onClick={() => openEdit(row.original)}
          >
            <Pencil className="w-4 h-4" />
          </Button>
        </div>
      )
    }
  ]

  return (
    <div className="p-6 space-y-6">
      <PageHeader
        title="تعزيزات المزادات"
        subtitle="إدارة تعزيزات المزادات"
        icon={Zap}
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

      <Modal isOpen={isModalOpen} onOpenChange={setIsModalOpen} title={editingBoost ? 'تعديل التعزيز' : 'تعزيز جديد'}>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="auction_id">معرف المزاد *</Label>
            <Input
              id="auction_id"
              value={form.auction_id}
              onChange={(e) => setForm({ ...form, auction_id: e.target.value })}
              placeholder="معرف المزاد"
              required
            />
          </div>

          <div>
            <Label htmlFor="boost_type">نوع التعزيز *</Label>
            <select
              id="boost_type"
              value={form.boost_type}
              onChange={(e) => setForm({ ...form, boost_type: e.target.value })}
              className="w-full px-3 py-2 bg-surface border border-surface-border rounded-md"
              required
            >
              <option value="featured">مميز</option>
              <option value="urgent">عاجل</option>
              <option value="top">أعلى</option>
            </select>
          </div>

          <div>
            <Label htmlFor="start_at">تاريخ البدء *</Label>
            <Input
              id="start_at"
              type="date"
              value={form.start_at}
              onChange={(e) => setForm({ ...form, start_at: e.target.value })}
              required
            />
          </div>

          <div>
            <Label htmlFor="end_at">تاريخ الانتهاء *</Label>
            <Input
              id="end_at"
              type="date"
              value={form.end_at}
              onChange={(e) => setForm({ ...form, end_at: e.target.value })}
              required
            />
          </div>

          <div>
            <Label htmlFor="amount">المبلغ (MRU)</Label>
            <Input
              id="amount"
              type="number"
              value={form.amount}
              onChange={(e) => setForm({ ...form, amount: e.target.value })}
              placeholder="0.00"
            />
          </div>

          <div className="flex justify-end gap-2 pt-4">
            <Button type="button" variant="outline" onClick={() => setIsModalOpen(false)} disabled={createMutation.isPending}>
              إلغاء
            </Button>
            <Button type="submit" disabled={createMutation.isPending}>
              {createMutation.isPending ? 'جاري التحميل...' : (editingBoost ? 'تعديل' : 'إنشاء')}
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  )
}
