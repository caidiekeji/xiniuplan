/**
 * SSE 流式客户端。
 *
 * 为什么不用 axios：axios 的响应是「一次性读完」模型，遇到永不结束的
 * 流式响应（SSE）会一直挂起，直到 timeout 触发。容器日志「跟随」功能
 * 之前就用 axios 请求 follow=true，每 3 秒轮询一次 → 每秒堆积一个
 * 挂起请求，页面越来越卡。
 *
 * 为什么不用 EventSource：EventSource 不支持自定义请求头，无法携带
 * Authorization: Bearer <token>（面板的鉴权方式）。
 *
 * 因此用 fetch + ReadableStream 手动解析 SSE 帧。
 */
import { useAuthStore } from '@/stores/auth'

/** 面板 API 基址（安全入口模式下带前缀） */
function apiBase(): string {
  return `${(window as any).__PANEL_BASE__ || ''}/api/v1`
}

/**
 * 订阅一个 SSE 流，逐行回调。
 *
 * @param path      API 路径（相对 /api/v1）
 * @param onLine    每收到一行数据回调
 * @param onError   出错回调
 * @returns 取消订阅函数
 */
export function subscribeSSE(
  path: string,
  onLine: (line: string) => void,
  onError?: (err: Error) => void
): () => void {
  const controller = new AbortController()
  const authStore = useAuthStore()

  ;(async () => {
    try {
      const resp = await fetch(`${apiBase()}${path}`, {
        method: 'GET',
        headers: {
          Accept: 'text/event-stream',
          ...(authStore.token ? { Authorization: `Bearer ${authStore.token}` } : {})
        },
        signal: controller.signal
      })

      if (!resp.ok || !resp.body) {
        throw new Error(`SSE 连接失败: HTTP ${resp.status}`)
      }

      const reader = resp.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      for (;;) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })

        // SSE 以空行分隔事件；这里按行拆分，data: 前缀的即为数据
        let idx: number
        while ((idx = buffer.indexOf('\n')) >= 0) {
          const line = buffer.slice(0, idx).replace(/\r$/, '')
          buffer = buffer.slice(idx + 1)
          if (!line.startsWith('data:')) continue
          onLine(line.slice(5).replace(/^ /, ''))
        }
      }
    } catch (e: any) {
      // 主动取消不算错误
      if (e?.name === 'AbortError') return
      onError?.(e instanceof Error ? e : new Error(String(e)))
    }
  })()

  return () => controller.abort()
}
