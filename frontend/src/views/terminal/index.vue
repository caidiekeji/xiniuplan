<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useMessage } from 'naive-ui'
import {
  NTabs, NTabPane, NButton, NSpace, NIcon,
  NForm, NFormItem, NInput, NInputNumber
} from 'naive-ui'
import type { FormRules } from 'naive-ui'
import type { StatusType } from '@/components'
import { PageCard, StatusTag } from '@/components'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'
import { useAuthStore } from '@/stores/auth'
import { useThemeTokens, hexToRgba } from '@/composables/useThemeTokens'
import {
  AddOutline, TrashOutline, CopyOutline, ClipboardOutline,
  ChevronDownOutline, ChevronUpOutline
} from '@vicons/ionicons5'

const message = useMessage()
const authStore = useAuthStore()
const { tokens, refresh: refreshTokens } = useThemeTokens()

/* ==================== 类型 ==================== */
type ConnStatus = 'connecting' | 'connected' | 'disconnected'

interface TabMeta {
  id: string
  title: string
  sessionId: string
  status: ConnStatus
  /** 收到服务端首条 output（SSH 真正建立）后置 true */
  connectedOnce?: boolean
  conn: ConnConfig
}

interface ConnConfig {
  host: string
  port: number
  username: string
  password: string
}

interface TermInstance {
  term: Terminal
  fit: FitAddon
  ws: WebSocket | null
  container: HTMLElement | null
  reconnectAttempts: number
  reconnectTimer: number | null
  resizeObserver: ResizeObserver | null
  disposed: boolean
}

/* ==================== 状态 ==================== */
const tabs = ref<TabMeta[]>([])
const activeId = ref('')
const quickBarOpen = ref(true)

/* SSH 连接配置表单 */
const connForm = ref<ConnConfig>({
  host: 'localhost',
  port: 22,
  username: '',
  password: ''
})
const connFormRef = ref()
const connRules: FormRules = {
  host: { required: true, message: '请输入主机地址', trigger: 'blur' },
  port: { required: true, type: 'number', min: 1, max: 65535, message: '端口范围 1-65535', trigger: 'blur' },
  username: { required: true, message: '请输入用户名', trigger: 'blur' }
}
const connecting = ref(false)

const instances = new Map<string, TermInstance>()

const activeTab = computed(() => tabs.value.find((t) => t.id === activeId.value) || null)
const activeStatusType = computed<StatusType>(() => {
  const s = activeTab.value?.status
  if (s === 'connected') return 'running'
  if (s === 'connecting') return 'warning'
  return 'error'
})
const activeStatusText = computed(() => {
  const s = activeTab.value?.status
  if (s === 'connected') return '已连接'
  if (s === 'connecting') return '连接中'
  return '已断开'
})

/* ==================== 工具 ==================== */
function genId() {
  return `t_${Date.now()}_${Math.random().toString(36).slice(2, 7)}`
}
function genSession() {
  return `sess_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`
}
function statusLedClass(s: ConnStatus) {
  if (s === 'connected') return 'led--ok'
  if (s === 'connecting') return 'led--warn'
  return 'led--err'
}

/* ==================== 终端创建 / 连接 ==================== */
function createTerminalOptions() {
  const t = tokens.value
  const bg = t.bg || '#0e1117'
  const ink = t.ink || '#e8ecf2'
  const acc = t.acc || '#1ba784'
  return {
    fontFamily: t.fontMono || '"JetBrains Mono", Consolas, "Courier New", monospace',
    fontSize: parseInt(t.termFontSize) || 14,
    lineHeight: 1.2,
    cursorBlink: true,
    cursorStyle: 'bar' as const,
    scrollback: 5000,
    convertEol: true,
    theme: {
      background: bg,
      foreground: ink,
      cursor: acc,
      cursorAccent: bg,
      selectionBackground: hexToRgba(acc, 0.3),
      black: t.surface || '#1a2030',
      red: t.err || '#e85d4d',
      green: t.ok || '#2dd4a8',
      yellow: t.warn || '#1ba784',
      blue: t.info || '#60a5fa',
      magenta: t.termMagenta || '#c084fc',
      cyan: t.termCyan || '#22d3ee',
      white: ink,
      brightBlack: t.faint || '#5c6678'
    }
  }
}

