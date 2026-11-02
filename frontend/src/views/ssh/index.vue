<script setup lang="ts">
import { ref, onMounted, computed, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { NTabs, NTabPane, NButton, NIcon, NForm, NFormItem, NInput, NInputNumber, NSwitch, NSelect, NAlert, NText, NSpace } from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import { AddOutline, TrashOutline, BanOutline, CheckmarkCircleOutline } from '@vicons/ionicons5'
import { useQueryState } from '@/composables/useQueryState'
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
import { isApiUnsupported } from '@/api/request'

const message = useMessage()
const { t } = useI18n()

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
    message.success(t('ssh.message.saveConfigSuccess'))
    showSaveConfirm.value = false
  } catch (e: any) {
    message.error(e.message || t('ssh.message.saveFailed'))
  } finally {
    configSaving.value = false
  }
}

/* ==================== Tab 2: 密钥管理 ==================== */

const keysLoading = ref(false)
const sshKeys = ref<SSHKey[]>([])
const keysApiAvailable = ref(true)
/** 临时性加载失败（非「接口未实现」）：渲染错误面板 + 重试 */
const keysError = ref<string | null>(null)

const showAddKeyModal = ref(false)
const addKeyLoading = ref(false)
const addKeyForm = ref({ key: '', comment: '' })

const deleteKeyDialog = ref({ show: false, fingerprint: '', comment: '' })
const deleteKeyLoading = ref(false)

