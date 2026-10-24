<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NButton, NIcon, NForm, NFormItem, NInput, NInputNumber, NSelect, NSpace,
  NTag, NModal
} from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import { AddOutline, PlayOutline, CreateOutline, TrashOutline, DocumentTextOutline } from '@vicons/ionicons5'
import { PageCard, DataTable, ModalForm, ContentModal, StatusTag, ConfirmDialog } from '@/components'
import type { StatusType } from '@/components'
import {
  listDeployApps, createDeployApp, updateDeployApp, deleteDeployApp, runDeploy, listDeployLogs,
  type DeployApp, type DeployLog
} from '@/api/features17'
import type { DataTableColumns } from 'naive-ui'

const { t } = useI18n()
const message = useMessage()

const buildTypeOptions = computed(() => [
  { label: t('deploy.buildTypes.static'), value: 'static' },
  { label: t('deploy.buildTypes.node'), value: 'node' },
  { label: t('deploy.buildTypes.dockerfile'), value: 'dockerfile' }
])
const buildTypeLabel = computed<Record<string, string>>(() => ({
  static: t('deploy.buildTypes.static'),
  node: t('deploy.buildTypes.node'),
  dockerfile: t('deploy.buildTypes.dockerfile')
}))

const statusMap = computed<Record<string, { status: StatusType; text: string }>>(() => ({
  success: { status: 'running', text: t('deploy.status.success') },
  failed: { status: 'error', text: t('deploy.status.failed') },
  deploying: { status: 'warning', text: t('deploy.status.deploying') }
}))

function fmtTime(v: string | null): string {
  if (!v) return '—'
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return v
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

/* ==================== 列表 ==================== */
const loading = ref(false)
const apps = ref<DeployApp[]>([])
const actionLoadingId = ref<number | null>(null)

async function load() {
  loading.value = true
  try {
    const res = await listDeployApps()
    apps.value = res.items || []
  } catch {
    apps.value = []
  } finally {
    loading.value = false
  }
}

const columns = computed<DataTableColumns>(() => [
  { title: t('deploy.columns.name'), key: 'name', width: 150, render: (row: any) => h('span', { class: 'app-name' }, row.name) },
  { title: t('deploy.columns.repo'), key: 'repo_url', ellipsis: { tooltip: true }, render: (row: any) => h('span', { class: 'mono' }, row.repo_url) },
  { title: t('deploy.columns.branch'), key: 'branch', width: 90, render: (row: any) => h(NTag, { size: 'small', bordered: false, round: true }, { default: () => row.branch }) },
  {
    title: t('deploy.columns.buildType'), key: 'build_type', width: 110,
    render: (row: any) => h(NTag, { size: 'small', type: 'info', bordered: false, round: true }, { default: () => buildTypeLabel.value[row.build_type] || row.build_type })
  },
  { title: t('deploy.columns.destPath'), key: 'dest_path', width: 180, ellipsis: { tooltip: true }, render: (row: any) => h('span', { class: 'mono faint' }, row.dest_path) },
  {
    title: t('deploy.columns.lastDeploy'), key: 'last_deploy_at', width: 150,
    render: (row: any) => {
      if (!row.last_status) return h('span', { class: 'faint' }, t('deploy.notDeployed'))
      const cfg = statusMap.value[row.last_status]
      return h(NSpace, { size: 'small', wrap: false, align: 'center' }, {
        default: () => [
          cfg ? h(StatusTag, { status: cfg.status, text: cfg.text }) : null,
          h('span', { class: 'mono faint' }, fmtTime(row.last_deploy_at))
        ]
      })
    }
  },
  {
    title: t('deploy.columns.actions'), key: 'actions', width: 240,
    render: (row: any) => {
      const busy = actionLoadingId.value === row.id
      return h(NSpace, { size: 'small', wrap: false }, {
        default: () => [
          h(NButton, {
            size: 'tiny', quaternary: true, type: 'primary', loading: busy,
            onClick: () => handleDeploy(row)
          }, { icon: () => h(NIcon, null, { default: () => h(PlayOutline) }), default: () => t('deploy.actions.deploy') }),
          h(NButton, {
            size: 'tiny', quaternary: true,
            onClick: () => openLogs(row)
          }, { icon: () => h(NIcon, null, { default: () => h(DocumentTextOutline) }), default: () => t('deploy.actions.logs') }),
          h(NButton, {
            size: 'tiny', quaternary: true,
            onClick: () => openEdit(row)
          }, { icon: () => h(NIcon, null, { default: () => h(CreateOutline) }), default: () => t('deploy.actions.edit') }),
          h(NButton, {
            size: 'tiny', quaternary: true, type: 'error', loading: busy,
            onClick: () => openDelete(row)
          }, { icon: () => h(NIcon, null, { default: () => h(TrashOutline) }), default: () => t('deploy.actions.delete') })
        ]
      })
    }
  }
])

async function handleDeploy(row: DeployApp) {
  actionLoadingId.value = row.id
  try {
    await runDeploy(row.id)
    message.success(t('deploy.messages.deployDone', { name: row.name }))
    load()
  } catch (err: any) {
    message.error(err.message || t('deploy.messages.deployFailed'))
    load()
  } finally {
    actionLoadingId.value = null
  }
}

/* ==================== 创建/编辑 ==================== */
const showForm = ref(false)
const formLoading = ref(false)
const formRef = ref()
const isEdit = ref(false)
const editId = ref<number | null>(null)

const form = ref({
  name: '',
  repo_url: '',
  branch: 'main',
  secret: '',
  build_type: 'static' as 'static' | 'node' | 'dockerfile',
  out_dir: 'dist',
  dest_path: '',
  container_port: 8080
})

const formRules = computed(() => ({
  name: { required: true, message: t('deploy.rules.nameRequired'), trigger: 'blur' },
  repo_url: { required: true, message: t('deploy.rules.repoRequired'), trigger: 'blur' },
  dest_path: { required: true, message: t('deploy.rules.destPathRequired'), trigger: 'blur' }
}))

function genSecret(): string {
  const arr = new Uint8Array(16)
  crypto.getRandomValues(arr)
  return Array.from(arr, (b) => b.toString(16).padStart(2, '0')).join('')
}

function resetForm() {
  form.value = {
    name: '', repo_url: '', branch: 'main', secret: genSecret(),
    build_type: 'static', out_dir: 'dist', dest_path: '', container_port: 8080
  }
}

function openCreate() {
  isEdit.value = false
  editId.value = null
  resetForm()
  showForm.value = true
}

function openEdit(row: DeployApp) {
  isEdit.value = true
  editId.value = row.id
  form.value = {
    name: row.name,
    repo_url: row.repo_url,
    branch: row.branch,
    secret: row.secret,
    build_type: row.build_type,
    out_dir: row.out_dir || 'dist',
    dest_path: row.dest_path,
    container_port: row.container_port || 8080
  }
  showForm.value = true
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }
  formLoading.value = true
  try {
    if (isEdit.value && editId.value !== null) {
      await updateDeployApp(editId.value, { ...form.value })
      message.success(t('deploy.messages.updated'))
    } else {
      await createDeployApp({ ...form.value, enabled: true })
      message.success(t('deploy.messages.created'))
    }
    showForm.value = false
    load()
  } catch (err: any) {
    message.error(err.message || t('deploy.messages.saveFailed'))
  } finally {
    formLoading.value = false
  }
}

