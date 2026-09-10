<script setup lang="ts">
import { ref, reactive, computed, onMounted, h, watch } from 'vue'
import {
  NButton, NIcon, NForm, NFormItem, NInput, NSelect, NSwitch,
  NSpace, NAvatar, useMessage
} from 'naive-ui'
import {
  AddOutline, CreateOutline, KeyOutline, TrashOutline,
  PlayOutline, StopOutline, SearchOutline
} from '@vicons/ionicons5'
import {
  PageCard, DataTable, ModalForm, StatusTag, ConfirmDialog
} from '@/components'
import type { StatusType } from '@/components'
import {
  getUsers, createUser, updateUser, deleteUser,
  updatePassword, enableUser, disableUser, getRoles,
  type User, type Role
} from '@/api/user'
import { useAuthStore } from '@/stores/auth'
import type { DataTableColumns } from 'naive-ui'

const message = useMessage()
const authStore = useAuthStore()

/* ========== 状态映射 ========== */
const roleStatusMap: Record<string, { status: StatusType; text: string }> = {
  super_admin: { status: 'warning', text: '超级管理员' },
  admin: { status: 'running', text: '管理员' },
  user: { status: 'stopped', text: '普通用户' }
}

/* ========== 列表状态 ========== */
const loading = ref(false)
const users = ref<User[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
let searchTimer: ReturnType<typeof setTimeout> | null = null

watch(keyword, (val) => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    loadUsers()
  }, 300)
})

async function loadUsers() {
  loading.value = true
  try {
    const res = await getUsers({
      page: page.value,
      pageSize: pageSize.value,
      keyword: keyword.value || undefined
    })
    if (res.data?.items) {
      users.value = res.data.items
      total.value = res.data.total
    } else {
      users.value = []
      total.value = 0
    }
  } catch (err: any) {
    message.error(err.message || '加载用户列表失败')
  } finally {
    loading.value = false
  }
}

const pagination = computed(() => ({
  page: page.value,
  pageSize: pageSize.value,
  itemCount: total.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  'onUpdate:page': (p: number) => { page.value = p; loadUsers() },
  'onUpdate:page-size': (ps: number) => { pageSize.value = ps; page.value = 1; loadUsers() }
}))

/* ========== 角色列表 ========== */
const roles = ref<Role[]>([])
const roleOptions = computed(() =>
  roles.value.map((r) => ({ label: r.name, value: r.id }))
)

async function loadRoles() {
  try {
    const res = await getRoles()
    if (res.data) {
      roles.value = res.data
    }
  } catch {
    // 角色加载失败不阻塞
  }
}

