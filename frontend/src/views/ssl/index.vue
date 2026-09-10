<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue'
import {
  NButton, NIcon, NForm, NFormItem, NInput, NSelect, NRadioGroup, NRadio,
  NSpace, useMessage
} from 'naive-ui'
import {
  AddOutline, EyeOutline, RefreshOutline, TrashOutline
} from '@vicons/ionicons5'
import {
  PageCard, DataTable, ModalForm, StatusTag, ConfirmDialog, CodeEditor
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

/* ========== 列表状态 ========== */
const loading = ref(false)
const sslList = ref<SSLInfo[]>([])
const sites = ref<Site[]>([])
const dnsAccounts = ref<DNSAccount[]>([])

const providerLabelMap: Record<SSLProvider, string> = {
  letsencrypt: "Let's Encrypt",
  manual: '手动上传'
}

const sslStatusMap: Record<SSLStatus, { status: StatusType; text: string }> = {
  valid: { status: 'running', text: '有效' },
  expiring: { status: 'warning', text: '即将到期' },
  expired: { status: 'error', text: '已过期' },
  pending: { status: 'pending', text: '签发中' }
}

function generateMockSSLList(): SSLInfo[] {
  const now = Date.now()
  return [
    {
      id: 1, siteId: 1, domain: 'example.com',
      provider: 'letsencrypt', email: 'admin@example.com',
      issuer: "Let's Encrypt Authority X3",
      expireAt: new Date(now + 60 * 86400000).toISOString().slice(0, 10),
      status: 'valid'
    },
    {
      id: 2, siteId: 2, domain: 'api.example.com',
      provider: 'manual',
      issuer: 'DigiCert Inc',
      expireAt: new Date(now + 15 * 86400000).toISOString().slice(0, 10),
      status: 'expiring'
    },
    {
      id: 3, siteId: 4, domain: 'shop.example.com',
      provider: 'letsencrypt', email: 'admin@example.com',
      issuer: "Let's Encrypt Authority X3",
      expireAt: new Date(now - 5 * 86400000).toISOString().slice(0, 10),
      status: 'expired'
    }
  ]
}

async function loadSSLList() {
  loading.value = true
  try {
    const res = await getSSLList()
    if (res.data && Array.isArray(res.data)) {
      sslList.value = res.data
    } else {
      sslList.value = generateMockSSLList()
    }
  } catch {
    sslList.value = generateMockSSLList()
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
    title: '域名',
    key: 'domain',
    width: 220,
    render: (row: any) => h('span', { class: 'mono', style: 'font-weight:500' }, row.domain)
  },
  {
    title: '证书提供方',
    key: 'provider',
    width: 160,
    render: (row: any) => h('span', { style: 'color:var(--sub)' }, providerLabelMap[row.provider as keyof typeof providerLabelMap])
  },
  {
    title: '到期时间',
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
          isExpired ? `已过期 ${Math.abs(days)} 天` : `剩余 ${days} 天`
        )
      ])
    }
  },
  {
    title: '状态',
    key: 'status',
    width: 120,
    render: (row: any) => {
      const cfg = sslStatusMap[row.status as keyof typeof sslStatusMap]
      return h(StatusTag, { status: cfg.status, text: cfg.text })
    }
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
          onClick: () => viewDetail(row)
        }, {
          icon: () => h(NIcon, null, { default: () => h(EyeOutline) }),
          default: () => '详情'
        }),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          type: 'primary',
          loading: renewLoadingId.value === row.id,
          onClick: () => handleRenew(row.id)
        }, {
          icon: () => h(NIcon, null, { default: () => h(RefreshOutline) }),
          default: () => '续期'
        }),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          type: 'error',
          onClick: () => openDelete(row)
        }, {
          icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
          default: () => '删除'
        })
      ]
    })
  }
])

/* ========== 续期 ========== */
const renewLoadingId = ref<number | null>(null)

