<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  NButton, NIcon, NCard, NForm, NFormItem, NInput, NSelect,
  NSpace, NTag
, NTabs, NTabPane } from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import { SaveOutline, SparklesOutline } from '@vicons/ionicons5'
import { PageCard } from '@/components'
import { diagnoseAI, type DiagnoseResult } from '@/api/features17'
import { getSecurityScore, type SecurityScore } from '@/api/newfeatures'
import { request } from '@/api/request'
import { useI18n } from 'vue-i18n'

const message = useMessage()
const { t } = useI18n()

/* 3 个功能区以标签页组织 */
const activeTab = ref('score')

/* ==================== 顶部安全评分 ==================== */
const score = ref<SecurityScore | null>(null)

async function loadScore() {
  try {
    score.value = await getSecurityScore()
  } catch {
    score.value = null
  }
}

const checkTagType: Record<string, 'success' | 'warning' | 'error'> = {
  good: 'success',
  warn: 'warning',
  bad: 'error'
}

/* ==================== AI 接口配置 ==================== */
const configLoading = ref(false)
const aiForm = ref({
  ai_api_url: '',
  ai_api_key: '',
  ai_model: ''
})

async function loadConfig() {
  try {
    const res = await request<{ settings: Record<string, string> }>({
      url: '/system/settings',
      method: 'get'
    })
    const settings = res.data?.settings || {}
    aiForm.value.ai_api_url = settings.ai_api_url || ''
    aiForm.value.ai_api_key = settings.ai_api_key || ''
    aiForm.value.ai_model = settings.ai_model || ''
  } catch {
    // 读取失败时保留默认空表单，不弹错
  }
}

async function saveConfig() {
  configLoading.value = true
  try {
    await request({
      url: '/system/settings',
      method: 'put',
      data: {
        settings: {
          ai_api_url: aiForm.value.ai_api_url,
          ai_api_key: aiForm.value.ai_api_key,
          ai_model: aiForm.value.ai_model
        }
      }
    })
    message.success(t('ai.saveSuccess'))
  } catch (err: any) {
    message.error(err.message || t('ai.saveFailed'))
  } finally {
    configLoading.value = false
  }
}

/* ==================== 日志诊断 ==================== */
const diagTypeOptions = computed(() => [
  { label: t('ai.tabLogs'), value: 'logs' },
  { label: t('ai.tabMetrics'), value: 'metrics' },
  { label: t('ai.tabConfig'), value: 'config' }
])

const diagLoading = ref(false)
const diagForm = ref({
  type: 'logs' as string,
  content: '',
  question: ''
})
const diagResult = ref<DiagnoseResult | null>(null)

async function handleDiagnose() {
  if (!diagForm.value.content.trim()) {
    message.warning(t('ai.emptyContent'))
    return
  }
  diagLoading.value = true
  diagResult.value = null
  try {
    diagResult.value = await diagnoseAI({
      type: diagForm.value.type,
      content: diagForm.value.content,
      question: diagForm.value.question || undefined
    })
  } catch (err: any) {
    message.error(err.message || t('ai.diagnoseFailed'))
  } finally {
    diagLoading.value = false
  }
}

onMounted(() => {
  loadScore()
  loadConfig()
})
</script>