function initTerminal(id: string) {
  const container = document.getElementById(`term-${id}`)
  if (!container) return
  const tab = tabs.value.find((t) => t.id === id)
  if (!tab) return

  const term = new Terminal(createTerminalOptions())
  const fit = new FitAddon()
  term.loadAddon(fit)
  term.open(container)

  const inst: TermInstance = {
    term,
    fit,
    ws: null,
    container,
    reconnectAttempts: 0,
    reconnectTimer: null,
    resizeObserver: null,
    disposed: false
  }
  instances.set(id, inst)

  term.onData((data) => {
    send(id, { type: 'input', data })
  })

  // 自适应大小
  fit.fit()
  const ro = new ResizeObserver(() => {
    if (inst.disposed) return
    if (activeId.value === id) {
      try { fit.fit() } catch { /* noop */ }
      sendResize(id)
    }
  })
  ro.observe(container)
  inst.resizeObserver = ro

  connect(id)
  // 下一帧再次 fit，确保容器尺寸稳定
  requestAnimationFrame(() => {
    if (!inst.disposed) {
      try { fit.fit() } catch { /* noop */ }
      sendResize(id)
    }
  })
}

function connect(id: string) {
  const inst = instances.get(id)
  const tab = tabs.value.find((t) => t.id === id)
  if (!inst || !tab || inst.disposed) return

  tab.status = 'connecting'
  const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const c = tab.conn
  const apiBase = (window as any).__PANEL_BASE__ || ''
  // SSH 连接参数改为 WS 首条消息传输（不再进 URL，避免密码出现在访问日志）
  const url =
    `${proto}://${window.location.host}${apiBase}/api/v1/ws/terminal` +
    `?session_id=${encodeURIComponent(tab.sessionId)}` +
    `&cols=${inst.term.cols}` +
    `&rows=${inst.term.rows}` +
    `&token=${encodeURIComponent(authStore.token || '')}`

  const ws = new WebSocket(url)
  inst.ws = ws

  ws.onopen = () => {
    if (inst.disposed) { ws.close(); return }
    inst.reconnectAttempts = 0
    // 首条消息：SSH 连接参数。恢复会话时服务端会忽略此消息
    send(id, {
      type: 'connect',
      host: c.host,
      port: Number(c.port) || 22,
      username: c.username,
      password: c.password
    })
  }

  ws.onmessage = (e) => {
    if (inst.disposed) return
    try {
      const msg = JSON.parse(e.data)
      if (msg.type === 'output') {
        if (!tab.connectedOnce) {
          tab.connectedOnce = true
          tab.status = 'connected'
          inst.term.write('\r\n\x1b[32m[云枢终端 · 已连接]\x1b[0m\r\n')
        }
        inst.term.write(msg.data)
      } else if (msg.type === 'error') {
        // SSH 认证/连接失败：显示错误且不计入自动重连（相同参数重连必然再失败）
        inst.term.write(`\r\n\x1b[31m[错误] ${msg.data}\x1b[0m\r\n`)
        inst.reconnectAttempts = 3
      } else if (msg.type === 'close') {
        inst.term.write('\r\n\x1b[33m[服务端已关闭连接]\x1b[0m\r\n')
        scheduleReconnect(id)
      }
    } catch {
      inst.term.write(e.data)
    }
  }

  ws.onclose = () => {
    if (inst.disposed) return
    tab.status = 'disconnected'
    scheduleReconnect(id)
  }

  ws.onerror = () => {
    try { ws.close() } catch { /* noop */ }
  }
}

