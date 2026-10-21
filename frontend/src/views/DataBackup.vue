<template>
  <div class="data-backup">
    <el-header class="app-header">
      <div class="header-left">
        <h1>MCloud</h1>
        <span>云平台资产管理</span>
      </div>
      <AppNav
        variant="backup"
      />
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
              <el-dropdown-item command="personnel">
                人员录入
              </el-dropdown-item>
              <el-dropdown-item
                command="publicIP"
              >
                公网IP录入
              </el-dropdown-item>
              <el-dropdown-item
                command="dataBackup"
                disabled
              >
                数据备份
              </el-dropdown-item>
              <el-dropdown-item
                command="aiSettings"
                :disabled="$route.path === '/ai-settings'"
              >
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
      <el-card shadow="never">
        <template #header>
          <span>数据备份（导出 / 导入）</span>
        </template>
        <el-alert
          type="info"
          :closable="false"
          show-icon
          title="单个 Excel 包含 10 张业务表（人员/公网IP/云资源/IP网段/主机/申请信息/零信任/端口映射/对象存储/对象存储主机），关联用自然键表达（人员姓名、内网IP），不包含账号密码。"
        />
        <el-alert
          type="warning"
          :closable="false"
          show-icon
          style="margin-top: 10px;"
          title="导出必须设置密码：文件为 Excel 标准加密的 .xlsx，可用 Excel/WPS 输密码打开、修改后另存再导入。导入时填写密码即先解密；未加密的 .xlsx 可直接导入。导入仅新增与更新（不删除已有数据），任一行失败将整体回滚。"
        />

        <div class="action-row">
          <el-input
            v-model="exportPassword"
            type="password"
            show-password
            clearable
            placeholder="导出密码（必填，至少 6 位，用于加密 Excel）"
            style="max-width: 320px;"
          />
          <el-button
            type="primary"
            :loading="exporting"
            @click="handleExport"
          >
            <el-icon><Download /></el-icon>
            导出全部数据（加密）
          </el-button>
        </div>

        <el-divider />

        <el-upload
          ref="uploadRef"
          drag
          :auto-upload="false"
          :limit="1"
          accept=".xlsx"
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
            将 .xlsx 备份文件（可加密）拖到此处，或<em>点击上传</em>
          </div>
          <template #tip>
            <div class="el-upload__tip">
              支持本页导出的加密 .xlsx 或未加密 .xlsx，最大 20MB
            </div>
          </template>
        </el-upload>

        <div class="action-row">
          <el-input
            v-model="importPassword"
            type="password"
            show-password
            clearable
            placeholder="打开密码（加密 Excel 必填，未加密可留空）"
            style="max-width: 320px;"
          />
          <el-button
            type="success"
            :loading="importing"
            :disabled="!file"
            @click="handleImport"
          >
            <el-icon><Upload /></el-icon>
            导入（覆盖更新）
          </el-button>
        </div>

        <div
          v-if="report"
          class="report"
        >
          <el-alert
            :type="report.committed ? 'success' : 'error'"
            :closable="false"
            show-icon
            :title="report.committed ? '导入成功，已提交' : '导入失败，已整体回滚，库数据未变化'"
            style="margin-bottom: 12px;"
          />
          <el-table
            v-if="report.sheets && report.sheets.length"
            :data="report.sheets"
            size="small"
            border
          >
            <el-table-column
              prop="sheet"
              label="工作表"
              width="160"
            />
            <el-table-column
              prop="rows"
              label="数据行"
              width="90"
            />
            <el-table-column
              prop="created"
              label="新增"
              width="90"
            />
            <el-table-column
              prop="updated"
              label="更新"
              width="90"
            />
            <el-table-column
              prop="skipped"
              label="未变化"
              width="90"
            />
          </el-table>
          <template v-if="report.errors && report.errors.length">
            <div class="error-title">
              逐行错误明细（共 {{ report.errors.length }} 条）
            </div>
            <el-table
              :data="report.errors"
              size="small"
              border
            >
              <el-table-column
                prop="sheet"
                label="工作表"
                width="160"
              />
              <el-table-column
                prop="row"
                label="行号"
                width="90"
              />
              <el-table-column
                prop="message"
                label="错误信息"
              />
            </el-table>
          </template>
        </div>
      </el-card>
    </div>

    <ChangePasswordDialog ref="changePasswordDialog" />
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import { exportAll, importAll } from '../api/dataExchange'
import { downloadBlob, formatDate } from '../utils'
import { encryptXlsx, decryptXlsx } from '../utils/xlsxCrypto'
import ChangePasswordDialog from '../components/ChangePasswordDialog.vue'
import AppNav from '../components/AppNav.vue'

