<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue'
import { NButton, NIcon, NForm, NFormItem, NInput, NSelect, NRadioGroup, NRadio, NSpace, NTag, NDrawer, NDrawerContent } from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import {
  AddOutline, PlayOutline, StopOutline, RefreshOutline,
  DocumentTextOutline, TrashOutline, CreateOutline, TimerOutline
} from '@vicons/ionicons5'
import {
  PageCard, DataTable, ModalForm, StatusTag, ConfirmDialog, CodeEditor
} from '@/components'
import type { StatusType } from '@/components'
import {
  getCronTasks, createCronTask, updateCronTask, deleteCronTask,
  enableCronTask, disableCronTask, triggerCronTask, getCronLogs,
  type CronTask, type CronTaskType, type CronTaskStatus,
  type CreateCronTaskData, type CronLog
} from '@/api/cron'
import type { DataTableColumns } from 'naive-ui'
import { useI18n } from 'vue-i18n'

const message = useMessage()
const { t } = useI18n()

/* ========== 常量映射 ========== */
const typeTagTypeMap: Record<CronTaskType, 'info' | 'success' | 'warning' | 'default' | 'error'> = {
  shell: 'info',
  backup: 'success',
  url: 'warning',
  app_restart: 'default',
  ssl_check: 'warning',
  docker_cleanup: 'info',
  recycle_cleanup: 'default',
  tamper_check: 'error',
  virus_scan: 'error'
}

/* 内置任务类型：无需 command/target 参数 */
const builtinTypes: CronTaskType[] = ['ssl_check', 'docker_cleanup', 'recycle_cleanup', 'tamper_check']

const statusMap: Record<CronTaskStatus, StatusType> = {
  enabled: 'running',
  disabled: 'stopped'
}

/* ========== Cron 预设 ========== */
const cronPresets = computed(() => [
  { label: t('cron.presets.minute'), value: '* * * * *' },
  { label: t('cron.presets.hourly'), value: '0 * * * *' },
  { label: t('cron.presets.daily'), value: '0 3 * * *' },
  { label: t('cron.presets.weekly'), value: '0 3 * * 0' },
  { label: t('cron.presets.monthly'), value: '0 3 1 * *' }
])

/* ========== Cron 人类可读描述 ========== */
function describeCron(expr: string): string {
  const parts = expr.trim().split(/\s+/)
  if (parts.length !== 5) return expr
  const [minute, hour, day, month, week] = parts

  if (minute === '*' && hour === '*' && day === '*' && month === '*' && week === '*') return t('cron.desc.everyMinute')
  if (minute === '0' && hour === '*' && day === '*' && month === '*' && week === '*') return t('cron.desc.everyHour')
  if (hour === '*' && day === '*' && month === '*' && week === '*') return t('cron.desc.hourlyAt', { minute })
  if (day === '*' && month === '*' && week === '*') {
    return t('cron.desc.dailyAt', { time: `${hour.padStart(2, '0')}:${minute.padStart(2, '0')}` })
  }
  if (month === '*' && week === '*' && day !== '*') {
    return t('cron.desc.monthlyAt', { day, time: `${hour.padStart(2, '0')}:${minute.padStart(2, '0')}` })
  }
  if (day === '*' && month === '*' && week !== '*') {
    const weekKeyMap: Record<string, string> = { '0': 'sun', '1': 'mon', '2': 'tue', '3': 'wed', '4': 'thu', '5': 'fri', '6': 'sat', '7': 'sun' }
    const weekDay = t(`cron.weekNames.${weekKeyMap[week] || 'sun'}`)
    return t('cron.desc.weeklyAt', { day: weekDay, time: `${hour.padStart(2, '0')}:${minute.padStart(2, '0')}` })
  }
  return expr
}

