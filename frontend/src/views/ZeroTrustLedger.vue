<template>
  <div class="zero-trust-ledger">
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
              <el-dropdown-item command="personnel">
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
          placeholder="搜索申请单位 / 账户名 / 联系方式 / 系统名称 / 主机 / 公网IP / 出口位置 / 备注"
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
          新增申请
        </el-button>
        <el-button
          type="primary"
          plain
          @click="zeroTrustBatchDialog?.open()"
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
          prop="apply_unit"
          label="申请单位"
          min-width="140"
          show-overflow-tooltip
        />
        <el-table-column
          prop="account_name"
          label="账户名"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column
          label="申请人联系方式"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.contact || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          label="申请主机"
          min-width="180"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <div v-if="row.targets?.length">
              <div
                v-for="(t, ti) in row.targets"
                :key="ti"
              >
                {{ t.host_name || `#${t.host_id}` }}
                <span
                  v-if="t.private_ip"
                  class="host-ip"
                >({{ t.private_ip }})</span>
              </div>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column
          label="申请端口"
          min-width="110"
          align="center"
        >
          <template #default="{ row }">
            <el-tag
              v-for="(t, ti) in row.targets || []"
              :key="'p' + ti"
              size="small"
              class="port-tag"
            >
              {{ t.port }}
            </el-tag>
            <span v-if="!row.targets?.length">-</span>
          </template>
        </el-table-column>
        <el-table-column
          label="系统名称"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.system_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          label="公网IP"
          min-width="130"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.public_ip || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          label="接入地区"
          min-width="100"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.exit_location || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          label="申请时间"
          width="170"
        >
          <template #default="{ row }">
            {{ formatTime(row.apply_time) }}
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
              :title="`确定删除账户 ${row.account_name} 的申请记录？`"
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
      width="520px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="110px"
      >
        <el-form-item
          label="申请单位"
          prop="apply_unit"
        >
          <el-input
            v-model="form.apply_unit"
            placeholder="请输入申请单位"
            maxlength="128"
          />
        </el-form-item>
        <el-form-item
          label="账户名"
          prop="account_name"
        >
          <el-input
            v-model="form.account_name"
            placeholder="请输入账户名"
            maxlength="64"
          />
        </el-form-item>
        <el-form-item
          label="联系方式"
          prop="contact"
        >
          <el-input
            v-model="form.contact"
            placeholder="请输入申请人联系方式"
            maxlength="64"
          />
        </el-form-item>
        <el-form-item
          label="接入主机与端口"
          required
        >
          <div class="target-rows">
            <div
              v-for="(t, idx) in form.targets"
              :key="idx"
              class="target-row"
            >
              <el-select
                v-model="t.host_id"
                placeholder="选择主机"
                filterable
                clearable
                class="target-host"
                :loading="hostsLoading"
              >
                <el-option
                  v-for="h in hosts"
                  :key="h.id"
                  :label="`${h.name} (${h.private_ip})`"
                  :value="h.id"
                />
              </el-select>
              <el-input-number
                v-model="t.port"
                :min="1"
                :max="65535"
                controls-position="right"
                class="target-port"
              />
              <el-button
                link
                type="danger"
                :disabled="form.targets.length <= 1"
                @click="form.targets.splice(idx, 1)"
              >
                删除
              </el-button>
            </div>
          </div>
          <el-button
            link
            type="primary"
            @click="addTargetRow"
          >
            + 添加一组
          </el-button>
        </el-form-item>
        <el-form-item
          label="系统名称"
          prop="system_name"
        >
          <el-input
            v-model="form.system_name"
            placeholder="请输入系统名称（选填）"
            maxlength="128"
          />
        </el-form-item>
        <el-form-item
          label="公网IP"
          prop="public_ip"
        >
          <el-select
            v-model="form.public_ip"
            placeholder="从公网IP资源池选择（选填）"
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
          <div
            v-if="selectedExitLocation"
            class="exit-hint"
          >
            接入地区：{{ selectedExitLocation }}
          </div>
        </el-form-item>
        <el-form-item
          label="申请时间"
          prop="apply_time"
        >
          <el-date-picker
            v-model="form.apply_time"
            type="datetime"
            placeholder="请选择申请时间"
            style="width: 100%"
            value-format="YYYY-MM-DDTHH:mm:ss"
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
    <ZeroTrustBatchAddDialog ref="zeroTrustBatchDialog" />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import { getZeroTrusts, createZeroTrust, updateZeroTrust, deleteZeroTrust } from '../api/zero_trust'
