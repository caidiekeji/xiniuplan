<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useMessage } from 'naive-ui'
import {
  NButton, NSpace, NIcon, NSelect, NInput, NFormItem, NEmpty,
  NCheckbox, NAlert, NTag, NUpload
} from 'naive-ui'
import { PlayOutline, CloudUploadOutline } from '@vicons/ionicons5'
import { DataTable, ModalForm, CodeEditor } from '@/components'
import * as dbApi from '@/api/database'
import type { DatabaseInfo, SQLResult } from '@/api/database'
import { rowsToObjects, buildDynamicColumns } from './shared'

const message = useMessage()

/* ==================== Props / Emits ==================== */

const props = defineProps<{
  instanceId: number | null
  instanceOptions: { label: string; value: number }[]
  databases: DatabaseInfo[]
}>()

const emit = defineEmits<{
  (e: 'update:instanceId', val: number | null): void
}>()

/* ==================== SQL 执行器 ==================== */

const sqlText = ref('-- 在此输入 SQL 语句\nSELECT 1;')
const sqlLoading = ref(false)
const allowDestructive = ref(false)
const sqlSelectedDb = ref('')
const sqlResult = ref<SQLResult | null>(null)
const sqlResultRows = ref<Record<string, any>[]>([])
const sqlResultColumns = ref<string[]>([])

const dbOptionsForSql = computed(() =>
  props.databases.map((d) => ({ label: d.name, value: d.name }))
)

watch(
  () => props.instanceId,
  () => {
    sqlSelectedDb.value = ''
    sqlResult.value = null
  }
)

watch(
  () => props.databases,
  (dbs) => {
    if (dbs.length > 0 && !sqlSelectedDb.value) {
      sqlSelectedDb.value = dbs[0].name
    }
  }
)

async function executeSql() {
  if (!props.instanceId) {
    message.warning('请先选择实例')
    return
  }
  if (!sqlSelectedDb.value) {
    message.warning('请选择数据库')
    return
  }
  if (!sqlText.value.trim()) {
    message.warning('请输入 SQL 语句')
    return
  }
  sqlLoading.value = true
  sqlResult.value = null
  try {
    const res = await dbApi.executeSQL(props.instanceId, {
      db_name: sqlSelectedDb.value,
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

/* ==================== 导入 SQL Modal ==================== */

const importModal = ref(false)
const importLoading = ref(false)
const importDbName = ref('')
const importContent = ref('')

function openImport() {
  importDbName.value = sqlSelectedDb.value || ''
  importContent.value = ''
  importModal.value = true
}

function handleImportFile(file: File) {
  const reader = new FileReader()
  reader.onload = (e) => {
    importContent.value = e.target?.result as string
    message.success(`已加载文件: ${file.name}`)
  }
  reader.readAsText(file)
  return false
}

async function submitImport() {
  if (!props.instanceId) return
  if (!importDbName.value) {
    message.warning('请选择数据库')
    return
  }
  if (!importContent.value.trim()) {
    message.warning('请粘贴 SQL 内容或上传文件')
    return
  }
  importLoading.value = true
  try {
    const res = await dbApi.importSQL(props.instanceId, {
      db_name: importDbName.value,
      content: importContent.value
    })
    message.success(`导入完成，影响 ${res.data.affected_rows} 行，耗时 ${res.data.duration_ms}ms`)
    importModal.value = false
  } catch (e: any) {
    message.error(e.message || '导入失败')
  } finally {
    importLoading.value = false
  }
}
</script>

<template>
  <div class="sql-executor-tab">
    <div class="tab-toolbar">
      <NSpace>
        <NSelect
          :value="instanceId"
          placeholder="选择 MySQL 实例"
          :options="instanceOptions"
          style="width: var(--input-2xl)"
          @update:value="(v: number | null) => $emit('update:instanceId', v)"
        />
        <NSelect
          v-model:value="sqlSelectedDb"
          placeholder="选择数据库"
          :options="dbOptionsForSql"
          style="width: var(--input-lg)"
          :disabled="!instanceId"
        />
        <NButton
          type="primary"
          :loading="sqlLoading"
          :disabled="!instanceId || !sqlSelectedDb"
          @click="executeSql"
        >
          <template #icon><NIcon><PlayOutline /></NIcon></template>执行
        </NButton>
        <NButton :disabled="!instanceId" @click="openImport">
          <template #icon><NIcon><CloudUploadOutline /></NIcon></template>导入
        </NButton>
        <NCheckbox v-model:checked="allowDestructive">
          <span class="checkbox-label">允许危险操作</span>
        </NCheckbox>
      </NSpace>
    </div>

    <div v-if="!instanceId" class="empty-hint">
      <NEmpty description="请先选择实例和数据库" />
    </div>

    <template v-else>
      <div class="sql-editor-wrap">
        <CodeEditor v-model="sqlText" language="sql" height="300px" :minimap="false" />
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

    <!-- 导入 SQL -->
    <ModalForm
      v-model:show="importModal"
      title="导入 SQL"
      :loading="importLoading"
      :width="640"
      @submit="submitImport"
    >
      <NFormItem label="目标数据库">
        <NSelect v-model:value="importDbName" :options="dbOptionsForSql" placeholder="选择数据库" />
      </NFormItem>
      <NFormItem label="上传 SQL 文件">
        <NUpload
          :custom-request="({ file }) => handleImportFile(file.file as File)"
          :show-file-list="false"
          accept=".sql"
        >
          <NButton>
            <template #icon><NIcon><CloudUploadOutline /></NIcon></template>选择 .sql 文件
          </NButton>
        </NUpload>
      </NFormItem>
      <NFormItem label="或粘贴 SQL 内容">
        <NInput
          v-model:value="importContent"
          type="textarea"
          :rows="8"
          placeholder="粘贴 SQL 语句..."
          style="font-family: var(--font-mono)"
        />
      </NFormItem>
    </ModalForm>
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

.checkbox-label {
  font-size: var(--fs-body);
  color: var(--sub);
}

.sql-editor-wrap {
  margin-bottom: var(--space-4);
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
</style>
