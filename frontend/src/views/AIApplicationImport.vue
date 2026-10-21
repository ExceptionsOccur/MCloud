<template>
  <div class="ai-import">
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
              <el-dropdown-item command="aiSettings">
                AI 接入设置
              </el-dropdown-item>
              <el-dropdown-item
                command="aiImport"
                :disabled="$route.path === '/ai-import'"
              >
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
      <el-steps
        :active="activeStep"
        finish-status="success"
        align-center
        class="steps"
      >
        <el-step title="上传表单" />
        <el-step title="确认信息" />
        <el-step title="完成" />
      </el-steps>

      <!-- 步骤 1：上传 -->
      <div
        v-if="activeStep === 0"
        class="step-panel"
      >
        <el-upload
          ref="uploadRef"
          drag
          :auto-upload="false"
          :limit="1"
          accept=".docx,.xlsx"
          :on-change="handleFileChange"
          :on-remove="handleFileRemove"
        >
          <el-icon
            class="el-icon--upload"
            :size="40"
          >
            <Upload />
          </el-icon>
          <div class="el-upload__text">
            将申请表单（.docx / .xlsx）拖到此处，或<em>点击上传</em>
          </div>
          <template #tip>
            <div class="el-upload__tip">
              支持 Word / Excel 申请表单，最大 20MB；解析后信息可编辑确认再入账
            </div>
          </template>
        </el-upload>

        <div
          v-if="uploading"
          class="progress-row"
        >
          <el-progress
            :percentage="uploadPercent"
            :stroke-width="10"
            :status="uploadPercent >= 100 ? 'success' : undefined"
          />
          <span class="progress-text">{{ progressLabel }}</span>
        </div>

        <div class="action-row">
          <el-button
            type="primary"
            :loading="uploading"
            :disabled="!file"
            @click="handleExtract"
          >
            <el-icon><MagicStick /></el-icon>
            解析表单
          </el-button>
        </div>
      </div>

      <!-- 步骤 2：确认 -->
      <div
        v-if="activeStep === 1 && result"
        class="step-panel"
      >
        <el-alert
          :title="`识别类型：${result.doc_type_name}（${result.doc_type}）`"
          type="info"
          :closable="false"
          show-icon
          style="margin-bottom: 12px;"
        />

        <el-alert
          v-if="warnings.length"
          type="warning"
          :closable="false"
          show-icon
          style="margin-bottom: 16px;"
        >
          <div
            v-for="(w, i) in warnings"
            :key="i"
          >
            {{ w }}
          </div>
        </el-alert>

        <template v-if="result.doc_type === 'host_apply'">
          <h3 class="section-title">
            主机申请信息（可编辑）
          </h3>
          <el-form
            label-width="130px"
            class="confirm-form"
          >
            <el-row :gutter="16">
              <el-col
                v-for="f in hostFields"
                :key="f.key"
                :span="12"
              >
                <el-form-item :label="f.label">
                  <el-input
                    v-model="hostForm[f.key]"
                    :placeholder="f.hint || ''"
                    clearable
                  />
                </el-form-item>
              </el-col>
            </el-row>
          </el-form>
        </template>

        <template v-else-if="result.doc_type === 'zero_trust_apply'">
          <h3 class="section-title">
            零信任申请信息（可编辑）
          </h3>
          <el-form
            label-width="130px"
            class="confirm-form"
          >
            <el-row :gutter="16">
              <el-col
                v-for="f in zeroTrustScalarFields"
                :key="f.key"
                :span="12"
              >
                <el-form-item :label="f.label">
                  <el-input
                    v-model="zeroTrustForm[f.key]"
                    :placeholder="f.hint || ''"
                    clearable
                  />
                </el-form-item>
              </el-col>
            </el-row>

            <el-form-item label="接入目标">
              <div class="targets">
                <div
                  v-for="(t, i) in zeroTrustTargets"
                  :key="i"
                  class="target-row"
                >
                  <el-input
                    v-model="t.host_ip"
                    placeholder="内网IP或主机名"
                    style="width: 200px"
                  />
                  <el-input
                    v-model="t.ports"
                    placeholder="端口，如 22 或 1000-1005"
                    style="width: 200px"
                  />
                  <el-select
                    v-model="t.protocol"
                    style="width: 100px"
                  >
                    <el-option
                      label="tcp"
                      value="tcp"
                    />
                    <el-option
                      label="udp"
                      value="udp"
                    />
                  </el-select>
                  <el-button
                    type="danger"
                    link
                    @click="removeTarget(i)"
                  >
                    删除
                  </el-button>
                </div>
                <el-button
                  type="primary"
                  link
                  @click="addTarget"
                >
                  + 添加接入目标
                </el-button>
              </div>
            </el-form-item>
          </el-form>
        </template>

        <el-alert
          v-else
          title="未能识别为受支持的申请表单类型，无法入账。可返回重新上传，或在对应模块手动录入。"
          type="error"
          :closable="false"
          show-icon
        />

        <div class="action-row">
          <el-button @click="resetAll">
            重新上传
          </el-button>
          <el-button
            v-if="isSupported"
            type="primary"
            :loading="submitting"
            @click="handleConfirm"
          >
            <el-icon><Check /></el-icon>
            确认入账
          </el-button>
        </div>
      </div>

      <!-- 步骤 3：完成 -->
      <div
        v-if="activeStep === 2"
        class="step-panel"
      >
        <el-result
          icon="success"
          title="入账成功"
          :sub-title="submitMessage"
        >
          <template #extra>
            <el-button
              type="primary"
              @click="goToList"
            >
              查看列表
            </el-button>
            <el-button @click="resetAll">
              继续入账
            </el-button>
          </template>
        </el-result>
      </div>
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
import { extractApplication, archiveApplication } from '../api/ai'
import { createHost, getHosts } from '../api/host'
import { createZeroTrust } from '../api/zero_trust'
import ChangePasswordDialog from '../components/ChangePasswordDialog.vue'
import CloudResourceDialog from '../components/CloudResourceDialog.vue'
import AppNav from '../components/AppNav.vue'

