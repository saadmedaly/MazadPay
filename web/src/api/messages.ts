import client from './client'
import type { APIResponse } from '@/types/api'

export interface Conversation {
  id: string
  type: 'direct' | 'group' | 'support'
  title?: string
  last_message_at?: string
  last_message_preview?: string
  last_message_sender_id?: string
  is_active: boolean
  unread_count: number
  participants?: Participant[]
}

export interface Participant {
  id: string
  conversation_id: string
  user_id: string
  role: string
  joined_at: string
  unread_count: number
  user?: {
    id: string
    full_name: string
    profile_pic_url: string
    is_active: boolean
  }
}

export interface Message {
  id: string
  conversation_id: string
  sender_id?: string
  type: 'text' | 'audio' | 'video' | 'image' | 'file' | 'system'
  content?: string
  file_name?: string
  file_url?: string
  file_size?: number
  file_duration?: number
  mime_type?: string
  thumbnail_url?: string
  created_at: string
  sender?: {
    id: string
    full_name: string
    profile_pic_url: string
  }
  status?: MessageStatus[]
}

export interface MessageStatus {
  id: string
  message_id: string
  user_id: string
  status: 'sent' | 'delivered' | 'read'
  updated_at: string
}

export async function fetchConversations() {
  const res = await client.get<APIResponse<Conversation[]>>('/v1/api/conversations/')
  return res.data.data
}

export async function createConversation(payload: { 
  type: 'direct' | 'group' | 'support', 
  user_ids: string[], 
  title?: string 
}) {
  const res = await client.post<APIResponse<Conversation>>('/v1/api/conversations/', payload)
  return res.data.data
}

export async function fetchMessages(conversationId: string) {
  const res = await client.get<APIResponse<Message[]>>(`/v1/api/conversations/${conversationId}/messages`)
  return res.data.data
}

export async function sendMessage(conversationId: string, payload: {
  type: 'text' | 'image' | 'file'
  content?: string
  file_url?: string
}) {
  const res = await client.post<APIResponse<Message>>(`/v1/api/conversations/${conversationId}/messages`, payload)
  return res.data.data
}

export async function markAsRead(conversationId: string) {
  const res = await client.post<APIResponse<void>>(`/v1/api/conversations/${conversationId}/read`)
  return res.data.data
}
