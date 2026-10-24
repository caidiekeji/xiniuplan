<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NButton, NIcon, NForm, NFormItem, NInput, NSwitch, NTag,
  NSpace
} from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import {
  AddOutline, PlayOutline, StopOutline, CreateOutline, TrashOutline, DocumentTextOutline
} from '@vicons/ionicons5'
import { PageCard, DataTable, ModalForm, ContentModal, StatusTag, ConfirmDialog } from '@/components'
import {
  listDaemons, createDaemon, updateDaemon, deleteDaemon, startDaemon, stopDaemon, getDaemonLogs,
  type DaemonProcess
} from '@/api/features17'

const { t } = useI18n()
const message = useMessage()

function fmtTime(v: string | null): string {
  if (!v) return '—'
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return v
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

/* ==================== 列表 ==================== */
const loading = ref(false)
const daemons = ref<DaemonProcess[]>([])
const actionLoadingId = ref<number | null>(null)

async function load() {
  loading.value = true
  try {
    const res = await listDaemons()
    daemons.value = res.items || []
  } catch {
    daemons.value = []
  } finally {
    loading.value = false
  }
}

async function handleStart(id: number) {
  actionLoadingId.value = id
  try {
    await startDaemon(id)
    message.success(t('daemon.messages.started'))
    load()
  } catch (err: any) {
    message.error(err.message || t('daemon.messages.startFailed'))
  } finally {
    actionLoadingId.value = null
  }
}

async function handleStop(id: number) {
  actionLoadingId.value = id
  try {
    await stopDaemon(id)
    message.success(t('daemon.messages.stopped'))
    load()
  } catch (err: any) {
    message.error(err.message || t('daemon.messages.stopFailed'))
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
  command: '',
  work_dir: '',
  auto_start: true,
  enabled: true
})

const formRules = computed(() => ({
  name: { required: true, message: t('daemon.rules.nameRequired'), trigger: 'blur' },
  command: { required: true, message: t('daemon.rules.commandRequired'), trigger: 'blur' }
}))

function resetForm() {
  form.value = { name: '', command: '', work_dir: '', auto_start: true, enabled: true }
}

function openCreate() {
  isEdit.value = false
  editId.value = null
  resetForm()
  showForm.value = true
}

function openEdit(row: DaemonProcess) {
  isEdit.value = true
  editId.value = row.id
  form.value = {
    name: row.name,
    command: row.command,
    work_dir: row.work_dir || '',
    auto_start: row.auto_start,
    enabled: row.enabled
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
      await updateDaemon(editId.value, { ...form.value })
      message.success(t('daemon.messages.updated'))
    } else {
      await createDaemon({ ...form.value })
      message.success(t('daemon.messages.created'))
    }
    showForm.value = false
    load()
  } catch (err: any) {
    message.error(err.message || t('daemon.messages.saveFailed'))
  } finally {
    formLoading.value = false
  }
}

/* ==================== 删除 ==================== */
const showDelete = ref(false)
const deleteLoading = ref(false)
const deleteTarget = ref<DaemonProcess | null>(null)

function openDelete(row: DaemonProcess) {
  deleteTarget.value = row
  showDelete.value = true
}

async function handleDeleteConfirm() {
  if (!deleteTarget.value) return
  deleteLoading.value = true
  try {
    await deleteDaemon(deleteTarget.value.id)
    message.success(t('daemon.messages.deleted'))
    showDelete.value = false
    load()
  } catch (err: any) {
    message.error(err.message || t('daemon.messages.deleteFailed'))
  } finally {
    deleteLoading.value = false
  }
}

/* ==================== 日志 ==================== */
const showLogs = ref(false)
const logsLoading = ref(false)
const logsContent = ref('')
const logsDaemon = ref<DaemonProcess | null>(null)

async function openLogs(row: DaemonProcess) {
  logsDaemon.value = row
  logsContent.value = ''
  showLogs.value = true
  logsLoading.value = true
  try {
    const res = await getDaemonLogs(row.id, 300)
    logsContent.value = res.content || ''
  } catch {
    logsContent.value = ''
  } finally {
    logsLoading.value = false
  }
}

