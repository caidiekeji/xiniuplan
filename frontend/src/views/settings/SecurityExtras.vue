<template>
  <PageCard :title="t('securityExtras.score.title')" :subtitle="t('securityExtras.score.subtitle')">
    <template #extra>
      <NSpace>
        <NButton size="small" quaternary :loading="loading" @click="loadAll">
          <template #icon><NIcon><RefreshOutline /></NIcon></template>
          {{ t('securityExtras.refresh') }}
        </NButton>
      </NSpace>
    </template>

    <!-- 评分概览 -->
    <div class="score-overview">
      <div class="score-ring" :class="gradeClass">
        <span class="score-ring__num">{{ score?.total ?? '—' }}</span>
        <span class="score-ring__grade">{{ score?.grade || '' }}</span>
      </div>
      <div class="score-checks">
        <div v-for="chk in score?.checks || []" :key="chk.key" class="score-check" :class="`score-check--${chk.status}`">
          <div class="score-check__head">
            <span>{{ chk.label }}</span>
            <span class="mono">{{ chk.score }}/{{ chk.max }}</span>
          </div>
          <div v-if="chk.advice" class="score-check__advice">{{ chk.advice }}</div>
        </div>
      </div>
    </div>
  </PageCard>

  <PageCard :title="t('securityExtras.updates.title')" :subtitle="t('securityExtras.updates.subtitle')" style="margin-top: var(--space-4)">
    <template #extra>
      <NSpace>
        <NButton size="small" :loading="checking" @click="doCheck">
          <template #icon><NIcon><DownloadOutline /></NIcon></template>
          {{ t('securityExtras.updates.check') }}
        </NButton>
        <NPopconfirm @positive-click="doInstall">
          <template #trigger>
            <NButton size="small" type="warning" :disabled="updates.length === 0" :loading="installing">
              {{ t('securityExtras.updates.installAll', { count: updates.length }) }}
            </NButton>
          </template>
          {{ t('securityExtras.updates.installConfirm') }}
        </NPopconfirm>
      </NSpace>
    </template>
    <div v-if="updates.length === 0" class="empty-hint">{{ t('securityExtras.updates.emptyHint') }}</div>
    <div v-else class="update-list">
      <div v-for="u in updates" :key="u.name" class="update-item">
        <span class="mono">{{ u.name }}</span>
        <span class="mono update-item__ver">{{ u.new }}</span>
        <span class="update-item__src">{{ u.source }}</span>
      </div>
    </div>
  </PageCard>

  <PageCard :title="t('securityExtras.ntp.title')" :subtitle="t('securityExtras.ntp.subtitle')" style="margin-top: var(--space-4)">
    <div class="ntp-row">
      <div class="version-item">
        <span class="version-item__label">{{ t('securityExtras.ntp.sync') }}</span>
        <NSwitch :value="ntp?.ntp_enabled" :loading="ntpSwitching" @update:value="toggleNTP" />
      </div>
      <div class="version-item">
        <span class="version-item__label">{{ t('securityExtras.ntp.clockSynced') }}</span>
        <NTag size="small" :type="ntp?.synced ? 'success' : 'warning'" :bordered="false">
          {{ ntp?.synced ? t('securityExtras.ntp.yes') : t('securityExtras.ntp.no') }}
        </NTag>
      </div>
      <div class="version-item">
        <span class="version-item__label">{{ t('securityExtras.ntp.timezone') }}</span>
        <span class="mono">{{ ntp?.timezone || '—' }}</span>
      </div>
      <div class="version-item">
        <span class="version-item__label">{{ t('securityExtras.ntp.localTime') }}</span>
        <span class="mono">{{ ntp?.local_time || '—' }}</span>
      </div>
    </div>
  </PageCard>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { NButton, NIcon, NSpace, NTag, NSwitch, NPopconfirm } from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import { RefreshOutline, DownloadOutline } from '@vicons/ionicons5'
