import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as api from '@/api/messages'
import toast from 'react-hot-toast'

export const messageKeys = {
  all: ['messages'] as const,
  conversations: () => [...messageKeys.all, 'conversations'] as const,
  list: (conversationId: string) => [...messageKeys.all, 'list', conversationId] as const,
}

export function useConversations() {
  return useQuery({
    queryKey: messageKeys.conversations(),
    queryFn: api.fetchConversations,
  })
}

export function useMessages(conversationId: string) {
  return useQuery({
    queryKey: messageKeys.list(conversationId),
    queryFn: () => api.fetchMessages(conversationId),
    enabled: !!conversationId,
    refetchInterval: 30000, // Fallback polling (WebSocket is primary)
  })
}

export function useSendMessage() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ conversationId, payload }: { 
      conversationId: string; 
      payload: { type: 'text' | 'image' | 'file'; content?: string } 
    }) => api.sendMessage(conversationId, payload),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: messageKeys.list(vars.conversationId) })
      qc.invalidateQueries({ queryKey: messageKeys.conversations() })
    },
    onError: (err: Error) => toast.error(err.message),
  })
}

export function useMarkAsRead() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (conversationId: string) => api.markAsRead(conversationId),
    onSuccess: (_, conversationId) => {
      qc.invalidateQueries({ queryKey: messageKeys.conversations() })
    },
  })
}

export function useCreateConversation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: { type: 'direct' | 'group' | 'support', user_ids: string[], title?: string }) => 
      api.createConversation(payload),
    onSuccess: (newConv) => {
      qc.invalidateQueries({ queryKey: messageKeys.conversations() })
      toast.success('تم إنشاء المحادثة بنجاح')
    },
    onError: (err: Error) => toast.error(err.message),
  })
}
