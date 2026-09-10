<script setup lang="ts">
import { ref } from 'vue'
import { NTabs, NTabPane } from 'naive-ui'
import UserListTab from './UserListTab.vue'
import RoleListTab from './RoleListTab.vue'
import TOTPTab from './TOTPTab.vue'

const activeTab = ref('users')
</script>

<template>
  <div class="users-page">
    <NTabs v-model:value="activeTab" type="line" animated class="users-tabs">
      <NTabPane name="users" tab="用户列表">
        <UserListTab />
      </NTabPane>
      <NTabPane name="roles" tab="角色权限">
        <RoleListTab />
      </NTabPane>
      <NTabPane name="totp" tab="TOTP 二次验证">
        <TOTPTab />
      </NTabPane>
    </NTabs>
  </div>
</template>

<style scoped>
.users-page {
  width: 100%;
}

.users-tabs :deep(.n-tabs-nav) {
  margin-bottom: var(--space-4);
}

.users-tabs :deep(.n-tab) {
  font-size: var(--fs-body);
  font-weight: 500;
  padding: var(--space-2) var(--space-4);
}

.users-tabs :deep(.n-tab--active) {
  color: var(--acc);
}

.users-tabs :deep(.n-tabs-tab .n-tabs-tab__label) {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}
</style>
