<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NCard, NForm, NFormItem, NInput, NButton, NAlert, NIcon } from 'naive-ui'
import { LockClosedOutline, KeyOutline } from '@vicons/ionicons5'
import { resetPassword } from '@/api/auth'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()

const token = computed(() => (route.query.token as string) || '')

const loading = ref(false)
const errorMsg = ref('')
const successMsg = ref(false)

const formRef = ref()
const form = reactive({
  newPassword: '',
  confirmPassword: ''
})

const rules = computed(() => ({
  newPassword: {
    required: true,
    trigger: ['blur', 'input'],
    validator: (_rule: any, value: string) => {
      if (!value) return new Error(t('resetPassword.rules.newPasswordRequired'))
      if (value.length < 6) return new Error(t('resetPassword.rules.newPasswordMin'))
      return true
    }
  },
  confirmPassword: {
    required: true,
    trigger: ['blur', 'input'],
    validator: (_rule: any, value: string) => {
      if (!value) return new Error(t('resetPassword.rules.confirmPasswordRequired'))
      if (value !== form.newPassword) return new Error(t('resetPassword.rules.passwordMismatch'))
      return true
    }
  }
}))

async function handleSubmit() {
  errorMsg.value = ''
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    await resetPassword({ token: token.value, new_password: form.newPassword })
    successMsg.value = true
    setTimeout(() => router.push('/login'), 2000)
  } catch (err: any) {
    errorMsg.value = err.message || t('resetPassword.failed')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="rp-page">
    <div class="rp-wrap">
      <NCard class="rp-card" :bordered="false">
        <div class="rp-card__header">
          <div class="rp-icon">
            <NIcon :size="30"><KeyOutline /></NIcon>
          </div>
          <h2 class="rp-card__title">{{ t('resetPassword.title') }}</h2>
          <p class="rp-card__subtitle">{{ t('resetPassword.subtitle') }}</p>
        </div>

        <NAlert v-if="!token" type="error" :show-icon="true" class="rp-alert">
          {{ t('resetPassword.invalidToken') }}
        </NAlert>

        <template v-else>
          <NAlert v-if="errorMsg" type="error" :show-icon="true" class="rp-alert" closable @close="errorMsg = ''">
            {{ errorMsg }}
          </NAlert>
          <NAlert v-if="successMsg" type="success" :show-icon="true" class="rp-alert">
            {{ t('resetPassword.success') }}
          </NAlert>

          <NForm v-if="!successMsg" ref="formRef" :model="form" :rules="rules" size="large" @submit.prevent="handleSubmit">
            <NFormItem path="newPassword">
              <NInput
                v-model:value="form.newPassword"
                type="password"
                show-password-on="click"
                :placeholder="t('resetPassword.placeholders.newPassword')"
                :input-props="{ autocomplete: 'new-password' }"
              >
                <template #prefix>
                  <NIcon><LockClosedOutline /></NIcon>
                </template>
              </NInput>
            </NFormItem>

            <NFormItem path="confirmPassword">
              <NInput
                v-model:value="form.confirmPassword"
                type="password"
                show-password-on="click"
                :placeholder="t('resetPassword.placeholders.confirmPassword')"
                :input-props="{ autocomplete: 'new-password' }"
                @keyup.enter="handleSubmit"
              >
                <template #prefix>
                  <NIcon><LockClosedOutline /></NIcon>
                </template>
              </NInput>
            </NFormItem>

            <NFormItem>
              <NButton type="primary" block size="large" :loading="loading" @click="handleSubmit">
                {{ t('resetPassword.submit') }}
              </NButton>
            </NFormItem>
          </NForm>
        </template>

        <div class="rp-back">
          <NButton text size="small" @click="router.push('/login')">{{ t('resetPassword.backToLogin') }}</NButton>
        </div>
      </NCard>
    </div>
  </div>
</template>

<style scoped lang="scss">
.rp-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg);
}

.rp-wrap {
  width: 420px;
  max-width: calc(100vw - 32px);
}

.rp-card {
  border-radius: 12px;
  box-shadow: var(--shadow-lg);

  &__header {
    text-align: center;
    margin-bottom: var(--space-5);
  }

  &__title {
    margin: var(--space-2) 0 var(--space-1);
    font-size: 20px;
    font-weight: 600;
  }

  &__subtitle {
    margin: 0;
    font-size: 13px;
    color: var(--text-quaternary);
  }
}

.rp-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 60px;
  height: 60px;
  border-radius: 16px;
  color: var(--on-color);
  background: linear-gradient(135deg, var(--info), var(--acc3-hover, var(--acc)));
}

.rp-alert {
  margin-bottom: var(--space-4);
}

.rp-back {
  text-align: center;
  margin-top: var(--space-2);
}
</style>
