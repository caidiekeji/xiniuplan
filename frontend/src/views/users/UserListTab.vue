<script setup lang="ts">
import { ref, reactive, computed, onMounted, h, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NButton, NIcon, NForm, NFormItem, NInput, NSelect, NSwitch,
  NSpace, NAvatar
} from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
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
const { t } = useI18n()
const authStore = useAuthStore()

/* ========== 状态映射 ========== */
const roleStatusMap = computed<Record<string, { status: StatusType; text: string }>>(() => ({
  super_admin: { status: 'warning', text: t('userList.roles.superAdmin') },
  admin: { status: 'running', text: t('userList.roles.admin') },
  user: { status: 'stopped', text: t('userList.roles.user') }
}))

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
    message.error(err.message || t('userList.loadError'))
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
    title: t('userList.columns.username'),
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
    title: t('userList.columns.role'),
    key: 'role',
    width: 130,
    render: (row: any) => {
      const cfg = roleStatusMap.value[row.role] || roleStatusMap.value.user
      return h(StatusTag, { status: cfg.status, text: cfg.text })
    }
  },
  {
    title: 'TOTP',
    key: 'totp_enabled',
    width: 110,
    render: (row: any) => h(StatusTag, {
      status: row.totp_enabled ? 'running' : 'stopped',
      text: row.totp_enabled ? t('userList.status.enabled') : t('userList.status.disabled')
    })
  },
  {
    title: t('userList.columns.status'),
    key: 'status',
    width: 100,
    render: (row: any) => h(StatusTag, {
      status: row.status === 'active' ? 'running' : 'error',
      text: row.status === 'active' ? t('userList.status.active') : t('userList.status.inactive')
    })
  },
  {
    title: t('userList.columns.lastLogin'),
    key: 'last_login_at',
    width: 170,
    render: (row: any) => h('span', {
      class: 'mono',
      style: 'color:var(--sub);font-size:var(--fs-caption)'
    }, row.last_login_at || '—')
  },
  {
    title: t('userList.columns.createdAt'),
    key: 'created_at',
    width: 170,
    render: (row: any) => h('span', {
      class: 'mono',
      style: 'color:var(--sub);font-size:var(--fs-caption)'
    }, row.created_at || '—')
  },
  {
    title: t('userList.columns.actions'),
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
            default: () => t('userList.actions.edit')
          }),
          h(NButton, {
            size: 'tiny',
            quaternary: true,
            onClick: () => openPassword(row)
          }, {
            icon: () => h(NIcon, null, { default: () => h(KeyOutline) }),
            default: () => t('userList.actions.password')
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
            default: () => row.status === 'active' ? t('userList.actions.disable') : t('userList.actions.enable')
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
            default: () => t('userList.actions.delete')
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
    message.success(t('userList.enableSuccess'))
    loadUsers()
  } catch (err: any) {
    message.error(err.message || t('userList.enableFailed'))
  } finally {
    actionLoadingId.value = null
  }
}

