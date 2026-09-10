<script setup lang="ts">
import { computed, h, ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  NLayout, NLayoutSider, NLayoutHeader, NLayoutContent,
  NMenu, NIcon, NButton, NDropdown, NAvatar, NSpace,
  NBadge, NDivider
} from 'naive-ui'
import {
  GridOutline, GlobeOutline, ServerOutline,
  CubeOutline, StorefrontOutline, FolderOpenOutline,
  TerminalOutline, AnalyticsOutline, SaveOutline,
  ShieldHalfOutline, KeyOutline, FileTrayFullOutline,
  SettingsOutline, SunnyOutline, MoonOutline,
  MenuOutline, NotificationsOutline, LogOutOutline,
  PersonOutline, ChevronBackOutline, ChevronForwardOutline,
  TimeOutline, LockClosedOutline, AlertCircleOutline, PeopleOutline,
  LanguageOutline, ExtensionPuzzleOutline
} from '@vicons/ionicons5'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { LOCALES, type Locale } from '@/locales'
import RhinoLogo from '@/components/RhinoLogo.vue'
import { getAlertHistory } from '@/api/alert'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()

const collapsed = computed(() => appStore.sidebarCollapsed)
const activeKey = computed(() => route.path)

/* ==================== 移动端适配 ==================== */
const isMobile = ref(false)
const mobileDrawerOpen = ref(false)

function checkMobile() {
  isMobile.value = window.innerWidth <= 768
  if (!isMobile.value) {
    mobileDrawerOpen.value = false
  }
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  fetchNotificationCount()
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', checkMobile)
})

// 路由变化时关闭移动端抽屉
watch(() => route.path, () => {
  if (isMobile.value) {
    mobileDrawerOpen.value = false
  }
})

function toggleMobileDrawer() {
  mobileDrawerOpen.value = !mobileDrawerOpen.value
}

const menuOptions = computed(() => [
  { label: t('menu.dashboard'), key: '/dashboard', icon: renderIcon(GridOutline) },
  { label: t('menu.website'), key: '/website', icon: renderIcon(GlobeOutline) },
  { label: t('menu.database'), key: '/database', icon: renderIcon(ServerOutline) },
  { label: t('menu.container'), key: '/container', icon: renderIcon(CubeOutline) },
  { label: t('menu.appstore'), key: '/appstore', icon: renderIcon(StorefrontOutline) },
  { label: t('menu.plugins'), key: '/plugins', icon: renderIcon(ExtensionPuzzleOutline) },
  { label: t('menu.file'), key: '/file', icon: renderIcon(FolderOpenOutline) },
  { label: t('menu.terminal'), key: '/terminal', icon: renderIcon(TerminalOutline) },
  { label: t('menu.cron'), key: '/cron', icon: renderIcon(TimeOutline) },
  { label: t('menu.monitor'), key: '/monitor', icon: renderIcon(AnalyticsOutline) },
  { label: t('menu.backup'), key: '/backup', icon: renderIcon(SaveOutline) },
  { label: t('menu.firewall'), key: '/firewall', icon: renderIcon(ShieldHalfOutline) },
  { label: t('menu.ssh'), key: '/ssh', icon: renderIcon(LockClosedOutline) },
  { label: t('menu.alert'), key: '/alert', icon: renderIcon(AlertCircleOutline) },
  { label: t('menu.ssl'), key: '/ssl', icon: renderIcon(KeyOutline) },
  { label: t('menu.users'), key: '/users', icon: renderIcon(PeopleOutline) },
  { label: t('menu.audit'), key: '/audit', icon: renderIcon(FileTrayFullOutline) },
  { label: t('menu.settings'), key: '/settings', icon: renderIcon(SettingsOutline) }
])

function renderIcon(icon: any) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

function handleMenuSelect(key: string) {
  router.push(key)
}

function toggleSidebar() {
  appStore.toggleSidebar()
}

const userDropdownOptions = computed(() => [
  { label: t('auth.profile'), key: 'profile', icon: renderIcon(PersonOutline) },
  { label: t('auth.systemSettings'), key: 'settings', icon: renderIcon(SettingsOutline) },
  { type: 'divider' as const, key: 'd1' },
  { label: t('auth.logout'), key: 'logout', icon: renderIcon(LogOutOutline) }
])

function handleUserSelect(key: string) {
  if (key === 'logout') {
    authStore.logout()
    router.push('/login')
  } else if (key === 'settings') {
    router.push('/settings')
  }
}

