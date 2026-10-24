<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { NTabs, NTabPane } from 'naive-ui'
import UserListTab from './UserListTab.vue'
import RoleListTab from './RoleListTab.vue'
import TOTPTab from './TOTPTab.vue'

const { t } = useI18n()
const activeTab = ref('users')
</script>

<template>
  <div class="users-page">
    <NTabs v-model:value="activeTab" type="line" animated class="users-tabs">
      <NTabPane name="users" :tab="t('usersIndex.tabs.users')">
        <UserListTab />
      </NTabPane>
      <NTabPane name="roles" :tab="t('usersIndex.tabs.roles')">
        <RoleListTab />
      </NTabPane>
      <NTabPane name="totp" :tab="t('usersIndex.tabs.totp')">
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
