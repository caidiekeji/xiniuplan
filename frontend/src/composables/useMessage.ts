import { useMessage as naiveUseMessage } from 'naive-ui'

/**
 * Toast 同类合并策略（审计 D12②）：
 * 同类型 + 同文案的 Toast 在 2 秒窗口内不再重复弹出（返回上一条的句柄，调用方 destroy 语义不变），
 * 防止轮询失败 / 批量操作连发时堆叠轰炸。
 * API 与 naive-ui 的 useMessage 同构，业务侧只需把 import 换到本文件。
 * 非字符串内容（VNode/函数）不做合并，原样透传。
 */
const DEDUP_WINDOW_MS = 2000
const recent = new Map<string, { at: number; handle: any }>()

export function useMessage() {
  const base = naiveUseMessage()

  function wrap(type: 'success' | 'error' | 'warning' | 'info' | 'loading') {
    return (content: any, options?: any) => {
      if (typeof content === 'string') {
        const key = `${type}:${content}`
        const now = Date.now()
        const hit = recent.get(key)
        if (hit && now - hit.at < DEDUP_WINDOW_MS) return hit.handle
        const handle = base[type](content, options)
        recent.set(key, { at: now, handle })
        return handle
      }
      return base[type](content, options)
    }
  }

  return {
    success: wrap('success'),
    error: wrap('error'),
    warning: wrap('warning'),
    info: wrap('info'),
    loading: wrap('loading')
  }
}
