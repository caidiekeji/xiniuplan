<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NButton, NIcon, NForm, NFormItem, NInput, NSelect, NRadioGroup,
  NRadio, NSpace, NTag, NEmpty
} from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import { AddOutline, DownloadOutline, RefreshOutline, TrashOutline, CloudUploadOutline } from '@vicons/ionicons5'
import {
  PageCard, DataTable, SearchForm, ModalForm, ContentModal, StatusTag, ConfirmDialog
} from '@/components'
import type { StatusType } from '@/components'
import {
  getBackups, createBackup, deleteBackup, restoreBackup, downloadBackup,
  type Backup, type BackupType, type BackupStatus, type BackupListParams, type CreateBackupData
} from '@/api/backup'
import { getSites, type Site } from '@/api/site'
import { getInstances, type DBInstance } from '@/api/database'
import type { DataTableColumns } from 'naive-ui'
import { useQueryState } from '@/composables/useQueryState'

const message = useMessage()
const { t } = useI18n()

/* ========== 常量映射 ========== */
function typeLabel(type: BackupType): string {
  return t('backup.types.' + type)
}

const typeTagTypeMap: Record<BackupType, 'info' | 'success' | 'warning'> = {
  site: 'info',
  database: 'success',
  system: 'warning',
  app: 'warning'
}

const statusMap = computed<Record<BackupStatus, { status: StatusType; text: string }>>(() => ({
  // 状态集对齐后端 model.Backup：pending | running | success | failed
  pending: { status: 'pending', text: t('backup.status.pending') },
  success: { status: 'running', text: t('backup.status.completed') },
  running: { status: 'warning', text: t('backup.status.running') },
  failed: { status: 'error', text: t('backup.status.failed') }
}))

const typeOptions = computed(() => [
  { label: t('backup.types.site'), value: 'site' },
  { label: t('backup.types.database'), value: 'database' },
  { label: t('backup.types.system'), value: 'system' },
  { label: t('backup.types.app'), value: 'app' }
])