const router = useRouter()
const authStore = useAuthStore()

const changePasswordDialog = ref(null)
const cloudResourceDialog = ref(null)

const activeStep = ref(0)
const uploadRef = ref(null)
const file = ref(null)

const uploading = ref(false)
const uploadPercent = ref(0)
const progressLabel = ref('')

const result = ref(null)
const warnings = ref([])
const submitting = ref(false)
const submitMessage = ref('')

// 主机申请字段（对齐后端 CreateHostRequest）
const hostFields = [
  { key: 'region', label: '区域', hint: '如 region-a' },
  { key: 'name', label: '主机名', hint: '必填' },
  { key: 'private_ip', label: '内网IP', hint: '必填' },
  { key: 'instance_id', label: '实例ID' },
  { key: 'asset_type', label: '资产类型', hint: '虚拟机/裸金属服务器/物理机' },
  { key: 'os', label: '操作系统' },
  { key: 'cpu', label: 'CPU核数' },
  { key: 'cpu_arch', label: 'CPU架构', hint: 'C86/X86/ARM' },
  { key: 'memory', label: '内存(GB)' },
  { key: 'object_storage', label: '对象存储(GB)' },
  { key: 'system_disk', label: '系统盘(GB)' },
  { key: 'data_disk', label: '数据盘(GB)' },
  { key: 'env_type', label: '环境类型', hint: '测试/生产' },
  { key: 'is_db_server', label: '是否数据库服务器', hint: 'true/false' },
  { key: 'is_external', label: '是否云外主机', hint: 'true/false，默认 false' },
  { key: 'status', label: '状态' },
  { key: 'open_ports', label: '开放端口' },
  { key: 'tags', label: '标签' },
  { key: 'apply_unit', label: '申请单位' },
  { key: 'applicant', label: '申请人' },
  { key: 'applicant_contact', label: '申请人联系方式' },
  { key: 'project', label: '项目名称' },
  { key: 'apply_reason', label: '申请理由' },
  { key: 'apply_config', label: '申请配置' },
  { key: 'apply_time', label: '申请时间', hint: 'YYYY-MM-DD' },
  { key: 'remark', label: '备注' }
]