/* ========== 列表状态 ========== */
const loading = ref(false)
const tasks = ref<CronTask[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

/* ========== Mock 数据兜底 ========== */
async function loadTasks() {
  loading.value = true
  try {
    const res = await getCronTasks(page.value, pageSize.value)
    if (res.data?.items) {
      tasks.value = res.data.items
      total.value = res.data.total
    } else {
      tasks.value = []
      total.value = 0
    }
  } catch (e: any) {
    tasks.value = []
    total.value = 0
    message.error(e.message || t('cron.loadFailed'))
  } finally {
    loading.value = false
  }
}

const pagination = computed(() => ({
  page: page.value,
  pageSize: pageSize.value,
  itemCount: total.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  'onUpdate:page': (p: number) => { page.value = p; loadTasks() },
  'onUpdate:page-size': (ps: number) => { pageSize.value = ps; page.value = 1; loadTasks() }
}))

/* ========== 表格列 ========== */
const columns = computed<DataTableColumns>(() => [
  {
    title: t('cron.nameCol'),
    key: 'name',
    width: 180,
    render: (row: any) => h('div', { class: 'cron-name' }, [
      h('span', { class: 'cron-name__main' }, row.name),
      row.remark
        ? h('span', { class: 'cron-name__remark' }, row.remark)
        : null
    ])
  },
  {
    title: t('cron.typeCol'),
    key: 'type',
    width: 100,
    render: (row: any) => h(NTag, {
      size: 'small',
      type: typeTagTypeMap[row.type as CronTaskType],
      bordered: false,
      round: true
    }, { default: () => t('cron.types.' + row.type) })
  },
  {
    title: t('cron.cronCol'),
    key: 'cron',
    width: 240,
    render: (row: any) => h('div', { class: 'cron-expr' }, [
      h('span', { class: 'cron-expr__code mono' }, row.schedule),
      h('span', { class: 'cron-expr__desc' }, describeCron(row.schedule))
    ])
  },
  {
    title: t('cron.statusCol'),
    key: 'status',
    width: 100,
    render: (row: any) => {
      const status = statusMap[row.status as CronTaskStatus]
      return h(StatusTag, { status, text: t('cron.statuses.' + row.status) })
    }
  },
  {
    title: t('cron.lastRunCol'),
    key: 'lastRunAt',
    width: 160,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.lastRunAt || '—')
  },
  {
    title: t('cron.nextRunCol'),
    key: 'nextRunAt',
    width: 160,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.nextRunAt || '—')
  },
  {
    title: t('cron.actionsCol'),
    key: 'actions',
    width: 320,
    render: (row: any) => {
      const isEnabled = row.status === 'enabled'
      const actionLoading = actionLoadingId.value === row.id
      return h(NSpace, { size: 'small', wrap: false }, {
        default: () => [
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            type: isEnabled ? 'warning' : 'success',
            loading: actionLoading,
            onClick: () => isEnabled ? handleDisable(row.id) : handleEnable(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(isEnabled ? StopOutline : PlayOutline) }),
            default: () => isEnabled ? t('cron.disable') : t('cron.enable')
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            loading: actionLoading,
            onClick: () => handleTrigger(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(RefreshOutline) }),
            default: () => t('cron.trigger')
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            onClick: () => openLogs(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(DocumentTextOutline) }),
            default: () => t('cron.logs')
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            onClick: () => openEdit(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(CreateOutline) }),
            default: () => t('cron.edit')
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            type: 'error',
            loading: actionLoading,
            onClick: () => openDelete(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
            default: () => t('cron.delete')
          })
        ]
      })
    }
  }
])

/* ========== 操作 loading ========== */
const actionLoadingId = ref<number | null>(null)

async function handleEnable(id: number) {
  actionLoadingId.value = id
  try {
    await enableCronTask(id)
    message.success(t('cron.enabledMsg'))
    loadTasks()
  } catch (err: any) {
    message.error(err.message || t('cron.enableFailed'))
  } finally {
    actionLoadingId.value = null
  }
}

async function handleDisable(id: number) {
  actionLoadingId.value = id
  try {
    await disableCronTask(id)
    message.success(t('cron.disabledMsg'))
    loadTasks()
  } catch (err: any) {
    message.error(err.message || t('cron.disableFailed'))
  } finally {
    actionLoadingId.value = null
  }
}

async function handleTrigger(id: number) {
  actionLoadingId.value = id
  try {
    await triggerCronTask(id)
    message.success(t('cron.triggeredMsg'))
  } catch (err: any) {
    message.error(err.message || t('cron.triggerFailed'))
  } finally {
    actionLoadingId.value = null
  }
}

/* ========== 创建/编辑任务对话框 ========== */
const showForm = ref(false)
const formLoading = ref(false)
const formRef = ref()
const isEdit = ref(false)
const editId = ref<number | null>(null)

const form = reactive({
  name: '',
  type: 'shell' as CronTaskType,
  cronMode: 'preset' as 'preset' | 'custom',
  cronPreset: '0 3 * * *',
  cronMinute: '0',
  cronHour: '3',
  cronDay: '*',
  cronMonth: '*',
  cronWeek: '*',
  command: '',
  target: '',
  remark: ''
})

