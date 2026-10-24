<script setup lang="ts">
import { ref, computed, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from '@/composables/useMessage'
import {
  NButton, NSpace, NIcon, NInput, NInputNumber, NRadioGroup, NRadio,
  NFormItem, NEmpty, NModal, NAlert
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {
  AddOutline, TrashOutline, PlayOutline, StopOutline,
  RefreshOutline, ConstructOutline, CloudDownloadOutline, SearchOutline
} from '@vicons/ionicons5'
import { DataTable, ModalForm, ContentModal, StatusTag, ConfirmDialog } from '@/components'
import type { StatusType } from '@/components'
import * as dbApi from '@/api/database'
import type { DBInstance, CreateInstanceData, DBType } from '@/api/database'
import { DB_VERSION_OPTIONS, DB_DEFAULT_PORT, DB_TYPE_LABELS } from '@/api/database'
import { monoEl, iconBtn, btnGroup, formatTime, typeTag } from './shared'

const message = useMessage()
const { t } = useI18n()

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

const instanceColumns = computed<DataTableColumns>(() => [
  { title: t('database.instances.columns.name'), key: 'name', width: 160 },
  {
    title: t('database.instances.columns.type'), key: 'type', width: 130,
    render: (row: any) => typeTag(row.type as DBType)
  },
  {
    title: t('database.instances.columns.version'), key: 'version', width: 100,
    render: (row: any) => monoEl(row.version)
  },
  { title: t('database.instances.columns.port'), key: 'port', width: 80, render: (row: any) => monoEl(row.port) },
  {
    title: t('database.instances.columns.status'), key: 'status', width: 100,
    render: (row: any) => h(StatusTag, { status: row.status as StatusType })
  },
  {
    title: t('database.instances.columns.containerId'), key: 'container_id', width: 140,
    render: (row: any) => monoEl(row.container_id?.substring(0, 12))
  },
  {
    title: t('database.instances.columns.createdAt'), key: 'created_at', width: 170,
    render: (row: any) => formatTime(row.created_at)
  },
  {
    title: t('database.instances.columns.actions'), key: 'actions', width: 280,
    render: (row: any) => {
      const isRunning = row.status === 'running'
      const children: any[] = []
      if (!isRunning) {
        children.push(iconBtn({ text: t('database.instances.start'), icon: PlayOutline, type: 'success', onClick: () => handleStart(row) }))
      } else {
        children.push(iconBtn({ text: t('database.instances.stop'), icon: StopOutline, type: 'warning', onClick: () => handleStop(row) }))
      }
      children.push(iconBtn({ text: t('database.instances.restart'), icon: RefreshOutline, onClick: () => handleRestart(row) }))
      children.push(iconBtn({ text: t('database.instances.manage'), icon: ConstructOutline, type: 'primary', onClick: () => handleManage(row) }))
      children.push(iconBtn({ text: t('database.common.delete'), icon: TrashOutline, type: 'error', onClick: () => confirmDeleteInstance(row) }))
      return btnGroup(children)
    }
  }
])

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
    message.success(t('database.instances.started', { name: row.name }))
    emit('refresh')
  } catch (e: any) {
    message.error(e.message || t('database.instances.startFailed'))
  }
}

async function handleStop(row: any) {
  try {
    await dbApi.stopInstance(row.id)
    message.success(t('database.instances.stopped', { name: row.name }))
    emit('refresh')
  } catch (e: any) {
    message.error(e.message || t('database.instances.stopFailed'))
  }
}

