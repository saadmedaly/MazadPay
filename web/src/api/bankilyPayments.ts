import client from './client'
import type { APIResponse, PaginatedResponse } from '@/types/api'

export type BankilyPaymentStatus = 'INITIATED' | 'SUBMITTED' | 'PENDING_VERIFICATION' | 'SUCCESS' | 'FAILED' | 'NEEDS_REVIEW'

// Bankily B-PAY automatic top-up attempt (admin view). Never contains the
// customer's B-PAY passcode -- the backend does not store it.
export interface BankilyPayment {
  id: string
  user_id: string
  user_full_name: string | null
  user_phone: string | null
  operation_id: string
  amount: string
  currency_code: string
  client_phone: string
  status: BankilyPaymentStatus
  bankily_transaction_id: string | null
  bankily_error_code: string | null
  bankily_error_message: string | null
  bankily_status: string | null
  failure_reason: string | null
  wallet_transaction_id: string | null
  check_attempts: number
  needs_review: boolean
  last_checked_at: string | null
  submitted_at: string | null
  completed_at: string | null
  created_at: string
}

export interface BankilyFilters {
  status?: string
  page?: number
  per_page?: number
}

export async function fetchBankilyPayments(filters: BankilyFilters) {
  const { data } = await client.get<PaginatedResponse<BankilyPayment>>(
    '/v1/api/admin/bankily/payments', { params: filters }
  )
  return { data: data.data, total: data.meta.total }
}

// Runs Bankily checkTransaction for a pending attempt. It can never credit a
// wallet by itself: the backend credits only on a Bankily-confirmed TS.
export async function recheckBankilyPayment(id: string): Promise<BankilyPayment> {
  const { data } = await client.post<APIResponse<BankilyPayment>>(`/v1/api/admin/bankily/payments/${id}/recheck`)
  return data.data
}
