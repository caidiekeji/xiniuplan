<script setup lang="ts">
import { computed, watch } from 'vue'
import { NConfigProvider, NMessageProvider, darkTheme, zhCN, dateZhCN } from 'naive-ui'
import { useAppStore } from '@/stores/app'
import { useThemeTokens } from '@/composables/useThemeTokens'
import ErrorBoundary from '@/components/ErrorBoundary.vue'

const appStore = useAppStore()
const { tokens, refresh } = useThemeTokens()

const theme = computed(() => (appStore.theme === 'dark' ? darkTheme : null))

// Naive UI 主题覆盖：全部从 CSS 变量读取，不硬编码色值
const themeOverrides = computed(() => ({
  common: {
    primaryColor: tokens.value.acc || '#1ba784',
    primaryColorHover: tokens.value.accHover || '#23c096',
    primaryColorPressed: tokens.value.accPressed || '#147760',
    primaryColorSuppl: tokens.value.acc || '#1ba784',
    successColor: tokens.value.ok || '#2dd4a8',
    warningColor: tokens.value.warn || '#1ba784',
    errorColor: tokens.value.err || '#e85d4d',
    infoColor: tokens.value.info || '#60a5fa',
    borderRadius: tokens.value.radiusMd || '8px',
    fontFamily: tokens.value.fontSans || 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif',
    fontFamilyMono: tokens.value.fontMono || '"JetBrains Mono", "SF Mono", "Cascadia Code", Consolas, monospace'
  },
  Menu: {
    itemColorActive: 'transparent',
    itemColorActiveHover: 'transparent',
    itemColorHover: 'rgba(255,255,255,0.04)'
  }
}))

watch(
  () => appStore.theme,
  (val) => {
    document.documentElement.setAttribute('data-theme', val)
    // 主题切换后重新读取 CSS 变量
    setTimeout(refresh, 50)
  },
  { immediate: true }
)
</script>

<template>
  <NConfigProvider :theme="theme" :theme-overrides="themeOverrides" :locale="zhCN" :date-locale="dateZhCN">
    <NMessageProvider>
      <ErrorBoundary>
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </ErrorBoundary>
    </NMessageProvider>
  </NConfigProvider>
</template>
