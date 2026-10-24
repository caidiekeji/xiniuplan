<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NButton, NIcon, NForm, NFormItem, NInput, NSelect, NRadioGroup, NRadio,
  NSpace
} from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import {
  AddOutline, EyeOutline, RefreshOutline, TrashOutline
} from '@vicons/ionicons5'
import {
  PageCard, DataTable, ModalForm, ContentModal, StatusTag, ConfirmDialog, CodeEditor
} from '@/components'
import type { StatusType } from '@/components'
import {
  getSSLList, applySSL, renewSSL, deleteSSL,
  type SSLInfo, type SSLProvider, type SSLStatus, type ApplySSLData
} from '@/api/ssl'
import { getSites, type Site } from '@/api/site'
import { getDNSAccounts, type DNSAccount } from '@/api/dns'
import type { DataTableColumns } from 'naive-ui'

const message = useMessage()
const { t } = useI18n()

/* ========== 列表状态 ========== */
const loading = ref(false)
const sslList = ref<SSLInfo[]>([])
const sites = ref<Site[]>([])
const dnsAccounts = ref<DNSAccount[]>([])

const providerLabelMap = computed<Record<SSLProvider, string>>(() => ({
  letsencrypt: "Let's Encrypt",
  manual: t('ssl.provider.manual')
}))

const sslStatusMap = computed<Record<SSLStatus, { status: StatusType; text: string }>>(() => ({
  valid: { status: 'running', text: t('ssl.status.valid') },
  expiring: { status: 'warning', text: t('ssl.status.expiring') },
  expired: { status: 'error', text: t('ssl.status.expired') },
  pending: { status: 'pending', text: t('ssl.status.pending') }
}))

async function loadSSLList() {
  loading.value = true
  try {
    const res = await getSSLList()
    if (res.data && Array.isArray(res.data)) {
      sslList.value = res.data
    } else {
      sslList.value = []
    }
  } catch {
    sslList.value = []
    message.error(t('ssl.messages.loadListFailed'))
  } finally {
    loading.value = false
  }
}

async function loadSites() {
  try {
    const res = await getSites({ page: 1, pageSize: 100 })
    if (res.data?.items) {
      sites.value = res.data.items
    }
  } catch {
    sites.value = [
      { id: 1, domain: 'example.com', aliases: [], type: 'static', rootDir: '/www/example.com', status: 'running', ssl: true, createdAt: '2026-08-15' },
      { id: 2, domain: 'api.example.com', aliases: [], type: 'reverse_proxy', rootDir: '/www/api', status: 'running', ssl: true, createdAt: '2026-08-20' },
      { id: 3, domain: 'blog.example.com', aliases: [], type: 'runtime', rootDir: '/www/blog', status: 'stopped', ssl: false, createdAt: '2026-09-01' }
    ]
  }
}

async function loadDNSAccounts() {
  try {
    const res = await getDNSAccounts()
    if (res.data && Array.isArray(res.data)) {
      dnsAccounts.value = res.data.filter((a) => a.enabled)
    }
  } catch {
    dnsAccounts.value = []
  }
}

function getDaysRemaining(expireAt: string): number {
  const expire = new Date(expireAt).getTime()
  const now = Date.now()
  return Math.ceil((expire - now) / 86400000)
}

