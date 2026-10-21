<template>
  <div class="ai-settings">
    <AppHeader />
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
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getAISettings, saveAISettings } from '../api/ai'
import AppHeader from '../components/AppHeader.vue'

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

</script>

<style scoped>
.ai-settings {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
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
