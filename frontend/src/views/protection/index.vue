<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import {
  NButton, NIcon, NForm, NFormItem, NInput, NInputNumber, NSwitch, NSpace,
  NSelect, NTag, NTabs, NTabPane
} from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import { RefreshOutline, ShieldCheckmarkOutline, SaveOutline } from '@vicons/ionicons5'
import { PageCard, StatusTag, ConfirmDialog, DataTable } from '@/components'
import type { StatusType } from '@/components'
import { getSites } from '@/api/site'
import {
  updateSiteWAF, getTamperStatus, buildTamperBaseline, checkTamper,
  type TamperChange
} from '@/api/features17'
import type { DataTableColumns } from 'naive-ui'
import { useI18n } from 'vue-i18n'

const message = useMessage()
const { t } = useI18n()

/* ==================== 站点列表 ==================== */
type SiteOption = { label: string; value: number }

const sites = ref<any[]>([])
const sitesLoading = ref(false)
const selectedSite = ref<number | null>(null)

async function loadSites() {
  sitesLoading.value = true
  try {
    const res = await getSites({ page: 1, pageSize: 200 })
    sites.value = res.data?.items || []
  } catch {
    sites.value = []
  } finally {
    sitesLoading.value = false
  }
}

const siteOptions = computed<SiteOption[]>(() =>
  sites.value.map((s: any) => ({ label: s.primary_domain || s.domain || `#${s.id}`, value: s.id }))
)

function siteById(id: number | null): any | null {
  return sites.value.find((s: any) => s.id === id) || null
}

const currentDomain = computed(() => {
  const s = siteById(selectedSite.value)
  return s ? (s.primary_domain || s.domain || '') : ''
})

/* ==================== Tab 1: WAF 防护 ==================== */
const wafForm = ref({ waf_enabled: false, cc_rpm: 120, ua_blacklist: '' })
const wafLoading = ref(false)

function syncWafForm() {
  const s = siteById(selectedSite.value)
  if (!s) return
  wafForm.value = {
    waf_enabled: !!s.waf_enabled,
    cc_rpm: s.cc_rpm || 120,
    ua_blacklist: s.ua_blacklist || ''
  }
}

async function handleSelectSite(v: number | null) {
  selectedSite.value = v
  syncWafForm()
  await loadTamperStatus()
}

async function saveWaf() {
  if (selectedSite.value === null) return
  wafLoading.value = true
  try {
    await updateSiteWAF(selectedSite.value, { ...wafForm.value })
    message.success(t('protection.wafSaved'))
    await loadSites()
  } catch (err: any) {
    message.error(err.message || t('protection.saveFailed'))
  } finally {
    wafLoading.value = false
  }
}

/* ==================== Tab 2: 防篡改 ==================== */
const tamperLoading = ref(false)
const hasBaseline = ref(false)
const baselineFiles = ref(0)
const checking = ref(false)
const changes = ref<TamperChange[]>([])
const checked = ref(false)
const showBaselineConfirm = ref(false)
const baselineActionLoading = ref(false)

async function loadTamperStatus() {
  changes.value = []
  checked.value = false
  if (selectedSite.value === null) {
    hasBaseline.value = false
    baselineFiles.value = 0
    return
  }
  tamperLoading.value = true
  try {
    const res = await getTamperStatus(selectedSite.value)
    hasBaseline.value = !!res.has_baseline
    baselineFiles.value = res.files || 0
  } catch {
    hasBaseline.value = false
    baselineFiles.value = 0
  } finally {
    tamperLoading.value = false
  }
}

async function doBuildBaseline() {
  if (selectedSite.value === null) return
  baselineActionLoading.value = true
  try {
    const res = await buildTamperBaseline(selectedSite.value)
    message.success(res.message || t('protection.baselineBuilt', { files: res.files }))
    await loadTamperStatus()
  } catch (err: any) {
    message.error(err.message || t('protection.baselineFailed'))
  } finally {
    baselineActionLoading.value = false
    showBaselineConfirm.value = false
  }
}

