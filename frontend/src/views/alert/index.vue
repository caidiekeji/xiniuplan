<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue'
import {
  NButton, NIcon, NForm, NFormItem, NInput, NSelect, NCheckboxGroup,
  NCheckbox, NSpace, NTag, NSwitch, NTabs, NTabPane
} from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import {
  AddOutline, PlayOutline, StopOutline, CreateOutline,
  TrashOutline, FlashOutline, ChevronDownOutline, ChevronForwardOutline
} from '@vicons/ionicons5'
import {
  PageCard, DataTable, ModalForm, StatusTag, ConfirmDialog
} from '@/components'
import type { StatusType } from '@/components'
import {
  getAlertRules, createAlertRule, updateAlertRule, deleteAlertRule,
  enableAlertRule, disableAlertRule, testAlertRule,
  getAlertChannels, createAlertChannel, updateAlertChannel, deleteAlertChannel,
  testAlertChannel, getAlertHistory,
  type AlertRule, type AlertChannel, type AlertHistory,
  type MetricType, type AlertCondition, type ChannelType,
  type AlertStatus, type NotificationStatus,
  type CreateAlertRuleData, type CreateAlertChannelData
} from '@/api/alert'
import type { DataTableColumns } from 'naive-ui'
import { useI18n } from 'vue-i18n'

const message = useMessage()
const { t } = useI18n()

/* ==================== 常量映射 ==================== */

const metricLabelMap = computed<Record<MetricType, string>>(() => ({
  cpu: t('alert.metrics.cpu'),
  memory: t('alert.metrics.memory'),
  disk: t('alert.metrics.disk'),
  network: t('alert.metrics.network')
}))

const metricTagTypeMap: Record<MetricType, 'info' | 'success' | 'warning' | 'error' | 'default'> = {
  cpu: 'info',
  memory: 'success',
  disk: 'warning',
  network: 'default'
}

const metricUnitMap: Record<MetricType, string> = {
  cpu: '%',
  memory: '%',
  disk: '%',
  network: 'MB/s'
}

const conditionLabelMap = computed<Record<AlertCondition, string>>(() => ({
  gt: t('alert.conditions.gt'),
  lt: t('alert.conditions.lt'),
  eq: t('alert.conditions.eq')
}))

const conditionSymbolMap: Record<AlertCondition, string> = {
  gt: '>',
  lt: '<',
  eq: '='
}

const channelLabelMap = computed<Record<ChannelType, string>>(() => ({
  email: t('alert.channels.email'),
  dingtalk: t('alert.channels.dingtalk'),
  wechat: t('alert.channels.wechat'),
  feishu: t('alert.channels.feishu'),
  telegram: t('alert.channels.telegram'),
  webhook: t('alert.channels.webhook')
}))

const channelTagTypeMap: Record<ChannelType, 'info' | 'success' | 'warning' | 'error' | 'default'> = {
  email: 'info',
  dingtalk: 'success',
  wechat: 'warning',
  feishu: 'default',
  telegram: 'error',
  webhook: 'info'
}

const historyStatusMap = computed<Record<AlertStatus, { status: StatusType; text: string }>>(() => ({
  triggered: { status: 'error', text: t('alert.ruleStatus.triggered') },
  recovered: { status: 'running', text: t('alert.ruleStatus.recovered') }
}))

const notificationStatusMap = computed<Record<NotificationStatus, { status: StatusType; text: string }>>(() => ({
  pending: { status: 'warning', text: t('alert.notifyStatus.pending') },
  sent: { status: 'running', text: t('alert.notifyStatus.sent') },
  failed: { status: 'error', text: t('alert.notifyStatus.failed') }
}))

const metricOptions = computed(() =>
  Object.entries(metricLabelMap.value).map(([value, label]) => ({ label, value }))
)
const conditionOptions = computed(() =>
  Object.entries(conditionLabelMap.value).map(([value, label]) => ({ label, value }))
)
const channelTypeOptions = computed(() =>
  Object.entries(channelLabelMap.value).map(([value, label]) => ({ label, value }))
)
const historyStatusOptions = computed<Array<{ label: string; value: string }>>(() => [
  { label: t('alert.ruleStatus.all'), value: '' },
  { label: t('alert.ruleStatus.triggered'), value: 'triggered' },
  { label: t('alert.ruleStatus.recovered'), value: 'recovered' }
])

/* ==================== Tab 1: 告警规则 ==================== */

const rulesLoading = ref(false)
const rules = ref<AlertRule[]>([])
const rulesTotal = ref(0)
const rulesPage = ref(1)
const rulesPageSize = ref(20)
const ruleActionLoadingId = ref<number | null>(null)

// 渠道列表（供规则表单多选）
const channelList = ref<AlertChannel[]>([])

async function loadChannels() {
  try {
    const res = await getAlertChannels()
    if (res.data?.items) {
      channelList.value = res.data.items
    }
  } catch {
    channelList.value = []
  }
}