/* ==================== 表格列 ==================== */
const columns = computed(() => [
  {
    title: t('daemon.columns.name'),
    key: 'name',
    width: 150,
    render: (row: any) => h('span', { class: 'daemon-name' }, row.name)
  },
  {
    title: t('daemon.columns.command'),
    key: 'command',
    ellipsis: { tooltip: true },
    render: (row: any) => h('span', { class: 'mono' }, row.command)
  },
  {
    title: t('daemon.columns.workDir'),
    key: 'work_dir',
    ellipsis: { tooltip: true },
    render: (row: any) =>
      row.work_dir ? h('span', { class: 'mono' }, row.work_dir) : h('span', { class: 'faint' }, '—')
  },
  {
    title: t('daemon.columns.status'),
    key: 'pid',
    width: 150,
    render: (row: any) => {
      const running = row.pid > 0
      return h(NSpace, { size: 'small', wrap: false, align: 'center' }, {
        default: () => [
          h(StatusTag, {
            status: running ? 'running' : 'stopped',
            text: running ? t('daemon.status.running') : t('daemon.status.stopped')
          }),
          running ? h('span', { class: 'mono faint' }, `PID ${row.pid}`) : null
        ]
      })
    }
  },
  {
    title: t('daemon.columns.restarts'),
    key: 'restarts',
    width: 90,
    render: (row: any) => h('span', { class: 'mono' }, String(row.restarts))
  },
  {
    title: t('daemon.columns.lastExitCode'),
    key: 'last_exit_code',
    width: 100,
    render: (row: any) =>
      row.pid > 0
        ? h('span', { class: 'faint' }, '—')
        : h('span', { class: 'mono' }, String(row.last_exit_code))
  },
  {
    title: t('daemon.columns.autoStart'),
    key: 'auto_start',
    width: 90,
    render: (row: any) => h(NTag, {
      size: 'small', bordered: false, round: true,
      type: row.auto_start ? 'success' : 'default'
    }, { default: () => (row.auto_start ? t('daemon.autoStart.on') : t('daemon.autoStart.off')) })
  },
  {
    title: t('daemon.columns.actions'),
    key: 'actions',
    width: 280,
    render: (row: any) => {
      const busy = actionLoadingId.value === row.id
      const running = row.pid > 0
      return h(NSpace, { size: 'small', wrap: false }, {
        default: () => [
          h(NButton, {
            size: 'tiny', quaternary: true, loading: busy, disabled: running,
            onClick: () => handleStart(row.id)
          }, { icon: () => h(NIcon, null, { default: () => h(PlayOutline) }), default: () => t('daemon.actions.start') }),
          h(NButton, {
            size: 'tiny', quaternary: true, loading: busy, disabled: !running,
            onClick: () => handleStop(row.id)
          }, { icon: () => h(NIcon, null, { default: () => h(StopOutline) }), default: () => t('daemon.actions.stop') }),
          h(NButton, {
            size: 'tiny', quaternary: true,
            onClick: () => openLogs(row)
          }, { icon: () => h(NIcon, null, { default: () => h(DocumentTextOutline) }), default: () => t('daemon.actions.logs') }),
          h(NButton, {
            size: 'tiny', quaternary: true,
            onClick: () => openEdit(row)
          }, { icon: () => h(NIcon, null, { default: () => h(CreateOutline) }), default: () => t('daemon.actions.edit') }),
          h(NButton, {
            size: 'tiny', quaternary: true, type: 'error',
            onClick: () => openDelete(row)
          }, { icon: () => h(NIcon, null, { default: () => h(TrashOutline) }), default: () => t('daemon.actions.delete') })
        ]
      })
    }
  }
])

onMounted(load)
</script>