const router = useRouter()
const authStore = useAuthStore()
const changePasswordDialog = ref(null)
const uploadRef = ref(null)

const exporting = ref(false)
const importing = ref(false)
const file = ref(null)
const report = ref(null)
const XLSX_MIME = 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'

const exportPassword = ref('')
const importPassword = ref('')

onMounted(() => {
  authStore.fetchUserInfo()
})

function handleCommand(command) {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
  } else if (command === 'changePassword') {
    changePasswordDialog.value?.open()
  } else if (command === 'cloudResource') {
    router.push('/')
  } else if (command === 'personnel') {
    router.push('/personnel')
  } else if (command === 'aiSettings') {
    router.push('/ai-settings')
  } else if (command === 'aiImport') {
    router.push('/ai-import')
  } else if (command === 'publicIP') {
    router.push('/public-ip')
  }
}

async function handleExport() {
  if (exportPassword.value.length < 6) {
    ElMessage.warning('请设置至少 6 位的导出密码')
    return
  }
  exporting.value = true
  try {
    const blob = await exportAll()
    const plain = await blob.arrayBuffer()
    let bytes
    try {
      bytes = encryptXlsx(plain, exportPassword.value)
    } catch (err) {
      ElMessage.error('加密失败：' + (err?.message || '未知错误'))
      return
    }
    downloadBlob(new Blob([bytes], { type: XLSX_MIME }), `mcloud_backup_${formatDate()}.xlsx`)
    ElMessage.success('导出成功（已加密）')
  } finally {
    exporting.value = false
  }
}

function handleFileChange(uploadFile) {
  file.value = uploadFile.raw
}

function handleFileRemove() {
  file.value = null
}

async function handleImport() {
  if (!file.value) return
  const name = file.value.name.toLowerCase()
  if (!name.endsWith('.xlsx')) {
    ElMessage.warning('仅支持 .xlsx 文件')
    return
  }
  try {
    await ElMessageBox.confirm(
      '导入仅新增与更新（不删除数据），任一行失败将整体回滚。是否继续？',
      '确认导入',
      { type: 'warning', confirmButtonText: '继续导入', cancelButtonText: '取消' }
    )
  } catch {
    return
  }

  importing.value = true
  report.value = null
  try {
    // 加密 Excel 先在前端解密（密码不正确时直接提示，不上传）；未加密文件原样上传
    let uploadFile = file.value
    if (importPassword.value) {
      try {
        const bytes = await decryptXlsx(await file.value.arrayBuffer(), importPassword.value)
        uploadFile = new File([bytes], file.value.name, { type: XLSX_MIME })
      } catch (err) {
        ElMessage.error('解密失败：' + (err?.message || '密码错误'))
        return
      }
    }
    const res = await importAll(uploadFile)
    // 拦截器返回响应体 {code, message, data}，导入报告在 data 内
    report.value = res.data
    if (report.value?.committed) {
      ElMessage.success('导入成功')
    }
    uploadRef.value?.clearFiles()
    file.value = null
    importPassword.value = ''
  } catch {
    // 结构性/执行错误已由拦截器弹出 ElMessage
  } finally {
    importing.value = false
  }
}
</script>

<style scoped>
.data-backup {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f5f7fa;
}

.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 60px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
}

.header-left h1 {
  display: inline;
  font-size: 18px;
  margin: 0 8px 0 0;
}

.header-left span {
  color: #909399;
  font-size: 13px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-info {
  color: #606266;
  font-size: 14px;
}

.main-content {
  flex: 1;
  padding: 24px;
  max-width: 1000px;
  width: 100%;
  margin: 0 auto;
  box-sizing: border-box;
}

.action-row {
  margin-top: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.report {
  margin-top: 20px;
}

.error-title {
  margin: 16px 0 8px;
  color: #f56c6c;
  font-size: 14px;
  font-weight: 600;
}
</style>
