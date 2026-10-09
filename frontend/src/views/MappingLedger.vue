<template>
  <div class="mapping-ledger">
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
          to="/mapping-ledger"
          class="nav-tab"
          :class="{ active: $route.path === '/mapping-ledger' }"
        >
          映射
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
              <el-dropdown-item command="dataBackup">
                数据备份
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
          placeholder="搜索公网IP / 内网主机 / 端口 / 域名 / 运营商 / 出口位置"
          clearable
          style="width: 320px"
          @keyup.enter="loadList"
          @clear="loadList"
        />
        <el-button
          type="primary"
          @click="loadList"
        >
          查询
        </el-button>
        <el-button
          type="primary"
          @click="openDialog()"
        >
          新增映射
        </el-button>
        <el-button
          type="primary"
          plain
          @click="batchDialog?.open()"
        >
          批量添加
        </el-button>
      </div>

      <el-table
        v-loading="loading"
        :data="records"
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
          prop="public_ip"
          label="公网IP"
          min-width="130"
          show-overflow-tooltip
        />
        <el-table-column
          label="内网主机"
          min-width="170"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <span v-if="row.host">
              {{ row.host.name }}
              <span class="host-ip">({{ row.host.private_ip }})</span>
            </span>
            <span v-else>#{{ row.host_id }}</span>
          </template>
        </el-table-column>
        <el-table-column
          label="外网端口"
          min-width="140"
        >
          <template #default="{ row }">
            <el-tag
              v-for="p in splitPorts(row.external_ports)"
              :key="'e' + p"
              size="small"
              class="port-tag"
            >
              {{ p }}
            </el-tag>
            <span v-if="!splitPorts(row.external_ports).length">-</span>
          </template>
        </el-table-column>
        <el-table-column
          label="内网端口"
          min-width="140"
        >
          <template #default="{ row }">
            <el-tag
              v-for="p in splitPorts(row.internal_ports)"
              :key="'i' + p"
              size="small"
              type="success"
              class="port-tag"
            >
              {{ p }}
            </el-tag>
            <span v-if="!splitPorts(row.internal_ports).length">-</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="domain"
          label="域名"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.domain || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="isp"
          label="运营商"
          min-width="100"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.isp || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="exit_location"
          label="出口位置"
          min-width="100"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.exit_location || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="remark"
          label="备注"
          min-width="120"
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
              title="确定删除该映射记录？"
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
      width="640px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item
          label="公网IP"
          prop="public_ip"
        >
          <el-select
            v-model="form.public_ip"
            placeholder="请从公网IP资源池选择"
            filterable
            clearable
            style="width: 100%"
            :loading="publicIPsLoading"
          >
            <el-option
              v-for="ip in publicIPs"
              :key="ip.ip"
              :label="ip.isp ? `${ip.ip}（${ip.isp}）` : ip.ip"
              :value="ip.ip"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          label="内网主机"
          prop="host_id"
        >
          <el-select
            v-model="form.host_id"
            placeholder="请选择内网主机"
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
          label="外网端口"
          prop="external_ports"
        >
          <el-input
            v-model="form.external_ports"
            placeholder="多个端口用逗号分隔，如 80,443"
          />
        </el-form-item>
        <el-form-item
          label="内网端口"
          prop="internal_ports"
        >
          <el-input
            v-model="form.internal_ports"
            placeholder="与外网端口数量一致、顺序对应，如 8080,8443"
          />
        </el-form-item>
        <el-form-item label="端口对应">
          <div class="port-preview">
            <template v-if="portPairs.length">
              <el-tag
                v-for="(pair, idx) in portPairs"
                :key="idx"
                size="small"
              >
                外网 {{ pair[0] }} → 内网 {{ pair[1] }}
              </el-tag>
            </template>
            <span
              v-else
              class="muted"
            >填写两端端口后预览对应关系</span>
          </div>
        </el-form-item>
        <el-form-item
          label="域名"
          prop="domain"
        >
          <el-input
            v-model="form.domain"
            placeholder="可选"
            maxlength="255"
          />
        </el-form-item>
        <el-form-item
          label="备注"
          prop="remark"
        >
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="2"
            placeholder="可选"
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

    <MappingBatchAddDialog ref="batchDialog" />
    <ChangePasswordDialog ref="changePasswordDialog" />
    <CloudResourceDialog ref="cloudResourceDialog" />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import {
  getPortMappings,
  createPortMapping,
  updatePortMapping,
  deletePortMapping
} from '../api/port_mapping'
import { getHosts } from '../api/host'
import { getPublicIPs } from '../api/public_ip'
import MappingBatchAddDialog from '../components/MappingBatchAddDialog.vue'
import ChangePasswordDialog from '../components/ChangePasswordDialog.vue'
import CloudResourceDialog from '../components/CloudResourceDialog.vue'

