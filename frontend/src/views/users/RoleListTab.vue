<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { NButton, NIcon, NForm, NFormItem, NInput, NCheckbox, NCheckboxGroup, NSpace } from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import {
  AddOutline, CreateOutline, TrashOutline
} from '@vicons/ionicons5'
import {
  PageCard, DataTable, ModalForm, StatusTag, ConfirmDialog
} from '@/components'
import type { StatusType } from '@/components'
import {
  getRoles, createRole, updateRole, deleteRole, getPermissions,
  type Role, type PermissionGroup
} from '@/api/user'
import type { DataTableColumns } from 'naive-ui'

const message = useMessage()
const { t } = useI18n()

/* ========== 列表状态 ========== */
const loading = ref(false)
const roles = ref<Role[]>([])

async function loadRoles() {
  loading.value = true
  try {
    const res = await getRoles()
    if (res.data) {
      roles.value = res.data
    }
  } catch (err: any) {
    message.error(err.message || t('roleList.messages.loadError'))
  } finally {
    loading.value = false
  }
}

/* ========== 权限树 ========== */
const permissionGroups = ref<PermissionGroup[]>([])
const allPermissionCodes = computed(() => {
  const codes: string[] = []
  permissionGroups.value.forEach((g) => {
    g.permissions.forEach((p) => codes.push(p.code))
  })
  return codes
})

async function loadPermissions() {
  try {
    const res = await getPermissions()
    if (res.data) {
      permissionGroups.value = res.data
    }
  } catch {
    // 权限加载失败时使用默认分组
    permissionGroups.value = defaultPermissionGroups()
  }
}

function defaultPermissionGroups(): PermissionGroup[] {
  const modules = [
    { module: 'site', label: t('roleList.permissions.modules.site') },
    { module: 'database', label: t('roleList.permissions.modules.database') },
    { module: 'file', label: t('roleList.permissions.modules.file') },
    { module: 'terminal', label: t('roleList.permissions.modules.terminal') },
    { module: 'container', label: t('roleList.permissions.modules.container') },
    { module: 'appstore', label: t('roleList.permissions.modules.appstore') },
    { module: 'firewall', label: t('roleList.permissions.modules.firewall') },
    { module: 'backup', label: t('roleList.permissions.modules.backup') },
    { module: 'monitor', label: t('roleList.permissions.modules.monitor') },
    { module: 'user', label: t('roleList.permissions.modules.user') },
    { module: 'audit', label: t('roleList.permissions.modules.audit') },
    { module: 'setting', label: t('roleList.permissions.modules.setting') },
    { module: 'alert', label: t('roleList.permissions.modules.alert') },
    { module: 'cron', label: t('roleList.permissions.modules.cron') }
  ]
  const actions = [
    { code: 'view', label: t('roleList.permissions.actions.view') },
    { code: 'create', label: t('roleList.permissions.actions.create') },
    { code: 'update', label: t('roleList.permissions.actions.update') },
    { code: 'delete', label: t('roleList.permissions.actions.delete') }
  ]
  return modules.map((m) => ({
    module: m.module,
    label: m.label,
    permissions: actions.map((a) => ({
      code: `${m.module}:${a.code}`,
      label: a.label
    }))
  }))
}

/* ========== 表格列 ========== */
const columns = computed(() => [
  {
    title: t('roleList.columns.name'),
    key: 'name',
    width: 160,
    render: (row: any) => h('span', { style: 'font-weight:500;color:var(--ink)' }, row.name)
  },
  {
    title: t('roleList.columns.code'),
    key: 'code',
    width: 180,
    render: (row: any) => h('span', {
      class: 'mono',
      style: 'color:var(--sub);font-size:var(--fs-mono)'
    }, row.code)
  },
  {
    title: t('roleList.columns.description'),
    key: 'description',
    render: (row: any) => h('span', { style: 'color:var(--sub)' }, row.description || '—')
  },
  {
    title: t('roleList.columns.permissions'),
    key: 'permissions',
    width: 110,
    render: (row: any) => h('span', {
      class: 'mono tabular-nums',
      style: 'color:var(--ink);font-weight:600'
    }, String(row.permissions?.length || 0))
  },
  {
    title: t('roleList.columns.type'),
    key: 'is_system',
    width: 100,
    render: (row: any) => h(StatusTag, {
      status: row.is_system ? 'warning' as StatusType : 'running' as StatusType,
      text: row.is_system ? t('roleList.status.builtIn') : t('roleList.status.custom')
    })
  },
  {
    title: t('roleList.columns.actions'),
    key: 'actions',
    width: 160,
    render: (row: any) => {
      const actionLoading = actionLoadingId.value === row.id
      return h(NSpace, { size: 'small', wrap: false }, {
        default: () => [
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            onClick: () => openEdit(row)
          }, {
            icon: () => h(NIcon, null, { default: () => h(CreateOutline) }),
            default: () => t('roleList.actions.edit')
          }),
          !row.is_system
            ? h(NButton, {
                size: 'tiny',
                quaternary: true,
                type: 'error',
                loading: actionLoading,
                onClick: () => openDelete(row)
              }, {
                icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
                default: () => t('roleList.actions.delete')
              })
            : null
        ]
      })
    }
  }
] as DataTableColumns)