const cronExpression = computed(() => {
  if (form.cronMode === 'preset') return form.cronPreset
  return `${form.cronMinute} ${form.cronHour} ${form.cronDay} ${form.cronMonth} ${form.cronWeek}`
})

const cronDescription = computed(() => describeCron(cronExpression.value))

const formRules = computed(() => ({
  name: {
    required: true,
    message: t('cron.nameRequired'),
    trigger: 'blur'
  },
  command: {
    required: true,
    message: t('cron.commandRequired'),
    trigger: 'blur'
  },
  target: {
    required: true,
    message: t('cron.targetRequired'),
    trigger: 'blur'
  }
}))

const backupTypeOptions = computed(() => [
  { label: t('cron.backupTypes.website'), value: 'website' },
  { label: t('cron.backupTypes.database'), value: 'database' },
  { label: t('cron.backupTypes.system'), value: 'system' }
])

function resetForm() {
  form.name = ''
  form.type = 'shell'
  form.cronMode = 'preset'
  form.cronPreset = '0 3 * * *'
  form.cronMinute = '0'
  form.cronHour = '3'
  form.cronDay = '*'
  form.cronMonth = '*'
  form.cronWeek = '*'
  form.command = ''
  form.target = ''
  form.remark = ''
}

function openCreate() {
  isEdit.value = false
  editId.value = null
  resetForm()
  showForm.value = true
}

function openEdit(id: number) {
  const task = tasks.value.find((tk) => tk.id === id)
  if (!task) return
  isEdit.value = true
  editId.value = id
  resetForm()
  form.name = task.name
  form.type = task.type
  form.command = task.command || ''
  form.target = task.target || ''
  form.remark = task.remark || ''
  // 解析 cron
  const parts = task.schedule.trim().split(/\s+/)
  if (parts.length === 5) {
    const matched = cronPresets.value.find((p) => p.value === task.schedule)
    if (matched) {
      form.cronMode = 'preset'
      form.cronPreset = task.schedule
    } else {
      form.cronMode = 'custom'
      form.cronMinute = parts[0]
      form.cronHour = parts[1]
      form.cronDay = parts[2]
      form.cronMonth = parts[3]
      form.cronWeek = parts[4]
    }
  }
  showForm.value = true
}

async function handleFormSubmit() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  const data: CreateCronTaskData = {
    name: form.name,
    type: form.type,
    schedule: cronExpression.value,
    remark: form.remark
  }

  if (form.type === 'shell') {
    if (!form.command.trim()) {
      message.warning(t('cron.commandRequired'))
      return
    }
    data.command = form.command
  } else if (form.type === 'backup') {
    if (!form.target) {
      message.warning(t('cron.selectBackupType'))
      return
    }
    data.target = form.target
  } else if (form.type === 'url') {
    if (!form.target.trim()) {
      message.warning(t('cron.urlRequired'))
      return
    }
    data.target = form.target
  } else if (form.type === 'app_restart') {
    if (!form.target.trim()) {
      message.warning(t('cron.containerRequired'))
      return
    }
    data.target = form.target
  } else if (form.type === 'virus_scan') {
    if (!form.target.trim()) {
      message.warning(t('cron.scanDirRequired'))
      return
    }
    data.target = form.target
  } else if (!builtinTypes.includes(form.type)) {
    message.warning(t('cron.unknownType'))
    return
  }

  formLoading.value = true
  try {
    if (isEdit.value && editId.value !== null) {
      await updateCronTask(editId.value, data)
      message.success(t('cron.updatedMsg'))
    } else {
      await createCronTask(data)
      message.success(t('cron.createdMsg'))
    }
    showForm.value = false
    loadTasks()
  } catch (err: any) {
    message.error(err.message || t('cron.saveFailed'))
  } finally {
    formLoading.value = false
  }
}

/* ========== 执行日志抽屉 ========== */
const showLogs = ref(false)
const logsLoading = ref(false)
const logsTaskId = ref<number | null>(null)
const logsTaskName = ref('')
const logs = ref<CronLog[]>([])
const logsTotal = ref(0)
const logsPage = ref(1)
const logsPageSize = ref(10)
const expandedLogId = ref<number | null>(null)

const logStatusMap: Record<string, StatusType> = {
  success: 'running',
  failed: 'error',
  running: 'warning'
}

