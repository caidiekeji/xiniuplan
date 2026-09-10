<script setup lang="ts">
import { ref, computed, h } from 'vue'
import { useMessage } from 'naive-ui'
import {
  NButton, NSpace, NIcon, NSelect, NInput, NFormItem, NEmpty,
  NCheckboxGroup, NCheckbox
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { AddOutline, TrashOutline, ConstructOutline, ShieldHalfOutline } from '@vicons/ionicons5'
import { DataTable, ModalForm, ConfirmDialog } from '@/components'
import * as dbApi from '@/api/database'
import type { DBUser, DatabaseInfo, CreateUserData, GrantUserData } from '@/api/database'
import { monoEl, iconBtn, btnGroup } from './shared'

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

/* ==================== 用户列表 ==================== */

const users = ref<DBUser[]>([])
const usersLoading = ref(false)

async function loadUsers() {
  if (!props.instanceId) return
  usersLoading.value = true
  try {
    const res = await dbApi.getUsers(props.instanceId)
    users.value = res.data
  } catch (e: any) {
    message.error(e.message || '加载用户列表失败')
  } finally {
    usersLoading.value = false
  }
}

const userColumns: DataTableColumns = [
  { title: '用户名', key: 'username', width: 200, render: (row: any) => monoEl(row.username) },
  { title: '主机', key: 'host', width: 200, render: (row: any) => monoEl(row.host) },
  {
    title: '操作', key: 'actions', width: 200,
    render: (row: any) =>
      btnGroup([
        iconBtn({ text: '授权', icon: ConstructOutline, type: 'primary', onClick: () => openGrantUser(row) }),
        iconBtn({ text: '回收权限', icon: ShieldHalfOutline, type: 'warning', onClick: () => openRevokeUser(row) }),
        iconBtn({ text: '删除', icon: TrashOutline, type: 'error', onClick: () => confirmDeleteUser(row) })
      ])
  }
]

/* ==================== 创建用户 Modal ==================== */

const createUserModal = ref(false)
const createUserLoading = ref(false)
const userForm = ref<CreateUserData>({ username: '', password: '', host: '%' })

function openCreateUser() {
  userForm.value = { username: '', password: '', host: '%' }
  createUserModal.value = true
}

defineExpose({ openCreateUser })

async function submitCreateUser() {
  if (!props.instanceId) return
  if (!userForm.value.username) {
    message.warning('请输入用户名')
    return
  }
  if (!userForm.value.password) {
    message.warning('请输入密码')
    return
  }
  createUserLoading.value = true
  try {
    await dbApi.createUser(props.instanceId, userForm.value)
    message.success('用户创建成功')
    createUserModal.value = false
    loadUsers()
  } catch (e: any) {
    message.error(e.message || '创建失败')
  } finally {
    createUserLoading.value = false
  }
}

/* ==================== 删除用户确认 ==================== */

const deleteUserConfirm = ref(false)
const deleteUserLoading = ref(false)
const targetUser = ref<DBUser | null>(null)

function confirmDeleteUser(row: any) {
  targetUser.value = row
  deleteUserConfirm.value = true
}

async function doDeleteUser() {
  if (!props.instanceId || !targetUser.value) return
  deleteUserLoading.value = true
  try {
    await dbApi.deleteUser(props.instanceId, targetUser.value.username, targetUser.value.host)
    message.success(`用户 ${targetUser.value.username} 已删除`)
    deleteUserConfirm.value = false
    loadUsers()
  } catch (e: any) {
    message.error(e.message || '删除失败')
  } finally {
    deleteUserLoading.value = false
  }
}

/* ==================== 授权 Modal ==================== */

const grantUserModal = ref(false)
const grantUserLoading = ref(false)
const grantTargetUsername = ref('')
const grantForm = ref<GrantUserData>({ host: '%', db_name: '', privileges: ['SELECT'] })

const dbOptionsForGrant = computed(() =>
  props.databases.map((d) => ({ label: d.name, value: d.name }))
)

function openGrantUser(row: any) {
  grantTargetUsername.value = row.username
  grantForm.value = { host: row.host, db_name: '', privileges: ['SELECT'] }
  grantUserModal.value = true
}

async function submitGrantUser() {
  if (!props.instanceId || !grantTargetUsername.value) return
  if (!grantForm.value.db_name) {
    message.warning('请选择数据库')
    return
  }
  grantUserLoading.value = true
  try {
    await dbApi.grantUser(props.instanceId, grantTargetUsername.value, grantForm.value)
    message.success('授权成功')
    grantUserModal.value = false
  } catch (e: any) {
    message.error(e.message || '授权失败')
  } finally {
    grantUserLoading.value = false
  }
}

/* ==================== 回收权限 Modal ==================== */

const revokeUserModal = ref(false)
const revokeUserLoading = ref(false)
const revokeForm = ref<{ host: string; db_name: string }>({ host: '%', db_name: '' })

function openRevokeUser(row: any) {
  revokeForm.value = { host: row.host, db_name: '' }
  grantTargetUsername.value = row.username
  revokeUserModal.value = true
}

async function submitRevokeUser() {
  if (!props.instanceId || !grantTargetUsername.value) return
  if (!revokeForm.value.db_name) {
    message.warning('请选择要回收权限的数据库')
    return
  }
  revokeUserLoading.value = true
  try {
    await dbApi.revokeUser(props.instanceId, grantTargetUsername.value, {
      host: revokeForm.value.host,
      db_name: revokeForm.value.db_name
    })
    message.success(`已回收 ${grantTargetUsername.value}@${revokeForm.value.host} 在 ${revokeForm.value.db_name} 上的权限`)
    revokeUserModal.value = false
  } catch (e: any) {
    message.error(e.message || '回收权限失败')
  } finally {
    revokeUserLoading.value = false
  }
}

/* ==================== 实例切换 ==================== */

function onInstanceChange(val: number | null) {
  emit('update:instanceId', val)
  users.value = []
}
</script>

<template>
  <div class="user-manager-tab">
    <div class="tab-toolbar">
      <NSpace>
        <NSelect
          :value="instanceId"
          placeholder="选择 MySQL 实例"
          :options="instanceOptions"
          style="width: var(--input-3xl)"
          @update:value="onInstanceChange"
        />
        <NButton type="primary" :disabled="!instanceId" @click="openCreateUser">
          <template #icon><NIcon><AddOutline /></NIcon></template>创建用户
        </NButton>
      </NSpace>
    </div>

    <div v-if="!instanceId" class="empty-hint">
      <NEmpty description="请先选择一个数据库实例" />
    </div>

    <template v-else>
      <DataTable
        :columns="userColumns"
        :data="users"
        :loading="usersLoading"
        :pagination="false"
      />
      <div v-if="users.length === 0 && !usersLoading" class="empty-hint">
        <NEmpty description="暂无用户，点击上方按钮创建" />
      </div>
    </template>

    <!-- 创建用户 -->
    <ModalForm
      v-model:show="createUserModal"
      title="创建数据库用户"
      :loading="createUserLoading"
      @submit="submitCreateUser"
    >
      <NFormItem label="用户名">
        <NInput v-model:value="userForm.username" placeholder="例如: app_user" />
      </NFormItem>
      <NFormItem label="密码">
        <NInput v-model:value="userForm.password" type="password" show-password-on="click" placeholder="请输入密码" />
      </NFormItem>
      <NFormItem label="主机">
        <NInput v-model:value="userForm.host" placeholder="%" />
        <div class="form-hint">% 表示允许任意主机连接</div>
      </NFormItem>
    </ModalForm>

    <!-- 授权 -->
    <ModalForm
      v-model:show="grantUserModal"
      :title="`授权用户: ${grantTargetUsername}`"
      :loading="grantUserLoading"
      @submit="submitGrantUser"
    >
      <NFormItem label="目标主机">
        <NInput v-model:value="grantForm.host" />
      </NFormItem>
      <NFormItem label="选择数据库">
        <NSelect v-model:value="grantForm.db_name" :options="dbOptionsForGrant" placeholder="选择数据库" />
      </NFormItem>
      <NFormItem label="权限">
        <NCheckboxGroup v-model:value="grantForm.privileges">
          <NSpace>
            <NCheckbox value="ALL">ALL</NCheckbox>
            <NCheckbox value="SELECT">SELECT</NCheckbox>
            <NCheckbox value="INSERT">INSERT</NCheckbox>
            <NCheckbox value="UPDATE">UPDATE</NCheckbox>
            <NCheckbox value="DELETE">DELETE</NCheckbox>
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
      <NFormItem label="目标主机">
        <NInput v-model:value="revokeForm.host" />
      </NFormItem>
      <NFormItem label="选择数据库">
        <NSelect v-model:value="revokeForm.db_name" :options="dbOptionsForGrant" placeholder="选择要回收权限的数据库" />
      </NFormItem>
    </ModalForm>

    <!-- 删除用户确认 -->
    <ConfirmDialog
      v-model:show="deleteUserConfirm"
      title="删除用户"
      :content="`确定要删除用户「${targetUser?.username}@${targetUser?.host}」吗？`"
      type="error"
      confirm-text="删除"
      :loading="deleteUserLoading"
      @confirm="doDeleteUser"
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

.form-hint {
  font-size: var(--fs-caption);
  color: var(--faint);
  margin-top: var(--space-1);
}
</style>
