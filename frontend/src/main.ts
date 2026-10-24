// 安全入口探测：在挂载应用前确定面板访问前缀
// - 未启用安全入口：GET /api/v1/panel/ping 返回 200，base = '/'
// - 已启用安全入口（如 admin-xyz）：只能通过 /admin-xyz 进入，
//   GET /admin-xyz/api/v1/panel/ping 返回 200，base = '/admin-xyz'
//
// 2026-10-23 优化：探测并行发起 + 根路径优先。
// 原实现是串行的——在 /dashboard 这类深层路径下，第一个探测（/dashboard/api/v1/panel/ping）
// 会命中 SPA 回退返回 index.html，解析 JSON 失败后才去试根路径，白等一个往返
// （线上实测该请求耗时 2 秒，直接卡在首屏关键路径上）。
async function probePanelBase(base: string): Promise<boolean> {
  try {
    const res = await fetch(`${base}/api/v1/panel/ping`, { method: 'GET' })
    if (!res.ok) return false
    const data = await res.json()
    return !!(data && data.ok)
  } catch {
    return false
  }
}

async function detectPanelBase(): Promise<string> {
  const firstSeg = window.location.pathname.split('/')[1] || ''

  // 根路径探测优先：命中即代表"未启用安全入口"，立即返回，
  // 不必等待入口探测的结果
  const rootProbe = probePanelBase('')
  if (!firstSeg) {
    return (await rootProbe) ? '' : 'BLOCKED'
  }
  if (await rootProbe) {
    return ''
  }
  // 根路径不通，再看是否走了安全入口前缀
  return (await probePanelBase(`/${firstSeg}`)) ? `/${firstSeg}` : 'BLOCKED'
}

/**
 * 解析面板访问前缀。
 *
 * 首选后端注入的 window.__PANEL_BASE__（后端在返回 index.html 时写入，
 * 见 backend/cmd/agent/main.go 的 panelIndexHTML）——零请求、零等待。
 * 仅当注入缺失（旧缓存的 HTML）时才回退到运行时探测。
 */
async function resolvePanelBase(): Promise<string> {
  const injected = (window as any).__PANEL_BASE__
  if (typeof injected === 'string') return injected
  return detectPanelBase()
}

async function bootstrap() {
  // 首屏 import 并行化（2026-10-22 性能修复）：
  // 这些模块的加载都不依赖 __PANEL_BASE__，先并行触发下载，避免串行瀑布
  // （此前 naive-ui 377KB 要等 locales 主包 316KB 下载完才发起，白等 ~700ms）。
  // 唯独 router 模块顶层会读 __PANEL_BASE__，必须等 base 就绪后再 import。
  const localesP = import('./locales')
  const vueP = import('vue')
  const piniaP = import('pinia')
  const naiveP = import('naive-ui')
  const appP = import('./App.vue')
  // 全局样式改由 index.html 静态 link 加载（2026-10-23）：动态 import 会让
  // 页面在 JS 就绪前完全无样式，深色主题下呈现为整屏黑屏

  const base = await resolvePanelBase()
  // i18n 实例来自 locales 模块（模块顶层创建并导出，与 main.ts 无循环依赖），
  // 需在 BLOCKED 分支渲染提示前就绪
  const { default: i18n, initLocale } = await localesP
  // 用户保存的语言非默认中文时，按需加载语言包并切换（默认语言同步可用，无闪烁）
  await initLocale()
  if (base === 'BLOCKED') {
    // 启用了安全入口但未通过入口访问：渲染提示页（不暴露面板存在性的详细信息）
    // 设计豁免：此页在 token/main.css 加载之前执行（挂载前拦截），内联 hex 无法引用
    // CSS 变量；且该隔离页不加载任何主题资源，不属于主题体系管辖范围
    document.body.innerHTML = `
      <div style="min-height:100vh;display:flex;align-items:center;justify-content:center;background:#0e1117;color:#8b8f98;font-family:system-ui,sans-serif;">
        <div style="text-align:center;">
          <div style="font-size:64px;margin-bottom:16px;">404</div>
          <div>${i18n.global.t('common.errors.pageNotFound')}</div>
        </div>
      </div>`
    return
  }
  ;(window as any).__PANEL_BASE__ = base

  const { createApp } = await vueP
  const { createPinia } = await piniaP
  const { default: naive } = await naiveP
  const App = (await appP).default
  const router = (await import('./router')).default

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
        globalMsg.error(i18n.global.t('common.errors.appError', { msg }))
      } else {
        window.alert(i18n.global.t('common.errors.appError', { msg }))
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
