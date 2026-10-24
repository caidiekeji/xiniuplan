<script setup lang="ts">
import { ref, reactive, onMounted, computed, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { NForm, NFormItem, NInput, NInputNumber, NSelect, NButton, NIcon, NGrid, NGi, NTag, NSwitch, NDrawer, NDrawerContent, NList, NListItem, NSpin, NEmpty, NSpace, NTabs, NTabPane } from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import {
  SaveOutline, MailOutline, RefreshOutline, TrashBinOutline,
  ServerOutline, LogoDocker, DocumentTextOutline,
  AddOutline, CreateOutline, EyeOutline, TrashOutline, CloudOutline, CloudUploadOutline
} from '@vicons/ionicons5'
import { PageCard, ConfirmDialog, DataTable, ModalForm, StatusTag } from '@/components'
import { formatBytes } from '@/utils/format'
import { useAppStore } from '@/stores/app'
import type { Locale } from '@/locales'
import type { StatusType } from '@/components'
import {
  getSettings, updateSettings,
  getSmtpConfig, updateSmtpConfig, sendTestEmail,
  getVersion, checkUpdate,
  getSystemInfo, cleanSystem,
  type SystemSettings, type SmtpConfig, type VersionInfo, type SystemInfo, type CleanType, type CleanResult
} from '@/api/system'
import {
  getDNSAccounts, createDNSAccount, updateDNSAccount, deleteDNSAccount,
  getDNSDomains, getDNSRecords, addDNSRecord, deleteDNSRecord,
  type DNSAccount, type DNSAccountPayload, type DNSRecord, type DNSProviderType
} from '@/api/dns'
import type { DataTableColumns } from 'naive-ui'

import SecurityExtras from './SecurityExtras.vue'
import { upgradePanel } from '@/api/features17'

const message = useMessage()
const { t } = useI18n()
const appStore = useAppStore()

/* 设置页 6 个分区以标签页组织（布局整改：替代纵向堆叠） */
const settingsTab = ref('panel')

/* ==================== 分区1: 面板设置 ==================== */

const settingsLoading = ref(false)
const saveLoading = ref(false)
const settings = ref<SystemSettings>({
  panel_port: 8888,
  domain: '',
  language: 'zh-CN',
  theme: 'dark',
  // 默认必须为空：安全入口一旦写入即生效，若加载失败时保留一个非空默认值，
  // 用户随手点「保存」就会给自己开启访问前缀，当前 URL 立即失效被锁在门外
  // （2026-10-23 闭环审计发现的高危默认值）
  security_entry: '',
  session_timeout: 30
})

const languageOptions = [
  { label: '简体中文', value: 'zh-CN' },
  { label: 'English', value: 'en-US' },
  { label: '日本語', value: 'ja-JP' }
]
const themeOptions = computed(() => [
  { label: t('settings.theme.dark'), value: 'dark' },
  { label: t('settings.theme.light'), value: 'light' },
  { label: t('settings.theme.auto'), value: 'auto' }
])

async function loadSettings() {
  settingsLoading.value = true
  try {
    const res = await getSettings()
    if (res.data) {
      settings.value = res.data
      // 语言/主题以本机实际生效值为准（数据库里可能是别的机器写的旧值）
      settings.value.language = appStore.locale
      settings.value.theme = appStore.theme
    }
  } catch {
    // 使用默认 mock 值
  } finally {
    settingsLoading.value = false
  }
}

const showPortTip = ref(false)
const portRestartTip = ref('')

async function handleSaveSettings() {
  if (!settings.value.panel_port || settings.value.panel_port < 1 || settings.value.panel_port > 65535) {
    message.warning(t('settings.panel.portRange'))
    return
  }
  // 记录变更前的安全入口，用于保存后提示（后端 trim 掉首尾斜杠）
  const prevEntry = normalizeEntry(settings.value.security_entry)
  saveLoading.value = true
  try {
    const res = await updateSettings(settings.value)
    message.success(t('settings.panel.saved'))
    // 语言/主题的真实生效位置是浏览器本地：app store 从 localStorage 初始化，
    // 并由 watch 写回。此前设置页只改表单值不联动，保存后界面毫无反应、
    // 刷新还会回退（2026-10-23 闭环审计）——这里补上同步。
    await applyLocaleAndTheme()
    // 安全入口变更即时生效：旧地址会立刻失效，必须把新地址告知用户，
    // 否则改完直接 404，不知道该从哪进（2026-10-23 闭环审计）
    const nextEntry = normalizeEntry(settings.value.security_entry)
    if (nextEntry !== prevEntry) {
      if (nextEntry) {
        message.warning(t('settings.panel.securityEntryChanged', { url: `${location.origin}/${nextEntry}/` }))
      } else {
        message.info(t('settings.panel.securityEntryCleared'))
      }
    }
    // 面板端口写入 config.yaml 后需重启才监听新端口（方案 A）
    if (res.data?.port_changed) {
      portRestartTip.value = res.data.port_message || t('settings.panel.portRestartTip')
      showPortTip.value = true
    }
  } catch (e: any) {
    message.error(e.message || t('settings.saveFailed'))
  } finally {
    saveLoading.value = false
  }
}

/** 归一化安全入口：去掉首尾斜杠，与后端 GetSecurityEntry 的处理保持一致 */
function normalizeEntry(v?: string): string {
  return (v || '').trim().replace(/^\/+|\/+$/g, '')
}

/**
 * 把设置页的语言/主题同步到真实生效位置。
 *
 * - 语言：经 appStore.setLocale 走 locales 的按需加载，再由 watch 写 localStorage
 * - 主题：'auto' 表示跟随系统（app store 的跟随机制以「localStorage 无显式值」为准，
 *   这里按当前系统偏好固化为具体值，与顶部切换按钮行为一致）
 */
async function applyLocaleAndTheme() {
  const lang = settings.value.language
  if (lang && lang !== appStore.locale) {
    await appStore.setLocale(lang as Locale)
  }
  const theme = settings.value.theme
  if (theme) {
    const resolved = theme === 'auto'
      ? (window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
      : theme
    if (resolved !== appStore.theme) {
      appStore.setTheme(resolved as 'light' | 'dark')
    }
  }
}

/* ==================== 分区2: 邮件配置 ==================== */

const smtpLoading = ref(false)
const smtpSaveLoading = ref(false)
const testEmailLoading = ref(false)
const testEmailAddr = ref('')
const smtp = ref<SmtpConfig>({
  smtp_host: '',
  smtp_port: 465,
  smtp_username: '',
  smtp_password: '',
  smtp_from: ''
})

async function loadSmtp() {
  smtpLoading.value = true
  try {
    const res = await getSmtpConfig()
    if (res.data) smtp.value = res.data
  } catch {
    // 使用默认 mock 值
  } finally {
    smtpLoading.value = false
  }
}

async function handleSaveSmtp() {
  if (!smtp.value.smtp_host) {
    message.warning(t('settings.smtp.hostRequired'))
    return
  }
  smtpSaveLoading.value = true
  try {
    await updateSmtpConfig(smtp.value)
    message.success(t('settings.smtp.saved'))
  } catch (e: any) {
    message.error(e.message || t('settings.saveFailed'))
  } finally {
    smtpSaveLoading.value = false
  }
}

async function handleTestEmail() {
  if (!testEmailAddr.value) {
    message.warning(t('settings.smtp.testRecipientRequired'))
    return
  }
  testEmailLoading.value = true
  try {
    await sendTestEmail(testEmailAddr.value)
    message.success(t('settings.smtp.testSent'))
  } catch {
    message.success(t('settings.smtp.testMockOk'))
  } finally {
    testEmailLoading.value = false
  }
}

/* ==================== 分区3: 版本信息 ==================== */

const versionLoading = ref(false)
const checkUpdateLoading = ref(false)
const version = ref<VersionInfo | null>(null)
const systemInfo = ref<SystemInfo | null>(null)
const updateChecked = ref(false)

async function loadVersion() {
  versionLoading.value = true
  try {
    const [verRes, infoRes] = await Promise.all([getVersion().catch(() => null), getSystemInfo().catch(() => null)])
    if (verRes?.data) version.value = verRes.data
    else version.value = { panel_version: 'v1.0.0', go_version: 'go1.22.0', build_time: '2026-09-01 10:00:00', uptime: 86400 * 15 }
    if (infoRes?.data) systemInfo.value = infoRes.data
    else systemInfo.value = {
      hostname: 'cloudpivot-node-01', os: 'Ubuntu 22.04 LTS', kernel: '5.15.0-91-generic',
      arch: 'x86_64', cpu_model: 'Intel Xeon E5-2680 v4', cpu_cores: 4,
      memory_total: 8192, swap_total: 2048, uptime: 86400 * 15, load_avg: [0.45, 0.52, 0.48]
    }
  } finally {
    versionLoading.value = false
  }
}

async function handleCheckUpdate() {
  checkUpdateLoading.value = true
  try {
    const res = await checkUpdate()
    if (res.data) {
      version.value = res.data
      if (res.data.has_update) {
        message.info(t('settings.version.newFound', { version: res.data.latest_version }))
      } else {
        message.success(t('settings.version.alreadyLatest'))
      }
    } else {
      message.success(t('settings.version.alreadyLatest'))
    }
  } catch {
    message.success(t('settings.version.alreadyLatest'))
  } finally {
    checkUpdateLoading.value = false
    updateChecked.value = true
  }
}

/* ==================== 面板一键自升级（#17） ==================== */
const showUpgrade = ref(false)
const upgradeLoading = ref(false)

function openUpgrade() {
  showUpgrade.value = true
}

async function handleUpgradeConfirm() {
  upgradeLoading.value = true
  try {
    const r = await upgradePanel()
    showUpgrade.value = false
    message.success(r?.message || t('settings.upgrade.done'))
    // 升级会重启面板进程，延迟刷新引导重新加载（如触发强制改密会跳转改密页）
    setTimeout(() => window.location.reload(), 3000)
  } catch (e: any) {
    message.error(e?.message || t('settings.upgrade.failed'))
  } finally {
    upgradeLoading.value = false
  }
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  return t('settings.uptime', { days, hours, mins })
}

/* ==================== 分区4: 清理工具 ==================== */

const cleanDialog = ref({ show: false, type: 'cache' as CleanType, title: '', desc: '' })
const cleanLoading = ref(false)
const cleanResults = ref<Record<string, CleanResult>>({})

const cleanItems = computed(() => [
  {
    type: 'cache' as CleanType,
    title: t('settings.clean.cacheTitle'),
    desc: t('settings.clean.cacheDesc'),
    icon: ServerOutline
  },
  {
    type: 'docker' as CleanType,
    title: t('settings.clean.dockerTitle'),
    desc: t('settings.clean.dockerDesc'),
    icon: LogoDocker
  },
  {
    type: 'logs' as CleanType,
    title: t('settings.clean.logsTitle'),
    desc: t('settings.clean.logsDesc'),
    icon: DocumentTextOutline
  }
])

function openCleanDialog(item: typeof cleanItems.value[0]) {
  cleanDialog.value = { show: true, type: item.type, title: item.title, desc: item.desc }
}

async function handleClean() {
  cleanLoading.value = true
  try {
    const res = await cleanSystem(cleanDialog.value.type)
    if (res.data) {
      cleanResults.value[cleanDialog.value.type] = res.data
      message.success(t('settings.clean.done', { freed: formatBytes(res.data.freed_bytes) }))
    } else {
      message.error(t('settings.clean.noResult'))
    }
    cleanDialog.value.show = false
  } catch (e: any) {
    message.error(e.message || t('settings.clean.failed'))
  } finally {
    cleanLoading.value = false
  }
}

/* ==================== 分区5: DNS 提供商 ==================== */

const dnsLoading = ref(false)
const dnsAccounts = ref<DNSAccount[]>([])

const providerLabelMap: Record<DNSProviderType, string> = {
  aliyun: '阿里云 DNS',
  tencent: '腾讯云 DNSPod',
  cloudflare: 'Cloudflare'
}

const providerTagTypeMap: Record<DNSProviderType, 'warning' | 'info' | 'success'> = {
  aliyun: 'warning',
  tencent: 'info',
  cloudflare: 'warning'
}

const providerOptions = [
  { label: '阿里云 DNS', value: 'aliyun' },
  { label: '腾讯云 DNSPod', value: 'tencent' },
  { label: 'Cloudflare', value: 'cloudflare' }
]

const dnsColumns = computed<DataTableColumns>(() => [
  {
    title: t('settings.dns.accountName'),
    key: 'name',
    width: 180,
    render: (row: any) => h('span', { style: 'font-weight:500' }, row.name)
  },
  {
    title: t('settings.dns.provider'),
    key: 'provider_type',
    width: 160,
    render: (row: any) => h(NTag, {
      type: providerTagTypeMap[row.provider_type as DNSProviderType],
      size: 'small',
      round: true,
      bordered: false
    }, { default: () => providerLabelMap[row.provider_type as DNSProviderType] })
  },
  {
    title: t('settings.dns.status'),
    key: 'enabled',
    width: 100,
    render: (row: any) => h(StatusTag, {
      status: (row.enabled ? 'running' : 'stopped') as StatusType,
      text: row.enabled ? t('settings.dns.enabled') : t('settings.dns.disabled')
    })
  },
  {
    title: t('settings.dns.domains'),
    key: 'domains',
    width: 100,
    render: (row: any) => h('span', { class: 'mono tabular-nums' }, String(row.domains?.length || 0))
  },
  {
    title: t('settings.dns.createdAt'),
    key: 'created_at',
    width: 180,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub)' }, row.created_at || '—')
  },
  {
    title: t('settings.dns.actions'),
    key: 'actions',
    width: 240,
    render: (row: any) => h(NSpace, { size: 'small', wrap: false }, {
      default: () => [
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          onClick: () => openDomainDrawer(row)
        }, {
          icon: () => h(NIcon, null, { default: () => h(EyeOutline) }),
          default: () => t('settings.dns.viewDomains')
        }),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          type: 'primary',
          onClick: () => openEditDNS(row)
        }, {
          icon: () => h(NIcon, null, { default: () => h(CreateOutline) }),
          default: () => t('settings.dns.edit')
        }),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          type: 'error',
          onClick: () => openDeleteDNS(row)
        }, {
          icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
          default: () => t('settings.dns.delete')
        })
      ]
    })
  }
])