const router = useRouter()
const authStore = useAuthStore()

const changePasswordDialog = ref(null)
const cloudResourceDialog = ref(null)
const batchDialog = ref(null)

const records = ref([])
const loading = ref(false)
const keyword = ref('')
const dialogVisible = ref(false)
const submitting = ref(false)
const editId = ref(null)
const formRef = ref(null)
const hosts = ref([])
const hostsLoading = ref(false)
const publicIPs = ref([])
const publicIPsLoading = ref(false)

const form = reactive({
  public_ip: '',
  host_id: null,
  external_ports: '',
  internal_ports: '',
  domain: '',
  remark: ''
})

const rules = {
  public_ip: [{ required: true, message: '请输入公网IP', trigger: 'blur' }],
  host_id: [{ required: true, message: '请选择内网主机', trigger: 'change' }],
  external_ports: [{ required: true, message: '请输入外网端口', trigger: 'blur' }],
  internal_ports: [{ required: true, message: '请输入内网端口', trigger: 'blur' }]
}

const dialogTitle = computed(() => (editId.value ? '编辑映射' : '新增映射'))

function splitPorts(val) {
  if (!val) return []
  return String(val).split(',').map(s => s.trim()).filter(Boolean)
}

const portPairs = computed(() => {
  const ext = splitPorts(form.external_ports)
  const intl = splitPorts(form.internal_ports)
  const n = Math.min(ext.length, intl.length)
  const pairs = []
  for (let i = 0; i < n; i++) pairs.push([ext[i], intl[i]])
  return pairs
})

onMounted(() => {
  authStore.fetchUserInfo()
  loadList()
  loadHosts()
  loadPublicIPs()
  window.addEventListener('ledger-batch-done', loadList)
})

onUnmounted(() => {
  window.removeEventListener('ledger-batch-done', loadList)
})

async function loadPublicIPs() {
  publicIPsLoading.value = true
  try {
    const res = await getPublicIPs('')
    if (res.code === 0) {
      publicIPs.value = res.data || []
    }
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  } finally {
    publicIPsLoading.value = false
  }
}

async function loadList() {
  loading.value = true
  try {
    const res = await getPortMappings(keyword.value)
    if (res.code === 0) {
      records.value = res.data || []
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
  form.public_ip = row?.public_ip || ''
  form.host_id = row?.host_id || null
  form.external_ports = row?.external_ports || ''
  form.internal_ports = row?.internal_ports || ''
  form.domain = row?.domain || ''
  form.remark = row?.remark || ''
  dialogVisible.value = true
  formRef.value?.clearValidate()
}

async function handleSubmit() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  const ext = splitPorts(form.external_ports)
  const intl = splitPorts(form.internal_ports)
  if (ext.length !== intl.length) {
    ElMessage.warning(`外网端口与内网端口数量必须一致（外网 ${ext.length} 个，内网 ${intl.length} 个）`)
    return
  }

  submitting.value = true
  try {
    const payload = {
      public_ip: form.public_ip,
      host_id: form.host_id,
      external_ports: ext.join(','),
      internal_ports: intl.join(','),
      domain: form.domain.trim(),
      remark: form.remark.trim()
    }
    const res = editId.value
      ? await updatePortMapping(editId.value, payload)
      : await createPortMapping(payload)
    if (res.code === 0) {
      ElMessage.success(editId.value ? '修改成功' : '新增成功')
      dialogVisible.value = false
      await loadList()
    }
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id) {
  try {
    const res = await deletePortMapping(id)
    if (res.code === 0) {
      ElMessage.success('删除成功')
      await loadList()
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
  } else if (command === 'dataBackup') {
    router.push('/data-backup')
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
.mapping-ledger {
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

.port-tag {
  margin: 2px 4px 2px 0;
}

.port-preview {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.muted {
  color: #909399;
  font-size: 12px;
}
</style>