function scheduleReconnect(id: string) {
  const inst = instances.get(id)
  if (!inst || inst.disposed) return
  if (inst.reconnectAttempts >= 3) {
    inst.term.write('\r\n\x1b[31m[重连失败，已达到最大次数]\x1b[0m\r\n')
    return
  }
  const delay = Math.pow(2, inst.reconnectAttempts) * 1000
  inst.reconnectAttempts += 1
  inst.term.write(`\r\n\x1b[33m[${delay / 1000}s 后第 ${inst.reconnectAttempts} 次重连…]\x1b[0m\r\n`)
  inst.reconnectTimer = window.setTimeout(() => connect(id), delay)
}

function send(id: string, msg: Record<string, unknown>) {
  const inst = instances.get(id)
  if (!inst?.ws || inst.ws.readyState !== WebSocket.OPEN) return
  inst.ws.send(JSON.stringify(msg))
}

function sendResize(id: string) {
  const inst = instances.get(id)
  if (!inst) return
  send(id, { type: 'resize', cols: inst.term.cols, rows: inst.term.rows })
}

/* ==================== 标签管理 ==================== */
function createTab(conn?: ConnConfig) {
  const id = genId()
  const sessionId = genSession()
  const idx = tabs.value.length + 1
  const c: ConnConfig = conn || { ...connForm.value }
  tabs.value.push({
    id,
    title: `${c.username}@${c.host}`,
    sessionId,
    status: 'connecting',
    conn: c
  })
  activeId.value = id
  nextTick(() => initTerminal(id))
}

async function handleConnect() {
  try {
    await connFormRef.value?.validate()
  } catch {
    return
  }
  connecting.value = true
  try {
    createTab({ ...connForm.value })
  } finally {
    connecting.value = false
  }
}

function handleClose(name: string | number) {
  closeTab(String(name))
}

function closeTab(id: string) {
  const inst = instances.get(id)
  if (inst) {
    inst.disposed = true
    if (inst.reconnectTimer) clearTimeout(inst.reconnectTimer)
    if (inst.ws) {
      try { inst.ws.send(JSON.stringify({ type: 'close' })) } catch { /* noop */ }
      try { inst.ws.close() } catch { /* noop */ }
    }
    if (inst.resizeObserver) inst.resizeObserver.disconnect()
    inst.term.dispose()
    instances.delete(id)
  }
  tabs.value = tabs.value.filter((t) => t.id !== id)
  if (activeId.value === id) {
    activeId.value = tabs.value.length ? tabs.value[tabs.value.length - 1].id : ''
  }
}

function closeActiveTab() {
  if (activeId.value) closeTab(activeId.value)
}

/* ==================== 快捷命令 / 操作 ==================== */
const quickCommands = [
  { label: 'top', cmd: 'top\n' },
  { label: 'df -h', cmd: 'df -h\n' },
  { label: 'free -m', cmd: 'free -m\n' },
  { label: 'docker ps', cmd: 'docker ps\n' },
  { label: 'nginx -t', cmd: 'nginx -t\n' }
]

function runQuick(cmd: string) {
  if (!activeId.value) {
    message.warning('请先打开一个终端标签')
    return
  }
  send(activeId.value, { type: 'input', data: cmd })
}

function clearTerminal() {
  const inst = instances.get(activeId.value)
  if (inst) inst.term.clear()
}

async function copySelection() {
  const inst = instances.get(activeId.value)
  if (!inst) return
  const text = inst.term.getSelection()
  if (!text) {
    message.warning('没有选中的文本')
    return
  }
  try {
    await navigator.clipboard.writeText(text)
    message.success('已复制到剪贴板')
  } catch {
    message.error('复制失败')
  }
}

async function pasteFromClipboard() {
  if (!activeId.value) return
  try {
    const text = await navigator.clipboard.readText()
    send(activeId.value, { type: 'input', data: text })
  } catch {
    message.warning('无法读取剪贴板，请使用 Ctrl+Shift+V')
  }
}

function onKeydown(e: KeyboardEvent) {
  if (!activeId.value) return
  if (e.ctrlKey && e.shiftKey && e.key.toLowerCase() === 'c') {
    e.preventDefault()
    copySelection()
  } else if (e.ctrlKey && e.shiftKey && e.key.toLowerCase() === 'v') {
    e.preventDefault()
    pasteFromClipboard()
  }
}

