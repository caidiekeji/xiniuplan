<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, h, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from '@/composables/useMessage'
import { NButton, NSpace, NIcon, NSelect, NInput, NTag, NEmpty, NSpin, NAlert, NSwitch, NDrawer, NDrawerContent, NRadioGroup, NRadio, NProgress, NDropdown } from 'naive-ui'
import type { DataTableColumns, DropdownOption } from 'naive-ui'
import { PlayOutline, StopOutline, RefreshOutline, DocumentTextOutline, CopyOutline, ImagesOutline, EllipsisHorizontalOutline, TrashOutline, DownloadOutline } from '@vicons/ionicons5'
import {
  PageCard, DataTable, StatusTag, ConfirmDialog, ContentModal, CodeEditor
} from '@/components'
import type { StatusType } from '@/components'
import * as containerApi from '@/api/container'
import { formatDate } from '@/utils/format'
import { subscribeSSE } from '@/utils/sse'
import { useQueryState } from '@/composables/useQueryState'
import type { Container, ContainerStatus, Image, RegistryOption, PullProgress } from '@/api/container'

import DockerResources from './DockerResources.vue'
import Compose from './Compose.vue'
import ExecTerminal from './ExecTerminal.vue'

const message = useMessage()
const { t, te } = useI18n()

/* ==================== h() 辅助函数 ==================== */

function monoEl(text: any) {
  return h('span', { class: 'mono' }, String(text ?? '-'))
}

