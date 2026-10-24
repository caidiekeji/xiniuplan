<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { NButton, NIcon, NForm, NFormItem, NInput, NSteps, NStep, NSpace, NAlert } from 'naive-ui'
import { useMessage } from '@/composables/useMessage'
import {
  ShieldCheckmarkOutline, QrCodeOutline, CopyOutline,
  CheckmarkCircleOutline, WarningOutline
} from '@vicons/ionicons5'
import { PageCard, StatusTag, ContentModal } from '@/components'
import {
  getTOTPStatus, bindTOTP, verifyTOTP, disableTOTP,
  type TOTPStatus
} from '@/api/totp'
import QRCode from 'qrcode'

const message = useMessage()
const { t } = useI18n()

/* ========== 状态 ========== */
const loading = ref(false)
const totpStatus = ref<TOTPStatus | null>(null)

async function loadStatus() {
  loading.value = true
  try {
    const res = await getTOTPStatus()
    if (res.data) {
      totpStatus.value = res.data
    }
  } catch (err: any) {
    message.error(err.message || t('totp.loadStatusError'))
  } finally {
    loading.value = false
  }
}

/* ========== 绑定流程 ========== */
const showBind = ref(false)
const bindStep = ref(0)
const bindLoading = ref(false)
const bindSecret = ref('')
const bindOtpauthUrl = ref('')
const qrCodeDataUrl = ref('')
const verifyCode = ref('')
const verifyFormRef = ref()
const recoveryCodes = ref<string[]>([])

const verifyRules = computed(() => ({
  code: {
    required: true,
    message: t('totp.validate.codeRequired'),
    trigger: 'blur',
    validator: (_rule: any, value: string) => {
      if (!/^\d{6}$/.test(value)) {
        return new Error(t('totp.validate.enterDigits'))
      }
      return true
    }
  }
}))

async function startBind() {
  bindStep.value = 0
  bindSecret.value = ''
  bindOtpauthUrl.value = ''
  qrCodeDataUrl.value = ''
  verifyCode.value = ''
  recoveryCodes.value = []
  showBind.value = true

  bindLoading.value = true
  try {
    const res = await bindTOTP()
    if (res.data) {
      bindSecret.value = res.data.secret
      bindOtpauthUrl.value = res.data.otpauth_url
      // 生成二维码
      qrCodeDataUrl.value = await QRCode.toDataURL(res.data.otpauth_url, {
        width: 200,
        margin: 2,
        color: {
          dark: '#1a2030',
          light: '#ffffff'
        }
      })
    }
  } catch (err: any) {
    message.error(err.message || t('totp.getBindInfoError'))
    showBind.value = false
  } finally {
    bindLoading.value = false
  }
}

function copySecret() {
  navigator.clipboard.writeText(bindSecret.value).then(() => {
    message.success(t('totp.secretCopied'))
  }).catch(() => {
    message.error(t('totp.copyFailed'))
  })
}

function copyRecoveryCodes() {
  const text = recoveryCodes.value.join('\n')
  navigator.clipboard.writeText(text).then(() => {
    message.success(t('totp.recoveryCopied'))
  }).catch(() => {
    message.error(t('totp.copyFailed'))
  })
}

async function handleVerify() {
  try {
    await verifyFormRef.value?.validate()
  } catch {
    return
  }

  bindLoading.value = true
  try {
    const res = await verifyTOTP({ secret: bindSecret.value, code: verifyCode.value })
    if (res.data) {
      recoveryCodes.value = res.data.recovery_codes || []
      bindStep.value = 1
      message.success(t('totp.bindSuccessMsg'))
    }
  } catch (err: any) {
    message.error(err.message || t('totp.verifyFailed'))
  } finally {
    bindLoading.value = false
  }
}

function finishBind() {
  showBind.value = false
  loadStatus()
}

/* ========== 禁用 TOTP ========== */
const showDisable = ref(false)
const disableLoading = ref(false)
const disableFormRef = ref()
const disableCode = ref('')

const disableRules = computed(() => ({
  code: {
    required: true,
    message: t('totp.validate.codeRequired'),
    trigger: 'blur',
    validator: (_rule: any, value: string) => {
      if (!/^\d{6}$/.test(value)) {
        return new Error(t('totp.validate.enterDigits'))
      }
      return true
    }
  }
}))

function openDisable() {
  disableCode.value = ''
  showDisable.value = true
}

async function handleDisableConfirm() {
  try {
    await disableFormRef.value?.validate()
  } catch {
    return
  }

  disableLoading.value = true
  try {
    await disableTOTP({ code: disableCode.value })
    message.success(t('totp.disableSuccessMsg'))
    showDisable.value = false
    loadStatus()
  } catch (err: any) {
    message.error(err.message || t('totp.disableFailed'))
  } finally {
    disableLoading.value = false
  }
}

onMounted(() => {
  loadStatus()
})
</script>

