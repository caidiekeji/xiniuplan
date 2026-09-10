<script setup lang="ts">
import { ref, computed, h } from 'vue'
import { useMessage } from 'naive-ui'
import {
  NButton, NSpace, NIcon, NInput, NInputNumber, NRadioGroup, NRadio,
  NFormItem, NEmpty, NModal, NAlert
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {
  AddOutline, TrashOutline, PlayOutline, StopOutline,
  RefreshOutline, ConstructOutline, CloudDownloadOutline, SearchOutline
} from '@vicons/ionicons5'
import { DataTable, ModalForm, StatusTag, ConfirmDialog } from '@/components'
import type { StatusType } from '@/components'
import * as dbApi from '@/api/database'
import type { DBInstance, CreateInstanceData, DBType } from '@/api/database'
import { DB_VERSION_OPTIONS, DB_DEFAULT_PORT, DB_TYPE_LABELS } from '@/api/database'
import { monoEl, iconBtn, btnGroup, formatTime, typeTag } from './shared'

const message = useMessage()

/* ==================== Props / Emits ==================== */

const props = defineProps<{
  instances: DBInstance[]
  loading: boolean
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{
  (e: 'refresh'): void
  (e: 'manage', instance: DBInstance): void
  (e: 'update:page', val: number): void
  (e: 'update:pageSize', val: number): void
}>()

/* ==================== 列定义 ==================== */

const instanceColumns: DataTableColumns = [
  { title: '实例名', key: 'name', width: 160 },
  {
    title: '类型', key: 'type', width: 130,
    render: (row: any) => typeTag(row.type as DBType)
  },
  {
    title: '版本', key: 'version', width: 100,
    render: (row: any) => monoEl(row.version)
  },
  { title: '端口', key: 'port', width: 80, render: (row: any) => monoEl(row.port) },
  {
    title: '状态', key: 'status', width: 100,
    render: (row: any) => h(StatusTag, { status: row.status as StatusType })
  },
  {
    title: '容器ID', key: 'container_id', width: 140,
    render: (row: any) => monoEl(row.container_id?.substring(0, 12))
  },
  {
    title: '创建时间', key: 'created_at', width: 170,
    render: (row: any) => formatTime(row.created_at)
  },
  {
    title: '操作', key: 'actions', width: 280,
    render: (row: any) => {
      const isRunning = row.status === 'running'
      const children: any[] = []
      if (!isRunning) {
        children.push(iconBtn({ text: '启动', icon: PlayOutline, type: 'success', onClick: () => handleStart(row) }))
      } else {
        children.push(iconBtn({ text: '停止', icon: StopOutline, type: 'warning', onClick: () => handleStop(row) }))
      }
      children.push(iconBtn({ text: '重启', icon: RefreshOutline, onClick: () => handleRestart(row) }))
      children.push(iconBtn({ text: '管理', icon: ConstructOutline, type: 'primary', onClick: () => handleManage(row) }))
      children.push(iconBtn({ text: '删除', icon: TrashOutline, type: 'error', onClick: () => confirmDeleteInstance(row) }))
      return btnGroup(children)
    }
  }
]

const pagination = computed(() => ({
  page: props.page,
  pageSize: props.pageSize,
  itemCount: props.total,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  'onUpdate:page': (p: number) => emit('update:page', p),
  'onUpdate:page-size': (s: number) => emit('update:pageSize', s)
}))

/* ==================== 操作 ==================== */

async function handleStart(row: any) {
  try {
    await dbApi.startInstance(row.id)
    message.success(`实例 ${row.name} 已启动`)
    emit('refresh')
  } catch (e: any) {
    message.error(e.message || '启动失败')
  }
}

async function handleStop(row: any) {
  try {
    await dbApi.stopInstance(row.id)
    message.success(`实例 ${row.name} 已停止`)
    emit('refresh')
  } catch (e: any) {
    message.error(e.message || '停止失败')
  }
}

async function handleRestart(row: any) {
  try {
    await dbApi.restartInstance(row.id)
    message.success(`实例 ${row.name} 已重启`)
    emit('refresh')
  } catch (e: any) {
    message.error(e.message || '重启失败')
  }
}

function handleManage(row: DBInstance) {
  emit('manage', row)
}

/* ==================== 创建实例 Modal ==================== */

const createInstanceModal = ref(false)
const createInstanceLoading = ref(false)
const instanceForm = ref<CreateInstanceData>({
  type: 'mysql',
  version: '8.0',
  port: undefined,
  root_password: '',
  remark: ''
})

function openCreateInstance() {
  instanceForm.value = {
    type: 'mysql',
    version: DB_VERSION_OPTIONS.mysql[0].value,
    port: DB_DEFAULT_PORT.mysql,
    root_password: '',
    remark: ''
  }
  createInstanceModal.value = true
}

function onInstanceTypeChange(type: DBType) {
  instanceForm.value.type = type
  instanceForm.value.version = DB_VERSION_OPTIONS[type][0].value
  instanceForm.value.port = DB_DEFAULT_PORT[type]
}

async function submitCreateInstance() {
  createInstanceLoading.value = true
  try {
    const data: CreateInstanceData = { ...instanceForm.value }
    if (!data.root_password) delete data.root_password
    if (!data.port) delete data.port
    await dbApi.createInstance(data)
    message.success('实例创建成功')
    createInstanceModal.value = false
    emit('refresh')
  } catch (e: any) {
    message.error(e.message || '创建失败')
  } finally {
    createInstanceLoading.value = false
  }
}

/* ==================== 导入已有实例 ==================== */

const importModal = ref(false)
const importLoading = ref(false)
const discoverLoading = ref(false)
const discovered = ref<dbApi.DiscoveredDB[]>([])
const importForm = ref<dbApi.ImportInstanceData>({
  type: 'mysql',
  host: '127.0.0.1',
  port: 3306,
  username: '',
  password: '',
  version: '',
  remark: ''
})

function openImportModal() {
  importForm.value = { type: 'mysql', host: '127.0.0.1', port: 3306, username: '', password: '', version: '', remark: '' }
  discovered.value = []
  importModal.value = true
}

async function handleDiscover() {
  discoverLoading.value = true
  try {
    const res = await dbApi.discoverInstances()
    discovered.value = res.data
    if (res.data.length === 0) {
      message.info('未在本机发现运行中的数据库（已探测 3306/5432/6379 端口）')
    }
  } catch (e: any) {
    message.error(e.message || '探测失败')
  } finally {
    discoverLoading.value = false
  }
}

function useDiscovered(d: dbApi.DiscoveredDB) {
  importForm.value.type = d.type
  importForm.value.host = d.host
  importForm.value.port = d.port
  importForm.value.version = d.version || ''
}

function onImportTypeChange(type: string) {
  importForm.value.type = type
  importForm.value.port = DB_DEFAULT_PORT[type as DBType]
}

async function submitImport() {
  importLoading.value = true
  try {
    const data: dbApi.ImportInstanceData = { ...importForm.value }
    if (!data.username) delete data.username
    if (!data.password) delete data.password
    if (!data.version) delete data.version
    if (!data.remark) delete data.remark
    await dbApi.importInstance(data)
    message.success('实例导入成功')
    importModal.value = false
    emit('refresh')
  } catch (e: any) {
    message.error(e.message || '导入失败')
  } finally {
    importLoading.value = false
  }
}

/* ==================== 删除实例确认 ==================== */

const deleteInstanceConfirm = ref(false)
const deleteInstanceLoading = ref(false)
const targetInstance = ref<DBInstance | null>(null)

function confirmDeleteInstance(row: any) {
  targetInstance.value = row
  deleteInstanceConfirm.value = true
}

async function doDeleteInstance() {
  if (!targetInstance.value) return
  deleteInstanceLoading.value = true
  try {
    await dbApi.deleteInstance(targetInstance.value.id)
    message.success(`实例 ${targetInstance.value.name} 已删除`)
    deleteInstanceConfirm.value = false
    emit('refresh')
  } catch (e: any) {
    message.error(e.message || '删除失败')
  } finally {
    deleteInstanceLoading.value = false
  }
}
</script>

<template>
  <div class="instance-tab">
    <div class="tab-toolbar">
      <NSpace>
        <NButton type="primary" @click="openCreateInstance">
          <template #icon><NIcon><AddOutline /></NIcon></template>创建实例
        </NButton>
        <NButton type="info" ghost @click="openImportModal">
          <template #icon><NIcon><CloudDownloadOutline /></NIcon></template>导入已有实例
        </NButton>
        <NButton @click="$emit('refresh')" :loading="loading">
          <template #icon><NIcon><RefreshOutline /></NIcon></template>刷新
        </NButton>
      </NSpace>
    </div>

    <DataTable
      :columns="instanceColumns"
      :data="instances"
      :loading="loading"
      :pagination="pagination"
    />

    <div v-if="instances.length === 0 && !loading" class="empty-hint">
      <NEmpty description="暂无数据库实例，点击上方按钮创建">
        <NButton type="primary" size="small" @click="openCreateInstance">创建实例</NButton>
      </NEmpty>
    </div>

    <!-- 导入已有实例 -->
    <NModal v-model:show="importModal" preset="card" title="导入已有数据库实例" :style="{ width: 'min(520px, 92vw)' }">
      <NAlert type="info" :show-icon="true" style="margin-bottom: 14px;">
        导入本机已安装但并非由面板创建的数据库实例（不会创建容器）。可先点击"自动探测"。
      </NAlert>

      <NSpace style="margin-bottom: 14px;">
        <NButton size="small" :loading="discoverLoading" @click="handleDiscover">
          <template #icon><NIcon><SearchOutline /></NIcon></template>自动探测本机数据库
        </NButton>
      </NSpace>

      <div v-if="discovered.length > 0" class="discover-list">
        <div
          v-for="(d, i) in discovered"
          :key="i"
          class="discover-item"
          @click="useDiscovered(d)"
        >
          <span>{{ d.type.toUpperCase() }} @ {{ d.host }}:{{ d.port }}</span>
          <span class="discover-item__ver">{{ d.version || '版本未知' }}</span>
          <span class="discover-item__use">点击使用</span>
        </div>
      </div>

      <NFormItem label="类型">
        <NRadioGroup :value="importForm.type" @update:value="onImportTypeChange">
          <NRadio value="mysql">MySQL</NRadio>
          <NRadio value="postgres">PostgreSQL</NRadio>
          <NRadio value="redis">Redis</NRadio>
        </NRadioGroup>
      </NFormItem>
      <NFormItem label="地址">
        <NInput v-model:value="importForm.host" placeholder="127.0.0.1" />
      </NFormItem>
      <NFormItem label="端口">
        <NInputNumber v-model:value="importForm.port" :min="1" :max="65535" style="width: 100%;" />
      </NFormItem>
      <NFormItem v-if="importForm.type !== 'redis'" label="管理员用户名">
        <NInput v-model:value="importForm.username" placeholder="root / postgres" />
      </NFormItem>
      <NFormItem v-if="importForm.type !== 'redis'" label="管理员密码">
        <NInput v-model:value="importForm.password" type="password" show-password-on="click" placeholder="可留空" />
      </NFormItem>
      <NFormItem label="备注">
        <NInput v-model:value="importForm.remark" placeholder="可选" />
      </NFormItem>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="importModal = false">取消</NButton>
          <NButton type="primary" :loading="importLoading" @click="submitImport">导入</NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- 创建实例 -->
    <ModalForm
      v-model:show="createInstanceModal"
      title="创建数据库实例"
      :loading="createInstanceLoading"
      @submit="submitCreateInstance"
    >
      <NFormItem label="类型">
        <div class="type-cards">
          <div
            v-for="t in (['mysql','postgres','redis'] as DBType[])"
            :key="t"
            class="type-card"
            :class="{ 'type-card--active': instanceForm.type === t }"
            @click="onInstanceTypeChange(t)"
          >
            <span class="type-card__name">{{ DB_TYPE_LABELS[t] }}</span>
            <span class="type-card__port">:{{ DB_DEFAULT_PORT[t] }}</span>
          </div>
        </div>
      </NFormItem>
      <NFormItem label="版本">
        <NRadioGroup v-model:value="instanceForm.version">
          <NSpace>
            <NRadio v-for="opt in DB_VERSION_OPTIONS[instanceForm.type]" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </NRadio>
          </NSpace>
        </NRadioGroup>
      </NFormItem>
      <NFormItem label="端口">
        <NInputNumber v-model:value="instanceForm.port" :placeholder="String(DB_DEFAULT_PORT[instanceForm.type])" :min="1" :max="65535" style="width: 100%" />
        <div class="form-hint">默认端口 {{ DB_DEFAULT_PORT[instanceForm.type] }}，留空将自动分配可用端口</div>
      </NFormItem>
      <NFormItem v-if="instanceForm.type !== 'redis'" label="root 密码">
        <NInput v-model:value="instanceForm.root_password" type="password" show-password-on="click" placeholder="留空自动生成" />
        <div class="form-hint">留空将自动生成强密码</div>
      </NFormItem>
      <NFormItem label="备注">
        <NInput v-model:value="instanceForm.remark" placeholder="可选" />
      </NFormItem>
    </ModalForm>

    <!-- 删除实例确认 -->
    <ConfirmDialog
      v-model:show="deleteInstanceConfirm"
      title="删除实例"
      :content="`确定要删除实例「${targetInstance?.name}」吗？此操作不可恢复，所有数据将丢失。`"
      type="error"
      confirm-text="删除"
      :loading="deleteInstanceLoading"
      @confirm="doDeleteInstance"
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

.form-hint {
  font-size: var(--fs-caption);
  color: var(--faint);
  margin-top: var(--space-1);
}

/* 类型选择卡片 */
.type-cards {
  display: flex;
  gap: var(--space-3);
}

.type-card {
  flex: 1;
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--t-fast) var(--ease);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-1);
  background: var(--surface);
}

.type-card:hover {
  border-color: var(--line-strong);
  background: var(--surface2);
}

.type-card--active {
  border-color: var(--acc);
  background: var(--acc-soft);
}

.type-card__name {
  font-size: var(--fs-body);
  font-weight: 600;
  color: var(--ink);
}

.type-card--active .type-card__name {
  color: var(--acc);
}

.type-card__port {
  font-size: var(--fs-caption);
  color: var(--faint);
  font-family: var(--font-mono);
}

.discover-list {
  margin-bottom: 14px;
  border: 1px solid var(--line);
  border-radius: 8px;
  overflow: hidden;
}

.discover-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  font-size: 13px;
  cursor: pointer;
  transition: background var(--t-fast) var(--ease);
}

.discover-item + .discover-item {
  border-top: 1px solid var(--line);
}

.discover-item:hover {
  background: var(--info-soft);
}

.discover-item__ver {
  flex: 1;
  color: var(--faint);
  font-family: var(--font-mono);
  font-size: 12px;
}

.discover-item__use {
  color: var(--acc);
  font-size: 12px;
}
</style>