async function handleRenew(id: number) {
  renewLoadingId.value = id
  try {
    await renewSSL(id)
    message.success('证书续期成功')
    loadSSLList()
  } catch (err: any) {
    message.error(err.message || '续期失败')
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

const applyRules = {
  siteId: {
    required: true,
    message: '请选择站点',
    trigger: 'change'
  },
  email: {
    required: true,
    message: '请输入邮箱地址',
    trigger: 'blur'
  },
  cert: {
    required: true,
    message: '请输入证书内容',
    trigger: 'blur'
  },
  key: {
    required: true,
    message: '请输入私钥内容',
    trigger: 'blur'
  },
  dns_account_id: {
    required: true,
    message: '请选择 DNS 提供商',
    trigger: 'change'
  }
}

const siteOptions = computed(() =>
  sites.value.map((s) => ({ label: s.domain, value: s.id }))
)

const dnsAccountOptions = computed(() =>
  dnsAccounts.value.map((a) => ({ label: `${a.name}（${a.provider_type}）`, value: a.id }))
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
    applyStatusText.value = '正在添加 DNS TXT 记录...'
  }
  try {
    await applySSL(applyForm.siteId, data)
    message.success('证书申请提交成功')
    showApply.value = false
    loadSSLList()
  } catch (err: any) {
    message.error(err.message || '申请失败')
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
    message.success('证书已删除')
    showDelete.value = false
    loadSSLList()
  } catch (err: any) {
    message.error(err.message || '删除失败')
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
  <PageCard title="SSL 证书" subtitle="HTTPS 证书申请、续期与管理">
    <template #extra>
      <NButton type="primary" @click="openApply">
        <template #icon>
          <NIcon><AddOutline /></NIcon>
        </template>
        申请证书
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
  <NModal
    :show="showDetail"
    :mask-closable="false"
    preset="card"
    title="证书详情"
    :style="{ width: 'var(--modal-width-md)' }"
    @update:show="(v: boolean) => (showDetail = v)"
  >
    <div v-if="detailData" class="ssl-detail">
      <div class="ssl-detail__row">
        <span class="ssl-detail__label">域名</span>
        <span class="ssl-detail__value mono">{{ detailData.domain }}</span>
      </div>
      <div class="ssl-detail__row">
        <span class="ssl-detail__label">提供方</span>
        <span class="ssl-detail__value">{{ providerLabelMap[detailData.provider] }}</span>
      </div>
      <div class="ssl-detail__row" v-if="detailData.issuer">
        <span class="ssl-detail__label">签发机构</span>
        <span class="ssl-detail__value">{{ detailData.issuer }}</span>
      </div>
      <div class="ssl-detail__row" v-if="detailData.email">
        <span class="ssl-detail__label">联系邮箱</span>
        <span class="ssl-detail__value mono">{{ detailData.email }}</span>
      </div>
      <div class="ssl-detail__row">
        <span class="ssl-detail__label">到期时间</span>
        <span class="ssl-detail__value mono">{{ detailData.expireAt }}</span>
      </div>
      <div class="ssl-detail__row">
        <span class="ssl-detail__label">状态</span>
        <StatusTag
          :status="sslStatusMap[detailData.status].status"
          :text="sslStatusMap[detailData.status].text"
        />
      </div>
    </div>
  </NModal>

  <!-- 申请证书对话框 -->
  <ModalForm
    :show="showApply"
    title="申请 SSL 证书"
    :loading="applyLoading"
    :width="640"
    @update:show="(v: boolean) => (showApply = v)"
    @submit="handleApplySubmit"
  >
    <NForm ref="applyFormRef" :model="applyForm" :rules="applyRules" label-placement="top">
      <NFormItem label="选择站点" path="siteId">
        <NSelect
          v-model:value="applyForm.siteId"
          :options="siteOptions"
          placeholder="选择要绑定证书的站点"
        />
      </NFormItem>

      <NFormItem label="申请方式" path="provider">
        <NRadioGroup v-model:value="applyForm.provider">
          <NSpace>
            <NRadio value="letsencrypt">Let's Encrypt（自动签发）</NRadio>
            <NRadio value="manual">手动上传</NRadio>
          </NSpace>
        </NRadioGroup>
      </NFormItem>

      <template v-if="applyForm.provider === 'letsencrypt'">
        <NFormItem label="邮箱地址" path="email">
          <NInput
            v-model:value="applyForm.email"
            type="text"
            placeholder="用于接收证书到期提醒"
          />
        </NFormItem>
        <NFormItem label="验证方式">
          <NRadioGroup v-model:value="applyForm.verification_method">
            <NSpace>
              <NRadio value="http">HTTP 验证</NRadio>
              <NRadio value="dns">DNS 验证</NRadio>
            </NSpace>
          </NRadioGroup>
        </NFormItem>

        <template v-if="applyForm.verification_method === 'http'">
          <NFormItem label="验证说明">
            <NInput value="将自动在站点目录放置验证文件完成 HTTP-01 验证" disabled />
          </NFormItem>
        </template>

        <template v-else>
          <NFormItem label="DNS 提供商" path="dns_account_id">
            <NSelect
              v-model:value="applyForm.dns_account_id"
              :options="dnsAccountOptions"
              placeholder="选择已配置的 DNS 提供商账号"
            />
          </NFormItem>
          <NFormItem label="验证说明">
            <div class="ssl-apply__dns-hint">
              <span class="led led--warn" />
              <span>将自动添加 TXT 记录完成 DNS-01 验证，无需手动操作</span>
            </div>
          </NFormItem>
        </template>
      </template>

      <template v-else>
        <NFormItem label="证书内容 (PEM)" path="cert">
          <CodeEditor
            v-model="applyForm.cert"
            language="plaintext"
            height="160px"
            :minimap="false"
            placeholder="-----BEGIN CERTIFICATE-----"
          />
        </NFormItem>
        <NFormItem label="私钥内容 (PEM)" path="key">
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
    title="删除证书"
    :content="`确定要删除域名「${deleteSSLDomain}」的 SSL 证书吗？删除后该站点将无法使用 HTTPS。`"
    confirm-text="删除"
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
  color: var(--faint);
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
