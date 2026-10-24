<template>
  <div class="page">
    <NTabs v-model:value="activeTab" type="line" animated class="session-tabs">
      <NTabPane name="sessions" :tab="t('sessions.title')">
      <PageCard :title="t('sessions.title')">
      <template #extra>
        <NButton size="small" quaternary @click="loadSessions">
          <template #icon><NIcon><RefreshOutline /></NIcon></template>
          {{ t('sessions.refresh') }}
        </NButton>
      </template>
      <DataTable :columns="sessionColumns" :data="sessions" :loading="sessionLoading" :row-key="(r: any) => r.token_hash" />
      </PageCard>
      </NTabPane>

      <NTabPane name="keys" :tab="t('sessions.apiKeyTab')">
      <PageCard title="API Key">
      <template #extra>
        <NSpace>
          <NButton size="small" type="primary" @click="showCreateKey = true">{{ t('sessions.createKey') }}</NButton>
          <NButton size="small" quaternary @click="loadKeys">
            <template #icon><NIcon><RefreshOutline /></NIcon></template>
            {{ t('sessions.refresh') }}
          </NButton>
        </NSpace>
      </template>
      <DataTable :columns="keyColumns" :data="apiKeys" :loading="keyLoading" :row-key="(r: any) => r.id" />
      </PageCard>
      </NTabPane>
    </NTabs>

    <ModalForm
      v-model:show="showCreateKey"
      :title="t('sessions.createKey')"
      :width="480"
      :loading="creating"
      :confirm-text="t('sessions.create')"
      :cancel-text="t('sessions.cancel')"
      @submit="doCreateKey"
    >
      <NForm label-placement="left" label-width="110">
        <NFormItem :label="t('sessions.form.name')">
          <NInput v-model:value="newKey.name" :placeholder="t('sessions.form.namePlaceholder')" />
        </NFormItem>
        <NFormItem :label="t('sessions.form.expiresInDays')">
          <NInputNumber v-model:value="newKey.expiresInDays" :min="0" style="width: 100%" />
          <span style="margin-left: var(--space-2); color: var(--text-quaternary)">{{ t('sessions.form.neverExpiresHint') }}</span>
        </NFormItem>
      </NForm>
    </ModalForm>

    <ContentModal v-model:show="showSecret" :title="t('sessions.secret.title')" size="md">
      <NAlert type="warning" style="margin-bottom: var(--space-3)">
        {{ t('sessions.secret.warning') }}
        <code>X-API-Key: keyID.secret</code>
      </NAlert>
      <NInput :value="createdSecret" readonly type="textarea" :rows="2" />
      <template #footer>
        <NSpace justify="end">
          <NButton type="primary" @click="showSecret = false">{{ t('sessions.secret.saved') }}</NButton>
        </NSpace>
      </template>
    </ContentModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { NButton, NIcon, NTag, NSpace, NForm, NFormItem, NInput, NInputNumber, NAlert, NTabs, NTabPane, useDialog } from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import { RefreshOutline, CloseCircleOutline, CopyOutline } from '@vicons/ionicons5'
import { PageCard, DataTable, ModalForm, ContentModal } from '@/components'
import {
  listSessions, revokeSession, listAPIKeys, createAPIKey, deleteAPIKey,
  type OnlineSession, type APIKey
} from '@/api/newfeatures'
import { useI18n } from 'vue-i18n'
import { formatDate } from '@/utils/format'

/* 两个功能区以标签页组织 */
const activeTab = ref('sessions')

const { t } = useI18n()
const message = useMessage()
const dialog = useDialog()

/* ---------- 会话 ---------- */
const sessions = ref<OnlineSession[]>([])
const sessionLoading = ref(false)

async function loadSessions() {
  sessionLoading.value = true
  try {
    sessions.value = await listSessions()
  } catch (e: any) {
    message.error(e.message || t('sessions.messages.loadSessionsFailed'))
  } finally {
    sessionLoading.value = false
  }
}

async function doRevoke(hash: string) {
  try {
    await revokeSession(hash)
    message.success(t('sessions.messages.revoked'))
    loadSessions()
  } catch (e: any) {
    message.error(e.message || t('sessions.messages.revokeFailed'))
  }
}

