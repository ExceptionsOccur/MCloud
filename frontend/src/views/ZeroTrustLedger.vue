<template>
  <div class="zero-trust-ledger">
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
      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="搜索申请单位 / 账户名 / 账号姓名 / 申请人 / 账户联系方式 / 系统名称 / 主机 / 接入地区 / 备注"
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
        @row-click="handleRowClick"
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
          min-width="100"
          show-overflow-tooltip
        />
        <el-table-column
          prop="account_name_zh"
          label="账号姓名"
          min-width="70"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.account_name_zh || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          label="账户联系方式"
          min-width="90"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.contact || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          label="申请资源"
          min-width="220"
        >
          <template #default="{ row }">
            <template v-if="targetGroupsOf(row).length">
              <div class="target-hosts">
                <el-popover
                  v-for="h in targetGroupsOf(row).slice(0, 3)"
                  :key="h.host_id"
                  placement="auto"
                  :width="260"
                  trigger="click"
                  popper-class="target-popper"
                  :popper-options="targetPopperOptions"
                >
                  <template #reference>
                    <el-tag
                      size="small"
                      class="target-tag"
                    >
                      {{ h.label }}
                    </el-tag>
                  </template>
                  <div class="target-detail-scroll">
                    <div class="target-detail-title">
                      {{ h.label }}
                    </div>
                    <div class="target-detail-ports">
                      <el-tag
                        v-for="(p, pi) in h.ports"
                        :key="pi"
                        size="small"
                        class="port-tag"
                      >
                        {{ p.ports }}/{{ p.protocol }}
                      </el-tag>
                    </div>
                  </div>
                </el-popover>
                <el-popover
                  v-if="targetGroupsOf(row).length > 3"
                  placement="auto"
                  :width="300"
                  trigger="click"
                  popper-class="target-popper"
                  :popper-options="targetPopperOptions"
                >
                  <template #reference>
                    <el-tag
                      size="small"
                      type="info"
                      class="target-tag"
                    >
                      +{{ targetGroupsOf(row).length - 3 }} 台
                    </el-tag>
                  </template>
                  <div class="target-detail-scroll">
                    <div
                      v-for="h in targetGroupsOf(row).slice(3)"
                      :key="h.host_id"
                      class="target-detail-row"
                    >
                      <div class="target-detail-title">
                        {{ h.label }}
                      </div>
                      <div class="target-detail-ports">
                        <el-tag
                          v-for="(p, pi) in h.ports"
                          :key="pi"
                          size="small"
                          class="port-tag"
                        >
                          {{ p.ports }}/{{ p.protocol }}
                        </el-tag>
                      </div>
                    </div>
                  </div>
                </el-popover>
              </div>
            </template>
            <span v-else>-</span>
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
          label="接入地区"
          min-width="80"
        >
          <template #default="{ row }">
            <div
              v-if="splitLocations(row.exit_location).length"
            >
              <el-tag
                v-for="(loc, li) in splitLocations(row.exit_location)"
                :key="li"
                size="small"
                type="info"
                class="loc-tag"
              >
                {{ loc }}
              </el-tag>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column
          label="申请时间"
          width="100"
        >
          <template #default="{ row }">
            {{ row.apply_time || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="applicant"
          label="申请人"
          min-width="70"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.applicant || '-' }}
          </template>
        </el-table-column>
        <el-table-column
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
          width="180"
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
            <el-button
              type="primary"
              link
              size="small"
              @click="openDialog(row, true)"
            >
              复制
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
      width="550px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
      >
        <el-form-item
          label="申请单位"
          prop="apply_unit"
          label-width="120px"
        >
          <el-input
            v-model="form.apply_unit"
            placeholder="请输入申请单位"
            maxlength="128"
            label-width="120px"
          />
        </el-form-item>
        <el-form-item
          label="账户名"
          prop="account_name"
          label-width="120px"
        >
          <el-input
            v-model="form.account_name"
            placeholder="请输入账户名"
            maxlength="64"
          />
        </el-form-item>
        <el-form-item
          label="账号姓名"
          prop="account_name_zh"
        >
          <el-input
            v-model="form.account_name_zh"
            placeholder="请输入账号姓名（选填）"
            maxlength="64"
          />
        </el-form-item>
        <el-form-item
          label="申请人"
          prop="applicant"
        >
          <el-input
            v-model="form.applicant"
            placeholder="请输入申请人（选填）"
            maxlength="64"
          />
        </el-form-item>
        <el-form-item
          label="账户联系方式"
          prop="contact"
        >
          <el-input
            v-model="form.contact"
            placeholder="请输入账户联系方式"
            maxlength="64"
          />
        </el-form-item>
        <el-form-item
          label="接入主机与端口"
          required
          label-width="120px"
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
              <el-input
                v-model="t.ports"
                placeholder="端口或范围，如 22、1000-10005、22,443,5432"
                class="target-port"
              />
              <el-select
                v-model="t.protocol"
                class="target-protocol"
              >
                <el-option
                  label="TCP"
                  value="tcp"
                />
                <el-option
                  label="UDP"
                  value="udp"
                />
              </el-select>
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
          label="接入地区"
          prop="exit_location"
        >
          <el-select
            v-model="form.exit_location"
            placeholder="从公网IP资源池筛选接入地区（可多选，选填）"
            filterable
            multiple
            clearable
            style="width: 100%"
            :loading="exitLocationsLoading"
          >
            <el-option
              v-for="loc in exitLocations"
              :key="loc"
              :label="loc"
              :value="loc"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          label="申请时间"
          prop="apply_time"
        >
          <el-date-picker
            v-model="form.apply_time"
            type="date"
            placeholder="请选择申请时间"
            style="width: 100%"
            value-format="YYYY-MM-DD"
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
    <ZeroTrustDetailDialog ref="zeroTrustDetailDialog" />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import { getZeroTrusts, getZeroTrustExitLocations, createZeroTrust, updateZeroTrust, deleteZeroTrust } from '../api/zero_trust'