const languageOptions = computed(() =>
  LOCALES.map((l) => ({
    label: l.label,
    key: l.value
  }))
)

const currentLocaleLabel = computed(() => {
  const found = LOCALES.find((l) => l.value === appStore.locale)
  return found ? found.label : '中文'
})

function handleLocaleSelect(key: string) {
  appStore.setLocale(key as Locale)
}

const breadcrumb = computed(() => {
  const matched = route.matched.filter((r) => r.meta.title)
  return matched.map((r) => {
    const key = r.meta.title as string
    return t(key)
  })
})

/* ==================== 消息提醒 ==================== */
const notificationCount = ref(0)

async function fetchNotificationCount() {
  try {
    const res = await getAlertHistory({ status: 'triggered', page: 1, pageSize: 1 })
    notificationCount.value = res.data.total || 0
  } catch {
    notificationCount.value = 0
  }
}

onMounted(() => {
  fetchNotificationCount()
})
</script>

<template>
  <NLayout has-sider style="height: 100vh" :native-scrollbar="false">
    <!-- 移动端遮罩层 -->
    <div
      v-if="isMobile && mobileDrawerOpen"
      class="mobile-mask"
      @click="mobileDrawerOpen = false"
    />

    <!-- 侧边栏 -->
    <NLayoutSider
      bordered
      :collapsed="isMobile ? false : collapsed"
      :collapsed-width="60"
      :width="220"
      :show-trigger="false"
      class="layout-sider"
      :class="{ 'layout-sider--mobile': isMobile, 'layout-sider--mobile-open': isMobile && mobileDrawerOpen }"
    >
      <div class="sider-logo" @click="collapsed && toggleSidebar()">
        <div class="sider-logo__mark">
          <RhinoLogo :size="34" />
        </div>
        <transition name="fade">
          <span v-if="!collapsed" class="sider-logo__text">{{ t('auth.brandName') }}</span>
        </transition>
        <NButton v-if="!collapsed" quaternary size="small" circle @click.stop="toggleSidebar" class="sider-collapse-btn" :title="t('common.collapse')">
          <template #icon>
            <NIcon><ChevronBackOutline /></NIcon>
          </template>
        </NButton>
      </div>

      <NMenu
        :value="activeKey"
        :options="menuOptions"
        :collapsed="collapsed"
        :collapsed-width="60"
        :collapsed-icon-size="20"
        @update:value="handleMenuSelect"
        class="sider-menu"
      />
    </NLayoutSider>

    <!-- 主区域 -->
    <NLayout>
      <!-- 顶栏 -->
      <NLayoutHeader bordered class="layout-header">
        <div class="header-left">
          <NButton
            quaternary
            size="small"
            @click="isMobile ? toggleMobileDrawer() : toggleSidebar()"
            class="header-collapse-btn"
          >
            <template #icon>
              <NIcon><MenuOutline /></NIcon>
            </template>
          </NButton>
          <div class="header-breadcrumb">
            <span v-for="(item, idx) in breadcrumb" :key="idx" class="breadcrumb-item">
              <span v-if="idx > 0" class="breadcrumb-sep">/</span>
              {{ item }}
            </span>
          </div>
        </div>

        <div class="header-right">
          <!-- 语言切换 -->
          <NDropdown
            :options="languageOptions"
            trigger="click"
            @select="handleLocaleSelect"
          >
            <NButton quaternary size="small" class="header-lang-btn">
              <template #icon>
                <NIcon><LanguageOutline /></NIcon>
              </template>
              <span class="header-lang__label">{{ currentLocaleLabel }}</span>
            </NButton>
          </NDropdown>

          <NButton quaternary size="small" circle @click="appStore.toggleTheme" :title="appStore.theme === 'dark' ? t('common.toggleLight') : t('common.toggleDark')">
            <NIcon><component :is="appStore.theme === 'dark' ? SunnyOutline : MoonOutline" /></NIcon>
          </NButton>

          <NButton quaternary size="small" circle :title="t('common.notifications')" @click="router.push('/alert')">
            <NBadge :value="notificationCount" :max="99" size="small" :show="notificationCount > 0">
              <NIcon><NotificationsOutline /></NIcon>
            </NBadge>
          </NButton>

          <NDivider vertical />

          <NDropdown
            :options="userDropdownOptions"
            trigger="click"
            @select="handleUserSelect"
          >
            <div class="header-user">
              <NAvatar round size="small" style="background: var(--acc2); color: var(--bg)">
                {{ authStore.userInfo?.username?.charAt(0) || 'A' }}
              </NAvatar>
              <span class="header-user__name">{{ authStore.userInfo?.username || 'admin' }}</span>
            </div>
          </NDropdown>
        </div>
      </NLayoutHeader>

      <!-- 内容区 -->
      <NLayoutContent class="layout-content">
        <router-view />
      </NLayoutContent>
    </NLayout>
  </NLayout>
