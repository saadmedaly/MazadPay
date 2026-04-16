import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import type { AdminUser } from '@/types/api'

interface AuthStore {
  token: string | null
  user: AdminUser | null
  setAuth: (token: string, user: AdminUser) => void
  logout: () => void
  isAuthenticated: () => boolean
}

export const useAuthStore = create<AuthStore>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,
      setAuth: (token, user) => set({ token, user }),
      logout: () => set({ token: null, user: null }),
      isAuthenticated: () => {
        const { token, user } = get()
        return !!token && user?.role === 'admin'
      },
    }),
    {
      name: 'mazadpay-admin-auth',
      storage: createJSONStorage(() => localStorage),
    }
  )
)
