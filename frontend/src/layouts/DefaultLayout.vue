<script setup lang="ts">
import { computed, h, ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NLayout, NLayoutSider, NMenu, NIcon, NButton, NDropdown, NAvatar, NBadge, NDivider } from 'naive-ui'
import { StorefrontOutline, FolderOpenOutline, FileTrayFullOutline, SettingsOutline, SunnyOutline, MoonOutline, MenuOutline, NotificationsOutline, LogOutOutline, PersonOutline, ChevronBackOutline, LockClosedOutline, AlertCircleOutline, PeopleOutline, LanguageOutline, ExtensionPuzzleOutline, PulseOutline, GitBranchOutline, ShieldCheckmarkOutline, HardwareChipOutline, LayersOutline, SparklesOutline, SearchOutline } from '@vicons/ionicons5'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { LOCALES, type Locale } from '@/locales'
import RhinoLogo from '@/components/RhinoLogo.vue'
import CommandPalette, { type CommandItem } from '@/components/CommandPalette.vue'
import {
  RackIcon, GlobeIcon, DbIcon, CubeIcon, TerminalIcon,
  GaugeIcon, TapeIcon, FirewallIcon, CertIcon, CronIcon
} from '@/components/PanelIcons'
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

// 侧栏 24 项按语义分 5 组（审计 P1-03：平铺导航无分组，找功能靠扫读）
// 组内顺序即菜单顺序；新增菜单项时归入对应组，组内 ≤6 项，超过则考虑再拆组
const menuOptions = computed(() => [
  {
    type: 'group' as const,
    label: t('menu.groups.overview'),
    key: 'group-overview',
    children: [
      { label: t('menu.dashboard'), key: '/dashboard', icon: renderIcon(RackIcon) },
      { label: t('menu.monitor'), key: '/monitor', icon: renderIcon(GaugeIcon) },
      { label: t('menu.probe'), key: '/probe', icon: renderIcon(PulseOutline) }
    ]
  },
  {
    type: 'group' as const,
    label: t('menu.groups.apps'),
    key: 'group-apps',
    children: [
      { label: t('menu.website'), key: '/website', icon: renderIcon(GlobeIcon) },
      { label: t('menu.deploy'), key: '/deploy', icon: renderIcon(GitBranchOutline) },
      { label: t('menu.database'), key: '/database', icon: renderIcon(DbIcon) },
      { label: t('menu.container'), key: '/container', icon: renderIcon(CubeIcon) },
      { label: t('menu.appstore'), key: '/appstore', icon: renderIcon(StorefrontOutline) },
      { label: t('menu.plugins'), key: '/plugins', icon: renderIcon(ExtensionPuzzleOutline) }
    ]
  },
  {
    type: 'group' as const,
    label: t('menu.groups.ops'),
    key: 'group-ops',
    children: [
      { label: t('menu.file'), key: '/file', icon: renderIcon(FolderOpenOutline) },
      { label: t('menu.terminal'), key: '/terminal', icon: renderIcon(TerminalIcon) },
      { label: t('menu.cron'), key: '/cron', icon: renderIcon(CronIcon) },
      { label: t('menu.daemon'), key: '/daemon', icon: renderIcon(HardwareChipOutline) },
      { label: t('menu.backup'), key: '/backup', icon: renderIcon(TapeIcon) }
    ]
  },
  {
    type: 'group' as const,
    label: t('menu.groups.security'),
    key: 'group-security',
    children: [
      { label: t('menu.firewall'), key: '/firewall', icon: renderIcon(FirewallIcon) },
      { label: t('menu.protection'), key: '/protection', icon: renderIcon(ShieldCheckmarkOutline) },
      { label: t('menu.ssh'), key: '/ssh', icon: renderIcon(LockClosedOutline) },
      { label: t('menu.alert'), key: '/alert', icon: renderIcon(AlertCircleOutline) },
      { label: t('menu.ssl'), key: '/ssl', icon: renderIcon(CertIcon) }
    ]
  },
  {
    type: 'group' as const,
    label: t('menu.groups.system'),
    key: 'group-system',
    children: [
      { label: t('menu.users'), key: '/users', icon: renderIcon(PeopleOutline) },
      { label: t('menu.audit'), key: '/audit', icon: renderIcon(FileTrayFullOutline) },
      { label: t('menu.nodes'), key: '/nodes', icon: renderIcon(LayersOutline) },
      { label: t('menu.ai'), key: '/ai', icon: renderIcon(SparklesOutline) },
      { label: t('menu.settings'), key: '/settings', icon: renderIcon(SettingsOutline) }
    ]
  }
])

function renderIcon(icon: any) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

