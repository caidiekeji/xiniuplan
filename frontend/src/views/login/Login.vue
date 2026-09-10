<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  NCard, NForm, NFormItem, NInput, NButton, NAlert,
  NIcon, NGrid, NGi, NModal
} from 'naive-ui'
import {
  PersonOutline, LockClosedOutline, LogInOutline,
  ShieldCheckmarkOutline, ArrowBackOutline, KeyOutline,
  MailOutline
} from '@vicons/ionicons5'
import { useAuthStore } from '@/stores/auth'
import { login as apiLogin, forgotPassword as apiForgotPassword, type LoginData } from '@/api/auth'
import { loginTOTP } from '@/api/totp'
import RhinoLogo from '@/components/RhinoLogo.vue'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const loading = ref(false)
const errorMsg = ref('')

/* ========== 登录表单 ========== */
const formRef = ref()
const form = reactive({
  username: '',
  password: ''
})

const rules = computed(() => ({
  username: {
    required: true,
    message: t('validation.usernameRequired'),
    trigger: 'blur'
  },
  password: {
    required: true,
    message: t('validation.passwordRequired'),
    trigger: 'blur'
  }
}))

/* ========== TOTP 二次验证 ========== */
const showTOTP = ref(false)
const tempToken = ref('')
const totpLoading = ref(false)
const totpFormRef = ref()
const totpCode = ref('')
const useRecoveryCode = ref(false)
const recoveryCode = ref('')

const totpRules = computed(() => ({
  code: useRecoveryCode.value
    ? {
        required: true,
        message: t('validation.recoveryCodeRequired'),
        trigger: 'blur'
      }
    : {
        required: true,
        message: t('validation.totpCodeRequired'),
        trigger: 'blur',
        validator: (_rule: any, value: string) => {
          if (!/^\d{6}$/.test(value)) {
            return new Error(t('validation.totpCodeInvalid'))
          }
          return true
        }
      }
}))

async function handleLogin() {
  errorMsg.value = ''
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    const res = await apiLogin({ username: form.username, password: form.password })
    const data = res.data as any

    // 检查是否需要 TOTP 二次验证
    if (data.require_totp && data.temp_token) {
      tempToken.value = data.temp_token
      totpCode.value = ''
      recoveryCode.value = ''
      useRecoveryCode.value = false
      showTOTP.value = true
      return
    }

    // 正常登录
    if (data.access_token) {
      authStore.setToken(data.access_token)
      if (data.user) {
        authStore.userInfo = data.user
      }
      // 首次登录/密码被重置：强制先修改密码
      if (data.require_password_change) {
        authStore.setMustChangePassword(true)
        router.push('/change-password')
        return
      }
      const redirect = (route.query.redirect as string) || '/dashboard'
      router.push(redirect)
    }
  } catch (err: any) {
    errorMsg.value = err.message || t('auth.loginFailed')
  } finally {
    loading.value = false
  }
}

async function handleTOTPVerify() {
  errorMsg.value = ''
  try {
    await totpFormRef.value?.validate()
  } catch {
    return
  }

  const code = useRecoveryCode.value ? recoveryCode.value : totpCode.value

  totpLoading.value = true
  try {
    const res = await loginTOTP({
      temp_token: tempToken.value,
      totp_code: code
    })
    const data = res.data as any

    if (data.access_token) {
      authStore.setToken(data.access_token)
      if (data.user) {
        authStore.userInfo = data.user
      }
      const redirect = (route.query.redirect as string) || '/dashboard'
      router.push(redirect)
    }
  } catch (err: any) {
    errorMsg.value = err.message || t('auth.verificationFailed')
  } finally {
    totpLoading.value = false
  }
}

function backToLogin() {
  showTOTP.value = false
  tempToken.value = ''
  totpCode.value = ''
  recoveryCode.value = ''
  errorMsg.value = ''
}

/* ========== 忘记密码 ========== */
const showForgot = ref(false)
const forgotLoading = ref(false)
const forgotSent = ref(false)
const forgotError = ref('')
const forgotForm = reactive({ username: '', email: '' })

