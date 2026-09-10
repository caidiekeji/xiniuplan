// 安全入口探测：在挂载应用前确定面板访问前缀
// - 未启用安全入口：GET /api/v1/panel/ping 返回 200，base = '/'
// - 已启用安全入口（如 admin-xyz）：只能通过 /admin-xyz 进入，
//   GET /admin-xyz/api/v1/panel/ping 返回 200，base = '/admin-xyz'
async function detectPanelBase(): Promise<string> {
  const firstSeg = window.location.pathname.split('/')[1] || ''

  // 1) 若当前路径带一级前缀，优先探测"安全入口模式"
  if (firstSeg) {
    try {
      const res = await fetch(`/${firstSeg}/api/v1/panel/ping`, { method: 'GET' })
      if (res.ok) {
        const data = await res.json()
        if (data && data.ok) return `/${firstSeg}`
      }
    } catch {
      /* 探测失败继续走默认探测 */
    }
  }

  // 2) 探测无入口模式
  try {
    const res = await fetch('/api/v1/panel/ping', { method: 'GET' })
    if (res.ok) {
      const data = await res.json()
      if (data && data.ok) return ''
    }
  } catch {
    /* ignore */
  }

  // 3) 都不通：可能启用了安全入口但用户访问了根路径
  return 'BLOCKED'
}

async function bootstrap() {
  const base = await detectPanelBase()
  if (base === 'BLOCKED') {
    // 启用了安全入口但未通过入口访问：渲染提示页（不暴露面板存在性的详细信息）
    document.body.innerHTML = `
      <div style="min-height:100vh;display:flex;align-items:center;justify-content:center;background:#0e1117;color:#8b8f98;font-family:system-ui,sans-serif;">
        <div style="text-align:center;">
          <div style="font-size:64px;margin-bottom:16px;">404</div>
          <div>页面不存在或访问路径无效</div>
        </div>
      </div>`
    return
  }
  ;(window as any).__PANEL_BASE__ = base

  const { createApp } = await import('vue')
  const { createPinia } = await import('pinia')
  const { default: naive } = await import('naive-ui')
  const App = (await import('./App.vue')).default
  const router = (await import('./router')).default
  const { default: i18n } = await import('./locales')
  await import('./styles/main.css')

  const app = createApp(App)

  app.use(createPinia())
  app.use(router)
  app.use(i18n)
  app.use(naive)

  // 全局错误捕获：未被 ErrorBoundary 捕获的 Vue 错误
  app.config.errorHandler = (err, instance, info) => {
    window.console.error('[GlobalErrorHandler]', err, info)
    // 尝试通过 Naive UI message 提示；若 provider 不可用则降级为 alert
    try {
      const msg = err instanceof Error ? err.message : String(err)
      const globalMsg = (app.config.globalProperties as any).$message
      if (globalMsg && typeof globalMsg.error === 'function') {
        globalMsg.error(`应用错误：${msg}`)
      } else {
        window.alert(`应用错误：${msg}`)
      }
    } catch {
      // 兜底：仅 console 输出
    }
  }

  // 全局未处理 Promise 拒绝
  window.addEventListener('unhandledrejection', (event) => {
    window.console.error('[UnhandledRejection]', event.reason)
  })

  app.mount('#app')
}

bootstrap()
