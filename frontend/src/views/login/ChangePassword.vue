<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, NAlert, NIcon } from 'naive-ui'
import { LockClosedOutline, ShieldCheckmarkOutline } from '@vicons/ionicons5'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const errorMsg = ref('')
const successMsg = ref('')

const formRef = ref()
const form = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const validatePasswordStrength = (_rule: any, value: string) => {
  if (!value) return new Error('请输入新密码')
  if (value.length < 8) return new Error('新密码至少 8 位')
  if (!/[a-zA-Z]/.test(value) || !/\d/.test(value)) {
    return new Error('新密码需同时包含字母和数字')
  }
  if (value === form.oldPassword) {
    return new Error('新密码不能与原密码相同')
  }
  return true
}

const rules = computed(() => ({
  oldPassword: {
    required: true,
    message: '请输入原密码',
    trigger: 'blur'
  },
  newPassword: {
    required: true,
    trigger: ['blur', 'input'],
    validator: validatePasswordStrength
  },
  confirmPassword: {
    required: true,
    trigger: ['blur', 'input'],
    validator: (_rule: any, value: string) => {
      if (!value) return new Error('请再次输入新密码')
      if (value !== form.newPassword) return new Error('两次输入的密码不一致')
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
    await authStore.changePassword({
      old_password: form.oldPassword,
      new_password: form.newPassword
    })
    successMsg.value = '密码修改成功，请使用新密码重新登录'
    setTimeout(() => {
      router.push('/login')
    }, 1500)
  } catch (err: any) {
    errorMsg.value = err.message || '密码修改失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="chg-page">
    <div class="chg-bg" aria-hidden="true">
      <div class="chg-bg__grid">
        <span v-for="i in 24" :key="i" class="chg-bg__dot" :style="{ animationDelay: (i * 0.15) + 's' }" />
      </div>
    </div>

    <div class="chg-wrap">
      <NCard class="chg-card" :bordered="false">
        <div class="chg-card__header">
          <div class="chg-icon">
            <NIcon :size="36"><ShieldCheckmarkOutline /></NIcon>
          </div>
          <h2 class="chg-card__title">修改密码</h2>
          <p class="chg-card__subtitle">出于安全考虑，首次登录（或密码被重置后）必须修改密码后才能继续使用面板</p>
        </div>

        <NAlert v-if="errorMsg" type="error" :show-icon="true" class="chg-alert" closable @close="errorMsg = ''">
          {{ errorMsg }}
        </NAlert>
        <NAlert v-if="successMsg" type="success" :show-icon="true" class="chg-alert">
          {{ successMsg }}
        </NAlert>

        <NForm ref="formRef" :model="form" :rules="rules" size="large" @submit.prevent="handleSubmit">
          <NFormItem path="oldPassword">
            <NInput
              v-model:value="form.oldPassword"
              type="password"
              show-password-on="click"
              placeholder="原密码（初始密码见服务启动日志）"
              :input-props="{ autocomplete: 'current-password' }"
            >
              <template #prefix>
                <NIcon><LockClosedOutline /></NIcon>
              </template>
            </NInput>
          </NFormItem>

          <NFormItem path="newPassword">
            <NInput
              v-model:value="form.newPassword"
              type="password"
              show-password-on="click"
              placeholder="新密码（至少 8 位，含字母和数字）"
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
              placeholder="确认新密码"
              :input-props="{ autocomplete: 'new-password' }"
              @keyup.enter="handleSubmit"
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
              @click="handleSubmit"
              class="chg-submit"
            >
              确认修改
            </NButton>
          </NFormItem>
        </NForm>
      </NCard>
    </div>
  </div>
</template>

<style scoped lang="scss">
.chg-page {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg);
  overflow: hidden;
}

.chg-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;

  &__grid {
    display: grid;
    grid-template-columns: repeat(12, 1fr);
    gap: 28px;
    padding: 40px;
    opacity: 0.25;
  }

  &__dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: currentColor;
    animation: chg-blink 3s ease-in-out infinite;
  }
}

@keyframes chg-blink {
  0%, 100% { opacity: 0.15; }
  50% { opacity: 0.8; }
}

.chg-wrap {
  position: relative;
  width: 420px;
  max-width: calc(100vw - 32px);
}

.chg-card {
  border-radius: 12px;
  box-shadow: var(--shadow-lg);

  &__header {
    text-align: center;
    margin-bottom: 20px;
  }

  &__title {
    margin: 8px 0 6px;
    font-size: 20px;
    font-weight: 600;
  }

  &__subtitle {
    margin: 0;
    font-size: 13px;
    color: var(--faint);
    line-height: 1.5;
  }
}

.chg-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: 16px;
  color: #fff;
  background: linear-gradient(135deg, var(--acc), var(--acc-hover));
}

.chg-alert {
  margin-bottom: 16px;
}

.chg-submit {
  margin-top: 4px;
}
</style>