/* ==================== 生命周期 ==================== */
watch(activeId, (id) => {
  if (!id) return
  nextTick(() => {
    const inst = instances.get(id)
    if (inst) {
      try { inst.fit.fit() } catch { /* noop */ }
      sendResize(id)
      inst.term.focus()
    }
  })
})

onMounted(() => {
  window.addEventListener('keydown', onKeydown)
  // 不再自动创建终端，用户通过连接表单发起连接
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
  for (const id of Array.from(instances.keys())) {
    closeTab(id)
  }
})
</script>

<template>
  <PageCard class="term-page" padding="0">
    <!-- 顶部栏 -->
    <div class="term-header">
      <div class="term-header__left">
        <span class="term-header__title">Web 终端</span>
        <StatusTag v-if="activeTab" :status="activeStatusType" :text="activeStatusText" />
      </div>
      <NSpace class="term-header__actions" size="small">
        <NButton size="small" quaternary @click="copySelection" title="复制 (Ctrl+Shift+C)">
          <template #icon><NIcon size="14"><CopyOutline /></NIcon></template>
          复制
        </NButton>
        <NButton size="small" quaternary @click="pasteFromClipboard" title="粘贴 (Ctrl+Shift+V)">
          <template #icon><NIcon size="14"><ClipboardOutline /></NIcon></template>
          粘贴
        </NButton>
        <NButton size="small" quaternary @click="clearTerminal" title="清屏">
          <template #icon><NIcon size="14"><TrashOutline /></NIcon></template>
          清屏
        </NButton>
        <NButton size="small" quaternary @click="closeActiveTab" title="关闭当前标签">
          关闭标签
        </NButton>
        <NButton size="small" @click="quickBarOpen = !quickBarOpen">
          <template #icon>
            <NIcon size="14">
              <component :is="quickBarOpen ? ChevronDownOutline : ChevronUpOutline" />
            </NIcon>
          </template>
          快捷命令
        </NButton>
      </NSpace>
    </div>

    <!-- SSH 连接配置表单 -->
    <div class="term-conn-form">
      <NForm
        ref="connFormRef"
        :model="connForm"
        :rules="connRules"
        label-placement="left"
        label-width="70"
        inline
        size="small"
      >
        <NFormItem label="主机" path="host">
          <NInput v-model:value="connForm.host" placeholder="localhost" style="width: 160px" />
        </NFormItem>
        <NFormItem label="端口" path="port">
          <NInputNumber v-model:value="connForm.port" :min="1" :max="65535" style="width: 100px" />
        </NFormItem>
        <NFormItem label="用户名" path="username">
          <NInput v-model:value="connForm.username" placeholder="root" style="width: 140px" />
        </NFormItem>
        <NFormItem label="密码" path="password">
          <NInput
            v-model:value="connForm.password"
            type="password"
            show-password-on="click"
            placeholder="密码"
            style="width: 160px"
            @keyup.enter="handleConnect"
          />
        </NFormItem>
        <NFormItem>
          <NButton type="primary" :loading="connecting" @click="handleConnect">
            连接
          </NButton>
        </NFormItem>
      </NForm>
    </div>

    <!-- 标签 + 终端区 -->
    <div class="term-body">
      <NTabs
        v-model:value="activeId"
        type="card"
        closable
        display-directive="show"
        class="term-tabs"
        @close="handleClose"
      >
        <NTabPane v-for="tab in tabs" :key="tab.id" :name="tab.id">
          <template #tab>
            <span class="tab-label">
              <span class="led" :class="statusLedClass(tab.status)" />
              <span class="tab-label__text">{{ tab.title }}</span>
            </span>
          </template>
          <div class="term-wrap" :id="`term-${tab.id}`" />
        </NTabPane>
        <template #suffix>
          <NButton size="small" quaternary circle @click="() => createTab()" title="新建终端标签">
            <template #icon><NIcon size="14"><AddOutline /></NIcon></template>
          </NButton>
        </template>
      </NTabs>
    </div>

    <!-- 快捷命令栏 -->
    <div v-if="quickBarOpen" class="term-quickbar">
      <span class="term-quickbar__label">常用命令</span>
      <NSpace size="small" wrap>
        <NButton
          v-for="c in quickCommands"
          :key="c.label"
          size="tiny"
          class="quick-cmd"
          @click="runQuick(c.cmd)"
        >
          {{ c.label }}
        </NButton>
      </NSpace>
    </div>
  </PageCard>