<template>
  <PageCard :title="t('totp.title')" :subtitle="t('totp.subtitle')">
    <div class="totp-container">
      <!-- 当前状态卡片 -->
      <div class="totp-status-card">
        <div class="totp-status-card__icon">
          <NIcon :size="32"><ShieldCheckmarkOutline /></NIcon>
        </div>
        <div class="totp-status-card__info">
          <div class="totp-status-card__row">
            <span class="totp-status-card__label">{{ t('totp.statusLabel') }}</span>
            <StatusTag
              :status="totpStatus?.enabled ? 'running' : 'stopped'"
              :text="totpStatus?.enabled ? t('totp.status.enabled') : t('totp.status.disabled')"
            />
          </div>
          <p v-if="totpStatus?.enabled && totpStatus?.bound_at" class="totp-status-card__time">
            {{ t('totp.boundAt') }} <span class="mono">{{ totpStatus.bound_at }}</span>
          </p>
          <p class="totp-status-card__desc">
            {{ t('totp.compatDesc') }}
            {{ t('totp.loginDesc') }}
          </p>
        </div>
        <div class="totp-status-card__action">
          <NButton
            v-if="!totpStatus?.enabled"
            type="primary"
            :loading="loading"
            @click="startBind"
          >
            <template #icon>
              <NIcon><QrCodeOutline /></NIcon>
            </template>
            {{ t('totp.bindNow') }}
          </NButton>
          <NButton
            v-else
            type="error"
            ghost
            @click="openDisable"
          >
            <template #icon>
              <NIcon><WarningOutline /></NIcon>
            </template>
            {{ t('totp.disableBtn') }}
          </NButton>
        </div>
      </div>

      <!-- 恢复码说明 -->
      <NAlert type="warning" :show-icon="true" class="totp-notice">
        <template #header>{{ t('totp.recoveryNotice.title') }}</template>
        {{ t('totp.recoveryNotice.line1') }}
        {{ t('totp.recoveryNotice.line2') }}
      </NAlert>
    </div>
  </PageCard>

  <!-- 绑定流程对话框 -->
  <ContentModal
    :show="showBind"
    :mask-closable="false"
    :title="t('totp.bindModal.title')"
    size="md"
    @update:show="(v: boolean) => (showBind = v)"
  >
    <NSteps :current="bindStep" class="totp-steps">
      <NStep :title="t('totp.bindModal.steps.scan.title')" :description="t('totp.bindModal.steps.scan.desc')" />
      <NStep :title="t('totp.bindModal.steps.verify.title')" :description="t('totp.bindModal.steps.verify.desc')" />
    </NSteps>

    <!-- 步骤1：二维码 -->
    <div v-if="bindStep === 0" class="totp-step">
      <NSpin :show="bindLoading">
        <div class="totp-qr-section">
          <div class="totp-qr">
            <img v-if="qrCodeDataUrl" :src="qrCodeDataUrl" :alt="t('totp.bindModal.qrAlt')" />
            <div v-else class="totp-qr__placeholder">
              <NIcon :size="48"><QrCodeOutline /></NIcon>
            </div>
          </div>
          <div class="totp-qr-info">
            <p class="totp-qr-info__tip">{{ t('totp.bindModal.qrTip') }}</p>
            <p class="totp-qr-info__sub">{{ t('totp.bindModal.qrManual') }}</p>
            <div class="totp-secret">
              <code class="mono">{{ bindSecret }}</code>
              <NButton size="tiny" quaternary @click="copySecret">
                <template #icon>
                  <NIcon><CopyOutline /></NIcon>
                </template>
                {{ t('totp.bindModal.copy') }}
              </NButton>
            </div>
          </div>
        </div>

        <NForm ref="verifyFormRef" :model="{ code: verifyCode }" :rules="verifyRules" label-placement="top">
          <NFormItem :label="t('totp.bindModal.codeLabel')" path="code">
            <NInput
              v-model:value="verifyCode"
              :placeholder="t('totp.validate.enterDigits')"
              maxlength="6"
              class="totp-code-input"
            />
          </NFormItem>
        </NForm>

        <div class="totp-step__actions">
          <NButton @click="showBind = false">{{ t('totp.cancel') }}</NButton>
          <NButton type="primary" :loading="bindLoading" @click="handleVerify">
            {{ t('totp.bindModal.confirm') }}
          </NButton>
        </div>
      </NSpin>
    </div>

    <!-- 步骤2：恢复码 -->
    <div v-else class="totp-step">
      <NAlert type="success" :show-icon="true" class="totp-success">
        <template #header>{{ t('totp.success.title') }}</template>
        {{ t('totp.success.body') }}
      </NAlert>

      <div class="totp-recovery">
        <div class="totp-recovery__header">
          <span class="totp-recovery__title">{{ t('totp.recoveryCount', { count: recoveryCodes.length }) }}</span>
          <NButton size="tiny" quaternary @click="copyRecoveryCodes">
            <template #icon>
              <NIcon><CopyOutline /></NIcon>
            </template>
            {{ t('totp.copyAll') }}
          </NButton>
        </div>
        <div class="totp-recovery__list">
          <div
            v-for="(code, idx) in recoveryCodes"
            :key="idx"
            class="totp-recovery__item mono"
          >
            {{ code }}
          </div>
        </div>
      </div>

      <div class="totp-step__actions">
        <NButton type="primary" @click="finishBind">
          <template #icon>
            <NIcon><CheckmarkCircleOutline /></NIcon>
          </template>
          {{ t('totp.savedRecovery') }}
        </NButton>
      </div>
    </div>
  </ContentModal>

  <!-- 禁用确认对话框 -->
  <ContentModal
    :show="showDisable"
    :mask-closable="false"
    :title="t('totp.disableModal.title')"
    size="sm"
    @update:show="(v: boolean) => (showDisable = v)"
  >
    <NAlert type="warning" :show-icon="true" class="totp-disable-alert">
      {{ t('totp.disableModal.desc') }}
    </NAlert>

    <NForm ref="disableFormRef" :model="{ code: disableCode }" :rules="disableRules" label-placement="top">
      <NFormItem :label="t('totp.disableModal.codeLabel')" path="code">
        <NInput
          v-model:value="disableCode"
          :placeholder="t('totp.disableModal.codePlaceholder')"
          maxlength="6"
        />
      </NFormItem>
    </NForm>

    <template #footer>
      <NSpace justify="end">
        <NButton @click="showDisable = false">{{ t('totp.cancel') }}</NButton>
        <NButton type="error" :loading="disableLoading" @click="handleDisableConfirm">
          {{ t('totp.disableModal.confirm') }}
        </NButton>
      </NSpace>
    </template>
  </ContentModal>
