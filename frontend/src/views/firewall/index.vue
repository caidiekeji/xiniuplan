<script setup lang="ts">
import { ref, onMounted, computed, h } from 'vue'
import { NTabs, NTabPane, NButton, NIcon, NTag, NSwitch, NFormItem, NInput, NSelect, NText, NEmpty } from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import { ShieldHalfOutline, AddOutline, TrashOutline } from '@vicons/ionicons5'
import { PageCard, DataTable, ModalForm, StatusTag, ConfirmDialog } from '@/components'
import type { DataTableColumns } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import {
  getFirewallRules,
  addFirewallRule,
  deleteFirewallRule,
  toggleFirewallRule,
  getSecuritySuggestions,
  type FirewallRule,
  type SecuritySuggestion,
  type RuleType,
  type RuleDirection,
  type RuleProtocol,
  type RuleAction
} from '@/api/firewall'

const message = useMessage()
const { t } = useI18n()

/* ==================== 端口规则 Tab ==================== */

const rulesLoading = ref(false)
const rules = ref<FirewallRule[]>([])
const rulesTotal = ref(0)
/** 规则加载失败信息：非空时表格区渲染错误面板（含重试） */
const rulesError = ref<string | null>(null)
const rulesPage = ref(1)
const rulesPageSize = ref(20)
const ruleTypeFilter = ref<RuleType | null>(null)

const showAddModal = ref(false)
const addLoading = ref(false)
const addForm = ref({
  type: 'port' as RuleType | 'forward',
  direction: 'inbound' as RuleDirection,
  protocol: 'tcp' as RuleProtocol,
  port: '',
  ip: '',
  dest_ip: '',
  dest_port: '',
  action: 'accept' as RuleAction,
  remark: ''
})

const deleteDialog = ref({ show: false, id: 0, port: '' })
const deleteLoading = ref(false)
const toggleLoading = ref<Set<number>>(new Set())

const typeOptions = computed(() => [
  { label: t('firewall.typeForward'), value: 'forward' },
  { label: t('firewall.typePort'), value: 'port' },
  { label: 'IP', value: 'ip' }
])
const directionOptions = computed(() => [
  { label: t('firewall.directionInbound'), value: 'inbound' },
  { label: t('firewall.directionOutbound'), value: 'outbound' }
])
const protocolOptions = [
  { label: 'TCP', value: 'tcp' },
  { label: 'UDP', value: 'udp' },
  { label: 'ICMP', value: 'icmp' },
  { label: 'ALL', value: 'all' }
]
const actionOptions = computed(() => [
  { label: t('firewall.actionAccept'), value: 'accept' },
  { label: t('firewall.actionDrop'), value: 'drop' }
])

const isPortType = computed(() => addForm.value.type === 'port')
const isForwardType = computed(() => addForm.value.type === 'forward')

const ruleColumns = computed<DataTableColumns>(() => [
  {
    title: t('firewall.colType'),
    key: 'type',
    width: 80,
    render: (row: any) =>
      h(NTag, { size: 'small', bordered: false, type: row.type === 'port' ? 'info' : 'default' }, () =>
        row.type === 'port' ? t('firewall.typePort') : 'IP'
      )
  },
  {
    title: t('firewall.colDirection'),
    key: 'direction',
    width: 80,
    render: (row: any) => (row.direction === 'inbound' ? t('firewall.directionInbound') : t('firewall.directionOutbound'))
  },
  {
    title: t('firewall.colProtocol'),
    key: 'protocol',
    width: 80,
    render: (row: any) => h('span', { class: 'mono' }, row.protocol)
  },
  {
    title: t('firewall.colPort'),
    key: 'port',
    width: 100,
    render: (row: any) => h('span', { class: 'mono' }, row.port || '—')
  },
  {
    title: t('firewall.colIp'),
    key: 'ip',
    width: 140,
    render: (row: any) => h('span', { class: 'mono' }, row.ip_address || '—')
  },
  {
    title: t('firewall.colAction'),
    key: 'action',
    width: 90,
    render: (row: any) =>
      h(StatusTag, {
        status: row.action === 'accept' ? 'running' : 'error',
        text: row.action === 'accept' ? t('firewall.actionAccept') : t('firewall.actionDrop')
      })
  },
  {
    title: t('firewall.colStatus'),
    key: 'enabled',
    width: 90,
    render: (row: any) =>
      h(NSwitch, {
        value: row.status === 'enabled',
        size: 'small',
        disabled: toggleLoading.value.has(row.id),
        'onUpdate:value': () => handleToggle(row)
      })
  },
  {
    title: t('firewall.colRemark'),
    key: 'remark',
    ellipsis: { tooltip: true },
    render: (row: any) => row.remark || '—'
  },
  {
    title: t('firewall.colActions'),
    key: 'actions',
    width: 100,
    render: (row: any) =>
      h(
        NButton,
        {
          size: 'tiny',
          type: 'error',
          ghost: true,
          onClick: () => {
            deleteDialog.value = { show: true, id: row.id, port: row.port || row.ip_address || '' }
          }
        },
        () => h(NIcon, { size: 14 }, () => h(TrashOutline))
      )
  }
])

