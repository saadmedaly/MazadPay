import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatPrice(price: number | string): string {
  const numPrice = typeof price === 'string' ? parseFloat(price) : price
  return new Intl.NumberFormat('ar-MA', {
    style: 'currency',
    currency: 'MRO',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(numPrice)
}

export function formatDate(dateString: string): string {
  const date = new Date(dateString)
  return new Intl.DateTimeFormat('ar-MA', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  }).format(date)
}

export function shortID(id: string): string {
  return id.slice(0, 8)
}
