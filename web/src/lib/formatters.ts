import { format, formatDistanceToNow } from 'date-fns'
import { ar } from 'date-fns/locale'

export const DEFAULT_CURRENCY_CODE = 'MRU'

const MINOR_UNITS: Record<string, number> = {
  MRU: 0,
  TND: 3,
  MAD: 2,
  XOF: 0,
  EUR: 2,
  USD: 2,
}

// Currency-aware price formatter (migration 000046, Phase 2). Displays the
// amount with its own ISO-4217 currency code -- NO exchange-rate conversion,
// NO assumed/global currency. currencyCode should come from the record's own
// field (or its Effective*CurrencyCode fallback server-side); defaults to
// DEFAULT_CURRENCY_CODE only when the API object legitimately has none.
export const formatPrice = (amount: number | string, currencyCode?: string | null): string => {
  const n = typeof amount === 'string' ? parseFloat(amount) : amount
  if (isNaN(n)) return '—'
  const code = currencyCode && currencyCode.trim() !== '' ? currencyCode : DEFAULT_CURRENCY_CODE
  const digits = MINOR_UNITS[code] ?? 2
  return '‎' + new Intl.NumberFormat('en-US', {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  }).format(n) + ' ' + code
}

export const formatDate = (date: string | Date): string => {
  if (!date) return '—'
  const d = new Date(date)
  if (isNaN(d.getTime())) return '—'
  return format(d, 'dd MMM yyyy, HH:mm', { locale: ar })
}

export const formatDateShort = (date: string | Date): string => {
  if (!date) return '—'
  const d = new Date(date)
  if (isNaN(d.getTime())) return '—'
  return format(d, 'yyyy/MM/dd', { locale: ar })
}

export const formatRelative = (date: string | Date): string => {
  if (!date) return '—'
  const d = new Date(date)
  if (isNaN(d.getTime())) return '—'
  return formatDistanceToNow(d, { addSuffix: true, locale: ar })
}

export const maskPhone = (phone: string): string => {
  if (!phone || phone.length < 4) return '####'
  return '####' + phone.slice(-4)
}

export const shortID = (id: string): string =>
  id?.slice(0, 8).toUpperCase() ?? '—'

export const formatPercent = (value: number): string =>
  `${value >= 0 ? '+' : ''}${value.toFixed(1)}%`