async function loadRules() {
  rulesLoading.value = true
  rulesError.value = null
  try {
    const res = await getFirewallRules(rulesPage.value, rulesPageSize.value, ruleTypeFilter.value || undefined)
    if (res.data?.items) {
      rules.value = res.data.items
      rulesTotal.value = res.data.total
    } else {
      rules.value = []
      rulesTotal.value = 0
    }
  } catch (e: any) {
    rules.value = []
    rulesTotal.value = 0
    rulesError.value = e.message || t('firewall.loadRulesFailed')
    message.error(e.message || t('firewall.loadRulesFailed'))
  } finally {
    rulesLoading.value = false
  }
}

async function handleAddRule() {
  if (isPortType.value && !addForm.value.port) {
    message.warning(t('firewall.enterPort'))
    return
  }
  if (isForwardType.value && !addForm.value.port) {
    message.warning(t('firewall.enterSourcePort'))
    return
  }
  if (isForwardType.value && !addForm.value.dest_ip) {
    message.warning(t('firewall.enterDestIp'))
    return
  }
  if (!isPortType.value && !isForwardType.value && !addForm.value.ip) {
    message.warning(t('firewall.enterIp'))
    return
  }
  addLoading.value = true
  try {
    await addFirewallRule({
      type: addForm.value.type,
      direction: addForm.value.direction,
      protocol: addForm.value.protocol,
      port: addForm.value.port ? parseInt(addForm.value.port) : undefined,
      ip_address: addForm.value.ip || undefined,
      dest_ip: addForm.value.dest_ip || undefined,
      dest_port: addForm.value.dest_port ? parseInt(addForm.value.dest_port) : undefined,
      action: addForm.value.action,
      remark: addForm.value.remark || undefined
    } as any)
    message.success(t('firewall.addSuccess'))
    showAddModal.value = false
    resetAddForm()
    loadRules()
  } catch (e: any) {
    message.error(e.message || t('firewall.addFailed'))
  } finally {
    addLoading.value = false
  }
}

function resetAddForm() {
  addForm.value = {
    type: 'port',
    direction: 'inbound',
    protocol: 'tcp',
    port: '',
    ip: '',
    dest_ip: '',
    dest_port: '',
    action: 'accept',
    remark: ''
  }
}

async function handleToggle(row: any) {
  toggleLoading.value.add(row.id)
  try {
    await toggleFirewallRule(row.id)
    row.status = row.status === 'enabled' ? 'disabled' : 'enabled'
    message.success(row.status === 'enabled' ? t('firewall.ruleEnabled') : t('firewall.ruleDisabled'))
  } catch (e: any) {
    message.error(e.message || t('firewall.opFailed'))
  } finally {
    toggleLoading.value.delete(row.id)
  }
}

async function handleDelete() {
  deleteLoading.value = true
  try {
    await deleteFirewallRule(deleteDialog.value.id)
    message.success(t('firewall.ruleDeleted'))
    deleteDialog.value.show = false
    loadRules()
  } catch (e: any) {
    message.error(e.message || t('firewall.deleteFailed'))
  } finally {
    deleteLoading.value = false
  }
}

/* ==================== 安全建议 Tab ==================== */

