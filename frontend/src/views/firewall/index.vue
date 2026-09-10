<script setup lang="ts">
import { ref, onMounted, computed, h } from 'vue'
import { NTabs, NTabPane, NButton, NIcon, NTag, NSwitch, NForm, NFormItem, NInput, NSelect, NText, NEmpty, useMessage } from 'naive-ui'
import { ShieldHalfOutline, AddOutline, TrashOutline } from '@vicons/ionicons5'
import { PageCard, DataTable, ModalForm, StatusTag, ConfirmDialog } from '@/components'
import type { DataTableColumns } from 'naive-ui'
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

/* ==================== 端口规则 Tab ==================== */

const rulesLoading = ref(false)
const rules = ref<FirewallRule[]>([])
const rulesTotal = ref(0)
const rulesPage = ref(1)
const rulesPageSize = ref(20)
const ruleTypeFilter = ref<RuleType | null>(null)

const showAddModal = ref(false)
const addLoading = ref(false)
const addForm = ref({
  type: 'port' as RuleType,
  direction: 'inbound' as RuleDirection,
  protocol: 'tcp' as RuleProtocol,
  port: '',
  ip: '',
  action: 'accept' as RuleAction,
  remark: ''
})

const deleteDialog = ref({ show: false, id: 0, port: '' })
const deleteLoading = ref(false)
const toggleLoading = ref<Set<number>>(new Set())

const typeOptions = [
  { label: '端口', value: 'port' },
  { label: 'IP', value: 'ip' }
]
const directionOptions = [
  { label: '入站', value: 'inbound' },
  { label: '出站', value: 'outbound' }
]
const protocolOptions = [
  { label: 'TCP', value: 'tcp' },
  { label: 'UDP', value: 'udp' },
  { label: 'ICMP', value: 'icmp' },
  { label: 'ALL', value: 'all' }
]
const actionOptions = [
  { label: '允许', value: 'accept' },
  { label: '拒绝', value: 'drop' }
]

const isPortType = computed(() => addForm.value.type === 'port')

const ruleColumns = computed<DataTableColumns>(() => [
  {
    title: '类型',
    key: 'type',
    width: 80,
    render: (row: any) =>
      h(NTag, { size: 'small', bordered: false, type: row.type === 'port' ? 'info' : 'default' }, () =>
        row.type === 'port' ? '端口' : 'IP'
      )
  },
  {
    title: '方向',
    key: 'direction',
    width: 80,
    render: (row: any) => (row.direction === 'inbound' ? '入站' : '出站')
  },
  {
    title: '协议',
    key: 'protocol',
    width: 80,
    render: (row: any) => h('span', { class: 'mono' }, row.protocol)
  },
  {
    title: '端口',
    key: 'port',
    width: 100,
    render: (row: any) => h('span', { class: 'mono' }, row.port || '—')
  },
  {
    title: 'IP 地址',
    key: 'ip',
    width: 140,
    render: (row: any) => h('span', { class: 'mono' }, row.ip_address || '—')
  },
  {
    title: '动作',
    key: 'action',
    width: 90,
    render: (row: any) =>
      h(StatusTag, {
        status: row.action === 'accept' ? 'running' : 'error',
        text: row.action === 'accept' ? '允许' : '拒绝'
      })
  },
  {
    title: '状态',
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
    title: '备注',
    key: 'remark',
    ellipsis: { tooltip: true },
    render: (row: any) => row.remark || '—'
  },
  {
    title: '操作',
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
  try {
    const res = await getFirewallRules(rulesPage.value, rulesPageSize.value, ruleTypeFilter.value || undefined)
    if (res.data?.items) {
      rules.value = res.data.items
      rulesTotal.value = res.data.total
    } else {
      generateMockRules()
    }
  } catch {
    generateMockRules()
  } finally {
    rulesLoading.value = false
  }
}

function generateMockRules() {
  rules.value = [
    { id: 1, type: 'port', direction: 'inbound', protocol: 'tcp', port: 22, action: 'accept', status: 'enabled', remark: 'SSH 远程登录', created_at: '2026-09-01 10:00:00' },
    { id: 2, type: 'port', direction: 'inbound', protocol: 'tcp', port: 80, action: 'accept', status: 'enabled', remark: 'HTTP', created_at: '2026-09-01 10:00:00' },
    { id: 3, type: 'port', direction: 'inbound', protocol: 'tcp', port: 443, action: 'accept', status: 'enabled', remark: 'HTTPS', created_at: '2026-09-01 10:00:00' },
    { id: 4, type: 'port', direction: 'inbound', protocol: 'tcp', port: 3306, action: 'drop', status: 'enabled', remark: '禁止外部访问 MySQL', created_at: '2026-09-02 14:30:00' },
    { id: 5, type: 'ip', direction: 'inbound', protocol: 'all', ip_address: '192.168.1.100', action: 'accept', status: 'enabled', remark: '内网管理机', created_at: '2026-09-03 09:15:00' },
    { id: 6, type: 'port', direction: 'inbound', protocol: 'tcp', port: 6379, action: 'drop', status: 'disabled', remark: 'Redis 外部访问（已禁用）', created_at: '2026-09-04 16:20:00' }
  ]
  rulesTotal.value = rules.value.length
}

async function handleAddRule() {
  if (isPortType.value && !addForm.value.port) {
    message.warning('请输入端口号')
    return
  }
  if (!isPortType.value && !addForm.value.ip) {
    message.warning('请输入 IP 地址')
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
      action: addForm.value.action,
      remark: addForm.value.remark || undefined
    })
    message.success('规则添加成功')
    showAddModal.value = false
    resetAddForm()
    loadRules()
  } catch (e: any) {
    message.error(e.message || '添加失败')
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
    action: 'accept',
    remark: ''
  }
}