function formatDuration(ms?: number): string {
  if (!ms && ms !== 0) return '—'
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

async function openLogs(id: number) {
  logsTaskId.value = id
  const task = tasks.value.find((tk) => tk.id === id)
  logsTaskName.value = task?.name || ''
  logsPage.value = 1
  expandedLogId.value = null
  showLogs.value = true
  await loadLogs()
}

async function loadLogs() {
  if (logsTaskId.value === null) return
  logsLoading.value = true
  try {
    const res = await getCronLogs(logsTaskId.value, logsPage.value, logsPageSize.value)
    if (res.data?.items) {
      logs.value = res.data.items
      logsTotal.value = res.data.total
    } else {
      const mock: CronLog[] = [
        { id: 1, startTime: '2026-09-08 03:00:00', endTime: '2026-09-08 03:02:15', durationMs: 135000, result: 'success', output: '[2026-09-08 03:00:00] Starting backup...\n[2026-09-08 03:00:01] Compressing files...\n[2026-09-08 03:02:10] Upload completed\n[2026-09-08 03:02:15] Backup finished successfully' },
        { id: 2, startTime: '2026-09-07 03:00:00', endTime: '2026-09-07 03:01:45', durationMs: 105000, result: 'success', output: '[2026-09-07 03:00:00] Starting backup...\n[2026-09-07 03:01:45] Backup finished successfully' },
        { id: 3, startTime: '2026-09-06 03:00:00', endTime: '2026-09-06 03:00:30', durationMs: 30000, result: 'failed', output: '[2026-09-06 03:00:00] Starting backup...\n[2026-09-06 03:00:30] ERROR: Connection refused to storage endpoint\n[2026-09-06 03:00:30] Backup failed' }
      ]
      logs.value = mock
      logsTotal.value = mock.length
    }
  } catch {
    logs.value = []
    logsTotal.value = 0
  } finally {
    logsLoading.value = false
  }
}

const logsPagination = computed(() => ({
  page: logsPage.value,
  pageSize: logsPageSize.value,
  itemCount: logsTotal.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
  'onUpdate:page': (p: number) => { logsPage.value = p; loadLogs() },
  'onUpdate:page-size': (ps: number) => { logsPageSize.value = ps; logsPage.value = 1; loadLogs() }
}))

const logColumns = computed<DataTableColumns>(() => [
  {
    title: t('cron.startTimeCol'),
    key: 'startTime',
    width: 160,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.startTime)
  },
  {
    title: t('cron.endTimeCol'),
    key: 'endTime',
    width: 160,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.endTime || '—')
  },
  {
    title: t('cron.durationCol'),
    key: 'durationMs',
    width: 100,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub)' }, formatDuration(row.durationMs))
  },
  {
    title: t('cron.resultCol'),
    key: 'result',
    width: 90,
    render: (row: any) => {
      const key = logStatusMap[row.result] ? row.result : 'failed'
      return h(StatusTag, { status: logStatusMap[key], text: t('cron.results.' + key) })
    }
  },
  {
    title: t('cron.outputCol'),
    key: 'output',
    render: (row: any) => {
      const isExpanded = expandedLogId.value === row.id
      const preview = (row.output || '').split('\n')[0] || '—'
      return h('div', { class: 'log-output' }, [
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          onClick: () => { expandedLogId.value = isExpanded ? null : row.id }
        }, {
          default: () => isExpanded ? t('cron.collapse') : t('cron.expand')
        }),
        isExpanded
          ? h('div', { class: 'log-output__detail' }, [
              h(CodeEditor, {
                modelValue: row.output || '',
                language: 'log',
                readOnly: true,
                height: '180px',
                minimap: false
              })
            ])
          : h('span', { class: 'log-output__preview mono' }, preview.length > 60 ? preview.slice(0, 60) + '...' : preview)
      ])
    }
  }
])

/* ========== 删除确认 ========== */
const showDelete = ref(false)
const deleteLoading = ref(false)
const deleteId = ref<number | null>(null)
const deleteName = ref('')

function openDelete(id: number) {
  deleteId.value = id
  const task = tasks.value.find((tk) => tk.id === id)
  deleteName.value = task?.name || ''
  showDelete.value = true
}

async function handleDeleteConfirm() {
  if (deleteId.value === null) return
  deleteLoading.value = true
  try {
    await deleteCronTask(deleteId.value)
    message.success(t('cron.deletedMsg'))
    showDelete.value = false
    loadTasks()
  } catch (err: any) {
    message.error(err.message || t('cron.deleteFailed'))
  } finally {
    deleteLoading.value = false
  }
}

