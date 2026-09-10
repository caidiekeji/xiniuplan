<script setup lang="ts">
import { ref, watch } from 'vue'
import { useMessage } from 'naive-ui'
import {
  NButton, NSpace, NIcon, NTag, NForm, NFormItem, NInput,
  NSelect, NAlert, NSteps, NStep, NProgress, NModal
} from 'naive-ui'
import { CheckmarkCircleOutline } from '@vicons/ionicons5'
import * as appstoreApi from '@/api/appstore'
import type { App, EnvParam, InstallResult } from '@/api/appstore'
import { getIconColor, categoryLabel, categoryTagType } from './shared'

const message = useMessage()

/* ==================== Props / Emits ==================== */

const props = defineProps<{
  visible: boolean
  app: App | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'success'): void
}>()

/* ==================== 安装状态 ==================== */

const installStep = ref(0)
const envForm = ref<Record<string, string>>({})
const installLoading = ref(false)
const installProgress = ref(0)
const installLogs = ref('')
const installResult = ref<InstallResult | null>(null)
const installError = ref<string | null>(null)

watch(
  () => props.visible,
  (val) => {
    if (val && props.app) {
      installStep.value = 0
      envForm.value = {}
      installLoading.value = false
      installProgress.value = 0
      installLogs.value = ''
      installResult.value = null
      installError.value = null
      ;(props.app.env_params || []).forEach((p: EnvParam) => {
        if (p.default !== undefined) {
          envForm.value[p.key] = String(p.default)
        }
      })
    }
  }
)

function closeModal() {
  emit('update:visible', false)
}

function goToStep(step: number) {
  // 注意：必填校验只放在"开始安装"时执行。
  // 之前校验写在 0→1 的"下一步"上，会导致用户还没进入参数配置页就被拦截，
  // 表现为"提示必填项但看不到填写界面"
  installStep.value = step
}

async function startInstall() {
  if (!props.app) return

  // 必填校验：进入安装前检查（此时用户已在参数配置页，缺什么能立刻补）
  const missing = (props.app.env_params || []).find(
    (p: EnvParam) => p.required && !envForm.value[p.key]
  )
  if (missing) {
    message.warning(`请填写必填项: ${missing.label}`)
    installStep.value = 1
    return
  }

  installStep.value = 2
  installLoading.value = true
  installProgress.value = 10
  installLogs.value = '正在准备安装环境...\n'

  try {
    installProgress.value = 30
    installLogs.value += `正在拉取镜像 ${props.app.name}:${props.app.version}...\n`

    const res = await appstoreApi.installApp(
      props.app.key,
      envForm.value,
      props.app.source_id
    )

    installProgress.value = 80
    installLogs.value += '正在配置容器和网络...\n'

    await new Promise((r) => setTimeout(r, 800))

    installProgress.value = 100
    installLogs.value += '安装完成！\n'

    installResult.value = res.data
    installLoading.value = false

    if (res.data.success) {
      message.success(`${props.app.name} 安装成功`)
      emit('success')
    } else {
      installError.value = res.data.message || '安装失败'
    }
  } catch (e: any) {
    installError.value = e.message || '安装失败'
    installLogs.value += `错误: ${e.message || '未知错误'}\n`
    installLoading.value = false
  }
}

async function copyPassword() {
  if (!installResult.value?.initial_password) return
  try {
    await navigator.clipboard.writeText(installResult.value.initial_password)
    message.success('密码已复制到剪贴板')
  } catch {
    message.error('复制失败，请手动复制')
  }
}
</script>

