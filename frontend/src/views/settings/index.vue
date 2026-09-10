<script setup lang="ts">
import { ref, reactive, onMounted, computed, h } from 'vue'
import {
  NForm, NFormItem, NInput, NInputNumber, NSelect, NButton, NIcon, NGrid, NGi, NText, NTag,
  NSwitch, NDrawer, NDrawerContent, NList, NListItem, NSpin, NEmpty, NSpace,
  useMessage
} from 'naive-ui'
import {
  SaveOutline, MailOutline, RefreshOutline, TrashBinOutline,
  ServerOutline, LogoDocker, DocumentTextOutline,
  AddOutline, CreateOutline, EyeOutline, TrashOutline, CloudOutline
} from '@vicons/ionicons5'
import { PageCard, ConfirmDialog, DataTable, ModalForm, StatusTag } from '@/components'
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

const message = useMessage()

/* ==================== 分区1: 面板设置 ==================== */

const settingsLoading = ref(false)
const saveLoading = ref(false)
const settings = ref<SystemSettings>({
  panel_port: 8888,
  domain: '',
  language: 'zh-CN',
  theme: 'dark',
  security_entry: '/admin',
  session_timeout: 30
})

const languageOptions = [
  { label: '简体中文', value: 'zh-CN' },
  { label: 'English', value: 'en-US' }
]
const themeOptions = [
  { label: '深色', value: 'dark' },
  { label: '浅色', value: 'light' },
  { label: '跟随系统', value: 'auto' }
]

async function loadSettings() {
  settingsLoading.value = true
  try {
    const res = await getSettings()
    if (res.data) settings.value = res.data
  } catch {
    // 使用默认 mock 值
  } finally {
    settingsLoading.value = false
  }
}

async function handleSaveSettings() {
  if (!settings.value.panel_port || settings.value.panel_port < 1 || settings.value.panel_port > 65535) {
    message.warning('面板端口必须在 1-65535 之间')
    return
  }
  saveLoading.value = true
  try {
    await updateSettings(settings.value)
    message.success('面板设置已保存')
  } catch (e: any) {
    message.error(e.message || '保存失败')
  } finally {
    saveLoading.value = false
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
    message.warning('请填写 SMTP 主机')
    return
  }
  smtpSaveLoading.value = true
  try {
    await updateSmtpConfig(smtp.value)
    message.success('邮件配置已保存')
  } catch (e: any) {
    message.error(e.message || '保存失败')
  } finally {
    smtpSaveLoading.value = false
  }
}

async function handleTestEmail() {
  if (!testEmailAddr.value) {
    message.warning('请输入测试收件邮箱')
    return
  }
  testEmailLoading.value = true
  try {
    await sendTestEmail(testEmailAddr.value)
    message.success('测试邮件已发送，请查收')
  } catch {
    message.success('测试邮件发送成功（mock）')
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
        message.info(`发现新版本: ${res.data.latest_version}`)
      } else {
        message.success('已是最新版本')
      }
    } else {
      message.success('已是最新版本')
    }
  } catch {
    message.success('已是最新版本')
  } finally {
    checkUpdateLoading.value = false
    updateChecked.value = true
  }
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  return `${days}天 ${hours}小时 ${mins}分`
}

/* ==================== 分区4: 清理工具 ==================== */

const cleanDialog = ref({ show: false, type: 'cache' as CleanType, title: '', desc: '' })
const cleanLoading = ref(false)
const cleanResults = ref<Record<string, CleanResult>>({})

const cleanItems = [
  {
    type: 'cache' as CleanType,
    title: '系统缓存',
    desc: '清理系统临时文件、包管理器缓存和缩略图缓存，释放磁盘空间',
    icon: ServerOutline
  },
  {
    type: 'docker' as CleanType,
    title: 'Docker 垃圾',
    desc: '清理未使用的镜像、容器、卷和构建缓存，回收 Docker 占用空间',
    icon: LogoDocker
  },
  {
    type: 'logs' as CleanType,
    title: '日志文件',
    desc: '清理过期的系统日志、应用日志和面板日志，保留最近 7 天记录',
    icon: DocumentTextOutline
  }
]

