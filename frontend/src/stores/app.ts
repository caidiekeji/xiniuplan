import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import i18n, { ensureLocale, type Locale } from '@/locales'

export type ThemeMode = 'light' | 'dark'

/** 系统是否偏好深色（无 API 环境回落 dark，与原默认一致） */
function systemPrefersDark(): boolean {
  return typeof window !== 'undefined' && window.matchMedia
    ? window.matchMedia('(prefers-color-scheme: dark)').matches
    : true
}

export const useAppStore = defineStore('app', () => {
  // 主题策略（审计 D21）：用户显式选择过（localStorage 有值）→ 以用户为准；
  // 从未选择 → 跟随系统 prefers-color-scheme，系统变化时实时跟随。
  const storedTheme = localStorage.getItem('cloudpivot_theme') as ThemeMode | null
  const theme = ref<ThemeMode>(storedTheme || (systemPrefersDark() ? 'dark' : 'light'))
  const sidebarCollapsed = ref<boolean>(localStorage.getItem('cloudpivot_sidebar_collapsed') === 'true')
  const locale = ref<Locale>((localStorage.getItem('cloudpivot_locale') as Locale) || 'zh-CN')

  // 未显式选择时监听系统主题；一旦用户手动切换/设置，监听即失效
  let systemListener: ((e: MediaQueryListEvent) => void) | null = null
  function watchSystemTheme() {
    if (!window.matchMedia || storedTheme !== null) return
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    systemListener = (e: MediaQueryListEvent) => { theme.value = e.matches ? 'dark' : 'light' }
    mq.addEventListener('change', systemListener)
  }
  function unwatchSystemTheme() {
    if (!systemListener || !window.matchMedia) return
    window.matchMedia('(prefers-color-scheme: dark)').removeEventListener('change', systemListener)
    systemListener = null
  }
  watchSystemTheme()

  function toggleTheme() {
    unwatchSystemTheme() // 用户显式选择后不再跟随系统
    theme.value = theme.value === 'dark' ? 'light' : 'dark'
  }

  function setTheme(t: ThemeMode) {
    unwatchSystemTheme()
    theme.value = t
  }

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  // 语言切换：先确保语言包已按需加载（locales 改为按需加载，2026-10-22），
  // 再写入 locale——watch 会把 i18n.global.locale 切过去
  async function setLocale(l: Locale) {
    await ensureLocale(l)
    locale.value = l
  }

  watch(theme, (val) => localStorage.setItem('cloudpivot_theme', val))
  watch(sidebarCollapsed, (val) => localStorage.setItem('cloudpivot_sidebar_collapsed', String(val)))
  watch(locale, (val) => {
    localStorage.setItem('cloudpivot_locale', val)
    i18n.global.locale.value = val
  })

  return {
    theme,
    sidebarCollapsed,
    locale,
    toggleTheme,
    setTheme,
    toggleSidebar,
    setLocale
  }
})