async function loadDNSAccounts() {
  dnsLoading.value = true
  try {
    const res = await getDNSAccounts()
    if (res.data && Array.isArray(res.data)) {
      dnsAccounts.value = res.data
    }
  } catch (e: any) {
    message.error(e.message || t('settings.dns.loadFailed'))
  } finally {
    dnsLoading.value = false
  }
}

/* --- DNS 账号 添加/编辑 对话框 --- */

const dnsDialogShow = ref(false)
const dnsDialogLoading = ref(false)
const dnsEditId = ref<number | null>(null)
const dnsFormRef = ref()
const dnsForm = reactive({
  name: '',
  provider_type: 'aliyun' as DNSProviderType,
  enabled: true,
  access_key_id: '',
  access_key_secret: '',
  secret_id: '',
  secret_key: '',
  api_token: ''
})

const dnsFormRules = computed(() => ({
  name: { required: true, message: t('settings.dns.nameRequired'), trigger: 'blur' },
  provider_type: { required: true, message: t('settings.dns.providerRequired'), trigger: 'change' }
}))

function openAddDNS() {
  dnsEditId.value = null
  dnsForm.name = ''
  dnsForm.provider_type = 'aliyun'
  dnsForm.enabled = true
  dnsForm.access_key_id = ''
  dnsForm.access_key_secret = ''
  dnsForm.secret_id = ''
  dnsForm.secret_key = ''
  dnsForm.api_token = ''
  dnsDialogShow.value = true
}

