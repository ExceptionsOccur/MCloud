<template>
  <div class="domain-ledger">
    <el-header class="app-header">
      <div class="header-left">
        <h1>MCloud</h1>
        <span>云平台资产管理</span>
      </div>
      <div class="header-center">
        <router-link
          to="/"
          class="nav-tab"
          :class="{ active: $route.path === '/' }"
        >
          主机管理
        </router-link>
        <router-link
          to="/statistics"
          class="nav-tab"
          :class="{ active: $route.path === '/statistics' }"
        >
          资源统计
        </router-link>
        <router-link
          to="/ip-statistics"
          class="nav-tab"
          :class="{ active: $route.path === '/ip-statistics' }"
        >
          IP统计
        </router-link>
        <router-link
          to="/business-statistics"
          class="nav-tab"
          :class="{ active: $route.path === '/business-statistics' }"
        >
          业务统计
        </router-link>
        <router-link
          to="/zero-trust"
          class="nav-tab"
          :class="{ active: $route.path === '/zero-trust' }"
        >
          零信任
        </router-link>
        <router-link
          to="/domain-ledger"
          class="nav-tab"
          :class="{ active: $route.path === '/domain-ledger' }"
        >
          域名
        </router-link>
      </div>
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
      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="搜索域名 / 公网IP / 运营商 / 主机 / 备注"
          clearable
          style="width: 280px"
          @keyup.enter="loadDomains"
          @clear="loadDomains"
        />
        <el-button
          type="primary"
          @click="loadDomains"
        >
          查询
        </el-button>
        <el-button
          type="primary"
          @click="openDialog()"
        >
          新增域名
        </el-button>
        <el-button
          type="primary"
          plain
          @click="domainBatchDialog?.open()"
        >
          批量添加
        </el-button>
      </div>

      <el-table
        v-loading="loading"
        :data="domains"
        border
        stripe
      >
        <el-table-column
          label="#"
          width="60"
        >
          <template #default="{ $index }">
            {{ $index + 1 }}
          </template>
        </el-table-column>
        <el-table-column
          prop="domain"
          label="域名"
          min-width="200"
          show-overflow-tooltip
        />
        <el-table-column
          prop="public_ip"
          label="解析公网IP"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.public_ip || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="isp"
          label="运营商"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.isp || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          label="内网主机"
          min-width="180"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <span v-if="row.host">
              {{ row.host.name }}
              <span class="host-ip">({{ row.host.private_ip }})</span>
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column
          label="主机端口"
          width="100"
          align="center"
        >
          <template #default="{ row }">
            {{ row.host_port || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          label="备注"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.remark || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          label="操作"
          width="140"
          align="center"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              type="primary"
              link
              size="small"
              @click="openDialog(row)"
            >
              编辑
            </el-button>
            <el-popconfirm
              :title="`确定删除域名 ${row.domain}？`"
              @confirm="handleDelete(row.id)"
            >
              <template #reference>
                <el-button
                  type="danger"
                  link
                  size="small"
                >
                  删除
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="480px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item
          label="域名"
          prop="domain"
        >
          <el-input
            v-model="form.domain"
            placeholder="请输入域名"
            maxlength="255"
          />
        </el-form-item>
        <el-form-item
          label="解析公网IP"
          prop="public_ip"
        >
          <el-input
            v-model="form.public_ip"
            placeholder="请输入解析公网IP"
            maxlength="45"
          />
        </el-form-item>
        <el-form-item
          label="运营商"
          prop="isp"
        >
          <el-input
            v-model="form.isp"
            placeholder="请输入运营商"
            maxlength="128"
          />
        </el-form-item>
        <el-form-item
          label="内网主机"
          prop="host_id"
        >
          <el-select
            v-model="form.host_id"
            placeholder="请选择内网主机（选填）"
            filterable
            clearable
            style="width: 100%"
            :loading="hostsLoading"
          >
            <el-option
              v-for="h in hosts"
              :key="h.id"
              :label="`${h.name} (${h.private_ip})`"
              :value="h.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          label="主机端口"
          prop="host_port"
        >
          <el-input-number
            v-model="form.host_port"
            :min="0"
            :max="65535"
            controls-position="right"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item
          label="备注"
          prop="remark"
        >
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入备注"
            maxlength="500"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">
          取消
        </el-button>
        <el-button
          type="primary"
          :loading="submitting"
          @click="handleSubmit"
        >
          {{ submitting ? '提交中...' : '确定' }}
        </el-button>
      </template>
    </el-dialog>

    <ChangePasswordDialog ref="changePasswordDialog" />
    <CloudResourceDialog ref="cloudResourceDialog" />
    <DomainBatchAddDialog ref="domainBatchDialog" />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import { getDomains, createDomain, updateDomain, deleteDomain } from '../api/domain'
import { getHosts } from '../api/host'
import ChangePasswordDialog from '../components/ChangePasswordDialog.vue'
import CloudResourceDialog from '../components/CloudResourceDialog.vue'
import DomainBatchAddDialog from '../components/DomainBatchAddDialog.vue'

const router = useRouter()
const authStore = useAuthStore()

const changePasswordDialog = ref(null)
const cloudResourceDialog = ref(null)
const domainBatchDialog = ref(null)

const domains = ref([])
const loading = ref(false)
const keyword = ref('')
const dialogVisible = ref(false)
const submitting = ref(false)
const editId = ref(null)
const formRef = ref(null)
const hosts = ref([])
const hostsLoading = ref(false)

const form = reactive({
  domain: '',
  public_ip: '',
  isp: '',
  host_id: null,
  host_port: 0,
  remark: ''
})

const rules = {
  domain: [{ required: true, message: '请输入域名', trigger: 'blur' }]
}

const dialogTitle = computed(() => (editId.value ? '编辑域名' : '新增域名'))

onMounted(() => {
  authStore.fetchUserInfo()
  loadDomains()
  loadHosts()
  window.addEventListener('ledger-batch-done', loadDomains)
})

import { onUnmounted } from 'vue'
onUnmounted(() => {
  window.removeEventListener('ledger-batch-done', loadDomains)
})

async function loadDomains() {
  loading.value = true
  try {
    const res = await getDomains(keyword.value)
    if (res.code === 0) {
      domains.value = res.data || []
    }
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  } finally {
    loading.value = false
  }
}

async function loadHosts() {
  hostsLoading.value = true
  try {
    const res = await getHosts({ page: 1, page_size: 100 })
    if (res.code === 0) {
      hosts.value = res.data?.hosts || []
    }
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  } finally {
    hostsLoading.value = false
  }
}

function openDialog(row) {
  editId.value = row?.id || null
  form.domain = row?.domain || ''
  form.public_ip = row?.public_ip || ''
  form.isp = row?.isp || ''
  form.host_id = row?.host_id || null
  form.host_port = row?.host_port || 0
  form.remark = row?.remark || ''
  dialogVisible.value = true
  formRef.value?.clearValidate()
}

async function handleSubmit() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const payload = {
      domain: form.domain.trim(),
      public_ip: form.public_ip.trim(),
      isp: form.isp.trim(),
      host_id: form.host_id || 0,
      host_port: form.host_port || 0,
      remark: form.remark.trim()
    }
    const res = editId.value
      ? await updateDomain(editId.value, payload)
      : await createDomain(payload)
    if (res.code === 0) {
      ElMessage.success(editId.value ? '修改成功' : '新增成功')
      dialogVisible.value = false
      await loadDomains()
    }
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id) {
  try {
    const res = await deleteDomain(id)
    if (res.code === 0) {
      ElMessage.success('删除成功')
      await loadDomains()
    }
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  }
}

function handleCommand(command) {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
  } else if (command === 'changePassword') {
    changePasswordDialog.value?.open()
  } else if (command === 'cloudResource') {
    cloudResourceDialog.value?.open()
  } else if (command === 'personnel') {
    router.push('/personnel')
  } else if (command === 'publicIP') {
    router.push('/public-ip')
  }
}
</script>

<style scoped>
.domain-ledger {
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

.header-center {
  display: flex;
  gap: 4px;
}

.nav-tab {
  padding: 6px 16px;
  font-size: 14px;
  color: #606266;
  text-decoration: none;
  border-radius: 4px;
  transition: all 0.2s;
}

.nav-tab:hover {
  color: #409eff;
  background: #ecf5ff;
}

.nav-tab.active {
  color: #409eff;
  font-weight: 600;
  background: #ecf5ff;
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

.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.host-ip {
  color: #909399;
  font-size: 12px;
}
</style>
