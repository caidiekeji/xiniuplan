<script setup lang="ts">
import { ref, onMounted, computed, h } from 'vue'
import {
  NTabs,
  NTabPane,
  NButton,
  NIcon,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSwitch,
  NSelect,
  NAlert,
  NText,
  NEmpty,
  NSpace,
  useMessage
} from 'naive-ui'
import {
  AddOutline,
  TrashOutline,
  ShieldHalfOutline,
  BanOutline,
  CheckmarkCircleOutline
} from '@vicons/ionicons5'
import {
  PageCard,
  DataTable,
  ModalForm,
  StatusTag,
  StatCard,
  ConfirmDialog
} from '@/components'
import type { DataTableColumns } from 'naive-ui'
import {
  getSSHConfig,
  updateSSHConfig,
  getSSHKeys,
  addSSHKey,
  deleteSSHKey,
  getSSHLogs,
  getFail2banStatus,
  updateFail2banConfig,
  unbanIP,
  banIP,
  type SSHConfig,
  type SSHKey,
  type SSHLoginLog,
  type Fail2banConfig,
  type Fail2banStatus,
  type BannedIP
} from '@/api/ssh'

const message = useMessage()

/* ==================== Tab 1: SSH 配置 ==================== */

const configLoading = ref(false)
const configSaving = ref(false)
const sshConfig = ref<SSHConfig>({
  port: 22,
  permit_root_login: false,
  permit_password_login: false,
  max_auth_tries: 6,
  client_alive_interval: 0,
  allow_users: ''
})
const configApiAvailable = ref(true)

const showSaveConfirm = ref(false)

async function loadSSHConfig() {
  configLoading.value = true
  try {
    const res = await getSSHConfig()
    if (res.data) {
      sshConfig.value = { ...res.data }
      configApiAvailable.value = true
    }
  } catch {
    configApiAvailable.value = false
  } finally {
    configLoading.value = false
  }
}

function handleSaveClick() {
  showSaveConfirm.value = true
}

async function handleSaveConfig() {
  configSaving.value = true
  try {
    await updateSSHConfig({ ...sshConfig.value })
    message.success('SSH 配置已保存')
    showSaveConfirm.value = false
  } catch (e: any) {
    message.error(e.message || '保存失败')
  } finally {
    configSaving.value = false
  }
}

/* ==================== Tab 2: 密钥管理 ==================== */

const keysLoading = ref(false)
const sshKeys = ref<SSHKey[]>([])
const keysApiAvailable = ref(true)

const showAddKeyModal = ref(false)
const addKeyLoading = ref(false)
const addKeyForm = ref({ key: '', comment: '' })

const deleteKeyDialog = ref({ show: false, fingerprint: '', comment: '' })
const deleteKeyLoading = ref(false)

const keyColumns = computed<DataTableColumns>(() => [
  {
    title: '密钥类型',
    key: 'type',
    width: 120,
    render: (row: any) => {
      const t = (row.type as string).toUpperCase()
      const color = t.includes('ED25519') ? 'var(--acc2)' : t.includes('RSA') ? 'var(--acc3)' : 'var(--warn)'
      return h('span', { class: 'mono', style: { color } }, t)
    }
  },
  {
    title: '注释/名称',
    key: 'comment',
    width: 180,
    render: (row: any) => row.comment || '—'
  },
  {
    title: 'Fingerprint',
    key: 'fingerprint',
    ellipsis: { tooltip: true },
    render: (row: any) => h('span', { class: 'mono' }, row.fingerprint)
  },
  {
    title: '添加时间',
    key: 'created_at',
    width: 180,
    render: (row: any) => h('span', { class: 'mono' }, row.created_at)
  },
  {
    title: '操作',
    key: 'actions',
    width: 80,
    render: (row: any) =>
      h(
        NButton,
        {
          size: 'tiny',
          type: 'error',
          ghost: true,
          onClick: () => {
            deleteKeyDialog.value = { show: true, fingerprint: row.fingerprint, comment: row.comment }
          }
        },
        () => h(NIcon, { size: 14 }, () => h(TrashOutline))
      )
  }
])

async function loadSSHKeys() {
  keysLoading.value = true
  try {
    const res = await getSSHKeys()
    if (res.data && Array.isArray(res.data)) {
      sshKeys.value = res.data
      keysApiAvailable.value = true
    }
  } catch {
    keysApiAvailable.value = false
  } finally {
    keysLoading.value = false
  }
}

