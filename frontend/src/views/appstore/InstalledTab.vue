<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from '@/composables/useMessage'
import { NButton, NIcon, NEmpty, NTag, NSpin } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {
  RefreshOutline, OpenOutline, DocumentTextOutline, TrashOutline, CloudUploadOutline
} from '@vicons/ionicons5'
import { DataTable, StatusTag, ConfirmDialog, ContentModal } from '@/components'
import * as appstoreApi from '@/api/appstore'
import { upgradeApp } from '@/api/newfeatures'
import { getContainerLogs } from '@/api/container'
import type { InstalledApp, AppCategory, AppStatus } from '@/api/appstore'
import {
  monoEl, iconBtn, btnGroup, formatTime,
  categoryLabel, categoryTagType, mapAppStatus, appStatusText
} from './shared'

const { t } = useI18n()
const message = useMessage()

/* ==================== Props / Emits ==================== */

defineProps<{
  active: boolean
}>()

/* ==================== 已安装应用列表 ==================== */

const installedApps = ref<InstalledApp[]>([])
const installedLoading = ref(false)

async function loadInstalledApps() {
  installedLoading.value = true
  try {
    const res = await appstoreApi.getInstalledApps()
    installedApps.value = res.data
  } catch (e: any) {
    message.error(e.message || t('appstore.installedTab.loadFailed'))
    installedApps.value = []
  } finally {
    installedLoading.value = false
  }
}

function openApp(row: InstalledApp) {
  if (row.access_url) {
    window.open(row.access_url, '_blank')
  } else {
    message.info(t('appstore.installedTab.noAccessUrl'))
  }
}

// 每次激活 tab 重新挂载时自行加载
onMounted(() => {
  loadInstalledApps()
})

defineExpose({ loadInstalledApps })

/* ==================== 容器日志查看 ==================== */

const logModal = ref(false)
const logLoading = ref(false)
const logAppName = ref('')
const logContent = ref('')

async function openLogs(row: InstalledApp) {
  if (!row.container_id) {
    message.warning(t('appstore.installedTab.noContainer'))
    return
  }
  logAppName.value = row.name
  logContent.value = ''
  logModal.value = true
  logLoading.value = true
  try {
    const res = await getContainerLogs(row.container_id, 300)
    logContent.value = res.data.content || t('appstore.installedTab.noLogs')
  } catch (e: any) {
    logContent.value = t('appstore.installedTab.loadLogsFailed', { error: e.message || e })
  } finally {
    logLoading.value = false
  }
}

/* ==================== 卸载确认 ==================== */

const uninstallConfirm = ref(false)
const uninstallLoading = ref(false)
const targetInstalledApp = ref<InstalledApp | null>(null)

async function doUpgrade(row: any) {
  try {
    const res = await upgradeApp(row.key)
    message.success(t('appstore.installedTab.upgraded', { name: row.name, version: res.version }))
    loadInstalledApps()
  } catch (e: any) {
    message.error(e.message || t('appstore.installedTab.upgradeFailed'))
  }
}

function confirmUninstall(row: InstalledApp) {
  targetInstalledApp.value = row
  uninstallConfirm.value = true
}

async function doUninstall() {
  if (!targetInstalledApp.value) return
  uninstallLoading.value = true
  try {
    await appstoreApi.uninstallApp(targetInstalledApp.value.key)
    message.success(
      t('appstore.installedTab.uninstalled', { name: targetInstalledApp.value.name })
    )
    uninstallConfirm.value = false
    loadInstalledApps()
  } catch (e: any) {
    message.error(e.message || t('appstore.installedTab.uninstallFailed'))
  } finally {
    uninstallLoading.value = false
  }
}

/* ==================== 列定义 ==================== */

const installedColumns = computed<DataTableColumns>(() => [
  { title: t('appstore.installedTab.columns.appName'), key: 'name', width: 180 },
  {
    title: t('appstore.installedTab.columns.version'), key: 'version', width: 120,
    render: (row: any) => monoEl(row.version)
  },
  {
    title: t('appstore.installedTab.columns.category'), key: 'category', width: 120,
    render: (row: any) =>
      h(
        NTag,
        { size: 'small', bordered: false, type: categoryTagType(row.category as AppCategory) as any },
        { default: () => categoryLabel(t, row.category as AppCategory) }
      )
  },
  {
    title: t('appstore.columns.status'), key: 'status', width: 110,
    render: (row: any) =>
      h(StatusTag, {
        status: mapAppStatus(row.status as AppStatus),
        text: appStatusText(t, row.status as AppStatus)
      })
  },
  {
    title: t('appstore.installedTab.columns.installedAt'), key: 'installed_at', width: 170,
    render: (row: any) => formatTime(row.installed_at)
  },
  {
    title: t('appstore.columns.actions'), key: 'actions', width: 220,
    render: (row: any) =>
      btnGroup([
        iconBtn({
          text: t('appstore.installedTab.actions.open'), icon: OpenOutline, type: 'primary',
          onClick: () => openApp(row)
        }),
        iconBtn({
          text: t('appstore.installedTab.actions.logs'), icon: DocumentTextOutline,
          onClick: () => openLogs(row)
        }),
        iconBtn({
          text: t('appstore.installedTab.actions.upgrade'), icon: CloudUploadOutline, type: 'warning',
          onClick: () => doUpgrade(row)
        }),
        iconBtn({
          text: t('appstore.installedTab.actions.uninstall'), icon: TrashOutline, type: 'error',
          onClick: () => confirmUninstall(row)
        })
      ])
  }
])
</script>

<template>
  <div class="installed-tab">
    <div class="tab-toolbar">
      <NButton @click="loadInstalledApps" :loading="installedLoading">
        <template #icon><NIcon><RefreshOutline /></NIcon></template>{{ t('appstore.installedTab.refresh') }}
      </NButton>
    </div>

    <DataTable
      :columns="installedColumns"
      :data="installedApps"
      :loading="installedLoading"
      :pagination="false"
      :row-key="(row: any) => row.key"
    />

    <div v-if="installedApps.length === 0 && !installedLoading" class="empty-hint">
      <NEmpty :description="t('appstore.installedTab.empty')">
        <NButton type="primary" size="small" @click="$emit('goStore')">
          {{ t('appstore.installedTab.browseApps') }}
        </NButton>
      </NEmpty>
    </div>

    <!-- 容器日志 -->
    <ContentModal v-model:show="logModal" :title="t('appstore.installedTab.logTitle', { name: logAppName })" size="xl">
      <NSpin :show="logLoading">
        <pre class="app-log-viewer">{{ logContent }}</pre>
      </NSpin>
    </ContentModal>

    <!-- 卸载确认 -->
    <ConfirmDialog
      v-model:show="uninstallConfirm"
      :title="t('appstore.installedTab.uninstallTitle')"
      :content="t('appstore.installedTab.uninstallContent', { name: targetInstalledApp?.name })"
      type="error"
      :confirm-text="t('appstore.installedTab.uninstallConfirm')"
      :loading="uninstallLoading"
      @confirm="doUninstall"
    />
  </div>
</template>

<style scoped>
.tab-toolbar {
  margin-bottom: var(--space-4);
  display: flex;
  align-items: center;
}

.empty-hint {
  padding: var(--space-8) 0;
  display: flex;
  justify-content: center;
}

.app-log-viewer {
  margin: 0;
  max-height: 60vh;
  overflow: auto;
  padding: var(--space-3);
  background: var(--log-viewer-bg);
  color: var(--log-viewer-text);
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 12px;
  line-height: 1.6;
  border-radius: 6px;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
