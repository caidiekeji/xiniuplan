<script setup lang="ts">
import { ref, computed, onMounted, watch, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from '@/composables/useMessage'
import {
  NButton, NSpace, NIcon, NTabs, NTabPane, NInput, NSelect,
  NCheckboxGroup, NCheckbox, NFormItem, NEmpty, NSpin, NAlert, NTag
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { AddOutline, TrashOutline, PlayOutline, RefreshOutline, ConstructOutline, FileTrayFullOutline, ChevronForwardOutline, CloseOutline, ServerOutline, ShieldHalfOutline } from '@vicons/ionicons5'
import {
  DataTable, ModalForm, ConfirmDialog, CodeEditor
} from '@/components'
import * as pgApi from '@/api/postgres'
import type {
  PgDatabase, PgTable, PgTableSchema, PgUser,
  CreatePgDatabaseData, CreatePgUserData, PgGrantData, PgSQLResult
} from '@/api/postgres'

const props = defineProps<{
  instanceId: number
}>()

const { t } = useI18n()
const message = useMessage()

/* ==================== 辅助函数 ==================== */

function monoEl(text: any) {
  return h('span', { class: 'mono' }, String(text ?? '-'))
}

function iconBtn(opts: { text: string; icon: any; onClick: () => void; type?: string }) {
  return h(NButton, {
    size: 'tiny',
    quaternary: true,
    type: (opts.type || 'default') as any,
    onClick: opts.onClick
  }, {
    icon: () => h(NIcon, null, { default: () => h(opts.icon) }),
    default: () => opts.text
  })
}

function btnGroup(children: any[]) {
  return h(NSpace, { size: 4 }, { default: () => children })
}

function rowsToObjects(columns: string[], rows: any[][]): Record<string, any>[] {
  return rows.map(row => {
    const obj: Record<string, any> = {}
    columns.forEach((col, i) => { obj[col] = row[i] })
    return obj
  })
}

function buildDynamicColumns(columns: string[] | null | undefined): DataTableColumns {
  if (!Array.isArray(columns)) return []
  return columns.map(col => ({ title: col, key: col }))
}

/* ==================== 子 Tab ==================== */

const activeSubTab = ref('schema')

/* ==================== Tab 1: 库表管理 ==================== */

const databases = ref<PgDatabase[]>([])
const databasesLoading = ref(false)
const selectedDbName = ref('')

async function loadDatabases() {
  databasesLoading.value = true
  try {
    const res = await pgApi.getDatabases(props.instanceId)
    databases.value = Array.isArray(res.data) ? res.data : []
    if (databases.value.length > 0 && !databases.value.find(d => d.name === selectedDbName.value)) {
      selectedDbName.value = databases.value[0].name
    }
  } catch (e: any) {
    message.error(e.message || t('pgPanel.loadDatabasesFailed'))
  } finally {
    databasesLoading.value = false
  }
}

watch(selectedDbName, () => {
  loadTables()
  expandedTableName.value = ''
})

/* --- 表列表 --- */

const tables = ref<PgTable[]>([])
const tablesLoading = ref(false)
const expandedTableName = ref('')
const tableSchema = ref<PgTableSchema | null>(null)
const tableSchemaLoading = ref(false)

async function loadTables() {
  if (!selectedDbName.value) return
  tablesLoading.value = true
  try {
    const res = await pgApi.getTables(props.instanceId, selectedDbName.value)
    tables.value = res.data
  } catch (e: any) {
    message.error(e.message || t('pgPanel.loadTablesFailed'))
  } finally {
    tablesLoading.value = false
  }
}

async function toggleTableSchema(tableName: string) {
  if (expandedTableName.value === tableName) {
    expandedTableName.value = ''
    return
  }
  expandedTableName.value = tableName
  tableSchemaLoading.value = true
  try {
    const res = await pgApi.getTableSchema(props.instanceId, selectedDbName.value, tableName)
    tableSchema.value = res.data
  } catch (e: any) {
    message.error(e.message || t('pgPanel.loadSchemaFailed'))
  } finally {
    tableSchemaLoading.value = false
  }
}

const pgTableColumns = computed<DataTableColumns>(() => [
  {
    title: t('pgPanel.columns.tableName'), key: 'name', width: 220,
    render: (row: any) => h(NButton, {
      text: true, size: 'small', onClick: () => toggleTableSchema(row.name)
    }, {
      default: () => h(NSpace, { size: 4 }, {
        default: () => [
          h(NIcon, { size: 14 }, { default: () => h(ChevronForwardOutline) }),
          monoEl(row.name)
        ]
      })
    })
  },
  {
    title: t('pgPanel.columns.rowsEstimate'), key: 'rows_estimate', width: 120,
    render: (row: any) => monoEl(row.rows_estimate?.toLocaleString() ?? '-')
  },
  { title: t('pgPanel.columns.size'), key: 'size', width: 120 },
  {
    title: t('pgPanel.columns.actions'), key: 'actions', width: 160,
    render: (row: any) => btnGroup([
      iconBtn({ text: t('pgPanel.columns.structure'), icon: ConstructOutline, onClick: () => toggleTableSchema(row.name) }),
      iconBtn({ text: t('pgPanel.columns.data'), icon: ServerOutline, type: 'primary', onClick: () => viewTableData(row.name) })
    ])
  }
])

const pgStructureColumns = computed<DataTableColumns>(() => [
  { title: t('pgPanel.columns.columnName'), key: 'name', width: 180, render: (row: any) => monoEl(row.name) },
  { title: t('pgPanel.columns.type'), key: 'type', width: 160, render: (row: any) => monoEl(row.type) },
  {
    title: t('pgPanel.columns.nullable'), key: 'nullable', width: 80,
    render: (row: any) => row.nullable ? 'YES' : 'NO'
  },
  { title: t('pgPanel.columns.defaultValue'), key: 'default', width: 160, render: (row: any) => row.default ?? '-' },
  {
    title: t('pgPanel.columns.primaryKey'), key: 'is_primary', width: 80,
    render: (row: any) => row.is_primary
      ? h(NTag, { size: 'small', type: 'warning', bordered: false }, { default: () => 'PK' })
      : '-'
  }
])

/* --- 表数据 --- */

const tableDataVisible = ref(false)
const tableDataTableName = ref('')
const tableDataColumns = ref<string[]>([])
const tableDataRows = ref<Record<string, any>[]>([])
const tableDataLoading = ref(false)
const tableDataPage = ref(1)
const tableDataPageSize = ref(20)
const tableDataTotal = ref(0)

async function viewTableData(tableName: string) {
  tableDataVisible.value = true
  tableDataTableName.value = tableName
  tableDataPage.value = 1
  await loadTableData()
}

async function loadTableData() {
  if (!selectedDbName.value || !tableDataTableName.value) return
  tableDataLoading.value = true
  try {
    const res = await pgApi.getTableData(
      props.instanceId, selectedDbName.value, tableDataTableName.value,
      tableDataPage.value, tableDataPageSize.value
    )
    tableDataColumns.value = res.data.columns
    tableDataRows.value = rowsToObjects(res.data.columns, res.data.rows)
    tableDataTotal.value = res.data.total
  } catch (e: any) {
    message.error(e.message || t('pgPanel.loadTableDataFailed'))
  } finally {
    tableDataLoading.value = false
  }
}

const tableDataDynamicColumns = computed(() => buildDynamicColumns(tableDataColumns.value))

const tableDataPagination = computed(() => ({
  page: tableDataPage.value,
  pageSize: tableDataPageSize.value,
  itemCount: tableDataTotal.value,
  'onUpdate:page': (p: number) => { tableDataPage.value = p; loadTableData() },
  'onUpdate:page-size': (s: number) => { tableDataPageSize.value = s; tableDataPage.value = 1; loadTableData() }
}))

/* --- 创建数据库 --- */

const createDbModal = ref(false)
const createDbLoading = ref(false)
const pgDbForm = ref<CreatePgDatabaseData>({
  name: '',
  encoding: 'UTF8',
  collate: 'en_US.UTF8',
  owner: ''
})

const encodingOptions = [
  { label: 'UTF8', value: 'UTF8' },
  { label: 'SQL_ASCII', value: 'SQL_ASCII' },
  { label: 'LATIN1', value: 'LATIN1' }
]

const collateOptions = [
  { label: 'en_US.UTF8', value: 'en_US.UTF8' },
  { label: 'C', value: 'C' },
  { label: 'POSIX', value: 'POSIX' }
]

function openCreateDb() {
  pgDbForm.value = { name: '', encoding: 'UTF8', collate: 'en_US.UTF8', owner: '' }
  createDbModal.value = true
}

async function submitCreateDb() {
  if (!pgDbForm.value.name) { message.warning(t('pgPanel.createDb.nameRequired')); return }
  createDbLoading.value = true
  try {
    const data: CreatePgDatabaseData = { ...pgDbForm.value }
    if (!data.owner) delete data.owner
    await pgApi.createDatabase(props.instanceId, data)
    message.success(t('pgPanel.createDb.success'))
    createDbModal.value = false
    await loadDatabases()
    selectedDbName.value = pgDbForm.value.name
  } catch (e: any) {
    message.error(e.message || t('pgPanel.createFailed'))
  } finally {
    createDbLoading.value = false
  }
}

/* --- 删除数据库 --- */

const deleteDbConfirm = ref(false)
const deleteDbLoading = ref(false)
const targetDbName = ref('')

function confirmDeleteDb(dbName: string) {
  targetDbName.value = dbName
  deleteDbConfirm.value = true
}

async function doDeleteDb() {
  if (!targetDbName.value) return
  deleteDbLoading.value = true
  try {
    await pgApi.dropDatabase(props.instanceId, targetDbName.value)
    message.success(t('pgPanel.dbDeleted', { name: targetDbName.value }))
    deleteDbConfirm.value = false
    if (selectedDbName.value === targetDbName.value) {
      selectedDbName.value = ''
    }
    await loadDatabases()
  } catch (e: any) {
    message.error(e.message || t('pgPanel.deleteFailed'))
  } finally {
    deleteDbLoading.value = false
  }
}

/* ==================== Tab 2: 用户权限 ==================== */

const users = ref<PgUser[]>([])
const usersLoading = ref(false)

async function loadUsers() {
  usersLoading.value = true
  try {
    const res = await pgApi.getUsers(props.instanceId)
    users.value = res.data
  } catch (e: any) {
    message.error(e.message || t('pgPanel.loadUsersFailed'))
  } finally {
    usersLoading.value = false
  }
}

const pgUserColumns = computed<DataTableColumns>(() => [
  { title: t('pgPanel.columns.username'), key: 'username', width: 200, render: (row: any) => monoEl(row.username) },
  {
    title: t('pgPanel.columns.attributes'), key: 'attributes',
    render: (row: any) => h(NSpace, { size: 4 }, {
      default: () => (row.attributes || []).map((attr: string) =>
        h(NTag, { size: 'small', type: 'info', bordered: false }, { default: () => attr })
      )
    })
  },
  { title: t('pgPanel.columns.validUntil'), key: 'valid_until', width: 180, render: (row: any) => row.valid_until || t('pgPanel.columns.permanent') },
  {
    title: t('pgPanel.columns.actions'), key: 'actions', width: 280,
    render: (row: any) => btnGroup([
      iconBtn({ text: t('pgPanel.columns.grant'), icon: ConstructOutline, type: 'primary', onClick: () => openGrantUser(row) }),
      iconBtn({ text: t('pgPanel.columns.revoke'), icon: ShieldHalfOutline, type: 'warning', onClick: () => openRevokeUser(row) }),
      iconBtn({ text: t('pgPanel.delete'), icon: TrashOutline, type: 'error', onClick: () => confirmDeleteUser(row) })
    ])
  }
])

/* --- 创建用户 --- */

const createUserModal = ref(false)
const createUserLoading = ref(false)
const pgUserForm = ref<{ username: string; password: string; attributes: string[] }>({
  username: '',
  password: '',
  attributes: ['LOGIN']
})

const pgRoleAttributes = ['SUPERUSER', 'CREATEDB', 'CREATEROLE', 'LOGIN']

function openCreateUser() {
  pgUserForm.value = { username: '', password: '', attributes: ['LOGIN'] }
  createUserModal.value = true
}

async function submitCreateUser() {
  if (!pgUserForm.value.username) { message.warning(t('pgPanel.createUser.nameRequired')); return }
  if (!pgUserForm.value.password) { message.warning(t('pgPanel.createUser.passwordRequired')); return }
  createUserLoading.value = true
  try {
    await pgApi.createUser(props.instanceId, {
      username: pgUserForm.value.username,
      password: pgUserForm.value.password,
      attributes: pgUserForm.value.attributes
    })
    message.success(t('pgPanel.createUser.success'))
    createUserModal.value = false
    loadUsers()
  } catch (e: any) {
    message.error(e.message || t('pgPanel.createFailed'))
  } finally {
    createUserLoading.value = false
  }
}

/* --- 删除用户 --- */

const deleteUserConfirm = ref(false)
const deleteUserLoading = ref(false)
const targetUser = ref<PgUser | null>(null)

function confirmDeleteUser(row: any) {
  targetUser.value = row
  deleteUserConfirm.value = true
}

async function doDeleteUser() {
  if (!targetUser.value) return
  deleteUserLoading.value = true
  try {
    await pgApi.dropUser(props.instanceId, targetUser.value.username)
    message.success(t('pgPanel.userDeleted', { name: targetUser.value.username }))
    deleteUserConfirm.value = false
    loadUsers()
  } catch (e: any) {
    message.error(e.message || t('pgPanel.deleteFailed'))
  } finally {
    deleteUserLoading.value = false
  }
}

/* --- 授权 --- */

const grantUserModal = ref(false)
const grantUserLoading = ref(false)
const grantTargetUsername = ref('')
const grantForm = ref<PgGrantData>({ db: '', privileges: ['CONNECT'] })

const pgPrivilegeOptions = ['CONNECT', 'CREATE', 'TEMPORARY', 'ALL']

const dbOptionsForGrant = computed(() =>
  databases.value.map(d => ({ label: d.name, value: d.name }))
)

function openGrantUser(row: any) {
  grantTargetUsername.value = row.username
  grantForm.value = { db: selectedDbName.value || '', privileges: ['CONNECT'] }
  if (databases.value.length === 0) {
    loadDatabases()
  }
  grantUserModal.value = true
}

async function submitGrantUser() {
  if (!grantTargetUsername.value) return
  if (!grantForm.value.db) { message.warning(t('pgPanel.selectDb')); return }
  grantUserLoading.value = true
  try {
    await pgApi.grantUser(props.instanceId, grantTargetUsername.value, grantForm.value)
    message.success(t('pgPanel.grant.success'))
    grantUserModal.value = false
  } catch (e: any) {
    message.error(e.message || t('pgPanel.grant.failed'))
  } finally {
    grantUserLoading.value = false
  }
}

/* --- 回收权限 --- */

const revokeUserModal = ref(false)
const revokeUserLoading = ref(false)
const revokeForm = ref<{ db: string }>({ db: '' })

function openRevokeUser(row: any) {
  grantTargetUsername.value = row.username
  revokeForm.value = { db: selectedDbName.value || '' }
  if (databases.value.length === 0) {
    loadDatabases()
  }
  revokeUserModal.value = true
}

async function submitRevokeUser() {
  if (!grantTargetUsername.value) return
  if (!revokeForm.value.db) { message.warning(t('pgPanel.selectDb')); return }
  revokeUserLoading.value = true
  try {
    await pgApi.revokePgUser(props.instanceId, grantTargetUsername.value, revokeForm.value.db)
    message.success(t('pgPanel.revoke.success', { user: grantTargetUsername.value, db: revokeForm.value.db }))
    revokeUserModal.value = false
  } catch (e: any) {
    message.error(e.message || t('pgPanel.revoke.failed'))
  } finally {
    revokeUserLoading.value = false
  }
}

/* ==================== Tab 3: SQL 执行器 ==================== */

const sqlText = ref(t('pgPanel.sql.defaultSql'))
const sqlLoading = ref(false)
const allowDestructive = ref(false)
const sqlSelectedDb = ref('')
const sqlResult = ref<PgSQLResult | null>(null)
const sqlResultRows = ref<Record<string, any>[]>([])
const sqlResultColumns = ref<string[]>([])

const dbOptionsForSql = computed(() =>
  databases.value.map(d => ({ label: d.name, value: d.name }))
)

watch(selectedDbName, (db) => {
  if (db && !sqlSelectedDb.value) {
    sqlSelectedDb.value = db
  }
})

async function executeSql() {
  if (!sqlSelectedDb.value) { message.warning(t('pgPanel.selectDb')); return }
  if (!sqlText.value.trim()) { message.warning(t('pgPanel.sql.enterSql')); return }
  sqlLoading.value = true
  sqlResult.value = null
  try {
    const res = await pgApi.execSQL(props.instanceId, {
      db: sqlSelectedDb.value,
      sql: sqlText.value,
      allow_destructive: allowDestructive.value
    })
    sqlResult.value = res.data
    if (res.data.columns && res.data.rows) {
      sqlResultColumns.value = res.data.columns
      sqlResultRows.value = rowsToObjects(res.data.columns, res.data.rows)
    } else {
      sqlResultColumns.value = []
      sqlResultRows.value = []
    }
    if (res.data.error) {
      message.error(res.data.error)
    } else {
      message.success(t('pgPanel.sql.executed'))
    }
  } catch (e: any) {
    message.error(e.message || t('pgPanel.sql.failed'))
  } finally {
    sqlLoading.value = false
  }
}

const sqlDynamicColumns = computed(() => buildDynamicColumns(sqlResultColumns.value))

/* ==================== 生命周期 ==================== */

onMounted(() => {
  loadDatabases()
})

watch(activeSubTab, (tab) => {
  if (tab === 'users') {
    loadUsers()
  } else if (tab === 'sql' && databases.value.length === 0) {
    loadDatabases()
  }
})
</script>

<template>
  <NTabs v-model:value="activeSubTab" type="line" animated>
    <!-- ==================== Tab 1: 库表管理 ==================== -->
    <NTabPane name="schema" :tab="t('pgPanel.tabs.schema')">
      <div class="tab-toolbar">
        <NSpace>
          <NButton type="primary" @click="openCreateDb">
            <template #icon><NIcon><AddOutline /></NIcon></template>{{ t('pgPanel.createDb.title') }}
          </NButton>
          <NButton @click="loadDatabases" :loading="databasesLoading">
            <template #icon><NIcon><RefreshOutline /></NIcon></template>{{ t('pgPanel.refresh') }}
          </NButton>
        </NSpace>
      </div>

      <div class="pg-layout">
        <!-- 左侧数据库列表 -->
        <div class="pg-sidebar">
          <div class="pg-sidebar__title">
            <NIcon :size="14"><FileTrayFullOutline /></NIcon>
            <span>{{ t('pgPanel.sidebarTitle') }}</span>
          </div>
          <NSpin :show="databasesLoading">
            <div v-if="databases.length === 0 && !databasesLoading" class="pg-sidebar__empty">
              {{ t('pgPanel.noDatabases') }}
            </div>
            <div
              v-for="db in databases"
              :key="db.name"
              class="db-item"
              :class="{ 'db-item--active': db.name === selectedDbName }"
              @click="selectedDbName = db.name"
            >
              <NIcon :size="14"><FileTrayFullOutline /></NIcon>
              <span class="mono db-item__name">{{ db.name }}</span>
              <NButton
                size="tiny"
                quaternary
                type="error"
                class="db-item__delete"
                @click.stop="confirmDeleteDb(db.name)"
              >
                <template #icon><NIcon :size="12"><TrashOutline /></NIcon></template>
              </NButton>
            </div>
          </NSpin>
        </div>

        <!-- 右侧表列表 -->
        <div class="pg-main">
          <div v-if="!selectedDbName" class="empty-hint">
            <NEmpty :description="t('pgPanel.selectDbLeft')" />
          </div>

          <template v-else>
            <div class="pg-main__header">
              <span class="mono pg-main__dbname">{{ selectedDbName }}</span>
              <NSpace>
                <NButton size="small" @click="loadTables" :loading="tablesLoading">
                  <template #icon><NIcon><RefreshOutline /></NIcon></template>{{ t('pgPanel.refresh') }}
                </NButton>
              </NSpace>
            </div>

            <DataTable
              :columns="pgTableColumns"
              :data="tables"
              :loading="tablesLoading"
              :pagination="false"
            />

            <div v-if="tables.length === 0 && !tablesLoading" class="empty-hint">
              <NEmpty :description="t('pgPanel.noTables')" />
            </div>

            <!-- 展开的表结构 -->
            <div v-if="expandedTableName && tableSchema" class="expanded-section">
              <div class="expanded-section__header">
                <NIcon :size="14"><ConstructOutline /></NIcon>
                <span class="mono">{{ t('pgPanel.tableSchema', { name: expandedTableName }) }}</span>
                <NButton size="tiny" text @click="expandedTableName = ''">
                  <template #icon><NIcon><CloseOutline /></NIcon></template>
                </NButton>
              </div>
              <NSpin :show="tableSchemaLoading">
                <DataTable
                  :columns="pgStructureColumns"
                  :data="tableSchema.columns"
                  :pagination="false"
                  size="small"
                />
              </NSpin>
              <div v-if="tableSchema.indexes && tableSchema.indexes.length > 0" class="pg-indexes">
                <div class="pg-indexes__title">{{ t('pgPanel.indexes') }}</div>
                <div v-for="idx in tableSchema.indexes" :key="idx.name" class="pg-indexes__item">
                  <span class="mono pg-indexes__name">{{ idx.name }}</span>
                  <span class="pg-indexes__def">{{ idx.definition }}</span>
                </div>
              </div>
            </div>

            <!-- 表数据查看 -->
            <div v-if="tableDataVisible" class="expanded-section">
              <div class="expanded-section__header">
                <NIcon :size="14"><ServerOutline /></NIcon>
                <span class="mono">{{ t('pgPanel.tableData', { name: tableDataTableName }) }}</span>
                <NButton size="tiny" text @click="tableDataVisible = false">
                  <template #icon><NIcon><CloseOutline /></NIcon></template>
                </NButton>
              </div>
              <DataTable
                :columns="tableDataDynamicColumns"
                :data="tableDataRows"
                :loading="tableDataLoading"
                :pagination="tableDataPagination"
                size="small"
              />
            </div>
          </template>
        </div>
      </div>
    </NTabPane>

    <!-- ==================== Tab 2: 用户权限 ==================== -->
    <NTabPane name="users" :tab="t('pgPanel.tabs.users')">
      <div class="tab-toolbar">
        <NSpace>
          <NButton type="primary" @click="openCreateUser">
            <template #icon><NIcon><AddOutline /></NIcon></template>{{ t('pgPanel.createUser.create') }}
          </NButton>
          <NButton @click="loadUsers" :loading="usersLoading">
            <template #icon><NIcon><RefreshOutline /></NIcon></template>{{ t('pgPanel.refresh') }}
          </NButton>
        </NSpace>
      </div>

      <DataTable
        :columns="pgUserColumns"
        :data="users"
        :loading="usersLoading"
        :pagination="false"
      />
      <div v-if="users.length === 0 && !usersLoading" class="empty-hint">
        <NEmpty :description="t('pgPanel.noUsers')" />
      </div>
    </NTabPane>

    <!-- ==================== Tab 3: SQL 执行器 ==================== -->
    <NTabPane name="sql" :tab="t('pgPanel.tabs.sql')">
      <div class="tab-toolbar">
        <NSpace>
          <NSelect
            v-model:value="sqlSelectedDb"
            :placeholder="t('pgPanel.selectDbPlaceholder')"
            :options="dbOptionsForSql"
            style="width: var(--input-lg)"
          />
          <NButton type="primary" :loading="sqlLoading" :disabled="!sqlSelectedDb" @click="executeSql">
            <template #icon><NIcon><PlayOutline /></NIcon></template>{{ t('pgPanel.execute') }}
          </NButton>
          <NCheckbox v-model:checked="allowDestructive">
            <span class="checkbox-label">{{ t('pgPanel.allowDestructive') }}</span>
          </NCheckbox>
        </NSpace>
      </div>

      <div v-if="!sqlSelectedDb" class="empty-hint">
        <NEmpty :description="t('pgPanel.selectDbFirst')" />
      </div>

      <template v-else>
        <div class="sql-editor-wrap">
          <CodeEditor
            v-model="sqlText"
            language="pgsql"
            height="300px"
            :minimap="false"
          />
        </div>

        <!-- 执行结果 -->
        <div v-if="sqlResult" class="sql-result">
          <div class="sql-result__header">
            <NSpace>
              <NTag size="small" :type="sqlResult.error ? 'error' : 'success'" bordered>
                {{ sqlResult.error ? t('pgPanel.sql.failed') : t('pgPanel.sql.success') }}
              </NTag>
              <span class="mono sql-result__meta">
                {{ t('pgPanel.sql.duration', { ms: sqlResult.duration_ms }) }}
                <template v-if="sqlResult.affected_rows !== undefined">
                  · {{ t('pgPanel.sql.affected', { rows: sqlResult.affected_rows }) }}
                </template>
              </span>
            </NSpace>
          </div>

          <NAlert v-if="sqlResult.error" type="error" :show-icon="true" class="sql-result__error">
            {{ sqlResult.error }}
          </NAlert>

          <div v-if="sqlResultRows.length > 0" class="sql-result__table">
            <DataTable
              :columns="sqlDynamicColumns"
              :data="sqlResultRows"
              :pagination="false"
              size="small"
            />
          </div>

          <div v-else-if="!sqlResult.error && !sqlResult.columns" class="sql-result__empty">
            {{ t('pgPanel.sql.noResultSet') }}
          </div>
        </div>
      </template>
    </NTabPane>
  </NTabs>

  <!-- ==================== Modals ==================== -->

  <!-- 创建数据库 -->
  <ModalForm
    v-model:show="createDbModal"
    :title="t('pgPanel.createDb.title')"
    :loading="createDbLoading"
    @submit="submitCreateDb"
  >
    <NFormItem :label="t('pgPanel.createDb.nameLabel')">
      <NInput v-model:value="pgDbForm.name" :placeholder="t('pgPanel.createDb.namePlaceholder')" />
    </NFormItem>
    <NFormItem :label="t('pgPanel.createDb.encodingLabel')">
      <NSelect v-model:value="pgDbForm.encoding" :options="encodingOptions" />
    </NFormItem>
    <NFormItem :label="t('pgPanel.createDb.collateLabel')">
      <NSelect v-model:value="pgDbForm.collate" :options="collateOptions" />
    </NFormItem>
    <NFormItem label="Owner">
      <NInput v-model:value="pgDbForm.owner" :placeholder="t('pgPanel.createDb.ownerPlaceholder')" />
    </NFormItem>
  </ModalForm>

  <!-- 创建用户 -->
  <ModalForm
    v-model:show="createUserModal"
    :title="t('pgPanel.createUser.title')"
    :loading="createUserLoading"
    @submit="submitCreateUser"
  >
    <NFormItem :label="t('pgPanel.createUser.nameLabel')">
      <NInput v-model:value="pgUserForm.username" :placeholder="t('pgPanel.createUser.namePlaceholder')" />
    </NFormItem>
    <NFormItem :label="t('pgPanel.createUser.passwordLabel')">
      <NInput v-model:value="pgUserForm.password" type="password" show-password-on="click" :placeholder="t('pgPanel.createUser.passwordPlaceholder')" />
    </NFormItem>
    <NFormItem :label="t('pgPanel.createUser.roleAttributesLabel')">
      <NCheckboxGroup v-model:value="pgUserForm.attributes">
        <NSpace>
          <NCheckbox v-for="attr in pgRoleAttributes" :key="attr" :value="attr">{{ attr }}</NCheckbox>
        </NSpace>
      </NCheckboxGroup>
    </NFormItem>
  </ModalForm>

  <!-- 授权 -->
  <ModalForm
    v-model:show="grantUserModal"
    :title="t('pgPanel.grant.title', { name: grantTargetUsername })"
    :loading="grantUserLoading"
    @submit="submitGrantUser"
  >
    <NFormItem :label="t('pgPanel.selectDbLabel')">
      <NSelect v-model:value="grantForm.db" :options="dbOptionsForGrant" :placeholder="t('pgPanel.selectDbPlaceholder')" />
    </NFormItem>
    <NFormItem :label="t('pgPanel.grant.privilegesLabel')">
      <NCheckboxGroup v-model:value="grantForm.privileges">
        <NSpace>
          <NCheckbox v-for="p in pgPrivilegeOptions" :key="p" :value="p">{{ p }}</NCheckbox>
        </NSpace>
      </NCheckboxGroup>
    </NFormItem>
  </ModalForm>

  <!-- 回收权限 -->
  <ModalForm
    v-model:show="revokeUserModal"
    :title="t('pgPanel.revoke.title', { name: grantTargetUsername })"
    :loading="revokeUserLoading"
    @submit="submitRevokeUser"
  >
    <NFormItem :label="t('pgPanel.selectDbLabel')">
      <NSelect v-model:value="revokeForm.db" :options="dbOptionsForGrant" :placeholder="t('pgPanel.revoke.dbPlaceholder')" />
    </NFormItem>
  </ModalForm>

  <!-- ==================== Confirm Dialogs ==================== -->

  <ConfirmDialog
    v-model:show="deleteDbConfirm"
    :title="t('pgPanel.deleteDb.title')"
    :content="t('pgPanel.deleteDb.content', { name: targetDbName })"
    type="error"
    :confirm-text="t('pgPanel.delete')"
    :loading="deleteDbLoading"
    @confirm="doDeleteDb"
  />

  <ConfirmDialog
    v-model:show="deleteUserConfirm"
    :title="t('pgPanel.deleteUser.title')"
    :content="t('pgPanel.deleteUser.content', { name: targetUser?.username ?? '' })"
    type="error"
    :confirm-text="t('pgPanel.delete')"
    :loading="deleteUserLoading"
    @confirm="doDeleteUser"
  />
</template>

<style scoped>
.tab-toolbar {
  margin-bottom: var(--space-4);
  display: flex;
  align-items: center;
}

.empty-hint {
  padding: var(--space-8) 0;
  display: flex;
  justify-content: center;
}

/* PG 布局 */
.pg-layout {
  display: flex;
  gap: var(--space-4);
  min-height: var(--chart-height);
}

.pg-sidebar {
  width: var(--input-xl);
  flex-shrink: 0;
  background: var(--surface2);
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  padding: var(--space-3);
  overflow-y: auto;
  max-height: 70vh;
}

.pg-sidebar__title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--fs-caption);
  font-weight: 600;
  color: var(--sub);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  margin-bottom: var(--space-3);
  padding-bottom: var(--space-2);
  border-bottom: 1px solid var(--line);
}

