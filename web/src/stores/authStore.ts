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

// Décode le payload d'un JWT et vérifie sa date d'expiration localement, sans appel
// réseau (durcissement sécurité — évite qu'un token expiré reste "valide" côté client
// jusqu'au prochain appel API qui échouerait avec 401).
function isTokenExpired(token: string): boolean {
  try {
    const payload = token.split('.')[1]
    if (!payload) return false // format inattendu : laisser l'API trancher
    const decoded = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')))
    if (typeof decoded.exp !== 'number') return false
    return Date.now() >= decoded.exp * 1000
  } catch {
    return false // décodage impossible : laisser l'API trancher
  }
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
        if (!token || isTokenExpired(token)) return false
        return user?.role === 'admin' || user?.role === 'super_admin'
      },
    }),
    {
      name: 'mazadpay-admin-auth',
      storage: createJSONStorage(() => localStorage),
    }
  )
)