/* ==================== 删除 ==================== */
const showDelete = ref(false)
const deleteLoading = ref(false)
const deleteTarget = ref<DeployApp | null>(null)

function openDelete(row: DeployApp) {
  deleteTarget.value = row
  showDelete.value = true
}

async function handleDeleteConfirm() {
  if (!deleteTarget.value) return
  deleteLoading.value = true
  try {
    await deleteDeployApp(deleteTarget.value.id)
    message.success(t('deploy.messages.deleted'))
    showDelete.value = false
    load()
  } catch (err: any) {
    message.error(err.message || t('deploy.messages.deleteFailed'))
  } finally {
    deleteLoading.value = false
  }
}

/* ==================== 日志 ==================== */
const showLogs = ref(false)
const logsLoading = ref(false)
const logs = ref<DeployLog[]>([])
const logsApp = ref<DeployApp | null>(null)
const expandedLogId = ref<number | null>(null)

async function openLogs(row: DeployApp) {
  logsApp.value = row
  expandedLogId.value = null
  showLogs.value = true
  logsLoading.value = true
  try {
    const res = await listDeployLogs(row.id, 20)
    logs.value = res.items || []
  } catch {
    logs.value = []
  } finally {
    logsLoading.value = false
  }
}

function webhookURL(id: number): string {
  const base = (window as any).__PANEL_BASE__ || ''
  return `${window.location.origin}${base}/api/v1/deploy/webhook/${id}?sign=<secret>`
}

onMounted(load)
</script>