async function handleDisable(id: number) {
  actionLoadingId.value = id
  try {
    await disableUser(id)
    message.success(t('userList.disableSuccess'))
    loadUsers()
  } catch (err: any) {
    message.error(err.message || t('userList.disableFailed'))
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
    message: t('userList.hints.enterUsername'),
    trigger: 'blur'
  },
  password: isEdit.value
    ? {}
    : {
        required: true,
        message: t('userList.hints.enterPassword'),
        trigger: 'blur'
      },
  confirmPassword: isEdit.value
    ? {}
    : {
        required: true,
        message: t('userList.hints.confirmPassword'),
        trigger: 'blur',
        validator: (_rule: any, value: string) => {
          if (value !== form.password) {
            return new Error(t('userList.hints.passwordMismatch'))
          }
          return true
        }
      },
  role_id: {
    required: true,
    message: t('userList.hints.selectRole'),
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
    message.warning(t('userList.hints.selectRole'))
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
      message.success(t('userList.messages.updated'))
    } else {
      await createUser({
        username: form.username,
        password: form.password,
        role_id: form.role_id,
        status: form.status ? 'active' : 'disabled'
      })
      message.success(t('userList.messages.created'))
    }
    showForm.value = false
    loadUsers()
  } catch (err: any) {
    message.error(err.message || t('userList.messages.saveFailed'))
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

const pwdRules = computed(() => ({
  newPassword: {
    required: true,
    message: t('userList.hints.enterNewPassword'),
    trigger: 'blur'
  },
  confirmPassword: {
    required: true,
    message: t('userList.hints.confirmNewPassword'),
    trigger: 'blur',
    validator: (_rule: any, value: string) => {
      if (value !== pwdForm.newPassword) {
        return new Error(t('userList.hints.passwordMismatch'))
      }
      return true
    }
  }
}))

const passwordStrength = computed(() => {
  const pwd = pwdForm.newPassword
  if (!pwd) return { level: 0, text: '', color: '' }
  let score = 0
  if (pwd.length >= 8) score++
  if (/[a-z]/.test(pwd) && /[A-Z]/.test(pwd)) score++
  if (/\d/.test(pwd)) score++
  if (/[^a-zA-Z0-9]/.test(pwd)) score++
  if (score <= 1) return { level: 1, text: t('userList.strength.weak'), color: 'var(--err)' }
  if (score <= 2) return { level: 2, text: t('userList.strength.medium'), color: 'var(--warn)' }
  return { level: 3, text: t('userList.strength.strong'), color: 'var(--ok)' }
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
    message.success(t('userList.messages.passwordChanged'))
    showPwd.value = false
  } catch (err: any) {
    message.error(err.message || t('userList.messages.passwordFailed'))
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
    message.success(t('userList.messages.deleted'))
    showDelete.value = false
    loadUsers()
  } catch (err: any) {
    message.error(err.message || t('userList.messages.deleteFailed'))
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
  <PageCard :title="t('userList.title')" :subtitle="t('userList.subtitle')">
    <template #extra>
      <NButton type="primary" @click="openCreate">
        <template #icon>
          <NIcon><AddOutline /></NIcon>
        </template>
        {{ t('userList.createUser') }}
      </NButton>
    </template>

    <!-- 搜索框 -->
    <div class="user-search">
      <NInput
        v-model:value="keyword"
        :placeholder="t('userList.searchPlaceholder')"
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
    :title="isEdit ? t('userList.editUser') : t('userList.createUser')"
    :loading="formLoading"
    :width="520"
    @update:show="(v: boolean) => (showForm = v)"
    @submit="handleFormSubmit"
  >
    <NForm ref="formRef" :model="form" :rules="formRules" label-placement="top">
      <NFormItem :label="t('userList.labels.username')" path="username">
        <NInput
          v-model:value="form.username"
          :placeholder="t('userList.hints.enterUsername')"
          :disabled="isEdit"
          clearable
        />
      </NFormItem>

      <NFormItem :label="isEdit ? t('userList.labels.passwordOptional') : t('userList.labels.password')" path="password">
        <NInput
          v-model:value="form.password"
          type="password"
          :placeholder="t('userList.hints.enterPassword')"
          show-password-on="click"
        />
      </NFormItem>

      <NFormItem v-if="!isEdit" :label="t('userList.labels.confirmPassword')" path="confirmPassword">
        <NInput
          v-model:value="form.confirmPassword"
          type="password"
          :placeholder="t('userList.hints.reenterPassword')"
          show-password-on="click"
        />
      </NFormItem>

      <NFormItem :label="t('userList.labels.role')" path="role_id">
        <NSelect
          v-model:value="form.role_id"
          :options="roleOptions"
          :placeholder="t('userList.hints.selectRole')"
          clearable
        />
      </NFormItem>

      <NFormItem :label="t('userList.labels.status')" path="status">
        <NSwitch
          v-model:value="form.status"
          :checked-value="true"
          :unchecked-value="false"
        />
        <span style="margin-left: var(--space-2); color: var(--sub); font-size: var(--fs-caption)">
          {{ form.status ? t('userList.status.on') : t('userList.status.off') }}
        </span>
      </NFormItem>
    </NForm>
  </ModalForm>

  <!-- 修改密码对话框 -->
  <ModalForm
    :show="showPwd"
    :title="t('userList.pwdTitle', { username: pwdUsername })"
    :loading="pwdLoading"
    :width="480"
    @update:show="(v: boolean) => (showPwd = v)"
    @submit="handlePwdSubmit"
  >
    <NForm ref="pwdRef" :model="pwdForm" :rules="pwdRules" label-placement="top">
      <NFormItem :label="t('userList.labels.newPassword')" path="newPassword">
        <NInput
          v-model:value="pwdForm.newPassword"
          type="password"
          :placeholder="t('userList.hints.enterNewPassword')"
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
          {{ t('userList.strength.label', { level: passwordStrength.text }) }}
        </span>
      </div>

      <NFormItem :label="t('userList.labels.confirmNewPassword')" path="confirmPassword">
        <NInput
          v-model:value="pwdForm.confirmPassword"
          type="password"
          :placeholder="t('userList.hints.reenterNewPassword')"
          show-password-on="click"
        />
      </NFormItem>
    </NForm>
  </ModalForm>

  <!-- 删除确认 -->
  <ConfirmDialog
    :show="showDelete"
    :title="t('userList.delete.title')"
    :content="t('userList.delete.content', { name: deleteName })"
    :confirm-text="t('userList.actions.delete')"
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