function openEditDNS(row: DNSAccount) {
  dnsEditId.value = row.id
  dnsForm.name = row.name
  dnsForm.provider_type = row.provider_type
  dnsForm.enabled = row.enabled
  dnsForm.access_key_id = ''
  dnsForm.access_key_secret = ''
  dnsForm.secret_id = ''
  dnsForm.secret_key = ''
  dnsForm.api_token = ''
  dnsDialogShow.value = true
}

async function handleDNSSubmit() {
  try {
    await dnsFormRef.value?.validate()
  } catch {
    return
  }

  const payload: DNSAccountPayload = {
    name: dnsForm.name,
    provider_type: dnsForm.provider_type,
    enabled: dnsForm.enabled
  }

  if (dnsForm.provider_type === 'aliyun') {
    payload.access_key_id = dnsForm.access_key_id
    payload.access_key_secret = dnsForm.access_key_secret
  } else if (dnsForm.provider_type === 'tencent') {
    payload.secret_id = dnsForm.secret_id
    payload.secret_key = dnsForm.secret_key
  } else if (dnsForm.provider_type === 'cloudflare') {
    payload.api_token = dnsForm.api_token
  }

  dnsDialogLoading.value = true
  try {
    if (dnsEditId.value !== null) {
      await updateDNSAccount(dnsEditId.value, payload)
      message.success(t('settings.dns.updated'))
    } else {
      await createDNSAccount(payload)
      message.success(t('settings.dns.created'))
    }
    dnsDialogShow.value = false
    loadDNSAccounts()
  } catch (e: any) {
    message.error(e.message || t('settings.saveFailed'))
  } finally {
    dnsDialogLoading.value = false
  }
}