</template>

<style scoped>
.totp-container {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.totp-status-card {
  display: flex;
  align-items: flex-start;
  gap: var(--space-5);
  padding: var(--space-5);
  background: var(--surface2);
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
}

.totp-status-card__icon {
  flex-shrink: 0;
  width: var(--size-lg);
  height: var(--size-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--acc2-soft);
  border-radius: var(--radius-md);
  color: var(--acc2);
}

.totp-status-card__info {
  flex: 1;
  min-width: 0;
}

.totp-status-card__row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-bottom: var(--space-2);
}

.totp-status-card__label {
  color: var(--sub);
  font-size: var(--fs-body);
}

.totp-status-card__time {
  font-size: var(--fs-caption);
  color: var(--text-quaternary);
  margin-bottom: var(--space-2);
}

.totp-status-card__desc {
  font-size: var(--fs-caption);
  color: var(--sub);
  line-height: var(--lh-body);
  margin: 0;
}

.totp-status-card__action {
  flex-shrink: 0;
}

.totp-notice {
  border-radius: var(--radius-md);
}

.totp-steps {
  margin-bottom: var(--space-5);
}

.totp-step {
  min-height: 280px;
}

.totp-qr-section {
  display: flex;
  gap: var(--space-5);
  margin-bottom: var(--space-5);
  padding: var(--space-4);
  background: var(--surface2);
  border-radius: var(--radius-md);
  border: 1px solid var(--line);
}

.totp-qr {
  flex-shrink: 0;
  width: 200px;
  height: 200px;
  background: var(--paper);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.totp-qr img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.totp-qr__placeholder {
  color: var(--text-quaternary);
}

.totp-qr-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: var(--space-2);
}

.totp-qr-info__tip {
  font-weight: 500;
  color: var(--ink);
  margin: 0;
}

.totp-qr-info__sub {
  font-size: var(--fs-caption);
  color: var(--sub);
  margin: 0;
}

.totp-secret {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: var(--radius-sm);
}

.totp-secret code {
  flex: 1;
  font-size: var(--fs-mono);
  color: var(--acc);
  word-break: break-all;
}

.totp-code-input :deep(.n-input__input-el) {
  font-family: var(--font-mono);
  font-size: 1.25rem;
  letter-spacing: 0.3em;
  text-align: center;
}

.totp-step__actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2);
  margin-top: var(--space-5);
}

.totp-success {
  margin-bottom: var(--space-4);
}

.totp-recovery {
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.totp-recovery__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-3) var(--space-4);
  background: var(--surface2);
  border-bottom: 1px solid var(--line);
}

.totp-recovery__title {
  font-weight: 600;
  color: var(--ink);
  font-size: var(--fs-body);
}

.totp-recovery__list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0;
  padding: var(--space-2);
}

.totp-recovery__item {
  padding: var(--space-2) var(--space-3);
  font-size: var(--fs-mono);
  color: var(--sub);
  text-align: center;
  border-radius: var(--radius-sm);
}

.totp-recovery__item:hover {
  background: var(--surface2);
  color: var(--ink);
}

.totp-disable-alert {
  margin-bottom: var(--space-4);
}

@media (max-width: 640px) {
  .totp-status-card {
    flex-direction: column;
  }
  .totp-qr-section {
    flex-direction: column;
    align-items: center;
  }
  .totp-recovery__list {
    grid-template-columns: 1fr;
  }
}

/* 移动端：恢复码单列 */
@media (max-width: 640px) {
  .totp-recovery__list {
    grid-template-columns: 1fr;
  }
}
</style>