const keyColumns = computed<DataTableColumns>(() => [
  {
    title: t('ssh.columns.type'),
    key: 'type',
    width: 120,
    render: (row: any) => {
      const t = (row.type as string).toUpperCase()
      const color = t.includes('ED25519') ? 'var(--acc2)' : t.includes('RSA') ? 'var(--acc3)' : 'var(--warn)'
      return h('span', { class: 'mono', style: { color } }, t)
    }
  },
  {
    title: t('ssh.columns.comment'),
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
    title: t('ssh.columns.createdAt'),
    key: 'created_at',
    width: 180,
    render: (row: any) => h('span', { class: 'mono' }, row.created_at)
  },
  {
    title: t('ssh.columns.actions'),
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
  keysError.value = null
  try {
    const res = await getSSHKeys()
    if (res.data && Array.isArray(res.data)) {
      sshKeys.value = res.data
      keysApiAvailable.value = true
    }
  } catch (e: any) {
    // 接口不存在（旧版后端）→ 提示版本不支持；其余失败 → 错误面板 + 重试
    if (isApiUnsupported(e)) {
      keysApiAvailable.value = false
    } else {
      keysError.value = e.message || t('common.loadFailed')
    }
    sshKeys.value = []
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
    message.warning(t('ssh.keys.enterPublicKey'))
    return
  }
  if (!validatePublicKey(addKeyForm.value.key)) {
    message.warning(t('ssh.keys.invalidFormat'))
    return
  }
  addKeyLoading.value = true
  try {
    await addSSHKey({
      public_key: addKeyForm.value.key.trim(),
      comment: addKeyForm.value.comment.trim() || undefined
    })
    message.success(t('ssh.keys.addSuccess'))
    showAddKeyModal.value = false
    addKeyForm.value = { key: '', comment: '' }
    loadSSHKeys()
  } catch (e: any) {
    message.error(e.message || t('ssh.keys.addFailed'))
  } finally {
    addKeyLoading.value = false
  }
}

async function handleDeleteKey() {
  deleteKeyLoading.value = true
  try {
    await deleteSSHKey(deleteKeyDialog.value.fingerprint)
    message.success(t('ssh.keys.deleteSuccess'))
    deleteKeyDialog.value.show = false
    loadSSHKeys()
  } catch (e: any) {
    message.error(e.message || t('ssh.keys.deleteFailed'))
  } finally {
    deleteKeyLoading.value = false
  }
}

/* ==================== Tab 3: 登录日志 ==================== */

// Tab / 分页 / 筛选 URL 化（D19）：刷新与深链可还原现场
const activeTab = useQueryState('tab', 'config')

const logsLoading = ref(false)
const loginLogs = ref<SSHLoginLog[]>([])
const logsTotal = ref(0)
const logsPage = useQueryState('page', 1)
const logsPageSize = ref(20)
const logStatusFilter = useQueryState<'all' | 'success' | 'failed'>('status', 'all')
const logsApiAvailable = ref(true)
/** 临时性加载失败（非「接口未实现」）：渲染错误面板 + 重试 */
const logsError = ref<string | null>(null)

const logColumns = computed<DataTableColumns>(() => [
  {
    title: t('ssh.logs.columns.time'),
    key: 'time',
    width: 180,
    render: (row: any) => h('span', { class: 'mono' }, row.time)
  },
  {
    title: t('ssh.logs.columns.user'),
    key: 'user',
    width: 120,
    render: (row: any) => h('span', { class: 'mono' }, row.user)
  },
  {
    title: t('ssh.logs.columns.ip'),
    key: 'ip',
    width: 160,
    render: (row: any) => h('span', { class: 'mono' }, row.ip)
  },
  {
    title: t('ssh.logs.columns.method'),
    key: 'method',
    width: 120,
    render: (row: any) => h('span', { class: 'mono' }, row.method)
  },
  {
    title: t('ssh.logs.columns.status'),
    key: 'status',
    width: 100,
    render: (row: any) =>
      h(StatusTag, {
        status: row.status === 'success' ? 'running' : 'error',
        text: row.status === 'success' ? t('ssh.status.success') : t('ssh.status.failed')
      })
  }
])

const logRowProps = computed(() => (row: any) => ({
  style: row.status === 'failed' ? 'background: var(--err-soft); opacity: 0.6;' : undefined
}))

const statusFilterOptions = computed(() => [
  { label: t('ssh.filter.all'), value: 'all' },
  { label: t('ssh.status.success'), value: 'success' },
  { label: t('ssh.status.failed'), value: 'failed' }
])

async function loadSSHLogs() {
  logsLoading.value = true
  logsError.value = null
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
  } catch (e: any) {
    if (isApiUnsupported(e)) {
      logsApiAvailable.value = false
    } else {
      logsError.value = e.message || t('common.loadFailed')
    }
    loginLogs.value = []
    logsTotal.value = 0
  } finally {
    logsLoading.value = false
  }
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
/** 临时性加载失败（非「接口未实现」）：渲染错误面板 + 重试 */
const f2bError = ref<string | null>(null)

const banIPInput = ref('')
const banIPLoading = ref(false)
const banReasonInput = ref('')

const unbanDialog = ref({ show: false, ip: '' })
const unbanLoading = ref(false)

const f2bStatCards = computed<Array<{ title: string; value: string | number; unit?: string; accent: string }>>(() => [
  {
    title: t('ssh.fail2ban.stats.protectionStatus'),
    value: f2bStatus.value.installed ? (f2bStatus.value.enabled ? t('ssh.fail2ban.stats.enabled') : t('ssh.fail2ban.stats.disabled')) : t('ssh.fail2ban.stats.notInstalled'),
    accent: f2bStatus.value.installed ? (f2bStatus.value.enabled ? 'ok' : 'warn') : 'info'
  },
  {
    title: t('ssh.fail2ban.stats.bannedIps'),
    value: f2bStatus.value.banned_ips.length,
    accent: 'err'
  },
  {
    title: t('ssh.form.maxRetry'),
    value: f2bConfig.value.max_retry,
    unit: t('ssh.fail2ban.stats.unitTimes'),
    accent: 'warn'
  },
  {
    title: t('ssh.fail2ban.stats.banDuration'),
    value: Math.round(f2bConfig.value.ban_time / 60),
    unit: t('ssh.fail2ban.stats.unitMinutes'),
    accent: 'acc'
  }
])

const bannedIPColumns = computed<DataTableColumns>(() => [
  {
    title: t('ssh.columns.ip'),
    key: 'ip',
    width: 160,
    render: (row: any) => h('span', { class: 'mono' }, row.ip)
  },
  {
    title: t('ssh.fail2ban.columns.jail'),
    key: 'jail',
    width: 100,
    render: (row: any) => h('span', { class: 'mono' }, row.jail || 'sshd')
  },
  {
    title: t('ssh.fail2ban.columns.bannedAt'),
    key: 'banned_at',
    width: 180,
    render: (row: any) => h('span', { class: 'mono' }, row.banned_at)
  },
  {
    title: t('ssh.fail2ban.columns.reason'),
    key: 'reason',
    ellipsis: { tooltip: true },
    render: (row: any) => row.reason || t('ssh.fail2ban.defaultReason')
  },
  {
    title: t('ssh.columns.actions'),
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
        () => t('ssh.actions.unban')
      )
  }
])

async function loadFail2ban() {
  f2bLoading.value = true
  f2bError.value = null
  try {
    const res = await getFail2banStatus()
    if (res.data) {
      f2bStatus.value = { ...res.data }
      f2bConfig.value.enabled = res.data.enabled
      f2bApiAvailable.value = true
    }
  } catch (e: any) {
    if (isApiUnsupported(e)) {
      f2bApiAvailable.value = false
      f2bStatus.value = { enabled: false, installed: false, banned_ips: [] }
    } else {
      f2bError.value = e.message || t('common.loadFailed')
    }
  } finally {
    f2bLoading.value = false
  }
}

async function handleSaveF2bConfig() {
  f2bConfigSaving.value = true
  try {
    await updateFail2banConfig({ ...f2bConfig.value })
    message.success(t('ssh.fail2ban.saveSuccess'))
    loadFail2ban()
  } catch (e: any) {
    message.error(e.message || t('ssh.message.saveFailed'))
  } finally {
    f2bConfigSaving.value = false
  }
}

async function handleUnban() {
  unbanLoading.value = true
  try {
    await unbanIP(unbanDialog.value.ip)
    message.success(t('ssh.message.unbanSuccess', { ip: unbanDialog.value.ip }))
    unbanDialog.value.show = false
    loadFail2ban()
  } catch (e: any) {
    message.error(e.message || t('ssh.message.unbanFailed'))
  } finally {
    unbanLoading.value = false
  }
}

async function handleBanIP() {
  if (!banIPInput.value.trim()) {
    message.warning(t('ssh.message.enterBanIp'))
    return
  }
  banIPLoading.value = true
  try {
    await banIP(banIPInput.value.trim(), banReasonInput.value.trim() || undefined)
    message.success(t('ssh.message.banSuccess', { ip: banIPInput.value }))
    banIPInput.value = ''
    banReasonInput.value = ''
    loadFail2ban()
  } catch (e: any) {
    message.error(e.message || t('ssh.message.banFailed'))
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
    <NTabs v-model:value="activeTab" type="line" animated>
      <!-- ==================== Tab 1: SSH 配置 ==================== -->
      <NTabPane name="config" :tab="t('ssh.tabs.config')">
        <PageCard :title="t('ssh.config.title')" :subtitle="t('ssh.config.subtitle')">
          <NAlert
            v-if="!configApiAvailable"
            type="warning"
            :show-icon="true"
            style="margin-bottom: var(--space-4)"
          >
            {{ t('ssh.config.apiUnavailable') }}
          </NAlert>

          <NForm label-placement="left" label-width="180" :show-feedback="false">
            <NFormItem :label="t('ssh.form.port')">
              <NInputNumber
                v-model:value="sshConfig.port"
                :min="1"
                :max="65535"
                placeholder="22"
                style="max-width: var(--input-lg)"
              />
              <NText depth="3" style="margin-left: var(--space-3); font-size: var(--fs-caption)">
                {{ t('ssh.form.portHint') }}
              </NText>
            </NFormItem>

            <NFormItem :label="t('ssh.form.permitRootLogin')">
              <NSwitch v-model:value="sshConfig.permit_root_login" />
              <StatusTag
                v-if="!sshConfig.permit_root_login"
                status="running"
                :text="t('ssh.status.safe')"
                style="margin-left: var(--space-3)"
              />
              <StatusTag
                v-else
                status="warning"
                :text="t('ssh.status.risk')"
                style="margin-left: var(--space-3)"
              />
            </NFormItem>

            <NFormItem :label="t('ssh.form.permitPasswordLogin')">
              <NSwitch v-model:value="sshConfig.permit_password_login" />
              <StatusTag
                v-if="!sshConfig.permit_password_login"
                status="running"
                :text="t('ssh.status.safeKeyOnly')"
                style="margin-left: var(--space-3)"
              />
              <StatusTag
                v-else
                status="warning"
                :text="t('ssh.status.risk')"
                style="margin-left: var(--space-3)"
              />
            </NFormItem>

            <NFormItem :label="t('ssh.form.maxAuthTries')">
              <NInputNumber
                v-model:value="sshConfig.max_auth_tries"
                :min="1"
                :max="20"
                placeholder="6"
                style="max-width: var(--input-lg)"
              />
            </NFormItem>

            <NFormItem :label="t('ssh.form.clientAliveInterval')">
              <NInputNumber
                v-model:value="sshConfig.client_alive_interval"
                :min="0"
                placeholder="0"
                style="max-width: var(--input-lg)"
              />
              <NText depth="3" style="margin-left: var(--space-3); font-size: var(--fs-caption)">
                {{ t('ssh.form.clientAliveHint') }}
              </NText>
            </NFormItem>

            <NFormItem :label="t('ssh.form.allowUsers')">
              <NInput
                v-model:value="sshConfig.allow_users"
                :placeholder="t('ssh.form.allowUsersPlaceholder')"
                style="max-width: var(--content-max-sm)"
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
                {{ t('ssh.actions.saveConfig') }}
              </NButton>
            </NFormItem>
          </NForm>
        </PageCard>
      </NTabPane>

      <!-- ==================== Tab 2: 密钥管理 ==================== -->
      <NTabPane name="keys" :tab="t('ssh.tabs.keys')">
        <PageCard :title="t('ssh.keys.title')" :subtitle="t('ssh.keys.subtitle')">
          <template #extra>
            <NButton type="primary" size="small" @click="showAddKeyModal = true">
              <template #icon>
                <NIcon><AddOutline /></NIcon>
              </template>
              {{ t('ssh.actions.addKey') }}
            </NButton>
          </template>

          <NAlert
            v-if="!keysApiAvailable"
            type="warning"
            :show-icon="true"
            style="margin-bottom: var(--space-4)"
          >
            {{ t('ssh.keys.apiUnavailable') }}
          </NAlert>

          <DataTable
            :columns="keyColumns"
            :data="sshKeys"
            :loading="keysLoading"
            :row-key="(row: SSHKey) => row.fingerprint"
            :pagination="false"
            :error="keysError"
            @retry="loadSSHKeys"
          />
        </PageCard>
      </NTabPane>

      <!-- ==================== Tab 3: 登录日志 ==================== -->
      <NTabPane name="logs" :tab="t('ssh.tabs.logs')">
        <PageCard :title="t('ssh.logs.title')" :subtitle="t('ssh.logs.subtitle')">
          <div class="ssh__filter">
            <NSelect
              v-model:value="logStatusFilter"
              :placeholder="t('ssh.logs.statusPlaceholder')"
              size="small"
              style="width: var(--input-sm)"
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
            {{ t('ssh.logs.apiUnavailable') }}
          </NAlert>

          <DataTable
            :columns="logColumns"
            :data="loginLogs"
            :loading="logsLoading"
            :row-key="(row: any) => `${row.time}-${row.ip}`"
            :row-props="logRowProps"
            :error="logsError"
            @retry="loadSSHLogs"
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
      <NTabPane name="fail2ban" :tab="t('ssh.tabs.fail2ban')">
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
          <template #header>{{ t('ssh.fail2ban.notInstalledHeader') }}</template>
          {{ t('ssh.fail2ban.installHint') }}<span class="mono">apt install fail2ban</span> {{ t('ssh.fail2ban.installOr') }}
          <span class="mono">yum install fail2ban</span>
        </NAlert>

        <NAlert
          v-if="!f2bApiAvailable"
          type="warning"
          :show-icon="true"
          style="margin-bottom: var(--space-4)"
        >
          {{ t('ssh.fail2ban.apiUnavailable') }}
        </NAlert>

        <PageCard :title="t('ssh.fail2ban.configTitle')" :subtitle="t('ssh.fail2ban.configSubtitle')" style="margin-bottom: var(--space-4)">
          <NForm label-placement="left" label-width="180" :show-feedback="false">
            <NFormItem :label="t('ssh.form.maxRetry')">
              <NInputNumber
                v-model:value="f2bConfig.max_retry"
                :min="1"
                :max="20"
                placeholder="5"
                style="max-width: var(--input-lg)"
              />
            </NFormItem>
            <NFormItem :label="t('ssh.form.banDuration')">
              <NInputNumber
                v-model:value="f2bConfig.ban_time"
                :min="1"
                placeholder="10"
                style="max-width: var(--input-lg)"
              />
            </NFormItem>
            <NFormItem :label="t('ssh.form.findWindow')">
              <NInputNumber
                v-model:value="f2bConfig.find_time"
                :min="1"
                placeholder="10"
                style="max-width: var(--input-lg)"
              />
            </NFormItem>
            <NFormItem>
              <NButton
                type="primary"
                :loading="f2bConfigSaving"
                :disabled="!f2bStatus.installed"
                @click="handleSaveF2bConfig"
              >
                {{ t('ssh.actions.saveConfig') }}
              </NButton>
            </NFormItem>
          </NForm>
        </PageCard>

        <PageCard :title="t('ssh.fail2ban.manualTitle')" :subtitle="t('ssh.fail2ban.manualSubtitle')" style="margin-bottom: var(--space-4)">
          <NSpace align="center">
            <NInput
              v-model:value="banIPInput"
              :placeholder="t('ssh.fail2ban.ipPlaceholder')"
              style="width: var(--input-2xl)"
            />
            <NInput
              v-model:value="banReasonInput"
              :placeholder="t('ssh.fail2ban.reasonPlaceholder')"
              style="width: var(--input-xl)"
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
              {{ t('ssh.actions.ban') }}
            </NButton>
          </NSpace>
        </PageCard>

        <PageCard :title="t('ssh.fail2ban.listTitle')" :subtitle="t('ssh.fail2ban.listSubtitle')">
          <DataTable
            :columns="bannedIPColumns"
            :data="f2bStatus.banned_ips"
            :loading="f2bLoading"
            :row-key="(row: BannedIP) => row.ip"
            :pagination="false"
            :error="f2bError"
            @retry="loadFail2ban"
          />
        </PageCard>
      </NTabPane>
    </NTabs>

    <!-- ==================== 添加公钥对话框 ==================== -->
    <ModalForm
      :show="showAddKeyModal"
      :title="t('ssh.keys.modalTitle')"
      :loading="addKeyLoading"
      :width="600"
      @update:show="(v: boolean) => (showAddKeyModal = v)"
      @submit="handleAddKey"
    >
      <NFormItem :label="t('ssh.keys.keyLabel')" required>
        <NInput
          v-model:value="addKeyForm.key"
          type="textarea"
          :placeholder="t('ssh.keys.keyPlaceholder')"
          :autosize="{ minRows: 4, maxRows: 8 }"
        />
      </NFormItem>
      <NFormItem :label="t('ssh.keys.commentLabel')">
        <NInput v-model:value="addKeyForm.comment" :placeholder="t('ssh.keys.commentPlaceholder')" />
      </NFormItem>
      <NAlert type="info" :show-icon="true" style="margin-top: var(--space-2)">
        {{ t('ssh.keys.pasteHintPrefix') }}<span class="mono">~/.ssh/id_rsa.pub</span> {{ t('ssh.keys.pasteHintOr') }}
        <span class="mono">~/.ssh/id_ed25519.pub</span>{{ t('ssh.keys.pasteHintSuffix') }}
      </NAlert>
    </ModalForm>

    <!-- ==================== 保存 SSH 配置确认 ==================== -->
    <ConfirmDialog
      :show="showSaveConfirm"
      :title="t('ssh.confirm.saveTitle')"
      :content="t('ssh.confirm.saveContent')"
      :confirm-text="t('ssh.confirm.saveConfirmText')"
      type="warning"
      :loading="configSaving"
      @update:show="(v: boolean) => (showSaveConfirm = v)"
      @confirm="handleSaveConfig"
    />

    <!-- ==================== 删除密钥确认 ==================== -->
    <ConfirmDialog
      :show="deleteKeyDialog.show"
      :title="t('ssh.confirm.deleteKeyTitle')"
      :content="t('ssh.confirm.deleteKeyContent', { name: deleteKeyDialog.comment || deleteKeyDialog.fingerprint })"
      :confirm-text="t('ssh.confirm.deleteKeyConfirmText')"
      type="error"
      :loading="deleteKeyLoading"
      @update:show="(v: boolean) => (deleteKeyDialog.show = v)"
      @confirm="handleDeleteKey"
    />

    <!-- ==================== 解封确认 ==================== -->
    <ConfirmDialog
      :show="unbanDialog.show"
      :title="t('ssh.confirm.unbanTitle')"
      :content="t('ssh.confirm.unbanContent', { ip: unbanDialog.ip })"
      :confirm-text="t('ssh.confirm.unbanConfirmText')"
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
  gap: var(--space-6);
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
