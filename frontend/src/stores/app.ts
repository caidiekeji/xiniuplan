import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import i18n, { type Locale } from '@/locales'

export type ThemeMode = 'light' | 'dark'

export const useAppStore = defineStore('app', () => {
  const theme = ref<ThemeMode>((localStorage.getItem('cloudpivot_theme') as ThemeMode) || 'dark')
  const sidebarCollapsed = ref<boolean>(localStorage.getItem('cloudpivot_sidebar_collapsed') === 'true')
  const locale = ref<Locale>((localStorage.getItem('cloudpivot_locale') as Locale) || 'zh-CN')

  function toggleTheme() {
    theme.value = theme.value === 'dark' ? 'light' : 'dark'
  }

  function setTheme(t: ThemeMode) {
    theme.value = t
  }

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setLocale(l: Locale) {
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
