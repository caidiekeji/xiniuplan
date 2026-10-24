<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, h } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NButton, NIcon, NForm, NFormItem, NInput, NInputNumber, NSelect, NSwitch,
  NSpace, NTag
} from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import { AddOutline, PlayOutline, CreateOutline, TrashOutline, TimeOutline } from '@vicons/ionicons5'
import { PageCard, DataTable, ModalForm, ContentModal, StatusTag, ConfirmDialog } from '@/components'
import type { StatusType } from '@/components'
import {
  listProbes, createProbe, updateProbe, deleteProbe, checkProbeNow, listProbeResults,
  type Probe, type ProbeResult
} from '@/api/features17'
import type { DataTableColumns } from 'naive-ui'

const { t } = useI18n()
const message = useMessage()

const probeTypeOptions = [
  { label: 'HTTP(S)', value: 'http' },
  { label: 'TCP', value: 'tcp' }
]

const statusMap = computed<Record<string, { status: StatusType; text: string }>>(() => ({
  up: { status: 'running', text: t('probe.status.up') },
  down: { status: 'error', text: t('probe.status.down') },
  unknown: { status: 'pending', text: t('probe.status.unknown') }
}))

function fmtTime(v: string | null): string {
  if (!v) return '—'
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return v
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getMonth() + 1}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

/* ==================== 列表 ==================== */
const loading = ref(false)
const probes = ref<Probe[]>([])
const actionLoadingId = ref<number | null>(null)

async function load() {
  loading.value = true
  try {
    const res = await listProbes()
    probes.value = res.items || []
  } catch {
    probes.value = []
  } finally {
    loading.value = false
  }
}

const columns = computed<DataTableColumns>(() => [
  {
    title: t('probe.columns.name'),
    key: 'name',
    width: 160,
    render: (row: any) => h('span', { class: 'probe-name' }, row.name)
  },
  {
    title: t('probe.columns.type'),
    key: 'type',
    width: 90,
    render: (row: any) => h(NTag, { size: 'small', type: 'info', bordered: false, round: true },
      { default: () => (row.type === 'http' ? 'HTTP(S)' : 'TCP') })
  },
  {
    title: t('probe.columns.target'),
    key: 'target',
    ellipsis: { tooltip: true },
    render: (row: any) => h('span', { class: 'mono' }, row.target)
  },
  {
    title: t('probe.columns.interval'),
    key: 'interval_sec',
    width: 80,
    render: (row: any) => h('span', { class: 'mono' }, `${row.interval_sec}s`)
  },
  {
    title: t('probe.columns.status'),
    key: 'last_status',
    width: 90,
    render: (row: any) => {
      const cfg = statusMap.value[row.last_status] || statusMap.value.unknown
      return h(StatusTag, { status: cfg.status, text: cfg.text })
    }
  },
  {
    title: t('probe.columns.latency'),
    key: 'last_latency_ms',
    width: 90,
    render: (row: any) =>
      row.last_status === 'unknown'
        ? h('span', { class: 'faint' }, '—')
        : h('span', { class: 'mono' }, `${Math.round(row.last_latency_ms)}ms`)
  },
  {
    title: t('probe.columns.httpCode'),
    key: 'last_code',
    width: 90,
    render: (row: any) =>
      row.type === 'http'
        ? h('span', { class: 'mono' }, row.last_code ? String(row.last_code) : '—')
        : h('span', { class: 'faint' }, '—')
  },
  {
    title: t('probe.columns.lastChecked'),
    key: 'last_checked',
    width: 140,
    render: (row: any) => h('span', { class: 'mono faint' }, fmtTime(row.last_checked))
  },
  {
    title: t('probe.columns.enabled'),
    key: 'enabled',
    width: 80,
    render: (row: any) => h(StatusTag, {
      status: row.enabled ? 'running' : 'stopped',
      text: row.enabled ? t('probe.switchStatus.enabled') : t('probe.switchStatus.disabled')
    })
  },
  {
    title: t('probe.columns.actions'),
    key: 'actions',
    width: 230,
    render: (row: any) => {
      const busy = actionLoadingId.value === row.id
      return h(NSpace, { size: 'small', wrap: false }, {
        default: () => [
          h(NButton, {
            size: 'tiny', quaternary: true, loading: busy,
            onClick: () => handleCheckNow(row.id)
          }, { icon: () => h(NIcon, null, { default: () => h(PlayOutline) }), default: () => t('probe.actions.checkNow') }),
          h(NButton, {
            size: 'tiny', quaternary: true,
            onClick: () => openEdit(row)
          }, { icon: () => h(NIcon, null, { default: () => h(CreateOutline) }), default: () => t('probe.actions.edit') }),
          h(NButton, {
            size: 'tiny', quaternary: true, type: 'error', loading: busy,
            onClick: () => openDelete(row)
          }, { icon: () => h(NIcon, null, { default: () => h(TrashOutline) }), default: () => t('probe.actions.delete') })
        ]
      })
    }
  }
])