async function handleRestart(row: any) {
  try {
    await dbApi.restartInstance(row.id)
    message.success(t('database.instances.restarted', { name: row.name }))
    emit('refresh')
  } catch (e: any) {
    message.error(e.message || t('database.instances.restartFailed'))
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
    message.success(t('database.instances.createSuccess'))
    createInstanceModal.value = false
    emit('refresh')
  } catch (e: any) {
    message.error(e.message || t('database.instances.createFailed'))
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
      message.info(t('database.instances.discoverNone'))
    }
  } catch (e: any) {
    message.error(e.message || t('database.instances.discoverFailed'))
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
    message.success(t('database.instances.importSuccess'))
    importModal.value = false
    emit('refresh')
  } catch (e: any) {
    message.error(e.message || t('database.instances.importFailed'))
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
    message.success(t('database.instances.deleted', { name: targetInstance.value.name }))
    deleteInstanceConfirm.value = false
    emit('refresh')
  } catch (e: any) {
    message.error(e.message || t('database.instances.deleteFailed'))
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
          <template #icon><NIcon><AddOutline /></NIcon></template>{{ t('database.instances.createInstance') }}
        </NButton>
        <NButton type="info" ghost @click="openImportModal">
          <template #icon><NIcon><CloudDownloadOutline /></NIcon></template>{{ t('database.instances.importExisting') }}
        </NButton>
        <NButton @click="$emit('refresh')" :loading="loading">
          <template #icon><NIcon><RefreshOutline /></NIcon></template>{{ t('database.common.refresh') }}
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
      <NEmpty :description="t('database.instances.empty')">
        <NButton type="primary" size="small" @click="openCreateInstance">{{ t('database.instances.createInstance') }}</NButton>
      </NEmpty>
    </div>

    <!-- 导入已有实例 -->
    <ContentModal v-model:show="importModal" :title="t('database.instances.importModalTitle')" size="sm">
      <NAlert type="info" :show-icon="true" style="margin-bottom: var(--space-4);">
        {{ t('database.instances.importAlert') }}
      </NAlert>

      <NSpace style="margin-bottom: var(--space-4);">
        <NButton size="small" :loading="discoverLoading" @click="handleDiscover">
          <template #icon><NIcon><SearchOutline /></NIcon></template>{{ t('database.instances.discover') }}
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
          <span class="discover-item__ver">{{ d.version || t('database.instances.versionUnknown') }}</span>
          <span class="discover-item__use">{{ t('database.instances.clickToUse') }}</span>
        </div>
      </div>

      <NFormItem :label="t('database.instances.form.type')">
        <NRadioGroup :value="importForm.type" @update:value="onImportTypeChange">
          <NRadio value="mysql">MySQL</NRadio>
          <NRadio value="postgres">PostgreSQL</NRadio>
          <NRadio value="redis">Redis</NRadio>
        </NRadioGroup>
      </NFormItem>
      <NFormItem :label="t('database.instances.form.host')">
        <NInput v-model:value="importForm.host" placeholder="127.0.0.1" />
      </NFormItem>
      <NFormItem :label="t('database.instances.form.port')">
        <NInputNumber v-model:value="importForm.port" :min="1" :max="65535" style="width: 100%;" />
      </NFormItem>
      <NFormItem v-if="importForm.type !== 'redis'" :label="t('database.instances.form.username')">
        <NInput v-model:value="importForm.username" placeholder="root / postgres" />
      </NFormItem>
      <NFormItem v-if="importForm.type !== 'redis'" :label="t('database.instances.form.password')">
        <NInput v-model:value="importForm.password" type="password" show-password-on="click" :placeholder="t('database.instances.form.passwordPlaceholder')" />
      </NFormItem>
      <NFormItem :label="t('database.instances.form.remark')">
        <NInput v-model:value="importForm.remark" :placeholder="t('database.instances.form.optionalPlaceholder')" />
      </NFormItem>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="importModal = false">{{ t('database.common.cancel') }}</NButton>
          <NButton type="primary" :loading="importLoading" @click="submitImport">{{ t('database.instances.import') }}</NButton>
        </NSpace>
      </template>
    </ContentModal>

    <!-- 创建实例 -->
    <ModalForm
      v-model:show="createInstanceModal"
      :title="t('database.instances.createModalTitle')"
      :loading="createInstanceLoading"
      @submit="submitCreateInstance"
    >
      <NFormItem :label="t('database.instances.form.type')">
        <div class="type-cards">
          <div
            v-for="tp in (['mysql','postgres','redis'] as DBType[])"
            :key="tp"
            class="type-card"
            :class="{ 'type-card--active': instanceForm.type === tp }"
            @click="onInstanceTypeChange(tp)"
          >
            <span class="type-card__name">{{ DB_TYPE_LABELS[tp] }}</span>
            <span class="type-card__port">:{{ DB_DEFAULT_PORT[tp] }}</span>
          </div>
        </div>
      </NFormItem>
      <NFormItem :label="t('database.instances.form.version')">
        <NRadioGroup v-model:value="instanceForm.version">
          <NSpace>
            <NRadio v-for="opt in DB_VERSION_OPTIONS[instanceForm.type]" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </NRadio>
          </NSpace>
        </NRadioGroup>
      </NFormItem>
      <NFormItem :label="t('database.instances.form.port')">
        <NInputNumber v-model:value="instanceForm.port" :placeholder="String(DB_DEFAULT_PORT[instanceForm.type])" :min="1" :max="65535" style="width: 100%" />
        <div class="form-hint">{{ t('database.instances.form.defaultPortHint', { port: DB_DEFAULT_PORT[instanceForm.type] }) }}</div>
      </NFormItem>
      <NFormItem v-if="instanceForm.type !== 'redis'" :label="t('database.instances.form.rootPassword')">
        <NInput v-model:value="instanceForm.root_password" type="password" show-password-on="click" :placeholder="t('database.instances.form.rootPasswordPlaceholder')" />
        <div class="form-hint">{{ t('database.instances.form.rootPasswordHint') }}</div>
      </NFormItem>
      <NFormItem :label="t('database.instances.form.remark')">
        <NInput v-model:value="instanceForm.remark" :placeholder="t('database.instances.form.optionalPlaceholder')" />
      </NFormItem>
    </ModalForm>

    <!-- 删除实例确认 -->
    <ConfirmDialog
      v-model:show="deleteInstanceConfirm"
      :title="t('database.instances.deleteModalTitle')"
      :content="t('database.instances.deleteModalContent', { name: targetInstance?.name ?? '' })"
      type="error"
      :confirm-text="t('database.common.delete')"
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
  color: var(--text-quaternary);
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
  transition: box-shadow var(--t-fast) var(--ease), transform var(--t-fast) var(--ease), border-color var(--t-fast) var(--ease);
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
  color: var(--text-quaternary);
  font-family: var(--font-mono);
}

.discover-list {
  margin-bottom: var(--space-4);
  border: 1px solid var(--line);
  border-radius: 8px;
  overflow: hidden;
}

.discover-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
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
  color: var(--text-quaternary);
  font-family: var(--font-mono);
  font-size: 12px;
}

.discover-item__use {
  color: var(--acc);
  font-size: 12px;
}
</style>