async function doCheck() {
  if (selectedSite.value === null) return
  checking.value = true
  try {
    const res = await checkTamper(selectedSite.value, false)
    changes.value = res.changes || []
    checked.value = true
    if (res.count === 0) {
      message.success(t('protection.checkOk'))
    } else {
      message.warning(t('protection.changesFound', { count: res.count }))
    }
  } catch (err: any) {
    message.error(err.message || t('protection.checkFailed'))
  } finally {
    checking.value = false
  }
}

const statusTagMap = computed<Record<string, { status: StatusType; text: string }>>(() => ({
  modified: { status: 'warning', text: t('protection.statusModified') },
  deleted: { status: 'error', text: t('protection.statusDeleted') },
  added: { status: 'pending', text: t('protection.statusAdded') }
}))

const changeColumns = computed<DataTableColumns<TamperChange>>(() => [
  {
    title: t('protection.colPath'),
    key: 'path',
    ellipsis: { tooltip: true },
    render: (row) => h('span', { class: 'mono path-cell' }, row.path)
  },
  {
    title: t('protection.colChangeType'),
    key: 'status',
    width: 110,
    render: (row) => {
      const cfg = statusTagMap.value[row.status]
      return cfg ? h(StatusTag, { status: cfg.status, text: cfg.text }) : h('span', {}, row.status)
    }
  },
  {
    title: t('protection.colOldHash'),
    key: 'old_hash',
    width: 140,
    render: (row) => row.old_hash
      ? h('span', { class: 'mono faint' }, row.old_hash.slice(0, 12))
      : h('span', { class: 'faint' }, '—')
  },
  {
    title: t('protection.colNewHash'),
    key: 'new_hash',
    width: 140,
    render: (row) => row.new_hash
      ? h('span', { class: 'mono faint' }, row.new_hash.slice(0, 12))
      : h('span', { class: 'faint' }, '—')
  }
])

onMounted(async () => {
  await loadSites()
  if (sites.value.length > 0) {
    selectedSite.value = sites.value[0].id
    syncWafForm()
    await loadTamperStatus()
  }
})
</script>