function validatePublicKey(key: string): boolean {
  const trimmed = key.trim()
  return /^(ssh-rsa|ssh-ed25519|ecdsa-sha2-nistp\d{3})\s+/.test(trimmed)
}

async function handleAddKey() {
  if (!addKeyForm.value.key.trim()) {
    message.warning('请输入公钥内容')
    return
  }
  if (!validatePublicKey(addKeyForm.value.key)) {
    message.warning('公钥格式不正确，应以 ssh-rsa / ssh-ed25519 / ecdsa-sha2- 开头')
    return
  }
  addKeyLoading.value = true
  try {
    await addSSHKey({
      public_key: addKeyForm.value.key.trim(),
      comment: addKeyForm.value.comment.trim() || undefined
    })
    message.success('公钥添加成功')
    showAddKeyModal.value = false
    addKeyForm.value = { key: '', comment: '' }
    loadSSHKeys()
  } catch (e: any) {
    message.error(e.message || '添加失败')
  } finally {
    addKeyLoading.value = false
  }
}

async function handleDeleteKey() {
  deleteKeyLoading.value = true
  try {
    await deleteSSHKey(deleteKeyDialog.value.fingerprint)
    message.success('密钥已删除')
    deleteKeyDialog.value.show = false
    loadSSHKeys()
  } catch (e: any) {
    message.error(e.message || '删除失败')
  } finally {
    deleteKeyLoading.value = false
  }
}

/* ==================== Tab 3: 登录日志 ==================== */

const logsLoading = ref(false)
const loginLogs = ref<SSHLoginLog[]>([])
const logsTotal = ref(0)
const logsPage = ref(1)
const logsPageSize = ref(20)
const logStatusFilter = ref<'all' | 'success' | 'failed'>('all')
const logsApiAvailable = ref(true)

const logColumns = computed<DataTableColumns>(() => [
  {
    title: '时间',
    key: 'time',
    width: 180,
    render: (row: any) => h('span', { class: 'mono' }, row.time)
  },
  {
    title: '用户',
    key: 'user',
    width: 120,
    render: (row: any) => h('span', { class: 'mono' }, row.user)
  },
  {
    title: 'IP 地址',
    key: 'ip',
    width: 160,
    render: (row: any) => h('span', { class: 'mono' }, row.ip)
  },
  {
    title: '登录方式',
    key: 'method',
    width: 120,
    render: (row: any) => h('span', { class: 'mono' }, row.method)
  },
  {
    title: '结果',
    key: 'status',
    width: 100,
    render: (row: any) =>
      h(StatusTag, {
        status: row.status === 'success' ? 'running' : 'error',
        text: row.status === 'success' ? '成功' : '失败'
      })
  }
])

const logRowProps = computed(() => (row: any) => ({
  style: row.status === 'failed' ? 'background: var(--err-soft); opacity: 0.6;' : undefined
}))

const statusFilterOptions = [
  { label: '全部', value: 'all' },
  { label: '成功', value: 'success' },
  { label: '失败', value: 'failed' }
]

async function loadSSHLogs() {
  logsLoading.value = true
  try {
    const res = await getSSHLogs(logsPage.value, logsPageSize.value)
    if (res.data?.items) {
      let items = res.data.items
      if (logStatusFilter.value !== 'all') {
        items = items.filter((l) => l.status === logStatusFilter.value)
      }
      loginLogs.value = items
      logsTotal.value = res.data.total
      logsApiAvailable.value = true
    }
  } catch {
    logsApiAvailable.value = false
    generateMockLogs()
  } finally {
    logsLoading.value = false
  }
}