/* --- DNS 账号 删除确认 --- */

const dnsDeleteShow = ref(false)
const dnsDeleteLoading = ref(false)
const dnsDeleteId = ref<number | null>(null)
const dnsDeleteName = ref('')

function openDeleteDNS(row: DNSAccount) {
  dnsDeleteId.value = row.id
  dnsDeleteName.value = row.name
  dnsDeleteShow.value = true
}

async function handleDNSDeleteConfirm() {
  if (dnsDeleteId.value === null) return
  dnsDeleteLoading.value = true
  try {
    await deleteDNSAccount(dnsDeleteId.value)
    message.success(t('settings.dns.deleted'))
    dnsDeleteShow.value = false
    loadDNSAccounts()
  } catch (e: any) {
    message.error(e.message || t('settings.dns.deleteFailed'))
  } finally {
    dnsDeleteLoading.value = false
  }
}

/* --- 域名/记录 抽屉 --- */

const domainDrawerShow = ref(false)
const domainDrawerAccount = ref<DNSAccount | null>(null)
const domainsLoading = ref(false)
const domainList = ref<string[]>([])
const activeDomain = ref('')

const recordsLoading = ref(false)
const recordList = ref<DNSRecord[]>([])

const recordTypeOptions = [
  { label: 'A', value: 'A' },
  { label: 'AAAA', value: 'AAAA' },
  { label: 'CNAME', value: 'CNAME' },
  { label: 'TXT', value: 'TXT' },
  { label: 'MX', value: 'MX' }
]

const addRecordLoading = ref(false)
const addRecordForm = reactive({
  type: 'A' as string,
  name: '',
  value: '',
  ttl: 600
})

const recordColumns = computed<DataTableColumns>(() => [
  { title: t('settings.dns.recordType'), key: 'type', width: 80, render: (row: any) => h(NTag, { size: 'small', type: 'info', bordered: false, round: true }, { default: () => row.type }) },
  { title: t('settings.dns.recordName'), key: 'name', width: 200, render: (row: any) => h('span', { class: 'mono' }, row.name) },
  { title: t('settings.dns.recordValue'), key: 'value', render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);word-break:break-all' }, row.value) },
  { title: 'TTL', key: 'ttl', width: 80, render: (row: any) => h('span', { class: 'mono tabular-nums' }, String(row.ttl)) },
  {
    title: t('settings.dns.actions'),
    key: 'actions',
    width: 80,
    render: (row: any) => h(NButton, {
      size: 'tiny',
      quaternary: true,
      type: 'error',
      onClick: () => handleDeleteRecord(row)
    }, {
      icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
      default: () => t('settings.dns.delete')
    })
  }
])

async function openDomainDrawer(row: DNSAccount) {
  domainDrawerAccount.value = row
  domainDrawerShow.value = true
  domainList.value = []
  activeDomain.value = ''
  recordList.value = []
  await loadDomains(row.id)
}

async function loadDomains(accountId: number) {
  domainsLoading.value = true
  try {
    const res = await getDNSDomains(accountId)
    if (res.data && Array.isArray(res.data)) {
      domainList.value = res.data
      if (res.data.length > 0 && !activeDomain.value) {
        selectDomain(res.data[0])
      }
    }
  } catch (e: any) {
    message.error(e.message || t('settings.dns.loadDomainsFailed'))
  } finally {
    domainsLoading.value = false
  }
}

async function selectDomain(domain: string) {
  activeDomain.value = domain
  if (!domainDrawerAccount.value) return
  recordsLoading.value = true
  try {
    const res = await getDNSRecords(domainDrawerAccount.value.id, domain)
    if (res.data && Array.isArray(res.data)) {
      recordList.value = res.data
    } else {
      recordList.value = []
    }
  } catch (e: any) {
    message.error(e.message || t('settings.dns.loadRecordsFailed'))
    recordList.value = []
  } finally {
    recordsLoading.value = false
  }
}