/* ========== 工具函数 ========== */
function formatFileSize(bytes: number): string {
  if (!bytes || bytes <= 0) return '—'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

/* ========== 列表状态 ========== */
// 分页 / 筛选 URL 化（D19）：刷新与深链可还原现场
const loading = ref(false)
const backups = ref<Backup[]>([])
const total = ref(0)
const page = useQueryState('page', 1)
const pageSize = ref(20)
const searchKeyword = useQueryState('kw', '')
const filterType = useQueryState<BackupType | ''>('type', '')

/* ========== Mock 数据兜底 ========== */
async function loadBackups() {
  loading.value = true
  try {
    const params: BackupListParams = {
      page: page.value,
      pageSize: pageSize.value
    }
    if (filterType.value) params.type = filterType.value as BackupType
    if (searchKeyword.value) params.keyword = searchKeyword.value

    const res = await getBackups(params)
    if (res.data?.items) {
      backups.value = res.data.items
      total.value = res.data.total
    } else {
      backups.value = []
      total.value = 0
    }
  } catch (e: any) {
    backups.value = []
    total.value = 0
    message.error(e.message || t('backup.messages.loadFailed'))
  } finally {
    loading.value = false
  }
}

function handleSearch(values: Record<string, any>) {
  searchKeyword.value = values.keyword || ''
  page.value = 1
  loadBackups()
}

function handleReset() {
  searchKeyword.value = ''
  filterType.value = ''
  page.value = 1
  loadBackups()
}

const pagination = computed(() => ({
  page: page.value,
  pageSize: pageSize.value,
  itemCount: total.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  'onUpdate:page': (p: number) => { page.value = p; loadBackups() },
  'onUpdate:page-size': (ps: number) => { pageSize.value = ps; page.value = 1; loadBackups() }
}))

/* ========== 表格列 ========== */
const columns = computed<DataTableColumns>(() => [
  {
    title: t('backup.columns.type'),
    key: 'type',
    width: 110,
    render: (row: any) => h(NTag, {
      size: 'small',
      type: typeTagTypeMap[row.type as BackupType],
      bordered: false,
      round: true
    }, { default: () => typeLabel(row.type as BackupType) })
  },
  {
    title: t('backup.columns.target'),
    key: 'target',
    width: 220,
    render: (row: any) => h('div', { class: 'backup-target' }, [
      h('span', { class: 'backup-target__main mono' }, row.target),
      row.remark
        ? h('span', { class: 'backup-target__remark' }, row.remark)
        : null
    ])
  },
  {
    title: t('backup.columns.size'),
    key: 'size',
    width: 120,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub)' }, row.formatted_size || formatFileSize(row.size))
  },
  {
    title: t('backup.columns.status'),
    key: 'status',
    width: 110,
    render: (row: any) => {
      const cfg = statusMap.value[row.status as BackupStatus] || { status: 'stopped', text: row.status || t('backup.unknown') }
      return h(StatusTag, { status: cfg.status, text: cfg.text })
    }
  },
  {
    title: t('backup.columns.createdAt'),
    key: 'created_at',
    width: 170,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.created_at)
  },
  {
    title: t('backup.columns.actions'),
    key: 'actions',
    width: 240,
    render: (row: any) => {
      // 与后端状态集一致（success 而非 completed）
      const canRestore = row.status === 'success'
      const canDownload = row.status === 'success'
      const actionLoading = actionLoadingId.value === row.id
      return h(NSpace, { size: 'small', wrap: false }, {
        default: () => [
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            type: 'warning',
            disabled: !canRestore,
            loading: actionLoading,
            onClick: () => openRestore(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(RefreshOutline) }),
            default: () => t('backup.actions.restore')
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            disabled: !canDownload,
            loading: actionLoading,
            onClick: () => handleDownload(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(DownloadOutline) }),
            default: () => t('backup.actions.download')
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            type: 'info',
            disabled: !canDownload,
            loading: actionLoading,
            onClick: () => openPush(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(CloudUploadOutline) }),
            default: () => t('backup.actions.push')
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            type: 'error',
            loading: actionLoading,
            onClick: () => openDelete(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
            default: () => t('backup.actions.delete')
          })
        ]
      })
    }
  }
])

/* ========== 操作 loading ========== */
const actionLoadingId = ref<number | null>(null)

/* ========== 创建备份对话框 ========== */
const showCreate = ref(false)
const createLoading = ref(false)
const createFormRef = ref()
const createForm = reactive<CreateBackupData & { targetInput?: string }>({
  type: 'site',
  target: '',
  remark: '',
  targetInput: ''
})

const siteOptions = ref<{ label: string; value: string }[]>([])
const dbOptions = ref<{ label: string; value: string }[]>([])

async function loadTargetOptions() {
  try {
    const [sitesRes, dbsRes] = await Promise.all([
      getSites({ page: 1, pageSize: 100 }),
      getInstances({ page: 1, pageSize: 100 })
    ])
    const sites: Site[] = sitesRes.data?.items || []
    const dbs: DBInstance[] = dbsRes.data?.items || []
    siteOptions.value = sites.map((s) => ({ label: s.domain, value: s.domain }))
    dbOptions.value = dbs.map((d) => ({ label: `${d.name || d.type}_${d.id}`, value: d.name || `${d.type}_${d.id}` }))
  } catch {
    siteOptions.value = [
      { label: 'example.com', value: 'example.com' },
      { label: 'blog.example.com', value: 'blog.example.com' }
    ]
    dbOptions.value = [
      { label: 'mysql_prod', value: 'mysql_prod' },
      { label: 'redis_cache', value: 'redis_cache' }
    ]
  }
}

const createRules = computed(() => ({
  target: {
    required: true,
    message: t('backup.rules.targetRequired'),
    trigger: 'blur'
  }
}))