async function handleForgot() {
  forgotError.value = ''
  if (!forgotForm.username.trim() || !forgotForm.email.trim()) {
    forgotError.value = '请输入用户名和绑定邮箱'
    return
  }
  forgotLoading.value = true
  try {
    await apiForgotPassword({ username: forgotForm.username.trim(), email: forgotForm.email.trim() })
    forgotSent.value = true
  } catch (err: any) {
    forgotError.value = err.message || '发送失败，请稍后重试'
  } finally {
    forgotLoading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <!-- 背景装饰：机柜指示灯网格 -->
    <div class="login-bg" aria-hidden="true">
      <div class="login-bg__grid">
        <span v-for="i in 48" :key="i" class="login-bg__dot" :style="{ animationDelay: (i * 0.15) + 's' }" />
      </div>
    </div>

    <div class="login-container">
      <!-- 左侧品牌区 -->
      <div class="login-brand">
        <div class="login-brand__logo">
          <RhinoLogo :size="40" />
          <span class="login-brand__name">{{ t('auth.brandName') }}</span>
        </div>
        <h1 class="login-brand__title">{{ t('auth.brandTitle') }}</h1>
        <p class="login-brand__desc">{{ t('auth.brandDesc') }}</p>
        <div class="login-brand__stats">
          <div class="login-brand__stat">
            <span class="mono">99.9%</span>
            <span class="login-brand__stat-label">{{ t('auth.uptimeLabel') }}</span>
          </div>
          <div class="login-brand__stat">
            <span class="mono">13</span>
            <span class="login-brand__stat-label">{{ t('auth.moduleLabel') }}</span>
          </div>
          <div class="login-brand__stat">
            <span class="mono">v1.0</span>
            <span class="login-brand__stat-label">{{ t('auth.versionLabel') }}</span>
          </div>
        </div>
      </div>

      <!-- 右侧登录表单 -->
      <div class="login-form-wrap">
        <NCard class="login-card" :bordered="false">
          <!-- TOTP 验证界面 -->
          <template v-if="showTOTP">
            <div class="login-card__header">
              <NButton text size="small" @click="backToLogin" class="login-back">
                <template #icon>
                  <NIcon><ArrowBackOutline /></NIcon>
                </template>
                {{ t('auth.back') }}
              </NButton>
              <h2 class="login-card__title">{{ t('auth.twoFactorAuth') }}</h2>
              <p class="login-card__subtitle">{{ t('auth.totpSubtitle') }}</p>
            </div>

            <div class="totp-icon">
              <NIcon :size="40"><ShieldCheckmarkOutline /></NIcon>
            </div>

            <NAlert v-if="errorMsg" type="error" :show-icon="true" class="login-alert" closable @close="errorMsg = ''">
              {{ errorMsg }}
            </NAlert>

            <NForm ref="totpFormRef" :model="{ code: useRecoveryCode ? recoveryCode : totpCode }" :rules="totpRules" size="large" @submit.prevent="handleTOTPVerify">
              <NFormItem path="code">
                <NInput
                  v-if="!useRecoveryCode"
                  v-model:value="totpCode"
                  :placeholder="t('auth.enterTotpCode')"
                  maxlength="6"
                  class="totp-code-input"
                  @keyup.enter="handleTOTPVerify"
                >
                  <template #prefix>
                    <NIcon><ShieldCheckmarkOutline /></NIcon>
                  </template>
                </NInput>
                <NInput
                  v-else
                  v-model:value="recoveryCode"
                  :placeholder="t('auth.enterRecoveryCode')"
                  class="totp-recovery-input"
                  @keyup.enter="handleTOTPVerify"
                >
                  <template #prefix>
                    <NIcon><KeyOutline /></NIcon>
                  </template>
                </NInput>
              </NFormItem>

              <NFormItem>
                <NButton
                  type="primary"
                  block
                  size="large"
                  :loading="totpLoading"
                  @click="handleTOTPVerify"
                  class="login-submit"
                >
                  <template #icon>
                    <NIcon><LogInOutline /></NIcon>
                  </template>
                  {{ t('auth.verifyAndLogin') }}
                </NButton>
              </NFormItem>
            </NForm>

            <div class="totp-recovery-link">
              <NButton text size="small" @click="useRecoveryCode = !useRecoveryCode">
                {{ useRecoveryCode ? t('auth.useTotpCode') : t('auth.useRecoveryCode') }}
              </NButton>
            </div>
          </template>

          <!-- 正常登录界面 -->
          <template v-else>
            <div class="login-card__header">
              <h2 class="login-card__title">{{ t('auth.welcomeBack') }}</h2>
              <p class="login-card__subtitle">{{ t('auth.loginSubtitle') }}</p>
            </div>

            <NAlert v-if="errorMsg" type="error" :show-icon="true" class="login-alert" closable @close="errorMsg = ''">
              {{ errorMsg }}
            </NAlert>

            <NForm ref="formRef" :model="form" :rules="rules" size="large" @submit.prevent="handleLogin">
              <NFormItem path="username">
                <NInput
                  v-model:value="form.username"
                  :placeholder="t('auth.username')"
                  clearable
                >
                  <template #prefix>
                    <NIcon><PersonOutline /></NIcon>
                  </template>
                </NInput>
              </NFormItem>

              <NFormItem path="password">
                <NInput
                  v-model:value="form.password"
                  type="password"
                  :placeholder="t('auth.password')"
                  show-password-on="click"
                  @keyup.enter="handleLogin"
                >
                  <template #prefix>
                    <NIcon><LockClosedOutline /></NIcon>
                  </template>
                </NInput>
              </NFormItem>

              <NFormItem>
                <NButton
                  type="primary"
                  block
                  size="large"
                  :loading="loading"
                  @click="handleLogin"
                  class="login-submit"
                >
                  <template #icon>
                    <NIcon><LogInOutline /></NIcon>
                  </template>
                  {{ t('auth.login') }}
                </NButton>
              </NFormItem>
              <div class="login-forgot">
                <NButton text size="small" @click="showForgot = true">忘记密码？</NButton>
              </div>
            </NForm>
          </template>

          <div class="login-card__footer">
            <span class="led led--ok" />
            <span>{{ t('auth.systemNormal') }}</span>
          </div>
        </NCard>

        <!-- 忘记密码弹窗 -->
        <NModal v-model:show="showForgot" preset="card" title="忘记密码" class="forgot-modal" :style="{ width: 'min(420px, 92vw)' }">
          <template v-if="!forgotSent">
            <p class="forgot-tip">输入用户名和绑定的邮箱，系统将发送密码重置链接（需先在"系统设置"中配置 SMTP）。</p>
            <NAlert v-if="forgotError" type="error" :show-icon="true" class="login-alert" closable @close="forgotError = ''">
              {{ forgotError }}
            </NAlert>
            <NForm :model="forgotForm" size="large">
              <NFormItem>
                <NInput v-model:value="forgotForm.username" placeholder="用户名" clearable>
                  <template #prefix>
                    <NIcon><PersonOutline /></NIcon>
                  </template>
                </NInput>
              </NFormItem>
              <NFormItem>
                <NInput v-model:value="forgotForm.email" placeholder="绑定邮箱" clearable @keyup.enter="handleForgot">
                  <template #prefix>
                    <NIcon><MailOutline /></NIcon>
                  </template>
                </NInput>
              </NFormItem>
              <NButton type="primary" block :loading="forgotLoading" @click="handleForgot">
                发送重置链接
              </NButton>
            </NForm>
          </template>
          <template v-else>
            <NAlert type="success" :show-icon="true">
              如果用户名和邮箱匹配，重置链接将发送到该邮箱（30 分钟内有效）。请查收邮件后重置密码。
            </NAlert>
          </template>
        </NModal>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-forgot {
  display: flex;
  justify-content: flex-end;
  margin-top: -8px;
}

.forgot-tip {
  margin: 0 0 14px;
  font-size: 13px;
  color: var(--faint);
  line-height: 1.6;
}

.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg);
  position: relative;
  overflow: hidden;
  padding: var(--space-5);
}

