import client from './client'
import type { APIResponse, AdminUser } from '@/types/api'

export interface LoginPayload {
  phone: string
  pin: string
}

export interface LoginResponse {
  token: string
  user: AdminUser
}

export async function loginAdmin(payload: LoginPayload): Promise<LoginResponse> {
  const { data } = await client.post<APIResponse<LoginResponse>>(
    '/v1/api/auth/login',
    payload
  )
  if (!data.success) throw new Error(data.error?.message)
  if (data.data.user.role !== 'admin') throw new Error('عذراً، هذا الدخول مخصص للمسؤولين فقط')
  return data.data
}
