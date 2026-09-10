<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
import { useMessage } from 'naive-ui'
import { NButton, NSpace, NIcon, NEmpty, NTag, NModal, NSpin } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {
  RefreshOutline, OpenOutline, DocumentTextOutline, TrashOutline
} from '@vicons/ionicons5'
import { DataTable, StatusTag, ConfirmDialog } from '@/components'
import * as appstoreApi from '@/api/appstore'
import { getContainerLogs } from '@/api/container'
import type { InstalledApp, AppCategory, AppStatus } from '@/api/appstore'
import {
  monoEl, iconBtn, btnGroup, formatTime,
  categoryLabel, categoryTagType, mapAppStatus, appStatusText
} from './shared'

const message = useMessage()

/* ==================== Props / Emits ==================== */

defineProps<{
  active: boolean
}>()

const emit = defineEmits<{
  (e: 'goStore'): void
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
    message.error(e.message || '加载已安装应用失败')
    installedApps.value = []
  } finally {
    installedLoading.value = false
  }
}

function openApp(row: InstalledApp) {
  if (row.access_url) {
    window.open(row.access_url, '_blank')
  } else {
    message.info('该应用暂无访问地址')
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
    message.warning('该应用没有关联的容器，无法查看日志')
    return
  }
  logAppName.value = row.name
  logContent.value = ''
  logModal.value = true
  logLoading.value = true
  try {
    const res = await getContainerLogs(row.container_id, 300)
    logContent.value = res.data.content || '（暂无日志）'
  } catch (e: any) {
    logContent.value = `加载日志失败：${e.message || e}`
  } finally {
    logLoading.value = false
  }
}

/* ==================== 卸载确认 ==================== */

const uninstallConfirm = ref(false)
const uninstallLoading = ref(false)
const targetInstalledApp = ref<InstalledApp | null>(null)

function confirmUninstall(row: InstalledApp) {
  targetInstalledApp.value = row
  uninstallConfirm.value = true
}

async function doUninstall() {
  if (!targetInstalledApp.value) return
  uninstallLoading.value = true
  try {
    await appstoreApi.uninstallApp(targetInstalledApp.value.key)
    message.success(`${targetInstalledApp.value.name} 已卸载`)
    uninstallConfirm.value = false
    loadInstalledApps()
  } catch (e: any) {
    message.error(e.message || '卸载失败')
  } finally {
    uninstallLoading.value = false
  }
}

/* ==================== 列定义 ==================== */

const installedColumns: DataTableColumns = [
  { title: '应用名', key: 'name', width: 180 },
  { title: '版本', key: 'version', width: 120, render: (row: any) => monoEl(row.version) },
  {
    title: '分类', key: 'category', width: 120,
    render: (row: any) =>
      h(
        NTag,
        { size: 'small', bordered: false, type: categoryTagType(row.category as AppCategory) as any },
        { default: () => categoryLabel(row.category as AppCategory) }
      )
  },
  {
    title: '状态', key: 'status', width: 110,
    render: (row: any) =>
      h(StatusTag, {
        status: mapAppStatus(row.status as AppStatus),
        text: appStatusText(row.status as AppStatus)
      })
  },
  {
    title: '安装时间', key: 'installed_at', width: 170,
    render: (row: any) => formatTime(row.installed_at)
  },
  {
    title: '操作', key: 'actions', width: 220,
    render: (row: any) =>
      btnGroup([
        iconBtn({
          text: '打开', icon: OpenOutline, type: 'primary',
          onClick: () => openApp(row)
        }),
        iconBtn({
          text: '日志', icon: DocumentTextOutline,
          onClick: () => openLogs(row)
        }),
        iconBtn({
          text: '卸载', icon: TrashOutline, type: 'error',
          onClick: () => confirmUninstall(row)
        })
      ])
  }
]
</script>

<template>
  <div class="installed-tab">
    <div class="tab-toolbar">
      <NButton @click="loadInstalledApps" :loading="installedLoading">
        <template #icon><NIcon><RefreshOutline /></NIcon></template>刷新
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
      <NEmpty description="暂无已安装应用，去应用商店看看吧">
        <NButton type="primary" size="small" @click="$emit('goStore')">
          浏览应用
        </NButton>
      </NEmpty>
    </div>

    <!-- 容器日志 -->
    <NModal v-model:show="logModal" preset="card" :title="`应用日志 - ${logAppName}`" :style="{ width: 'min(860px, 92vw)' }">
      <NSpin :show="logLoading">
        <pre class="app-log-viewer">{{ logContent }}</pre>
      </NSpin>
    </NModal>

    <!-- 卸载确认 -->
    <ConfirmDialog
      v-model:show="uninstallConfirm"
      title="卸载应用"
      :content="`确定要卸载「${targetInstalledApp?.name}」吗？容器及相关数据将被删除，此操作不可恢复。`"
      type="error"
      confirm-text="卸载"
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
  padding: 12px;
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
