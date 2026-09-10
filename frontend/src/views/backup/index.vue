<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue'
import {
  NButton, NIcon, NForm, NFormItem, NInput, NSelect, NRadioGroup,
  NRadio, NSpace, NTag, useMessage
} from 'naive-ui'
import {
  AddOutline, DownloadOutline, RefreshOutline, TrashOutline,
  SaveOutline, CloudUploadOutline
} from '@vicons/ionicons5'
import {
  PageCard, DataTable, SearchForm, ModalForm, StatusTag, ConfirmDialog
} from '@/components'
import type { StatusType } from '@/components'
import {
  getBackups, createBackup, deleteBackup, restoreBackup, downloadBackup,
  type Backup, type BackupType, type BackupStatus, type BackupListParams, type CreateBackupData
} from '@/api/backup'
import { getSites, type Site } from '@/api/site'
import { getInstances, type DBInstance } from '@/api/database'
import type { DataTableColumns } from 'naive-ui'

const message = useMessage()

/* ========== 常量映射 ========== */
const typeLabelMap: Record<BackupType, string> = {
  site: '网站',
  database: '数据库',
  system: '系统'
}

const typeTagTypeMap: Record<BackupType, 'info' | 'success' | 'warning'> = {
  site: 'info',
  database: 'success',
  system: 'warning'
}

const statusMap: Record<BackupStatus, { status: StatusType; text: string }> = {
  completed: { status: 'running', text: '已完成' },
  running: { status: 'warning', text: '进行中' },
  failed: { status: 'error', text: '失败' }
}

const typeOptions = [
  { label: '网站', value: 'site' },
  { label: '数据库', value: 'database' },
  { label: '系统', value: 'system' }
]