const suggestionsLoading = ref(false)
const suggestions = ref<SecuritySuggestion[]>([])
const closingPorts = ref<Set<number>>(new Set())

async function loadSuggestions() {
  suggestionsLoading.value = true
  try {
    const res = await getSecuritySuggestions()
    if (res.data && Array.isArray(res.data)) {
      suggestions.value = res.data
    } else {
      suggestions.value = []
    }
  } catch {
    suggestions.value = []
    message.error(t('firewall.loadSuggestionsFailed'))
  } finally {
    suggestionsLoading.value = false
  }
}

function levelTag(level: string) {
  if (level === 'high') return { status: 'error' as const, text: t('firewall.levelHigh') }
  if (level === 'medium') return { status: 'warning' as const, text: t('firewall.levelMedium') }
  return { status: 'running' as const, text: t('firewall.levelLow') }
}

async function handleOneClickClose(s: SecuritySuggestion) {
  closingPorts.value.add(s.port)
  try {
    await addFirewallRule({
      type: 'port',
      direction: 'inbound',
      protocol: 'all',
      port: s.port,
      action: 'drop',
      remark: t('firewall.autoCloseRemark', { service: s.service })
    })
    message.success(t('firewall.portClosed', { port: s.port }))
    loadRules()
  } catch (e: any) {
    message.error(e.message || t('firewall.opFailed'))
  } finally {
    closingPorts.value.delete(s.port)
  }
}

/* ==================== 生命周期 ==================== */

onMounted(() => {
  loadRules()
  loadSuggestions()
})
</script>