function openCreate() {
  createForm.type = 'site'
  createForm.target = ''
  createForm.targetInput = ''
  createForm.remark = ''
  loadTargetOptions()
  showCreate.value = true
}

async function handleCreateSubmit() {
  try {
    await createFormRef.value?.validate()
  } catch {
    return
  }

  const target = createForm.type === 'system' ? 'full-system' : (createForm.target || createForm.targetInput || '')
  if (!target) {
    message.warning(t('backup.messages.enterTarget'))
    return
  }

  createLoading.value = true
  try {
    await createBackup({
      type: createForm.type,
      target,
      remark: createForm.remark
    })
    message.success(t('backup.messages.createSuccess'))
    showCreate.value = false
    loadBackups()
  } catch (err: any) {
    message.error(err.message || t('backup.messages.createFailed'))
  } finally {
    createLoading.value = false
  }
}

/* ========== 恢复确认 ========== */
const showRestore = ref(false)
const restoreLoading = ref(false)
const restoreId = ref<number | null>(null)
const restoreTarget = ref('')

function openRestore(id: number) {
  restoreId.value = id
  const item = backups.value.find((b) => b.id === id)
  restoreTarget.value = item?.target || ''
  showRestore.value = true
}

async function handleRestoreConfirm() {
  if (restoreId.value === null) return
  restoreLoading.value = true
  try {
    await restoreBackup(restoreId.value)
    message.success(t('backup.messages.restoreSuccess'))
    showRestore.value = false
    loadBackups()
  } catch (err: any) {
    message.error(err.message || t('backup.messages.restoreFailed'))
  } finally {
    restoreLoading.value = false
  }
}

/* ========== 下载 ========== */
async function handleDownload(id: number) {
  actionLoadingId.value = id
  try {
    const res = await downloadBackup(id)
    const blob = res.data as Blob
    const item = backups.value.find((b) => b.id === id)
    const fileName = item?.file_name || `backup_${id}.tar.gz`
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = fileName
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    message.success(t('backup.messages.downloadStart'))
  } catch (err: any) {
    message.error(err.message || t('backup.messages.downloadFailed'))
  } finally {
    actionLoadingId.value = null
  }
}

/* ========== 删除确认 ========== */
const showDelete = ref(false)
const deleteLoading = ref(false)
const deleteId = ref<number | null>(null)
const deleteTarget = ref('')

function openDelete(id: number) {
  deleteId.value = id
  const item = backups.value.find((b) => b.id === id)
  deleteTarget.value = item?.target || ''
  showDelete.value = true
}

async function handleDeleteConfirm() {
  if (deleteId.value === null) return
  deleteLoading.value = true
  try {
    await deleteBackup(deleteId.value)
    message.success(t('backup.messages.deleteSuccess'))
    showDelete.value = false
    loadBackups()
  } catch (err: any) {
    message.error(err.message || t('backup.messages.deleteFailed'))
  } finally {
    deleteLoading.value = false
  }
}

onMounted(() => {
  loadBackups()
})

/* ========== 远程推送与远程存储管理（BK-003） ========== */
import {
  listRemoteStorages, createRemoteStorage, deleteRemoteStorage, testRemoteStorage, pushBackup,
  type RemoteStorage
} from '@/api/newfeatures'

const showPush = ref(false)
const pushBackupId = ref<number | null>(null)
const pushRemoteId = ref<number | null>(null)
const remotes = ref<RemoteStorage[]>([])
const pushLoading = ref(false)
const showRemoteMgr = ref(false)
const showRemoteCreate = ref(false)
const remoteLoading = ref(false)
const newRemote = reactive({ name: '', type: 's3' as RemoteStorage['type'], endpoint: '', bucket: '', region: '', access_key: '', secret_key: '', url: '', username: '', password: '', token: '' })
const creatingRemote = ref(false)