function getChannelName(id: string): string {
  const ch = channelList.value.find((c) => String(c.id) === String(id))
  return ch?.name || id
}

function describeRule(row: AlertRule): string {
  const metric = metricLabelMap.value[row.metric_type]
  const symbol = conditionSymbolMap[row.condition]
  const unit = metricUnitMap[row.metric_type]
  return `${metric} ${symbol} ${row.threshold}${unit}`
}

async function loadRules() {
  rulesLoading.value = true
  try {
    const res = await getAlertRules(rulesPage.value, rulesPageSize.value)
    if (res.data?.items) {
      rules.value = res.data.items
      rulesTotal.value = res.data.total
    } else {
      rules.value = []
      rulesTotal.value = 0
    }
  } catch {
    rules.value = []
    rulesTotal.value = 0
  } finally {
    rulesLoading.value = false
  }
}

const rulesPagination = computed(() => ({
  page: rulesPage.value,
  pageSize: rulesPageSize.value,
  itemCount: rulesTotal.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  'onUpdate:page': (p: number) => { rulesPage.value = p; loadRules() },
  'onUpdate:page-size': (ps: number) => { rulesPageSize.value = ps; rulesPage.value = 1; loadRules() }
}))

const ruleColumns = computed<DataTableColumns>(() => [
  {
    title: t('alert.rules.columns.name'),
    key: 'name',
    width: 180,
    render: (row: any) => h('div', { class: 'rule-name' }, [
      h('span', { class: 'rule-name__main' }, row.name),
      row.remark
        ? h('span', { class: 'rule-name__remark' }, row.remark)
        : null
    ])
  },
  {
    title: t('alert.rules.columns.metric'),
    key: 'metric_type',
    width: 100,
    render: (row: any) => h(NTag, {
      size: 'small',
      type: metricTagTypeMap[row.metric_type as MetricType],
      bordered: false,
      round: true
    }, { default: () => metricLabelMap.value[row.metric_type as MetricType] })
  },
  {
    title: t('alert.rules.columns.condition'),
    key: 'condition',
    width: 160,
    render: (row: any) => h('span', { class: 'mono rule-condition' }, describeRule(row as AlertRule))
  },
  {
    title: t('alert.rules.columns.duration'),
    key: 'duration',
    width: 100,
    render: (row: any) => h('span', { class: 'mono' }, t('alert.durationUnit', { duration: row.duration }))
  },
  {
    title: t('alert.rules.columns.channels'),
    key: 'channels',
    width: 200,
    render: (row: any) => {
      const ids: string[] = row.channels || []
      if (ids.length === 0) return h('span', { style: 'color: var(--text-quaternary);font-size:var(--fs-caption)' }, '—')
      return h(NSpace, { size: 'small', wrap: true }, {
        default: () => ids.map((id) => h(NTag, {
          size: 'small',
          type: 'default',
          bordered: false,
          round: true
        }, { default: () => getChannelName(id) }))
      })
    }
  },
  {
    title: t('alert.rules.columns.status'),
    key: 'enabled',
    width: 100,
    render: (row: any) => h(StatusTag, {
      status: row.enabled ? 'running' : 'stopped',
      text: row.enabled ? t('alert.status.enabled') : t('alert.status.disabled')
    })
  },
  {
    title: t('alert.rules.columns.actions'),
    key: 'actions',
    width: 280,
    render: (row: any) => {
      const isEnabled = row.enabled
      const actionLoading = ruleActionLoadingId.value === row.id
      return h(NSpace, { size: 'small', wrap: false }, {
        default: () => [
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            type: isEnabled ? 'warning' : 'success',
            loading: actionLoading,
            onClick: () => isEnabled ? handleDisableRule(row.id) : handleEnableRule(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(isEnabled ? StopOutline : PlayOutline) }),
            default: () => isEnabled ? t('alert.status.disabled') : t('alert.status.enabled')
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            loading: actionLoading,
            onClick: () => handleTestRule(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(FlashOutline) }),
            default: () => t('alert.actions.test')
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            onClick: () => openRuleEdit(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(CreateOutline) }),
            default: () => t('alert.actions.edit')
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            type: 'error',
            loading: actionLoading,
            onClick: () => openRuleDelete(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
            default: () => t('alert.actions.delete')
          })
        ]
      })
    }
  }
])

async function handleEnableRule(id: number) {
  ruleActionLoadingId.value = id
  try {
    await enableAlertRule(id)
    message.success(t('alert.rules.messages.enabled'))
    loadRules()
  } catch (err: any) {
    message.error(err.message || t('alert.rules.messages.enableFailed'))
  } finally {
    ruleActionLoadingId.value = null
  }
}

async function handleDisableRule(id: number) {
  ruleActionLoadingId.value = id
  try {
    await disableAlertRule(id)
    message.success(t('alert.rules.messages.disabled'))
    loadRules()
  } catch (err: any) {
    message.error(err.message || t('alert.rules.messages.disableFailed'))
  } finally {
    ruleActionLoadingId.value = null
  }
}