.pg-sidebar__empty {
  color: var(--text-quaternary);
  font-size: var(--fs-caption);
  text-align: center;
  padding: var(--space-4) 0;
}

.db-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background var(--t-fast) var(--ease);
  color: var(--sub);
  position: relative;
}

.db-item:hover {
  background: var(--surface3);
  color: var(--ink);
}

.db-item--active {
  background: var(--acc-soft);
  color: var(--acc);
  font-weight: 500;
}

.db-item__name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--fs-body);
}

.db-item__delete {
  opacity: 0;
  transition: opacity var(--t-fast) var(--ease);
}

.db-item:hover .db-item__delete {
  opacity: 1;
}

.pg-main {
  flex: 1;
  min-width: 0;
}

.pg-main__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-3);
}

.pg-main__dbname {
  font-size: var(--fs-title);
  font-weight: 600;
  color: var(--ink);
}

/* 展开区域 */
.expanded-section {
  margin-top: var(--space-4);
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.expanded-section__header {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-4);
  background: var(--surface2);
  border-bottom: 1px solid var(--line);
  font-size: var(--fs-body);
  font-weight: 500;
  color: var(--ink);
}

.expanded-section__header > :nth-child(3) {
  margin-left: auto;
}

/* 索引 */
.pg-indexes {
  padding: var(--space-3) var(--space-4);
  border-top: 1px solid var(--line);
}

