<template>
  <PageCard :title="t('dockerResources.title')" :subtitle="t('dockerResources.subtitle')">
    <template #extra>
      <NButton size="small" quaternary @click="loadAll">
        <template #icon><NIcon><RefreshOutline /></NIcon></template>
        {{ t('dockerResources.refresh') }}
      </NButton>
    </template>

    <NTabs type="line" animated>
      <NTabPane name="networks" :tab="t('dockerResources.tabs.networks')">
        <div class="res-toolbar">
          <NButton size="small" type="primary" @click="showCreateNetwork = true">{{ t('dockerResources.createNetwork') }}</NButton>
        </div>
        <DataTable
          :columns="networkColumns"
          :data="networks"
          :loading="networkLoading"
          :pagination="false"
          :row-key="(r: any) => r.id"
        />
      </NTabPane>

      <NTabPane name="volumes" :tab="t('dockerResources.tabs.volumes')">
        <div class="res-toolbar">
          <NButton size="small" type="primary" @click="showCreateVolume = true">{{ t('dockerResources.createVolume') }}</NButton>
        </div>
        <DataTable
          :columns="volumeColumns"
          :data="volumes"
          :loading="volumeLoading"
          :pagination="false"
          :row-key="(r: any) => r.name"
        />
      </NTabPane>
    </NTabs>

    <!-- 创建网络 -->
    <ModalForm
      v-model:show="showCreateNetwork"
      :title="t('dockerResources.dialogs.createNetworkTitle')"
      :width="460"
      :loading="creating"
      :confirm-text="t('dockerResources.form.confirm')"
      :cancel-text="t('dockerResources.form.cancel')"
      @submit="doCreateNetwork"
    >
      <NForm label-placement="left" label-width="90">
        <NFormItem :label="t('dockerResources.form.nameLabel')" required><NInput v-model:value="netForm.name" :placeholder="t('dockerResources.form.namePlaceholderNetwork')" /></NFormItem>
        <NFormItem :label="t('dockerResources.form.driverLabel')"><NInput v-model:value="netForm.driver" placeholder="bridge" /></NFormItem>
        <NFormItem :label="t('dockerResources.form.subnetLabel')"><NInput v-model:value="netForm.subnet" :placeholder="t('dockerResources.form.subnetPlaceholder')" /></NFormItem>
        <NFormItem :label="t('dockerResources.form.gatewayLabel')"><NInput v-model:value="netForm.gateway" :placeholder="t('dockerResources.form.gatewayPlaceholder')" /></NFormItem>
      </NForm>
    </ModalForm>

    <!-- 创建存储卷 -->
    <ModalForm
      v-model:show="showCreateVolume"
      :title="t('dockerResources.dialogs.createVolumeTitle')"
      :width="420"
      :loading="creating"
      :confirm-text="t('dockerResources.form.confirm')"
      :cancel-text="t('dockerResources.form.cancel')"
      @submit="doCreateVolume"
    >
      <NForm label-placement="left" label-width="90">
        <NFormItem :label="t('dockerResources.form.nameLabel')" required><NInput v-model:value="volForm.name" :placeholder="t('dockerResources.form.namePlaceholderVolume')" /></NFormItem>
        <NFormItem :label="t('dockerResources.form.driverLabel')"><NInput v-model:value="volForm.driver" placeholder="local" /></NFormItem>
      </NForm>
    </ModalForm>
  </PageCard>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NTabs, NTabPane, NButton, NIcon, NForm, NFormItem, NInput, useDialog
} from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import { RefreshOutline, TrashOutline } from '@vicons/ionicons5'
import { PageCard, DataTable, ModalForm } from '@/components'
import {
  listNetworks, createNetwork, removeNetwork, listVolumes, createVolume, removeVolume,
  type DockerNetwork, type DockerVolume
} from '@/api/newfeatures'

const message = useMessage()
const dialog = useDialog()
const { t } = useI18n()

const networks = ref<DockerNetwork[]>([])
const volumes = ref<DockerVolume[]>([])
const networkLoading = ref(false)
const volumeLoading = ref(false)
const showCreateNetwork = ref(false)
const showCreateVolume = ref(false)
const creating = ref(false)
const netForm = ref({ name: '', driver: '', subnet: '', gateway: '' })
const volForm = ref({ name: '', driver: '' })