/* ========== 表格列 ========== */
const columns = computed(() => [
  {
    title: '用户名',
    key: 'username',
    width: 200,
    render: (row: any) => h('div', { class: 'user-cell' }, [
      h(NAvatar, {
        size: 'small',
        round: true,
        style: 'background:var(--acc);color:var(--on-color);font-weight:600;'
      }, { default: () => row.username.charAt(0).toUpperCase() }),
      h('span', { class: 'user-cell__name' }, row.username)
    ])
  },
  {
    title: '角色',
    key: 'role',
    width: 130,
    render: (row: any) => {
      const cfg = roleStatusMap[row.role] || roleStatusMap.user
      return h(StatusTag, { status: cfg.status, text: cfg.text })
    }
  },
  {
    title: 'TOTP',
    key: 'totp_enabled',
    width: 110,
    render: (row: any) => h(StatusTag, {
      status: row.totp_enabled ? 'running' : 'stopped',
      text: row.totp_enabled ? '已启用' : '未启用'
    })
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row: any) => h(StatusTag, {
      status: row.status === 'active' ? 'running' : 'error',
      text: row.status === 'active' ? '正常' : '已禁用'
    })
  },
  {
    title: '最后登录',
    key: 'last_login_at',
    width: 170,
    render: (row: any) => h('span', {
      class: 'mono',
      style: 'color:var(--sub);font-size:var(--fs-caption)'
    }, row.last_login_at || '—')
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 170,
    render: (row: any) => h('span', {
      class: 'mono',
      style: 'color:var(--sub);font-size:var(--fs-caption)'
    }, row.created_at || '—')
  },
  {
    title: '操作',
    key: 'actions',
    width: 300,
    render: (row: any) => {
      const isSelf = authStore.userInfo?.id === row.id
      const isSuperAdmin = row.role === 'super_admin'
      const actionLoading = actionLoadingId.value === row.id
      const canDelete = !isSelf && !isSuperAdmin
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
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            onClick: () => openPassword(row)
          }, {
            icon: () => h(NIcon, null, { default: () => h(KeyOutline) }),
            default: () => '改密'
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            type: row.status === 'active' ? 'warning' : 'success',
            loading: actionLoading,
            disabled: isSelf,
            onClick: () => row.status === 'active' ? handleDisable(row.id) : handleEnable(row.id)
          }, {
            icon: () => h(NIcon, null, { default: () => h(row.status === 'active' ? StopOutline : PlayOutline) }),
            default: () => row.status === 'active' ? '禁用' : '启用'
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            type: 'error',
            loading: actionLoading,
            disabled: !canDelete,
            onClick: () => openDelete(row)
          }, {
            icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
            default: () => '删除'
          })
        ]
      })
    }
  }
] as DataTableColumns)

/* ========== 操作 loading ========== */
const actionLoadingId = ref<number | null>(null)

async function handleEnable(id: number) {
  actionLoadingId.value = id
  try {
    await enableUser(id)
    message.success('用户已启用')
    loadUsers()
  } catch (err: any) {
    message.error(err.message || '启用失败')
  } finally {
    actionLoadingId.value = null
  }
}

async function handleDisable(id: number) {
  actionLoadingId.value = id
  try {
    await disableUser(id)
    message.success('用户已禁用')
    loadUsers()
  } catch (err: any) {
    message.error(err.message || '禁用失败')
  } finally {
    actionLoadingId.value = null
  }
}

/* ========== 创建/编辑用户对话框 ========== */
const showForm = ref(false)
const formLoading = ref(false)
const formRef = ref()
const isEdit = ref(false)
const editId = ref<number | null>(null)

const form = reactive({
  username: '',
  password: '',
  confirmPassword: '',
  role_id: null as number | null,
  status: true as boolean
})

const formRules = computed(() => ({
  username: {
    required: true,
    message: '请输入用户名',
    trigger: 'blur'
  },
  password: isEdit.value
    ? {}
    : {
        required: true,
        message: '请输入密码',
        trigger: 'blur'
      },
  confirmPassword: isEdit.value
    ? {}
    : {
        required: true,
        message: '请确认密码',
        trigger: 'blur',
        validator: (_rule: any, value: string) => {
          if (value !== form.password) {
            return new Error('两次输入的密码不一致')
          }
          return true
        }
      },
  role_id: {
    required: true,
    message: '请选择角色',
    trigger: 'change'
  }
}))

function resetForm() {
  form.username = ''
  form.password = ''
  form.confirmPassword = ''
  form.role_id = null
  form.status = true
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
  form.username = row.username
  form.role_id = row.role_id
  form.status = row.status === 'active'
  showForm.value = true
}

async function handleFormSubmit() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  if (form.role_id === null) {
    message.warning('请选择角色')
    return
  }

  formLoading.value = true
  try {
    if (isEdit.value && editId.value !== null) {
      const data: any = {
        role_id: form.role_id,
        status: form.status ? 'active' : 'disabled'
      }
      if (form.password) {
        data.password = form.password
      }
      await updateUser(editId.value, data)
      message.success('用户已更新')
    } else {
      await createUser({
        username: form.username,
        password: form.password,
        role_id: form.role_id,
        status: form.status ? 'active' : 'disabled'
      })
      message.success('用户已创建')
    }
    showForm.value = false
    loadUsers()
  } catch (err: any) {
    message.error(err.message || '保存失败')
  } finally {
    formLoading.value = false
  }
}