<template>
  <PageCard :title="t('deploy.pageTitle')" :subtitle="t('deploy.pageSubtitle')">
    <template #extra>
      <NSpace>
        <NButton @click="load">{{ t('deploy.refresh') }}</NButton>
        <NButton type="primary" @click="openCreate">
          <template #icon><NIcon><AddOutline /></NIcon></template>
          {{ t('deploy.addApp') }}
        </NButton>
      </NSpace>
    </template>

    <DataTable :columns="columns" :data="apps" :loading="loading">
      <template #empty>
        <div class="empty-state">
          <p>{{ t('deploy.empty.title') }}</p>
          <span>{{ t('deploy.empty.desc') }}</span>
        </div>
      </template>
    </DataTable>

    <div v-if="apps.length > 0" class="webhook-tip">
      {{ t('deploy.webhook.tipPrefix') }}<span class="mono">{{ webhookURL(apps[0].id) }}</span>{{ t('deploy.webhook.tipSuffix') }}
    </div>
  </PageCard>

  <!-- 创建/编辑 -->
  <ModalForm
    :show="showForm"
    :title="isEdit ? t('deploy.form.editTitle') : t('deploy.form.addTitle')"
    :loading="formLoading"
    :width="620"
    @update:show="(v: boolean) => (showForm = v)"
    @submit="handleSubmit"
  >
    <NForm ref="formRef" :model="form" :rules="formRules" label-placement="top">
      <NSpace size="medium" style="width: 100%">
        <NFormItem :label="t('deploy.form.nameLabel')" path="name" style="flex: 1">
          <NInput v-model:value="form.name" :placeholder="t('deploy.form.namePlaceholder')" clearable />
        </NFormItem>
        <NFormItem :label="t('deploy.form.branchLabel')" path="branch" style="flex: 1">
          <NInput v-model:value="form.branch" placeholder="main" clearable />
        </NFormItem>
      </NSpace>
      <NFormItem :label="t('deploy.form.repoLabel')" path="repo_url">
        <NInput v-model:value="form.repo_url" :placeholder="t('deploy.form.repoPlaceholder')" clearable />
      </NFormItem>
      <NSpace size="medium" style="width: 100%">
        <NFormItem :label="t('deploy.form.buildTypeLabel')" path="build_type" style="flex: 1">
          <NSelect v-model:value="form.build_type" :options="buildTypeOptions" />
        </NFormItem>
        <NFormItem v-if="form.build_type === 'dockerfile'" :label="t('deploy.form.containerPortLabel')" path="container_port" style="flex: 1">
          <NInputNumber v-model:value="form.container_port" :min="1" :max="65535" style="width: 100%" />
        </NFormItem>
        <NFormItem v-else-if="form.build_type === 'node'" :label="t('deploy.form.outDirLabel')" path="out_dir" style="flex: 1">
          <NInput v-model:value="form.out_dir" placeholder="dist" clearable />
        </NFormItem>
      </NSpace>
      <NFormItem :label="t('deploy.form.destPathLabel')" path="dest_path">
        <NInput v-model:value="form.dest_path" :placeholder="t('deploy.form.destPathPlaceholder')" clearable />
      </NFormItem>
      <NFormItem label="Webhook Secret" path="secret">
        <NInput v-model:value="form.secret" clearable />
      </NFormItem>
      <div class="form-hint">{{ t('deploy.form.hint') }}</div>
    </NForm>
  </ModalForm>

  <!-- 日志 -->
  <ContentModal
    v-model:show="showLogs"
    :title="t('deploy.logs.title', { name: logsApp?.name || '' })"
    size="lg"
  >
    <div v-if="logs.length === 0" class="empty-state">
      <p>{{ t('deploy.logs.empty') }}</p>
    </div>
    <div v-for="log in logs" :key="log.id" class="log-item">
      <div class="log-item__head" @click="expandedLogId = expandedLogId === log.id ? null : log.id">
        <StatusTag
          :status="log.status === 'success' ? 'running' : 'error'"
          :text="log.status === 'success' ? t('deploy.status.success') : t('deploy.status.failed')"
        />
        <span class="mono faint">{{ fmtTime(log.created_at) }}</span>
        <span class="log-item__toggle">{{ expandedLogId === log.id ? t('deploy.logs.collapse') : t('deploy.logs.expand') }}</span>
      </div>
      <pre v-if="expandedLogId === log.id" class="log-output">{{ log.output }}</pre>
    </div>
  </ContentModal>

  <!-- 删除确认 -->
  <ConfirmDialog
    :show="showDelete"
    :title="t('deploy.deleteDialog.title')"
    :content="t('deploy.deleteDialog.content', { name: deleteTarget?.name || '' })"
    :confirm-text="t('deploy.deleteDialog.confirm')"
    type="error"
    :loading="deleteLoading"
    @update:show="(v: boolean) => (showDelete = v)"
    @confirm="handleDeleteConfirm"
  />
</template>

<style scoped>
.app-name {
  font-weight: 500;
  color: var(--ink);
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
  margin: 0 0 var(--space-2);
}
.empty-state span {
  color: var(--text-quaternary);
  font-size: var(--fs-caption);
}
.webhook-tip {
  margin-top: var(--space-3);
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
  word-break: break-all;
}
.form-hint {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
  margin-bottom: var(--space-2);
}
.log-item {
  border: 1px solid var(--line);
  border-radius: var(--radius-sm);
  margin-bottom: var(--space-2);
  overflow: hidden;
}
.log-item__head {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  cursor: pointer;
}
.log-item__toggle {
  margin-left: auto;
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
}
.log-output {
  margin: 0;
  padding: var(--space-3);
  background: var(--surface2);
  font-family: var(--font-mono, 'Cascadia Code', Consolas, monospace);
  font-size: var(--fs-caption);
  color: var(--sub);
  max-height: 320px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
.mono {
  font-family: var(--font-mono, 'Cascadia Code', Consolas, monospace);
}
</style>
