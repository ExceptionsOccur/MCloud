<template>
  <div class="mapping-ledger">
    <AppHeader />
    <div class="main-content">
      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="搜索公网IP / 内网主机 / 端口 / 协议 / 域名 / 运营商 / 出口位置"
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
              v-for="p in portTags(row.external_ports, row.external_protocols)"
              :key="'e' + p"
              size="small"
              class="port-tag"
            >
              {{ p }}
            </el-tag>
            <span v-if="!portTags(row.external_ports, row.external_protocols).length">-</span>
          </template>
        </el-table-column>
        <el-table-column
          label="内网端口"
          min-width="140"
        >
          <template #default="{ row }">
            <el-tag
              v-for="p in portTags(row.internal_ports, row.internal_protocols)"
              :key="'i' + p"
              size="small"
              type="success"
              class="port-tag"
            >
              {{ p }}
            </el-tag>
            <span v-if="!portTags(row.internal_ports, row.internal_protocols).length">-</span>
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
            placeholder="多个端口用逗号分隔，支持连续范围如 1000-1005；可带协议如 53/udp，缺省 tcp"
          />
        </el-form-item>
        <el-form-item
          label="内网端口"
          prop="internal_ports"
        >
          <el-input
            v-model="form.internal_ports"
            placeholder="与外网端口展开后数量一致、顺序对应，支持范围如 8080-8085；可带协议如 5353/udp"
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
              <span
                v-if="portPairTotal > portPairs.length"
                class="muted"
              >共 {{ portPairTotal }} 对，仅预览前 {{ portPairs.length }} 对</span>
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
    <MappingDetailDialog ref="mappingDetailDialog" />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  getPortMappings,
  createPortMapping,
  updatePortMapping,
  deletePortMapping
} from '../api/port_mapping'
import { getHosts } from '../api/host'
import { getPublicIPs } from '../api/public_ip'
import MappingBatchAddDialog from '../components/MappingBatchAddDialog.vue'
import MappingDetailDialog from '../components/MappingDetailDialog.vue'
import AppHeader from '../components/AppHeader.vue'

const batchDialog = ref(null)
const mappingDetailDialog = ref(null)

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

const dialogTitle = computed(() => (editId.value ? '编辑映射' : (copyMode.value ? '复制新增映射' : '新增映射')))

function splitPorts(val) {
  if (!val) return []
  return String(val).split(',').map(s => s.trim()).filter(Boolean)
}

// 列表标签：端口/协议配对展示，协议缺省 tcp（T-055）
function portTags(ports, protocols) {
  const ps = splitPorts(ports)
  const qs = splitPorts(protocols)
  return ps.map((p, i) => `${p}/${qs[i] || 'tcp'}`)
}

// 解析端口输入：逗号分隔，token 支持「端口」「起始-结束」及可选 /tcp|udp（缺省 tcp，T-055）
function parsePortTokens(val, field) {
  const tokens = splitPorts(val)
  const ports = []
  const protocols = []
  for (const t of tokens) {
    const m = t.match(/^(\d{1,5})(?:-(\d{1,5}))?(?:\/(tcp|udp))?$/i)
    if (!m) {
      throw new Error(`${field}中的“${t}”格式不正确（应为 端口、起始-结束，可加 /tcp|udp）`)
    }
    const a = Number(m[1])
    const b = m[2] ? Number(m[2]) : a
    if (a < 1 || b > 65535 || a > b) {
      throw new Error(`${field}中的“${t}”端口范围不合法（1-65535 且起始 ≤ 结束）`)
    }
    ports.push(m[2] ? `${a}-${b}` : `${a}`)
    protocols.push((m[3] || 'tcp').toLowerCase())
  }
  return { ports, protocols }
}

// 表单回填：端口列表 + 协议列表 → 「端口/协议」token 串
function joinPortTokens(ports, protocols) {
  const ps = splitPorts(ports)
  const qs = splitPorts(protocols)
  return ps.map((p, i) => `${p}/${qs[i] || 'tcp'}`).join(',')
}

// 端口 token 展开后的实际端口数（单口 1，A-B 为 B-A+1）
function expandCount(ports) {
  let n = 0
  for (const t of ports) {
    const m = t.match(/^(\d+)-(\d+)$/)
    n += m ? Number(m[2]) - Number(m[1]) + 1 : 1
  }
  return n
}

// 预览用尽力展开（跳过编辑中/非法 token，范围 token 展开为逐端口带协议）
function expandForPreview(val) {
  const out = []
  for (const t of splitPorts(val)) {
    const m = t.match(/^(\d{1,5})(?:-(\d{1,5}))?(?:\/(tcp|udp))?$/i)
    if (!m) {
      out.push(t)
      continue
    }
    const proto = (m[3] || 'tcp').toLowerCase()
    if (m[2]) {
      const a = Number(m[1])
      const b = Number(m[2])
      if (a < 1 || b > 65535 || a > b) {
        out.push(t)
        continue
      }
      for (let p = a; p <= b; p++) out.push(`${p}/${proto}`)
    } else {
      out.push(`${m[1]}/${proto}`)
    }
  }
  return out
}

// 预览上限：大范围（如 1-65535）只渲染前 N 对并给出总对数
const PORT_PAIR_PREVIEW_LIMIT = 100

const portPairTotal = computed(() =>
  Math.min(
    expandForPreview(form.external_ports).length,
    expandForPreview(form.internal_ports).length
  )
)

const portPairs = computed(() => {
  const ext = expandForPreview(form.external_ports)
  const intl = expandForPreview(form.internal_ports)
  const n = Math.min(ext.length, intl.length, PORT_PAIR_PREVIEW_LIMIT)
  const pairs = []
  for (let i = 0; i < n; i++) pairs.push([ext[i], intl[i]])
  return pairs
})

onMounted(() => {
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

function openDialog(row, copy = false) {
  copyMode.value = copy
  editId.value = copy ? null : (row?.id || null)
  // 复制时唯一键相关字段（公网IP、内网主机、外网/内网端口）置空待填，域名/备注保留
  form.public_ip = copy ? '' : (row?.public_ip || '')
  form.host_id = copy ? null : (row?.host_id || null)
  form.external_ports = copy ? '' : joinPortTokens(row?.external_ports, row?.external_protocols)
  form.internal_ports = copy ? '' : joinPortTokens(row?.internal_ports, row?.internal_protocols)
  form.domain = row?.domain || ''
  form.remark = row?.remark || ''
  dialogVisible.value = true
  formRef.value?.clearValidate()
}

async function handleSubmit() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  let ext
  let intl
  try {
    ext = parsePortTokens(form.external_ports, '外网端口')
    intl = parsePortTokens(form.internal_ports, '内网端口')
  } catch (e) {
    ElMessage.warning(e.message)
    return
  }
  const extCount = expandCount(ext.ports)
  const intlCount = expandCount(intl.ports)
  if (extCount !== intlCount) {
    ElMessage.warning(`外网端口与内网端口数量必须一致（外网 ${extCount} 个，内网 ${intlCount} 个）`)
    return
  }

  submitting.value = true
  try {
    const payload = {
      public_ip: form.public_ip,
      host_id: form.host_id,
      external_ports: ext.ports.join(','),
      external_protocols: ext.protocols.join(','),
      internal_ports: intl.ports.join(','),
      internal_protocols: intl.protocols.join(','),
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

// 点击数据行打开只读详情；操作列不触发（T-066）
function handleRowClick(row, column) {
  if (column?.label === '操作') return
  mappingDetailDialog.value?.open(row)
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

</script>

<style scoped>
.mapping-ledger {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
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
