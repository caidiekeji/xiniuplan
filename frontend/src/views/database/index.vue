<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { NTabs, NTabPane, NButton, NSpace, NIcon, NEmpty } from 'naive-ui'
import { ChevronBackOutline } from '@vicons/ionicons5'
import { PageCard, StatusTag } from '@/components'
import * as dbApi from '@/api/database'
import type { DBInstance, DatabaseInfo } from '@/api/database'
import InstanceListTab from './InstanceListTab.vue'
import DatabaseManagerTab from './DatabaseManagerTab.vue'
import UserManagerTab from './UserManagerTab.vue'
import SqlExecutorTab from './SqlExecutorTab.vue'
import RedisPanel from './RedisPanel.vue'
import PgPanel from './PgPanel.vue'

const { t } = useI18n()

/* ==================== 共享状态 ==================== */

const activeTab = ref<string>('instances')

// 实例列表（所有 tab 共用 selector）
const instances = ref<DBInstance[]>([])
const instancesLoading = ref(false)
const instancePage = ref(1)
const instancePageSize = ref(20)
const instanceTotal = ref(0)
let pollTimer: ReturnType<typeof setInterval> | null = null

// 当前选中实例（schema/users/sql tab 共用）
const selectedInstanceId = ref<number | null>(null)
const selectedInstanceName = ref('')

// 数据库列表（schema/sql/users tab 共用）
const databases = ref<DatabaseInfo[]>([])
const databasesLoading = ref(false)

/* ==================== 实例加载 / 轮询 ==================== */

async function loadInstances() {
  instancesLoading.value = true
  try {
    const res = await dbApi.getInstances({
      page: instancePage.value,
      pageSize: instancePageSize.value
    })
    instances.value = res.data.items
    instanceTotal.value = res.data.total
  } catch {
    // 静默失败，轮询时不弹窗
  } finally {
    instancesLoading.value = false
  }
}