async function handleTestRule(id: number) {
  ruleActionLoadingId.value = id
  try {
    await testAlertRule(id)
    message.success(t('alert.rules.messages.tested'))
  } catch (err: any) {
    message.error(err.message || t('alert.rules.messages.testFailed'))
  } finally {
    ruleActionLoadingId.value = null
  }
}

/* --- 规则创建/编辑对话框 --- */

const showRuleForm = ref(false)
const ruleFormLoading = ref(false)
const ruleFormRef = ref()
const ruleIsEdit = ref(false)
const ruleEditId = ref<number | null>(null)

const ruleForm = reactive({
  name: '',
  metric_type: 'cpu' as MetricType,
  condition: 'gt' as AlertCondition,
  threshold: '80',
  duration: '5',
  channels: [] as string[],
  remark: ''
})

const ruleFormRules = {
  name: { required: true, message: t('alert.rules.messages.nameRequired'), trigger: 'blur' },
  threshold: { required: true, message: t('alert.rules.messages.thresholdRequired'), trigger: 'blur' }
}

function resetRuleForm() {
  ruleForm.name = ''
  ruleForm.metric_type = 'cpu'
  ruleForm.condition = 'gt'
  ruleForm.threshold = '80'
  ruleForm.duration = '5'
  ruleForm.channels = []
  ruleForm.remark = ''
}

function openRuleCreate() {
  ruleIsEdit.value = false
  ruleEditId.value = null
  resetRuleForm()
  showRuleForm.value = true
}

function openRuleEdit(id: number) {
  const rule = rules.value.find((r) => r.id === id)
  if (!rule) return
  ruleIsEdit.value = true
  ruleEditId.value = id
  resetRuleForm()
  ruleForm.name = rule.name
  ruleForm.metric_type = rule.metric_type
  ruleForm.condition = rule.condition
  ruleForm.threshold = String(rule.threshold)
  ruleForm.duration = String(rule.duration)
  ruleForm.channels = [...rule.channels]
  ruleForm.remark = rule.remark || ''
  showRuleForm.value = true
}

async function handleRuleFormSubmit() {
  try {
    await ruleFormRef.value?.validate()
  } catch {
    return
  }

  const data: CreateAlertRuleData = {
    name: ruleForm.name,
    metric_type: ruleForm.metric_type,
    condition: ruleForm.condition,
    threshold: Number(ruleForm.threshold),
    duration: Number(ruleForm.duration),
    channels: ruleForm.channels,
    remark: ruleForm.remark
  }

  ruleFormLoading.value = true
  try {
    if (ruleIsEdit.value && ruleEditId.value !== null) {
      await updateAlertRule(ruleEditId.value, data)
      message.success(t('alert.rules.messages.updated'))
    } else {
      await createAlertRule(data)
      message.success(t('alert.rules.messages.created'))
    }
    showRuleForm.value = false
    loadRules()
  } catch (err: any) {
    message.error(err.message || t('alert.rules.messages.saveFailed'))
  } finally {
    ruleFormLoading.value = false
  }
}

/* --- 规则删除确认 --- */

const showRuleDelete = ref(false)
const ruleDeleteLoading = ref(false)
const ruleDeleteId = ref<number | null>(null)
const ruleDeleteName = ref('')

function openRuleDelete(id: number) {
  ruleDeleteId.value = id
  const rule = rules.value.find((r) => r.id === id)
  ruleDeleteName.value = rule?.name || ''
  showRuleDelete.value = true
}

async function handleRuleDeleteConfirm() {
  if (ruleDeleteId.value === null) return
  ruleDeleteLoading.value = true
  try {
    await deleteAlertRule(ruleDeleteId.value)
    message.success(t('alert.rules.messages.deleted'))
    showRuleDelete.value = false
    loadRules()
  } catch (err: any) {
    message.error(err.message || t('alert.rules.messages.deleteFailed'))
  } finally {
    ruleDeleteLoading.value = false
  }
}

/* ==================== Tab 2: 通知渠道 ==================== */

const channelsLoading = ref(false)
const channels = ref<AlertChannel[]>([])
const channelsTotal = ref(0)
const channelsPage = ref(1)
const channelsPageSize = ref(20)
const channelActionLoadingId = ref<number | null>(null)

async function loadChannelsList() {
  channelsLoading.value = true
  try {
    const res = await getAlertChannels()
    if (res.data?.items) {
      channels.value = res.data.items
      channelsTotal.value = res.data.total
    } else {
      channels.value = []
      channelsTotal.value = 0
    }
  } catch {
    channels.value = []
    channelsTotal.value = 0
  } finally {
    channelsLoading.value = false
  }
}

const channelsPagination = computed(() => ({
  page: channelsPage.value,
  pageSize: channelsPageSize.value,
  itemCount: channelsTotal.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  'onUpdate:page': (p: number) => { channelsPage.value = p; loadChannelsList() },
  'onUpdate:page-size': (ps: number) => { channelsPageSize.value = ps; channelsPage.value = 1; loadChannelsList() }
}))