/** 单行截断文本：超宽以省略号收尾并用 title 提示完整值，防止溢出压到相邻列（fixed 表格布局下 td 不会自动裁剪） */
function clipEl(text: any, mono = true) {
  const s = String(text ?? '-')
  return h('div', { class: mono ? 'cell-clip mono' : 'cell-clip', title: s }, s)
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

function statusText(state: ContainerStatus): string {
  // state 是归一化枚举（key 必存在）；te() 防御后端新增枚举时 key 泄漏给用户
  const key = 'container.status.' + state
  return te(key) ? t(key) : state
}

/* ==================== 筛选与搜索 ==================== */

// 筛选 / 搜索 / 分页 URL 化（D19）：刷新与深链可还原现场
const statusFilter = useQueryState<string>('status', 'all')
const searchKeyword = useQueryState('kw', '')

const statusOptions = computed(() => [
  { label: t('container.filter.all'), value: 'all' },
  { label: t('container.status.running'), value: 'running' },
  { label: t('container.status.stopped'), value: 'stopped' }
])

/* ==================== 容器列表 ==================== */

const containers = ref<Container[]>([])
const containersLoading = ref(false)
const dockerError = ref<string | null>(null)
let pollTimer: ReturnType<typeof setInterval> | null = null

const filteredContainers = computed(() => {
  let list = containers.value
  if (statusFilter.value === 'running') {
    list = list.filter(c => c.state === 'running')
  } else if (statusFilter.value === 'stopped') {
    list = list.filter(c => c.state !== 'running')
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
    dockerError.value = e.message || t('container.messages.dockerUnavailable')
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
    message.success(t('container.messages.started', { name: row.name }))
    loadContainers(true)
  } catch (e: any) {
    message.error(e.message || t('container.messages.startFailed'))
  } finally {
    actionLoading.value[row.id] = false
  }
}

async function handleStop(row: Container) {
  actionLoading.value[row.id] = true
  try {
    await containerApi.stopContainer(row.id)
    message.success(t('container.messages.stopped', { name: row.name }))
    loadContainers(true)
  } catch (e: any) {
    message.error(e.message || t('container.messages.stopFailed'))
  } finally {
    actionLoading.value[row.id] = false
  }
}

async function handleRestart(row: Container) {
  actionLoading.value[row.id] = true
  try {
    await containerApi.restartContainer(row.id)
    message.success(t('container.messages.restarted', { name: row.name }))
    loadContainers(true)
  } catch (e: any) {
    message.error(e.message || t('container.messages.restartFailed'))
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
    message.success(t('container.messages.deleted', { name: targetContainer.value.name }))
    deleteConfirm.value = false
    loadContainers()
  } catch (e: any) {
    message.error(e.message || t('container.messages.deleteFailed'))
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
// SSE 连接的取消函数（跟随模式下保持单条长连接）
let logUnsubscribe: (() => void) | null = null

const tailOptions = computed(() => [
  { label: t('container.logs.tailLines', { n: 100 }), value: 100 },
  { label: t('container.logs.tailLines', { n: 500 }), value: 500 },
  { label: t('container.logs.tailLines', { n: 1000 }), value: 1000 }
])

async function openLogs(row: Container) {
  logContainer.value = row
  logContent.value = ''
  logDrawerVisible.value = true
  await loadLogs()
  // 跟随开关已打开时，需要为新容器重建 SSE 连接
  if (logFollow.value) {
    startLogFollow()
  }
}

/* ==================== 镜像拉取 ==================== */

const pullModalVisible = ref(false)
const pullLoading = ref(false)
const pullRegistries = ref<RegistryOption[]>([])
const pullHasPrivate = ref(false)
const pullPrivateUser = ref('')
const pullForm = reactive({
  image: '',
  mirror: '',
  username: '',
  password: ''
})
const pullProgress = ref<PullProgress[]>([])
const pullOverall = ref(0)
const pullStatusText = ref('')
let pullUnsubscribe: (() => void) | null = null

/** 打开拉取弹窗：重置表单并加载可用仓库列表 */
async function openPull() {
  pullModalVisible.value = true
  pullForm.image = ''
  pullForm.mirror = ''
  pullForm.username = ''
  pullForm.password = ''
  pullProgress.value = []
  pullOverall.value = 0
  pullStatusText.value = ''
  try {
    const res = await containerApi.getRegistries()
    pullRegistries.value = res.data.registries || []
    pullHasPrivate.value = !!res.data.has_private
    pullPrivateUser.value = res.data.private_user || ''
    // 默认直连 Docker Hub：私有仓库镜像自带 registry，若默认选中加速源反而会被
    // 后端拒绝改写（P1-3）。用户需加速时再手动从下拉选择即可。
    pullForm.mirror = ''
    pullForm.username = ''
  } catch {
    pullRegistries.value = []
  }
}

/** 开始拉取 */
async function handlePull() {
  const image = pullForm.image.trim()
  if (!image) {
    message.warning(t('container.pull.imageRequired'))
    return
  }
  if (pullLoading.value) return

  pullLoading.value = true
  pullProgress.value = []
  pullOverall.value = 0
  pullStatusText.value = t('container.pull.pulling')

  // 凭据仅在选择私有仓库（@private）时随请求发送，避免把私有仓库账号误发给加速源
  const isPrivate = pullForm.mirror === '@private'
  pullUnsubscribe = containerApi.pullImage(
    {
      image,
      mirror: pullForm.mirror,
      username: isPrivate ? pullForm.username || undefined : undefined,
      password: isPrivate ? pullForm.password || undefined : undefined
    },
    (p) => {
      if (p.error) {
        pullStatusText.value = p.error
        message.error(p.error)
        pullLoading.value = false
        return
      }
      if (p.status === 'pulling') return // 心跳帧
      if (p.message || p.status === 'Pull complete') {
        pullStatusText.value = p.message || t('container.pull.complete')
        pullOverall.value = 100
        pullLoading.value = false
        message.success(p.message || t('container.pull.complete'))
        loadImages()
        return
      }
      if (p.status) pullStatusText.value = p.status
      // 汇总总进度：按各层百分比的最大值近似（各层并行下载，没有全局真值）
      if (p.percent !== undefined && p.percent >= 0) {
        const idx = pullProgress.value.findIndex((x) => x.layer_id === p.layer_id)
        if (idx >= 0) pullProgress.value[idx] = p
        else pullProgress.value.push(p)
        pullProgress.value = [...pullProgress.value]
        const max = Math.max(...pullProgress.value.map((x) => x.percent ?? 0))
        pullOverall.value = Math.min(100, Math.round(max))
      }
    },
    (e) => {
      pullStatusText.value = e.message
      message.error(e.message)
      pullLoading.value = false
    }
  )
}

/** 取消拉取 */
function cancelPull() {
  if (pullUnsubscribe) {
    pullUnsubscribe()
    pullUnsubscribe = null
  }
  pullLoading.value = false
  pullStatusText.value = t('container.pull.cancelled')
}

/** 关闭弹窗时确保取消未完成的拉取 */
function closePull() {
  if (pullLoading.value) cancelPull()
  pullModalVisible.value = false
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
    message.error(e.message || t('container.messages.statsFailed'))
    statsModal.value = false
  } finally {
    statsLoading.value = false
  }
}

async function loadLogs() {
  if (!logContainer.value) return
  logLoading.value = true
  try {
    // 跟随模式下走 SSE 流，不能用 axios（响应永不结束会挂起）
    if (logFollow.value) {
      logContent.value = ''
      return
    }
    const res = await containerApi.getContainerLogs(logContainer.value.id, logTail.value, false)
    logContent.value = res.data.content || t('container.logs.empty')
  } catch (e: any) {
    logContent.value = t('container.logs.loadFailed', { msg: e.message || t('container.unknownError') })
  } finally {
    logLoading.value = false
  }
}

/**
 * 启动日志跟随。
 *
 * 早期实现用 setInterval 每 3 秒用 axios 请求一次 follow=true —— SSE 响应
 * 永不结束，每个请求都会一直挂起直到超时，请求越堆越多，页面越来越卡。
 * 现在改为建立单条 SSE 长连接，服务端逐行推送。
 */
function startLogFollow() {
  stopLogFollow()
  if (!logContainer.value) return

  const containerId = logContainer.value.id
  const lines: string[] = []

  logUnsubscribe = subscribeSSE(
    `/containers/${encodeURIComponent(containerId)}/logs?tail=${logTail.value}&follow=true`,
    (line) => {
      lines.push(line)
      // 只保留最近 2000 行，避免长时间跟随后内存无限增长
      if (lines.length > 2000) lines.splice(0, lines.length - 2000)
      logContent.value = lines.join('\n')
    },
    () => {
      // 连接中断：结束跟随状态，避免界面显示「跟随中」但实际没在跟
      logFollow.value = false
      stopLogFollow()
    }
  )
}

function stopLogFollow() {
  if (logUnsubscribe) {
    logUnsubscribe()
    logUnsubscribe = null
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
    message.success(t('container.messages.copySuccess'))
  } catch {
    message.error(t('container.messages.copyFailed'))
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
    message.error(e.message || t('container.messages.loadImagesFailed'))
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

/* --- 镜像删除确认 --- */

const imageDeleteConfirm = ref(false)
const imageDeleteLoading = ref(false)
const targetImage = ref<Image | null>(null)

/** 镜像展示名：优先取第一个 tag，悬空镜像回退到短 ID */
function imageDisplayName(img: Image | null): string {
  if (!img) return ''
  return img.tags?.[0] || shortId(img.id)
}

function confirmImageDelete(row: Image) {
  targetImage.value = row
  imageDeleteConfirm.value = true
}

async function doImageDelete() {
  if (!targetImage.value) return
  imageDeleteLoading.value = true
  try {
    await containerApi.removeImage(targetImage.value.id, true)
    message.success(t('container.messages.imageDeleted', { name: imageDisplayName(targetImage.value) }))
    imageDeleteConfirm.value = false
    loadImages()
  } catch (e: any) {
    message.error(e.message || t('container.messages.imageDeleteFailed'))
  } finally {
    imageDeleteLoading.value = false
  }
}

const imageColumns = computed<DataTableColumns>(() => [
  {
    // 后端镜像 DTO 字段是 tags（backend/internal/dto/container.go），
    // 此前误写 repo_tags，导致镜像名称列恒为 <none>
    title: t('container.imageColumns.image'), key: 'tags',
    render: (row: any) => monoEl(row.tags?.[0] || '<none>')
  },
  { title: t('container.imageColumns.id'), key: 'id', width: 160, render: (row: any) => monoEl(shortId(row.id)) },
  { title: t('container.imageColumns.size'), key: 'size', width: 120, render: (row: any) => monoEl(formatSize(row.size)) },
  {
    title: t('container.imageColumns.createdAt'), key: 'created', width: 150,
    // 与容器表同源契约：镜像 created 也是 unix 秒，此前误写 created_at 一直显示 "-"
    render: (row: any) => row.created
      ? clipEl(formatDate(row.created * 1000), false)
      : h('span', { class: 'faint-text' }, '-')
  },
  {
    title: t('container.columns.actions'), key: 'actions', width: 90,
    render: (row: any) => btnGroup([
      iconBtn({
        text: t('container.actions.delete'), icon: TrashOutline, type: 'error',
        disabled: imageDeleteLoading.value && targetImage.value?.id === row.id,
        onClick: () => confirmImageDelete(row)
      })
    ])
  }
])

/* ==================== 表格列定义 ==================== */

const containerColumns = computed<DataTableColumns>(() => [
  {
    title: t('container.columns.name'), key: 'name', width: 200,
    render: (row: any) => h('div', { class: 'cell-name', title: row.name }, [
      h('div', { class: 'cell-name__main cell-clip' }, row.name),
      h('div', { class: 'cell-name__sub mono cell-clip' }, shortId(row.id))
    ])
  },
  {
    title: t('container.columns.image'), key: 'image', width: 180,
    render: (row: any) => clipEl(row.image)
  },
  {
    title: t('container.columns.status'), key: 'status', width: 110,
    render: (row: any) => h(StatusTag, {
      // state 是归一化枚举；status 是 Docker 原始描述（"Exited (0) 39 hours ago"），传给 key 会泄漏原文并撑爆单元格
      status: mapStatus(row.state as ContainerStatus),
      text: statusText(row.state as ContainerStatus)
    })
  },
  {
    title: t('container.columns.ports'), key: 'ports', width: 170,
    render: (row: any) => {
      // 后端契约：ports 为逗号分隔字符串（如 "0.0.0.0:8080->80/tcp"），非数组
      const list = String(row.ports || '').split(',').map((s) => s.trim()).filter(Boolean)
      if (list.length === 0) return h('span', { class: 'faint-text' }, '-')
      return h(NSpace, { size: 4, wrapItem: true, wrap: true }, {
        default: () => list.map((p, i) =>
          h(NTag, {
            key: i,
            size: 'small',
            bordered: false,
            type: 'info' as any,
            class: 'port-tag',
            title: p
          }, {
            default: () => p
          })
        )
      })
    }
  },
  {
    title: t('container.columns.createdAt'), key: 'created', width: 140,
    // 契约字段是 created（unix 秒）；此前误写 created_at，该列一直显示 "-"
    render: (row: any) => row.created
      ? clipEl(formatDate(row.created * 1000), false)
      : h('span', { class: 'faint-text' }, '-')
  },
  {
    title: t('container.columns.actions'), key: 'actions', width: 120,
    render: (row: any) => {
      const isRunning = row.state === 'running'
      const loading = isActionLoading(row.id)
      // 主位只留最高频的一个动作，其余收进 ⋯ 下拉（与文件管理页同款范式），避免 6 按钮换行挤压
      const primary = isRunning
        ? iconBtn({
            text: t('container.actions.stop'), icon: StopOutline, type: 'warning',
            disabled: loading,
            onClick: () => handleStop(row)
          })
        : iconBtn({
            text: t('container.actions.start'), icon: PlayOutline, type: 'success',
            disabled: loading,
            onClick: () => handleStart(row)
          })
      const menu: DropdownOption[] = [
        { key: 'restart', label: t('container.actions.restart') },
        { key: 'logs', label: t('container.actions.logs') },
        ...(isRunning ? [
          { key: 'terminal', label: t('container.actions.terminal') },
          { key: 'stats', label: t('container.actions.stats') }
        ] : []),
        { key: 'delete', label: () => h('span', { class: 'menu-danger' }, t('container.actions.delete')) }
      ]
      const more = h(NDropdown, {
        options: menu,
        trigger: 'click',
        disabled: loading,
        onSelect: (key: string | number) => handleRowAction(String(key), row)
      }, {
        default: () => h(NButton, {
          size: 'tiny', quaternary: true, disabled: loading,
          title: t('container.actions.more')
        }, {
          default: () => h(NIcon, { size: 14 }, { default: () => h(EllipsisHorizontalOutline) })
        })
      })
      return btnGroup([primary, more])
    }
  }
])

/** ⋯ 下拉菜单分发 */
function handleRowAction(key: string, row: Container) {
  if (key === 'restart') handleRestart(row)
  else if (key === 'logs') openLogs(row)
  else if (key === 'terminal') openExec(row)
  else if (key === 'stats') openStats(row)
  else if (key === 'delete') confirmDelete(row)
}

/* ==================== 分页 ==================== */

const containerPage = useQueryState('page', 1)
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

/* ==================== 容器终端（C-004） ==================== */
const execVisible = ref(false)
const execContainerId = ref('')
const execContainerName = ref('')

function openExec(row: any) {
  execContainerId.value = row.id
  execContainerName.value = row.name
  execVisible.value = true
}
</script>

<template>
  <PageCard :title="t('container.title')" :subtitle="t('container.subtitle')">
    <template #extra>
      <NButton @click="() => loadContainers()" :loading="containersLoading">
        <template #icon><NIcon><RefreshOutline /></NIcon></template>{{ t('container.refresh') }}
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
            :placeholder="t('container.searchPlaceholder')"
            clearable
            style="width: var(--input-3xl)"
          >
            <template #prefix>
              <NIcon :size="14"><DocumentTextOutline /></NIcon>
            </template>
          </NInput>
        </NSpace>
        <NSpace align="center">
          <span class="filter-label">{{ t('container.showImagesLabel') }}</span>
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
        <NEmpty :description="searchKeyword || statusFilter !== 'all' ? t('container.empty.noMatch') : t('container.empty.noContainer')">
          <NButton size="small" @click="() => loadContainers()">
            <template #icon><NIcon><RefreshOutline /></NIcon></template>{{ t('container.refresh') }}
          </NButton>
        </NEmpty>
      </div>

      <!-- 镜像列表（可折叠） -->
      <div v-if="showImages" class="images-section">
        <div class="images-section__header">
          <NIcon :size="16"><ImagesOutline /></NIcon>
          <span class="images-section__title">{{ t('container.imagesTitle') }}</span>
          <NButton size="tiny" @click="openPull">
            <template #icon><NIcon><DownloadOutline /></NIcon></template>{{ t('container.pull.button') }}
          </NButton>
          <NButton size="tiny" text @click="loadImages" :loading="imagesLoading">
            <template #icon><NIcon><RefreshOutline /></NIcon></template>{{ t('container.refresh') }}
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
          <NEmpty :description="t('container.empty.noImages')" />
        </div>
      </div>
    </template>

    <!-- 日志抽屉 -->
    <NDrawer
      :show="logDrawerVisible"
      :width="'var(--drawer-width)'"
      placement="right"
      @update:show="(v: boolean) => { if (!v) closeLogDrawer() }"
    >
      <NDrawerContent
        :title="logContainer ? t('container.logs.title', { name: logContainer.name }) : t('container.logs.titlePlain')"
        closable
      >
        <div class="log-toolbar">
          <NSpace align="center">
            <span class="log-label">{{ t('container.logs.tail') }}</span>
            <NRadioGroup v-model:value="logTail" @update:value="handleLogTailChange">
              <NSpace>
                <NRadio v-for="opt in tailOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </NRadio>
              </NSpace>
            </NRadioGroup>
          </NSpace>
          <NSpace align="center">
            <span class="log-label">{{ t('container.logs.follow') }}</span>
            <NSwitch :value="logFollow" @update:value="handleFollowChange" />
            <NButton size="small" @click="loadLogs" :loading="logLoading">
              <template #icon><NIcon><RefreshOutline /></NIcon></template>{{ t('container.refresh') }}
            </NButton>
            <NButton size="small" @click="copyLogs">
              <template #icon><NIcon><CopyOutline /></NIcon></template>{{ t('container.logs.copy') }}
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
      :title="t('container.deleteDialog.title')"
      :content="t('container.deleteDialog.content', { name: targetContainer?.name || '' })"
      type="error"
      :confirm-text="t('container.deleteDialog.confirm')"
      :loading="deleteLoading"
      @confirm="doDelete"
    />

    <!-- 镜像删除确认 -->
    <ConfirmDialog
      v-model:show="imageDeleteConfirm"
      :title="t('container.imageDeleteDialog.title')"
      :content="t('container.imageDeleteDialog.content', { name: imageDisplayName(targetImage) })"
      type="error"
      :confirm-text="t('container.imageDeleteDialog.confirm')"
      :loading="imageDeleteLoading"
      @confirm="doImageDelete"
    />

    <!-- 资源统计 -->
    <ContentModal v-model:show="statsModal" :title="t('container.stats.title', { name: statsContainer?.name || '' })" size="sm">
      <NSpin :show="statsLoading">
        <div v-if="statsData" class="stats-body">
          <div class="stats-row">
            <span class="stats-label">{{ t('container.stats.cpuUsage') }}</span>
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
            <span class="stats-label">{{ t('container.stats.memoryUsage') }}</span>
            <NProgress
              type="line"
              :percentage="Math.min(100, statsData.memory_percent)"
              :height="10"
              indicator-placement="inside"
              :color="statsData.memory_percent > 80 ? 'var(--err)' : 'var(--acc)'"
            />
            <span class="mono stats-value">{{ t('container.stats.memoryValue', { usage: statsData.memory_usage, limit: statsData.memory_limit, percent: statsData.memory_percent.toFixed(2) }) }}</span>
          </div>
        </div>
      </NSpin>
    </ContentModal>

    <!-- 拉取镜像弹窗：表单 + 流式进度 -->
    <ContentModal
      :show="pullModalVisible"
      :title="t('container.pull.title')"
      size="md"
      :mask-closable="false"
      @update:show="(v: boolean) => { if (!v) closePull() }"
    >
      <div class="pull-form">
        <div class="pull-form__row">
          <label class="pull-form__label">{{ t('container.pull.imageLabel') }}</label>
          <NInput
            v-model:value="pullForm.image"
            :placeholder="t('container.pull.imagePlaceholder')"
            :disabled="pullLoading"
            @keydown.enter="!pullLoading && handlePull()"
          />
          <div class="form-hint">{{ t('container.pull.imageHint') }}</div>
        </div>

        <div class="pull-form__row">
          <label class="pull-form__label">{{ t('container.pull.registryLabel') }}</label>
          <NSelect
            v-model:value="pullForm.mirror"
            :options="pullRegistries.map((r) => ({ label: r.name, value: r.id }))"
            :disabled="pullLoading"
            :placeholder="t('container.pull.registryPlaceholder')"
          />
          <div v-if="pullForm.mirror === '@private'" class="pull-private">
            <NInput
              v-model:value="pullForm.username"
              :placeholder="t('container.pull.usernamePlaceholder')"
              :disabled="pullLoading"
            />
            <NInput
              v-model:value="pullForm.password"
              type="password"
              show-password-on="click"
              :placeholder="t('container.pull.passwordPlaceholder')"
              :disabled="pullLoading"
            />
          </div>
        </div>

        <!-- 进度区：拉取中显示 -->
        <div v-if="pullLoading || pullOverall > 0" class="pull-progress">
          <NProgress
            type="line"
            :percentage="pullOverall"
            :height="10"
            indicator-placement="inside"
            :processing="pullLoading"
          />
          <div class="pull-progress__status">{{ pullStatusText }}</div>
          <div v-if="pullProgress.length > 0" class="pull-progress__layers">
            <div v-for="p in pullProgress.slice(0, 5)" :key="p.layer_id" class="pull-progress__layer">
              <span class="mono pull-progress__id">{{ p.layer_id?.slice(0, 10) }}</span>
              <span class="pull-progress__text">{{ p.status }}</span>
            </div>
            <div v-if="pullProgress.length > 5" class="pull-progress__more">
              {{ t('container.pull.moreLayers', { n: pullProgress.length - 5 }) }}
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <NSpace justify="end">
          <NButton v-if="pullLoading" @click="cancelPull">{{ t('container.pull.cancel') }}</NButton>
          <NButton v-else @click="closePull">{{ t('container.pull.close') }}</NButton>
          <NButton
            type="primary"
            :loading="pullLoading"
            :disabled="!pullForm.image.trim()"
            @click="handlePull"
          >
            {{ t('container.pull.start') }}
          </NButton>
        </NSpace>
      </template>
    </ContentModal>
  </PageCard>

  <DockerResources />
  <Compose />
  <ExecTerminal v-model:show="execVisible" :container-id="execContainerId" :name="execContainerName" />
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

/* 单行截断：fixed 表格布局下 td 不裁剪内容，长文本必须由自身收敛，否则会压到相邻列 */
.cell-clip {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.menu-danger {
  color: var(--err);
}

.port-tag {
  max-width: 100%;
}

.port-tag :deep(.n-tag__content) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cell-name__main {
  font-weight: 500;
  color: var(--ink);
  font-size: var(--fs-body);
}

.cell-name__sub {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
  margin-top: var(--space-0);
}

.faint-text {
  color: var(--text-quaternary);
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
  gap: var(--space-4);
  padding: var(--space-1) 0;
}

.stats-row {
  display: grid;
  grid-template-columns: 90px 1fr auto;
  align-items: center;
  gap: var(--space-3);
}

.stats-label {
  font-size: 13px;
  color: var(--text-quaternary);
}

.stats-value {
  font-size: 12px;
  min-width: 150px;
  text-align: right;
}

/* 拉取镜像弹窗 */
.pull-form__row {
  margin-bottom: var(--space-4);
}

.pull-form__label {
  display: block;
  font-size: var(--fs-caption);
  color: var(--sub);
  margin-bottom: var(--space-2);
}

.pull-private {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-2);
  margin-top: var(--space-2);
}

.pull-progress {
  margin-top: var(--space-5);
  padding: var(--space-3);
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  background: var(--surface2);
}

.pull-progress__status {
  font-size: var(--fs-caption);
  color: var(--sub);
  margin-top: var(--space-2);
  word-break: break-all;
}

.pull-progress__layers {
  margin-top: var(--space-2);
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.pull-progress__layer {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--fs-caption);
}

.pull-progress__id {
  color: var(--text-quaternary);
  flex-shrink: 0;
}

.pull-progress__text {
  color: var(--sub);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pull-progress__more {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
}
</style>