<template>
  <NModal
    :show="visible"
    :mask-closable="false"
    preset="card"
    :title="app ? `安装 ${app.name}` : '安装应用'"
    :style="{ width: 'min(640px, 92vw)' }"
    @update:show="(v: boolean) => { if (!v) closeModal() }"
  >
    <NSteps :current="installStep" class="install-steps">
      <NStep title="确认信息" description="查看应用详情" />
      <NStep title="配置参数" description="设置环境变量" />
      <NStep title="安装中" description="等待部署完成" />
    </NSteps>

    <!-- 步骤1: 应用信息确认 -->
    <div v-if="installStep === 0 && app" class="install-step-content">
      <div class="install-info">
        <div class="install-info__icon" :style="{ background: getIconColor(app.key) }">
          {{ app.name.charAt(0).toUpperCase() }}
        </div>
        <div class="install-info__body">
          <div class="install-info__name">{{ app.name }}</div>
          <div class="install-info__version mono">版本 v{{ app.version }}</div>
          <NTag size="small" bordered :type="categoryTagType(app.category) as any">
            {{ categoryLabel(app.category) }}
          </NTag>
        </div>
      </div>

      <div class="install-info__desc">{{ app.description }}</div>

      <div v-if="app.resources" class="install-info__resources">
        <div class="install-info__resources-title">最低资源要求</div>
        <NSpace>
          <NTag v-if="app.resources.memory" size="small" bordered type="info">
            内存: {{ app.resources.memory }}
          </NTag>
          <NTag v-if="app.resources.disk" size="small" bordered type="info">
            磁盘: {{ app.resources.disk }}
          </NTag>
          <NTag v-if="app.resources.cpu" size="small" bordered type="info">
            CPU: {{ app.resources.cpu }}
          </NTag>
        </NSpace>
      </div>

      <div v-if="app.ports && app.ports.length > 0" class="install-info__ports">
        <span class="install-info__label">暴露端口:</span>
        <span class="mono">{{ (app.ports || []).join(', ') }}</span>
      </div>
    </div>

    <!-- 步骤2: 环境参数配置 -->
    <div v-if="installStep === 1 && app" class="install-step-content">
      <NForm label-placement="top">
        <NFormItem
          v-for="param in (app.env_params || [])"
          :key="param.key"
          :label="param.label + (param.required ? ' *' : '')"
        >
          <template v-if="param.type === 'password'">
            <NInput
              v-model:value="envForm[param.key]"
              type="password"
              show-password-on="click"
              :placeholder="param.placeholder || `请输入${param.label}`"
            />
          </template>
          <template v-else-if="param.type === 'number'">
            <NInput
              v-model:value="envForm[param.key]"
              :placeholder="param.placeholder || `请输入${param.label}`"
            />
          </template>
          <template v-else-if="param.type === 'select'">
            <NSelect
              v-model:value="envForm[param.key]"
              :options="param.options || []"
              :placeholder="param.placeholder || `请选择${param.label}`"
            />
          </template>
          <template v-else>
            <NInput
              v-model:value="envForm[param.key]"
              :placeholder="param.placeholder || `请输入${param.label}`"
            />
          </template>
          <div v-if="param.description" class="form-hint">{{ param.description }}</div>
        </NFormItem>

        <NAlert
          v-if="app && (!app.env_params || app.env_params.length === 0)"
          type="info"
          :show-icon="true"
        >
          该应用无需配置环境参数，可直接安装
        </NAlert>
      </NForm>
    </div>

    <!-- 步骤3: 安装中 -->
    <div v-if="installStep === 2" class="install-step-content">
      <NProgress
        v-if="installLoading"
        :percentage="installProgress"
        indicator-placement="inside"
        class="install-progress"
      />

      <div v-if="installLoading" class="install-logs">
        <div class="install-logs__title">安装日志</div>
        <pre class="install-logs__content mono">{{ installLogs }}</pre>
      </div>

      <!-- 安装成功 -->
      <div v-if="installResult && installResult.success" class="install-success">
        <NIcon :size="48" class="install-success__icon"><CheckmarkCircleOutline /></NIcon>
        <div class="install-success__title">安装成功</div>
        <div v-if="installResult.access_url" class="install-success__item">
          <span class="install-success__label">访问地址:</span>
          <a :href="installResult.access_url" target="_blank" class="mono">
            {{ installResult.access_url }}
          </a>
        </div>
        <div v-if="installResult.initial_password" class="install-success__item">
          <span class="install-success__label">初始密码:</span>
          <span class="mono install-success__password">{{ installResult.initial_password }}</span>
          <NButton size="tiny" text @click="copyPassword">复制</NButton>
        </div>
        <div v-if="installResult.container_id" class="install-success__item">
          <span class="install-success__label">容器ID:</span>
          <span class="mono">{{ installResult.container_id.substring(0, 12) }}</span>
        </div>
      </div>

      <!-- 安装失败 -->
      <NAlert v-if="installError" type="error" :show-icon="true" class="install-error">
        {{ installError }}
      </NAlert>
    </div>

    <!-- 对话框底部按钮 -->
    <template #footer>
      <NSpace justify="end">
        <NButton v-if="installStep < 2 || (installStep === 2 && !installLoading)" @click="closeModal">
          {{ installStep === 2 && installResult?.success ? '完成' : '取消' }}
        </NButton>
        <NButton v-if="installStep === 0" type="primary" @click="goToStep(1)">下一步</NButton>
        <template v-if="installStep === 1">
          <NButton @click="installStep = 0">上一步</NButton>
          <NButton type="primary" @click="startInstall">开始安装</NButton>
        </template>
        <NButton
          v-if="installStep === 2 && installResult?.success"
          type="primary"
          @click="closeModal(); $emit('success')"
        >
          查看已安装
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped>
.install-steps {
  margin-bottom: var(--space-5);
}

