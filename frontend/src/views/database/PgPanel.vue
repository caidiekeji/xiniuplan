<script setup lang="ts">
import { ref, computed, onMounted, watch, h } from 'vue'
import { useMessage } from 'naive-ui'
import {
  NButton, NSpace, NIcon, NTabs, NTabPane, NInput, NSelect,
  NCheckboxGroup, NCheckbox, NFormItem, NEmpty, NSpin, NAlert, NTag
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {
  AddOutline, TrashOutline, PlayOutline, RefreshOutline,
  ConstructOutline, PeopleOutline, FileTrayFullOutline,
  ChevronForwardOutline, CloseOutline, ServerOutline, ShieldHalfOutline
} from '@vicons/ionicons5'
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

function buildDynamicColumns(columns: string[]): DataTableColumns {
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
    message.error(e.message || '加载数据库列表失败')
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
    message.error(e.message || '加载表列表失败')
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
    message.error(e.message || '加载表结构失败')
  } finally {
    tableSchemaLoading.value = false
  }
}

const pgTableColumns: DataTableColumns = [
  {
    title: '表名', key: 'name', width: 220,
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
    title: '行数（估算）', key: 'rows_estimate', width: 120,
    render: (row: any) => monoEl(row.rows_estimate?.toLocaleString() ?? '-')
  },
  { title: '大小', key: 'size', width: 120 },
  {
    title: '操作', key: 'actions', width: 160,
    render: (row: any) => btnGroup([
      iconBtn({ text: '结构', icon: ConstructOutline, onClick: () => toggleTableSchema(row.name) }),
      iconBtn({ text: '数据', icon: ServerOutline, type: 'primary', onClick: () => viewTableData(row.name) })
    ])
  }
]

const pgStructureColumns: DataTableColumns = [
  { title: '列名', key: 'name', width: 180, render: (row: any) => monoEl(row.name) },
  { title: '类型', key: 'type', width: 160, render: (row: any) => monoEl(row.type) },
  {
    title: '可空', key: 'nullable', width: 80,
    render: (row: any) => row.nullable ? 'YES' : 'NO'
  },
  { title: '默认值', key: 'default', width: 160, render: (row: any) => row.default ?? '-' },
  {
    title: '主键', key: 'is_primary', width: 80,
    render: (row: any) => row.is_primary
      ? h(NTag, { size: 'small', type: 'warning', bordered: false }, { default: () => 'PK' })
      : '-'
  }
]

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
    message.error(e.message || '加载表数据失败')
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
  if (!pgDbForm.value.name) { message.warning('请输入数据库名'); return }
  createDbLoading.value = true
  try {
    const data: CreatePgDatabaseData = { ...pgDbForm.value }
    if (!data.owner) delete data.owner
    await pgApi.createDatabase(props.instanceId, data)
    message.success('数据库创建成功')
    createDbModal.value = false
    await loadDatabases()
    selectedDbName.value = pgDbForm.value.name
  } catch (e: any) {
    message.error(e.message || '创建失败')
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
    message.success(`数据库 ${targetDbName.value} 已删除`)
    deleteDbConfirm.value = false
    if (selectedDbName.value === targetDbName.value) {
      selectedDbName.value = ''
    }
    await loadDatabases()
  } catch (e: any) {
    message.error(e.message || '删除失败')
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
    message.error(e.message || '加载用户列表失败')
  } finally {
    usersLoading.value = false
  }
}