function startPolling() {
  stopPolling()
  pollTimer = setInterval(() => {
    dbApi
      .getInstances({ page: instancePage.value, pageSize: instancePageSize.value })
      .then((res) => {
        instances.value = res.data.items
        instanceTotal.value = res.data.total
      })
      .catch(() => {})
  }, 10000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

/* ==================== 数据库加载（共享） ==================== */

async function loadDatabases() {
  if (!selectedInstanceId.value) return
  databasesLoading.value = true
  try {
    const res = await dbApi.getDatabases(selectedInstanceId.value)
    databases.value = res.data
  } catch {
    // 静默
  } finally {
    databasesLoading.value = false
  }
}

/* ==================== Computed ==================== */

const mysqlInstanceOptions = computed(() =>
  instances.value
    .filter((ins) => ins.type === 'mysql')
    .map((ins) => ({
      label: `${ins.name} (${ins.version}, :${ins.port})`,
      value: ins.id
    }))
)

const selectedInstance = computed(
  () => instances.value.find((i) => i.id === selectedInstanceId.value) || null
)

const showMysqlTabs = computed(
  () => !selectedInstance.value || selectedInstance.value.type === 'mysql'
)
const showPgTab = computed(() => selectedInstance.value?.type === 'postgres')
const showRedisTab = computed(() => selectedInstance.value?.type === 'redis')

/* ==================== 事件处理 ==================== */

const userManagerTabRef = ref<InstanceType<typeof UserManagerTab> | null>(null)

function handleManage(instance: DBInstance) {
  selectedInstanceId.value = instance.id
  selectedInstanceName.value = instance.name || `${instance.type}-${instance.version}`
  if (instance.type === 'postgres') {
    activeTab.value = 'pg'
  } else if (instance.type === 'redis') {
    activeTab.value = 'redis'
  } else {
    activeTab.value = 'schema'
    loadDatabases()
  }
}

async function handleCreateUser() {
  activeTab.value = 'users'
  // 等待 tab 面板挂载后再调用子组件方法（原 setTimeout hack 的规范写法）
  await nextTick()
  userManagerTabRef.value?.openCreateUser()
}

function onInstanceIdChange(val: number | null) {
  selectedInstanceId.value = val
  const ins = instances.value.find((i) => i.id === val)
  if (ins) selectedInstanceName.value = ins.name || `${ins.type}-${ins.version}`
  databases.value = []
  if (val) loadDatabases()
}

/* ==================== 生命周期 ==================== */

onMounted(() => {
  loadInstances()
  startPolling()
})

onBeforeUnmount(() => {
  stopPolling()
})

watch(activeTab, (tab) => {
  if (tab === 'schema' && selectedInstanceId.value && databases.value.length === 0) {
    loadDatabases()
  } else if (tab === 'users' && selectedInstanceId.value) {
    // UserManagerTab 自己加载用户列表
  } else if (tab === 'sql' && selectedInstanceId.value && databases.value.length === 0) {
    loadDatabases()
  }
})
</script>

<template>
  <PageCard :title="t('database.title')" :subtitle="t('database.subtitle')">
    <NTabs v-model:value="activeTab" type="line" animated>
      <!-- Tab 1: 实例列表 -->
      <NTabPane name="instances" :tab="t('database.tabs.instances')">
        <InstanceListTab
          :instances="instances"
          :loading="instancesLoading"
          :page="instancePage"
          :page-size="instancePageSize"
          :total="instanceTotal"
          @refresh="loadInstances"
          @manage="handleManage"
          @update:page="(v: number) => { instancePage = v; loadInstances() }"
          @update:page-size="(v: number) => { instancePageSize = v; instancePage = 1; loadInstances() }"
        />
      </NTabPane>

      <!-- Tab 2: 库表管理 -->
      <NTabPane v-if="showMysqlTabs" name="schema" :tab="t('database.tabs.schema')">
        <DatabaseManagerTab
          :instance-id="selectedInstanceId"
          :instance-options="mysqlInstanceOptions"
          :databases="databases"
          :databases-loading="databasesLoading"
          @update:instance-id="onInstanceIdChange"
          @load-databases="loadDatabases"
          @create-user="handleCreateUser"
        />
      </NTabPane>

      <!-- Tab 3: 用户权限 -->
      <NTabPane v-if="showMysqlTabs" name="users" :tab="t('database.tabs.users')">
        <UserManagerTab
          ref="userManagerTabRef"
          :instance-id="selectedInstanceId"
          :instance-options="mysqlInstanceOptions"
          :databases="databases"
          @update:instance-id="onInstanceIdChange"
        />
      </NTabPane>

      <!-- Tab 4: SQL 执行器 -->
      <NTabPane v-if="showMysqlTabs" name="sql" :tab="t('database.tabs.sql')">
        <SqlExecutorTab
          :instance-id="selectedInstanceId"
          :instance-options="mysqlInstanceOptions"
          :databases="databases"
          @update:instance-id="onInstanceIdChange"
        />
      </NTabPane>

      <!-- Tab 5: PostgreSQL 管理 -->
      <NTabPane v-if="showPgTab" name="pg" :tab="t('database.tabs.pg')">
        <div v-if="selectedInstanceId" class="mgmt-panel">
          <div class="mgmt-panel__header">
            <NSpace>
              <span class="mgmt-panel__title">{{ selectedInstanceName }}</span>
              <StatusTag status="running" text="PostgreSQL" />
            </NSpace>
            <NButton size="small" @click="activeTab = 'instances'">
              <template #icon><NIcon><ChevronBackOutline /></NIcon></template>{{ t('database.backToInstances') }}
            </NButton>
          </div>
          <PgPanel :instance-id="selectedInstanceId" :key="selectedInstanceId" />
        </div>
        <div v-else class="empty-hint">
          <NEmpty :description="t('database.selectPgInstance')" />
        </div>
      </NTabPane>

      <!-- Tab 6: Redis 管理 -->
      <NTabPane v-if="showRedisTab" name="redis" :tab="t('database.tabs.redis')">
        <div v-if="selectedInstanceId" class="mgmt-panel">
          <div class="mgmt-panel__header">
            <NSpace>
              <span class="mgmt-panel__title">{{ selectedInstanceName }}</span>
              <StatusTag status="error" text="Redis" />
            </NSpace>
            <NButton size="small" @click="activeTab = 'instances'">
              <template #icon><NIcon><ChevronBackOutline /></NIcon></template>{{ t('database.backToInstances') }}
            </NButton>
          </div>
          <RedisPanel :instance-id="selectedInstanceId" :key="selectedInstanceId" />
        </div>
        <div v-else class="empty-hint">
          <NEmpty :description="t('database.selectRedisInstance')" />
        </div>
      </NTabPane>
    </NTabs>
  </PageCard>
</template>

<style scoped>
.empty-hint {
  padding: var(--space-8) 0;
  display: flex;
  justify-content: center;
}

/* NTabs 样式微调 */
:deep(.n-tabs-nav) {
  margin-bottom: var(--space-4);
}

:deep(.n-tab) {
  font-size: var(--fs-body);
}

/* 管理面板头部 */
.mgmt-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-4);
  padding-bottom: var(--space-3);
  border-bottom: 1px solid var(--line);
}

.mgmt-panel__title {
  font-size: var(--fs-title);
  font-weight: 600;
  color: var(--ink);
}
</style>