import { getHosts } from '../api/host'
import { formatDate } from '../utils'
import ChangePasswordDialog from '../components/ChangePasswordDialog.vue'
import CloudResourceDialog from '../components/CloudResourceDialog.vue'
import ZeroTrustBatchAddDialog from '../components/ZeroTrustBatchAddDialog.vue'
import ZeroTrustDetailDialog from '../components/ZeroTrustDetailDialog.vue'

const router = useRouter()
const authStore = useAuthStore()

const changePasswordDialog = ref(null)
const cloudResourceDialog = ref(null)
const zeroTrustBatchDialog = ref(null)
const zeroTrustDetailDialog = ref(null)

// 申请资源 popover 的 popper 选项（T-065）：auto 定位 + flip 自动换方向 + preventOverflow 保证不出界
const targetPopperOptions = {
  modifiers: [
    { name: 'flip', options: { fallbackPlacements: ['top', 'bottom', 'right', 'left'] } },
    { name: 'preventOverflow', options: { padding: 8, altAxis: true } }
  ]
}

const records = ref([])
const loading = ref(false)
const keyword = ref('')
const dialogVisible = ref(false)
const submitting = ref(false)
const editId = ref(null)
const copyMode = ref(false)
const formRef = ref(null)
const hosts = ref([])
const hostsLoading = ref(false)
const exitLocations = ref([])
const exitLocationsLoading = ref(false)

const form = reactive({
  apply_unit: '',
  account_name: '',
  account_name_zh: '',
  applicant: '',
  contact: '',
  targets: [{ host_id: null, ports: '22', protocol: 'tcp' }],
  system_name: '',
  exit_location: [],
  apply_time: '',
  remark: ''
})

const rules = {
  apply_unit: [{ required: true, message: '请输入申请单位', trigger: 'blur' }],
  account_name: [{ required: true, message: '请输入账户名', trigger: 'blur' }],
  apply_time: [{ required: true, message: '请选择申请时间', trigger: 'change' }]
}

const dialogTitle = computed(() => (editId.value ? '编辑申请' : (copyMode.value ? '复制新增申请' : '新增申请')))

onMounted(() => {
  authStore.fetchUserInfo()
  loadList()
  loadHosts()
  loadExitLocations()
  window.addEventListener('ledger-batch-done', loadList)
})

import { onUnmounted } from 'vue'
import AppNav from '../components/AppNav.vue'
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
    // 后端 page_size 上限 100，全量主机可能超过一页 → 循环拉取全部分页（T-065）
    const pageSize = 100
    let page = 1
    const all = []
    let total = Infinity
    while (all.length < total) {
      const res = await getHosts({ page, page_size: pageSize })
      if (res.code !== 0) break
      const batch = res.data?.hosts || []
      all.push(...batch)
      total = res.data?.total ?? all.length
      if (batch.length === 0) break
      page += 1
    }
    hosts.value = all
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  } finally {
    hostsLoading.value = false
  }
}

async function loadExitLocations() {
  exitLocationsLoading.value = true
  try {
    const res = await getZeroTrustExitLocations()
    if (res.code === 0) {
      exitLocations.value = res.data || []
    }
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  } finally {
    exitLocationsLoading.value = false
  }
}