const pgUserColumns: DataTableColumns = [
  { title: '用户名', key: 'username', width: 200, render: (row: any) => monoEl(row.username) },
  {
    title: '属性', key: 'attributes',
    render: (row: any) => h(NSpace, { size: 4 }, {
      default: () => (row.attributes || []).map((attr: string) =>
        h(NTag, { size: 'small', type: 'info', bordered: false }, { default: () => attr })
      )
    })
  },
  { title: '有效时间', key: 'valid_until', width: 180, render: (row: any) => row.valid_until || '永久' },
  {
    title: '操作', key: 'actions', width: 280,
    render: (row: any) => btnGroup([
      iconBtn({ text: '授权', icon: ConstructOutline, type: 'primary', onClick: () => openGrantUser(row) }),
      iconBtn({ text: '回收权限', icon: ShieldHalfOutline, type: 'warning', onClick: () => openRevokeUser(row) }),
      iconBtn({ text: '删除', icon: TrashOutline, type: 'error', onClick: () => confirmDeleteUser(row) })
    ])
  }
]

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
  if (!pgUserForm.value.username) { message.warning('请输入用户名'); return }
  if (!pgUserForm.value.password) { message.warning('请输入密码'); return }
  createUserLoading.value = true
  try {
    await pgApi.createUser(props.instanceId, {
      username: pgUserForm.value.username,
      password: pgUserForm.value.password,
      attributes: pgUserForm.value.attributes
    })
    message.success('用户创建成功')
    createUserModal.value = false
    loadUsers()
  } catch (e: any) {
    message.error(e.message || '创建失败')
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
    message.success(`用户 ${targetUser.value.username} 已删除`)
    deleteUserConfirm.value = false
    loadUsers()
  } catch (e: any) {
    message.error(e.message || '删除失败')
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
  if (!grantForm.value.db) { message.warning('请选择数据库'); return }
  grantUserLoading.value = true
  try {
    await pgApi.grantUser(props.instanceId, grantTargetUsername.value, grantForm.value)
    message.success('授权成功')
    grantUserModal.value = false
  } catch (e: any) {
    message.error(e.message || '授权失败')
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
  if (!revokeForm.value.db) { message.warning('请选择数据库'); return }
  revokeUserLoading.value = true
  try {
    await pgApi.revokePgUser(props.instanceId, grantTargetUsername.value, revokeForm.value.db)
    message.success(`已回收 ${grantTargetUsername.value} 在 ${revokeForm.value.db} 上的权限`)
    revokeUserModal.value = false
  } catch (e: any) {
    message.error(e.message || '回收权限失败')
  } finally {
    revokeUserLoading.value = false
  }
}

/* ==================== Tab 3: SQL 执行器 ==================== */

const sqlText = ref('-- 在此输入 SQL 语句\nSELECT 1;')
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
  if (!sqlSelectedDb.value) { message.warning('请选择数据库'); return }
  if (!sqlText.value.trim()) { message.warning('请输入 SQL 语句'); return }
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
      message.success('执行完成')
    }
  } catch (e: any) {
    message.error(e.message || '执行失败')
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
    <NTabPane name="schema" tab="库表管理">
      <div class="tab-toolbar">
        <NSpace>
          <NButton type="primary" @click="openCreateDb">
            <template #icon><NIcon><AddOutline /></NIcon></template>创建数据库
          </NButton>
          <NButton @click="loadDatabases" :loading="databasesLoading">
            <template #icon><NIcon><RefreshOutline /></NIcon></template>刷新
          </NButton>
        </NSpace>
      </div>

      <div class="pg-layout">
        <!-- 左侧数据库列表 -->
        <div class="pg-sidebar">
          <div class="pg-sidebar__title">
            <NIcon :size="14"><FileTrayFullOutline /></NIcon>
            <span>数据库</span>
          </div>
          <NSpin :show="databasesLoading">
            <div v-if="databases.length === 0 && !databasesLoading" class="pg-sidebar__empty">
              暂无数据库
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
            <NEmpty description="请选择左侧数据库" />
          </div>

          <template v-else>
            <div class="pg-main__header">
              <span class="mono pg-main__dbname">{{ selectedDbName }}</span>
              <NSpace>
                <NButton size="small" @click="loadTables" :loading="tablesLoading">
                  <template #icon><NIcon><RefreshOutline /></NIcon></template>刷新
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
              <NEmpty description="该数据库暂无表" />
            </div>

            <!-- 展开的表结构 -->
            <div v-if="expandedTableName && tableSchema" class="expanded-section">
              <div class="expanded-section__header">
                <NIcon :size="14"><ConstructOutline /></NIcon>
                <span class="mono">表结构: {{ expandedTableName }}</span>
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
                <div class="pg-indexes__title">索引</div>
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
                <span class="mono">表数据: {{ tableDataTableName }}</span>
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
    <NTabPane name="users" tab="用户权限">
      <div class="tab-toolbar">
        <NSpace>
          <NButton type="primary" @click="openCreateUser">
            <template #icon><NIcon><AddOutline /></NIcon></template>创建用户
          </NButton>
          <NButton @click="loadUsers" :loading="usersLoading">
            <template #icon><NIcon><RefreshOutline /></NIcon></template>刷新
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
        <NEmpty description="暂无用户，点击上方按钮创建" />
      </div>
    </NTabPane>

    <!-- ==================== Tab 3: SQL 执行器 ==================== -->
    <NTabPane name="sql" tab="SQL 执行器">
      <div class="tab-toolbar">
        <NSpace>
          <NSelect
            v-model:value="sqlSelectedDb"
            placeholder="选择数据库"
            :options="dbOptionsForSql"
            style="width: var(--input-lg)"
          />
          <NButton type="primary" :loading="sqlLoading" :disabled="!sqlSelectedDb" @click="executeSql">
            <template #icon><NIcon><PlayOutline /></NIcon></template>执行
          </NButton>
          <NCheckbox v-model:checked="allowDestructive">
            <span class="checkbox-label">允许危险操作</span>
          </NCheckbox>
        </NSpace>
      </div>

      <div v-if="!sqlSelectedDb" class="empty-hint">
        <NEmpty description="请先选择数据库" />
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
                {{ sqlResult.error ? '执行失败' : '执行成功' }}
              </NTag>
              <span class="mono sql-result__meta">
                耗时 <span class="num">{{ sqlResult.duration_ms }}</span> ms
                <template v-if="sqlResult.affected_rows !== undefined">
                  · 影响 <span class="num">{{ sqlResult.affected_rows }}</span> 行
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
            无返回结果集（非 SELECT 语句）
          </div>
        </div>
      </template>
    </NTabPane>
  </NTabs>

  <!-- ==================== Modals ==================== -->

  <!-- 创建数据库 -->
  <ModalForm
    v-model:show="createDbModal"
    title="创建数据库"
    :loading="createDbLoading"
    @submit="submitCreateDb"
  >
    <NFormItem label="数据库名">
      <NInput v-model:value="pgDbForm.name" placeholder="例如: myapp_db" />
    </NFormItem>
    <NFormItem label="编码">
      <NSelect v-model:value="pgDbForm.encoding" :options="encodingOptions" />
    </NFormItem>
    <NFormItem label="排序规则">
      <NSelect v-model:value="pgDbForm.collate" :options="collateOptions" />
    </NFormItem>
    <NFormItem label="Owner">
      <NInput v-model:value="pgDbForm.owner" placeholder="留空使用当前用户" />
    </NFormItem>
  </ModalForm>

  <!-- 创建用户 -->
  <ModalForm
    v-model:show="createUserModal"
    title="创建数据库用户"
    :loading="createUserLoading"
    @submit="submitCreateUser"
  >
    <NFormItem label="用户名">
      <NInput v-model:value="pgUserForm.username" placeholder="例如: app_user" />
    </NFormItem>
    <NFormItem label="密码">
      <NInput v-model:value="pgUserForm.password" type="password" show-password-on="click" placeholder="请输入密码" />
    </NFormItem>
    <NFormItem label="角色属性">
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
    :title="`授权用户: ${grantTargetUsername}`"
    :loading="grantUserLoading"
    @submit="submitGrantUser"
  >
    <NFormItem label="选择数据库">
      <NSelect v-model:value="grantForm.db" :options="dbOptionsForGrant" placeholder="选择数据库" />
    </NFormItem>
    <NFormItem label="权限">
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
    :title="`回收权限: ${grantTargetUsername}`"
    :loading="revokeUserLoading"
    @submit="submitRevokeUser"
  >
    <NFormItem label="选择数据库">
      <NSelect v-model:value="revokeForm.db" :options="dbOptionsForGrant" placeholder="选择要回收权限的数据库" />
    </NFormItem>
  </ModalForm>

  <!-- ==================== Confirm Dialogs ==================== -->

  <ConfirmDialog
    v-model:show="deleteDbConfirm"
    title="删除数据库"
    :content="`确定要删除数据库「${targetDbName}」吗？此操作不可恢复。`"
    type="error"
    confirm-text="删除"
    :loading="deleteDbLoading"
    @confirm="doDeleteDb"
  />

  <ConfirmDialog
    v-model:show="deleteUserConfirm"
    title="删除用户"
    :content="`确定要删除用户「${targetUser?.username}」吗？`"
    type="error"
    confirm-text="删除"
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
  color: var(--faint);
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
  color: var(--faint);
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
  color: var(--faint);
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
  color: var(--faint);
  font-size: var(--fs-caption);
}

:deep(.n-tabs-nav) {
  margin-bottom: var(--space-4);
}
</style>