.pg-indexes__title {
  font-size: var(--fs-caption);
  font-weight: 600;
  color: var(--sub);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  margin-bottom: var(--space-2);
}

.pg-indexes__item {
  display: flex;
  gap: var(--space-3);
  padding: var(--space-1) 0;
  font-size: var(--fs-caption);
}

.pg-indexes__name {
  color: var(--acc);
  font-weight: 500;
  flex-shrink: 0;
}

.pg-indexes__def {
  color: var(--text-quaternary);
  font-family: var(--font-mono);
}

/* SQL 执行器 */
.sql-editor-wrap {
  margin-bottom: var(--space-4);
}

.checkbox-label {
  font-size: var(--fs-body);
  color: var(--sub);
}

.sql-result {
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.sql-result__header {
  display: flex;
  align-items: center;
  padding: var(--space-3) var(--space-4);
  background: var(--surface2);
  border-bottom: 1px solid var(--line);
}

.sql-result__meta {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
}

.sql-result__error {
  margin: var(--space-3);
}

.sql-result__table {
  padding: var(--space-3);
}

.sql-result__empty {
  padding: var(--space-6);
  text-align: center;
  color: var(--text-quaternary);
  font-size: var(--fs-caption);
}

:deep(.n-tabs-nav) {
  margin-bottom: var(--space-4);
}
</style>
