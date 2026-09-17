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
  const { data } = await client.get<APIResponse<Transaction>>(`/v1/api/admin/transactions/${id}`)
  return data.data
}

// Le lien public direct du reçu n'est plus exposé par l'API (audit de sécurité — le
// bucket R2 est public, donc l'ancien receipt_url fonctionnait sans authentification
// pour quiconque le connaissait). Cette fonction récupère une URL présignée temporaire
// (valable quelques minutes) générée à la demande côté serveur.
export async function fetchReceiptURL(id: string): Promise<{ url: string; expires_in: number }> {
  const { data } = await client.get<APIResponse<{ url: string; expires_in: number }>>(
    `/v1/api/admin/transactions/${id}/receipt-url`
  )
  return data.data
}

export async function validateTransaction(payload: {
  id: string
  approve: boolean
  notes: string
  attachmentUrl?: string
}): Promise<void> {
  await client.put(`/v1/api/admin/transactions/${payload.id}/validate`, {
    approve: payload.approve,
    notes:   payload.notes,
    attachment_url: payload.attachmentUrl,
  })
}

// uploadReviewAttachment (Customer #36): an optional image the admin
// attaches while approving/rejecting a transaction. Not tied to a specific
// transaction id -- returns a URL the client then submits alongside
// approve/notes in the validateTransaction call above.
export async function uploadReviewAttachment(file: File): Promise<{ url: string }> {
  const formData = new FormData()
  formData.append('file', file)
  const { data } = await client.post<APIResponse<{ url: string }>>(
    '/v1/api/admin/transactions/upload',
    formData,
    { headers: { 'Content-Type': 'multipart/form-data' } }
  )
  return data.data
}

// addBalance (Customer #35): admin credits a user's wallet directly from a
// transaction's detail page. Only the anchor transaction id + amount are
// sent -- the backend derives the target user_id from that transaction's
// own user_id server-side, never from client input.
export async function addBalance(payload: {
  id: string
  amount: string
  notes?: string
}): Promise<{ transaction_id: string; amount: string }> {
  const { data } = await client.post<APIResponse<{ transaction_id: string; amount: string }>>(
    `/v1/api/admin/transactions/${payload.id}/add-balance`,
    { amount: payload.amount, notes: payload.notes }
  )
  return data.data
}

export async function exportTransactions(filters: { status?: string, start_date?: string, end_date?: string }): Promise<void> {
  const { data } = await client.get('/v1/api/admin/reports/transactions/export', {
    params: filters,
    responseType: 'blob'
  })
  
  const url = window.URL.createObjectURL(new Blob([data]))
  const link = document.createElement('a')
  link.href = url
  link.setAttribute('download', `transactions_${new Date().toISOString().split('T')[0]}.csv`)
  document.body.appendChild(link)
  link.click()
  link.remove()
}