</template>

<style scoped>
.term-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 104px);
  background: var(--bg);
}

.term-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-3) var(--space-5);
  border-bottom: 1px solid var(--line);
  background: var(--surface);
  flex-wrap: wrap;
}

.term-header__left {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.term-header__title {
  font-size: var(--fs-title);
  font-weight: 600;
  color: var(--ink);
}

.term-header__actions {
  flex-shrink: 0;
}

.term-conn-form {
  padding: var(--space-3) var(--space-5);
  border-bottom: 1px solid var(--line);
  background: var(--surface);
}

.term-conn-form :deep(.n-form-item) {
  margin-bottom: 0;
}

.term-body {
  flex: 1;
  min-height: 0;
  display: flex;
  padding: var(--space-3);
  background: var(--bg);
}

.term-tabs {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.term-tabs :deep(.n-tabs-wrapper) {
  margin-bottom: var(--space-2);
}

.term-tabs :deep(.n-tabs-panels) {
  flex: 1;
  min-height: 0;
}

.term-tabs :deep(.n-tabs-panel) {
  height: 100%;
  padding: 0;
}

.term-wrap {
  width: 100%;
  height: 100%;
  background: var(--bg);
  padding: var(--space-2);
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.term-wrap :deep(.xterm) {
  height: 100% !important;
}

.term-wrap :deep(.xterm-viewport) {
  overflow-y: auto;
}

.tab-label {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
}

.tab-label__text {
  font-size: var(--fs-body);
}

.term-quickbar {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-3) var(--space-5);
  border-top: 1px solid var(--line);
  background: var(--surface);
  flex-wrap: wrap;
}

.term-quickbar__label {
  font-size: var(--fs-caption);
  color: var(--faint);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  flex-shrink: 0;
}

.quick-cmd {
  font-family: var(--font-mono);
  font-size: var(--fs-caption);
}

/* 指示灯（复用全局 .led，这里确保深色下可见） */
.led {
  width: var(--led-size);
  height: var(--led-size);
  border-radius: 50%;
  display: inline-block;
  flex-shrink: 0;
}

/* ========== 移动端适配 ========== */
@media (max-width: 768px) {
  .term-conn-form {
    padding: var(--space-3);
  }

  /* 连接表单改为垂直排列 */
  .term-conn-form :deep(.n-form) {
    flex-direction: column;
    align-items: stretch;
  }

  .term-conn-form :deep(.n-form-item) {
    margin-right: 0 !important;
    margin-bottom: var(--space-2) !important;
    width: 100%;
  }

  .term-conn-form :deep(.n-form-item-label) {
    width: 60px !important;
  }

  .term-conn-form :deep(.n-input),
  .term-conn-form :deep(.n-input-number) {
    width: 100% !important;
  }

  .term-conn-form :deep(.n-form-item:last-child) {
    margin-bottom: 0 !important;
  }

  .term-conn-form :deep(.n-form-item:last-child .n-button) {
    width: 100%;
  }

  /* 终端区域 */
  .term-body {
    padding: var(--space-2);
  }

  .term-wrap {
    padding: var(--space-1);
  }

  /* 快捷命令栏可横向滚动 */
  .term-quickbar {
    flex-wrap: nowrap;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
    padding: var(--space-2) var(--space-3);
  }

  .term-quickbar :deep(.n-space) {
    flex-wrap: nowrap !important;
  }
}

@media (max-width: 480px) {
  .term-conn-form {
    padding: var(--space-2);
  }

  .term-header {
    padding: var(--space-2) var(--space-3);
  }

  .term-header__title {
    font-size: var(--fs-title);
  }
}
</style>