/* ========== 修改密码对话框 ========== */
const showPwd = ref(false)
const pwdLoading = ref(false)
const pwdRef = ref()
const pwdUserId = ref<number | null>(null)
const pwdUsername = ref('')

const pwdForm = reactive({
  newPassword: '',
  confirmPassword: ''
})

const pwdRules = {
  newPassword: {
    required: true,
    message: '请输入新密码',
    trigger: 'blur'
  },
  confirmPassword: {
    required: true,
    message: '请确认新密码',
    trigger: 'blur',
    validator: (_rule: any, value: string) => {
      if (value !== pwdForm.newPassword) {
        return new Error('两次输入的密码不一致')
      }
      return true
    }
  }
}

const passwordStrength = computed(() => {
  const pwd = pwdForm.newPassword
  if (!pwd) return { level: 0, text: '', color: '' }
  let score = 0
  if (pwd.length >= 8) score++
  if (/[a-z]/.test(pwd) && /[A-Z]/.test(pwd)) score++
  if (/\d/.test(pwd)) score++
  if (/[^a-zA-Z0-9]/.test(pwd)) score++
  if (score <= 1) return { level: 1, text: '弱', color: 'var(--err)' }
  if (score <= 2) return { level: 2, text: '中', color: 'var(--warn)' }
  return { level: 3, text: '强', color: 'var(--ok)' }
})

function openPassword(row: any) {
  pwdUserId.value = row.id
  pwdUsername.value = row.username
  pwdForm.newPassword = ''
  pwdForm.confirmPassword = ''
  showPwd.value = true
}

async function handlePwdSubmit() {
  try {
    await pwdRef.value?.validate()
  } catch {
    return
  }
  if (pwdUserId.value === null) return

  pwdLoading.value = true
  try {
    await updatePassword(pwdUserId.value, {
      new_password: pwdForm.newPassword
    })
    message.success('密码已修改')
    showPwd.value = false
  } catch (err: any) {
    message.error(err.message || '修改密码失败')
  } finally {
    pwdLoading.value = false
  }
}

/* ========== 删除确认 ========== */
const showDelete = ref(false)
const deleteLoading = ref(false)
const deleteId = ref<number | null>(null)
const deleteName = ref('')

function openDelete(row: any) {
  deleteId.value = row.id
  deleteName.value = row.username
  showDelete.value = true
}

async function handleDeleteConfirm() {
  if (deleteId.value === null) return
  deleteLoading.value = true
  try {
    await deleteUser(deleteId.value)
    message.success('用户已删除')
    showDelete.value = false
    loadUsers()
  } catch (err: any) {
    message.error(err.message || '删除失败')
  } finally {
    deleteLoading.value = false
  }
}

onMounted(() => {
  loadRoles()
  loadUsers()
})
</script>