const channelColumns = computed<DataTableColumns>(() => [
  {
    title: t('alert.channelsTab.columns.name'),
    key: 'name',
    width: 180,
    render: (row: any) => h('span', { class: 'channel-name' }, row.name)
  },
  {
    title: t('alert.channelsTab.columns.type'),
    key: 'type',
    width: 120,
    render: (row: any) => h(NTag, {
      size: 'small',
      type: channelTagTypeMap[row.type as ChannelType],
      bordered: false,
      round: true
    }, { default: () => channelLabelMap.value[row.type as ChannelType] })
  },
  {
    title: t('alert.channelsTab.columns.status'),
    key: 'enabled',
    width: 100,
    render: (row: any) => h(StatusTag, {
      status: row.enabled ? 'running' : 'stopped',
      text: row.enabled ? t('alert.status.enabled') : t('alert.status.disabled')
    })
  },
  {
    title: t('alert.channelsTab.columns.createdAt'),
    key: 'created_at',
    width: 180,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.created_at || '—')
  },
  {
    title: t('alert.channelsTab.columns.actions'),
    key: 'actions',
    width: 200,
    render: (row: any) => {
      const actionLoading = channelActionLoadingId.value === row.id
      return h(NSpace, { size: 'small', wrap: false }, {
        default: () => [
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            loading: actionLoading,
            onClick: () => handleTestChannel(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(FlashOutline) }),
            default: () => t('alert.actions.test')
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            onClick: () => openChannelEdit(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(CreateOutline) }),
            default: () => t('alert.actions.edit')
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            type: 'error',
            loading: actionLoading,
            onClick: () => openChannelDelete(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
            default: () => t('alert.actions.delete')
          })
        ]
      })
    }
  }
])

async function handleTestChannel(id: number) {
  channelActionLoadingId.value = id
  try {
    await testAlertChannel(id)
    message.success(t('alert.rules.messages.tested'))
  } catch (err: any) {
    message.error(err.message || t('alert.rules.messages.testFailed'))
  } finally {
    channelActionLoadingId.value = null
  }
}

/* --- 渠道创建/编辑对话框 --- */

const showChannelForm = ref(false)
const channelFormLoading = ref(false)
const channelFormRef = ref()
const channelIsEdit = ref(false)
const channelEditId = ref<number | null>(null)

const channelForm = reactive({
  name: '',
  type: 'email' as ChannelType,
  enabled: true,
  // 动态配置字段
  email_recipients: '',
  dingtalk_webhook: '',
  dingtalk_secret: '',
  wechat_webhook: '',
  feishu_webhook: '',
  telegram_bot_token: '',
  telegram_chat_id: '',
  webhook_url: '',
  webhook_headers: ''
})

const channelFormRules = {
  name: { required: true, message: t('alert.channelsTab.messages.nameRequired'), trigger: 'blur' }
}

function resetChannelForm() {
  channelForm.name = ''
  channelForm.type = 'email'
  channelForm.enabled = true
  channelForm.email_recipients = ''
  channelForm.dingtalk_webhook = ''
  channelForm.dingtalk_secret = ''
  channelForm.wechat_webhook = ''
  channelForm.feishu_webhook = ''
  channelForm.telegram_bot_token = ''
  channelForm.telegram_chat_id = ''
  channelForm.webhook_url = ''
  channelForm.webhook_headers = ''
}

function buildChannelConfig(): Record<string, any> {
  switch (channelForm.type) {
    case 'email':
      return { recipients: channelForm.email_recipients }
    case 'dingtalk':
      return {
        webhook: channelForm.dingtalk_webhook,
        secret: channelForm.dingtalk_secret || undefined
      }
    case 'wechat':
      return { webhook: channelForm.wechat_webhook }
    case 'feishu':
      return { webhook: channelForm.feishu_webhook }
    case 'telegram':
      return {
        bot_token: channelForm.telegram_bot_token,
        chat_id: channelForm.telegram_chat_id
      }
    case 'webhook':
      return {
        url: channelForm.webhook_url,
        headers: channelForm.webhook_headers || undefined
      }
    default:
      return {}
  }
}

function loadChannelConfig(ch: AlertChannel) {
  const cfg = ch.config || {}
  switch (ch.type) {
    case 'email':
      channelForm.email_recipients = cfg.recipients || ''
      break
    case 'dingtalk':
      channelForm.dingtalk_webhook = cfg.webhook || ''
      channelForm.dingtalk_secret = cfg.secret || ''
      break
    case 'wechat':
      channelForm.wechat_webhook = cfg.webhook || ''
      break
    case 'feishu':
      channelForm.feishu_webhook = cfg.webhook || ''
      break
    case 'telegram':
      channelForm.telegram_bot_token = cfg.bot_token || ''
      channelForm.telegram_chat_id = cfg.chat_id || ''
      break
    case 'webhook':
      channelForm.webhook_url = cfg.url || ''
      channelForm.webhook_headers = cfg.headers || ''
      break
  }
}