import { getHosts } from '../api/host'
import { getPublicIPs } from '../api/public_ip'
import { formatTime, toLocalPickerValue, toRFC3339 } from '../utils'
import ChangePasswordDialog from '../components/ChangePasswordDialog.vue'
import CloudResourceDialog from '../components/CloudResourceDialog.vue'
import ZeroTrustBatchAddDialog from '../components/ZeroTrustBatchAddDialog.vue'

const router = useRouter()
const authStore = useAuthStore()

const changePasswordDialog = ref(null)
const cloudResourceDialog = ref(null)
const zeroTrustBatchDialog = ref(null)

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
  apply_unit: '',
  account_name: '',
  contact: '',
  targets: [{ host_id: null, port: 22 }],
  system_name: '',
  public_ip: '',
  apply_time: '',
  remark: ''
})

const rules = {
  apply_unit: [{ required: true, message: '请输入申请单位', trigger: 'blur' }],
  account_name: [{ required: true, message: '请输入账户名', trigger: 'blur' }],
  apply_time: [{ required: true, message: '请选择申请时间', trigger: 'change' }]
}

const selectedExitLocation = computed(() => {
  if (!form.public_ip) return ''
  return publicIPs.value.find(ip => ip.ip === form.public_ip)?.exit_location || ''
})

const dialogTitle = computed(() => (editId.value ? '编辑申请' : '新增申请'))

onMounted(() => {
  authStore.fetchUserInfo()
  loadList()
  loadHosts()
  loadPublicIPs()
  window.addEventListener('ledger-batch-done', loadList)
})

import { onUnmounted } from 'vue'
onUnmounted(() => {
  window.removeEventListener('ledger-batch-done', loadList)
})

async function loadList() {
  loading.value = true
  try {
    const res = await getZeroTrusts(keyword.value)
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

function addTargetRow() {
  form.targets.push({ host_id: null, port: 22 })
}

function openDialog(row) {
  editId.value = row?.id || null
  form.apply_unit = row?.apply_unit || ''
  form.account_name = row?.account_name || ''
  form.contact = row?.contact || ''
  form.targets = (row?.targets || []).map(t => ({ host_id: t.host_id, port: t.port }))
  if (!form.targets.length) form.targets = [{ host_id: null, port: 22 }]
  form.system_name = row?.system_name || ''
  form.public_ip = row?.public_ip || ''
  form.apply_time = row?.apply_time ? toLocalPickerValue(row.apply_time) : ''
  form.remark = row?.remark || ''
  dialogVisible.value = true
  formRef.value?.clearValidate()
}

async function handleSubmit() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  for (let i = 0; i < form.targets.length; i++) {
    const t = form.targets[i]
    if (!t.host_id) {
      ElMessage.warning(`第 ${i + 1} 组请选择主机`)
      return
    }
    if (!t.port || t.port < 1 || t.port > 65535) {
      ElMessage.warning(`第 ${i + 1} 组端口必须在 1-65535 之间`)
      return
    }
  }

  submitting.value = true
  try {
    const payload = {
      apply_unit: form.apply_unit.trim(),
      account_name: form.account_name.trim(),
      contact: form.contact.trim(),
      targets: form.targets.map(t => ({ host_id: t.host_id, port: t.port })),
      system_name: form.system_name.trim(),
      public_ip: form.public_ip || '',
      apply_time: toRFC3339(form.apply_time),
      remark: form.remark.trim()
    }
    const res = editId.value
      ? await updateZeroTrust(editId.value, payload)
      : await createZeroTrust(payload)
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
    const res = await deleteZeroTrust(id)
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
.zero-trust-ledger {
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
  margin-right: 4px;
}

.target-rows {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.target-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.target-host {
  flex: 1;
}

.target-port {
  width: 130px;
}

.exit-hint {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
  line-height: 1.4;
}
</style>