async function handleCheckNow(id: number) {
  actionLoadingId.value = id
  try {
    await checkProbeNow(id)
    message.success(t('probe.messages.checkDone'))
    load()
  } catch (err: any) {
    message.error(err.message || t('probe.messages.checkFailed'))
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
  type: 'http' as 'http' | 'tcp',
  target: '',
  interval_sec: 60,
  keyword: '',
  enabled: true
})

const formRules = computed(() => ({
  name: { required: true, message: t('probe.rules.nameRequired'), trigger: 'blur' },
  target: { required: true, message: t('probe.rules.targetRequired'), trigger: 'blur' }
}))

function resetForm() {
  form.value = { name: '', type: 'http', target: '', interval_sec: 60, keyword: '', enabled: true }
}

function openCreate() {
  isEdit.value = false
  editId.value = null
  resetForm()
  showForm.value = true
}

function openEdit(row: Probe) {
  isEdit.value = true
  editId.value = row.id
  form.value = {
    name: row.name,
    type: row.type,
    target: row.target,
    interval_sec: row.interval_sec,
    keyword: row.keyword || '',
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
      await updateProbe(editId.value, { ...form.value })
      message.success(t('probe.messages.updated'))
    } else {
      await createProbe({ ...form.value })
      message.success(t('probe.messages.created'))
    }
    showForm.value = false
    load()
  } catch (err: any) {
    message.error(err.message || t('probe.messages.saveFailed'))
  } finally {
    formLoading.value = false
  }
}

/* ==================== 删除 ==================== */
const showDelete = ref(false)
const deleteLoading = ref(false)
const deleteTarget = ref<Probe | null>(null)

function openDelete(row: Probe) {
  deleteTarget.value = row
  showDelete.value = true
}

async function handleDeleteConfirm() {
  if (!deleteTarget.value) return
  deleteLoading.value = true
  try {
    await deleteProbe(deleteTarget.value.id)
    message.success(t('probe.messages.deleted'))
    showDelete.value = false
    load()
  } catch (err: any) {
    message.error(err.message || t('probe.messages.deleteFailed'))
  } finally {
    deleteLoading.value = false
  }
}

/* ==================== 历史结果 ==================== */
const showResults = ref(false)
const resultsLoading = ref(false)
const results = ref<ProbeResult[]>([])
const resultsProbe = ref<Probe | null>(null)

const resultColumns = computed<DataTableColumns<ProbeResult>>(() => [
  {
    title: t('probe.columns.checkedAt'),
    key: 'checked_at',
    width: 170,
    render: (row) => h('span', { class: 'mono faint' }, fmtTime(row.checked_at))
  },
  {
    title: t('probe.columns.result'),
    key: 'status',
    width: 90,
    render: (row) => h(StatusTag, {
      status: row.status === 'up' ? 'running' : 'error',
      text: row.status === 'up' ? t('probe.status.up') : t('probe.status.down')
    })
  },
  { title: t('probe.columns.httpCode'), key: 'code', width: 90, render: (row) => h('span', { class: 'mono' }, row.code || '—') },
  {
    title: t('probe.columns.latency'),
    key: 'latency_ms',
    width: 90,
    render: (row) => h('span', { class: 'mono' }, `${Math.round(row.latency_ms)}ms`)
  },
  { title: t('probe.columns.message'), key: 'message', ellipsis: { tooltip: true } }
])

async function openResults(row: Probe) {
  resultsProbe.value = row
  showResults.value = true
  resultsLoading.value = true
  try {
    const res = await listProbeResults(row.id, 100)
    results.value = res.items || []
  } catch {
    results.value = []
  } finally {
    resultsLoading.value = false
  }
}