function openChannelCreate() {
  channelIsEdit.value = false
  channelEditId.value = null
  resetChannelForm()
  showChannelForm.value = true
}

function openChannelEdit(id: number) {
  const ch = channels.value.find((c) => c.id === id)
  if (!ch) return
  channelIsEdit.value = true
  channelEditId.value = id
  resetChannelForm()
  channelForm.name = ch.name
  channelForm.type = ch.type
  channelForm.enabled = ch.enabled
  loadChannelConfig(ch)
  showChannelForm.value = true
}

async function handleChannelFormSubmit() {
  try {
    await channelFormRef.value?.validate()
  } catch {
    return
  }

  const data: CreateAlertChannelData = {
    name: channelForm.name,
    type: channelForm.type,
    config: buildChannelConfig(),
    enabled: channelForm.enabled
  }

  channelFormLoading.value = true
  try {
    if (channelIsEdit.value && channelEditId.value !== null) {
      await updateAlertChannel(channelEditId.value, data)
      message.success(t('alert.channelsTab.messages.updated'))
    } else {
      await createAlertChannel(data)
      message.success(t('alert.channelsTab.messages.created'))
    }
    showChannelForm.value = false
    loadChannelsList()
    loadChannels()
  } catch (err: any) {
    message.error(err.message || t('alert.channelsTab.messages.saveFailed'))
  } finally {
    channelFormLoading.value = false
  }
}

/* --- 渠道删除确认 --- */

const showChannelDelete = ref(false)
const channelDeleteLoading = ref(false)
const channelDeleteId = ref<number | null>(null)
const channelDeleteName = ref('')

function openChannelDelete(id: number) {
  channelDeleteId.value = id
  const ch = channels.value.find((c) => c.id === id)
  channelDeleteName.value = ch?.name || ''
  showChannelDelete.value = true
}

async function handleChannelDeleteConfirm() {
  if (channelDeleteId.value === null) return
  channelDeleteLoading.value = true
  try {
    await deleteAlertChannel(channelDeleteId.value)
    message.success(t('alert.channelsTab.messages.deleted'))
    showChannelDelete.value = false
    loadChannelsList()
    loadChannels()
  } catch (err: any) {
    message.error(err.message || t('alert.channelsTab.messages.deleteFailed'))
  } finally {
    channelDeleteLoading.value = false
  }
}

/* ==================== Tab 3: 告警历史 ==================== */

const historyLoading = ref(false)
const history = ref<AlertHistory[]>([])
const historyTotal = ref(0)
const historyPage = ref(1)
const historyPageSize = ref(20)
const historyFilterRule = ref<string | null>('')
const historyFilterStatus = ref<string | null>('')
const expandedHistoryId = ref<number | null>(null)

// 规则下拉选项（供历史筛选）
const ruleOptions = computed(() => [
  { label: t('alert.history.allRules'), value: '' },
  ...rules.value.map((r) => ({ label: r.name, value: String(r.id) }))
])

async function loadHistory() {
  historyLoading.value = true
  try {
    const params: any = {
      page: historyPage.value,
      pageSize: historyPageSize.value
    }
    if (historyFilterRule.value) params.rule_id = Number(historyFilterRule.value)
    if (historyFilterStatus.value) params.status = historyFilterStatus.value
    const res = await getAlertHistory(params)
    if (res.data?.items) {
      history.value = res.data.items
      historyTotal.value = res.data.total
    } else {
      history.value = []
      historyTotal.value = 0
    }
  } catch {
    history.value = []
    historyTotal.value = 0
  } finally {
    historyLoading.value = false
  }
}

function handleHistoryFilter() {
  historyPage.value = 1
  loadHistory()
}

function handleHistoryReset() {
  historyFilterRule.value = ''
  historyFilterStatus.value = ''
  historyPage.value = 1
  loadHistory()
}

const historyPagination = computed(() => ({
  page: historyPage.value,
  pageSize: historyPageSize.value,
  itemCount: historyTotal.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  'onUpdate:page': (p: number) => { historyPage.value = p; loadHistory() },
  'onUpdate:page-size': (ps: number) => { historyPageSize.value = ps; historyPage.value = 1; loadHistory() }
}))

