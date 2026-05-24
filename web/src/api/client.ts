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
  console.log(
    `%c[API Request] ${config.method?.toUpperCase()} ${config.url}`,
    'color: #60a5fa; font-weight: bold',
    config.data ? '\nPayload:' : '',
    config.data ?? ''
  )
  return config
})

client.interceptors.response.use(
  (res) => {
    console.log(
      `%c[API Response] ${res.status} ${res.config.method?.toUpperCase()} ${res.config.url}`,
      'color: #34d399; font-weight: bold',
      '\nData:', res.data
    )
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

    console.error(
      `%c[API Error] ${status} ${err.config?.method?.toUpperCase()} ${err.config?.url}`,
      'color: #f87171; font-weight: bold',
      '\nMessage:', message,
      '\nServer response:', serverData,
      '\nFull error:', err
    )

    if (err.response?.status === 401) {
      useAuthStore.getState().logout()
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(new Error(message))
  }
)

export default client