const remoteTypeOptions = computed(() => [
  { label: t('backup.remoteCreate.types.s3'), value: 's3' },
  { label: t('backup.remoteCreate.types.webdav'), value: 'webdav' },
  { label: t('backup.remoteCreate.types.sftp'), value: 'sftp' },
  { label: t('backup.remoteCreate.types.panel'), value: 'panel' }
])

async function loadRemotes() {
  remoteLoading.value = true
  try {
    remotes.value = await listRemoteStorages()
  } catch (e: any) {
    message.error(e.message || t('backup.messages.loadRemoteFailed'))
  } finally {
    remoteLoading.value = false
  }
}

function openPush(backupId: number) {
  pushBackupId.value = backupId
  pushRemoteId.value = null
  loadRemotes()
  showPush.value = true
}

async function doPush() {
  if (!pushRemoteId.value || !pushBackupId.value) {
    message.warning(t('backup.messages.selectRemote'))
    return
  }
  pushLoading.value = true
  try {
    await pushBackup(pushBackupId.value, pushRemoteId.value)
    message.success(t('backup.messages.pushSuccess'))
    showPush.value = false
    loadBackups()
  } catch (e: any) {
    message.error(e.message || t('backup.messages.pushFailed'))
  } finally {
    pushLoading.value = false
  }
}

async function doCreateRemote() {
  if (!newRemote.name.trim()) {
    message.warning(t('backup.messages.enterName'))
    return
  }
  creatingRemote.value = true
  try {
    const config: Record<string, string> = {}
    if (newRemote.type === 's3') {
      Object.assign(config, { endpoint: newRemote.endpoint, bucket: newRemote.bucket, region: newRemote.region, access_key: newRemote.access_key, secret_key: newRemote.secret_key, path_style: 'true' })
    } else if (newRemote.type === 'webdav') {
      Object.assign(config, { url: newRemote.url, username: newRemote.username, password: newRemote.password })
    } else if (newRemote.type === 'sftp') {
      Object.assign(config, { host: newRemote.url, port: '22', username: newRemote.username, password: newRemote.password })
    } else if (newRemote.type === 'panel') {
      Object.assign(config, { url: newRemote.url, token: newRemote.token })
    }
    await createRemoteStorage({ name: newRemote.name.trim(), type: newRemote.type, config, enabled: true })
    message.success(t('backup.messages.remoteCreateSuccess'))
    showRemoteCreate.value = false
    loadRemotes()
  } catch (e: any) {
    message.error(e.message || t('backup.messages.remoteCreateFailed'))
  } finally {
    creatingRemote.value = false
  }
}
</script>