const historyColumns = computed<DataTableColumns>(() => [
  {
    title: '',
    key: 'expand',
    width: 40,
    render: (row: any) => {
      const hasError = !!row.notification_error
      if (!hasError) return null
      const isExpanded = expandedHistoryId.value === row.id
      return h(NButton, {
        size: 'tiny',
        quaternary: true,
        circle: true,
        onClick: () => { expandedHistoryId.value = isExpanded ? null : row.id }
      }, {
        icon: () => h(NIcon, null, { default: () => h(isExpanded ? ChevronDownOutline : ChevronForwardOutline) })
      })
    }
  },
  {
    title: t('alert.history.columns.triggeredAt'),
    key: 'triggered_at',
    width: 170,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.triggered_at || '—')
  },
  {
    title: t('alert.history.columns.ruleName'),
    key: 'rule_name',
    width: 160,
    render: (row: any) => h('span', { class: 'history-rule-name' }, row.rule_name)
  },
  {
    title: t('alert.history.columns.metric'),
    key: 'metric_type',
    width: 90,
    render: (row: any) => h(NTag, {
      size: 'small',
      type: metricTagTypeMap[row.metric_type as MetricType],
      bordered: false,
      round: true
    }, { default: () => metricLabelMap.value[row.metric_type as MetricType] })
  },
  {
    title: t('alert.history.columns.triggerValue'),
    key: 'trigger_value',
    width: 100,
    render: (row: any) => h('span', { class: 'mono history-value' }, `${Number(row.trigger_value).toFixed(1)}${metricUnitMap[row.metric_type as MetricType]}`)
  },
  {
    title: t('alert.history.columns.threshold'),
    key: 'threshold',
    width: 90,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub)' }, `${row.threshold}${metricUnitMap[row.metric_type as MetricType]}`)
  },
  {
    title: t('alert.history.columns.status'),
    key: 'status',
    width: 100,
    render: (row: any) => {
      const cfg = historyStatusMap.value[row.status as AlertStatus]
      return h(StatusTag, { status: cfg.status, text: cfg.text })
    }
  },
  {
    title: t('alert.history.columns.notifyStatus'),
    key: 'notification_status',
    width: 100,
    render: (row: any) => {
      const cfg = notificationStatusMap.value[row.notification_status as NotificationStatus]
      return h(StatusTag, { status: cfg.status, text: cfg.text })
    }
  },
  {
    title: t('alert.history.columns.recoveredAt'),
    key: 'recovered_at',
    width: 170,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);font-size:var(--fs-caption)' }, row.recovered_at || '—')
  }
])

// 展开行内容（notification_error）

/* ==================== 生命周期 ==================== */

onMounted(() => {
  loadChannels()
  loadRules()
  loadChannelsList()
  loadHistory()
})
</script>

