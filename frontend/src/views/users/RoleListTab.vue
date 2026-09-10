<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue'
import {
  NButton, NIcon, NForm, NFormItem, NInput, NCheckbox,
  NCheckboxGroup, NSpace, NTag, useMessage
} from 'naive-ui'
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
    message.error(err.message || '加载角色列表失败')
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
    { module: 'site', label: '网站' },
    { module: 'database', label: '数据库' },
    { module: 'file', label: '文件' },
    { module: 'terminal', label: '终端' },
    { module: 'container', label: '容器' },
    { module: 'appstore', label: '应用商店' },
    { module: 'firewall', label: '防火墙' },
    { module: 'backup', label: '备份' },
    { module: 'monitor', label: '监控' },
    { module: 'user', label: '用户' },
    { module: 'audit', label: '审计' },
    { module: 'setting', label: '设置' },
    { module: 'alert', label: '告警' },
    { module: 'cron', label: '计划任务' }
  ]
  const actions = [
    { code: 'view', label: '查看' },
    { code: 'create', label: '创建' },
    { code: 'update', label: '修改' },
    { code: 'delete', label: '删除' }
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
    title: '角色名称',
    key: 'name',
    width: 160,
    render: (row: any) => h('span', { style: 'font-weight:500;color:var(--ink)' }, row.name)
  },
  {
    title: '角色编码',
    key: 'code',
    width: 180,
    render: (row: any) => h('span', {
      class: 'mono',
      style: 'color:var(--sub);font-size:var(--fs-mono)'
    }, row.code)
  },
  {
    title: '描述',
    key: 'description',
    render: (row: any) => h('span', { style: 'color:var(--sub)' }, row.description || '—')
  },
  {
    title: '权限数量',
    key: 'permissions',
    width: 110,
    render: (row: any) => h('span', {
      class: 'mono tabular-nums',
      style: 'color:var(--ink);font-weight:600'
    }, String(row.permissions?.length || 0))
  },
  {
    title: '类型',
    key: 'is_system',
    width: 100,
    render: (row: any) => h(StatusTag, {
      status: row.is_system ? 'warning' as StatusType : 'running' as StatusType,
      text: row.is_system ? '内置' : '自定义'
    })
  },
  {
    title: '操作',
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
            default: () => '编辑'
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
                default: () => '删除'
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

const formRules = {
  name: {
    required: true,
    message: '请输入角色名称',
    trigger: 'blur'
  },
  code: {
    required: true,
    message: '请输入角色编码',
    trigger: 'blur'
  }
}

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
      message.success('角色已更新')
    } else {
      await createRole({
        name: form.name,
        code: form.code,
        description: form.description,
        permissions: form.permissions
      })
      message.success('角色已创建')
    }
    showForm.value = false
    loadRoles()
  } catch (err: any) {
    message.error(err.message || '保存失败')
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
    message.success('角色已删除')
    showDelete.value = false
    loadRoles()
  } catch (err: any) {
    message.error(err.message || '删除失败')
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
  <PageCard title="角色权限" subtitle="管理系统角色及其权限分配">
    <template #extra>
      <NButton type="primary" @click="openCreate">
        <template #icon>
          <NIcon><AddOutline /></NIcon>
        </template>
        创建角色
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
    :title="isEdit ? '编辑角色' : '创建角色'"
    :loading="formLoading"
    :width="720"
    @update:show="(v: boolean) => (showForm = v)"
    @submit="handleFormSubmit"
  >
    <NForm ref="formRef" :model="form" :rules="formRules" label-placement="top">
      <div class="role-form__grid">
        <NFormItem label="角色名称" path="name">
          <NInput v-model:value="form.name" placeholder="例如：运维工程师" clearable />
        </NFormItem>

        <NFormItem label="角色编码" path="code">
          <NInput
            v-model:value="form.code"
            placeholder="例如：ops_engineer"
            :disabled="isEdit"
            clearable
          />
        </NFormItem>
      </div>

      <NFormItem label="描述" path="description">
        <NInput v-model:value="form.description" placeholder="可选描述信息" clearable />
      </NFormItem>

      <!-- 权限矩阵 -->
      <NFormItem label="权限分配">
        <div class="perm-matrix">
          <!-- 全选 -->
          <div class="perm-matrix__header">
            <NCheckbox
              :checked="isAllChecked"
              :indeterminate="isAllIndeterminate"
              @update:checked="toggleAll"
            >
              全选
            </NCheckbox>
            <span class="perm-matrix__count mono tabular-nums">
              已选 {{ form.permissions.length }} / {{ allPermissionCodes.length }}
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
    title="删除角色"
    :content="`确定要删除角色「${deleteName}」吗？此操作不可撤销。`"
    confirm-text="删除"
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
  color: var(--faint);
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