<template>
  <PageCard title="用户列表" subtitle="管理系统用户、角色分配与账户状态">
    <template #extra>
      <NButton type="primary" @click="openCreate">
        <template #icon>
          <NIcon><AddOutline /></NIcon>
        </template>
        创建用户
      </NButton>
    </template>

    <!-- 搜索框 -->
    <div class="user-search">
      <NInput
        v-model:value="keyword"
        placeholder="搜索用户名..."
        clearable
        style="width: var(--input-2xl)"
      >
        <template #prefix>
          <NIcon><SearchOutline /></NIcon>
        </template>
      </NInput>
    </div>

    <!-- 数据表格 -->
    <DataTable
      :columns="columns"
      :data="users"
      :loading="loading"
      :pagination="pagination"
    />
  </PageCard>

  <!-- 创建/编辑用户对话框 -->
  <ModalForm
    :show="showForm"
    :title="isEdit ? '编辑用户' : '创建用户'"
    :loading="formLoading"
    :width="520"
    @update:show="(v: boolean) => (showForm = v)"
    @submit="handleFormSubmit"
  >
    <NForm ref="formRef" :model="form" :rules="formRules" label-placement="top">
      <NFormItem label="用户名" path="username">
        <NInput
          v-model:value="form.username"
          placeholder="请输入用户名"
          :disabled="isEdit"
          clearable
        />
      </NFormItem>

      <NFormItem :label="isEdit ? '密码（留空不修改）' : '密码'" path="password">
        <NInput
          v-model:value="form.password"
          type="password"
          placeholder="请输入密码"
          show-password-on="click"
        />
      </NFormItem>

      <NFormItem v-if="!isEdit" label="确认密码" path="confirmPassword">
        <NInput
          v-model:value="form.confirmPassword"
          type="password"
          placeholder="请再次输入密码"
          show-password-on="click"
        />
      </NFormItem>

      <NFormItem label="角色" path="role_id">
        <NSelect
          v-model:value="form.role_id"
          :options="roleOptions"
          placeholder="请选择角色"
          clearable
        />
      </NFormItem>

      <NFormItem label="状态" path="status">
        <NSwitch
          v-model:value="form.status"
          :checked-value="true"
          :unchecked-value="false"
        />
        <span style="margin-left: var(--space-2); color: var(--sub); font-size: var(--fs-caption)">
          {{ form.status ? '启用' : '禁用' }}
        </span>
      </NFormItem>
    </NForm>
  </ModalForm>

  <!-- 修改密码对话框 -->
  <ModalForm
    :show="showPwd"
    :title="`修改密码 — ${pwdUsername}`"
    :loading="pwdLoading"
    :width="480"
    @update:show="(v: boolean) => (showPwd = v)"
    @submit="handlePwdSubmit"
  >
    <NForm ref="pwdRef" :model="pwdForm" :rules="pwdRules" label-placement="top">
      <NFormItem label="新密码" path="newPassword">
        <NInput
          v-model:value="pwdForm.newPassword"
          type="password"
          placeholder="请输入新密码"
          show-password-on="click"
        />
      </NFormItem>

      <!-- 密码强度提示 -->
      <div v-if="pwdForm.newPassword" class="pwd-strength">
        <div class="pwd-strength__bar">
          <span
            v-for="i in 3"
            :key="i"
            class="pwd-strength__seg"
            :style="{ background: i <= passwordStrength.level ? passwordStrength.color : 'var(--surface3)' }"
          />
        </div>
        <span class="pwd-strength__text" :style="{ color: passwordStrength.color }">
          密码强度：{{ passwordStrength.text }}
        </span>
      </div>

      <NFormItem label="确认新密码" path="confirmPassword">
        <NInput
          v-model:value="pwdForm.confirmPassword"
          type="password"
          placeholder="请再次输入新密码"
          show-password-on="click"
        />
      </NFormItem>
    </NForm>
  </ModalForm>

  <!-- 删除确认 -->
  <ConfirmDialog
    :show="showDelete"
    title="删除用户"
    :content="`确定要删除用户「${deleteName}」吗？此操作不可撤销。`"
    confirm-text="删除"
    type="error"
    :loading="deleteLoading"
    @update:show="(v: boolean) => (showDelete = v)"
    @confirm="handleDeleteConfirm"
  />
</template>

<style scoped>
.user-search {
  margin-bottom: var(--space-4);
}

.user-cell {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.user-cell__name {
  font-weight: 500;
  color: var(--ink);
}

.pwd-strength {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
  margin-top: calc(var(--space-2) * -1);
}

.pwd-strength__bar {
  display: flex;
  gap: var(--space-1);
}

.pwd-strength__seg {
  width: var(--size-sm);
  height: 4px;
  border-radius: 2px;
  transition: background var(--t-fast) var(--ease);
}

.pwd-strength__text {
  font-size: var(--fs-caption);
  font-weight: 500;
}
</style>
