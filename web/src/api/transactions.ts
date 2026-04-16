import client from './client'
import type { PaginatedResponse, Transaction, APIResponse } from '@/types/api'

export interface TxnFilters {
  status?: string
  type?: string
  page?: number
  per_page?: number
}

export async function fetchTransactions(filters: TxnFilters) {
  const { data } = await client.get<PaginatedResponse<Transaction>>(
    '/v1/api/admin/transactions', { params: filters }
  )
  return { data: data.data, total: data.meta.total }
}

export async function fetchTransaction(id: string): Promise<Transaction> {
  const { data } = await client.get<APIResponse<Transaction>>(`/v1/api/transactions/${id}`)
  return data.data
}

export async function validateTransaction(payload: {
  id: string
  approve: boolean
  notes: string
}): Promise<void> {
  await client.put(`/v1/api/admin/transactions/${payload.id}/validate`, {
    approve: payload.approve,
    notes:   payload.notes,
  })
}
