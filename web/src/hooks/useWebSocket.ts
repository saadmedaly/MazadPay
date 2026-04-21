import { useEffect, useRef, useState, useCallback } from 'react'
import { useAuthStore } from '@/stores/authStore'

export interface WebSocketMessage {
  type: string
  payload: any
}

export function useWebSocket(url: string) {
  const [isConnected, setIsConnected] = useState(false)
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null)
  const ws = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const { token } = useAuthStore()

  const connect = useCallback(() => {
    if (!token) return

    try {
      // Ajouter le token comme paramètre de query
      const wsUrl = `${url}?token=${token}`
      ws.current = new WebSocket(wsUrl)

      ws.current.onopen = () => {
        setIsConnected(true)
        console.log('WebSocket connected')
      }

      ws.current.onclose = () => {
        setIsConnected(false)
        console.log('WebSocket disconnected')
        // Reconnexion automatique après 5 secondes
        reconnectTimeoutRef.current = setTimeout(() => {
          connect()
        }, 5000)
      }

      ws.current.onerror = (error) => {
        console.error('WebSocket error:', error)
      }

      ws.current.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data)
          setLastMessage(message)
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err)
        }
      }
    } catch (err) {
      console.error('Failed to connect WebSocket:', err)
    }
  }, [url, token])

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }
    if (ws.current) {
      ws.current.close()
      ws.current = null
    }
  }, [])

  const sendMessage = useCallback((message: WebSocketMessage) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(message))
    } else {
      console.warn('WebSocket is not connected')
    }
  }, [])

  useEffect(() => {
    connect()
    return () => {
      disconnect()
    }
  }, [connect, disconnect])

  return {
    isConnected,
    lastMessage,
    sendMessage,
    connect,
    disconnect
  }
}

// Hook spécifique pour les notifications admin
export function useAdminNotifications() {
  const { isConnected, lastMessage } = useWebSocket(
    `${import.meta.env.VITE_WS_URL || 'ws://localhost:8080'}/ws/admin`
  )

  return {
    isConnected,
    lastMessage,
    // Helpers pour les types de messages spécifiques
    isNewRequest: lastMessage?.type === 'new_request',
    isRequestUpdated: lastMessage?.type === 'request_updated',
    isRequestReviewed: lastMessage?.type === 'request_reviewed',
    payload: lastMessage?.payload
  }
}