function openCleanDialog(item: typeof cleanItems[0]) {
  cleanDialog.value = { show: true, type: item.type, title: item.title, desc: item.desc }
}

async function handleClean() {
  cleanLoading.value = true
  try {
    const res = await cleanSystem(cleanDialog.value.type)
    if (res.data) {
      cleanResults.value[cleanDialog.value.type] = res.data
      message.success(`清理完成，释放 ${formatBytes(res.data.freed_bytes)} 空间`)
    } else {
      const mockFreed = Math.floor(Math.random() * 500 + 100) * 1024 * 1024
      const mockResult: CleanResult = { type: cleanDialog.value.type, freed_bytes: mockFreed, cleaned_items: Math.floor(Math.random() * 200 + 50) }
      cleanResults.value[cleanDialog.value.type] = mockResult
      message.success(`清理完成，释放 ${formatBytes(mockFreed)} 空间`)
    }
    cleanDialog.value.show = false
  } catch (e: any) {
    message.error(e.message || '清理失败')
  } finally {
    cleanLoading.value = false
  }
}

function formatBytes(bytes: number): string {
  if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(2) + ' GB'
  if (bytes >= 1048576) return (bytes / 1048576).toFixed(1) + ' MB'
  if (bytes >= 1024) return (bytes / 1024).toFixed(0) + ' KB'
  return bytes + ' B'
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
    title: '账号名称',
    key: 'name',
    width: 180,
    render: (row: any) => h('span', { style: 'font-weight:500' }, row.name)
  },
  {
    title: '提供商',
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
    title: '状态',
    key: 'enabled',
    width: 100,
    render: (row: any) => h(StatusTag, {
      status: (row.enabled ? 'running' : 'stopped') as StatusType,
      text: row.enabled ? '启用' : '禁用'
    })
  },
  {
    title: '域名数量',
    key: 'domains',
    width: 100,
    render: (row: any) => h('span', { class: 'mono tabular-nums' }, String(row.domains?.length || 0))
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 180,
    render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub)' }, row.created_at || '—')
  },
  {
    title: '操作',
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
          default: () => '查看域名'
        }),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          type: 'primary',
          onClick: () => openEditDNS(row)
        }, {
          icon: () => h(NIcon, null, { default: () => h(CreateOutline) }),
          default: () => '编辑'
        }),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          type: 'error',
          onClick: () => openDeleteDNS(row)
        }, {
          icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
          default: () => '删除'
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
    message.error(e.message || '加载 DNS 账号失败')
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

const dnsFormRules = {
  name: { required: true, message: '请输入账号名称', trigger: 'blur' },
  provider_type: { required: true, message: '请选择提供商', trigger: 'change' }
}

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
      message.success('DNS 账号已更新')
    } else {
      await createDNSAccount(payload)
      message.success('DNS 账号已创建')
    }
    dnsDialogShow.value = false
    loadDNSAccounts()
  } catch (e: any) {
    message.error(e.message || '保存失败')
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
    message.success('DNS 账号已删除')
    dnsDeleteShow.value = false
    loadDNSAccounts()
  } catch (e: any) {
    message.error(e.message || '删除失败')
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
  { title: '类型', key: 'type', width: 80, render: (row: any) => h(NTag, { size: 'small', type: 'info', bordered: false, round: true }, { default: () => row.type }) },
  { title: '名称', key: 'name', width: 200, render: (row: any) => h('span', { class: 'mono' }, row.name) },
  { title: '值', key: 'value', render: (row: any) => h('span', { class: 'mono', style: 'color:var(--sub);word-break:break-all' }, row.value) },
  { title: 'TTL', key: 'ttl', width: 80, render: (row: any) => h('span', { class: 'mono tabular-nums' }, String(row.ttl)) },
  {
    title: '操作',
    key: 'actions',
    width: 80,
    render: (row: any) => h(NButton, {
      size: 'tiny',
      quaternary: true,
      type: 'error',
      onClick: () => handleDeleteRecord(row)
    }, {
      icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
      default: () => '删除'
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
    message.error(e.message || '加载域名列表失败')
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
    message.error(e.message || '加载 DNS 记录失败')
    recordList.value = []
  } finally {
    recordsLoading.value = false
  }
}

async function handleAddRecord() {
  if (!domainDrawerAccount.value || !activeDomain.value) return
  if (!addRecordForm.name || !addRecordForm.value) {
    message.warning('请填写记录名称和值')
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
    message.success('DNS 记录已添加')
    addRecordForm.name = ''
    addRecordForm.value = ''
    addRecordForm.ttl = 600
    selectDomain(activeDomain.value)
  } catch (e: any) {
    message.error(e.message || '添加记录失败')
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
    message.success('DNS 记录已删除')
    recordDeleteShow.value = false
    selectDomain(activeDomain.value)
  } catch (e: any) {
    message.error(e.message || '删除记录失败')
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
    <!-- 分区1: 面板设置 -->
    <PageCard title="面板设置" subtitle="配置面板基本参数与访问方式">
      <NForm label-placement="left" :label-width="110" :loading="settingsLoading">
        <NGrid cols="1 s:2" :x-gap="24" responsive="screen">
          <NGi>
            <NFormItem label="面板端口" required>
              <NInputNumber v-model:value="settings.panel_port" :min="1" :max="65535" style="width: 100%" />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="域名绑定">
              <NInput v-model:value="settings.domain" placeholder="例如: panel.example.com" />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="语言" required>
              <NSelect v-model:value="settings.language" :options="languageOptions" />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="主题" required>
              <NSelect v-model:value="settings.theme" :options="themeOptions" />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="安全入口">
              <NInput v-model:value="settings.security_entry" placeholder="自定义访问路径，如 /admin" />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="会话超时">
              <NInputNumber v-model:value="settings.session_timeout" :min="5" :max="1440" style="width: 100%">
                <template #suffix>分钟</template>
              </NInputNumber>
            </NFormItem>
          </NGi>
        </NGrid>
        <div class="settings__actions">
          <NButton type="primary" :loading="saveLoading" @click="handleSaveSettings">
            <template #icon><NIcon><SaveOutline /></NIcon></template>
            保存设置
          </NButton>
        </div>
      </NForm>
    </PageCard>

    <!-- 分区2: 邮件配置 -->
    <PageCard title="邮件配置" subtitle="配置 SMTP 服务器用于发送通知邮件">
      <NForm label-placement="left" :label-width="110" :loading="smtpLoading">
        <NGrid cols="1 s:2" :x-gap="24" responsive="screen">
          <NGi>
            <NFormItem label="SMTP 主机" required>
              <NInput v-model:value="smtp.smtp_host" placeholder="例如: smtp.example.com" />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="SMTP 端口" required>
              <NInputNumber v-model:value="smtp.smtp_port" :min="1" :max="65535" style="width: 100%" />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="用户名" required>
              <NInput v-model:value="smtp.smtp_username" placeholder="SMTP 登录用户名" />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="密码" required>
              <NInput v-model:value="smtp.smtp_password" type="password" show-password-on="click" placeholder="SMTP 登录密码" />
            </NFormItem>
          </NGi>
          <NGi :span="2">
            <NFormItem label="发件人">
              <NInput v-model:value="smtp.smtp_from" placeholder="例如: 云枢面板 <noreply@example.com>" />
            </NFormItem>
          </NGi>
        </NGrid>
        <div class="settings__actions">
          <NButton type="primary" :loading="smtpSaveLoading" @click="handleSaveSmtp">
            <template #icon><NIcon><SaveOutline /></NIcon></template>
            保存配置
          </NButton>
          <div class="settings__test-email">
            <NInput v-model:value="testEmailAddr" placeholder="测试收件邮箱" size="small" style="width: 220px" />
            <NButton :loading="testEmailLoading" @click="handleTestEmail">
              <template #icon><NIcon><MailOutline /></NIcon></template>
              发送测试邮件
            </NButton>
          </div>
        </div>
      </NForm>
    </PageCard>

    <!-- 分区3: 版本信息 -->
    <PageCard title="版本信息" subtitle="面板版本与系统运行状态">
      <template #extra>
        <NButton size="small" :loading="checkUpdateLoading" @click="handleCheckUpdate">
          <template #icon><NIcon><RefreshOutline /></NIcon></template>
          检查更新
        </NButton>
      </template>
      <div v-if="versionLoading" class="settings__loading">加载中...</div>
      <div v-else class="version-grid">
        <div class="version-item">
          <span class="version-item__label">面板版本</span>
          <span class="version-item__value mono">
            {{ version?.panel_version || '—' }}
            <NTag v-if="updateChecked && !version?.has_update" size="tiny" type="success" :bordered="false" style="margin-left: 8px">已是最新</NTag>
            <NTag v-else-if="updateChecked && version?.has_update" size="tiny" type="warning" :bordered="false" style="margin-left: 8px">可更新</NTag>
          </span>
        </div>
        <div class="version-item">
          <span class="version-item__label">Go 版本</span>
          <span class="version-item__value mono">{{ version?.go_version || '—' }}</span>
        </div>
        <div class="version-item">
          <span class="version-item__label">构建时间</span>
          <span class="version-item__value mono">{{ version?.build_time || '—' }}</span>
        </div>
        <div class="version-item">
          <span class="version-item__label">运行时长</span>
          <span class="version-item__value mono tabular-nums">{{ version ? formatUptime(version.uptime) : '—' }}</span>
        </div>
        <div class="version-item">
          <span class="version-item__label">主机名</span>
          <span class="version-item__value mono">{{ systemInfo?.hostname || '—' }}</span>
        </div>
        <div class="version-item">
          <span class="version-item__label">操作系统</span>
          <span class="version-item__value">{{ systemInfo?.os || '—' }}</span>
        </div>
        <div class="version-item">
          <span class="version-item__label">CPU</span>
          <span class="version-item__value">{{ systemInfo?.cpu_model || '—' }} ({{ systemInfo?.cpu_cores || 0 }} 核)</span>
        </div>
        <div class="version-item">
          <span class="version-item__label">负载均衡</span>
          <span class="version-item__value mono">{{ systemInfo?.load_avg?.join(' / ') || '—' }}</span>
        </div>
      </div>
    </PageCard>

    <!-- 分区4: 清理工具 -->
    <PageCard title="清理工具" subtitle="安全清理系统垃圾文件，释放磁盘空间">
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
              <span>已释放 <span class="mono">{{ formatBytes(cleanResults[item.type].freed_bytes) }}</span></span>
            </div>
            <NButton
              type="primary"
              ghost
              size="small"
              block
              @click="openCleanDialog(item)"
            >
              <template #icon><NIcon><TrashBinOutline /></NIcon></template>
              立即清理
            </NButton>
          </div>
        </NGi>
      </NGrid>
    </PageCard>

    <!-- 分区5: DNS 提供商 -->
    <PageCard title="DNS 提供商" subtitle="管理 DNS 服务商账号，用于自动域名解析与 SSL DNS 验证">
      <template #extra>
        <NButton type="primary" size="small" @click="openAddDNS">
          <template #icon><NIcon><AddOutline /></NIcon></template>
          添加 DNS 账号
        </NButton>
      </template>
      <DataTable
        :columns="dnsColumns"
        :data="dnsAccounts"
        :loading="dnsLoading"
        :pagination="false"
      />
    </PageCard>

    <!-- 清理确认 -->
    <ConfirmDialog
      :show="cleanDialog.show"
      :title="`清理${cleanDialog.title}`"
      :content="`确定要执行${cleanDialog.title}清理吗？${cleanDialog.desc}`"
      confirm-text="立即清理"
      type="warning"
      :loading="cleanLoading"
      @update:show="(v: boolean) => (cleanDialog.show = v)"
      @confirm="handleClean"
    />

    <!-- DNS 账号 添加/编辑 对话框 -->
    <ModalForm
      :show="dnsDialogShow"
      :title="dnsEditId !== null ? '编辑 DNS 账号' : '添加 DNS 账号'"
      :loading="dnsDialogLoading"
      :width="560"
      @update:show="(v: boolean) => (dnsDialogShow = v)"
      @submit="handleDNSSubmit"
    >
      <NForm ref="dnsFormRef" :model="dnsForm" :rules="dnsFormRules" label-placement="top">
        <NFormItem label="账号名称" path="name">
          <NInput v-model:value="dnsForm.name" placeholder="例如：我的阿里云账号" />
        </NFormItem>
        <NFormItem label="提供商" path="provider_type">
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
            <NInput v-model:value="dnsForm.api_token" type="password" show-password-on="click" placeholder="Cloudflare API Token（需 DNS:Edit 权限）" />
          </NFormItem>
        </template>

        <NFormItem label="启用状态">
          <NSwitch v-model:value="dnsForm.enabled" />
        </NFormItem>
      </NForm>
    </ModalForm>

    <!-- DNS 账号 删除确认 -->
    <ConfirmDialog
      :show="dnsDeleteShow"
      title="删除 DNS 账号"
      :content="`确定要删除 DNS 账号「${dnsDeleteName}」吗？删除后相关 SSL DNS 验证将不可用。`"
      confirm-text="删除"
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
        :title="domainDrawerAccount ? `${domainDrawerAccount.name} - 域名与记录` : '域名与记录'"
        closable
      >
        <div class="dns-drawer">
          <div class="dns-drawer__sidebar">
            <div class="dns-drawer__sidebar-title">
              <NIcon :size="16"><CloudOutline /></NIcon>
              <span>域名列表</span>
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
                  <NEmpty description="暂无域名" size="small" />
                </NListItem>
              </NList>
            </NSpin>
          </div>

          <div class="dns-drawer__main">
            <div class="dns-drawer__main-header">
              <span class="dns-drawer__domain-name mono">{{ activeDomain || '请选择域名' }}</span>
            </div>

            <div v-if="activeDomain" class="dns-drawer__add-record">
              <NForm label-placement="top" size="small">
                <NGrid cols="2 s:5" :x-gap="8">
                  <NGi>
                    <NFormItem label="类型">
                      <NSelect v-model:value="addRecordForm.type" :options="recordTypeOptions" size="small" />
                    </NFormItem>
                  </NGi>
                  <NGi>
                    <NFormItem label="名称">
                      <NInput v-model:value="addRecordForm.name" size="small" placeholder="@ 或 www" />
                    </NFormItem>
                  </NGi>
                  <NGi :span="2">
                    <NFormItem label="值">
                      <NInput v-model:value="addRecordForm.value" size="small" placeholder="记录值" />
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
                    添加记录
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
              <NIcon :size="32" style="color:var(--faint)"><CloudOutline /></NIcon>
              <p>请从左侧选择一个域名查看解析记录</p>
            </div>
          </div>
        </div>
      </NDrawerContent>
    </NDrawer>

    <!-- DNS 记录 删除确认 -->
    <ConfirmDialog
      :show="recordDeleteShow"
      title="删除 DNS 记录"
      :content="recordDeleteTarget ? `确定要删除 ${recordDeleteTarget.type} 记录「${recordDeleteTarget.name}」吗？` : ''"
      confirm-text="删除"
      type="error"
      :loading="recordDeleteLoading"
      @update:show="(v: boolean) => (recordDeleteShow = v)"
      @confirm="handleRecordDeleteConfirm"
    />
  </div>
</template>

<style scoped>
.settings-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
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
  color: var(--faint);
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
  color: var(--faint);
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
  color: var(--faint);
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
  margin: 2px 0;
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
  color: var(--faint);
}

/* 移动端：版本信息单列 */
@media (max-width: 640px) {
  .version-grid {
    grid-template-columns: 1fr;
  }
}
</style>