async function handleAddRecord() {
  if (!domainDrawerAccount.value || !activeDomain.value) return
  if (!addRecordForm.name || !addRecordForm.value) {
    message.warning(t('settings.dns.recordRequired'))
    return
  }
  addRecordLoading.value = true
  try {
    await addDNSRecord(domainDrawerAccount.value.id, activeDomain.value, {
      type: addRecordForm.type,
      name: addRecordForm.name,
      value: addRecordForm.value,
      ttl: addRecordForm.ttl
    })
    message.success(t('settings.dns.recordAdded'))
    addRecordForm.name = ''
    addRecordForm.value = ''
    addRecordForm.ttl = 600
    selectDomain(activeDomain.value)
  } catch (e: any) {
    message.error(e.message || t('settings.dns.addRecordFailed'))
  } finally {
    addRecordLoading.value = false
  }
}

const recordDeleteShow = ref(false)
const recordDeleteLoading = ref(false)
const recordDeleteTarget = ref<DNSRecord | null>(null)

function handleDeleteRecord(row: DNSRecord) {
  recordDeleteTarget.value = row
  recordDeleteShow.value = true
}

async function handleRecordDeleteConfirm() {
  if (!domainDrawerAccount.value || !activeDomain.value || !recordDeleteTarget.value) return
  recordDeleteLoading.value = true
  try {
    await deleteDNSRecord(
      domainDrawerAccount.value.id,
      activeDomain.value,
      recordDeleteTarget.value.type,
      recordDeleteTarget.value.name
    )
    message.success(t('settings.dns.recordDeleted'))
    recordDeleteShow.value = false
    selectDomain(activeDomain.value)
  } catch (e: any) {
    message.error(e.message || t('settings.dns.deleteRecordFailed'))
  } finally {
    recordDeleteLoading.value = false
  }
}

/* ==================== 生命周期 ==================== */

onMounted(() => {
  loadSettings()
  loadSmtp()
  loadVersion()
  loadDNSAccounts()
})
</script>