function addTargetRow() {
  form.targets.push({ host_id: null, ports: '22', protocol: 'tcp' })
}

// splitLocations 多值接入地区字符串 → 去空数组（T-065）
function splitLocations(raw) {
  return String(raw || '')
    .split(',')
    .map(s => s.trim())
    .filter(Boolean)
}

// groupTargetsByHost 接入目标按主机聚合（T-065）：一主机一组端口，供列表「主机 tag + 端口 popover」展示
function groupTargetsByHost(targets) {
  const map = new Map()
  for (const t of targets || []) {
    if (!map.has(t.host_id)) {
      map.set(t.host_id, {
        host_id: t.host_id,
        label: t.host_name ? (t.private_ip ? `${t.host_name}(${t.private_ip})` : t.host_name) : `#${t.host_id}`,
        ports: []
      })
    }
    map.get(t.host_id).ports.push({ ports: t.ports || '', protocol: t.protocol || 'tcp' })
  }
  const list = [...map.values()]
  for (const h of list) {
    h.ports.sort((a, b) => String(a.ports).localeCompare(String(b.ports), 'zh-CN', { numeric: true }))
  }
  return list.sort((a, b) => a.label.localeCompare(b.label, 'zh-CN', { numeric: true, sensitivity: 'base' }))
}

// targetGroupsByRowId 列表每行接入目标按主机聚合的缓存（避免模板内重复计算）
const targetGroupsByRowId = computed(() => {
  const cache = {}
  for (const row of records.value) {
    cache[row.id] = groupTargetsByHost(row.targets)
  }
  return cache
})

function targetGroupsOf(row) {
  return targetGroupsByRowId.value[row.id] || []
}

function openDialog(row, copy = false) {
  copyMode.value = copy
  editId.value = copy ? null : (row?.id || null)
  form.apply_unit = row?.apply_unit || ''
  // 复制时唯一字段（账户名）置空待填
  form.account_name = copy ? '' : (row?.account_name || '')
  form.account_name_zh = row?.account_name_zh || ''
  form.applicant = row?.applicant || ''
  form.contact = row?.contact || ''
  form.targets = (row?.targets || []).map(t => ({ host_id: t.host_id, ports: t.ports || '', protocol: t.protocol || 'tcp' }))
  if (!form.targets.length) form.targets = [{ host_id: null, ports: '22' }]
  form.system_name = row?.system_name || ''
  form.exit_location = splitLocations(row?.exit_location)
  // T-082：申请时间为空（新增或历史缺值）时默认当天，有值保留原值
  form.apply_time = row?.apply_time || formatDate()
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
    if (!t.ports || !String(t.ports).trim()) {
      ElMessage.warning(`第 ${i + 1} 组请输入端口或范围表达式（如 22、1000-10005、22,443）`)
      return
    }
  }

  submitting.value = true
  try {
    const payload = {
      apply_unit: form.apply_unit.trim(),
      account_name: form.account_name.trim(),
      account_name_zh: form.account_name_zh.trim(),
      applicant: form.applicant.trim(),
      contact: form.contact.trim(),
      targets: form.targets.map(t => ({ host_id: t.host_id, ports: String(t.ports).trim(), protocol: t.protocol || 'tcp' })),
      system_name: form.system_name.trim(),
      exit_location: (form.exit_location || []).join(','),
      apply_time: form.apply_time || '',
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

// 点击数据行打开只读详情；申请资源/操作列不触发（T-066）
function handleRowClick(row, column) {
  if (column?.label === '申请资源' || column?.label === '操作') return
  zeroTrustDetailDialog.value?.open(row)
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
  } else if (command === 'aiSettings') {
    router.push('/ai-settings')
  } else if (command === 'aiImport') {
    router.push('/ai-import')
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

.target-hosts {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.target-tag {
  max-width: 100%;
  cursor: pointer;
}

.loc-tag {
  max-width: 100%;
  color: #409eff;
  border: 1px solid #d9ecff;
  background: #ecf5ff;
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

.target-protocol {
  width: 88px;
  flex-shrink: 0;
}
</style>

<style>
/* 接入资源 popover 内容在 teleport 容器内，需用非 scoped 样式（T-065） */
.target-popper {
  max-width: 360px;
}

.target-detail-scroll {
  max-height: 55vh;
  overflow-y: auto;
}

.target-detail-title {
  font-weight: 600;
  color: #303133;
  margin-bottom: 6px;
  word-break: break-all;
}

.target-detail-ports {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.target-detail-row {
  margin-top: 10px;
}

.target-detail-row:first-child {
  margin-top: 0;
}
</style>