// 命令面板条目：菜单项与分组同源展开，保证侧栏改分组时面板自动同步
const paletteItems = computed<CommandItem[]>(() => {
  const items: CommandItem[] = []
  menuOptions.value.forEach((group: any) => {
    const groupLabel = group.label as string
    ;(group.children || []).forEach((child: any) => {
      items.push({ key: child.key, label: child.label, icon: child.icon, group: groupLabel })
    })
  })
  // 动作类命令：非路由，key 以 action: 前缀区分
  items.push(
    { key: 'action:theme', label: appStore.theme === 'dark' ? t('common.toggleLight') : t('common.toggleDark'), icon: renderIcon(appStore.theme === 'dark' ? SunnyOutline : MoonOutline) }
  )
  return items
})

function handlePaletteSelect(key: string) {
  if (key === 'action:theme') {
    appStore.toggleTheme()
    return
  }
  router.push(key)
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

/* ==================== 命令面板 ==================== */
const paletteOpen = ref(false)

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
  <NLayout has-sider style="height: 100vh; overflow: hidden" :native-scrollbar="true">
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

      <nav class="sider-nav" :aria-label="t('menu.navLabel')">
        <NMenu
          :value="activeKey"
          :options="menuOptions"
          :collapsed="collapsed"
          :collapsed-width="60"
          :collapsed-icon-size="20"
          @update:value="handleMenuSelect"
          class="sider-menu"
        />
      </nav>
    </NLayoutSider>

    <!-- 主区域：显式锁高 + 禁外层滚动，保证侧栏永不随内容滚动 -->
    <NLayout style="height: 100vh; overflow: hidden">
      <!-- 顶栏（语义 landmark：header，D22） -->
      <header class="layout-header">
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
          <!-- 命令面板入口（Cmd/Ctrl+K），桌面端显示 -->
          <NButton
            quaternary
            size="small"
            class="header-cmd-btn"
            :title="t('common.commandPalette')"
            @click="() => paletteOpen = true"
          >
            <template #icon>
              <NIcon><SearchOutline /></NIcon>
            </template>
            <span class="header-cmd__kbd">Ctrl K</span>
          </NButton>

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
      </header>

      <!-- 内容区（语义 landmark：main；每页由布局统一注入 h1 = 当前面包屑末级标题，D22） -->
      <main class="layout-content">
        <h1 class="sr-only">{{ breadcrumb[breadcrumb.length - 1] || t('auth.brandName') }}</h1>
        <router-view />
      </main>
    </NLayout>

    <!-- 命令面板（Ctrl/Cmd+K 或顶栏入口唤起） -->
    <CommandPalette
      v-model:show="paletteOpen"
      :items="paletteItems"
      @select="handlePaletteSelect"
    />
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

.sider-nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.sider-menu {
  flex: 1;
  background: transparent;
  padding: var(--space-2) 0;
}

/* 分组标签：小字 + 弱化色，折叠时隐藏 */
.sider-menu :deep(.n-menu-group-label) {
  font-size: var(--fs-caption);
  font-weight: 500;
  color: var(--text-quaternary);
  letter-spacing: 0.02em;
}

.sider-menu :deep(.n-menu-group-label > .n-menu-group-label__icon) {
  display: none;
}

/* 折叠态：60px 宽容不下分组文字，整条隐藏，仅靠间距区隔 */
.sider-menu.n-menu--collapsed :deep(.n-menu-group-label) {
  display: none;
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
  color: var(--text-quaternary);
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

.header-cmd-btn {
  padding-left: var(--space-2);
  padding-right: var(--space-2);
  color: var(--text-quaternary);
  gap: var(--space-2);
}

.header-cmd__kbd {
  font-family: var(--font-mono);
  font-size: var(--fs-caption);
  background: var(--surface2);
  border: 1px solid var(--line);
  border-radius: var(--radius-sm);
  padding: 0 var(--space-1);
  line-height: 1.6;
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
  .layout-header {
    /* 刘海/挖孔屏竖屏避让（viewport-fit=cover 后生效，D20） */
    height: calc(var(--header-height) + env(safe-area-inset-top, 0px));
    padding-top: env(safe-area-inset-top, 0px);
  }
  .header-collapse-btn {
    display: inline-flex;
  }
  .header-breadcrumb {
    display: none;
  }
  .layout-content {
    padding: var(--space-3);
    /* 全面屏底部横条避让：保证滚动到底后最后一行操作不被遮挡（D20） */
    padding-bottom: calc(var(--space-3) + env(safe-area-inset-bottom, 0px));
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
  /* 触屏无键盘，命令面板入口不展示（Ctrl+K 仍可用外接键盘） */
  .header-cmd-btn {
    display: none;
  }
}

/* 移动端抽屉侧边栏 */
.mobile-mask {
  position: fixed;
  inset: 0;
  background: var(--overlay);
  z-index: var(--z-drawer);
  animation: fadeIn var(--t-fast) var(--ease);
}

.layout-sider--mobile {
  position: fixed !important;
  left: 0;
  top: 0;
  bottom: 0;
  z-index: var(--z-modal);
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
