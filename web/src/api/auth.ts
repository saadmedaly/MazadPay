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
  // Authorization contract (matches the backend's own model -- see
  // AuthService.GenerateJWT(userID, role, isSuperAdmin) and AdminUser's
  // existing `role`/`is_super_admin` fields, types/api.ts): role === 'admin'
  // was the ONLY path previously accepted here, silently rejecting a
  // legitimate role === 'super_admin' account (a strictly higher privilege
  // level, not a different one) even after a correct login. Checking
  // is_super_admin directly, in addition to role === 'admin', matches how
  // the backend itself distinguishes privilege rather than re-deriving it
  // from an incomplete role-string enumeration on the frontend.
  const { role, is_super_admin } = data.data.user
  if (role !== 'admin' && !is_super_admin) {
    throw new Error('عذراً، هذا الدخول مخصص للمسؤولين فقط')
  }
  return data.data
}
