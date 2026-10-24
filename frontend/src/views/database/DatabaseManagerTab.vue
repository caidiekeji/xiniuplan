<script setup lang="ts">
import { ref, computed, watch, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from '@/composables/useMessage'
import {
  NButton, NSpace, NIcon, NSelect, NInput, NFormItem, NEmpty, NSpin
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import {
  AddOutline, TrashOutline, RefreshOutline, ConstructOutline,
  ServerOutline, ChevronForwardOutline, CloseOutline, FileTrayFullOutline, PeopleOutline
} from '@vicons/ionicons5'
import { DataTable, ModalForm, ConfirmDialog } from '@/components'
import * as dbApi from '@/api/database'
import type { DatabaseInfo, TableInfo, TableColumn, CreateDatabaseData } from '@/api/database'
import { monoEl, iconBtn, btnGroup, rowsToObjects, buildDynamicColumns, formatTime } from './shared'

const message = useMessage()
const { t } = useI18n()

/* ==================== Props / Emits ==================== */

const props = defineProps<{
  instanceId: number | null
  instanceOptions: { label: string; value: number }[]
  databases: DatabaseInfo[]
  databasesLoading: boolean
}>()

const emit = defineEmits<{
  (e: 'update:instanceId', val: number | null): void
  (e: 'loadDatabases'): void
  (e: 'createUser'): void
}>()

/* ==================== 数据库选择 ==================== */

const selectedDbName = ref<string>('')

watch(
  () => props.databases,
  (dbs) => {
    if (dbs.length > 0 && !dbs.find((d) => d.name === selectedDbName.value)) {
      selectedDbName.value = dbs[0].name
    }
  }
)

watch(selectedDbName, () => {
  loadTables()
  expandedTableName.value = ''
})

/* ==================== 表列表 ==================== */

const tables = ref<TableInfo[]>([])
const tablesLoading = ref(false)
const expandedTableName = ref('')
const tableStructure = ref<TableColumn[]>([])
const tableStructureLoading = ref(false)

async function loadTables() {
  if (!props.instanceId || !selectedDbName.value) return
  tablesLoading.value = true
  try {
    const res = await dbApi.getTables(props.instanceId, selectedDbName.value)
    tables.value = res.data
  } catch (e: any) {
    message.error(e.message || t('database.manager.loadTablesFailed'))
  } finally {
    tablesLoading.value = false
  }
}

async function toggleTableStructure(tableName: string) {
  if (expandedTableName.value === tableName) {
    expandedTableName.value = ''
    return
  }
  expandedTableName.value = tableName
  if (!props.instanceId) return
  tableStructureLoading.value = true
  try {
    const res = await dbApi.getTableStructure(props.instanceId, selectedDbName.value, tableName)
    tableStructure.value = res.data
  } catch (e: any) {
    message.error(e.message || t('database.manager.loadStructureFailed'))
  } finally {
    tableStructureLoading.value = false
  }
}

const tableColumnsDef = computed<DataTableColumns>(() => [
  {
    title: t('database.manager.columns.tableName'), key: 'name', width: 220,
    render: (row: any) =>
      h(
        NButton,
        { text: true, size: 'small', onClick: () => toggleTableStructure(row.name) },
        {
          default: () =>
            h(NSpace, { size: 4 }, {
              default: () => [
                h(NIcon, { size: 14 }, { default: () => h(ChevronForwardOutline) }),
                monoEl(row.name)
              ]
            })
        }
      )
  },
  { title: t('database.manager.columns.engine'), key: 'engine', width: 100 },
  {
    title: t('database.manager.columns.rows'), key: 'rows', width: 100,
    render: (row: any) => monoEl(row.rows?.toLocaleString() ?? 0)
  },
  { title: t('database.manager.columns.collation'), key: 'collation', width: 160 },
  {
    title: t('database.manager.columns.createdAt'), key: 'created_at', width: 170,
    render: (row: any) => formatTime(row.created_at)
  },
  {
    title: t('database.manager.columns.actions'), key: 'actions', width: 200,
    render: (row: any) =>
      btnGroup([
        iconBtn({ text: t('database.manager.structure'), icon: ConstructOutline, onClick: () => toggleTableStructure(row.name) }),
        iconBtn({ text: t('database.manager.data'), icon: ServerOutline, type: 'primary', onClick: () => viewTableData(row.name) }),
        iconBtn({ text: t('database.common.delete'), icon: TrashOutline, type: 'error', onClick: () => confirmDeleteTable(row) })
      ])
  }
])

const structureColumns = computed<DataTableColumns>(() => [
  { title: t('database.manager.columns.field'), key: 'field', width: 180, render: (row: any) => monoEl(row.field) },
  { title: t('database.manager.columns.type'), key: 'type', width: 160, render: (row: any) => monoEl(row.type) },
  { title: t('database.manager.columns.nullable'), key: 'null', width: 80 },
  { title: t('database.manager.columns.key'), key: 'key', width: 80 },
  { title: t('database.manager.columns.defaultValue'), key: 'default', width: 140, render: (row: any) => row.default ?? '-' },
  { title: t('database.manager.columns.extra'), key: 'extra', width: 120 }
])

/* ==================== 表数据 ==================== */

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
  if (!props.instanceId || !selectedDbName.value || !tableDataTableName.value) return
  tableDataLoading.value = true
  try {
    const res = await dbApi.getTableData(
      props.instanceId,
      selectedDbName.value,
      tableDataTableName.value,
      tableDataPage.value,
      tableDataPageSize.value
    )
    tableDataColumns.value = res.data.columns
    tableDataRows.value = rowsToObjects(res.data.columns, res.data.rows)
    tableDataTotal.value = res.data.total
  } catch (e: any) {
    message.error(e.message || t('database.manager.loadDataFailed'))
  } finally {
    tableDataLoading.value = false
  }
}

const tableDataDynamicColumns = computed(() => buildDynamicColumns(tableDataColumns.value))

const tableDataPagination = computed(() => ({
  page: tableDataPage.value,
  pageSize: tableDataPageSize.value,
  itemCount: tableDataTotal.value,
  'onUpdate:page': (p: number) => {
    tableDataPage.value = p
    loadTableData()
  },
  'onUpdate:page-size': (s: number) => {
    tableDataPageSize.value = s
    tableDataPage.value = 1
    loadTableData()
  }
}))

/* ==================== 创建数据库 Modal ==================== */

const createDbModal = ref(false)
const createDbLoading = ref(false)
const dbForm = ref<CreateDatabaseData>({ name: '', charset: 'utf8mb4', collation: 'utf8mb4_general_ci' })

const charsetOptions = [
  { label: 'utf8mb4', value: 'utf8mb4' },
  { label: 'utf8', value: 'utf8' },
  { label: 'latin1', value: 'latin1' }
]

const collationOptions = [
  { label: 'utf8mb4_general_ci', value: 'utf8mb4_general_ci' },
  { label: 'utf8mb4_unicode_ci', value: 'utf8mb4_unicode_ci' },
  { label: 'utf8_general_ci', value: 'utf8_general_ci' },
  { label: 'latin1_swedish_ci', value: 'latin1_swedish_ci' }
]

function openCreateDb() {
  dbForm.value = { name: '', charset: 'utf8mb4', collation: 'utf8mb4_general_ci' }
  createDbModal.value = true
}

async function submitCreateDb() {
  if (!props.instanceId) return
  if (!dbForm.value.name) {
    message.warning(t('database.manager.enterDbName'))
    return
  }
  createDbLoading.value = true
  try {
    await dbApi.createDatabase(props.instanceId, dbForm.value)
    message.success(t('database.manager.createSuccess'))
    createDbModal.value = false
    emit('loadDatabases')
    selectedDbName.value = dbForm.value.name
  } catch (e: any) {
    message.error(e.message || t('database.manager.createFailed'))
  } finally {
    createDbLoading.value = false
  }
}

/* ==================== 删除数据库确认 ==================== */

const deleteDbConfirm = ref(false)
const deleteDbLoading = ref(false)
const targetDbName = ref('')

function confirmDeleteDb(dbName: string) {
  targetDbName.value = dbName
  deleteDbConfirm.value = true
}

async function doDeleteDb() {
  if (!props.instanceId || !targetDbName.value) return
  deleteDbLoading.value = true
  try {
    await dbApi.deleteDatabase(props.instanceId, targetDbName.value)
    message.success(t('database.manager.dbDeleted', { name: targetDbName.value }))
    deleteDbConfirm.value = false
    if (selectedDbName.value === targetDbName.value) {
      selectedDbName.value = ''
    }
    emit('loadDatabases')
  } catch (e: any) {
    message.error(e.message || t('database.manager.deleteFailed'))
  } finally {
    deleteDbLoading.value = false
  }
}

/* ==================== 删除表确认 ==================== */

const deleteTableConfirm = ref(false)
const deleteTableLoading = ref(false)
const targetTableName = ref('')

function confirmDeleteTable(row: any) {
  targetTableName.value = row.name
  deleteTableConfirm.value = true
}

async function doDeleteTable() {
  message.info(t('database.manager.useSqlExecutor'))
  deleteTableConfirm.value = false
}

/* ==================== 实例切换 ==================== */

function onInstanceChange(val: number | null) {
  emit('update:instanceId', val)
  selectedDbName.value = ''
  tables.value = []
  tableDataVisible.value = false
}
</script>

<template>
  <div class="db-manager-tab">
    <div class="tab-toolbar">
      <NSpace>
        <NSelect
          :value="instanceId"
          :placeholder="t('database.common.selectMysqlInstance')"
          :options="instanceOptions"
          style="width: var(--input-3xl)"
          @update:value="onInstanceChange"
        />
        <NButton type="primary" :disabled="!instanceId" @click="openCreateDb">
          <template #icon><NIcon><AddOutline /></NIcon></template>{{ t('database.manager.createDb') }}
        </NButton>
        <NButton :disabled="!instanceId" @click="$emit('createUser')">
          <template #icon><NIcon><PeopleOutline /></NIcon></template>{{ t('database.manager.createUser') }}
        </NButton>
      </NSpace>
    </div>

    <div v-if="!instanceId" class="empty-hint">
      <NEmpty :description="t('database.common.selectInstance')" />
    </div>

    <div v-else class="schema-layout">
      <!-- 左侧数据库列表 -->
      <div class="schema-sidebar">
        <div class="schema-sidebar__title">
          <NIcon :size="14"><FileTrayFullOutline /></NIcon>
          <span>{{ t('database.manager.sidebarTitle') }}</span>
        </div>
        <NSpin :show="databasesLoading">
          <div v-if="databases.length === 0 && !databasesLoading" class="schema-sidebar__empty">
            {{ t('database.manager.noDatabases') }}
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
      <div class="schema-main">
        <div v-if="!selectedDbName" class="empty-hint">
          <NEmpty :description="t('database.manager.selectDbLeft')" />
        </div>

        <template v-else>
          <div class="schema-main__header">
            <span class="mono schema-main__dbname">{{ selectedDbName }}</span>
            <NSpace>
              <NButton size="small" @click="loadTables" :loading="tablesLoading">
                <template #icon><NIcon><RefreshOutline /></NIcon></template>{{ t('database.common.refresh') }}
              </NButton>
            </NSpace>
          </div>

          <DataTable
            :columns="tableColumnsDef"
            :data="tables"
            :loading="tablesLoading"
            :pagination="false"
          />

          <div v-if="tables.length === 0 && !tablesLoading" class="empty-hint">
            <NEmpty :description="t('database.manager.noTables')" />
          </div>

          <!-- 展开的表结构 -->
          <div v-if="expandedTableName" class="expanded-section">
            <div class="expanded-section__header">
              <NIcon :size="14"><ConstructOutline /></NIcon>
              <span class="mono">{{ t('database.manager.tableSchema', { name: expandedTableName }) }}</span>
              <NButton size="tiny" text @click="expandedTableName = ''">
                <template #icon><NIcon><CloseOutline /></NIcon></template>
              </NButton>
            </div>
            <DataTable
              :columns="structureColumns"
              :data="tableStructure"
              :loading="tableStructureLoading"
              :pagination="false"
              size="small"
            />
          </div>

          <!-- 表数据查看（虚拟滚动） -->
          <div v-if="tableDataVisible" class="expanded-section">
            <div class="expanded-section__header">
              <NIcon :size="14"><ServerOutline /></NIcon>
              <span class="mono">{{ t('database.manager.tableData', { name: tableDataTableName }) }}</span>
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
              :virtual-scroll="true"
              :max-height="400"
            />
          </div>
        </template>
      </div>
    </div>

    <!-- 创建数据库 -->
    <ModalForm
      v-model:show="createDbModal"
      :title="t('database.manager.createDbModalTitle')"
      :loading="createDbLoading"
      @submit="submitCreateDb"
    >
      <NFormItem :label="t('database.manager.form.dbName')">
        <NInput v-model:value="dbForm.name" :placeholder="t('database.manager.form.dbNamePlaceholder')" />
      </NFormItem>
      <NFormItem :label="t('database.manager.form.charset')">
        <NSelect v-model:value="dbForm.charset" :options="charsetOptions" />
      </NFormItem>
      <NFormItem :label="t('database.manager.form.collation')">
        <NSelect v-model:value="dbForm.collation" :options="collationOptions" />
      </NFormItem>
    </ModalForm>

    <!-- 删除数据库确认 -->
    <ConfirmDialog
      v-model:show="deleteDbConfirm"
      :title="t('database.manager.deleteDbTitle')"
      :content="t('database.manager.deleteDbContent', { name: targetDbName })"
      type="error"
      :confirm-text="t('database.common.delete')"
      :loading="deleteDbLoading"
      @confirm="doDeleteDb"
    />

    <!-- 删除表确认 -->
    <ConfirmDialog
      v-model:show="deleteTableConfirm"
      :title="t('database.manager.deleteTableTitle')"
      :content="t('database.manager.deleteTableContent', { name: targetTableName })"
      type="error"
      :confirm-text="t('database.common.delete')"
      :loading="deleteTableLoading"
      @confirm="doDeleteTable"
    />
  </div>
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

/* 库表管理布局 */
.schema-layout {
  display: flex;
  gap: var(--space-4);
  min-height: var(--chart-height);
}

.schema-sidebar {
  width: var(--input-xl);
  flex-shrink: 0;
  background: var(--surface2);
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  padding: var(--space-3);
  overflow-y: auto;
  max-height: 70vh;
}

/* 移动端：侧栏与内容上下堆叠 */
@media (max-width: 768px) {
  .schema-layout {
    flex-direction: column;
    min-height: auto;
  }

  .schema-sidebar {
    width: 100%;
    max-height: 40vh;
  }
}

.schema-sidebar__title {
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

.schema-sidebar__empty {
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

.schema-main {
  flex: 1;
  min-width: 0;
}

.schema-main__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-3);
}

.schema-main__dbname {
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

/* 移动端适配 */
@media (max-width: 768px) {
  .schema-layout {
    flex-direction: column;
  }
  .schema-sidebar {
    width: 100%;
    max-height: 200px;
  }
  .tab-toolbar {
    flex-wrap: wrap;
    gap: var(--space-2);
  }
}
</style>