onMounted(() => {
  loadTasks()
})
</script>

<template>
  <PageCard :title="t('cron.title')" :subtitle="t('cron.subtitle')">
    <template #extra>
      <NButton type="primary" @click="openCreate">
        <template #icon>
          <NIcon><AddOutline /></NIcon>
        </template>
        {{ t('cron.createTask') }}
      </NButton>
    </template>

    <!-- 数据表格 -->
    <DataTable
      :columns="columns"
      :data="tasks"
      :loading="loading"
      :pagination="pagination"
    />
  </PageCard>

  <!-- 创建/编辑任务对话框 -->
  <ModalForm
    :show="showForm"
    :title="isEdit ? t('cron.editTask') : t('cron.createTask')"
    :loading="formLoading"
    :width="640"
    @update:show="(v: boolean) => (showForm = v)"
    @submit="handleFormSubmit"
  >
    <NForm ref="formRef" :model="form" :rules="formRules" label-placement="top">
      <NFormItem :label="t('cron.nameLabel')" path="name">
        <NInput v-model:value="form.name" :placeholder="t('cron.namePlaceholder')" clearable />
      </NFormItem>

      <NFormItem :label="t('cron.typeLabel')" path="type">
        <NRadioGroup v-model:value="form.type">
          <NSpace size="medium">
            <NRadio value="shell">{{ t('cron.typeOptions.shell') }}</NRadio>
            <NRadio value="backup">{{ t('cron.typeOptions.backup') }}</NRadio>
            <NRadio value="url">{{ t('cron.typeOptions.url') }}</NRadio>
            <NRadio value="app_restart">{{ t('cron.typeOptions.app_restart') }}</NRadio>
            <NRadio value="ssl_check">{{ t('cron.typeOptions.ssl_check') }}</NRadio>
            <NRadio value="docker_cleanup">{{ t('cron.typeOptions.docker_cleanup') }}</NRadio>
            <NRadio value="recycle_cleanup">{{ t('cron.typeOptions.recycle_cleanup') }}</NRadio>
            <NRadio value="tamper_check">{{ t('cron.typeOptions.tamper_check') }}</NRadio>
            <NRadio value="virus_scan">{{ t('cron.typeOptions.virus_scan') }}</NRadio>
          </NSpace>
        </NRadioGroup>
      </NFormItem>

      <!-- Cron 配置 -->
      <NFormItem :label="t('cron.scheduleLabel')">
        <div class="cron-config">
          <div class="cron-config__mode">
            <NRadioGroup v-model:value="form.cronMode" size="small">
              <NSpace size="small">
                <NRadio value="preset">{{ t('cron.presetMode') }}</NRadio>
                <NRadio value="custom">{{ t('cron.customMode') }}</NRadio>
              </NSpace>
            </NRadioGroup>
          </div>

          <div v-if="form.cronMode === 'preset'" class="cron-config__preset">
            <NSelect
              v-model:value="form.cronPreset"
              :options="cronPresets"
              :placeholder="t('cron.presetPlaceholder')"
              style="width: var(--input-2xl)"
            />
          </div>

          <div v-else class="cron-config__custom">
            <NSpace size="small" align="center">
              <div class="cron-field">
                <span class="cron-field__label">{{ t('cron.minuteLabel') }}</span>
                <NInput v-model:value="form.cronMinute" placeholder="*" class="cron-field__input" />
              </div>
              <div class="cron-field">
                <span class="cron-field__label">{{ t('cron.hourLabel') }}</span>
                <NInput v-model:value="form.cronHour" placeholder="*" class="cron-field__input" />
              </div>
              <div class="cron-field">
                <span class="cron-field__label">{{ t('cron.dayLabel') }}</span>
                <NInput v-model:value="form.cronDay" placeholder="*" class="cron-field__input" />
              </div>
              <div class="cron-field">
                <span class="cron-field__label">{{ t('cron.monthLabel') }}</span>
                <NInput v-model:value="form.cronMonth" placeholder="*" class="cron-field__input" />
              </div>
              <div class="cron-field">
                <span class="cron-field__label">{{ t('cron.weekLabel') }}</span>
                <NInput v-model:value="form.cronWeek" placeholder="*" class="cron-field__input" />
              </div>
            </NSpace>
          </div>

          <div class="cron-config__preview">
            <NIcon><TimerOutline /></NIcon>
            <span class="cron-config__expr mono">{{ cronExpression }}</span>
            <span class="cron-config__desc">{{ cronDescription }}</span>
          </div>
        </div>
      </NFormItem>

      <!-- 命令/目标 -->
      <NFormItem v-if="form.type === 'shell'" :label="t('cron.commandLabel')" path="command">
        <NInput
          v-model:value="form.command"
          type="textarea"
          :placeholder="t('cron.commandPlaceholder')"
          :rows="4"
        />
      </NFormItem>

      <NFormItem v-if="form.type === 'backup'" :label="t('cron.backupTypeLabel')" path="target">
        <NSelect
          v-model:value="form.target"
          :options="backupTypeOptions"
          :placeholder="t('cron.backupPlaceholder')"
          clearable
        />
      </NFormItem>

      <NFormItem v-if="form.type === 'url'" :label="t('cron.urlLabel')" path="target">
        <NInput
          v-model:value="form.target"
          :placeholder="t('cron.urlPlaceholder')"
          clearable
        />
      </NFormItem>

      <NFormItem v-if="form.type === 'app_restart'" :label="t('cron.containerLabel')" path="target">
        <NInput
          v-model:value="form.target"
          :placeholder="t('cron.containerPlaceholder')"
          clearable
        />
      </NFormItem>

      <NFormItem v-if="form.type === 'virus_scan'" :label="t('cron.scanDirLabel')" path="target">
        <NInput
          v-model:value="form.target"
          :placeholder="t('cron.scanDirPlaceholder')"
          clearable
        />
      </NFormItem>

      <NFormItem v-if="builtinTypes.includes(form.type)" :label="t('cron.paramsLabel')">
        <div class="cron-builtin-hint">{{ t('cron.builtinHint') }}</div>
      </NFormItem>

      <NFormItem :label="t('cron.remarkLabel')" path="remark">
        <NInput v-model:value="form.remark" :placeholder="t('cron.remarkPlaceholder')" clearable />
      </NFormItem>
    </NForm>
  </ModalForm>

  <!-- 执行日志抽屉 -->
  <NDrawer
    :show="showLogs"
    :width="780"
    placement="right"
    @update:show="(v: boolean) => (showLogs = v)"
  >
    <NDrawerContent
      :title="t('cron.logsTitle', { name: logsTaskName })"
      closable
    >
      <div class="logs-drawer">
        <DataTable
          :columns="logColumns"
          :data="logs"
          :loading="logsLoading"
          :pagination="logsPagination"
          size="small"
        />
      </div>
    </NDrawerContent>
  </NDrawer>

  <!-- 删除确认 -->
  <ConfirmDialog
    :show="showDelete"
    :title="t('cron.deleteTitle')"
    :content="t('cron.deleteContent', { name: deleteName })"
    :confirm-text="t('cron.delete')"
    type="error"
    :loading="deleteLoading"
    @update:show="(v: boolean) => (showDelete = v)"
    @confirm="handleDeleteConfirm"
  />