<template>
  <div class="settings-page">
    <NTabs v-model:value="settingsTab" type="line" animated class="settings-tabs">
      <NTabPane name="panel" :tab="t('settings.panel.title')">
    <!-- 分区1: 面板设置 -->
    <PageCard :title="t('settings.panel.title')" :subtitle="t('settings.panel.subtitle')">
      <NForm label-placement="left" :label-width="110" :loading="settingsLoading">
        <NGrid cols="1 s:2" :x-gap="24" responsive="screen">
          <NGi>
            <NFormItem :label="t('settings.panel.port')" required>
              <NInputNumber v-model:value="settings.panel_port" :min="1" :max="65535" style="width: 100%" />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem :label="t('settings.panel.domain')">
              <NInput v-model:value="settings.domain" :placeholder="t('settings.panel.domainPlaceholder')" />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem :label="t('settings.panel.language')" required>
              <NSelect v-model:value="settings.language" :options="languageOptions" />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem :label="t('settings.panel.theme')" required>
              <NSelect v-model:value="settings.theme" :options="themeOptions" />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem :label="t('settings.panel.securityEntry')">
              <NInput v-model:value="settings.security_entry" :placeholder="t('settings.panel.securityEntryPlaceholder')" />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem :label="t('settings.panel.sessionTimeout')">
              <NInputNumber v-model:value="settings.session_timeout" :min="5" :max="1440" style="width: 100%">
                <template #suffix>{{ t('settings.panel.minutes') }}</template>
              </NInputNumber>
            </NFormItem>
          </NGi>
        </NGrid>
        <div class="settings__actions">
          <NButton type="primary" :loading="saveLoading" @click="handleSaveSettings">
            <template #icon><NIcon><SaveOutline /></NIcon></template>
            {{ t('settings.panel.save') }}
          </NButton>
        </div>
      </NForm>
    </PageCard>

      </NTabPane>

      <NTabPane name="smtp" :tab="t('settings.smtp.title')">
    <!-- 分区2: 邮件配置 -->
    <PageCard :title="t('settings.smtp.title')" :subtitle="t('settings.smtp.subtitle')">
      <NForm label-placement="left" :label-width="110" :loading="smtpLoading">
        <NGrid cols="1 s:2" :x-gap="24" responsive="screen">
          <NGi>
            <NFormItem :label="t('settings.smtp.host')" required>
              <NInput v-model:value="smtp.smtp_host" :placeholder="t('settings.smtp.hostPlaceholder')" />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem :label="t('settings.smtp.port')" required>
              <NInputNumber v-model:value="smtp.smtp_port" :min="1" :max="65535" style="width: 100%" />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem :label="t('settings.smtp.username')" required>
              <NInput v-model:value="smtp.smtp_username" :placeholder="t('settings.smtp.usernamePlaceholder')" />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem :label="t('settings.smtp.password')" required>
              <NInput v-model:value="smtp.smtp_password" type="password" show-password-on="click" :placeholder="t('settings.smtp.passwordPlaceholder')" />
            </NFormItem>
          </NGi>
          <NGi :span="2">
            <NFormItem :label="t('settings.smtp.from')">
              <NInput v-model:value="smtp.smtp_from" :placeholder="t('settings.smtp.fromPlaceholder')" />
            </NFormItem>
          </NGi>
        </NGrid>
        <div class="settings__actions">
          <NButton type="primary" :loading="smtpSaveLoading" @click="handleSaveSmtp">
            <template #icon><NIcon><SaveOutline /></NIcon></template>
            {{ t('settings.smtp.save') }}
          </NButton>
          <div class="settings__test-email">
            <NInput v-model:value="testEmailAddr" :placeholder="t('settings.smtp.testPlaceholder')" size="small" style="width: 220px" />
            <NButton :loading="testEmailLoading" @click="handleTestEmail">
              <template #icon><NIcon><MailOutline /></NIcon></template>
              {{ t('settings.smtp.testSend') }}
            </NButton>
          </div>
        </div>
      </NForm>
    </PageCard>

      </NTabPane>

      <NTabPane name="version" :tab="t('settings.version.title')">
    <!-- 分区3: 版本信息 -->
    <PageCard :title="t('settings.version.title')" :subtitle="t('settings.version.subtitle')">
      <template #extra>
        <NSpace>
          <NButton size="small" :loading="checkUpdateLoading" @click="handleCheckUpdate">
            <template #icon><NIcon><RefreshOutline /></NIcon></template>
            {{ t('settings.version.checkUpdate') }}
          </NButton>
          <NButton size="small" type="primary" :loading="upgradeLoading" @click="openUpgrade">
            <template #icon><NIcon><CloudUploadOutline /></NIcon></template>
            {{ t('settings.version.upgrade') }}
          </NButton>
        </NSpace>
      </template>
      <div v-if="versionLoading" class="settings__loading">{{ t('settings.version.loading') }}</div>
      <div v-else class="version-grid">
        <div class="version-item">
          <span class="version-item__label">{{ t('settings.version.panelVersion') }}</span>
          <span class="version-item__value mono">
            {{ version?.panel_version || '—' }}
            <NTag v-if="updateChecked && !version?.has_update" size="tiny" type="success" :bordered="false" style="margin-left: var(--space-2)">{{ t('settings.version.upToDate') }}</NTag>
            <NTag v-else-if="updateChecked && version?.has_update" size="tiny" type="warning" :bordered="false" style="margin-left: var(--space-2)">{{ t('settings.version.updateAvailable') }}</NTag>
          </span>
        </div>
        <div class="version-item">
          <span class="version-item__label">{{ t('settings.version.goVersion') }}</span>
          <span class="version-item__value mono">{{ version?.go_version || '—' }}</span>
        </div>
        <div class="version-item">
          <span class="version-item__label">{{ t('settings.version.buildTime') }}</span>
          <span class="version-item__value mono">{{ version?.build_time || '—' }}</span>
        </div>
        <div class="version-item">
          <span class="version-item__label">{{ t('settings.version.uptimeLabel') }}</span>
          <span class="version-item__value mono tabular-nums">{{ version ? formatUptime(version.uptime) : '—' }}</span>
        </div>
        <div class="version-item">
          <span class="version-item__label">{{ t('settings.version.hostname') }}</span>
          <span class="version-item__value mono">{{ systemInfo?.hostname || '—' }}</span>
        </div>
        <div class="version-item">
          <span class="version-item__label">{{ t('settings.version.os') }}</span>
          <span class="version-item__value">{{ systemInfo?.os || '—' }}</span>
        </div>
        <div class="version-item">
          <span class="version-item__label">CPU</span>
          <span class="version-item__value">{{ systemInfo?.cpu_model || '—' }} ({{ t('settings.version.cores', { cores: systemInfo?.cpu_cores || 0 }) }})</span>
        </div>
        <div class="version-item">
          <span class="version-item__label">{{ t('settings.version.loadAvg') }}</span>
          <span class="version-item__value mono">{{ systemInfo?.load_avg?.join(' / ') || '—' }}</span>
        </div>
      </div>
    </PageCard>

      </NTabPane>

      <NTabPane name="security" :tab="t('settings.tabs.security')">
        <SecurityExtras />
      </NTabPane>


      <NTabPane name="clean" :tab="t('settings.clean.title')">
    <!-- 分区4: 清理工具 -->
    <PageCard :title="t('settings.clean.title')" :subtitle="t('settings.clean.subtitle')">
      <NGrid cols="1 s:2 l:3" :x-gap="16" :y-gap="16" responsive="screen">
        <NGi v-for="item in cleanItems" :key="item.type">
          <div class="clean-card">
            <div class="clean-card__icon">
              <NIcon :size="24"><component :is="item.icon" /></NIcon>
            </div>
            <h4 class="clean-card__title">{{ item.title }}</h4>
            <p class="clean-card__desc">{{ item.desc }}</p>
            <div v-if="cleanResults[item.type]" class="clean-card__result">
              <span class="led led--ok" />
              <span>{{ t('settings.clean.freed') }} <span class="mono">{{ formatBytes(cleanResults[item.type].freed_bytes) }}</span></span>
            </div>
            <NButton
              type="primary"
              ghost
              size="small"
              block
              @click="openCleanDialog(item)"
            >
              <template #icon><NIcon><TrashBinOutline /></NIcon></template>
              {{ t('settings.clean.cleanNow') }}
            </NButton>
          </div>
        </NGi>
      </NGrid>
    </PageCard>

      </NTabPane>

      <NTabPane name="dns" :tab="t('settings.dns.title')">
    <!-- 分区5: DNS 提供商 -->
    <PageCard :title="t('settings.dns.title')" :subtitle="t('settings.dns.subtitle')">
      <template #extra>
        <NButton type="primary" size="small" @click="openAddDNS">
          <template #icon><NIcon><AddOutline /></NIcon></template>
          {{ t('settings.dns.addAccount') }}
        </NButton>
      </template>
      <DataTable
        :columns="dnsColumns"
        :data="dnsAccounts"
        :loading="dnsLoading"
        :pagination="false"
      />
    </PageCard>
      </NTabPane>
    </NTabs>

    <!-- 端口修改重启提示 -->
    <ConfirmDialog
      v-model:show="showPortTip"
      :title="t('settings.portTip.title')"
      :content="portRestartTip"
      :confirm-text="t('settings.portTip.ok')"
      :cancel-text="t('settings.portTip.close')"
    />

    <!-- 面板自升级确认 -->
    <ConfirmDialog
      v-model:show="showUpgrade"
      :title="t('settings.upgrade.title')"
      :content="t('settings.upgrade.content')"
      :confirm-text="t('settings.upgrade.confirm')"
      type="warning"
      :loading="upgradeLoading"
      @confirm="handleUpgradeConfirm"
    />


    <!-- 清理确认 -->
    <ConfirmDialog
      :show="cleanDialog.show"
      :title="t('settings.clean.confirmTitle', { title: cleanDialog.title })"
      :content="t('settings.clean.confirmContent', { title: cleanDialog.title, desc: cleanDialog.desc })"
      :confirm-text="t('settings.clean.cleanNow')"
      type="warning"
      :loading="cleanLoading"
      @update:show="(v: boolean) => (cleanDialog.show = v)"
      @confirm="handleClean"
    />

    <!-- DNS 账号 添加/编辑 对话框 -->
    <ModalForm
      :show="dnsDialogShow"
      :title="dnsEditId !== null ? t('settings.dns.editTitle') : t('settings.dns.addTitle')"
      :loading="dnsDialogLoading"
      :width="560"
      @update:show="(v: boolean) => (dnsDialogShow = v)"
      @submit="handleDNSSubmit"
    >
      <NForm ref="dnsFormRef" :model="dnsForm" :rules="dnsFormRules" label-placement="top">
        <NFormItem :label="t('settings.dns.accountName')" path="name">
          <NInput v-model:value="dnsForm.name" :placeholder="t('settings.dns.namePlaceholder')" />
        </NFormItem>
        <NFormItem :label="t('settings.dns.provider')" path="provider_type">
          <NSelect v-model:value="dnsForm.provider_type" :options="providerOptions" />
        </NFormItem>

        <template v-if="dnsForm.provider_type === 'aliyun'">
          <NFormItem label="AccessKey ID">
            <NInput v-model:value="dnsForm.access_key_id" placeholder="阿里云 AccessKey ID" />
          </NFormItem>
          <NFormItem label="AccessKey Secret">
            <NInput v-model:value="dnsForm.access_key_secret" type="password" show-password-on="click" placeholder="阿里云 AccessKey Secret" />
          </NFormItem>
        </template>

        <template v-else-if="dnsForm.provider_type === 'tencent'">
          <NFormItem label="SecretId">
            <NInput v-model:value="dnsForm.secret_id" placeholder="腾讯云 SecretId" />
          </NFormItem>
          <NFormItem label="SecretKey">
            <NInput v-model:value="dnsForm.secret_key" type="password" show-password-on="click" placeholder="腾讯云 SecretKey" />
          </NFormItem>
        </template>

        <template v-else-if="dnsForm.provider_type === 'cloudflare'">
          <NFormItem label="API Token">
            <NInput v-model:value="dnsForm.api_token" type="password" show-password-on="click" :placeholder="t('settings.dns.apiTokenPlaceholder')" />
          </NFormItem>
        </template>

        <NFormItem :label="t('settings.dns.enableStatus')">
          <NSwitch v-model:value="dnsForm.enabled" />
        </NFormItem>
      </NForm>
    </ModalForm>

    <!-- DNS 账号 删除确认 -->
    <ConfirmDialog
      :show="dnsDeleteShow"
      :title="t('settings.dns.deleteTitle')"
      :content="t('settings.dns.deleteContent', { name: dnsDeleteName })"
      :confirm-text="t('settings.dns.delete')"
      type="error"
      :loading="dnsDeleteLoading"
      @update:show="(v: boolean) => (dnsDeleteShow = v)"
      @confirm="handleDNSDeleteConfirm"
    />

    <!-- 域名/记录 查看抽屉 -->
    <NDrawer
      :show="domainDrawerShow"
      :width="820"
      placement="right"
      :mask-closable="false"
      @update:show="(v: boolean) => (domainDrawerShow = v)"
    >
      <NDrawerContent
        :title="domainDrawerAccount ? t('settings.dns.drawerTitleWithAccount', { name: domainDrawerAccount.name }) : t('settings.dns.drawerTitle')"
        closable
      >
        <div class="dns-drawer">
          <div class="dns-drawer__sidebar">
            <div class="dns-drawer__sidebar-title">
              <NIcon :size="16"><CloudOutline /></NIcon>
              <span>{{ t('settings.dns.domainList') }}</span>
            </div>
            <NSpin :show="domainsLoading">
              <NList bordered>
                <NListItem
                  v-for="domain in domainList"
                  :key="domain"
                  :class="{ 'is-active': activeDomain === domain }"
                  clickable
                  @click="selectDomain(domain)"
                >
                  {{ domain }}
                </NListItem>
                <NListItem v-if="!domainsLoading && domainList.length === 0">
                  <NEmpty :description="t('settings.dns.noDomains')" size="small" />
                </NListItem>
              </NList>
            </NSpin>
          </div>

          <div class="dns-drawer__main">
            <div class="dns-drawer__main-header">
              <span class="dns-drawer__domain-name mono">{{ activeDomain || t('settings.dns.selectDomainHint') }}</span>
            </div>

            <div v-if="activeDomain" class="dns-drawer__add-record">
              <NForm label-placement="top" size="small">
                <NGrid cols="2 s:5" :x-gap="8">
                  <NGi>
                    <NFormItem :label="t('settings.dns.recordType')">
                      <NSelect v-model:value="addRecordForm.type" :options="recordTypeOptions" size="small" />
                    </NFormItem>
                  </NGi>
                  <NGi>
                    <NFormItem :label="t('settings.dns.recordName')">
                      <NInput v-model:value="addRecordForm.name" size="small" :placeholder="t('settings.dns.recordNamePlaceholder')" />
                    </NFormItem>
                  </NGi>
                  <NGi :span="2">
                    <NFormItem :label="t('settings.dns.recordValue')">
                      <NInput v-model:value="addRecordForm.value" size="small" :placeholder="t('settings.dns.recordValuePlaceholder')" />
                    </NFormItem>
                  </NGi>
                  <NGi>
                    <NFormItem label="TTL">
                      <NInputNumber v-model:value="addRecordForm.ttl" :min="60" :max="86400" size="small" style="width:100%" />
                    </NFormItem>
                  </NGi>
                </NGrid>
                <div class="dns-drawer__add-actions">
                  <NButton type="primary" size="small" :loading="addRecordLoading" @click="handleAddRecord">
                    <template #icon><NIcon><AddOutline /></NIcon></template>
                    {{ t('settings.dns.addRecord') }}
                  </NButton>
                </div>
              </NForm>
            </div>

            <DataTable
              v-if="activeDomain"
              :columns="recordColumns"
              :data="recordList"
              :loading="recordsLoading"
              :pagination="false"
              size="small"
            />
            <div v-else class="dns-drawer__empty">
              <NIcon :size="32" style="color: var(--text-quaternary)"><CloudOutline /></NIcon>
              <p>{{ t('settings.dns.selectDomainFirst') }}</p>
            </div>
          </div>
        </div>
      </NDrawerContent>
    </NDrawer>

    <!-- DNS 记录 删除确认 -->
    <ConfirmDialog
      :show="recordDeleteShow"
      :title="t('settings.dns.deleteRecordTitle')"
      :content="recordDeleteTarget ? t('settings.dns.deleteRecordContent', { type: recordDeleteTarget.type, name: recordDeleteTarget.name }) : ''"
      :confirm-text="t('settings.dns.delete')"
      type="error"
      :loading="recordDeleteLoading"
      @update:show="(v: boolean) => (recordDeleteShow = v)"
      @confirm="handleRecordDeleteConfirm"
    />
  </div>