// 零信任标量字段（对齐后端 ZeroTrustRequest）
const zeroTrustFields = [
  { key: 'apply_unit', label: '申请单位', hint: '必填' },
  { key: 'account_name', label: '账户名', hint: '必填' },
  { key: 'account_name_zh', label: '账号姓名' },
  { key: 'applicant', label: '申请人' },
  { key: 'contact', label: '账户联系方式' },
  { key: 'exit_location', label: '接入地区', hint: '多值逗号分隔' },
  { key: 'system_name', label: '系统名称' },
  { key: 'apply_time', label: '申请时间', hint: 'YYYY-MM-DD' },
  { key: 'remark', label: '备注' }
]
const zeroTrustScalarFields = zeroTrustFields

const hostForm = reactive({})
const zeroTrustForm = reactive({})
const zeroTrustTargets = ref([])

const isSupported = computed(
  () => result.value && ['host_apply', 'zero_trust_apply'].includes(result.value.doc_type)
)

onMounted(() => {
  authStore.fetchUserInfo()
})

function handleFileChange(uploadFile) {
  file.value = uploadFile.raw
}

function handleFileRemove() {
  file.value = null
}

function resetForms() {
  Object.keys(hostForm).forEach((k) => delete hostForm[k])
  Object.keys(zeroTrustForm).forEach((k) => delete zeroTrustForm[k])
  zeroTrustTargets.value = []
}

function fillForms(data) {
  resetForms()
  const fields = data.fields || {}

  if (data.doc_type === 'host_apply') {
    hostFields.forEach((f) => {
      const v = fields[f.key]
      hostForm[f.key] = v === undefined || v === null ? '' : String(v)
    })
  } else if (data.doc_type === 'zero_trust_apply') {
    zeroTrustFields.forEach((f) => {
      const v = fields[f.key]
      zeroTrustForm[f.key] = v === undefined || v === null ? '' : String(v)
    })
    const targets = Array.isArray(fields.targets) ? fields.targets : []
    zeroTrustTargets.value = targets.map((t) => ({
      host_ip: t.host_ip || t.private_ip || '',
      ports: t.ports || (t.port === undefined || t.port === null ? '' : String(t.port)),
      protocol: t.protocol || 'tcp'
    }))
    if (zeroTrustTargets.value.length === 0) {
      addTarget()
    }
  }
}

function addTarget() {
  zeroTrustTargets.value.push({ host_ip: '', ports: '', protocol: 'tcp' })
}

function removeTarget(i) {
  zeroTrustTargets.value.splice(i, 1)
}

async function handleExtract() {
  if (!file.value) return
  uploading.value = true
  uploadPercent.value = 0
  progressLabel.value = '上传中...'
  try {
    const res = await extractApplication(file.value, (e) => {
      if (e.total) {
        // 上传占 0-90%，后端解析阶段展示 90-99%
        uploadPercent.value = Math.min(90, Math.round((e.loaded / e.total) * 90))
      }
    })
    uploadPercent.value = 99
    progressLabel.value = '解析中...'
    if (res.code === 0) {
      uploadPercent.value = 100
      progressLabel.value = '解析完成'
      result.value = res.data
      warnings.value = res.data.warnings || []
      fillForms(res.data)
      activeStep.value = 1
    }
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  } finally {
    uploading.value = false
  }
}

function buildHostPayload() {
  const numericKeys = ['cpu', 'memory', 'object_storage', 'system_disk', 'data_disk']
  const payload = {}
  hostFields.forEach((f) => {
    const v = (hostForm[f.key] || '').trim()
    if (v === '') return
    if (numericKeys.includes(f.key)) {
      const n = parseInt(v, 10)
      if (!Number.isNaN(n)) payload[f.key] = n
    } else if (f.key === 'is_db_server' || f.key === 'is_external') {
      payload[f.key] = v === 'true' || v === '1' || v === '是'
    } else {
      payload[f.key] = v
    }
  })
  return payload
}