function generateMockLogs() {
  loginLogs.value = [
    { time: '2026-09-09 10:23:15', user: 'root', ip: '192.168.1.100', status: 'success', method: 'publickey' },
    { time: '2026-09-09 09:45:22', user: 'admin', ip: '10.0.0.55', status: 'success', method: 'password' },
    { time: '2026-09-09 08:12:03', user: 'root', ip: '45.33.32.156', status: 'failed', method: 'password' },
    { time: '2026-09-09 08:11:58', user: 'root', ip: '45.33.32.156', status: 'failed', method: 'password' },
    { time: '2026-09-09 08:11:52', user: 'admin', ip: '45.33.32.156', status: 'failed', method: 'password' },
    { time: '2026-09-08 22:30:10', user: 'deploy', ip: '172.16.0.10', status: 'success', method: 'publickey' },
    { time: '2026-09-08 18:05:44', user: 'root', ip: '203.0.113.50', status: 'failed', method: 'password' },
    { time: '2026-09-08 15:20:33', user: 'ubuntu', ip: '192.168.1.200', status: 'success', method: 'publickey' }
  ]
  logsTotal.value = loginLogs.value.length
}

/* ==================== Tab 4: Fail2ban 防护 ==================== */

const f2bLoading = ref(false)
const f2bStatus = ref<Fail2banStatus>({
  enabled: false,
  installed: true,
  banned_ips: []
})
const f2bConfig = ref<Fail2banConfig>({
  enabled: false,
  max_retry: 5,
  ban_time: 600,
  find_time: 600
})
const f2bConfigSaving = ref(false)
const f2bApiAvailable = ref(true)

const banIPInput = ref('')
const banIPLoading = ref(false)
const banReasonInput = ref('')

const unbanDialog = ref({ show: false, ip: '' })
const unbanLoading = ref(false)

const f2bStatCards = computed<Array<{ title: string; value: string | number; unit?: string; accent: string }>>(() => [
  {
    title: '防护状态',
    value: f2bStatus.value.installed ? (f2bStatus.value.enabled ? '已启用' : '已禁用') : '未安装',
    accent: f2bStatus.value.installed ? (f2bStatus.value.enabled ? 'ok' : 'warn') : 'info'
  },
  {
    title: '当前封禁 IP',
    value: f2bStatus.value.banned_ips.length,
    accent: 'err'
  },
  {
    title: '最大重试次数',
    value: f2bConfig.value.max_retry,
    unit: '次',
    accent: 'warn'
  },
  {
    title: '封禁时长',
    value: Math.round(f2bConfig.value.ban_time / 60),
    unit: '分钟',
    accent: 'acc'
  }
])

const bannedIPColumns = computed<DataTableColumns>(() => [
  {
    title: 'IP 地址',
    key: 'ip',
    width: 160,
    render: (row: any) => h('span', { class: 'mono' }, row.ip)
  },
  {
    title: '监狱',
    key: 'jail',
    width: 100,
    render: (row: any) => h('span', { class: 'mono' }, row.jail || 'sshd')
  },
  {
    title: '封禁时间',
    key: 'banned_at',
    width: 180,
    render: (row: any) => h('span', { class: 'mono' }, row.banned_at)
  },
  {
    title: '原因',
    key: 'reason',
    ellipsis: { tooltip: true },
    render: (row: any) => row.reason || '暴力破解'
  },
  {
    title: '操作',
    key: 'actions',
    width: 80,
    render: (row: any) =>
      h(
        NButton,
        {
          size: 'tiny',
          type: 'primary',
          ghost: true,
          onClick: () => {
            unbanDialog.value = { show: true, ip: row.ip }
          }
        },
        () => '解封'
      )
  }
])

async function loadFail2ban() {
  f2bLoading.value = true
  try {
    const res = await getFail2banStatus()
    if (res.data) {
      f2bStatus.value = { ...res.data }
      f2bConfig.value.enabled = res.data.enabled
      f2bApiAvailable.value = true
    }
  } catch {
    f2bApiAvailable.value = false
    generateMockFail2ban()
  } finally {
    f2bLoading.value = false
  }
}

function generateMockFail2ban() {
  f2bStatus.value = {
    enabled: true,
    installed: true,
    banned_ips: [
      { ip: '45.33.32.156', jail: 'sshd', banned_at: '2026-09-09 08:12:10', reason: '多次密码登录失败' },
      { ip: '203.0.113.50', jail: 'sshd', banned_at: '2026-09-08 18:06:00', reason: '暴力破解尝试' }
    ]
  }
  f2bConfig.value.enabled = true
}

async function handleSaveF2bConfig() {
  f2bConfigSaving.value = true
  try {
    await updateFail2banConfig({ ...f2bConfig.value })
    message.success('Fail2ban 配置已保存')
    loadFail2ban()
  } catch (e: any) {
    message.error(e.message || '保存失败')
  } finally {
    f2bConfigSaving.value = false
  }
}

