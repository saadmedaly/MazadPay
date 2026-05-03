import { useEffect, useRef, useCallback } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useAuthStore } from '@/stores/authStore'
import { messageKeys } from './useMessages'
import toast from 'react-hot-toast'

export function useChatWebSocket(conversationId?: string | null) {
  const socketRef = useRef<WebSocket | null>(null)
  const { token } = useAuthStore()
  const queryClient = useQueryClient()
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const lastJoinedConvRef = useRef<string | null>(null)

  const connect = useCallback(() => {
    if (!token || socketRef.current?.readyState === WebSocket.OPEN) return

    const apiUrl = import.meta.env.VITE_API_URL ?? 'http://localhost:8082'
    const wsBase = apiUrl.replace(/^http/, 'ws')
    const wsUrl = `${wsBase}/v1/api/chat/ws?token=${token}`

    console.log('Connecting to Chat WebSocket...')
    const socket = new WebSocket(wsUrl)

    socket.onopen = () => {
      console.log('Chat WebSocket connected')
      // Reset last joined ref so we re-join the current conv
      lastJoinedConvRef.current = null

      // If we have an active conversation, join it
      if (conversationId) {
        socket.send(JSON.stringify({
          event: 'conversation:join',
          conversation_id: conversationId,
          timestamp: new Date().toISOString()
        }))
        lastJoinedConvRef.current = conversationId
      }
    }

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)

        switch (data.event) {
          case 'message:new':
            // Invalidate the message list if it's for the current conversation
            if (data.conversation_id === conversationId || (data.data && data.data.conversation_id === conversationId)) {
              queryClient.invalidateQueries({ queryKey: messageKeys.list(conversationId || '') })
            }
            // Always invalidate conversation list to show latest preview/unread count
            queryClient.invalidateQueries({ queryKey: messageKeys.conversations() })
            break

          case 'message:read':
            queryClient.invalidateQueries({ queryKey: messageKeys.conversations() })
            break

          case 'error':
            console.error('WebSocket Error:', data.data)
            break
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message', err)
      }
    }

    let isUnmounting = false

    socket.onclose = (event) => {
      socketRef.current = null
      // Ne pas reconnecter si la page se décharge ou si c'est une fermeture propre
      if (isUnmounting || event.wasClean) {
        console.log('Chat WebSocket closed cleanly')
        return
      }
      console.log('Chat WebSocket disconnected, retrying in 3s...')
      reconnectTimeoutRef.current = setTimeout(connect, 3000)
    }

    socket.onerror = (err) => {
      // Ne pas logger d'erreur si c'est juste une déconnexion pendant le unload
      if (!isUnmounting) {
        console.warn('WebSocket error (will retry):', err.type || 'connection error')
      }
    }

    // Marquer comme démontage quand la page se décharge
    const handleBeforeUnload = () => { isUnmounting = true }
    window.addEventListener('beforeunload', handleBeforeUnload)

    socketRef.current = socket

    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload)
    }
  }, [token, queryClient, conversationId]) // still need conversationId for the message listener closures

  useEffect(() => {
    connect()
    return () => {
      if (reconnectTimeoutRef.current) clearTimeout(reconnectTimeoutRef.current)
      if (socketRef.current) {
        socketRef.current.close()
      }
    }
  }, [token]) // Only reconnect if token changes

  // Handle joining/leaving conversations
  useEffect(() => {
    const socket = socketRef.current
    if (socket?.readyState === WebSocket.OPEN && conversationId !== lastJoinedConvRef.current) {
      // Leave old
      if (lastJoinedConvRef.current) {
        socket.send(JSON.stringify({
          event: 'conversation:leave',
          conversation_id: lastJoinedConvRef.current,
          timestamp: new Date().toISOString()
        }))
      }
      // Join new
      if (conversationId) {
        socket.send(JSON.stringify({
          event: 'conversation:join',
          conversation_id: conversationId,
          timestamp: new Date().toISOString()
        }))
      }
      lastJoinedConvRef.current = conversationId || null
    }
  }, [conversationId])

  const sendEvent = useCallback((event: string, data: any) => {
    if (socketRef.current?.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify({
        event,
        conversation_id: conversationId,
        data,
        timestamp: new Date().toISOString()
      }))
    }
  }, [conversationId])

  return { sendEvent }
}