/* ========== 操作 loading ========== */
const actionLoadingId = ref<number | null>(null)

/* ========== 创建/编辑角色对话框 ========== */
const showForm = ref(false)
const formLoading = ref(false)
const formRef = ref()
const isEdit = ref(false)
const editId = ref<number | null>(null)

const form = reactive({
  name: '',
  code: '',
  description: '',
  permissions: [] as string[]
})

const formRules = computed(() => ({
  name: {
    required: true,
    message: t('roleList.hints.nameRequired'),
    trigger: 'blur'
  },
  code: {
    required: true,
    message: t('roleList.hints.codeRequired'),
    trigger: 'blur'
  }
}))

/* 权限矩阵全选/半选状态 */
function getModulePermissions(moduleCode: string): string[] {
  const group = permissionGroups.value.find((g) => g.module === moduleCode)
  return group ? group.permissions.map((p) => p.code) : []
}

function isModuleAllChecked(moduleCode: string): boolean {
  const codes = getModulePermissions(moduleCode)
  return codes.length > 0 && codes.every((c) => form.permissions.includes(c))
}

function isModuleIndeterminate(moduleCode: string): boolean {
  const codes = getModulePermissions(moduleCode)
  const checked = codes.filter((c) => form.permissions.includes(c))
  return checked.length > 0 && checked.length < codes.length
}

function toggleModule(moduleCode: string, checked: boolean) {
  const codes = getModulePermissions(moduleCode)
  if (checked) {
    codes.forEach((c) => {
      if (!form.permissions.includes(c)) {
        form.permissions.push(c)
      }
    })
  } else {
    form.permissions = form.permissions.filter((c) => !codes.includes(c))
  }
}

const isAllChecked = computed(() => {
  return allPermissionCodes.value.length > 0 &&
    allPermissionCodes.value.every((c) => form.permissions.includes(c))
})

const isAllIndeterminate = computed(() => {
  const checked = allPermissionCodes.value.filter((c) => form.permissions.includes(c))
  return checked.length > 0 && checked.length < allPermissionCodes.value.length
})

function toggleAll(checked: boolean) {
  form.permissions = checked ? [...allPermissionCodes.value] : []
}

function resetForm() {
  form.name = ''
  form.code = ''
  form.description = ''
  form.permissions = []
}

function openCreate() {
  isEdit.value = false
  editId.value = null
  resetForm()
  showForm.value = true
}

function openEdit(row: any) {
  isEdit.value = true
  editId.value = row.id
  resetForm()
  form.name = row.name
  form.code = row.code
  form.description = row.description || ''
  form.permissions = [...(row.permissions || [])]
  showForm.value = true
}

async function handleFormSubmit() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  formLoading.value = true
  try {
    if (isEdit.value && editId.value !== null) {
      await updateRole(editId.value, {
        name: form.name,
        description: form.description,
        permissions: form.permissions
      })
      message.success(t('roleList.messages.updated'))
    } else {
      await createRole({
        name: form.name,
        code: form.code,
        description: form.description,
        permissions: form.permissions
      })
      message.success(t('roleList.messages.created'))
    }
    showForm.value = false
    loadRoles()
  } catch (err: any) {
    message.error(err.message || t('roleList.messages.saveFailed'))
  } finally {
    formLoading.value = false
  }
}

/* ========== 删除确认 ========== */
const showDelete = ref(false)
const deleteLoading = ref(false)
const deleteId = ref<number | null>(null)
const deleteName = ref('')

function openDelete(row: any) {
  deleteId.value = row.id
  deleteName.value = row.name
  showDelete.value = true
}

async function handleDeleteConfirm() {
  if (deleteId.value === null) return
  deleteLoading.value = true
  try {
    await deleteRole(deleteId.value)
    message.success(t('roleList.messages.deleted'))
    showDelete.value = false
    loadRoles()
  } catch (err: any) {
    message.error(err.message || t('roleList.messages.deleteFailed'))
  } finally {
    deleteLoading.value = false
  }
}

onMounted(() => {
  loadPermissions()
  loadRoles()
})
</script>