<template>
  <NTabs type="line" animated>
    <!-- ==================== Tab 1: 告警规则 ==================== -->
    <NTabPane name="rules" :tab="t('alert.tabs.rules')">
      <PageCard :title="t('alert.rules.title')" :subtitle="t('alert.rules.subtitle')">
        <template #extra>
          <NButton type="primary" @click="openRuleCreate">
            <template #icon>
              <NIcon><AddOutline /></NIcon>
            </template>
            {{ t('alert.rules.create') }}
          </NButton>
        </template>

        <DataTable
          :columns="ruleColumns"
          :data="rules"
          :loading="rulesLoading"
          :pagination="rulesPagination"
        />
      </PageCard>
    </NTabPane>

    <!-- ==================== Tab 2: 通知渠道 ==================== -->
    <NTabPane name="channels" :tab="t('alert.tabs.channels')">
      <PageCard :title="t('alert.channelsTab.title')" :subtitle="t('alert.channelsTab.subtitle')">
        <template #extra>
          <NButton type="primary" @click="openChannelCreate">
            <template #icon>
              <NIcon><AddOutline /></NIcon>
            </template>
            {{ t('alert.channelsTab.create') }}
          </NButton>
        </template>

        <DataTable
          :columns="channelColumns"
          :data="channels"
          :loading="channelsLoading"
          :pagination="channelsPagination"
        />
      </PageCard>
    </NTabPane>

    <!-- ==================== Tab 3: 告警历史 ==================== -->
    <NTabPane name="history" :tab="t('alert.tabs.history')">
      <PageCard :title="t('alert.history.title')" :subtitle="t('alert.history.subtitle')">
        <!-- 筛选栏 -->
        <div class="history-filter">
          <NSpace size="medium" align="center">
            <div class="history-filter__item">
              <span class="history-filter__label">{{ t('alert.history.filterRule') }}</span>
              <NSelect
                v-model:value="historyFilterRule"
                :options="ruleOptions"
                :placeholder="t('alert.history.filterRulePlaceholder')"
                clearable
                style="width: 200px"
                @update:value="handleHistoryFilter"
              />
            </div>
            <div class="history-filter__item">
              <span class="history-filter__label">{{ t('alert.history.filterStatus') }}</span>
              <NSelect
                v-model:value="historyFilterStatus"
                :options="historyStatusOptions"
                :placeholder="t('alert.history.filterStatusPlaceholder')"
                clearable
                style="width: 140px"
                @update:value="handleHistoryFilter"
              />
            </div>
            <NButton size="small" @click="handleHistoryReset">{{ t('alert.history.reset') }}</NButton>
          </NSpace>
        </div>

        <DataTable
          :columns="historyColumns"
          :data="history"
          :loading="historyLoading"
          :pagination="historyPagination"
        >
          <template #empty>
            <div class="history-empty">
              <p>{{ t('alert.history.empty') }}</p>
              <span>{{ t('alert.history.emptyHint') }}</span>
            </div>
          </template>
        </DataTable>

        <!-- 展开行错误信息 -->
        <div v-for="row in history" :key="`expand-${row.id}`">
          <div v-if="expandedHistoryId === row.id && row.notification_error" class="history-expand-row">
            <span class="history-expand__label">{{ t('alert.history.errorLabel') }}</span>
            <span class="history-expand__error mono">{{ row.notification_error }}</span>
          </div>
        </div>
      </PageCard>
    </NTabPane>
  </NTabs>

  <!-- ==================== 规则创建/编辑对话框 ==================== -->
  <ModalForm
    :show="showRuleForm"
    :title="ruleIsEdit ? t('alert.rules.form.editTitle') : t('alert.rules.form.createTitle')"
    :loading="ruleFormLoading"
    :width="560"
    @update:show="(v: boolean) => (showRuleForm = v)"
    @submit="handleRuleFormSubmit"
  >
    <NForm ref="ruleFormRef" :model="ruleForm" :rules="ruleFormRules" label-placement="top">
      <NFormItem :label="t('alert.rules.form.name')" path="name">
        <NInput v-model:value="ruleForm.name" :placeholder="t('alert.rules.form.namePlaceholder')" clearable />
      </NFormItem>

      <NSpace size="medium" align="end" style="width: 100%">
        <NFormItem :label="t('alert.rules.form.metric')" path="metric_type" style="flex: 1">
          <NSelect v-model:value="ruleForm.metric_type" :options="metricOptions" />
        </NFormItem>
        <NFormItem :label="t('alert.rules.form.condition')" path="condition" style="flex: 1">
          <NSelect v-model:value="ruleForm.condition" :options="conditionOptions" />
        </NFormItem>
        <NFormItem :label="t('alert.rules.form.threshold')" path="threshold" style="flex: 1">
          <NInput
            v-model:value="ruleForm.threshold"
            :placeholder="t('alert.rules.form.thresholdPlaceholder', { unit: metricUnitMap[ruleForm.metric_type] })"
          />
        </NFormItem>
      </NSpace>

      <NFormItem :label="t('alert.rules.form.duration')" path="duration">
        <NInput v-model:value="ruleForm.duration" :placeholder="t('alert.rules.form.durationPlaceholder')" />
      </NFormItem>

      <NFormItem :label="t('alert.rules.form.channels')">
        <NCheckboxGroup v-model:value="ruleForm.channels">
          <NSpace size="medium" wrap>
            <NCheckbox v-for="ch in channelList" :key="ch.id" :value="String(ch.id)">
              {{ ch.name }}
              <NTag size="tiny" type="default" :bordered="false" round style="margin-left: var(--space-1)">
                {{ channelLabelMap[ch.type] }}
              </NTag>
            </NCheckbox>
          </NSpace>
        </NCheckboxGroup>
        <div v-if="channelList.length === 0" class="form-hint">
          {{ t('alert.rules.form.noChannels') }}
        </div>
      </NFormItem>

      <NFormItem :label="t('alert.rules.form.remark')" path="remark">
        <NInput v-model:value="ruleForm.remark" :placeholder="t('alert.rules.form.remarkPlaceholder')" clearable />
      </NFormItem>
    </NForm>
  </ModalForm>

  <!-- ==================== 渠道创建/编辑对话框 ==================== -->
  <ModalForm
    :show="showChannelForm"
    :title="channelIsEdit ? t('alert.channelsTab.form.editTitle') : t('alert.channelsTab.form.createTitle')"
    :loading="channelFormLoading"
    :width="560"
    @update:show="(v: boolean) => (showChannelForm = v)"
    @submit="handleChannelFormSubmit"
  >
    <NForm ref="channelFormRef" :model="channelForm" :rules="channelFormRules" label-placement="top">
      <NFormItem :label="t('alert.channelsTab.form.name')" path="name">
        <NInput v-model:value="channelForm.name" :placeholder="t('alert.channelsTab.form.namePlaceholder')" clearable />
      </NFormItem>

      <NFormItem :label="t('alert.channelsTab.form.type')" path="type">
        <NSelect v-model:value="channelForm.type" :options="channelTypeOptions" />
      </NFormItem>

      <!-- 邮件 -->
      <template v-if="channelForm.type === 'email'">
        <NFormItem :label="t('alert.channelsTab.form.emailRecipients')">
          <NInput
            v-model:value="channelForm.email_recipients"
            type="textarea"
            placeholder="admin@example.com, ops@example.com"
            :rows="2"
          />
        </NFormItem>
        <div class="form-hint">{{ t('alert.channelsTab.form.emailHint') }}</div>
      </template>

      <!-- 钉钉 -->
      <template v-if="channelForm.type === 'dingtalk'">
        <NFormItem label="Webhook URL">
          <NInput v-model:value="channelForm.dingtalk_webhook" placeholder="https://oapi.dingtalk.com/robot/send?access_token=..." />
        </NFormItem>
        <NFormItem :label="t('alert.channelsTab.form.dingtalkSecret')">
          <NInput v-model:value="channelForm.dingtalk_secret" placeholder="SEC..." />
        </NFormItem>
      </template>

      <!-- 企业微信 -->
      <template v-if="channelForm.type === 'wechat'">
        <NFormItem label="Webhook URL">
          <NInput v-model:value="channelForm.wechat_webhook" placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=..." />
        </NFormItem>
      </template>

      <!-- 飞书 -->
      <template v-if="channelForm.type === 'feishu'">
        <NFormItem label="Webhook URL">
          <NInput v-model:value="channelForm.feishu_webhook" placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/..." />
        </NFormItem>
      </template>

      <!-- Telegram -->
      <template v-if="channelForm.type === 'telegram'">
        <NFormItem label="Bot Token">
          <NInput v-model:value="channelForm.telegram_bot_token" type="password" show-password-on="click" placeholder="123456:ABC-DEF..." />
        </NFormItem>
        <NFormItem :label="t('alert.channelsTab.form.telegramChatId')">
          <NInput v-model:value="channelForm.telegram_chat_id" :placeholder="t('alert.channelsTab.form.telegramChatIdPlaceholder')" />
        </NFormItem>
      </template>

      <!-- Webhook -->
      <template v-if="channelForm.type === 'webhook'">
        <NFormItem label="URL">
          <NInput v-model:value="channelForm.webhook_url" placeholder="https://example.com/webhook" />
        </NFormItem>
        <NFormItem :label="t('alert.channelsTab.form.webhookHeaders')">
          <NInput
            v-model:value="channelForm.webhook_headers"
            type="textarea"
            placeholder='{"Authorization": "Bearer xxx", "X-Custom": "value"}'
            :rows="3"
          />
        </NFormItem>
      </template>

      <NFormItem :label="t('alert.channelsTab.form.enabled')">
        <NSwitch v-model:value="channelForm.enabled" />
      </NFormItem>
    </NForm>
  </ModalForm>

  <!-- ==================== 规则删除确认 ==================== -->
  <ConfirmDialog
    :show="showRuleDelete"
    :title="t('alert.rules.deleteConfirm.title')"
    :content="t('alert.rules.deleteConfirm.content', { name: ruleDeleteName })"
    :confirm-text="t('alert.rules.deleteConfirm.confirm')"
    type="error"
    :loading="ruleDeleteLoading"
    @update:show="(v: boolean) => (showRuleDelete = v)"
    @confirm="handleRuleDeleteConfirm"
  />

  <!-- ==================== 渠道删除确认 ==================== -->
  <ConfirmDialog
    :show="showChannelDelete"
    :title="t('alert.channelsTab.deleteConfirm.title')"
    :content="t('alert.channelsTab.deleteConfirm.content', { name: channelDeleteName })"
    :confirm-text="t('alert.channelsTab.deleteConfirm.confirm')"
    type="error"
    :loading="channelDeleteLoading"
    @update:show="(v: boolean) => (showChannelDelete = v)"
    @confirm="handleChannelDeleteConfirm"
  />