import { PageCard } from '@/components'
import {
  getSecurityScore, checkSystemUpdates, installSystemUpdates, getNTPStatus, setNTPStatus,
  type SecurityScore, type PackageUpdate, type NTPStatus
} from '@/api/newfeatures'

const message = useMessage()
const { t } = useI18n()
const loading = ref(false)
const score = ref<SecurityScore | null>(null)
const checking = ref(false)
const installing = ref(false)
const updates = ref<PackageUpdate[]>([])
const ntp = ref<NTPStatus | null>(null)
const ntpSwitching = ref(false)

const gradeClass = computed(() => {
  switch (score.value?.grade) {
    case 'A': return 'score-ring--good'
    case 'B': return 'score-ring--warn'
    default: return 'score-ring--bad'
  }
})

async function loadScore() {
  loading.value = true
  try {
    score.value = await getSecurityScore()
  } catch (e: any) {
    // 权限不足时静默
  } finally {
    loading.value = false
  }
}

async function loadNTP() {
  try {
    ntp.value = await getNTPStatus()
  } catch (e) { /* ignore */ }
}

function loadAll() {
  loadScore()
  loadNTP()
}

async function doCheck() {
  checking.value = true
  try {
    const res = await checkSystemUpdates()
    updates.value = res.items || []
    message.info(t('securityExtras.updates.found', { count: updates.value.length }))
  } catch (e: any) {
    message.error(e.message || t('securityExtras.updates.checkFailed'))
  } finally {
    checking.value = false
  }
}

async function doInstall() {
  installing.value = true
  try {
    await installSystemUpdates(true)
    message.success(t('securityExtras.updates.installDone'))
    updates.value = []
  } catch (e: any) {
    message.error(e.message || t('securityExtras.updates.installFailed'))
  } finally {
    installing.value = false
  }
}

async function toggleNTP(val: boolean) {
  ntpSwitching.value = true
  try {
    await setNTPStatus(val)
    message.success(val ? t('securityExtras.ntp.enabled') : t('securityExtras.ntp.disabled'))
    loadNTP()
  } catch (e: any) {
    message.error(e.message || t('securityExtras.ntp.opFailed'))
  } finally {
    ntpSwitching.value = false
  }
}

onMounted(loadAll)
</script>

<style scoped>
.score-overview { display: flex; gap: var(--space-5); align-items: flex-start; }
.score-ring {
  width: 110px; height: 110px; border-radius: 50%;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  flex-shrink: 0; border: 3px solid;
}
.score-ring--good { border-color: var(--ok); color: var(--ok); }
.score-ring--warn { border-color: var(--warn); color: var(--warn); }
.score-ring--bad { border-color: var(--err); color: var(--err); }
.score-ring__num { font-size: 34px; font-weight: 700; line-height: 1; }
.score-ring__grade { font-size: 14px; margin-top: var(--space-1); }
.score-checks { flex: 1; display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: var(--space-3); }
.score-check { border: 1px solid var(--line); border-radius: var(--radius-md); padding: var(--space-3); }
.score-check__head { display: flex; justify-content: space-between; font-size: 13px; font-weight: 500; }
.score-check__advice { font-size: 12px; color: var(--text-quaternary); margin-top: var(--space-2); line-height: 1.5; }
.score-check--bad { border-left: 3px solid var(--err); }
.score-check--warn { border-left: 3px solid var(--warn); }
.score-check--good { border-left: 3px solid var(--ok); }
.mono { font-family: monospace; }
.empty-hint { color: var(--text-quaternary); padding: var(--space-4) 0; text-align: center; }
.update-list { max-height: 260px; overflow: auto; }
.update-item {
  display: flex; gap: var(--space-4); padding: var(--space-1) var(--space-1); font-size: var(--fs-caption);
  border-bottom: 1px dashed var(--line);
}
.update-item__ver { color: var(--acc); }
.update-item__src { color: var(--text-quaternary); margin-left: auto; }
.ntp-row { display: flex; gap: var(--space-7); flex-wrap: wrap; }
</style>