<template>
  <div class="firewall-page">
    <NTabs type="line" animated>
      <!-- Tab 1: 端口规则 -->
      <NTabPane name="rules" :tab="t('firewall.tabRules')">
        <PageCard :title="t('firewall.rulesTitle')" :subtitle="t('firewall.rulesSubtitle')">
          <template #extra>
            <NButton type="primary" size="small" @click="showAddModal = true">
              <template #icon>
                <NIcon><AddOutline /></NIcon>
              </template>
              {{ t('firewall.addRule') }}
            </NButton>
          </template>

          <div class="firewall__filter">
            <NSelect
              v-model:value="ruleTypeFilter"
              :placeholder="t('firewall.allTypes')"
              clearable
              size="small"
              style="width: var(--input-sm)"
              :options="typeOptions"
              @update:value="() => { rulesPage = 1; loadRules() }"
            />
          </div>

          <DataTable
            :columns="ruleColumns"
            :data="rules"
            :loading="rulesLoading"
            :row-key="(row: any) => row.id"
            :error="rulesError"
            @retry="loadRules"
            :pagination="{
              page: rulesPage,
              pageSize: rulesPageSize,
              itemCount: rulesTotal,
              'onUpdate:page': (p: number) => { rulesPage = p; loadRules() },
              'onUpdate:pageSize': (s: number) => { rulesPageSize = s; rulesPage = 1; loadRules() }
            }"
          />
        </PageCard>
      </NTabPane>

      <!-- Tab 2: 安全建议 -->
      <NTabPane name="suggestions" :tab="t('firewall.tabSuggestions')">
        <PageCard :title="t('firewall.suggestionsTitle')" :subtitle="t('firewall.suggestionsSubtitle')">
          <div v-if="suggestionsLoading" class="firewall__loading">
            <NText>{{ t('firewall.loading') }}</NText>
          </div>
          <div v-else-if="suggestions.length === 0" class="firewall__empty">
            <NEmpty :description="t('firewall.emptySuggestions')" />
          </div>
          <div v-else class="suggestion-list">
            <div v-for="s in suggestions" :key="s.id" class="suggestion-card">
              <div class="suggestion-card__header">
                <div class="suggestion-card__port">
                  <span class="mono suggestion-card__port-num">{{ s.port }}</span>
                  <span class="suggestion-card__service">{{ s.service }}</span>
                </div>
                <StatusTag :status="levelTag(s.level).status" :text="levelTag(s.level).text" />
              </div>
              <p class="suggestion-card__desc">{{ s.description }}</p>
              <div class="suggestion-card__footer">
                <span class="suggestion-card__suggestion">
                  <NIcon size="14" style="color: var(--acc2)"><ShieldHalfOutline /></NIcon>
                  {{ s.suggestion }}
                </span>
                <NButton
                  size="tiny"
                  type="error"
                  ghost
                  :loading="closingPorts.has(s.port)"
                  @click="handleOneClickClose(s)"
                >
                  {{ t('firewall.oneClickClose') }}
                </NButton>
              </div>
            </div>
          </div>
        </PageCard>
      </NTabPane>
    </NTabs>

    <!-- 添加规则 Modal -->
    <ModalForm
      :show="showAddModal"
      :title="t('firewall.addRuleTitle')"
      :loading="addLoading"
      @update:show="(v: boolean) => (showAddModal = v)"
      @submit="handleAddRule"
    >
      <NFormItem :label="t('firewall.formType')" required>
        <NSelect v-model:value="addForm.type" :options="typeOptions" />
      </NFormItem>
      <NFormItem :label="t('firewall.colDirection')" required>
        <NSelect v-model:value="addForm.direction" :options="directionOptions" />
      </NFormItem>
      <NFormItem :label="t('firewall.colProtocol')" required>
        <NSelect v-model:value="addForm.protocol" :options="protocolOptions" />
      </NFormItem>
      <NFormItem v-if="isPortType || isForwardType" :label="t('firewall.colPort')" required>
        <NInput v-model:value="addForm.port" :placeholder="t('firewall.portPlaceholder')" />
      </NFormItem>
      <NFormItem v-else-if="isForwardType" :label="t('firewall.formDestIp')" required>
        <NInput v-model:value="addForm.dest_ip" :placeholder="t('firewall.destIpPlaceholder')" />
      </NFormItem>
      <NFormItem v-if="isForwardType" :label="t('firewall.formDestPort')">
        <NInput v-model:value="addForm.dest_port" :placeholder="t('firewall.destPortPlaceholder')" />
      </NFormItem>
      <NFormItem v-else :label="t('firewall.colIp')" required>
        <NInput v-model:value="addForm.ip" :placeholder="t('firewall.ipPlaceholder')" />
      </NFormItem>
      <NFormItem :label="t('firewall.colAction')" required>
        <NSelect v-model:value="addForm.action" :options="actionOptions" />
      </NFormItem>
      <NFormItem :label="t('firewall.colRemark')">
        <NInput v-model:value="addForm.remark" :placeholder="t('firewall.remarkPlaceholder')" />
      </NFormItem>
    </ModalForm>

    <!-- 删除确认 -->
    <ConfirmDialog
      :show="deleteDialog.show"
      :title="t('firewall.deleteTitle')"
      :content="t('firewall.deleteContent', { port: deleteDialog.port })"
      :confirm-text="t('firewall.deleteConfirm')"
      type="error"
      :loading="deleteLoading"
      @update:show="(v: boolean) => (deleteDialog.show = v)"
      @confirm="handleDelete"
    />
  </div>
</template>

<style scoped>
.firewall-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.firewall__filter {
  margin-bottom: var(--space-4);
}

.firewall__loading,
.firewall__empty {
  padding: var(--space-8) 0;
  text-align: center;
  color: var(--text-quaternary);
}

.suggestion-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.suggestion-card {
  padding: var(--space-4);
  background: var(--surface2);
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  transition: border-color var(--t-fast) var(--ease);
}

.suggestion-card:hover {
  border-color: var(--line-strong);
}

.suggestion-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-2);
}

.suggestion-card__port {
  display: flex;
  align-items: baseline;
  gap: var(--space-3);
}

.suggestion-card__port-num {
  font-size: var(--fs-headline);
  font-weight: 700;
  color: var(--acc);
  line-height: 1;
}

.suggestion-card__service {
  font-size: var(--fs-body);
  font-weight: 600;
  color: var(--ink);
}

.suggestion-card__desc {
  font-size: var(--fs-caption);
  color: var(--sub);
  line-height: var(--lh-body);
  margin-bottom: var(--space-3);
}

.suggestion-card__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.suggestion-card__suggestion {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--fs-caption);
  color: var(--sub);
}
</style>