/* ========== 表格列 ========== */
const columns = computed<DataTableColumns>(() => [
  {
    title: t('ssl.columns.domain'),
    key: 'domain',
    width: 220,
    render: (row: any) => h('span', { class: 'mono', style: 'font-weight:500' }, row.domain)
  },
  {
    title: t('ssl.columns.provider'),
    key: 'provider',
    width: 160,
    render: (row: any) => h('span', { style: 'color:var(--sub)' }, providerLabelMap.value[row.provider as keyof typeof providerLabelMap.value])
  },
  {
    title: t('ssl.columns.expireAt'),
    key: 'expireAt',
    width: 200,
    render: (row: any) => {
      const days = getDaysRemaining(row.expireAt)
      const isWarning = days < 30 && days >= 0
      const isExpired = days < 0
      const color = isExpired ? 'var(--err)' : isWarning ? 'var(--warn)' : 'var(--sub)'
      return h('div', null, [
        h('div', { class: 'mono', style: `color:${color};font-weight:500` }, row.expireAt),
        h('div', { style: `font-size:var(--fs-caption);color:${color}` },
          isExpired ? t('ssl.expireText.expired', { days: Math.abs(days) }) : t('ssl.expireText.remaining', { days })
        )
      ])
    }
  },
  {
    title: t('ssl.columns.status'),
    key: 'status',
    width: 120,
    render: (row: any) => {
      const cfg = sslStatusMap.value[row.status as keyof typeof sslStatusMap.value]
      return h(StatusTag, { status: cfg.status, text: cfg.text })
    }
  },
  {
    title: t('ssl.columns.actions'),
    key: 'actions',
    width: 240,
    render: (row: any) => h(NSpace, { size: 'small', wrap: false }, {
      default: () => [
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          onClick: () => viewDetail(row)
        }, {
          icon: () => h(NIcon, null, { default: () => h(EyeOutline) }),
          default: () => t('ssl.actions.detail')
        }),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          type: 'primary',
          loading: renewLoadingId.value === row.id,
          onClick: () => handleRenew(row.id)
        }, {
          icon: () => h(NIcon, null, { default: () => h(RefreshOutline) }),
          default: () => t('ssl.actions.renew')
        }),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          type: 'error',
          onClick: () => openDelete(row)
        }, {
          icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
          default: () => t('ssl.actions.delete')
        })
      ]
    })
  }
])

/* ========== 续期 ========== */
const renewLoadingId = ref<number | null>(null)

async function handleRenew(id: number) {
  // 后端要求 email 必填，缺失时先提示，不再发必然失败的请求
  if (!applyForm.email) {
    message.warning(t('ssl.messages.emailRequired'))
    return
  }
  renewLoadingId.value = id
  try {
    const res = await renewSSL(id, applyForm.email)
    // Let's Encrypt 目前是 MVP 模拟：后端只写申请日志、不执行 certbot，
    // 返回 status=pending_implementation —— 如实提示，不能报「续期成功」
    if (res.data?.status === 'pending_implementation') {
      message.warning(t('ssl.messages.renewPending'))
    } else {
      message.success(t('ssl.messages.renewSuccess'))
    }
    loadSSLList()
  } catch (err: any) {
    message.error(err.message || t('ssl.messages.renewFailed'))
  } finally {
    renewLoadingId.value = null
  }
}

/* ========== 证书详情 ========== */
const showDetail = ref(false)
const detailData = ref<SSLInfo | null>(null)

function viewDetail(row: any) {
  detailData.value = row
  showDetail.value = true
}

/* ========== 申请证书对话框 ========== */
const showApply = ref(false)
const applyLoading = ref(false)
const applyStatusText = ref('')
const applyFormRef = ref()
const applyForm = reactive({
  siteId: null as number | null,
  provider: 'letsencrypt' as SSLProvider,
  email: '',
  cert: '',
  key: '',
  verification_method: 'http' as 'http' | 'dns',
  dns_account_id: null as number | null
})

const applyRules = computed(() => ({
  siteId: {
    required: true,
    message: t('ssl.rules.selectSite'),
    trigger: 'change'
  },
  email: {
    required: true,
    message: t('ssl.rules.email'),
    trigger: 'blur'
  },
  cert: {
    required: true,
    message: t('ssl.rules.cert'),
    trigger: 'blur'
  },
  key: {
    required: true,
    message: t('ssl.rules.key'),
    trigger: 'blur'
  },
  dns_account_id: {
    required: true,
    message: t('ssl.rules.dnsProvider'),
    trigger: 'change'
  }
}))

const siteOptions = computed(() =>
  sites.value.map((s) => ({ label: s.domain, value: s.id }))
)

const dnsAccountOptions = computed(() =>
  dnsAccounts.value.map((a) => ({ label: t('ssl.apply.dnsAccountLabel', { name: a.name, type: a.provider_type }), value: a.id }))
)

function openApply() {
  applyForm.siteId = null
  applyForm.provider = 'letsencrypt'
  applyForm.email = ''
  applyForm.cert = ''
  applyForm.key = ''
  applyForm.verification_method = 'http'
  applyForm.dns_account_id = null
  showApply.value = true
}