<template>
  <NTabs type="line" animated>
    <!-- ==================== Tab 1: WAF 防护 ==================== -->
    <NTabPane name="waf" :tab="t('protection.tabWaf')">
      <PageCard :title="t('protection.wafTitle')" :subtitle="t('protection.wafSubtitle')">
        <div class="toolbar">
          <span class="toolbar__label">{{ t('protection.siteLabel') }}</span>
          <NSelect
            :value="selectedSite"
            :options="siteOptions"
            :loading="sitesLoading"
            :placeholder="t('protection.selectSite')"
            filterable
            style="width: var(--input-2xl)"
            @update:value="handleSelectSite"
          />
        </div>

        <NForm v-if="selectedSite !== null" label-placement="top" style="max-width: var(--content-max-md); margin-top: var(--space-4)">
          <NFormItem :label="t('protection.enableWaf')">
            <NSwitch v-model:value="wafForm.waf_enabled" />
            <span class="form-note">{{ t('protection.wafNote') }}</span>
          </NFormItem>
          <NFormItem :label="t('protection.ccLabel')">
            <NInputNumber v-model:value="wafForm.cc_rpm" :min="1" :max="10000" style="width: var(--input-xl)">
              <template #suffix>req/s</template>
            </NInputNumber>
            <span class="form-note">{{ t('protection.ccNote') }}</span>
          </NFormItem>
          <NFormItem :label="t('protection.uaLabel')">
            <NInput
              v-model:value="wafForm.ua_blacklist"
              type="textarea"
              :rows="5"
              placeholder="sqlmap&#10;nikto&#10;curl/7.0"
            />
          </NFormItem>
          <NSpace>
            <NButton type="primary" :loading="wafLoading" @click="saveWaf">
              <template #icon><NIcon><SaveOutline /></NIcon></template>
              {{ t('protection.saveConfig') }}
            </NButton>
          </NSpace>
        </NForm>

        <div v-else class="empty-state">
          <p>{{ t('protection.noSites') }}</p>
          <span>{{ t('protection.createSiteHint') }}</span>
        </div>
      </PageCard>
    </NTabPane>

    <!-- ==================== Tab 2: 防篡改 ==================== -->
    <NTabPane name="tamper" :tab="t('protection.tabTamper')">
      <PageCard :title="t('protection.tamperTitle')" :subtitle="t('protection.tamperSubtitle')">
        <div class="toolbar">
          <span class="toolbar__label">{{ t('protection.siteLabel') }}</span>
          <NSelect
            :value="selectedSite"
            :options="siteOptions"
            :loading="sitesLoading"
            :placeholder="t('protection.selectSite')"
            filterable
            style="width: var(--input-2xl)"
            @update:value="handleSelectSite"
          />
          <template v-if="selectedSite !== null">
            <StatusTag
              v-if="hasBaseline"
              status="running"
              :text="t('protection.baselineExists', { files: baselineFiles })"
            />
            <StatusTag v-else status="warning" :text="t('protection.noBaseline')" />
          </template>
        </div>

        <template v-if="selectedSite !== null">
          <NSpace style="margin: var(--space-4) 0">
            <NButton type="primary" :loading="baselineActionLoading" @click="showBaselineConfirm = true">
              <template #icon><NIcon><ShieldCheckmarkOutline /></NIcon></template>
              {{ hasBaseline ? t('protection.rebuildBaseline') : t('protection.createBaseline') }}
            </NButton>
            <NButton :loading="checking" :disabled="!hasBaseline" @click="doCheck">
              <template #icon><NIcon><RefreshOutline /></NIcon></template>
              {{ t('protection.checkNow') }}
            </NButton>
          </NSpace>

          <div v-if="checked" class="result-block">
            <DataTable
              :columns="changeColumns"
              :data="changes"
              :pagination="false"
              size="small"
              :max-height="420"
            >
              <template #empty>
                <div class="empty-state">
                  <p>{{ t('protection.integrityOk') }}</p>
                  <span>{{ t('protection.integrityOkHint', { domain: currentDomain }) }}</span>
                </div>
              </template>
            </DataTable>
          </div>
          <div v-else class="hint-block">
            {{ t('protection.tamperHint') }}
          </div>
        </template>

        <div v-else class="empty-state">
          <p>{{ t('protection.noSites') }}</p>
          <span>{{ t('protection.createSiteHint') }}</span>
        </div>
      </PageCard>
    </NTabPane>
  </NTabs>

  <ConfirmDialog
    :show="showBaselineConfirm"
    :title="hasBaseline ? t('protection.rebuildBaseline') : t('protection.createBaseline')"
    :content="hasBaseline
      ? t('protection.rebuildContent', { domain: currentDomain })
      : t('protection.createContent', { domain: currentDomain })"
    :confirm-text="t('protection.confirm')"
    :loading="baselineActionLoading"
    @update:show="(v: boolean) => (showBaselineConfirm = v)"
    @confirm="doBuildBaseline"
  />
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background: var(--surface2);
  border-radius: var(--radius-md);
}
.toolbar__label {
  font-size: var(--fs-caption);
  color: var(--sub);
  white-space: nowrap;
}
.form-note {
  margin-left: var(--space-3);
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
  white-space: nowrap;
}
.empty-state {
  padding: var(--space-8) 0;
  text-align: center;
}
.empty-state p {
  color: var(--sub);
  margin: 0 0 var(--space-2);
}
.empty-state span {
  color: var(--text-quaternary);
  font-size: var(--fs-caption);
}
.hint-block {
  padding: var(--space-4);
  background: var(--surface2);
  border-radius: var(--radius-md);
  font-size: var(--fs-caption);
  color: var(--sub);
  line-height: 1.7;
}
.path-cell {
  font-size: var(--fs-caption);
  color: var(--ink);
}
.faint {
  color: var(--text-quaternary);
  font-size: var(--fs-caption);
}
.mono {
  font-family: var(--font-mono, 'Cascadia Code', Consolas, monospace);
}
</style>