<template>
  <PageCard :title="t('backup.title')" :subtitle="t('backup.subtitle')">
    <template #extra>
      <NButton type="primary" @click="openCreate">
        <template #icon>
          <NIcon><AddOutline /></NIcon>
        </template>
        {{ t('backup.createBackup') }}
        <NButton size="small" quaternary @click="showRemoteMgr = true; loadRemotes()">{{ t('backup.remoteStorage') }}</NButton>
      </NButton>
    </template>

    <!-- 搜索筛选 -->
    <SearchForm @search="handleSearch" @reset="handleReset">
      <NFormItem :label="t('backup.search.type')">
        <NSelect
          v-model:value="filterType"
          :options="typeOptions as any"
          :placeholder="t('backup.search.allTypes')"
          clearable
          style="width: var(--input-md)"
          @update:value="() => { page = 1; loadBackups() }"
        />
      </NFormItem>
      <NFormItem :label="t('backup.search.keyword')" path="keyword">
        <NInput
          v-model:value="searchKeyword"
          name="keyword"
          :placeholder="t('backup.search.keywordPlaceholder')"
          clearable
          style="width: var(--input-lg)"
        />
      </NFormItem>
    </SearchForm>

    <!-- 数据表格 -->
    <DataTable
      :columns="columns"
      :data="backups"
      :loading="loading"
      :pagination="pagination"
    >
      <!-- 空态引导（词表 D15）：无筛选条件的首次空态给创建入口；筛选空结果只显示「暂无数据」 -->
      <template #empty>
        <NEmpty :description="t('common.noData')">
          <template v-if="!searchKeyword && !filterType" #extra>
            <NButton size="small" @click="openCreate">{{ t('common.create') }}</NButton>
          </template>
        </NEmpty>
      </template>
    </DataTable>
  </PageCard>

  <!-- 创建备份对话框 -->
  <ModalForm
    :show="showCreate"
    :title="t('backup.create.title')"
    :loading="createLoading"
    :width="560"
    @update:show="(v: boolean) => (showCreate = v)"
    @submit="handleCreateSubmit"
  >
    <NForm ref="createFormRef" :model="createForm" :rules="createRules" label-placement="top">
      <NFormItem :label="t('backup.create.typeLabel')" path="type">
        <NRadioGroup v-model:value="createForm.type">
          <NSpace size="medium">
            <NRadio value="site">{{ t('backup.types.site') }}</NRadio>
            <NRadio value="database">{{ t('backup.types.database') }}</NRadio>
            <NRadio value="system">{{ t('backup.types.system') }}</NRadio>
          </NSpace>
        </NRadioGroup>
      </NFormItem>

      <NFormItem v-if="createForm.type === 'site'" :label="t('backup.create.siteTarget')" path="target">
        <NSelect
          v-model:value="createForm.target"
          :options="siteOptions"
          :placeholder="t('backup.create.sitePlaceholder')"
          filterable
          clearable
        />
      </NFormItem>

      <NFormItem v-if="createForm.type === 'database'" :label="t('backup.create.dbTarget')" path="target">
        <NSelect
          v-model:value="createForm.target"
          :options="dbOptions"
          :placeholder="t('backup.create.dbPlaceholder')"
          filterable
          clearable
        />
      </NFormItem>

      <NFormItem v-if="createForm.type === 'system'" :label="t('backup.create.systemLabel')">
        <div class="system-backup-hint">
          <NIcon><CloudUploadOutline /></NIcon>
          <span>{{ t('backup.create.systemHint') }}</span>
        </div>
      </NFormItem>

      <NFormItem :label="t('backup.create.remark')" path="remark">
        <NInput
          v-model:value="createForm.remark"
          :placeholder="t('backup.create.remarkPlaceholder')"
          clearable
        />
      </NFormItem>
    </NForm>
  </ModalForm>

  <!-- 恢复确认 -->
  <ConfirmDialog
    :show="showRestore"
    :title="t('backup.restore.title')"
    :content="t('backup.restore.content', { name: restoreTarget })"
    :confirm-text="t('backup.restore.confirm')"
    type="warning"
    :loading="restoreLoading"
    @update:show="(v: boolean) => (showRestore = v)"
    @confirm="handleRestoreConfirm"
  />

  <!-- 删除确认 -->
  <ConfirmDialog
    :show="showDelete"
    :title="t('backup.remove.title')"
    :content="t('backup.remove.content', { name: deleteTarget })"
    :confirm-text="t('backup.remove.confirm')"
    type="error"
    :loading="deleteLoading"
    @update:show="(v: boolean) => (showDelete = v)"
    @confirm="handleDeleteConfirm"
  />

    <!-- 推送到远程存储 -->
    <ModalForm
      v-model:show="showPush"
      :title="t('backup.push.title')"
      :width="460"
      :loading="pushLoading"
      :confirm-text="t('backup.push.confirm')"
      :cancel-text="t('backup.push.cancel')"
      @submit="doPush"
    >
      <NSelect v-model:value="pushRemoteId" :options="remotes.map((r: RemoteStorage) => ({ label: r.name + ' (' + r.type + ')', value: r.id }))" :placeholder="t('backup.push.placeholder')" />
    </ModalForm>

    <!-- 远程存储管理 -->
    <ContentModal v-model:show="showRemoteMgr" :title="t('backup.remote.title')" size="lg">
      <div v-for="r in remotes" :key="r.id" style="display:flex;align-items:center;gap:var(--space-3);padding:var(--space-2) 0;border-bottom:1px dashed var(--line)">
        <NTag size="small">{{ r.type }}</NTag>
        <span style="flex:1">{{ r.name }}</span>
        <NTag size="tiny" :type="r.enabled ? 'success' : 'default'">{{ r.enabled ? t('backup.remote.enabled') : t('backup.remote.disabled') }}</NTag>
        <NButton size="tiny" quaternary @click="testRemoteStorage(r.id).then(() => message.success(t('backup.messages.testSuccess'))).catch((e:any) => message.error(e.message || t('backup.messages.testFailed')))">{{ t('backup.remote.test') }}</NButton>
        <NButton size="tiny" quaternary type="error" @click="deleteRemoteStorage(r.id).then(loadRemotes)">{{ t('backup.remote.delete') }}</NButton>
      </div>
      <NSpace style="margin-top:var(--space-3)">
        <NButton size="small" @click="showRemoteMgr = false">{{ t('backup.remote.close') }}</NButton>
        <NButton size="small" type="primary" @click="showRemoteCreate = true">{{ t('backup.remote.create') }}</NButton>
      </NSpace>
    </ContentModal>

    <!-- 新建远程存储 -->
    <ModalForm
      v-model:show="showRemoteCreate"
      :title="t('backup.remoteCreate.title')"
      :width="520"
      :loading="creatingRemote"
      :confirm-text="t('backup.remoteCreate.create')"
      :cancel-text="t('backup.remoteCreate.cancel')"
      @submit="doCreateRemote"
    >
      <NForm label-placement="left" label-width="110">
        <NFormItem :label="t('backup.remoteCreate.name')"><NInput v-model:value="newRemote.name" /></NFormItem>
        <NFormItem :label="t('backup.remoteCreate.type')">
          <NSelect v-model:value="newRemote.type" :options="remoteTypeOptions" />
        </NFormItem>
        <NFormItem v-if="newRemote.type === 's3'" label="Endpoint"><NInput v-model:value="newRemote.endpoint" placeholder="https://s3.example.com" /></NFormItem>
        <NFormItem v-if="newRemote.type === 's3'" label="Bucket"><NInput v-model:value="newRemote.bucket" /></NFormItem>
        <NFormItem v-if="newRemote.type === 's3'" label="Region"><NInput v-model:value="newRemote.region" placeholder="us-east-1" /></NFormItem>
        <NFormItem v-if="newRemote.type !== 's3'" :label="t('backup.remoteCreate.address')"><NInput v-model:value="newRemote.url" /></NFormItem>
        <NFormItem v-if="newRemote.type === 's3'" label="Access Key"><NInput v-model:value="newRemote.access_key" /></NFormItem>
        <NFormItem v-if="newRemote.type === 's3'" label="Secret Key"><NInput v-model:value="newRemote.secret_key" type="password" /></NFormItem>
        <NFormItem v-if="newRemote.type === 'webdav' || newRemote.type === 'sftp'" :label="t('backup.remoteCreate.username')"><NInput v-model:value="newRemote.username" /></NFormItem>
        <NFormItem v-if="newRemote.type !== 'panel'" :label="t('backup.remoteCreate.password')"><NInput v-model:value="newRemote.password" type="password" /></NFormItem>
        <NFormItem v-if="newRemote.type === 'panel'" label="Token"><NInput v-model:value="newRemote.token" /></NFormItem>
      </NForm>
    </ModalForm>
</template>

<style scoped>
.backup-target {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.backup-target__main {
  font-weight: 500;
  color: var(--ink);
}

.backup-target__remark {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
}

.system-backup-hint {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-4);
  background: var(--surface2);
  border-radius: var(--radius-md);
  color: var(--sub);
  font-size: var(--fs-caption);
}

.system-backup-hint .n-icon {
  color: var(--acc);
  flex-shrink: 0;
}
</style>