<template>
  <PageCard :title="t('ai.title')" :subtitle="t('ai.subtitle')">
    <NTabs v-model:value="activeTab" type="line" animated class="ai-tabs">
    <!-- 安全评分 -->
    <NTabPane name="score" :tab="t('ai.tabs.score')">
    <NCard v-if="score" class="section-card" size="small">
      <div class="score-wrap">
        <div class="score-main">
          <span class="score-total">{{ score.total }}</span>
          <NTag size="large" :bordered="false" round>{{ score.grade }}</NTag>
        </div>
        <NSpace size="small">
          <NTag
            v-for="check in score.checks"
            :key="check.key"
            size="small"
            :bordered="false"
            round
            :type="checkTagType[check.status] || 'default'"
          >
            {{ check.label }} {{ check.score }}/{{ check.max }}
          </NTag>
        </NSpace>
      </div>
    </NCard>
    </NTabPane>

    <!-- AI 接口配置 -->
    <NTabPane name="config" :tab="t('ai.configTitle')">
    <NCard class="section-card" size="small">
      <NForm label-placement="left" label-width="90" :show-feedback="false">
        <NFormItem :label="t('ai.labelApiUrl')">
          <NInput v-model:value="aiForm.ai_api_url" placeholder="https://api.deepseek.com" clearable />
        </NFormItem>
        <NFormItem label="API Key">
          <NInput
            v-model:value="aiForm.ai_api_key"
            type="password"
            show-password-on="click"
            placeholder="sk-..."
            clearable
          />
        </NFormItem>
        <NFormItem :label="t('ai.labelModel')">
          <NInput v-model:value="aiForm.ai_model" placeholder="deepseek-chat" clearable />
        </NFormItem>
      </NForm>
      <div class="config-footer">
        <span class="hint">{{ t('ai.privacyHint') }}</span>
        <NButton type="primary" :loading="configLoading" @click="saveConfig">
          <template #icon><NIcon><SaveOutline /></NIcon></template>
          {{ t('ai.btnSave') }}
        </NButton>
      </div>
    </NCard>
    </NTabPane>

    <!-- 日志诊断 -->
    <NTabPane name="diag" :tab="t('ai.diagTitle')">
    <NCard class="section-card" size="small">
      <NForm label-placement="top">
        <NSpace size="medium" style="width: 100%; margin-bottom: var(--space-3)">
          <NFormItem :label="t('ai.labelDiagType')" style="flex: 1; margin-bottom: 0">
            <NSelect v-model:value="diagForm.type" :options="diagTypeOptions" style="width: 160px" />
          </NFormItem>
        </NSpace>
        <NFormItem :label="t('ai.labelContent')">
          <NInput
            v-model:value="diagForm.content"
            type="textarea"
            :rows="8"
            class="mono-input"
            :placeholder="t('ai.placeholderContent')"
          />
        </NFormItem>
        <NFormItem :label="t('ai.labelQuestion')">
          <NInput
            v-model:value="diagForm.question"
            :placeholder="t('ai.placeholderQuestion')"
            clearable
          />
        </NFormItem>
        <NSpace justify="end">
          <NButton type="primary" :loading="diagLoading" @click="handleDiagnose">
            <template #icon><NIcon><SparklesOutline /></NIcon></template>
            {{ t('ai.btnDiagnose') }}
          </NButton>
        </NSpace>
      </NForm>

      <NCard v-if="diagResult" class="result-card" size="small">
        <div class="result-body">{{ diagResult.analysis }}</div>
        <div class="result-meta">
          <span>{{ t('ai.resultModel', { model: diagResult.model }) }}</span>
          <span>{{ t('ai.resultLatency', { ms: diagResult.latency_ms }) }}</span>
        </div>
      </NCard>
    </NCard>
    </NTabPane>
    </NTabs>
  </PageCard>
</template>

<style scoped>
.ai-tabs :deep(.n-tabs-nav) {
  margin-bottom: var(--space-4);
}
.section-card {
  margin-bottom: 0;
}
.score-wrap {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  flex-wrap: wrap;
}
.score-main {
  display: flex;
  align-items: baseline;
  gap: var(--space-2);
}
.score-total {
  font-size: 28px;
  font-weight: 600;
  color: var(--ink);
  line-height: 1;
}
.config-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: var(--space-3);
  gap: var(--space-3);
  flex-wrap: wrap;
}
.hint {
  color: var(--text-quaternary);
  font-size: var(--fs-caption);
}
.mono-input :deep(textarea) {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: var(--fs-caption);
}
.result-card {
  margin-top: var(--space-3);
  background: var(--surface2);
}
.result-body {
  color: var(--sub);
  font-size: var(--fs-body);
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}
.result-meta {
  display: flex;
  gap: var(--space-4);
  margin-top: var(--space-3);
  padding-top: var(--space-2);
  border-top: 1px solid var(--space-3);
  color: var(--text-quaternary);
  font-size: var(--fs-caption);
}
</style>
