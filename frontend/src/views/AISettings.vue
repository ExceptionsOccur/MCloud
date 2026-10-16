<template>
  <div class="ai-settings">
    <el-header class="app-header">
      <div class="header-left">
        <h1>MCloud</h1>
        <span>云平台资产管理</span>
      </div>
      <AppNav />
      <div class="header-right">
        <span
          v-if="authStore.user"
          class="user-info"
        >{{ authStore.user.username }}</span>
        <el-dropdown @command="handleCommand">
          <el-button text>
            <el-icon><Setting /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="cloudResource">
                云资源录入
              </el-dropdown-item>
              <el-dropdown-item
                command="personnel"
                :disabled="$route.path === '/personnel'"
              >
                人员录入
              </el-dropdown-item>
              <el-dropdown-item
                command="publicIP"
                :disabled="$route.path === '/public-ip'"
              >
                公网IP录入
              </el-dropdown-item>
              <el-dropdown-item command="dataBackup">
                数据备份
              </el-dropdown-item>
              <el-dropdown-item
                command="aiSettings"
                :disabled="$route.path === '/ai-settings'"
              >
                AI 接入设置
              </el-dropdown-item>
              <el-dropdown-item command="aiImport">
                AI 申请入账
              </el-dropdown-item>
              <el-dropdown-item command="changePassword">
                修改密码
              </el-dropdown-item>
              <el-dropdown-item
                command="logout"
                divided
              >
                退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <div class="main-content">
      <el-card
        v-loading="loading"
        class="settings-card"
        shadow="never"
      >
        <template #header>
          <div class="card-header">
            <span>AI 接入设置</span>
            <span class="hint">用于上传申请表单后调用大模型自动提取信息（OpenAI 兼容接口）</span>
          </div>
        </template>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-width="120px"
          class="settings-form"
        >
          <el-form-item
            label="API 地址"
            prop="api_url"
          >
            <el-input
              v-model="form.api_url"
              placeholder="例如 https://api.openai.com/v1"
              maxlength="255"
            />
          </el-form-item>
          <el-form-item
            label="API Key"
            prop="api_key"
          >
            <el-input
              v-model="form.api_key"
              type="password"
              show-password
              :placeholder="keyPlaceholder"
            />
            <div
              v-if="apiKeySet"
              class="field-tip"
            >
              已配置密钥（{{ apiKeyMasked }}），留空则保留原密钥
            </div>
          </el-form-item>
          <el-form-item
            label="模型名称"
            prop="model"
          >
            <el-input
              v-model="form.model"
              placeholder="例如 gpt-4o-mini"
              maxlength="128"
            />
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              :loading="submitting"
              @click="handleSubmit"
            >
              {{ submitting ? '保存中...' : '保存' }}
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>

    <ChangePasswordDialog ref="changePasswordDialog" />
    <CloudResourceDialog ref="cloudResourceDialog" />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import { getAISettings, saveAISettings } from '../api/ai'
import ChangePasswordDialog from '../components/ChangePasswordDialog.vue'
import CloudResourceDialog from '../components/CloudResourceDialog.vue'
import AppNav from '../components/AppNav.vue'

const router = useRouter()
const authStore = useAuthStore()

const changePasswordDialog = ref(null)
const cloudResourceDialog = ref(null)

const loading = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const apiKeySet = ref(false)
const apiKeyMasked = ref('')

const form = reactive({ api_url: '', api_key: '', model: '' })

const rules = {
  api_url: [{ required: true, message: '请输入 API 地址', trigger: 'blur' }],
  model: [{ required: true, message: '请输入模型名称', trigger: 'blur' }]
}

const keyPlaceholder = computed(() =>
  apiKeySet.value ? '留空则保留原密钥' : '请输入 API Key'
)

onMounted(() => {
  authStore.fetchUserInfo()
  loadSettings()
})

async function loadSettings() {
  loading.value = true
  try {
    const res = await getAISettings()
    if (res.code === 0) {
      form.api_url = res.data.api_url || ''
      form.model = res.data.model || ''
      form.api_key = ''
      apiKeySet.value = !!res.data.api_key_set
      apiKeyMasked.value = res.data.api_key_masked || ''
    }
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const res = await saveAISettings({
      api_url: form.api_url.trim(),
      api_key: form.api_key.trim(),
      model: form.model.trim()
    })
    if (res.code === 0) {
      ElMessage.success('保存成功')
      await loadSettings()
    }
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  } finally {
    submitting.value = false
  }
}

function handleCommand(command) {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
  } else if (command === 'changePassword') {
    changePasswordDialog.value?.open()
  } else if (command === 'dataBackup') {
    router.push('/data-backup')
  } else if (command === 'cloudResource') {
    cloudResourceDialog.value?.open()
  } else if (command === 'personnel') {
    router.push('/personnel')
  } else if (command === 'publicIP') {
    router.push('/public-ip')
  } else if (command === 'aiImport') {
    router.push('/ai-import')
  }
}
</script>

<style scoped>
.ai-settings {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  height: 60px;
  background: #fff;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  z-index: 10;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-left h1 {
  font-size: 20px;
  color: #409eff;
  font-weight: 700;
}

.header-left span {
  font-size: 14px;
  color: #909399;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-info {
  font-size: 14px;
  color: #606266;
}

.main-content {
  flex: 1;
  padding: 20px;
  overflow: auto;
}

.settings-card {
  max-width: 640px;
}

.card-header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.card-header .hint {
  font-size: 12px;
  color: #909399;
  font-weight: normal;
}

.settings-form {
  max-width: 520px;
}

.field-tip {
  font-size: 12px;
  color: #909399;
  line-height: 1.4;
  margin-top: 4px;
}
</style>
