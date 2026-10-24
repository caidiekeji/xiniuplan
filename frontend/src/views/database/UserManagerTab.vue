<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from '@/composables/useMessage'
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
const { t } = useI18n()

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
    message.error(e.message || t('database.users.loadFailed'))
  } finally {
    usersLoading.value = false
  }
}

const userColumns = computed<DataTableColumns>(() => [
  { title: t('database.users.columns.username'), key: 'username', width: 200, render: (row: any) => monoEl(row.username) },
  { title: t('database.users.columns.host'), key: 'host', width: 200, render: (row: any) => monoEl(row.host) },
  {
    title: t('database.users.columns.actions'), key: 'actions', width: 200,
    render: (row: any) =>
      btnGroup([
        iconBtn({ text: t('database.users.grant'), icon: ConstructOutline, type: 'primary', onClick: () => openGrantUser(row) }),
        iconBtn({ text: t('database.users.revoke'), icon: ShieldHalfOutline, type: 'warning', onClick: () => openRevokeUser(row) }),
        iconBtn({ text: t('database.common.delete'), icon: TrashOutline, type: 'error', onClick: () => confirmDeleteUser(row) })
      ])
  }
])

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
    message.warning(t('database.users.enterUsername'))
    return
  }
  if (!userForm.value.password) {
    message.warning(t('database.users.enterPassword'))
    return
  }
  createUserLoading.value = true
  try {
    await dbApi.createUser(props.instanceId, userForm.value)
    message.success(t('database.users.createSuccess'))
    createUserModal.value = false
    loadUsers()
  } catch (e: any) {
    message.error(e.message || t('database.users.createFailed'))
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
    message.success(t('database.users.deleted', { name: targetUser.value.username }))
    deleteUserConfirm.value = false
    loadUsers()
  } catch (e: any) {
    message.error(e.message || t('database.users.deleteFailed'))
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
    message.warning(t('database.users.selectDb'))
    return
  }
  grantUserLoading.value = true
  try {
    await dbApi.grantUser(props.instanceId, grantTargetUsername.value, grantForm.value)
    message.success(t('database.users.grantSuccess'))
    grantUserModal.value = false
  } catch (e: any) {
    message.error(e.message || t('database.users.grantFailed'))
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
    message.warning(t('database.users.selectRevokeDb'))
    return
  }
  revokeUserLoading.value = true
  try {
    await dbApi.revokeUser(props.instanceId, grantTargetUsername.value, {
      host: revokeForm.value.host,
      db_name: revokeForm.value.db_name
    })
    message.success(t('database.users.revoked', { user: grantTargetUsername.value, host: revokeForm.value.host, db: revokeForm.value.db_name }))
    revokeUserModal.value = false
  } catch (e: any) {
    message.error(e.message || t('database.users.revokeFailed'))
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
          :placeholder="t('database.common.selectMysqlInstance')"
          :options="instanceOptions"
          style="width: var(--input-3xl)"
          @update:value="onInstanceChange"
        />
        <NButton type="primary" :disabled="!instanceId" @click="openCreateUser">
          <template #icon><NIcon><AddOutline /></NIcon></template>{{ t('database.users.createUser') }}
        </NButton>
      </NSpace>
    </div>

    <div v-if="!instanceId" class="empty-hint">
      <NEmpty :description="t('database.common.selectInstance')" />
    </div>

    <template v-else>
      <DataTable
        :columns="userColumns"
        :data="users"
        :loading="usersLoading"
        :pagination="false"
      />
      <div v-if="users.length === 0 && !usersLoading" class="empty-hint">
        <NEmpty :description="t('database.users.empty')" />
      </div>
    </template>

    <!-- 创建用户 -->
    <ModalForm
      v-model:show="createUserModal"
      :title="t('database.users.createModalTitle')"
      :loading="createUserLoading"
      @submit="submitCreateUser"
    >
      <NFormItem :label="t('database.users.form.username')">
        <NInput v-model:value="userForm.username" :placeholder="t('database.users.form.usernamePlaceholder')" />
      </NFormItem>
      <NFormItem :label="t('database.users.form.password')">
        <NInput v-model:value="userForm.password" type="password" show-password-on="click" :placeholder="t('database.users.form.passwordPlaceholder')" />
      </NFormItem>
      <NFormItem :label="t('database.users.form.host')">
        <NInput v-model:value="userForm.host" placeholder="%" />
        <div class="form-hint">{{ t('database.users.form.hostHint') }}</div>
      </NFormItem>
    </ModalForm>

    <!-- 授权 -->
    <ModalForm
      v-model:show="grantUserModal"
      :title="t('database.users.grantTitle', { name: grantTargetUsername })"
      :loading="grantUserLoading"
      @submit="submitGrantUser"
    >
      <NFormItem :label="t('database.users.form.targetHost')">
        <NInput v-model:value="grantForm.host" />
      </NFormItem>
      <NFormItem :label="t('database.users.form.selectDb')">
        <NSelect v-model:value="grantForm.db_name" :options="dbOptionsForGrant" :placeholder="t('database.users.form.selectDbPlaceholder')" />
      </NFormItem>
      <NFormItem :label="t('database.users.form.privileges')">
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
      :title="t('database.users.revokeTitle', { name: grantTargetUsername })"
      :loading="revokeUserLoading"
      @submit="submitRevokeUser"
    >
      <NFormItem :label="t('database.users.form.targetHost')">
        <NInput v-model:value="revokeForm.host" />
      </NFormItem>
      <NFormItem :label="t('database.users.form.selectDb')">
        <NSelect v-model:value="revokeForm.db_name" :options="dbOptionsForGrant" :placeholder="t('database.users.revokeDbPlaceholder')" />
      </NFormItem>
    </ModalForm>

    <!-- 删除用户确认 -->
    <ConfirmDialog
      v-model:show="deleteUserConfirm"
      :title="t('database.users.deleteModalTitle')"
      :content="t('database.users.deleteModalContent', { name: `${targetUser?.username}@${targetUser?.host}` })"
      type="error"
      :confirm-text="t('database.common.delete')"
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
  color: var(--text-quaternary);
  margin-top: var(--space-1);
}
</style>