/* 背景指示灯网格 */
.login-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
  opacity: 0.4;
}

.login-bg__grid {
  display: grid;
  grid-template-columns: repeat(16, 1fr);
  gap: var(--space-4);
  padding: var(--space-6);
  height: 100%;
}

.login-bg__dot {
  width: var(--led-dot-size);
  height: var(--led-dot-size);
  border-radius: 50%;
  background: var(--acc2);
  opacity: 0.15;
  animation: dot-pulse 4s ease-in-out infinite;
}

@keyframes dot-pulse {
  0%, 100% { opacity: 0.1; }
  50% { opacity: 0.5; }
}

.login-container {
  display: flex;
  width: 100%;
  max-width: var(--login-max-width);
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  overflow: hidden;
  position: relative;
  z-index: 1;
}

/* 左侧品牌 */
.login-brand {
  flex: 1;
  padding: var(--space-7) var(--space-6);
  background: var(--bg2);
  border-right: 1px solid var(--line);
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.login-brand__logo {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-bottom: var(--space-6);
}

.login-brand__name {
  font-size: var(--fs-title);
  font-weight: 700;
  color: var(--ink);
}

.login-brand__title {
  font-size: var(--fs-hero);
  font-weight: 700;
  line-height: 1.3;
  color: var(--ink);
  margin-bottom: var(--space-4);
  letter-spacing: -0.02em;
  white-space: pre-line;
}

.login-brand__desc {
  font-size: var(--fs-body);
  color: var(--sub);
  line-height: var(--lh-body);
  margin-bottom: var(--space-6);
  white-space: pre-line;
}

.login-brand__stats {
  display: flex;
  gap: var(--space-5);
}

.login-brand__stat {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.login-brand__stat .mono {
  font-size: var(--fs-stat);
  font-weight: 700;
  color: var(--acc2);
}

.login-brand__stat-label {
  font-size: var(--fs-caption);
  color: var(--faint);
}

/* 右侧表单 */
.login-form-wrap {
  flex: 1;
  padding: var(--space-7) var(--space-6);
  display: flex;
  align-items: center;
}

.login-card {
  width: 100%;
  background: transparent;
  box-shadow: none;
}

.login-card__header {
  margin-bottom: var(--space-5);
  position: relative;
}

.login-card__title {
  font-size: var(--fs-headline);
  font-weight: 600;
  color: var(--ink);
  margin-bottom: var(--space-1);
}

.login-card__subtitle {
  font-size: var(--fs-body);
  color: var(--faint);
}

.login-alert {
  margin-bottom: var(--space-4);
}

.login-submit {
  margin-top: var(--space-2);
  font-weight: 600;
}

.login-card__footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  margin-top: var(--space-5);
  font-size: var(--fs-caption);
  color: var(--faint);
}

/* TOTP 样式 */
.login-back {
  position: absolute;
  top: 0;
  left: 0;
  color: var(--sub);
  padding: 0;
}

.totp-icon {
  display: flex;
  justify-content: center;
  margin-bottom: var(--space-5);
  color: var(--acc2);
}

.totp-code-input :deep(.n-input__input-el) {
  font-family: var(--font-mono);
  font-size: 1.25rem;
  letter-spacing: 0.3em;
  text-align: center;
}

.totp-recovery-input :deep(.n-input__input-el) {
  font-family: var(--font-mono);
  text-align: center;
  text-transform: uppercase;
}

.totp-recovery-link {
  display: flex;
  justify-content: center;
  margin-top: var(--space-2);
}

.totp-recovery-link .n-button {
  color: var(--acc);
  font-size: var(--fs-caption);
}

@media (max-width: 768px) {
  .login-container {
    flex-direction: column;
  }
  .login-brand {
    border-right: none;
    border-bottom: 1px solid var(--line);
    padding: var(--space-5);
  }
  .login-brand__title {
    font-size: var(--fs-stat);
  }
  .login-brand__stats {
    display: none;
  }
  .login-form-wrap {
    padding: var(--space-5);
  }
}
</style>
