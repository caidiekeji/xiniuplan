<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, h } from 'vue'
import { useMessage } from 'naive-ui'
import {
  NButton, NSpace, NIcon, NSelect, NInput, NTag, NEmpty, NSpin,
  NAlert, NSwitch, NDrawer, NDrawerContent, NRadioGroup, NRadio,
  NFormItem, NForm, NModal, NProgress
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {
  PlayOutline, StopOutline, RefreshOutline, TrashOutline,
  DocumentTextOutline, CubeOutline, CopyOutline, ImagesOutline, PulseOutline
} from '@vicons/ionicons5'
import {
  PageCard, DataTable, StatusTag, ConfirmDialog, CodeEditor
} from '@/components'
import type { StatusType } from '@/components'
import * as containerApi from '@/api/container'
import type { Container, ContainerStatus, Image, PortMapping } from '@/api/container'

const message = useMessage()

/* ==================== h() 辅助函数 ==================== */

function monoEl(text: any) {
  return h('span', { class: 'mono' }, String(text ?? '-'))
}

function iconBtn(opts: { text: string; icon: any; onClick: () => void; type?: string; disabled?: boolean }) {
  return h(NButton, {
    size: 'tiny',
    quaternary: true,
    type: (opts.type || 'default') as any,
    disabled: opts.disabled,
    onClick: opts.onClick
  }, {
    icon: () => h(NIcon, null, { default: () => h(opts.icon) }),
    default: () => opts.text
  })
}

function btnGroup(children: any[]) {
  return h(NSpace, { size: 4 }, { default: () => children })
}

/* ==================== 工具函数 ==================== */

function formatTime(s: string): string {
  if (!s) return '-'
  return s.replace('T', ' ').substring(0, 19)
}

function shortId(id: string): string {
  return id ? id.substring(0, 12) : '-'
}

function formatSize(bytes: number): string {
  if (!bytes) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
  return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
}

/** Docker 状态映射到 StatusTag */
function mapStatus(status: ContainerStatus): StatusType {
  switch (status) {
    case 'running':
      return 'running'
    case 'exited':
      return 'error'
    case 'paused':
    case 'created':
    case 'restarting':
      return 'warning'
    case 'stopped':
    default:
      return 'stopped'
  }
}

function statusText(status: ContainerStatus): string {
  const map: Record<ContainerStatus, string> = {
    running: '运行中',
    stopped: '已停止',
    exited: '已退出',
    created: '已创建',
    paused: '已暂停',
    restarting: '重启中'
  }
  return map[status] || status
}

/* ==================== 筛选与搜索 ==================== */

const statusFilter = ref<string>('all')
const searchKeyword = ref('')

const statusOptions = [
  { label: '全部', value: 'all' },
  { label: '运行中', value: 'running' },
  { label: '已停止', value: 'stopped' }
]

/* ==================== 容器列表 ==================== */

const containers = ref<Container[]>([])
const containersLoading = ref(false)
const dockerError = ref<string | null>(null)
let pollTimer: ReturnType<typeof setInterval> | null = null

const filteredContainers = computed(() => {
  let list = containers.value
  if (statusFilter.value === 'running') {
    list = list.filter(c => c.status === 'running')
  } else if (statusFilter.value === 'stopped') {
    list = list.filter(c => c.status !== 'running')
  }
  if (searchKeyword.value.trim()) {
    const kw = searchKeyword.value.trim().toLowerCase()
    list = list.filter(c =>
      c.name.toLowerCase().includes(kw) ||
      c.image.toLowerCase().includes(kw) ||
      c.id.toLowerCase().includes(kw)
    )
  }
  return list
})

async function loadContainers(silent = false) {
  if (!silent) containersLoading.value = true
  dockerError.value = null
  try {
    const res = await containerApi.getContainers(true)
    containers.value = res.data.items
  } catch (e: any) {
    dockerError.value = e.message || 'Docker 服务不可用，请检查 Docker 是否正常运行'
    containers.value = []
  } finally {
    if (!silent) containersLoading.value = false
  }
}

function startPolling() {
  stopPolling()
  pollTimer = setInterval(() => {
    loadContainers(true)
  }, 10000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

/* ==================== 容器操作 ==================== */

const actionLoading = ref<Record<string, boolean>>({})

function isActionLoading(id: string): boolean {
  return !!actionLoading.value[id]
}

async function handleStart(row: Container) {
  actionLoading.value[row.id] = true
  try {
    await containerApi.startContainer(row.id)
    message.success(`容器 ${row.name} 已启动`)
    loadContainers(true)
  } catch (e: any) {
    message.error(e.message || '启动失败')
  } finally {
    actionLoading.value[row.id] = false
  }
}

async function handleStop(row: Container) {
  actionLoading.value[row.id] = true
  try {
    await containerApi.stopContainer(row.id)
    message.success(`容器 ${row.name} 已停止`)
    loadContainers(true)
  } catch (e: any) {
    message.error(e.message || '停止失败')
  } finally {
    actionLoading.value[row.id] = false
  }
}

async function handleRestart(row: Container) {
  actionLoading.value[row.id] = true
  try {
    await containerApi.restartContainer(row.id)
    message.success(`容器 ${row.name} 已重启`)
    loadContainers(true)
  } catch (e: any) {
    message.error(e.message || '重启失败')
  } finally {
    actionLoading.value[row.id] = false
  }
}

/* --- 删除确认 --- */

const deleteConfirm = ref(false)
const deleteLoading = ref(false)
const targetContainer = ref<Container | null>(null)

function confirmDelete(row: Container) {
  targetContainer.value = row
  deleteConfirm.value = true
}

async function doDelete() {
  if (!targetContainer.value) return
  deleteLoading.value = true
  try {
    await containerApi.removeContainer(targetContainer.value.id, true)
    message.success(`容器 ${targetContainer.value.name} 已删除`)
    deleteConfirm.value = false
    loadContainers()
  } catch (e: any) {
    message.error(e.message || '删除失败')
  } finally {
    deleteLoading.value = false
  }
}

/* ==================== 日志对话框 ==================== */

const logDrawerVisible = ref(false)
const logContainer = ref<Container | null>(null)
const logContent = ref('')
const logLoading = ref(false)
const logTail = ref(500)
const logFollow = ref(false)
let logFollowTimer: ReturnType<typeof setInterval> | null = null

const tailOptions = [
  { label: '100 行', value: 100 },
  { label: '500 行', value: 500 },
  { label: '1000 行', value: 1000 }
]

async function openLogs(row: Container) {
  logContainer.value = row
  logContent.value = ''
  logDrawerVisible.value = true
  await loadLogs()
}

/* ==================== 容器资源统计 ==================== */

const statsModal = ref(false)
const statsLoading = ref(false)
const statsContainer = ref<Container | null>(null)
const statsData = ref<{ cpu_percent: number; memory_usage: string; memory_limit: string; memory_percent: number } | null>(null)

async function openStats(row: Container) {
  statsContainer.value = row
  statsData.value = null
  statsModal.value = true
  statsLoading.value = true
  try {
    const res = await containerApi.getContainerStats(row.id)
    statsData.value = res.data
  } catch (e: any) {
    message.error(e.message || '获取容器资源统计失败')
    statsModal.value = false
  } finally {
    statsLoading.value = false
  }
}

async function loadLogs() {
  if (!logContainer.value) return
  logLoading.value = true
  try {
    const res = await containerApi.getContainerLogs(logContainer.value.id, logTail.value, logFollow.value)
    logContent.value = res.data.content || '(无日志输出)'
  } catch (e: any) {
    logContent.value = `获取日志失败: ${e.message || '未知错误'}`
  } finally {
    logLoading.value = false
  }
}

function startLogFollow() {
  stopLogFollow()
  logFollowTimer = setInterval(() => {
    if (logDrawerVisible.value && logContainer.value) {
      containerApi.getContainerLogs(logContainer.value.id, logTail.value, true)
        .then(res => {
          logContent.value = res.data.content || '(无日志输出)'
        })
        .catch(() => {})
    }
  }, 3000)
}

function stopLogFollow() {
  if (logFollowTimer) {
    clearInterval(logFollowTimer)
    logFollowTimer = null
  }
}

function handleLogTailChange() {
  if (logDrawerVisible.value) {
    loadLogs()
  }
}

function handleFollowChange(val: boolean) {
  if (val) {
    startLogFollow()
  } else {
    stopLogFollow()
  }
}

function closeLogDrawer() {
  logDrawerVisible.value = false
  stopLogFollow()
  logFollow.value = false
  logContainer.value = null
}

async function copyLogs() {
  try {
    await navigator.clipboard.writeText(logContent.value)
    message.success('日志已复制到剪贴板')
  } catch {
    message.error('复制失败，请手动选择复制')
  }
}

/* ==================== 镜像列表 ==================== */

const showImages = ref(false)
const images = ref<Image[]>([])
const imagesLoading = ref(false)

async function loadImages() {
  imagesLoading.value = true
  try {
    const res = await containerApi.getImages()
    images.value = res.data
  } catch (e: any) {
    message.error(e.message || '加载镜像列表失败')
    images.value = []
  } finally {
    imagesLoading.value = false
  }
}

function toggleImages(val: boolean) {
  showImages.value = val
  if (val && images.value.length === 0) {
    loadImages()
  }
}

const imageColumns: DataTableColumns = [
  {
    title: '镜像', key: 'repo_tags',
    render: (row: any) => monoEl(row.repo_tags?.[0] || '<none>')
  },
  { title: 'ID', key: 'id', width: 160, render: (row: any) => monoEl(shortId(row.id)) },
  { title: '大小', key: 'size', width: 120, render: (row: any) => monoEl(formatSize(row.size)) },
  { title: '创建时间', key: 'created_at', width: 180, render: (row: any) => formatTime(row.created_at) }
]

/* ==================== 表格列定义 ==================== */

const containerColumns: DataTableColumns = [
  {
    title: '名称', key: 'name', width: 220,
    render: (row: any) => h('div', { class: 'cell-name' }, [
      h('div', { class: 'cell-name__main' }, row.name),
      h('div', { class: 'cell-name__sub mono' }, shortId(row.id))
    ])
  },
  {
    title: '镜像', key: 'image', width: 200,
    render: (row: any) => monoEl(row.image)
  },
  {
    title: '状态', key: 'status', width: 110,
    render: (row: any) => h(StatusTag, {
      status: mapStatus(row.status as ContainerStatus),
      text: statusText(row.status as ContainerStatus)
    })
  },
  {
    title: '端口映射', key: 'ports', width: 200,
    render: (row: any) => {
      const ports: PortMapping[] = row.ports || []
      if (ports.length === 0) return h('span', { class: 'faint-text' }, '-')
      return h(NSpace, { size: 4, wrapItem: true, wrap: true }, {
        default: () => ports.map((p, i) =>
          h(NTag, {
            key: i,
            size: 'small',
            bordered: false,
            type: 'info' as any
          }, {
            default: () => `${p.host_port}:${p.container_port}/${p.protocol}`
          })
        )
      })
    }
  },
  {
    title: '创建时间', key: 'created_at', width: 170,
    render: (row: any) => formatTime(row.created_at)
  },
  {
    title: '操作', key: 'actions', width: 280,
    render: (row: any) => {
      const isRunning = row.status === 'running'
      const loading = isActionLoading(row.id)
      const children: any[] = []
      if (!isRunning) {
        children.push(iconBtn({
          text: '启动', icon: PlayOutline, type: 'success',
          disabled: loading,
          onClick: () => handleStart(row)
        }))
      } else {
        children.push(iconBtn({
          text: '停止', icon: StopOutline, type: 'warning',
          disabled: loading,
          onClick: () => handleStop(row)
        }))
      }
      children.push(iconBtn({
        text: '重启', icon: RefreshOutline,
        disabled: loading,
        onClick: () => handleRestart(row)
      }))
      children.push(iconBtn({
        text: '日志', icon: DocumentTextOutline, type: 'primary',
        onClick: () => openLogs(row)
      }))
      if (isRunning) {
        children.push(iconBtn({
          text: '资源', icon: PulseOutline, type: 'info',
          onClick: () => openStats(row)
        }))
      }
      children.push(iconBtn({
        text: '删除', icon: TrashOutline, type: 'error',
        disabled: loading,
        onClick: () => confirmDelete(row)
      }))
      return btnGroup(children)
    }
  }
]

/* ==================== 分页 ==================== */

const containerPage = ref(1)
const containerPageSize = ref(20)

const containerPagination = computed(() => ({
  page: containerPage.value,
  pageSize: containerPageSize.value,
  itemCount: filteredContainers.value.length,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  'onUpdate:page': (p: number) => { containerPage.value = p },
  'onUpdate:page-size': (s: number) => { containerPageSize.value = s; containerPage.value = 1 }
}))

/* ==================== 生命周期 ==================== */

onMounted(() => {
  loadContainers()
  startPolling()
})

onBeforeUnmount(() => {
  stopPolling()
  stopLogFollow()
})
</script>

<template>
  <PageCard title="容器管理" subtitle="Docker 容器生命周期管理 · 实时日志 · 镜像浏览">
    <template #extra>
      <NButton @click="() => loadContainers()" :loading="containersLoading">
        <template #icon><NIcon><RefreshOutline /></NIcon></template>刷新
      </NButton>
    </template>

    <!-- Docker 不可用错误提示 -->
    <NAlert
      v-if="dockerError"
      type="error"
      :show-icon="true"
      class="docker-error"
    >
      {{ dockerError }}
    </NAlert>

    <template v-else>
      <!-- 筛选栏 -->
      <div class="filter-bar">
        <NSpace>
          <NSelect
            v-model:value="statusFilter"
            :options="statusOptions"
            style="width: var(--input-md)"
          />
          <NInput
            v-model:value="searchKeyword"
            placeholder="搜索容器名 / 镜像 / ID"
            clearable
            style="width: var(--input-3xl)"
          >
            <template #prefix>
              <NIcon :size="14"><DocumentTextOutline /></NIcon>
            </template>
          </NInput>
        </NSpace>
        <NSpace align="center">
          <span class="filter-label">显示镜像</span>
          <NSwitch :value="showImages" @update:value="toggleImages" />
        </NSpace>
      </div>

      <!-- 容器列表 -->
      <DataTable
        :columns="containerColumns"
        :data="filteredContainers"
        :loading="containersLoading"
        :pagination="containerPagination"
        :row-key="(row: any) => row.id"
        :virtual-scroll="true"
        :max-height="500"
      />

      <div v-if="filteredContainers.length === 0 && !containersLoading" class="empty-hint">
        <NEmpty :description="searchKeyword || statusFilter !== 'all' ? '没有匹配的容器' : '暂无容器'">
          <NButton size="small" @click="() => loadContainers()">
            <template #icon><NIcon><RefreshOutline /></NIcon></template>刷新
          </NButton>
        </NEmpty>
      </div>

      <!-- 镜像列表（可折叠） -->
      <div v-if="showImages" class="images-section">
        <div class="images-section__header">
          <NIcon :size="16"><ImagesOutline /></NIcon>
          <span class="images-section__title">镜像列表</span>
          <NButton size="tiny" text @click="loadImages" :loading="imagesLoading">
            <template #icon><NIcon><RefreshOutline /></NIcon></template>刷新
          </NButton>
        </div>
        <DataTable
          :columns="imageColumns"
          :data="images"
          :loading="imagesLoading"
          :pagination="false"
          size="small"
          :row-key="(row: any) => row.id"
        />
        <div v-if="images.length === 0 && !imagesLoading" class="empty-hint empty-hint--sm">
          <NEmpty description="暂无镜像" />
        </div>
      </div>
    </template>

    <!-- 日志抽屉 -->
    <NDrawer
      :show="logDrawerVisible"
      :width="720"
      placement="right"
      @update:show="(v: boolean) => { if (!v) closeLogDrawer() }"
    >
      <NDrawerContent
        :title="logContainer ? `容器日志 - ${logContainer.name}` : '容器日志'"
        closable
      >
        <div class="log-toolbar">
          <NSpace align="center">
            <span class="log-label">尾部行数</span>
            <NRadioGroup v-model:value="logTail" @update:value="handleLogTailChange">
              <NSpace>
                <NRadio v-for="opt in tailOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </NRadio>
              </NSpace>
            </NRadioGroup>
          </NSpace>
          <NSpace align="center">
            <span class="log-label">实时跟随</span>
            <NSwitch :value="logFollow" @update:value="handleFollowChange" />
            <NButton size="small" @click="loadLogs" :loading="logLoading">
              <template #icon><NIcon><RefreshOutline /></NIcon></template>刷新
            </NButton>
            <NButton size="small" @click="copyLogs">
              <template #icon><NIcon><CopyOutline /></NIcon></template>复制
            </NButton>
          </NSpace>
        </div>

        <div class="log-editor-wrap">
          <NSpin :show="logLoading">
            <CodeEditor
              v-model="logContent"
              language="shell"
              read-only
              height="calc(100vh - 220px)"
              :minimap="false"
            />
          </NSpin>
        </div>
      </NDrawerContent>
    </NDrawer>

    <!-- 删除确认 -->
    <ConfirmDialog
      v-model:show="deleteConfirm"
      title="删除容器"
      :content="`确定要删除容器「${targetContainer?.name}」吗？此操作不可恢复，容器内未持久化的数据将丢失。`"
      type="error"
      confirm-text="删除"
      :loading="deleteLoading"
      @confirm="doDelete"
    />

    <!-- 资源统计 -->
    <NModal v-model:show="statsModal" preset="card" :title="`资源统计 - ${statsContainer?.name || ''}`" :style="{ width: 'min(460px, 92vw)' }">
      <NSpin :show="statsLoading">
        <div v-if="statsData" class="stats-body">
          <div class="stats-row">
            <span class="stats-label">CPU 使用率</span>
            <NProgress
              type="line"
              :percentage="Math.min(100, statsData.cpu_percent)"
              :height="10"
              indicator-placement="inside"
              :color="statsData.cpu_percent > 80 ? 'var(--err)' : 'var(--acc3)'"
            />
            <span class="mono stats-value">{{ statsData.cpu_percent.toFixed(2) }}%</span>
          </div>
          <div class="stats-row">
            <span class="stats-label">内存使用</span>
            <NProgress
              type="line"
              :percentage="Math.min(100, statsData.memory_percent)"
              :height="10"
              indicator-placement="inside"
              :color="statsData.memory_percent > 80 ? 'var(--err)' : 'var(--acc)'"
            />
            <span class="mono stats-value">{{ statsData.memory_usage }} / {{ statsData.memory_limit }}（{{ statsData.memory_percent.toFixed(2) }}%）</span>
          </div>
        </div>
      </NSpin>
    </NModal>
  </PageCard>
</template>

<style scoped>
.filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-4);
  flex-wrap: wrap;
  gap: var(--space-3);
}

.filter-label {
  font-size: var(--fs-caption);
  color: var(--sub);
}

.docker-error {
  margin-bottom: var(--space-4);
}

.empty-hint {
  padding: var(--space-8) 0;
  display: flex;
  justify-content: center;
}

.empty-hint--sm {
  padding: var(--space-5) 0;
}

.cell-name__main {
  font-weight: 500;
  color: var(--ink);
  font-size: var(--fs-body);
}

.cell-name__sub {
  font-size: var(--fs-caption);
  color: var(--faint);
  margin-top: 2px;
}

.faint-text {
  color: var(--faint);
}

/* 镜像区域 */
.images-section {
  margin-top: var(--space-6);
  border-top: 1px solid var(--line);
  padding-top: var(--space-4);
}

.images-section__header {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-bottom: var(--space-3);
  color: var(--sub);
}

.images-section__title {
  font-size: var(--fs-body);
  font-weight: 600;
  color: var(--ink);
  margin-right: var(--space-2);
}

/* 日志抽屉 */
.log-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-3);
  flex-wrap: wrap;
  gap: var(--space-3);
}

.log-label {
  font-size: var(--fs-caption);
  color: var(--sub);
}

.log-editor-wrap {
  margin-top: var(--space-2);
}

:deep(.n-drawer .n-drawer-content-body) {
  padding: var(--space-4);
}

/* 移动端适配 */
@media (max-width: 768px) {
  .filter-bar {
    gap: var(--space-2);
  }
  .filter-bar :deep(.n-input),
  .filter-bar :deep(.n-select) {
    width: 100% !important;
  }
}

.stats-body {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 4px 0;
}

.stats-row {
  display: grid;
  grid-template-columns: 90px 1fr auto;
  align-items: center;
  gap: 12px;
}

.stats-label {
  font-size: 13px;
  color: var(--faint);
}

.stats-value {
  font-size: 12px;
  min-width: 150px;
  text-align: right;
}
</style>