function buildZeroTrustPayload() {
  const payload = {}
  zeroTrustFields.forEach((f) => {
    const v = (zeroTrustForm[f.key] || '').trim()
    if (v !== '') payload[f.key] = v
  })
  return payload
}

// loadAllHosts 分页拉取全量主机（内网IP/主机名 → id 映射，供接入目标解析）
async function loadAllHosts() {
  const map = new Map()
  let page = 1
  const pageSize = 100
  for (;;) {
    const res = await getHosts({ page, page_size: pageSize })
    if (res.code !== 0) break
    const hosts = res.data.hosts || []
    hosts.forEach((h) => {
      map.set(h.private_ip, h.id)
      map.set(h.name, h.id)
    })
    if (hosts.length < pageSize) break
    page += 1
  }
  return map
}

// buildZeroTrustTargets 解析接入目标 host_ip → host_id
async function buildZeroTrustTargets() {
  const valid = zeroTrustTargets.value.filter(
    (t) => (t.host_ip || '').trim() !== '' && (t.ports || '').trim() !== ''
  )
  if (valid.length === 0) {
    throw new Error('请至少填写一条有效的接入目标（主机 + 端口）')
  }
  const hostMap = await loadAllHosts()
  const targets = []
  const missing = []
  valid.forEach((t) => {
    const key = t.host_ip.trim()
    const id = hostMap.get(key)
    if (!id) {
      missing.push(key)
      return
    }
    targets.push({ host_id: id, ports: t.ports.trim(), protocol: t.protocol || 'tcp' })
  })
  if (missing.length > 0) {
    throw new Error(`未找到主机：${missing.join('、')}，请先在主机管理中录入或修正内网IP`)
  }
  return targets
}

async function handleConfirm() {
  if (!result.value) return
  submitting.value = true
  try {
    let res
    if (result.value.doc_type === 'host_apply') {
      res = await createHost(buildHostPayload())
    } else {
      let targets
      try {
        targets = await buildZeroTrustTargets()
      } catch (e) {
        ElMessage.error(e.message)
        return
      }
      const payload = buildZeroTrustPayload()
      payload.targets = targets
      res = await createZeroTrust(payload)
    }

    if (res.code === 0) {
      // 入库成功后归档源文件（失败不阻断入账结果）
      if (result.value.file_token) {
        try {
          await archiveApplication(result.value.file_token)
        } catch {
          ElMessage.warning('入账成功，但源文件归档失败')
        }
      }
      submitMessage.value = result.value.doc_type === 'host_apply'
        ? '主机申请已入账'
        : '零信任申请已入账'
      activeStep.value = 2
    }
  } finally {
    submitting.value = false
  }
}

function goToList() {
  if (result.value?.doc_type === 'zero_trust_apply') {
    router.push('/zero-trust')
  } else {
    router.push('/')
  }
}

function resetAll() {
  activeStep.value = 0
  file.value = null
  uploadRef.value?.clearFiles()
  result.value = null
  warnings.value = []
  submitMessage.value = ''
  uploadPercent.value = 0
  progressLabel.value = ''
  resetForms()
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
  } else if (command === 'aiSettings') {
    router.push('/ai-settings')
  }
}
</script>

<style scoped>
.ai-import {
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

.steps {
  max-width: 900px;
  margin: 0 auto 24px;
}

.step-panel {
  max-width: 900px;
  margin: 0 auto;
}

.progress-row {
  margin-top: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.progress-text {
  font-size: 13px;
  color: #909399;
  white-space: nowrap;
}

.action-row {
  margin-top: 20px;
  display: flex;
  gap: 12px;
}

.section-title {
  font-size: 15px;
  color: #303133;
  margin: 8px 0 16px;
}

.confirm-form {
  max-width: 820px;
}

.targets {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.target-row {
  display: flex;
  gap: 8px;
  align-items: center;
}
</style>
