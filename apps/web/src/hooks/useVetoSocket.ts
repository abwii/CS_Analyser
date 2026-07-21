import { useEffect, useRef, useState } from 'react'
import type { VetoActionType, VetoState } from '../types'

interface ActionPayload {
  team: string
  action: VetoActionType
  map: string
}

interface ErrorMessage {
  error: string
}

function isErrorMessage(data: VetoState | ErrorMessage): data is ErrorMessage {
  return 'error' in data
}

export function useVetoSocket(wsUrl: string) {
  const [state, setState] = useState<VetoState | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [connected, setConnected] = useState(false)
  const socketRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    const socket = new WebSocket(wsUrl)
    socketRef.current = socket

    socket.onopen = () => setConnected(true)
    socket.onclose = () => setConnected(false)
    socket.onerror = () => setError('websocket connection error')

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data as string) as VetoState | ErrorMessage
      if (isErrorMessage(data)) {
        setError(data.error)
        return
      }
      setError(null)
      setState(data)
    }

    return () => {
      socket.close()
      socketRef.current = null
    }
  }, [wsUrl])

  function sendAction(action: ActionPayload) {
    const socket = socketRef.current
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify(action))
    }
  }

  return { state, error, connected, sendAction }
}