.install-step-content {
  min-height: 200px;
}

.install-info {
  display: flex;
  gap: var(--space-4);
  margin-bottom: var(--space-4);
  padding: var(--space-4);
  background: var(--surface2);
  border-radius: var(--radius-md);
}

.install-info__icon {
  width: var(--size-lg);
  height: var(--size-lg);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--fs-display);
  font-weight: 700;
  color: var(--on-color);
  flex-shrink: 0;
}

.install-info__body {
  flex: 1;
}

.install-info__name {
  font-size: var(--fs-headline);
  font-weight: 600;
  color: var(--ink);
  margin-bottom: var(--space-1);
}

.install-info__version {
  font-size: var(--fs-caption);
  color: var(--faint);
  margin-bottom: var(--space-2);
}

.install-info__desc {
  font-size: var(--fs-body);
  color: var(--sub);
  line-height: var(--lh-body);
  margin-bottom: var(--space-4);
  padding: 0 var(--space-1);
}

.install-info__resources {
  margin-bottom: var(--space-3);
}

.install-info__resources-title {
  font-size: var(--fs-caption);
  color: var(--faint);
  margin-bottom: var(--space-2);
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.install-info__ports {
  font-size: var(--fs-body);
  color: var(--sub);
}

.install-info__label {
  color: var(--faint);
  margin-right: var(--space-2);
}

.form-hint {
  font-size: var(--fs-caption);
  color: var(--faint);
  margin-top: var(--space-1);
}

/* 安装进度 */
.install-progress {
  margin-bottom: var(--space-4);
}

.install-logs {
  border: 1px solid var(--line);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.install-logs__title {
  padding: var(--space-2) var(--space-3);
  background: var(--surface2);
  font-size: var(--fs-caption);
  color: var(--sub);
  border-bottom: 1px solid var(--line);
}

.install-logs__content {
  padding: var(--space-3);
  font-size: var(--fs-mono);
  color: var(--sub);
  max-height: 200px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
}

/* 安装成功 */
.install-success {
  text-align: center;
  padding: var(--space-5) 0;
}

.install-success__icon {
  color: var(--ok);
  margin-bottom: var(--space-3);
}

.install-success__title {
  font-size: var(--fs-headline);
  font-weight: 600;
  color: var(--ink);
  margin-bottom: var(--space-4);
}

.install-success__item {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  margin-bottom: var(--space-2);
  font-size: var(--fs-body);
}

.install-success__label {
  color: var(--faint);
}

.install-success__password {
  color: var(--acc);
  font-weight: 600;
}

.install-error {
  margin-top: var(--space-3);
}
</style>