/* ========== 工具函数 ========== */
function formatFileSize(bytes: number): string {
  if (!bytes || bytes <= 0) return '—'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

/* ========== 列表状态 ========== */
const loading = ref(false)
const backups = ref<Backup[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const searchKeyword = ref('')
const filterType = ref<BackupType | null>(null)

/* ========== Mock 数据兜底 ========== */
function generateMockBackups(): { items: Backup[]; total: number } {
  const mock: Backup[] = [
    { id: 1, type: 'site', target: 'example.com', file_name: 'example.com_20260908.tar.gz', size: 52428800, status: 'completed', remark: '每日自动备份', created_at: '2026-09-08 03:00:00' },
    { id: 2, type: 'database', target: 'mysql_prod', file_name: 'mysql_prod_20260908.sql.gz', size: 10485760, status: 'completed', remark: '生产数据库', created_at: '2026-09-08 03:10:00' },
    { id: 3, type: 'system', target: 'full-system', file_name: 'system_20260907.tar.gz', size: 1073741824, status: 'completed', remark: '每周全量', created_at: '2026-09-07 02:00:00' },
    { id: 4, type: 'site', target: 'blog.example.com', file_name: 'blog_20260908.tar.gz', size: 0, status: 'running', remark: '手动备份', created_at: '2026-09-08 14:30:00' },
    { id: 5, type: 'database', target: 'redis_cache', file_name: 'redis_20260906.rdb', size: 2097152, status: 'failed', remark: '', created_at: '2026-09-06 03:00:00' }
  ]
  return { items: mock, total: mock.length }
}

async function loadBackups() {
  loading.value = true
  try {
    const params: BackupListParams = {
      page: page.value,
      pageSize: pageSize.value
    }
    if (filterType.value) params.type = filterType.value
    if (searchKeyword.value) params.keyword = searchKeyword.value

    const res = await getBackups(params)
    if (res.data?.items) {
      backups.value = res.data.items
      total.value = res.data.total
    } else {
      const mock = generateMockBackups()
      backups.value = mock.items
      total.value = mock.total
    }
  } catch {
    const mock = generateMockBackups()
    backups.value = mock.items
    total.value = mock.total
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
  filterType.value = null
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
    title: '备份类型',
    key: 'type',
    width: 110,
    render: (row: any) => h(NTag, {
      size: 'small',
      type: typeTagTypeMap[row.type as BackupType],
      bordered: false,
      round: true
    }, { default: () => typeLabelMap[row.type as BackupType] })
  },
  {
    title: '目标',
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
    title: '文件大小',
    key: 'size',
    width: 120,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub)' }, row.formatted_size || formatFileSize(row.size))
  },
  {
    title: '状态',
    key: 'status',
    width: 110,
    render: (row: any) => {
      const cfg = statusMap[row.status as BackupStatus] || { status: 'stopped', text: row.status || '未知' }
      return h(StatusTag, { status: cfg.status, text: cfg.text })
    }
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 170,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.created_at)
  },
  {
    title: '操作',
    key: 'actions',
    width: 240,
    render: (row: any) => {
      const canRestore = row.status === 'completed'
      const canDownload = row.status === 'completed'
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
            default: () => '恢复'
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            disabled: !canDownload,
            loading: actionLoading,
            onClick: () => handleDownload(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(DownloadOutline) }),
            default: () => '下载'
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            type: 'error',
            loading: actionLoading,
            onClick: () => openDelete(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
            default: () => '删除'
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

const createRules = {
  target: {
    required: true,
    message: '请选择或输入备份目标',
    trigger: 'blur'
  }
}

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
    message.warning('请输入备份目标')
    return
  }

  createLoading.value = true
  try {
    await createBackup({
      type: createForm.type,
      target,
      remark: createForm.remark
    })
    message.success('备份任务已创建，正在进行中...')
    showCreate.value = false
    loadBackups()
  } catch (err: any) {
    message.error(err.message || '创建备份失败')
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
    message.success('恢复任务已启动，恢复前已自动创建快照')
    showRestore.value = false
    loadBackups()
  } catch (err: any) {
    message.error(err.message || '恢复失败')
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
    message.success('开始下载')
  } catch (err: any) {
    message.error(err.message || '下载失败')
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
    message.success('备份已删除')
    showDelete.value = false
    loadBackups()
  } catch (err: any) {
    message.error(err.message || '删除失败')
  } finally {
    deleteLoading.value = false
  }
}

onMounted(() => {
  loadBackups()
})
</script>

<template>
  <PageCard title="备份恢复" subtitle="网站 / 数据库 / 系统数据备份与一键恢复">
    <template #extra>
      <NButton type="primary" @click="openCreate">
        <template #icon>
          <NIcon><AddOutline /></NIcon>
        </template>
        创建备份
      </NButton>
    </template>

    <!-- 搜索筛选 -->
    <SearchForm @search="handleSearch" @reset="handleReset">
      <NFormItem label="类型">
        <NSelect
          v-model:value="filterType"
          :options="typeOptions as any"
          placeholder="全部类型"
          clearable
          style="width: var(--input-md)"
          @update:value="() => { page = 1; loadBackups() }"
        />
      </NFormItem>
      <NFormItem label="搜索" path="keyword">
        <NInput
          v-model:value="searchKeyword"
          name="keyword"
          placeholder="输入目标/备注搜索"
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
    />
  </PageCard>

  <!-- 创建备份对话框 -->
  <ModalForm
    :show="showCreate"
    title="创建备份"
    :loading="createLoading"
    :width="560"
    @update:show="(v: boolean) => (showCreate = v)"
    @submit="handleCreateSubmit"
  >
    <NForm ref="createFormRef" :model="createForm" :rules="createRules" label-placement="top">
      <NFormItem label="备份类型" path="type">
        <NRadioGroup v-model:value="createForm.type">
          <NSpace size="medium">
            <NRadio value="site">网站</NRadio>
            <NRadio value="database">数据库</NRadio>
            <NRadio value="system">系统</NRadio>
          </NSpace>
        </NRadioGroup>
      </NFormItem>

      <NFormItem v-if="createForm.type === 'site'" label="选择站点" path="target">
        <NSelect
          v-model:value="createForm.target"
          :options="siteOptions"
          placeholder="选择要备份的站点"
          filterable
          clearable
        />
      </NFormItem>

      <NFormItem v-if="createForm.type === 'database'" label="选择数据库" path="target">
        <NSelect
          v-model:value="createForm.target"
          :options="dbOptions"
          placeholder="选择要备份的数据库实例"
          filterable
          clearable
        />
      </NFormItem>

      <NFormItem v-if="createForm.type === 'system'" label="系统备份">
        <div class="system-backup-hint">
          <NIcon><CloudUploadOutline /></NIcon>
          <span>将对整个系统进行全量备份，包括配置、应用数据与系统文件</span>
        </div>
      </NFormItem>

      <NFormItem label="备注" path="remark">
        <NInput
          v-model:value="createForm.remark"
          placeholder="可选备注信息"
          clearable
        />
      </NFormItem>
    </NForm>
  </ModalForm>

  <!-- 恢复确认 -->
  <ConfirmDialog
    :show="showRestore"
    title="恢复备份"
    :content="`确定要恢复备份「${restoreTarget}」吗？恢复将覆盖当前数据，系统会在恢复前自动创建快照。`"
    confirm-text="确认恢复"
    type="warning"
    :loading="restoreLoading"
    @update:show="(v: boolean) => (showRestore = v)"
    @confirm="handleRestoreConfirm"
  />

  <!-- 删除确认 -->
  <ConfirmDialog
    :show="showDelete"
    title="删除备份"
    :content="`确定要删除备份「${deleteTarget}」吗？此操作不可撤销。`"
    confirm-text="删除"
    type="error"
    :loading="deleteLoading"
    @update:show="(v: boolean) => (showDelete = v)"
    @confirm="handleDeleteConfirm"
  />
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
  color: var(--faint);
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
