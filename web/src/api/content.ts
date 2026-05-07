import client from './client'
import type { APIResponse } from '@/types/api'

// FAQ Types
export interface FAQItem {
  id: number
  question_ar: string
  question_fr?: string
  answer_ar: string
  answer_fr?: string
  display_order: number
}

// Tutorial Types
export interface Tutorial {
  id: number
  title_ar: string
  title_fr?: string
  video_url: string
  thumbnail_url?: string
  category?: string
  display_order: number
}

// Banner Types
export interface Banner {
  id: number
  title_ar: string
  title_fr: string
  title_en: string
  image_url: string
  target_url: string
  is_active: boolean
  starts_at?: string
  ends_at?: string
  display_order: number
}

// FAQ API
export async function fetchFAQ() {
  const res = await client.get<APIResponse<FAQItem[]>>('/v1/api/faq')
  return res.data.data
}

export async function createFAQ(payload: Omit<FAQItem, 'id'>) {
  const res = await client.post<APIResponse<FAQItem>>('/v1/api/admin/faq', payload)
  return res.data.data
}

export async function updateFAQ(id: number, payload: Omit<FAQItem, 'id'>) {
  const res = await client.put<APIResponse<FAQItem>>(`/v1/api/admin/faq/${id}`, payload)
  return res.data.data
}

export async function deleteFAQ(id: number) {
  const res = await client.delete<APIResponse<void>>(`/v1/api/admin/faq/${id}`)
  return res.data.data
}

// Upload FAQ image to R2
export async function uploadFAQImage(file: File) {
  const formData = new FormData()
  formData.append('file', file)

  const res = await client.post<APIResponse<{ url: string }>>(
    '/v1/api/admin/faq/upload',
    formData,
    {
      timeout: 30_000,
      headers: {
        'Content-Type': undefined,
      },
    }
  )
  return res.data.data
}

// Tutorials API
export async function fetchTutorials() {
  const res = await client.get<APIResponse<Tutorial[]>>('/v1/api/tutorials')
  return res.data.data
}

export async function createTutorial(payload: Omit<Tutorial, 'id'>) {
  const res = await client.post<APIResponse<Tutorial>>('/v1/api/admin/tutorials', payload)
  return res.data.data
}

export async function updateTutorial(id: number, payload: Omit<Tutorial, 'id'>) {
  const res = await client.put<APIResponse<Tutorial>>(`/v1/api/admin/tutorials/${id}`, payload)
  return res.data.data
}

export async function deleteTutorial(id: number) {
  const res = await client.delete<APIResponse<void>>(`/v1/api/admin/tutorials/${id}`)
  return res.data.data
}

// Upload tutorial video to R2
export async function uploadTutorialVideo(file: File) {
  const formData = new FormData()
  formData.append('file', file)

  const res = await client.post<APIResponse<{ url: string }>>(
    '/v1/api/admin/tutorials/upload-video',
    formData,
    {
      timeout: 120_000, // 2 min for videos
      headers: {
        'Content-Type': undefined,
      },
    }
  )
  return res.data.data
}

// Upload tutorial thumbnail to R2
export async function uploadTutorialThumbnail(file: File) {
  const formData = new FormData()
  formData.append('file', file)

  const res = await client.post<APIResponse<{ url: string }>>(
    '/v1/api/admin/tutorials/upload-thumbnail',
    formData,
    {
      timeout: 30_000,
      headers: {
        'Content-Type': undefined,
      },
    }
  )
  return res.data.data
}

// Banners API
export async function fetchBanners() {
  const res = await client.get<APIResponse<Banner[]>>('/v1/api/banners')
  return res.data.data
}

export async function fetchAdminBanners() {
  const res = await client.get<APIResponse<Banner[]>>('/v1/api/admin/banners')
  return res.data.data
}

export async function createBanner(payload: Omit<Banner, 'id'>) {
  const res = await client.post<APIResponse<Banner>>('/v1/api/admin/banners', payload)
  return res.data.data
}

export async function updateBanner(id: number, payload: Partial<Banner>) {
  const res = await client.put<APIResponse<Banner>>(`/v1/api/admin/banners/${id}`, payload)
  return res.data.data
}

export async function deleteBanner(id: number) {
  const res = await client.delete<APIResponse<void>>(`/v1/api/admin/banners/${id}`)
  return res.data.data
}

export async function toggleBanner(id: number, isActive: boolean) {
  const res = await client.put<APIResponse<void>>(`/v1/api/admin/banners/${id}/toggle`, { is_active: isActive })
  return res.data.data
}

// Upload banner image to R2
export async function uploadBannerImage(file: File) {
  const formData = new FormData()
  formData.append('file', file)

  const res = await client.post<APIResponse<{ url: string }>>(
    '/v1/api/admin/banners/upload',
    formData,
    {
      timeout: 30_000,
      headers: {
        'Content-Type': undefined,
      },
    }
  )
  return res.data.data
}