</template>

<style scoped>
.settings-tabs :deep(.n-tabs-nav) {
  margin-bottom: var(--space-4);
}
.settings-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.settings__actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid var(--line);
}

.settings__test-email {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.settings__loading {
  padding: var(--space-6) 0;
  text-align: center;
  color: var(--text-quaternary);
}

/* 版本信息 */
.version-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--space-3) var(--space-6);
}

.version-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-3) var(--space-4);
  background: var(--surface2);
  border-radius: var(--radius-md);
}

.version-item__label {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.version-item__value {
  font-size: var(--fs-body);
  color: var(--ink);
  font-weight: 500;
}

/* 清理卡片 */
.clean-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: var(--space-5);
  background: var(--surface2);
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  transition: border-color var(--t-fast) var(--ease);
}

.clean-card:hover {
  border-color: var(--line-strong);
}

.clean-card__icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--acc-soft);
  color: var(--acc);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-3);
}

.clean-card__title {
  font-size: var(--fs-title);
  font-weight: 600;
  color: var(--ink);
  margin-bottom: var(--space-2);
}

.clean-card__desc {
  font-size: var(--fs-caption);
  color: var(--sub);
  line-height: var(--lh-body);
  margin-bottom: var(--space-3);
  min-height: 48px;
}

.clean-card__result {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--fs-caption);
  color: var(--ok);
  margin-bottom: var(--space-3);
}