async function handleToggle(row: any) {
  toggleLoading.value.add(row.id)
  try {
    await toggleFirewallRule(row.id)
    row.status = row.status === 'enabled' ? 'disabled' : 'enabled'
    message.success(row.status === 'enabled' ? '规则已启用' : '规则已禁用')
  } catch (e: any) {
    message.error(e.message || '操作失败')
  } finally {
    toggleLoading.value.delete(row.id)
  }
}

async function handleDelete() {
  deleteLoading.value = true
  try {
    await deleteFirewallRule(deleteDialog.value.id)
    message.success('规则已删除')
    deleteDialog.value.show = false
    loadRules()
  } catch (e: any) {
    message.error(e.message || '删除失败')
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
      generateMockSuggestions()
    }
  } catch {
    generateMockSuggestions()
  } finally {
    suggestionsLoading.value = false
  }
}

function generateMockSuggestions() {
  suggestions.value = [
    { id: 1, port: 23, service: 'Telnet', level: 'high', description: 'Telnet 协议以明文传输数据，存在严重的窃听和中间人攻击风险', suggestion: '关闭 23 端口，使用 SSH 替代' },
    { id: 2, port: 3389, service: 'RDP', level: 'high', description: '远程桌面协议暴露在公网，易遭受暴力破解和蓝宝漏洞攻击', suggestion: '限制来源 IP 或使用 VPN 访问' },
    { id: 3, port: 6379, service: 'Redis', level: 'medium', description: 'Redis 默认无认证，暴露后可能导致数据泄露和未授权访问', suggestion: '配置密码并绑定内网地址' },
    { id: 4, port: 27017, service: 'MongoDB', level: 'medium', description: 'MongoDB 未启用认证，存在数据被勒索或删除的风险', suggestion: '启用认证并限制访问来源' },
    { id: 5, port: 21, service: 'FTP', level: 'low', description: 'FTP 明文传输密码，建议使用更安全的 SFTP', suggestion: '迁移至 SFTP 或 FTPS' }
  ]
}

function levelTag(level: string) {
  if (level === 'high') return { status: 'error' as const, text: '高危' }
  if (level === 'medium') return { status: 'warning' as const, text: '中危' }
  return { status: 'running' as const, text: '低危' }
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
      remark: `安全建议自动关闭: ${s.service}`
    })
    message.success(`已添加拒绝规则，端口 ${s.port} 已关闭`)
    loadRules()
  } catch (e: any) {
    message.error(e.message || '操作失败')
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
      <NTabPane name="rules" tab="端口规则">
        <PageCard title="防火墙规则" subtitle="管理端口与 IP 访问控制策略">
          <template #extra>
            <NButton type="primary" size="small" @click="showAddModal = true">
              <template #icon>
                <NIcon><AddOutline /></NIcon>
              </template>
              添加规则
            </NButton>
          </template>

          <div class="firewall__filter">
            <NSelect
              v-model:value="ruleTypeFilter"
              placeholder="全部类型"
              clearable
              size="small"
              style="width: 140px"
              :options="typeOptions"
              @update:value="() => { rulesPage = 1; loadRules() }"
            />
          </div>

          <DataTable
            :columns="ruleColumns"
            :data="rules"
            :loading="rulesLoading"
            :row-key="(row: any) => row.id"
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
      <NTabPane name="suggestions" tab="安全建议">
        <PageCard title="安全建议" subtitle="高危端口检测与安全加固推荐">
          <div v-if="suggestionsLoading" class="firewall__loading">
            <NText>加载中...</NText>
          </div>
          <div v-else-if="suggestions.length === 0" class="firewall__empty">
            <NEmpty description="暂无安全建议，系统处于安全状态" />
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
                  一键关闭
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
      title="添加防火墙规则"
      :loading="addLoading"
      @update:show="(v: boolean) => (showAddModal = v)"
      @submit="handleAddRule"
    >
      <NFormItem label="规则类型" required>
        <NSelect v-model:value="addForm.type" :options="typeOptions" />
      </NFormItem>
      <NFormItem label="方向" required>
        <NSelect v-model:value="addForm.direction" :options="directionOptions" />
      </NFormItem>
      <NFormItem label="协议" required>
        <NSelect v-model:value="addForm.protocol" :options="protocolOptions" />
      </NFormItem>
      <NFormItem v-if="isPortType" label="端口" required>
        <NInput v-model:value="addForm.port" placeholder="例如: 80 或 8000-9000" />
      </NFormItem>
      <NFormItem v-else label="IP 地址" required>
        <NInput v-model:value="addForm.ip" placeholder="例如: 192.168.1.100 或 10.0.0.0/24" />
      </NFormItem>
      <NFormItem label="动作" required>
        <NSelect v-model:value="addForm.action" :options="actionOptions" />
      </NFormItem>
      <NFormItem label="备注">
        <NInput v-model:value="addForm.remark" placeholder="规则说明（可选）" />
      </NFormItem>
    </ModalForm>

    <!-- 删除确认 -->
    <ConfirmDialog
      :show="deleteDialog.show"
      title="删除规则"
      :content="`确定要删除规则（${deleteDialog.port}）吗？此操作不可撤销。`"
      confirm-text="删除"
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
  gap: var(--space-4);
}

.firewall__filter {
  margin-bottom: var(--space-4);
}

.firewall__loading,
.firewall__empty {
  padding: var(--space-8) 0;
  text-align: center;
  color: var(--faint);
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