async function handleApplySubmit() {
  try {
    await applyFormRef.value?.validate()
  } catch {
    return
  }

  if (applyForm.siteId === null) return

  const data: ApplySSLData = {
    provider: applyForm.provider
  }
  if (applyForm.provider === 'letsencrypt') {
    data.email = applyForm.email
    data.verification_method = applyForm.verification_method
    if (applyForm.verification_method === 'dns') {
      data.dns_account_id = applyForm.dns_account_id ?? undefined
    }
  } else {
    data.cert = applyForm.cert
    data.key = applyForm.key
  }

  applyLoading.value = true
  if (applyForm.verification_method === 'dns' && applyForm.provider === 'letsencrypt') {
    applyStatusText.value = t('ssl.applyStatus.dnsAdding')
  }
  try {
    await applySSL(applyForm.siteId, data)
    message.success(t('ssl.messages.applySuccess'))
    showApply.value = false
    loadSSLList()
  } catch (err: any) {
    message.error(err.message || t('ssl.messages.applyFailed'))
  } finally {
    applyLoading.value = false
    applyStatusText.value = ''
  }
}

/* ========== 删除确认 ========== */
const showDelete = ref(false)
const deleteLoading = ref(false)
const deleteSSLId = ref<number | null>(null)
const deleteSSLDomain = ref('')

function openDelete(row: any) {
  deleteSSLId.value = row.id
  deleteSSLDomain.value = row.domain
  showDelete.value = true
}

async function handleDeleteConfirm() {
  if (deleteSSLId.value === null) return
  deleteLoading.value = true
  try {
    await deleteSSL(deleteSSLId.value)
    message.success(t('ssl.messages.deleteSuccess'))
    showDelete.value = false
    loadSSLList()
  } catch (err: any) {
    message.error(err.message || t('ssl.messages.deleteFailed'))
  } finally {
    deleteLoading.value = false
  }
}

onMounted(() => {
  loadSSLList()
  loadSites()
  loadDNSAccounts()
})
</script>