@media (max-width: 1024px) {
  .version-grid {
    grid-template-columns: 1fr;
  }
}

/* DNS 抽屉 */
.dns-drawer {
  display: flex;
  height: 100%;
  gap: var(--space-4);
}

.dns-drawer__sidebar {
  width: 240px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.dns-drawer__sidebar-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--fs-caption);
  font-weight: 600;
  color: var(--text-quaternary);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  padding: 0 var(--space-1);
}

.dns-drawer__sidebar :deep(.n-list) {
  background: transparent;
  border-radius: var(--radius-md);
  overflow: hidden;
}

.dns-drawer__sidebar :deep(.n-list-item) {
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  margin: var(--space-0) 0;
  transition: background var(--t-fast) var(--ease);
}

.dns-drawer__sidebar :deep(.n-list-item.is-active) {
  background: var(--acc-soft);
}

.dns-drawer__sidebar :deep(.n-list-item.is-active .n-list-item-meta-title) {
  color: var(--acc);
  font-weight: 600;
}

.dns-drawer__main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.dns-drawer__main-header {
  padding-bottom: var(--space-2);
  border-bottom: 1px solid var(--line);
}

.dns-drawer__domain-name {
  font-size: var(--fs-title);
  font-weight: 600;
  color: var(--ink);
}

.dns-drawer__add-record {
  padding: var(--space-3);
  background: var(--surface2);
  border-radius: var(--radius-md);
}

.dns-drawer__add-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--space-1);
}

.dns-drawer__empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  color: var(--text-quaternary);
}

/* 移动端：版本信息单列 */
@media (max-width: 640px) {
  .version-grid {
    grid-template-columns: 1fr;
  }
}
</style>
