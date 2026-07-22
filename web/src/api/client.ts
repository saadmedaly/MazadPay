import axios from 'axios'
import { useAuthStore } from '@/stores/authStore'

const client = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? 'http://localhost:8082',
  timeout: 10_000,
  headers: { 'Content-Type': 'application/json' },
})

client.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  // Ne jamais logger le payload complet des requêtes (peut contenir des données
  // sensibles : mots de passe, montants, reçus...) — durcissement sécurité.
  if (import.meta.env.DEV) {
    console.log(`[API] ${config.method?.toUpperCase()} ${config.url}`)
  }
  return config
})

client.interceptors.response.use(
  (res) => {
    if (import.meta.env.DEV) {
      console.log(`[API] ${res.status} ${res.config.method?.toUpperCase()} ${res.config.url}`)
    }
    return res
  },
  (err) => {
    const status = err.response?.status ?? 'NETWORK_ERROR'
    const serverData = err.response?.data
    let message = err.message
    if (typeof serverData?.error === 'string') {
      message = serverData.error
    } else if (serverData?.error?.message) {
      message = serverData.error.message
    } else if (serverData?.message) {
      message = serverData.message
    }

    // Ne jamais logger le corps de la réponse serveur ni l'objet erreur complet — peut
    // contenir des jetons, des détails internes ou des données utilisateur sensibles.
    if (import.meta.env.DEV) {
      console.error(`[API Error] ${status} ${err.config?.method?.toUpperCase()} ${err.config?.url}: ${message}`)
    }

    // Session invalide ou accès refusé : déconnexion automatique et retour au login.
    if (err.response?.status === 401 || err.response?.status === 403) {
      useAuthStore.getState().logout()
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(new Error(message))
  }
)

export default client