</template>

<style scoped>
.layout-sider {
  background: var(--bg2);
  border-right: 1px solid var(--line) !important;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.sider-logo {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-3);
  height: var(--header-height);
  border-bottom: 1px solid var(--line);
  cursor: default;
}

.layout-sider.n-layout-sider--collapsed .sider-logo {
  justify-content: center;
  padding: var(--space-3);
  cursor: pointer;
}

.sider-logo__mark {
  width: var(--logo-size);
  height: var(--logo-size);
  border-radius: var(--radius-md);
  background: var(--surface2);
  border: 1px solid var(--line);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.sider-logo__text {
  font-size: var(--fs-title);
  font-weight: 700;
  color: var(--ink);
  letter-spacing: -0.01em;
  white-space: nowrap;
  flex: 1;
}

.sider-collapse-btn {
  flex-shrink: 0;
  margin-left: auto;
}

.sider-menu {
  flex: 1;
  background: transparent;
  padding: var(--space-2) 0;
}

.sider-menu :deep(.n-menu-item) {
  margin: var(--space-1) var(--space-2);
  border-radius: var(--radius-md);
}

.sider-menu :deep(.n-menu-item-content) {
  height: var(--menu-item-height);
  border-radius: var(--radius-md);
}

.sider-menu :deep(.n-menu-item-content--selected) {
  background: var(--acc-soft);
}

.sider-menu :deep(.n-menu-item-content--selected .n-menu-item-content__icon) {
  color: var(--acc);
}

.sider-menu :deep(.n-menu-item-content--selected .n-menu-item-content-header) {
  color: var(--acc);
  font-weight: 600;
}

.layout-header {
  height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--space-5);
  background: var(--bg);
  border-bottom: 1px solid var(--line) !important;
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  /* 允许面包屑截断，防止长路径把 header-right 挤出视口 */
  min-width: 0;
  flex: 1;
}

.header-collapse-btn {
  display: none;
}

.header-breadcrumb {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.breadcrumb-item {
  font-size: var(--fs-body);
  color: var(--sub);
  white-space: nowrap;
}

.breadcrumb-item:last-child {
  max-width: 420px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.breadcrumb-item:last-child {
  color: var(--ink);
  font-weight: 500;
}

.breadcrumb-sep {
  color: var(--faint);
  margin-right: var(--space-2);
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.header-lang-btn {
  padding-left: var(--space-2);
  padding-right: var(--space-3);
}

.header-lang__label {
  font-size: var(--fs-body);
  color: var(--ink);
}

.header-icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.header-user {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-1) var(--space-3);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: background var(--t-fast) var(--ease);
}

.header-user:hover {
  background: var(--surface2);
}

.header-user__name {
  font-size: var(--fs-body);
  color: var(--ink);
}

.layout-content {
  /* 关键：给内容区固定可视高度让它自己滚动。
     否则内容会把内层 NLayout 撑高，导致最外层滚动、侧栏跟着内容滚走 */
  height: calc(100vh - var(--header-height));
  padding: var(--space-5);
  background: var(--bg);
  overflow-y: auto;
}

@media (max-width: 768px) {
  .header-collapse-btn {
    display: inline-flex;
  }
  .header-breadcrumb {
    display: none;
  }
  .layout-content {
    padding: var(--space-3);
  }
  .header-lang__label {
    display: none;
  }
  .header-user__name {
    display: none;
  }
  .header-right {
    gap: var(--space-1);
  }
}

/* 移动端抽屉侧边栏 */
.mobile-mask {
  position: fixed;
  inset: 0;
  background: var(--overlay);
  z-index: 999;
  animation: fadeIn var(--t-fast) var(--ease);
}

.layout-sider--mobile {
  position: fixed !important;
  left: 0;
  top: 0;
  bottom: 0;
  z-index: 1000;
  transform: translateX(-100%);
  transition: transform var(--t-normal) var(--ease);
  box-shadow: var(--shadow-lg);
}

.layout-sider--mobile-open {
  transform: translateX(0);
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>