async function loadNetworks() {
  networkLoading.value = true
  try {
    networks.value = (await listNetworks()).items || []
  } catch (e: any) {
    message.error(e.message || t('dockerResources.messages.loadNetworksFailed'))
  } finally {
    networkLoading.value = false
  }
}

async function loadVolumes() {
  volumeLoading.value = true
  try {
    volumes.value = (await listVolumes()).items || []
  } catch (e: any) {
    message.error(e.message || t('dockerResources.messages.loadVolumesFailed'))
  } finally {
    volumeLoading.value = false
  }
}

function loadAll() {
  loadNetworks()
  loadVolumes()
}

async function doCreateNetwork() {
  creating.value = true
  try {
    await createNetwork(netForm.value)
    message.success(t('dockerResources.messages.networkCreated'))
    showCreateNetwork.value = false
    netForm.value = { name: '', driver: '', subnet: '', gateway: '' }
    loadNetworks()
  } catch (e: any) {
    message.error(e.message || t('dockerResources.messages.createFailed'))
  } finally {
    creating.value = false
  }
}

async function doCreateVolume() {
  creating.value = true
  try {
    await createVolume(volForm.value)
    message.success(t('dockerResources.messages.volumeCreated'))
    showCreateVolume.value = false
    volForm.value = { name: '', driver: '' }
    loadVolumes()
  } catch (e: any) {
    message.error(e.message || t('dockerResources.messages.createFailed'))
  } finally {
    creating.value = false
  }
}

function confirmRemoveNetwork(row: DockerNetwork) {
  dialog.warning({
    title: t('dockerResources.deleteNetwork.title'),
    content: t('dockerResources.deleteNetwork.content', { name: row.name }),
    positiveText: t('dockerResources.delete'),
    negativeText: t('dockerResources.deleteDialog.cancel'),
    onPositiveClick: async () => {
      await removeNetwork(row.id)
      message.success(t('dockerResources.messages.deleted'))
      loadNetworks()
    }
  })
}

function confirmRemoveVolume(row: DockerVolume) {
  if (row.in_use) {
    message.warning(t('dockerResources.messages.volumeInUse'))
    return
  }
  dialog.warning({
    title: t('dockerResources.deleteVolume.title'),
    content: t('dockerResources.deleteVolume.content', { name: row.name }),
    positiveText: t('dockerResources.delete'),
    negativeText: t('dockerResources.deleteDialog.cancel'),
    onPositiveClick: async () => {
      await removeVolume(row.name, true)
      message.success(t('dockerResources.messages.deleted'))
      loadVolumes()
    }
  })
}

const networkColumns = computed(() => [
  { title: t('dockerResources.columns.name'), key: 'name', width: 180 },
  { title: t('dockerResources.columns.driver'), key: 'driver', width: 100 },
  { title: t('dockerResources.columns.subnet'), key: 'subnet', width: 160 },
  { title: t('dockerResources.columns.containers'), key: 'containers', render: (r: any) => (r.containers || []).length },
  {
    title: t('dockerResources.columns.actions'), key: 'actions', width: 90,
    render: (row: any) => h(
      NButton, {
        size: 'tiny', type: 'error', quaternary: true, disabled: row.built_in,
        onClick: () => confirmRemoveNetwork(row)
      }, {
        icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
        default: () => t('dockerResources.delete')
      }
    )
  }
] as any)

const volumeColumns = computed(() => [
  { title: t('dockerResources.columns.name'), key: 'name', width: 220, ellipsis: { tooltip: true } },
  { title: t('dockerResources.columns.driver'), key: 'driver', width: 90 },
  {
    title: t('dockerResources.columns.inUse'), key: 'in_use', width: 90,
    render: (r: any) => (r.in_use ? t('dockerResources.yes') : t('dockerResources.no'))
  },
  { title: t('dockerResources.columns.usedBy'), key: 'used_by', ellipsis: { tooltip: true } },
  {
    title: t('dockerResources.columns.actions'), key: 'actions', width: 90,
    render: (row: any) => h(
      NButton, {
        size: 'tiny', type: 'error', quaternary: true, disabled: row.in_use,
        onClick: () => confirmRemoveVolume(row)
      }, {
        icon: () => h(NIcon, null, { default: () => h(TrashOutline) }),
        default: () => t('dockerResources.delete')
      }
    )
  }
] as any)

onMounted(loadAll)
</script>

<style scoped>
.res-toolbar { margin-bottom: var(--space-3); display: flex; gap: var(--space-2); }
</style>