<template>
  <PageCard :title="t('roleList.title')" :subtitle="t('roleList.subtitle')">
    <template #extra>
      <NButton type="primary" @click="openCreate">
        <template #icon>
          <NIcon><AddOutline /></NIcon>
        </template>
        {{ t('roleList.createRole') }}
      </NButton>
    </template>

    <!-- 数据表格 -->
    <DataTable
      :columns="columns"
      :data="roles"
      :loading="loading"
      :pagination="false"
    />
  </PageCard>

  <!-- 创建/编辑角色对话框 -->
  <ModalForm
    :show="showForm"
    :title="isEdit ? t('roleList.editRole') : t('roleList.createRole')"
    :loading="formLoading"
    :width="720"
    @update:show="(v: boolean) => (showForm = v)"
    @submit="handleFormSubmit"
  >
    <NForm ref="formRef" :model="form" :rules="formRules" label-placement="top">
      <div class="role-form__grid">
        <NFormItem :label="t('roleList.labels.name')" path="name">
          <NInput v-model:value="form.name" :placeholder="t('roleList.hints.name')" clearable />
        </NFormItem>

        <NFormItem :label="t('roleList.labels.code')" path="code">
          <NInput
            v-model:value="form.code"
            :placeholder="t('roleList.hints.code')"
            :disabled="isEdit"
            clearable
          />
        </NFormItem>
      </div>

      <NFormItem :label="t('roleList.labels.description')" path="description">
        <NInput v-model:value="form.description" :placeholder="t('roleList.hints.description')" clearable />
      </NFormItem>

      <!-- 权限矩阵 -->
      <NFormItem :label="t('roleList.labels.permissions')">
        <div class="perm-matrix">
          <!-- 全选 -->
          <div class="perm-matrix__header">
            <NCheckbox
              :checked="isAllChecked"
              :indeterminate="isAllIndeterminate"
              @update:checked="toggleAll"
            >
              {{ t('roleList.matrix.selectAll') }}
            </NCheckbox>
            <span class="perm-matrix__count mono tabular-nums">
              {{ t('roleList.matrix.selected', { selected: form.permissions.length, total: allPermissionCodes.length }) }}
            </span>
          </div>

          <!-- 按模块分组 -->
          <div class="perm-matrix__groups">
            <div
              v-for="group in permissionGroups"
              :key="group.module"
              class="perm-group"
            >
              <div class="perm-group__header">
                <NCheckbox
                  :checked="isModuleAllChecked(group.module)"
                  :indeterminate="isModuleIndeterminate(group.module)"
                  @update:checked="(v: boolean) => toggleModule(group.module, v)"
                >
                  <span class="perm-group__title">{{ group.label }}</span>
                </NCheckbox>
              </div>
              <NCheckboxGroup v-model:value="form.permissions" class="perm-group__items">
                <NCheckbox
                  v-for="perm in group.permissions"
                  :key="perm.code"
                  :value="perm.code"
                >
                  {{ perm.label }}
                </NCheckbox>
              </NCheckboxGroup>
            </div>
          </div>
        </div>
      </NFormItem>
    </NForm>
  </ModalForm>

  <!-- 删除确认 -->
  <ConfirmDialog
    :show="showDelete"
    :title="t('roleList.delete.title')"
    :content="t('roleList.delete.content', { name: deleteName })"
    :confirm-text="t('roleList.actions.delete')"
    type="error"
    :loading="deleteLoading"
    @update:show="(v: boolean) => (showDelete = v)"
    @confirm="handleDeleteConfirm"
  />
</template>

<style scoped>
.role-form__grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-4);
}

.perm-matrix {
  width: 100%;
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.perm-matrix__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-3) var(--space-4);
  background: var(--surface2);
  border-bottom: 1px solid var(--line);
}

.perm-matrix__count {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
}

.perm-matrix__groups {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0;
}

.perm-group {
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--line);
  border-right: 1px solid var(--line);
}

.perm-group:nth-child(2n) {
  border-right: none;
}

.perm-group__header {
  margin-bottom: var(--space-2);
}

.perm-group__title {
  font-weight: 600;
  color: var(--ink);
  font-size: var(--fs-body);
}

.perm-group__items {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  padding-left: var(--space-5);
}

@media (max-width: 640px) {
  .perm-matrix__groups {
    grid-template-columns: 1fr;
  }
  .perm-group {
    border-right: none;
  }
  .role-form__grid {
    grid-template-columns: 1fr;
  }
}

/* 移动端：权限分组单列 */
@media (max-width: 640px) {
  .perm-matrix__groups {
    grid-template-columns: 1fr;
  }
}
</style>