<template>
  <PageCard :title="t('daemon.pageTitle')" :subtitle="t('daemon.pageSubtitle')">
    <template #extra>
      <NSpace>
        <NButton @click="load">{{ t('daemon.refresh') }}</NButton>
        <NButton type="primary" @click="openCreate">
          <template #icon><NIcon><AddOutline /></NIcon></template>
          {{ t('daemon.addDaemon') }}
        </NButton>
      </NSpace>
    </template>

    <DataTable :columns="columns" :data="daemons" :loading="loading">
      <template #empty>
        <div class="empty-state">
          <p>{{ t('daemon.empty.title') }}</p>
          <span>{{ t('daemon.empty.desc') }}</span>
        </div>
      </template>
    </DataTable>
  </PageCard>

  <!-- 创建/编辑 -->
  <ModalForm
    :show="showForm"
    :title="isEdit ? t('daemon.form.editTitle') : t('daemon.form.addTitle')"
    :loading="formLoading"
    :width="560"
    @update:show="(v: boolean) => (showForm = v)"
    @submit="handleSubmit"
  >
    <NForm ref="formRef" :model="form" :rules="formRules" label-placement="top">
      <NFormItem :label="t('daemon.form.nameLabel')" path="name">
        <NInput v-model:value="form.name" :placeholder="t('daemon.form.namePlaceholder')" clearable />
      </NFormItem>
      <NFormItem :label="t('daemon.form.commandLabel')" path="command">
        <NInput
          v-model:value="form.command"
          type="textarea"
          :rows="2"
          :placeholder="t('daemon.form.commandPlaceholder')"
        />
      </NFormItem>
      <div class="form-hint">{{ t('daemon.form.commandHint') }}</div>
      <NFormItem :label="t('daemon.form.workDirLabel')" path="work_dir">
        <NInput v-model:value="form.work_dir" :placeholder="t('daemon.form.workDirPlaceholder')" clearable />
      </NFormItem>
      <NSpace size="medium" style="width: 100%">
        <NFormItem :label="t('daemon.form.autoStartLabel')" style="flex: 1">
          <NSwitch v-model:value="form.auto_start" />
        </NFormItem>
        <NFormItem :label="t('daemon.form.enabledLabel')" style="flex: 0 0 auto">
          <NSwitch v-model:value="form.enabled" />
        </NFormItem>
      </NSpace>
    </NForm>
  </ModalForm>

  <!-- 日志 -->
  <ContentModal
    v-model:show="showLogs"
    :title="t('daemon.logs.title', { name: logsDaemon?.name || '' })"
    size="xl"
  >
    <div v-if="logsLoading" class="log-loading">{{ t('daemon.logs.loading') }}</div>
    <pre v-else class="mono log-block">{{ logsContent || t('daemon.logs.empty') }}</pre>
  </ContentModal>

  <!-- 删除确认 -->
  <ConfirmDialog
    :show="showDelete"
    :title="t('daemon.deleteDialog.title')"
    :content="t('daemon.deleteDialog.content', { name: deleteTarget?.name || '' })"
    :confirm-text="t('daemon.deleteDialog.confirm')"
    type="error"
    :loading="deleteLoading"
    @update:show="(v: boolean) => (showDelete = v)"
    @confirm="handleDeleteConfirm"
  />
</template>

<style scoped>
.daemon-name {
  font-weight: 500;
  color: var(--ink);
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}
.faint {
  color: var(--text-quaternary);
  font-size: var(--fs-caption);
}
.form-hint {
  color: var(--text-quaternary);
  font-size: var(--fs-caption);
  margin: calc(var(--space-2) * -1) 0 var(--space-3);
}
.log-loading {
  padding: var(--space-8) 0;
  text-align: center;
  color: var(--sub);
  font-size: var(--fs-body);
}
.log-block {
  margin: 0;
  padding: var(--space-3);
  max-height: 420px;
  overflow: auto;
  background: var(--surface2);
  border-radius: var(--radius-sm);
  font-size: var(--fs-caption);
  color: var(--sub);
  white-space: pre-wrap;
  word-break: break-all;
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
