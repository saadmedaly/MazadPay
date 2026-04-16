import { format, formatDistanceToNow } from 'date-fns'
import { ar } from 'date-fns/locale'

export const formatPrice = (amount: number | string): string => {
  const n = typeof amount === 'string' ? parseFloat(amount) : amount
  return new Intl.NumberFormat('ar-MR', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(n) + ' أ.م'
}

export const formatDate = (date: string | Date): string =>
  format(new Date(date), 'dd MMM yyyy, HH:mm', { locale: ar })

export const formatDateShort = (date: string | Date): string =>
  format(new Date(date), 'yyyy/MM/dd', { locale: ar })

export const formatRelative = (date: string | Date): string =>
  formatDistanceToNow(new Date(date), { addSuffix: true, locale: ar })

export const maskPhone = (phone: string): string => {
  if (!phone || phone.length < 4) return '####'
  return '####' + phone.slice(-4)
}

export const shortID = (id: string): string =>
  id?.slice(0, 8).toUpperCase() ?? '—'

export const formatPercent = (value: number): string =>
  `${value >= 0 ? '+' : ''}${value.toFixed(1)}%`