<template>
  <PageCard :title="t('ssl.title')" :subtitle="t('ssl.subtitle')">
    <template #extra>
      <NButton type="primary" @click="openApply">
        <template #icon>
          <NIcon><AddOutline /></NIcon>
        </template>
        {{ t('ssl.actions.apply') }}
      </NButton>
    </template>

    <DataTable
      :columns="columns"
      :data="sslList"
      :loading="loading"
      :pagination="false"
    />
  </PageCard>

  <!-- 证书详情对话框 -->
  <ContentModal
    :show="showDetail"
    :mask-closable="false"
    :title="t('ssl.detail.title')"
    size="md"
    @update:show="(v: boolean) => (showDetail = v)"
  >
    <div v-if="detailData" class="ssl-detail">
      <div class="ssl-detail__row">
        <span class="ssl-detail__label">{{ t('ssl.columns.domain') }}</span>
        <span class="ssl-detail__value mono">{{ detailData.domain }}</span>
      </div>
      <div class="ssl-detail__row">
        <span class="ssl-detail__label">{{ t('ssl.detail.providerLabel') }}</span>
        <span class="ssl-detail__value">{{ providerLabelMap[detailData.provider] }}</span>
      </div>
      <div class="ssl-detail__row" v-if="detailData.issuer">
        <span class="ssl-detail__label">{{ t('ssl.detail.issuer') }}</span>
        <span class="ssl-detail__value">{{ detailData.issuer }}</span>
      </div>
      <div class="ssl-detail__row" v-if="detailData.email">
        <span class="ssl-detail__label">{{ t('ssl.detail.email') }}</span>
        <span class="ssl-detail__value mono">{{ detailData.email }}</span>
      </div>
      <div class="ssl-detail__row">
        <span class="ssl-detail__label">{{ t('ssl.columns.expireAt') }}</span>
        <span class="ssl-detail__value mono">{{ detailData.expireAt }}</span>
      </div>
      <div class="ssl-detail__row">
        <span class="ssl-detail__label">{{ t('ssl.columns.status') }}</span>
        <StatusTag
          :status="sslStatusMap[detailData.status].status"
          :text="sslStatusMap[detailData.status].text"
        />
      </div>
    </div>
  </ContentModal>

  <!-- 申请证书对话框 -->
  <ModalForm
    :show="showApply"
    :title="t('ssl.apply.title')"
    :loading="applyLoading"
    :width="640"
    @update:show="(v: boolean) => (showApply = v)"
    @submit="handleApplySubmit"
  >
    <NForm ref="applyFormRef" :model="applyForm" :rules="applyRules" label-placement="top">
      <NFormItem :label="t('ssl.apply.selectSite')" path="siteId">
        <NSelect
          v-model:value="applyForm.siteId"
          :options="siteOptions"
          :placeholder="t('ssl.apply.sitePlaceholder')"
        />
      </NFormItem>

      <NFormItem :label="t('ssl.apply.methodLabel')" path="provider">
        <NRadioGroup v-model:value="applyForm.provider">
          <NSpace>
            <NRadio value="letsencrypt">{{ t('ssl.apply.letsencryptOption') }}</NRadio>
            <NRadio value="manual">{{ t('ssl.provider.manual') }}</NRadio>
          </NSpace>
        </NRadioGroup>
      </NFormItem>

      <template v-if="applyForm.provider === 'letsencrypt'">
        <NFormItem :label="t('ssl.apply.emailLabel')" path="email">
          <NInput
            v-model:value="applyForm.email"
            type="text"
            :placeholder="t('ssl.apply.emailPlaceholder')"
          />
        </NFormItem>
        <NFormItem :label="t('ssl.apply.verificationLabel')">
          <NRadioGroup v-model:value="applyForm.verification_method">
            <NSpace>
              <NRadio value="http">{{ t('ssl.apply.httpOption') }}</NRadio>
              <NRadio value="dns">{{ t('ssl.apply.dnsOption') }}</NRadio>
            </NSpace>
          </NRadioGroup>
        </NFormItem>

        <template v-if="applyForm.verification_method === 'http'">
          <NFormItem :label="t('ssl.apply.verificationNoteLabel')">
            <NInput :value="t('ssl.apply.httpNote')" disabled />
          </NFormItem>
        </template>

        <template v-else>
          <NFormItem :label="t('ssl.apply.dnsProviderLabel')" path="dns_account_id">
            <NSelect
              v-model:value="applyForm.dns_account_id"
              :options="dnsAccountOptions"
              :placeholder="t('ssl.apply.dnsProviderPlaceholder')"
            />
          </NFormItem>
          <NFormItem :label="t('ssl.apply.verificationNoteLabel')">
            <div class="ssl-apply__dns-hint">
              <span class="led led--warn" />
              <span>{{ t('ssl.apply.dnsNote') }}</span>
            </div>
          </NFormItem>
        </template>
      </template>

      <template v-else>
        <NFormItem :label="t('ssl.apply.certLabel')" path="cert">
          <CodeEditor
            v-model="applyForm.cert"
            language="plaintext"
            height="160px"
            :minimap="false"
            placeholder="-----BEGIN CERTIFICATE-----"
          />
        </NFormItem>
        <NFormItem :label="t('ssl.apply.keyLabel')" path="key">
          <CodeEditor
            v-model="applyForm.key"
            language="plaintext"
            height="160px"
            :minimap="false"
            placeholder="-----BEGIN PRIVATE KEY-----"
          />
        </NFormItem>
      </template>
    </NForm>
    <div v-if="applyStatusText" class="ssl-apply__status">
      <span class="led led--warn" />
      <span>{{ applyStatusText }}</span>
    </div>
  </ModalForm>

  <!-- 删除确认 -->
  <ConfirmDialog
    :show="showDelete"
    :title="t('ssl.confirm.deleteTitle')"
    :content="t('ssl.confirm.deleteContent', { domain: deleteSSLDomain })"
    :confirm-text="t('ssl.confirm.deleteConfirmText')"
    type="error"
    :loading="deleteLoading"
    @update:show="(v: boolean) => (showDelete = v)"
    @confirm="handleDeleteConfirm"
  />
</template>

<style scoped>
.ssl-detail {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.ssl-detail__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-2) 0;
  border-bottom: 1px solid var(--line);
}

.ssl-detail__row:last-child {
  border-bottom: none;
}

.ssl-detail__label {
  color: var(--text-quaternary);
  font-size: var(--fs-body);
}

.ssl-detail__value {
  color: var(--ink);
  font-weight: 500;
}

.ssl-apply__dns-hint {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--fs-caption);
  color: var(--sub);
  padding: var(--space-2) var(--space-3);
  background: var(--surface2);
  border-radius: var(--radius-sm);
}

.ssl-apply__status {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-top: var(--space-3);
  padding: var(--space-2) var(--space-3);
  background: var(--acc-soft);
  border-radius: var(--radius-sm);
  font-size: var(--fs-body);
  color: var(--acc);
}
</style>
