<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import {
  NButton, NIcon, NForm, NFormItem, NInput, NSpace, NCard, NSpin, NSkeleton
} from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import { AddOutline, RefreshOutline, PlayOutline, PauseOutline, TrashOutline } from '@vicons/ionicons5'
import { PageCard, ModalForm, StatusTag, ConfirmDialog } from '@/components'
import type { StatusType } from '@/components'
import {
  listNodes, addNode, deleteNode, refreshNode, enableNode, disableNode,
  type NodeItem
} from '@/api/features17'
import { useI18n } from 'vue-i18n'

const message = useMessage()
const { t } = useI18n()

function fmtTime(v: string | null): string {
  if (!v) return '—'
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return v
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function relativeTime(v: string | null): string {
  if (!v) return t('nodes.timeNever')
  const tk = new Date(v).getTime()
  if (Number.isNaN(tk)) return v
  const diff = Date.now() - tk
  if (diff < 60_000) return t('nodes.timeJustNow')
  if (diff < 3_600_000) return t('nodes.minutesAgo', { n: Math.floor(diff / 60_000) })
  if (diff < 86_400_000) return t('nodes.hoursAgo', { n: Math.floor(diff / 3_600_000) })
  return fmtTime(v)
}

function nodeStatus(s: NodeItem['last_status']): { status: StatusType; text: string } {
  if (s === 'online') return { status: 'running', text: t('nodes.statusOnline') }
  if (s === 'offline') return { status: 'error', text: t('nodes.statusOffline') }
  return { status: 'stopped', text: t('nodes.statusUnknown') }
}

function usageColor(pct: number): string {
  const v = Number(pct) || 0
  if (v > 85) return 'var(--err)'
  if (v >= 70) return 'var(--warn)'
  return 'var(--ok)'
}

function usagePct(v: number): number {
  return Math.min(100, Math.max(0, Math.round(Number(v) || 0)))
}

/**
 * 进度条填充比例。
 * 用 transform: scaleX 而非 width —— width 动画每帧触发布局重排（主线程开销），
 * scaleX 只走合成器，动画更顺滑（审查报告 P1）。
 */
function usageScale(v: number): string {
  return `scaleX(${usagePct(v) / 100})`
}

/* ==================== 列表 ==================== */
const loading = ref(false)
const nodes = ref<NodeItem[]>([])
const actionId = ref<number | null>(null)

async function load() {
  loading.value = true
  try {
    const res = await listNodes()
    nodes.value = res.items || []
  } catch {
    nodes.value = []
  } finally {
    loading.value = false
  }
}

async function handleRefresh(id: number) {
  actionId.value = id
  try {
    await refreshNode(id)
    message.success(t('nodes.refreshed'))
    load()
  } catch (err: any) {
    message.error(err.message || t('nodes.refreshFailed'))
  } finally {
    actionId.value = null
  }
}

async function handleToggle(row: NodeItem) {
  actionId.value = row.id
  try {
    if (row.enabled) {
      await disableNode(row.id)
      message.success(t('nodes.disabled'))
    } else {
      await enableNode(row.id)
      message.success(t('nodes.enabled'))
    }
    load()
  } catch (err: any) {
    message.error(err.message || t('nodes.actionFailed'))
  } finally {
    actionId.value = null
  }
}

/* ==================== 添加 ==================== */
const showAdd = ref(false)
const addLoading = ref(false)
const addFormRef = ref()
const addForm = ref({ name: '', base_url: '', api_key: '' })

const addRules = computed(() => ({
  name: { required: true, message: t('nodes.ruleName'), trigger: 'blur' },
  base_url: { required: true, message: t('nodes.ruleBaseUrl'), trigger: 'blur' },
  api_key: { required: true, message: t('nodes.ruleApiKey'), trigger: 'blur' }
}))

function openAdd() {
  addForm.value = { name: '', base_url: '', api_key: '' }
  showAdd.value = true
}

async function handleAddSubmit() {
  try {
    await addFormRef.value?.validate()
  } catch {
    return
  }
  addLoading.value = true
  try {
    await addNode({ ...addForm.value })
    message.success(t('nodes.added'))
    showAdd.value = false
    load()
  } catch (err: any) {
    message.error(err.message || t('nodes.addFailed'))
  } finally {
    addLoading.value = false
  }
}

/* ==================== 删除 ==================== */
const showDelete = ref(false)
const deleteLoading = ref(false)
const deleteTarget = ref<NodeItem | null>(null)

function openDelete(row: NodeItem) {
  deleteTarget.value = row
  showDelete.value = true
}

async function handleDeleteConfirm() {
  if (!deleteTarget.value) return
  deleteLoading.value = true
  try {
    await deleteNode(deleteTarget.value.id)
    message.success(t('nodes.deleted'))
    showDelete.value = false
    load()
  } catch (err: any) {
    message.error(err.message || t('nodes.deleteFailed'))
  } finally {
    deleteLoading.value = false
  }
}

/* ==================== 生命周期 ==================== */
let timer: number | null = null
onMounted(() => {
  load()
  timer = window.setInterval(load, 60_000)
})
onBeforeUnmount(() => {
  if (timer !== null) window.clearInterval(timer)
})
</script>

<template>
  <PageCard :title="t('nodes.title')" :subtitle="t('nodes.subtitle')">
    <template #extra>
      <NButton type="primary" @click="openAdd">
        <template #icon><NIcon><AddOutline /></NIcon></template>
        {{ t('nodes.addNode') }}
      </NButton>
    </template>

    <p class="page-tip">{{ t('nodes.pageTip') }}</p>

    <NSpin :show="loading">
      <!-- 加载期骨架卡片网格：结构与真实节点卡一致，避免首屏空白（审查报告骨架屏维度） -->
      <div v-if="loading" class="node-grid" aria-hidden="true">
        <NCard v-for="i in 4" :key="`sk-${i}`" size="small" class="node-card">
          <NSkeleton height="17px" width="45%" :sharp="false" />
          <NSkeleton height="13px" width="70%" :sharp="false" style="margin-top: var(--space-2)" />
          <NSkeleton height="13px" width="55%" :sharp="false" style="margin-top: var(--space-1)" />
        </NCard>
      </div>
      <div v-else-if="nodes.length" class="node-grid">
        <NCard v-for="node in nodes" :key="node.id" size="small" class="node-card">
          <div class="node-head">
            <span class="node-name">{{ node.name }}</span>
            <StatusTag v-bind="nodeStatus(node.last_status)" />
          </div>
          <div class="node-meta">
            <span>{{ node.hostname || '—' }}</span>
            <span class="faint">v{{ node.version || '—' }}</span>
          </div>
          <div class="node-url mono">{{ node.base_url }}</div>

          <div class="usage-list">
            <div class="usage-row">
              <span class="usage-label">CPU</span>
              <div class="usage-bar"><div class="usage-fill" :style="{ transform: usageScale(node.cpu_percent), background: usageColor(node.cpu_percent) }" /></div>
              <span class="usage-val mono">{{ usagePct(node.cpu_percent) }}%</span>
            </div>
            <div class="usage-row">
              <span class="usage-label">{{ t('nodes.labelMem') }}</span>
              <div class="usage-bar"><div class="usage-fill" :style="{ transform: usageScale(node.mem_percent), background: usageColor(node.mem_percent) }" /></div>
              <span class="usage-val mono">{{ usagePct(node.mem_percent) }}%</span>
            </div>
            <div class="usage-row">
              <span class="usage-label">{{ t('nodes.labelDisk') }}</span>
              <div class="usage-bar"><div class="usage-fill" :style="{ transform: usageScale(node.disk_percent), background: usageColor(node.disk_percent) }" /></div>
              <span class="usage-val mono">{{ usagePct(node.disk_percent) }}%</span>
            </div>
          </div>

          <div class="node-foot">
            <span class="faint">{{ t('nodes.lastHeartbeat', { time: relativeTime(node.last_seen_at) }) }}</span>
          </div>
          <div v-if="node.last_error" class="node-error">{{ node.last_error }}</div>

          <NSpace size="small" class="node-actions">
            <NButton size="tiny" quaternary :loading="actionId === node.id" @click="handleRefresh(node.id)">
              <template #icon><NIcon><RefreshOutline /></NIcon></template>
              {{ t('nodes.btnRefresh') }}
            </NButton>
            <NButton size="tiny" quaternary :loading="actionId === node.id" @click="handleToggle(node)">
              <template #icon><NIcon><component :is="node.enabled ? PauseOutline : PlayOutline" /></NIcon></template>
              {{ node.enabled ? t('nodes.btnDisable') : t('nodes.btnEnable') }}
            </NButton>
            <NButton size="tiny" quaternary type="error" @click="openDelete(node)">
              <template #icon><NIcon><TrashOutline /></NIcon></template>
              {{ t('nodes.btnDelete') }}
            </NButton>
          </NSpace>
        </NCard>
      </div>

      <div v-else class="empty-state">
        <p>{{ loading ? t('nodes.loading') : t('nodes.empty') }}</p>
        <span>{{ t('nodes.emptyHint') }}</span>
      </div>
    </NSpin>
  </PageCard>

  <!-- 添加节点 -->
  <ModalForm
    :show="showAdd"
    :title="t('nodes.addNode')"
    :loading="addLoading"
    :width="520"
    @update:show="(v: boolean) => (showAdd = v)"
    @submit="handleAddSubmit"
  >
    <NForm ref="addFormRef" :model="addForm" :rules="addRules" label-placement="top">
      <NFormItem :label="t('nodes.formNameLabel')" path="name">
        <NInput v-model:value="addForm.name" :placeholder="t('nodes.formNamePlaceholder')" clearable />
      </NFormItem>
      <NFormItem :label="t('nodes.formUrlLabel')" path="base_url">
        <NInput v-model:value="addForm.base_url" placeholder="https://192.168.1.20:8888" clearable />
      </NFormItem>
      <NFormItem label="API Key" path="api_key">
        <NInput v-model:value="addForm.api_key" type="password" show-password-on="click" :placeholder="t('nodes.formApiKeyPlaceholder')" />
      </NFormItem>
    </NForm>
  </ModalForm>

  <!-- 删除确认 -->
  <ConfirmDialog
    :show="showDelete"
    :title="t('nodes.deleteTitle')"
    :content="t('nodes.deleteContent', { name: deleteTarget?.name || '' })"
    :confirm-text="t('nodes.deleteConfirm')"
    type="error"
    :loading="deleteLoading"
    @update:show="(v: boolean) => (showDelete = v)"
    @confirm="handleDeleteConfirm"
  />
</template>

<style scoped>
.page-tip {
  color: var(--text-quaternary);
  font-size: var(--fs-caption);
  margin: 0 0 var(--space-4);
}
.node-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--space-4);
}
.node-card {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}
.node-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}
.node-name {
  font-weight: 600;
  color: var(--ink);
}
.node-meta {
  display: flex;
  justify-content: space-between;
  font-size: var(--fs-caption);
  color: var(--sub);
}
.node-url {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
  word-break: break-all;
}
.mono {
  font-family: 'JetBrains Mono', 'Cascadia Code', Consolas, monospace;
}
.usage-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  margin-top: var(--space-1);
}
.usage-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}
.usage-label {
  width: 36px;
  flex-shrink: 0;
  font-size: var(--fs-caption);
  color: var(--sub);
}
.usage-bar {
  flex: 1;
  height: 6px;
  border-radius: var(--radius-sm);
  background: var(--surface2);
  overflow: hidden;
}
.usage-fill {
  height: 100%;
  width: 100%;
  border-radius: var(--radius-sm);
  transform-origin: left center;
  transition: transform var(--t-normal) var(--ease);
  will-change: transform;
}
.usage-val {
  width: 42px;
  flex-shrink: 0;
  text-align: right;
  font-size: var(--fs-caption);
  color: var(--sub);
}
.node-foot {
  font-size: var(--fs-caption);
}
.node-error {
  color: var(--err);
  font-size: var(--fs-caption);
  word-break: break-all;
}
.node-actions {
  margin-top: var(--space-1);
}
.faint {
  color: var(--text-quaternary);
  font-size: var(--fs-caption);
}
.empty-state {
  padding: var(--space-8) 0;
  text-align: center;
}
.empty-state p {
  color: var(--sub);
  font-size: var(--fs-body);
  margin: 0 0 var(--space-2);
}
.empty-state span {
  color: var(--text-quaternary);
  font-size: var(--fs-caption);
}
</style>