/* ==================== 生命周期 ==================== */
let timer: number | null = null
onMounted(() => {
  load()
  timer = window.setInterval(load, 30000)
})
onBeforeUnmount(() => {
  if (timer !== null) window.clearInterval(timer)
})
</script>

<template>
  <PageCard :title="t('probe.pageTitle')" :subtitle="t('probe.pageSubtitle')">
    <template #extra>
      <NSpace>
        <NButton @click="load">{{ t('probe.refresh') }}</NButton>
        <NButton type="primary" @click="openCreate">
          <template #icon><NIcon><AddOutline /></NIcon></template>
          {{ t('probe.addProbe') }}
        </NButton>
      </NSpace>
    </template>

    <DataTable :columns="columns" :data="probes" :loading="loading">
      <template #empty>
        <div class="empty-state">
          <p>{{ t('probe.empty.title') }}</p>
          <span>{{ t('probe.empty.desc') }}</span>
        </div>
      </template>
    </DataTable>
  </PageCard>

  <!-- 历史结果入口（放在列表下方，点击「状态」行操作） -->
  <div class="probe-tip">
    <NButton size="small" quaternary :disabled="probes.length === 0" @click="probes.length && openResults(probes[0])">
      <template #icon><NIcon><TimeOutline /></NIcon></template>
      {{ t('probe.historyButton') }}
    </NButton>
  </div>

  <!-- 创建/编辑 -->
  <ModalForm
    :show="showForm"
    :title="isEdit ? t('probe.form.editTitle') : t('probe.form.addTitle')"
    :loading="formLoading"
    :width="560"
    @update:show="(v: boolean) => (showForm = v)"
    @submit="handleSubmit"
  >
    <NForm ref="formRef" :model="form" :rules="formRules" label-placement="top">
      <NFormItem :label="t('probe.form.nameLabel')" path="name">
        <NInput v-model:value="form.name" :placeholder="t('probe.form.namePlaceholder')" clearable />
      </NFormItem>
      <NSpace size="medium" style="width: 100%">
        <NFormItem :label="t('probe.form.typeLabel')" path="type" style="flex: 1">
          <NSelect v-model:value="form.type" :options="probeTypeOptions" />
        </NFormItem>
        <NFormItem :label="t('probe.form.intervalLabel')" path="interval_sec" style="flex: 1">
          <NInputNumber v-model:value="form.interval_sec" :min="30" :max="3600" style="width: 100%" />
        </NFormItem>
      </NSpace>
      <NFormItem :label="t('probe.form.targetLabel')" path="target">
        <NInput
          v-model:value="form.target"
          :placeholder="form.type === 'http' ? 'https://example.com/health' : '192.168.1.1:3306'"
          clearable
        />
      </NFormItem>
      <NFormItem v-if="form.type === 'http'" :label="t('probe.form.keywordLabel')">
        <NInput v-model:value="form.keyword" :placeholder="t('probe.form.keywordPlaceholder')" clearable />
      </NFormItem>
      <NFormItem :label="t('probe.form.enabledLabel')">
        <NSwitch v-model:value="form.enabled" />
      </NFormItem>
    </NForm>
  </ModalForm>

  <!-- 检测历史 -->
  <ContentModal
    v-model:show="showResults"
    :title="t('probe.results.title', { name: resultsProbe?.name || '' })"
    size="xl"
  >
    <DataTable
      :columns="resultColumns"
      :data="results"
      :loading="resultsLoading"
      :max-height="420"
      :pagination="false"
      size="small"
    />
  </ContentModal>

  <!-- 删除确认 -->
  <ConfirmDialog
    :show="showDelete"
    :title="t('probe.deleteDialog.title')"
    :content="t('probe.deleteDialog.content', { name: deleteTarget?.name || '' })"
    :confirm-text="t('probe.deleteDialog.confirm')"
    type="error"
    :loading="deleteLoading"
    @update:show="(v: boolean) => (showDelete = v)"
    @confirm="handleDeleteConfirm"
  />
</template>

<style scoped>
.probe-name {
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
  font-size: var(--fs-body);
  margin: 0 0 var(--space-2);
}
.empty-state span {
  color: var(--text-quaternary);
  font-size: var(--fs-caption);
}
.probe-tip {
  margin-top: var(--space-3);
}
</style>