const sessionColumns = computed(() => [
  { title: t('sessions.columns.user'), key: 'username', width: 100 },
  { title: 'IP', key: 'ip', width: 140 },
  {
    title: t('sessions.columns.type'), key: 'token_type', width: 80,
    render: (row: any) => h(NTag, { size: 'small', type: row.token_type === 'access' ? 'info' : 'default' }, { default: () => row.token_type })
  },
  { title: 'User-Agent', key: 'user_agent', ellipsis: { tooltip: true } },
  { title: t('sessions.columns.loginAt'), key: 'login_at', width: 170, render: (r: any) => fmtTime(r.login_at) },
  { title: t('sessions.columns.lastActive'), key: 'last_active', width: 170, render: (r: any) => fmtTime(r.last_active) },
  {
    title: t('sessions.columns.actions'), key: 'actions', width: 90,
    render: (row: any) => h(
      NButton, {
        size: 'tiny', type: 'error', quaternary: true,
        onClick: () => doRevoke(row.token_hash)
      }, {
        icon: () => h(NIcon, null, { default: () => h(CloseCircleOutline) }),
        default: () => t('sessions.actions.revoke')
      }
    )
  }
])

/* ---------- API Key ---------- */
const apiKeys = ref<APIKey[]>([])
const keyLoading = ref(false)
const showCreateKey = ref(false)
const creating = ref(false)
const showSecret = ref(false)
const createdSecret = ref('')
const newKey = ref({ name: '', expiresInDays: 0 })

async function loadKeys() {
  keyLoading.value = true
  try {
    apiKeys.value = await listAPIKeys()
  } catch (e: any) {
    message.error(e.message || t('sessions.messages.loadKeysFailed'))
  } finally {
    keyLoading.value = false
  }
}

async function doCreateKey() {
  if (!newKey.value.name.trim()) {
    message.warning(t('sessions.messages.nameRequired'))
    return
  }
  creating.value = true
  try {
    const key = await createAPIKey({
      name: newKey.value.name.trim(),
      expires_in_days: newKey.value.expiresInDays
    })
    if (key.secret) {
      createdSecret.value = key.secret
      showSecret.value = true
    }
    showCreateKey.value = false
    newKey.value = { name: '', expiresInDays: 0 }
    loadKeys()
  } catch (e: any) {
    message.error(e.message || t('sessions.messages.createFailed'))
  } finally {
    creating.value = false
  }
}

function confirmDeleteKey(row: APIKey) {
  dialog.warning({
    title: t('sessions.deleteDialog.title'),
    content: t('sessions.deleteDialog.content', { name: row.name }),
    positiveText: t('sessions.actions.delete'),
    negativeText: t('sessions.cancel'),
    onPositiveClick: async () => {
      await deleteAPIKey(row.id)
      message.success(t('sessions.messages.deleted'))
      loadKeys()
    }
  })
}

const keyColumns = computed(() => [
  { title: t('sessions.columns.name'), key: 'name', width: 140 },
  { title: 'KeyID', key: 'key_id', width: 150, ellipsis: { tooltip: true } },
  {
    title: t('sessions.columns.status'), key: 'status', width: 90,
    render: (row: any) => h(NTag, { size: 'small', type: row.status === 'active' ? 'success' : 'error' }, { default: () => (row.status === 'active' ? t('sessions.status.active') : t('sessions.status.inactive')) })
  },
  { title: t('sessions.columns.lastUsed'), key: 'last_used_at', width: 170, render: (r: any) => (r.last_used_at ? fmtTime(r.last_used_at) : t('sessions.status.neverUsed')) },
  { title: t('sessions.columns.expiresAt'), key: 'expires_at', width: 170, render: (r: any) => (r.expires_at ? fmtTime(r.expires_at) : t('sessions.status.neverExpires')) },
  {
    title: t('sessions.columns.actions'), key: 'actions', width: 100,
    render: (row: any) => h(
      NButton, {
        size: 'tiny', type: 'error', quaternary: true,
        onClick: () => confirmDeleteKey(row)
      }, {
        icon: () => h(NIcon, null, { default: () => h(CopyOutline) }),
        default: () => t('sessions.actions.delete')
      }
    )
  }
])

function fmtTime(v: string) {
  if (!v) return '-'
  return formatDate(v)
}

onMounted(() => {
  loadSessions()
  loadKeys()
})
</script>

<style scoped>
.page {
  /* tabs 导航与卡片间的间距 */
}

.session-tabs :deep(.n-tabs-nav) {
  margin-bottom: var(--space-4);
}
</style>
