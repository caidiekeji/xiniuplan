<script setup lang="ts">
import { computed, watch } from 'vue'
import {
  NConfigProvider, NMessageProvider, NDialogProvider, darkTheme,
  zhCN, dateZhCN, enUS, dateEnUS, jaJP, dateJaJP
} from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useThemeTokens } from '@/composables/useThemeTokens'
import ErrorBoundary from '@/components/ErrorBoundary.vue'

const appStore = useAppStore()
const { tokens, refresh } = useThemeTokens()
const { locale } = useI18n()

// Naive UI 组件内置文案（分页"条/页"、日期选择器等）随面板语言切换
const naiveLocale = computed(() => {
  switch (locale.value) {
    case 'en-US': return { locale: enUS, dateLocale: dateEnUS }
    case 'ja-JP': return { locale: jaJP, dateLocale: dateJaJP }
    default: return { locale: zhCN, dateLocale: dateZhCN }
  }
})

const theme = computed(() => (appStore.theme === 'dark' ? darkTheme : null))

// Naive UI 主题覆盖：全部从 CSS 变量读取，不硬编码色值
const themeOverrides = computed(() => ({
  common: {
    primaryColor: tokens.value.acc || '#1ba784',
    primaryColorHover: tokens.value.accHover || '#23c096',
    primaryColorPressed: tokens.value.accPressed || '#147760',
    primaryColorSuppl: tokens.value.acc || '#1ba784',
    successColor: tokens.value.ok || '#2dd4a8',
    warningColor: tokens.value.warn || '#f59e0b',
    errorColor: tokens.value.err || '#e85d4d',
    infoColor: tokens.value.info || '#60a5fa',
    borderRadius: tokens.value.radiusMd || '8px',
    fontFamily: tokens.value.fontSans || 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif',
    fontFamilyMono: tokens.value.fontMono || '"JetBrains Mono", "SF Mono", "Cascadia Code", Consolas, monospace'
  },
  Button: {
    // 主按钮底色走 --btn-bg 组件级 token（浅色加深保证白字对比 ≥4.5:1，深色回归品牌主色）
    // 注意：主按钮文字色由 naive 内置 textColorPrimary（白色）提供，勿覆盖 textColor——
    // textColor 管的是 default 型（透明底）按钮，覆盖成白色会让浅色主题下"刷新"等按钮白字不可见
    colorPrimary: tokens.value.btnBg || '#147760',
    colorHover: tokens.value.btnBgHover || '#105c4a',
    colorPressed: tokens.value.btnBgHover || '#105c4a',
    colorFocus: tokens.value.btnBgHover || '#105c4a'
  },
  Menu: {
    itemColorActive: 'transparent',
    itemColorActiveHover: 'transparent',
    // 浅色主题下 4% 白叠加 ≈ 无感，改用 surface2 保证双主题 hover 反馈都可见
    itemColorHover: tokens.value.surface2 || '#f0f3f8'
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
  <NConfigProvider :theme="theme" :theme-overrides="themeOverrides" :locale="naiveLocale.locale" :date-locale="naiveLocale.dateLocale">
    <NMessageProvider>
      <NDialogProvider>
        <ErrorBoundary>
          <router-view v-slot="{ Component }">
            <transition name="fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </ErrorBoundary>
      </NDialogProvider>
    </NMessageProvider>
  </NConfigProvider>
</template>