async function handleUnban() {
  unbanLoading.value = true
  try {
    await unbanIP(unbanDialog.value.ip)
    message.success(`IP ${unbanDialog.value.ip} 已解封`)
    unbanDialog.value.show = false
    loadFail2ban()
  } catch (e: any) {
    message.error(e.message || '解封失败')
  } finally {
    unbanLoading.value = false
  }
}

async function handleBanIP() {
  if (!banIPInput.value.trim()) {
    message.warning('请输入要封禁的 IP 地址')
    return
  }
  banIPLoading.value = true
  try {
    await banIP(banIPInput.value.trim(), banReasonInput.value.trim() || undefined)
    message.success(`IP ${banIPInput.value} 已封禁`)
    banIPInput.value = ''
    banReasonInput.value = ''
    loadFail2ban()
  } catch (e: any) {
    message.error(e.message || '封禁失败')
  } finally {
    banIPLoading.value = false
  }
}

/* ==================== 生命周期 ==================== */

onMounted(() => {
  loadSSHConfig()
  loadSSHKeys()
  loadSSHLogs()
  loadFail2ban()
})
</script>

<template>
  <div class="ssh-page">
    <NTabs type="line" animated>
      <!-- ==================== Tab 1: SSH 配置 ==================== -->
      <NTabPane name="config" tab="SSH 配置">
        <PageCard title="SSH 服务配置" subtitle="管理 SSH 服务端安全参数">
          <NAlert
            v-if="!configApiAvailable"
            type="warning"
            :show-icon="true"
            style="margin-bottom: var(--space-4)"
          >
            当前环境无法连接 SSH 服务（Windows 开发环境无 SSH 服务），以下为默认配置展示。
          </NAlert>

          <NForm label-placement="left" label-width="180" :show-feedback="false">
            <NFormItem label="SSH 端口">
              <NInputNumber
                v-model:value="sshConfig.port"
                :min="1"
                :max="65535"
                placeholder="22"
                style="max-width: 200px"
              />
              <NText depth="3" style="margin-left: var(--space-3); font-size: var(--fs-caption)">
                修改端口后需确保防火墙已放行新端口
              </NText>
            </NFormItem>

            <NFormItem label="允许 Root 登录">
              <NSwitch v-model:value="sshConfig.permit_root_login" />
              <StatusTag
                v-if="!sshConfig.permit_root_login"
                status="running"
                text="安全"
                style="margin-left: var(--space-3)"
              />
              <StatusTag
                v-else
                status="warning"
                text="风险"
                style="margin-left: var(--space-3)"
              />
            </NFormItem>

            <NFormItem label="允许密码登录">
              <NSwitch v-model:value="sshConfig.permit_password_login" />
              <StatusTag
                v-if="!sshConfig.permit_password_login"
                status="running"
                text="安全（仅密钥）"
                style="margin-left: var(--space-3)"
              />
              <StatusTag
                v-else
                status="warning"
                text="风险"
                style="margin-left: var(--space-3)"
              />
            </NFormItem>

            <NFormItem label="最大认证尝试次数">
              <NInputNumber
                v-model:value="sshConfig.max_auth_tries"
                :min="1"
                :max="20"
                placeholder="6"
                style="max-width: 200px"
              />
            </NFormItem>

            <NFormItem label="客户端保活间隔(秒)">
              <NInputNumber
                v-model:value="sshConfig.client_alive_interval"
                :min="0"
                placeholder="0"
                style="max-width: 200px"
              />
              <NText depth="3" style="margin-left: var(--space-3); font-size: var(--fs-caption)">
                0 表示不发送保活包
              </NText>
            </NFormItem>

            <NFormItem label="允许登录用户">
              <NInput
                v-model:value="sshConfig.allow_users"
                placeholder="逗号分隔，留空表示允许全部用户"
                style="max-width: 400px"
              />
            </NFormItem>

            <NFormItem>
              <NButton
                type="primary"
                :loading="configSaving"
                @click="handleSaveClick"
              >
                <template #icon>
                  <NIcon><CheckmarkCircleOutline /></NIcon>
                </template>
                保存配置
              </NButton>
            </NFormItem>
          </NForm>
        </PageCard>
      </NTabPane>

      <!-- ==================== Tab 2: 密钥管理 ==================== -->
      <NTabPane name="keys" tab="密钥管理">
        <PageCard title="SSH 公钥管理" subtitle="管理授权登录的公钥">
          <template #extra>
            <NButton type="primary" size="small" @click="showAddKeyModal = true">
              <template #icon>
                <NIcon><AddOutline /></NIcon>
              </template>
              添加公钥
            </NButton>
          </template>

          <NAlert
            v-if="!keysApiAvailable"
            type="warning"
            :show-icon="true"
            style="margin-bottom: var(--space-4)"
          >
            当前环境无法读取 SSH 密钥列表（Windows 开发环境无 SSH 服务）。
          </NAlert>

          <DataTable
            :columns="keyColumns"
            :data="sshKeys"
            :loading="keysLoading"
            :row-key="(row: SSHKey) => row.fingerprint"
            :pagination="false"
          />
        </PageCard>
      </NTabPane>

      <!-- ==================== Tab 3: 登录日志 ==================== -->
      <NTabPane name="logs" tab="登录日志">
        <PageCard title="SSH 登录日志" subtitle="查看 SSH 登录记录与失败尝试">
          <div class="ssh__filter">
            <NSelect
              v-model:value="logStatusFilter"
              placeholder="全部状态"
              size="small"
              style="width: 140px"
              :options="statusFilterOptions"
              @update:value="() => { logsPage = 1; loadSSHLogs() }"
            />
          </div>

          <NAlert
            v-if="!logsApiAvailable"
            type="warning"
            :show-icon="true"
            style="margin-bottom: var(--space-4)"
          >
            当前环境无法读取 SSH 日志（Windows 开发环境无 SSH 服务），以下为示例数据。
          </NAlert>

          <DataTable
            :columns="logColumns"
            :data="loginLogs"
            :loading="logsLoading"
            :row-key="(row: any) => `${row.time}-${row.ip}`"
            :row-props="logRowProps"
            :pagination="{
              page: logsPage,
              pageSize: logsPageSize,
              itemCount: logsTotal,
              'onUpdate:page': (p: number) => { logsPage = p; loadSSHLogs() },
              'onUpdate:pageSize': (s: number) => { logsPageSize = s; logsPage = 1; loadSSHLogs() }
            }"
          />
        </PageCard>
      </NTabPane>

      <!-- ==================== Tab 4: Fail2ban 防护 ==================== -->
      <NTabPane name="fail2ban" tab="Fail2ban 防护">
        <div class="ssh__f2b-stats">
          <StatCard
            v-for="(stat, idx) in f2bStatCards"
            :key="idx"
            :title="stat.title"
            :value="stat.value"
            :unit="(stat as any).unit"
            :accent="stat.accent as any"
            :loading="f2bLoading"
          />
        </div>

        <NAlert
          v-if="!f2bStatus.installed"
          type="info"
          :show-icon="true"
          style="margin-bottom: var(--space-4)"
        >
          <template #header>Fail2ban 未安装</template>
          请先安装 Fail2ban：<span class="mono">apt install fail2ban</span> 或
          <span class="mono">yum install fail2ban</span>
        </NAlert>

        <NAlert
          v-if="!f2bApiAvailable"
          type="warning"
          :show-icon="true"
          style="margin-bottom: var(--space-4)"
        >
          当前环境无法连接 Fail2ban 服务（Windows 开发环境），以下为示例数据。
        </NAlert>

        <PageCard title="封禁配置" subtitle="配置 Fail2ban 自动封禁策略" style="margin-bottom: var(--space-4)">
          <NForm label-placement="left" label-width="180" :show-feedback="false">
            <NFormItem label="最大重试次数">
              <NInputNumber
                v-model:value="f2bConfig.max_retry"
                :min="1"
                :max="20"
                placeholder="5"
                style="max-width: 200px"
              />
            </NFormItem>
            <NFormItem label="封禁时长(分钟)">
              <NInputNumber
                v-model:value="f2bConfig.ban_time"
                :min="1"
                placeholder="10"
                style="max-width: 200px"
              />
            </NFormItem>
            <NFormItem label="检测窗口(分钟)">
              <NInputNumber
                v-model:value="f2bConfig.find_time"
                :min="1"
                placeholder="10"
                style="max-width: 200px"
              />
            </NFormItem>
            <NFormItem>
              <NButton
                type="primary"
                :loading="f2bConfigSaving"
                :disabled="!f2bStatus.installed"
                @click="handleSaveF2bConfig"
              >
                保存配置
              </NButton>
            </NFormItem>
          </NForm>
        </PageCard>

        <PageCard title="手动封禁" subtitle="主动封禁可疑 IP 地址" style="margin-bottom: var(--space-4)">
          <NSpace align="center">
            <NInput
              v-model:value="banIPInput"
              placeholder="输入 IP 地址，例如 192.168.1.100"
              style="width: 260px"
            />
            <NInput
              v-model:value="banReasonInput"
              placeholder="封禁原因（可选）"
              style="width: 220px"
            />
            <NButton
              type="error"
              :loading="banIPLoading"
              :disabled="!f2bStatus.installed"
              @click="handleBanIP"
            >
              <template #icon>
                <NIcon><BanOutline /></NIcon>
              </template>
              封禁
            </NButton>
          </NSpace>
        </PageCard>

        <PageCard title="封禁 IP 列表" subtitle="当前被 Fail2ban 封禁的 IP">
          <DataTable
            :columns="bannedIPColumns"
            :data="f2bStatus.banned_ips"
            :loading="f2bLoading"
            :row-key="(row: BannedIP) => row.ip"
            :pagination="false"
          />
        </PageCard>
      </NTabPane>
    </NTabs>

    <!-- ==================== 添加公钥对话框 ==================== -->
    <ModalForm
      :show="showAddKeyModal"
      title="添加 SSH 公钥"
      :loading="addKeyLoading"
      :width="600"
      @update:show="(v: boolean) => (showAddKeyModal = v)"
      @submit="handleAddKey"
    >
      <NFormItem label="公钥内容" required>
        <NInput
          v-model:value="addKeyForm.key"
          type="textarea"
          placeholder="ssh-rsa AAAA... 或 ssh-ed25519 AAAA..."
          :autosize="{ minRows: 4, maxRows: 8 }"
        />
      </NFormItem>
      <NFormItem label="注释">
        <NInput v-model:value="addKeyForm.comment" placeholder="例如：我的笔记本电脑" />
      </NFormItem>
      <NAlert type="info" :show-icon="true" style="margin-top: var(--space-2)">
        将 <span class="mono">~/.ssh/id_rsa.pub</span> 或
        <span class="mono">~/.ssh/id_ed25519.pub</span> 文件内容粘贴到此处
      </NAlert>
    </ModalForm>

    <!-- ==================== 保存 SSH 配置确认 ==================== -->
    <ConfirmDialog
      :show="showSaveConfirm"
      title="保存 SSH 配置"
      content="修改 SSH 配置可能导致无法连接服务器，请确保已测试新配置。确认继续？"
      confirm-text="确认保存"
      type="warning"
      :loading="configSaving"
      @update:show="(v: boolean) => (showSaveConfirm = v)"
      @confirm="handleSaveConfig"
    />

    <!-- ==================== 删除密钥确认 ==================== -->
    <ConfirmDialog
      :show="deleteKeyDialog.show"
      title="删除公钥"
      :content="`确定要删除公钥（${deleteKeyDialog.comment || deleteKeyDialog.fingerprint}）吗？删除后该密钥将无法用于登录。`"
      confirm-text="删除"
      type="error"
      :loading="deleteKeyLoading"
      @update:show="(v: boolean) => (deleteKeyDialog.show = v)"
      @confirm="handleDeleteKey"
    />

    <!-- ==================== 解封确认 ==================== -->
    <ConfirmDialog
      :show="unbanDialog.show"
      title="解封 IP"
      :content="`确定要解封 IP ${unbanDialog.ip} 吗？解封后该 IP 可再次尝试登录。`"
      confirm-text="解封"
      type="info"
      :loading="unbanLoading"
      @update:show="(v: boolean) => (unbanDialog.show = v)"
      @confirm="handleUnban"
    />
  </div>
</template>

<style scoped>
.ssh-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.ssh__filter {
  margin-bottom: var(--space-4);
}

.ssh__f2b-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}

@media (max-width: 1024px) {
  .ssh__f2b-stats {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .ssh__f2b-stats {
    grid-template-columns: 1fr;
  }
}
</style>
