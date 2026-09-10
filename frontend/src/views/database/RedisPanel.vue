<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue'
import { useMessage } from 'naive-ui'
import {
  NButton, NSpace, NIcon, NTabs, NTabPane, NInput, NInputNumber,
  NSelect, NRadioGroup, NRadio, NEmpty, NSpin, NAlert, NTag,
  NProgress, NFormItem, NCheckbox, NModal
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {
  SearchOutline, RefreshOutline, AddOutline, TrashOutline,
  SaveOutline, TerminalOutline, HardwareChipOutline, KeyOutline,
  ChevronForwardOutline, ChevronBackOutline, CloseOutline, TimerOutline
} from '@vicons/ionicons5'
import {
  PageCard, DataTable, ModalForm, StatusTag, ConfirmDialog, CodeEditor, StatCard
} from '@/components'
import type { StatusType } from '@/components'
import * as redisApi from '@/api/redis'
import type {
  RedisKeyInfo, RedisKeyValue, RedisKeyType, RedisMemoryInfo, RedisBigKey
} from '@/api/redis'

const props = defineProps<{
  instanceId: number
}>()

const message = useMessage()

/* ==================== 辅助函数 ==================== */

function monoEl(text: any) {
  return h('span', { class: 'mono' }, String(text ?? '-'))
}

function iconBtn(opts: { text: string; icon: any; onClick: () => void; type?: string }) {
  return h(NButton, {
    size: 'tiny',
    quaternary: true,
    type: (opts.type || 'default') as any,
    onClick: opts.onClick
  }, {
    icon: () => h(NIcon, null, { default: () => h(opts.icon) }),
    default: () => opts.text
  })
}

function btnGroup(children: any[]) {
  return h(NSpace, { size: 4 }, { default: () => children })
}

const keyTypeStatusMap: Record<RedisKeyType, StatusType> = {
  string: 'running',
  hash: 'running',
  list: 'warning',
  set: 'pending',
  zset: 'error',
  none: 'stopped'
}

function formatBytes(bytes?: number): string {
  if (bytes === undefined || bytes === null) return '-'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`
}

function formatTTL(ttl: number): string {
  if (ttl === -1) return '永久'
  if (ttl === -2) return '不存在'
  if (ttl < 60) return `${ttl}s`
  if (ttl < 3600) return `${Math.floor(ttl / 60)}m ${ttl % 60}s`
  return `${Math.floor(ttl / 3600)}h ${Math.floor((ttl % 3600) / 60)}m`
}

/* ==================== 子 Tab ==================== */

const activeSubTab = ref('keys')

/* ==================== Tab 1: Key 浏览器 ==================== */

const searchPattern = ref('*')
const scanCursor = ref(0)
const scanCursorStack = ref<number[]>([])
const keyList = ref<RedisKeyInfo[]>([])
const keyListLoading = ref(false)
const hasMoreKeys = ref(false)
const totalKeys = ref(0)

const selectedKey = ref<RedisKeyValue | null>(null)
const selectedKeyLoading = ref(false)
const keyDetailDirty = ref(false)

// 编辑状态
const editStringValue = ref('')
const editTTL = ref<number | null>(null)
const editHashPairs = ref<Array<{ field: string; value: string }>>([])
const editListItems = ref<string[]>([])
const editSetMembers = ref<string[]>([])
const editZSetItems = ref<Array<{ member: string; score: number }>>([])
const newListItem = ref('')
const newSetMember = ref('')
const newZSetMember = ref('')
const newZSetScore = ref(0)

const nextCursor = ref(0)

async function loadKeys() {
  keyListLoading.value = true
  try {
    const res = await redisApi.scanKeys(props.instanceId, {
      pattern: searchPattern.value || '*',
      cursor: scanCursor.value,
      count: 50
    })
    keyList.value = res.data.keys
    nextCursor.value = res.data.cursor
    hasMoreKeys.value = res.data.cursor !== 0
    totalKeys.value = res.data.total ?? 0
  } catch (e: any) {
    message.error(e.message || '加载 Key 列表失败')
  } finally {
    keyListLoading.value = false
  }
}

function searchKeys() {
  scanCursor.value = 0
  scanCursorStack.value = []
  selectedKey.value = null
  loadKeys()
}

function nextPage() {
  if (!hasMoreKeys.value) return
  scanCursorStack.value.push(scanCursor.value)
  scanCursor.value = nextCursor.value
  loadKeys()
}

function prevPage() {
  if (scanCursorStack.value.length === 0) return
  scanCursor.value = scanCursorStack.value.pop()!
  loadKeys()
}

const keyColumns: DataTableColumns = [
  {
    title: 'Key', key: 'key',
    render: (row: any) => h(NButton, {
      text: true, size: 'small', onClick: () => selectKey(row.key)
    }, {
      default: () => h(NSpace, { size: 4 }, {
        default: () => [
          h(NIcon, { size: 14 }, { default: () => h(KeyOutline) }),
          monoEl(row.key)
        ]
      })
    })
  },
  {
    title: '类型', key: 'type', width: 100,
    render: (row: any) => h(StatusTag, {
      status: keyTypeStatusMap[row.type as RedisKeyType] || 'stopped',
      text: row.type?.toUpperCase()
    })
  },
  { title: 'TTL', key: 'ttl', width: 100, render: (row: any) => formatTTL(row.ttl) },
  { title: '大小', key: 'size', width: 100, render: (row: any) => formatBytes(row.size) }
]

async function selectKey(key: string) {
  selectedKeyLoading.value = true
  keyDetailDirty.value = false
  try {
    const res = await redisApi.getKeyValue(props.instanceId, key)
    selectedKey.value = res.data
    editTTL.value = res.data.ttl >= 0 ? res.data.ttl : null
    // 初始化编辑数据
    if (res.data.type === 'string') {
      editStringValue.value = String(res.data.value ?? '')
    } else if (res.data.type === 'hash') {
      const obj = res.data.value as Record<string, string>
      editHashPairs.value = Object.entries(obj || {}).map(([field, value]) => ({ field, value: String(value) }))
    } else if (res.data.type === 'list') {
      editListItems.value = (res.data.value as string[]) || []
    } else if (res.data.type === 'set') {
      editSetMembers.value = (res.data.value as string[]) || []
    } else if (res.data.type === 'zset') {
      editZSetItems.value = (res.data.value as Array<{ member: string; score: number }>) || []
    }
  } catch (e: any) {
    message.error(e.message || '加载 Key 详情失败')
  } finally {
    selectedKeyLoading.value = false
  }
}

function markDirty() {
  keyDetailDirty.value = true
}

async function saveKey() {
  if (!selectedKey.value) return
  let value: any
  const type = selectedKey.value.type
  if (type === 'string') {
    value = editStringValue.value
  } else if (type === 'hash') {
    value = {}
    editHashPairs.value.forEach(p => { if (p.field) value[p.field] = p.value })
  } else if (type === 'list') {
    value = editListItems.value
  } else if (type === 'set') {
    value = editSetMembers.value
  } else if (type === 'zset') {
    value = editZSetItems.value
  }
  try {
    await redisApi.setKeyValue(props.instanceId, selectedKey.value.key, {
      type,
      value,
      ttl: editTTL.value ?? undefined
    })
    message.success('保存成功')
    keyDetailDirty.value = false
    selectKey(selectedKey.value.key)
    loadKeys()
  } catch (e: any) {
    message.error(e.message || '保存失败')
  }
}

/* --- Hash 编辑 --- */

function addHashPair() {
  editHashPairs.value.push({ field: '', value: '' })
  markDirty()
}

function removeHashPair(index: number) {
  editHashPairs.value.splice(index, 1)
  markDirty()
}

/* --- List 编辑 --- */

function lpushItem() {
  if (!newListItem.value) return
  editListItems.value.unshift(newListItem.value)
  newListItem.value = ''
  markDirty()
}

function rpushItem() {
  if (!newListItem.value) return
  editListItems.value.push(newListItem.value)
  newListItem.value = ''
  markDirty()
}

function removeListItem(index: number) {
  editListItems.value.splice(index, 1)
  markDirty()
}

/* --- Set 编辑 --- */

function addSetMember() {
  if (!newSetMember.value) return
  if (!editSetMembers.value.includes(newSetMember.value)) {
    editSetMembers.value.push(newSetMember.value)
  }
  newSetMember.value = ''
  markDirty()
}

function removeSetMember(index: number) {
  editSetMembers.value.splice(index, 1)
  markDirty()
}

/* --- ZSet 编辑 --- */

function addZSetItem() {
  if (!newZSetMember.value) return
  const existing = editZSetItems.value.find(i => i.member === newZSetMember.value)
  if (existing) {
    existing.score = newZSetScore.value
  } else {
    editZSetItems.value.push({ member: newZSetMember.value, score: newZSetScore.value })
  }
  newZSetMember.value = ''
  newZSetScore.value = 0
  markDirty()
}

function removeZSetItem(index: number) {
  editZSetItems.value.splice(index, 1)
  markDirty()
}

/* --- 删除 Key --- */

const deleteKeyConfirm = ref(false)
const deleteKeyLoading = ref(false)
const targetKeyName = ref('')

function confirmDeleteKey(key: string) {
  targetKeyName.value = key
  deleteKeyConfirm.value = true
}

async function doDeleteKey() {
  if (!targetKeyName.value) return
  deleteKeyLoading.value = true
  try {
    await redisApi.deleteKey(props.instanceId, targetKeyName.value)
    message.success(`Key ${targetKeyName.value} 已删除`)
    deleteKeyConfirm.value = false
    if (selectedKey.value?.key === targetKeyName.value) {
      selectedKey.value = null
    }
    loadKeys()
  } catch (e: any) {
    message.error(e.message || '删除失败')
  } finally {
    deleteKeyLoading.value = false
  }
}

/* --- 设置 Key 过期时间 --- */

const expireKeyModal = ref(false)
const expireKeyLoading = ref(false)
const expireForm = reactive({ key: '', ttl: 3600 })

function openExpireKey(key: string) {
  expireForm.key = key
  expireForm.ttl = 3600
  expireKeyModal.value = true
}

async function submitExpireKey() {
  if (!expireForm.key) return
  if (!Number.isFinite(expireForm.ttl) || expireForm.ttl <= 0) {
    message.warning('TTL 必须为正整数（秒），-1 表示永久')
    return
  }
  expireKeyLoading.value = true
  try {
    await redisApi.expireKey(props.instanceId, expireForm.key, expireForm.ttl)
    message.success(expireForm.ttl < 0 ? `Key ${expireForm.key} 已设为永久` : `Key ${expireForm.key} 将在 ${expireForm.ttl} 秒后过期`)
    expireKeyModal.value = false
    loadKeys()
  } catch (e: any) {
    message.error(e.message || '设置过期时间失败')
  } finally {
    expireKeyLoading.value = false
  }
}

/* --- 添加 Key --- */

const addKeyModal = ref(false)
const addKeyLoading = ref(false)
const addKeyForm = ref<{ type: RedisKeyType; key: string; value: string; ttl: number | null }>({
  type: 'string',
  key: '',
  value: '',
  ttl: null
})

const addKeyTypeOptions = [
  { label: 'String', value: 'string' },
  { label: 'Hash', value: 'hash' },
  { label: 'List', value: 'list' },
  { label: 'Set', value: 'set' },
  { label: 'ZSet', value: 'zset' }
]

function openAddKey() {
  addKeyForm.value = { type: 'string', key: '', value: '', ttl: null }
  addKeyModal.value = true
}

async function submitAddKey() {
  if (!addKeyForm.value.key) { message.warning('请输入 Key 名'); return }
  addKeyLoading.value = true
  try {
    let value: any = addKeyForm.value.value
    if (addKeyForm.value.type === 'hash') {
      value = {}
      // 简单解析 field=value,field2=value2 格式
      addKeyForm.value.value.split(',').forEach((pair: string) => {
        const [k, ...v] = pair.split('=')
        if (k) value[k.trim()] = v.join('=').trim()
      })
    } else if (addKeyForm.value.type === 'list' || addKeyForm.value.type === 'set') {
      value = addKeyForm.value.value.split(',').map((s: string) => s.trim()).filter(Boolean)
    } else if (addKeyForm.value.type === 'zset') {
      value = addKeyForm.value.value.split(',').map((pair: string) => {
        const [member, score] = pair.split(':')
        return { member: member?.trim() || '', score: Number(score) || 0 }
      })
    }
    await redisApi.setKeyValue(props.instanceId, addKeyForm.value.key, {
      type: addKeyForm.value.type,
      value,
      ttl: addKeyForm.value.ttl ?? undefined
    })
    message.success('Key 创建成功')
    addKeyModal.value = false
    loadKeys()
  } catch (e: any) {
    message.error(e.message || '创建失败')
  } finally {
    addKeyLoading.value = false
  }
}

/* ==================== Tab 2: Redis CLI ==================== */

const cliCommand = ref('')
const cliLoading = ref(false)
const cliResult = ref('')
const cliHistory = ref<string[]>([])

const quickCommands = ['INFO', 'DBSIZE', 'KEYS *', 'CLIENT LIST', 'CONFIG GET maxmemory']

async function runCLI(cmd?: string) {
  const command = (cmd ?? cliCommand.value).trim()
  if (!command) { message.warning('请输入命令'); return }
  cliLoading.value = true
  try {
    const res = await redisApi.execCLI(props.instanceId, command)
    cliResult.value = res.data.result || res.data.error || '(empty)'
    if (!cliHistory.value.includes(command)) {
      cliHistory.value.unshift(command)
      if (cliHistory.value.length > 20) cliHistory.value.pop()
    }
    if (res.data.error) {
      message.error(res.data.error)
    }
  } catch (e: any) {
    cliResult.value = `ERROR: ${e.message}`
    message.error(e.message || '执行失败')
  } finally {
    cliLoading.value = false
  }
}

function useHistoryCommand(cmd: string) {
  cliCommand.value = cmd
}

/* ==================== Tab 3: 内存分析 ==================== */

const memoryInfo = ref<RedisMemoryInfo | null>(null)
const memoryLoading = ref(false)
const bigKeys = ref<RedisBigKey[]>([])
const bigKeysLoading = ref(false)
const infoContent = ref('')
const infoLoading = ref(false)
const showInfo = ref(false)

async function loadMemory() {
  memoryLoading.value = true
  try {
    const res = await redisApi.getMemory(props.instanceId)
    memoryInfo.value = res.data
  } catch (e: any) {
    message.error(e.message || '加载内存信息失败')
  } finally {
    memoryLoading.value = false
  }
}

async function loadInfo() {
  infoLoading.value = true
  try {
    const res = await redisApi.getInfo(props.instanceId)
    if (typeof res.data === 'string') {
      infoContent.value = res.data
    } else {
      infoContent.value = JSON.stringify(res.data, null, 2)
    }
  } catch (e: any) {
    infoContent.value = `加载失败: ${e.message}`
  } finally {
    infoLoading.value = false
  }
}

const memoryUsagePercent = computed(() => {
  if (!memoryInfo.value || !memoryInfo.value.maxmemory || memoryInfo.value.maxmemory === 0) return 0
  return Math.round((memoryInfo.value.used_memory / memoryInfo.value.maxmemory) * 100)
})

const bigKeyColumns: DataTableColumns = [
  { title: 'Key', key: 'key', render: (row: any) => monoEl(row.key) },
  {
    title: '类型', key: 'type', width: 100,
    render: (row: any) => h(StatusTag, {
      status: keyTypeStatusMap[row.type as RedisKeyType] || 'stopped',
      text: row.type?.toUpperCase()
    })
  },
  { title: '大小', key: 'size_human', width: 120 },
  {
    title: '占比', key: 'ratio', width: 120,
    render: (row: any) => row.ratio !== undefined ? `${row.ratio.toFixed(1)}%` : '-'
  }
]

function toggleInfo() {
  showInfo.value = !showInfo.value
  if (showInfo.value && !infoContent.value) {
    loadInfo()
  }
}

/* ==================== 生命周期 ==================== */

onMounted(() => {
  loadKeys()
})
</script>

<template>
  <NTabs v-model:value="activeSubTab" type="line" animated>
    <!-- ==================== Tab 1: Key 浏览器 ==================== -->
    <NTabPane name="keys" tab="Key 浏览器">
      <div class="tab-toolbar">
        <NSpace>
          <NInput
            v-model:value="searchPattern"
            placeholder="pattern，支持 * 通配，如 user:*"
            style="width: var(--input-3xl)"
            @keyup.enter="searchKeys"
          >
            <template #prefix><NIcon :size="16"><SearchOutline /></NIcon></template>
          </NInput>
          <NButton type="primary" @click="searchKeys" :loading="keyListLoading">
            <template #icon><NIcon><SearchOutline /></NIcon></template>搜索
          </NButton>
          <NButton @click="loadKeys" :loading="keyListLoading">
            <template #icon><NIcon><RefreshOutline /></NIcon></template>刷新
          </NButton>
          <NButton type="primary" @click="openAddKey">
            <template #icon><NIcon><AddOutline /></NIcon></template>添加 Key
          </NButton>
        </NSpace>
      </div>

      <div class="redis-layout">
        <!-- 左侧 Key 列表 -->
        <div class="redis-keylist">
          <NSpin :show="keyListLoading">
            <DataTable
              :columns="keyColumns"
              :data="keyList"
              :pagination="false"
              size="small"
              :row-key="(row: any) => row.key"
            />
          </NSpin>
          <div v-if="keyList.length === 0 && !keyListLoading" class="empty-hint">
            <NEmpty description="暂无 Key" />
          </div>
          <div class="redis-pagination">
            <NSpace>
              <NButton size="tiny" :disabled="scanCursorStack.length === 0" @click="prevPage">
                <template #icon><NIcon><ChevronBackOutline /></NIcon></template>上一页
              </NButton>
              <span class="mono redis-pagination__info">cursor: {{ scanCursor }}</span>
              <NButton size="tiny" :disabled="!hasMoreKeys" @click="nextPage">
                下一页<template #icon><NIcon><ChevronForwardOutline /></NIcon></template>
              </NButton>
              <span v-if="totalKeys > 0" class="mono redis-pagination__total">共 {{ totalKeys }} 个 Key</span>
            </NSpace>
          </div>
        </div>

        <!-- 右侧 Key 详情 -->
        <div class="redis-detail">
          <div v-if="!selectedKey" class="empty-hint">
            <NEmpty description="点击左侧 Key 查看详情" />
          </div>

          <template v-else>
            <NSpin :show="selectedKeyLoading">
              <div class="redis-detail__header">
                <NSpace>
                  <NIcon :size="16"><KeyOutline /></NIcon>
                  <span class="mono redis-detail__keyname">{{ selectedKey.key }}</span>
                  <StatusTag
                    :status="keyTypeStatusMap[selectedKey.type] || 'stopped'"
                    :text="selectedKey.type.toUpperCase()"
                  />
                </NSpace>
                <NSpace>
                  <NButton size="tiny" @click="openExpireKey(selectedKey.key)">
                    <template #icon><NIcon><TimerOutline /></NIcon></template>TTL
                  </NButton>
                  <NButton size="tiny" type="primary" :loading="false" @click="saveKey" :disabled="!keyDetailDirty">
                    <template #icon><NIcon><SaveOutline /></NIcon></template>保存
                  </NButton>
                  <NButton size="tiny" type="error" @click="confirmDeleteKey(selectedKey.key)">
                    <template #icon><NIcon><TrashOutline /></NIcon></template>删除
                  </NButton>
                </NSpace>
              </div>

              <div class="redis-detail__meta">
                <NSpace>
                  <span class="mono">TTL: {{ formatTTL(selectedKey.ttl) }}</span>
                  <span class="mono">大小: {{ formatBytes(selectedKey.size) }}</span>
                </NSpace>
              </div>

              <!-- TTL 设置 -->
              <div class="redis-detail__section">
                <span class="redis-detail__label">TTL（秒，-1 为永久）</span>
                <NInputNumber
                  v-model:value="editTTL"
                  placeholder="-1 永久"
                  :min="-1"
                  style="width: var(--input-md)"
                  @update:value="markDirty"
                />
              </div>

              <!-- String 编辑器 -->
              <div v-if="selectedKey.type === 'string'" class="redis-detail__section">
                <span class="redis-detail__label">值</span>
                <NInput
                  v-model:value="editStringValue"
                  type="textarea"
                  :rows="6"
                  placeholder="字符串值"
                  @update:value="markDirty"
                  style="font-family: var(--font-mono)"
                />
              </div>

              <!-- Hash 编辑器 -->
              <div v-else-if="selectedKey.type === 'hash'" class="redis-detail__section">
                <div class="redis-detail__section-header">
                  <span class="redis-detail__label">字段 (Field / Value)</span>
                  <NButton size="tiny" @click="addHashPair">
                    <template #icon><NIcon><AddOutline /></NIcon></template>添加字段
                  </NButton>
                </div>
                <div class="hash-editor">
                  <div v-for="(pair, idx) in editHashPairs" :key="idx" class="hash-editor__row">
                    <NInput v-model:value="pair.field" placeholder="field" size="small" @update:value="markDirty" />
                    <NInput v-model:value="pair.value" placeholder="value" size="small" @update:value="markDirty" />
                    <NButton size="tiny" type="error" quaternary @click="removeHashPair(idx)">
                      <template #icon><NIcon><TrashOutline /></NIcon></template>
                    </NButton>
                  </div>
                  <div v-if="editHashPairs.length === 0" class="empty-hint">
                    <NEmpty description="暂无字段" />
                  </div>
                </div>
              </div>

              <!-- List 编辑器 -->
              <div v-else-if="selectedKey.type === 'list'" class="redis-detail__section">
                <div class="redis-detail__section-header">
                  <span class="redis-detail__label">列表项</span>
                  <NSpace>
                    <NInput v-model:value="newListItem" placeholder="新元素" size="small" style="width: var(--input-lg)" @keyup.enter="lpushItem" />
                    <NButton size="tiny" @click="lpushItem">LPUSH</NButton>
                    <NButton size="tiny" @click="rpushItem">RPUSH</NButton>
                  </NSpace>
                </div>
                <div class="list-editor">
                  <div v-for="(item, idx) in editListItems" :key="idx" class="list-editor__row">
                    <span class="mono list-editor__index">[{{ idx }}]</span>
                    <NInput v-model:value="editListItems[idx]" size="small" @update:value="markDirty" />
                    <NButton size="tiny" type="error" quaternary @click="removeListItem(idx)">
                      <template #icon><NIcon><TrashOutline /></NIcon></template>
                    </NButton>
                  </div>
                  <div v-if="editListItems.length === 0" class="empty-hint">
                    <NEmpty description="列表为空" />
                  </div>
                </div>
              </div>

              <!-- Set 编辑器 -->
              <div v-else-if="selectedKey.type === 'set'" class="redis-detail__section">
                <div class="redis-detail__section-header">
                  <span class="redis-detail__label">成员</span>
                  <NSpace>
                    <NInput v-model:value="newSetMember" placeholder="新成员" size="small" style="width: var(--input-lg)" @keyup.enter="addSetMember" />
                    <NButton size="tiny" @click="addSetMember">
                      <template #icon><NIcon><AddOutline /></NIcon></template>添加
                    </NButton>
                  </NSpace>
                </div>
                <div class="set-editor">
                  <div v-for="(member, idx) in editSetMembers" :key="idx" class="set-editor__row">
                    <NInput v-model:value="editSetMembers[idx]" size="small" @update:value="markDirty" />
                    <NButton size="tiny" type="error" quaternary @click="removeSetMember(idx)">
                      <template #icon><NIcon><TrashOutline /></NIcon></template>
                    </NButton>
                  </div>
                  <div v-if="editSetMembers.length === 0" class="empty-hint">
                    <NEmpty description="集合为空" />
                  </div>
                </div>
              </div>

              <!-- ZSet 编辑器 -->
              <div v-else-if="selectedKey.type === 'zset'" class="redis-detail__section">
                <div class="redis-detail__section-header">
                  <span class="redis-detail__label">成员 / 分数</span>
                  <NSpace>
                    <NInput v-model:value="newZSetMember" placeholder="成员" size="small" style="width: var(--input-md)" />
                    <NInputNumber v-model:value="newZSetScore" size="small" style="width: var(--input-sm)" />
                    <NButton size="tiny" @click="addZSetItem">
                      <template #icon><NIcon><AddOutline /></NIcon></template>添加
                    </NButton>
                  </NSpace>
                </div>
                <div class="zset-editor">
                  <div v-for="(item, idx) in editZSetItems" :key="idx" class="zset-editor__row">
                    <NInput v-model:value="item.member" size="small" @update:value="markDirty" />
                    <NInputNumber v-model:value="item.score" size="small" style="width: var(--input-sm)" @update:value="markDirty" />
                    <NButton size="tiny" type="error" quaternary @click="removeZSetItem(idx)">
                      <template #icon><NIcon><TrashOutline /></NIcon></template>
                    </NButton>
                  </div>
                  <div v-if="editZSetItems.length === 0" class="empty-hint">
                    <NEmpty description="有序集合为空" />
                  </div>
                </div>
              </div>
            </NSpin>
          </template>
        </div>
      </div>
    </NTabPane>

    <!-- ==================== Tab 2: Redis CLI ==================== -->
    <NTabPane name="cli" tab="Redis CLI">
      <div class="tab-toolbar">
        <NSpace>
          <NSelect
            v-model:value="cliCommand"
            placeholder="历史命令"
            :options="cliHistory.map(c => ({ label: c, value: c }))"
            style="width: var(--input-3xl)"
            @update:value="useHistoryCommand"
          />
          <NButton type="primary" :loading="cliLoading" @click="runCLI()">
            <template #icon><NIcon><TerminalOutline /></NIcon></template>执行
          </NButton>
        </NSpace>
      </div>

      <div class="cli-quick">
        <NSpace>
          <span class="cli-quick__label">快捷命令:</span>
          <NButton
            v-for="cmd in quickCommands"
            :key="cmd"
            size="tiny"
            text
            @click="runCLI(cmd)"
          >
            {{ cmd }}
          </NButton>
        </NSpace>
      </div>

      <div class="cli-editor">
        <CodeEditor
          v-model="cliCommand"
          language="shell"
          height="120px"
          :minimap="false"
        />
      </div>

      <div class="cli-result">
        <div class="cli-result__header">
          <NIcon :size="14"><TerminalOutline /></NIcon>
          <span>输出</span>
        </div>
        <pre class="cli-result__output">{{ cliResult || '(等待执行...)' }}</pre>
      </div>
    </NTabPane>

    <!-- ==================== Tab 3: 内存分析 ==================== -->
    <NTabPane name="memory" tab="内存分析">
      <div class="tab-toolbar">
        <NSpace>
          <NButton type="primary" @click="loadMemory" :loading="memoryLoading">
            <template #icon><NIcon><RefreshOutline /></NIcon></template>刷新
          </NButton>
          <NButton @click="toggleInfo">
            <template #icon><NIcon><HardwareChipOutline /></NIcon></template>{{ showInfo ? '隐藏' : '查看' }} INFO
          </NButton>
        </NSpace>
      </div>

      <NSpin :show="memoryLoading">
        <div class="stat-row" v-if="memoryInfo">
          <StatCard title="已用内存" :value="memoryInfo.used_memory_human" accent="acc" />
          <StatCard title="最大内存" :value="memoryInfo.maxmemory_human" accent="info" />
          <StatCard title="Key 数量" :value="memoryInfo.total_keys ?? '-'" accent="acc2" />
          <StatCard title="内存策略" :value="memoryInfo.maxmemory_policy" accent="warn" />
        </div>
      </NSpin>

      <div v-if="memoryInfo && memoryInfo.maxmemory > 0" class="memory-progress">
        <div class="memory-progress__label">
          <span>内存使用率</span>
          <span class="mono">{{ memoryUsagePercent }}%</span>
        </div>
        <NProgress
          :percentage="memoryUsagePercent"
          :color="memoryUsagePercent > 80 ? 'var(--err)' : memoryUsagePercent > 60 ? 'var(--warn)' : 'var(--acc2)'"
          :rail-color="'var(--surface3)'"
          :height="8"
        />
      </div>

      <div v-if="memoryInfo?.mem_fragmentation_ratio" class="memory-frag">
        <NAlert type="info" :show-icon="true">
          内存碎片率: <span class="mono">{{ memoryInfo.mem_fragmentation_ratio }}</span>
        </NAlert>
      </div>

      <!-- 大 Key 排行 -->
      <div class="bigkeys-section" v-if="bigKeys.length > 0">
        <div class="bigkeys-section__title">大 Key 排行</div>
        <DataTable :columns="bigKeyColumns" :data="bigKeys" :pagination="false" size="small" />
      </div>

      <!-- INFO 全量信息 -->
      <div v-if="showInfo" class="info-section">
        <div class="info-section__header">
          <span>INFO 全量信息</span>
          <NButton size="tiny" text @click="showInfo = false">
            <template #icon><NIcon><CloseOutline /></NIcon></template>
          </NButton>
        </div>
        <NSpin :show="infoLoading">
          <CodeEditor
            :model-value="infoContent"
            language="ini"
            height="400px"
            read-only
            :minimap="false"
          />
        </NSpin>
      </div>

      <div v-if="!memoryInfo && !memoryLoading" class="empty-hint">
        <NEmpty description="点击刷新加载内存信息" />
      </div>
    </NTabPane>
  </NTabs>

  <!-- ==================== Modals ==================== -->

  <!-- 添加 Key -->
  <ModalForm
    v-model:show="addKeyModal"
    title="添加 Key"
    :loading="addKeyLoading"
    @submit="submitAddKey"
  >
    <NFormItem label="类型">
      <NSelect v-model:value="addKeyForm.type" :options="addKeyTypeOptions" />
    </NFormItem>
    <NFormItem label="Key 名">
      <NInput v-model:value="addKeyForm.key" placeholder="例如: user:1" />
    </NFormItem>
    <NFormItem label="值">
      <NInput
        v-model:value="addKeyForm.value"
        type="textarea"
        :rows="4"
        :placeholder="addKeyForm.type === 'string' ? '字符串值' : addKeyForm.type === 'hash' ? 'field1=value1,field2=value2' : addKeyForm.type === 'zset' ? 'member1:1.0,member2:2.0' : 'item1,item2,item3'"
        style="font-family: var(--font-mono)"
      />
    </NFormItem>
    <NFormItem label="TTL（秒）">
      <NInputNumber v-model:value="addKeyForm.ttl" placeholder="留空为永久" :min="-1" style="width: 100%" />
    </NFormItem>
  </ModalForm>

  <!-- 删除 Key 确认 -->
  <ConfirmDialog
    v-model:show="deleteKeyConfirm"
    title="删除 Key"
    :content="`确定要删除 Key「${targetKeyName}」吗？此操作不可恢复。`"
    type="error"
    confirm-text="删除"
    :loading="deleteKeyLoading"
    @confirm="doDeleteKey"
  />

  <!-- 设置 Key 过期时间 -->
  <NModal v-model:show="expireKeyModal" preset="card" title="设置 Key 过期时间" :style="{ width: 'min(420px, 92vw)' }">
    <NFormItem label="Key">
      <NInput :value="expireForm.key" disabled />
    </NFormItem>
    <NFormItem label="TTL（秒）">
      <NInputNumber v-model:value="expireForm.ttl" :min="-1" style="width: 100%;" placeholder="秒数，-1 表示永久" />
    </NFormItem>
    <template #footer>
      <NSpace justify="end">
        <NButton @click="expireKeyModal = false">取消</NButton>
        <NButton type="primary" :loading="expireKeyLoading" @click="submitExpireKey">确定</NButton>
      </NSpace>
    </template>
  </NModal>
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

/* Redis 布局 */
.redis-layout {
  display: flex;
  gap: var(--space-4);
  min-height: var(--chart-height);
}

.redis-keylist {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.redis-pagination {
  margin-top: var(--space-3);
  display: flex;
  justify-content: center;
}

.redis-pagination__info,
.redis-pagination__total {
  font-size: var(--fs-caption);
  color: var(--faint);
}

.redis-detail {
  width: 480px;
  flex-shrink: 0;
  background: var(--surface2);
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  padding: var(--space-4);
  overflow-y: auto;
  max-height: 70vh;
}

/* 移动端：键列表与详情上下堆叠，宽度铺满 */
@media (max-width: 768px) {
  .redis-layout {
    flex-direction: column;
    min-height: auto;
  }

  .redis-detail {
    width: 100%;
    max-height: 60vh;
  }

  .redis-keylist {
    max-height: 45vh;
  }
}

.redis-detail__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-3);
  padding-bottom: var(--space-3);
  border-bottom: 1px solid var(--line);
}

.redis-detail__keyname {
  font-size: var(--fs-title);
  font-weight: 600;
  color: var(--ink);
  word-break: break-all;
}

.redis-detail__meta {
  margin-bottom: var(--space-4);
  font-size: var(--fs-caption);
  color: var(--faint);
}

.redis-detail__section {
  margin-bottom: var(--space-4);
}

.redis-detail__section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-2);
}

.redis-detail__label {
  font-size: var(--fs-caption);
  font-weight: 600;
  color: var(--sub);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  display: block;
  margin-bottom: var(--space-2);
}

/* Hash 编辑器 */
.hash-editor__row,
.list-editor__row,
.set-editor__row,
.zset-editor__row {
  display: flex;
  gap: var(--space-2);
  margin-bottom: var(--space-2);
  align-items: center;
}

.hash-editor__row .n-input,
.set-editor__row .n-input {
  flex: 1;
}

.list-editor__index {
  width: 40px;
  flex-shrink: 0;
  font-size: var(--fs-caption);
  color: var(--faint);
}

.list-editor__row .n-input {
  flex: 1;
}

/* CLI */
.cli-quick {
  margin-bottom: var(--space-3);
}

.cli-quick__label {
  font-size: var(--fs-caption);
  color: var(--faint);
}

.cli-editor {
  margin-bottom: var(--space-4);
}

.cli-result {
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.cli-result__header {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  background: var(--surface2);
  border-bottom: 1px solid var(--line);
  font-size: var(--fs-caption);
  font-weight: 600;
  color: var(--sub);
}

.cli-result__output {
  padding: var(--space-3);
  margin: 0;
  font-family: var(--font-mono);
  font-size: var(--fs-mono);
  color: var(--ink);
  background: var(--bg);
  max-height: 400px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

/* 内存分析 */
.stat-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
  gap: var(--space-4);
  margin-bottom: var(--space-4);
}

.memory-progress {
  margin-bottom: var(--space-4);
  padding: var(--space-4);
  background: var(--surface2);
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
}

.memory-progress__label {
  display: flex;
  justify-content: space-between;
  margin-bottom: var(--space-2);
  font-size: var(--fs-body);
  color: var(--sub);
}

.memory-frag {
  margin-bottom: var(--space-4);
}

.bigkeys-section {
  margin-bottom: var(--space-4);
}

.bigkeys-section__title {
  font-size: var(--fs-body);
  font-weight: 600;
  color: var(--ink);
  margin-bottom: var(--space-3);
}

.info-section {
  margin-top: var(--space-4);
}

.info-section__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-2);
  font-size: var(--fs-body);
  font-weight: 600;
  color: var(--ink);
}

:deep(.n-tabs-nav) {
  margin-bottom: var(--space-4);
}
</style>
