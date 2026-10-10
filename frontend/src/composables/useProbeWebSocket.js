import { ref, onUnmounted } from 'vue'
import { useAuthStore } from '../stores/auth'

/**
 * 探测 WebSocket 连接管理：连接、3s 断线重连、发送与卸载清理
 * @param {(data: object) => void} onMessage 收到合法帧时的回调（业务处理留在调用方）
 */
export function useProbeWebSocket(onMessage) {
  const authStore = useAuthStore()
  const wsConnected = ref(false)
  let ws = null
  let reconnectTimer = null
  let unmounted = false

  function connect() {
    if (unmounted) return
    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const token = authStore.token
    const url = `${protocol}//${location.host}/api/ws/probe?token=${encodeURIComponent(token)}`

    try {
      ws = new WebSocket(url)
    } catch {
      scheduleReconnect()
      return
    }

    ws.onopen = () => {
      wsConnected.value = true
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (onMessage) onMessage(data)
      } catch {
        // 非法帧静默丢弃
      }
    }

    ws.onclose = () => {
      wsConnected.value = false
      scheduleReconnect()
    }

    ws.onerror = () => {
      if (ws) ws.close()
    }
  }

  function scheduleReconnect() {
    if (unmounted || reconnectTimer) return
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      connect()
    }, 3000)
  }

  function send(payload) {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(payload))
      return true
    }
    return false
  }

  function disconnect() {
    unmounted = true
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (ws) {
      ws.onclose = null
      ws.close()
      ws = null
    }
  }

  onUnmounted(disconnect)

  return { wsConnected, connect, send, disconnect, isUnmounted: () => unmounted }
}