</template>

<style scoped>
.cron-name {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.cron-name__main {
  font-weight: 500;
  color: var(--ink);
}

.cron-name__remark {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
}

.cron-expr {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.cron-expr__code {
  font-weight: 500;
  color: var(--ink);
  font-size: var(--fs-mono);
}

.cron-expr__desc {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
}

/* Cron 配置区 */
.cron-config {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.cron-config__preset,
.cron-config__custom {
  width: 100%;
}

.cron-field {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-1);
}

.cron-field__label {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
}

.cron-field__input {
  width: 70px;
  text-align: center;
}

.cron-field__input :deep(.n-input__input-el) {
  text-align: center;
  font-family: var(--font-mono);
}

.cron-config__preview {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-4);
  background: var(--surface2);
  border-radius: var(--radius-md);
  border: 1px solid var(--line);
}

.cron-config__preview .n-icon {
  color: var(--acc);
  flex-shrink: 0;
}

.cron-config__expr {
  font-weight: 600;
  color: var(--ink);
  font-size: var(--fs-mono);
}

.cron-config__desc {
  font-size: var(--fs-caption);
  color: var(--sub);
  margin-left: var(--space-2);
  padding-left: var(--space-3);
  border-left: 1px solid var(--line);
}

/* 日志抽屉 */
.logs-drawer {
  height: 100%;
}

.log-output {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.log-output__preview {
  font-size: var(--fs-caption);
  color: var(--sub);
  max-width: 400px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.log-output__detail {
  width: 100%;
}

.cron-builtin-hint {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
  line-height: 1.6;
}
</style>