</template>

<style scoped>
/* 规则名称 */
.rule-name {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.rule-name__main {
  font-weight: 500;
  color: var(--ink);
}

.rule-name__remark {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
}

.rule-condition {
  font-weight: 500;
  color: var(--ink);
  font-size: var(--fs-mono);
}

/* 渠道名称 */
.channel-name {
  font-weight: 500;
  color: var(--ink);
}

/* 历史筛选 */
.history-filter {
  padding: var(--space-4);
  background: var(--surface2);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-4);
}

.history-filter__item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.history-filter__label {
  font-size: var(--fs-caption);
  color: var(--sub);
  white-space: nowrap;
}

/* 历史值 */
.history-value {
  font-weight: 600;
  color: var(--ink);
  font-variant-numeric: tabular-nums;
}

.history-rule-name {
  font-weight: 500;
  color: var(--ink);
}

/* 展开行 */
.history-expand-row {
  padding: var(--space-3) var(--space-4);
  background: var(--surface2);
  border-left: 3px solid var(--err);
  margin: 0 var(--space-1) var(--space-2);
  border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
  display: flex;
  gap: var(--space-2);
  align-items: flex-start;
}

.history-expand__label {
  font-size: var(--fs-caption);
  color: var(--err);
  font-weight: 600;
  flex-shrink: 0;
}

.history-expand__error {
  font-size: var(--fs-caption);
  color: var(--sub);
  word-break: break-all;
}

/* 空态 */
.history-empty {
  padding: var(--space-8) 0;
  text-align: center;
}

.history-empty p {
  color: var(--sub);
  font-size: var(--fs-body);
  margin: 0 0 var(--space-2);
}

.history-empty span {
  color: var(--text-quaternary);
  font-size: var(--fs-caption);
}

/* .form-hint 已收敛为全局样式（src/styles/main.css），此处不再重复定义 */
</style>
