<template>
  <div
    v-loading="loading"
    class="audit-timeline"
  >
    <el-empty
      v-if="!loading && !items.length"
      description="暂无变更记录"
      :image-size="60"
    />
    <el-timeline v-else>
      <el-timeline-item
        v-for="item in items"
        :key="item.id"
        :timestamp="formatTime(item.created_at)"
        :type="actionType(item.action)"
        placement="top"
      >
        <div class="timeline-head">
          <el-tag
            size="small"
            :type="actionType(item.action)"
          >
            {{ actionLabel(item.action) }}
          </el-tag>
          <span class="operator">{{ item.operator_name || '-' }}</span>
        </div>

        <div
          v-if="rowsOf(item).length"
          class="diff-block"
        >
          <div
            v-for="row in rowsOf(item)"
            :key="row.key"
            class="diff-row"
          >
            <span class="diff-key">{{ row.label }}</span>
            <template v-if="row.changed">
              <span class="diff-before">{{ row.before }}</span>
              <span class="diff-arrow">→</span>
            </template>
            <span class="diff-after">{{ row.after }}</span>
          </div>
        </div>
        <div
          v-else
          class="diff-empty"
        >
          无字段变化
        </div>
      </el-timeline-item>
    </el-timeline>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { getResourceHistory } from '../api/audit'
import { formatTime } from '../utils'

const props = defineProps({
  resourceType: { type: String, required: true },
  resourceId: { type: [Number, String], default: null }
})

const loading = ref(false)
const items = ref([])

const ACTION_LABELS = {
  create: '新增',
  update: '修改',
  delete: '删除',
  batch_update: '批量修改',
  import_csv: 'CSV 导入',
  import_xlsx: '数据备份导入'
}

const ACTION_TYPES = {
  create: 'success',
  update: 'primary',
  delete: 'danger',
  batch_update: 'warning',
  import_csv: 'info',
  import_xlsx: 'info'
}

// 字段键 → 中文标签（与后端快照键一致）
const FIELD_LABELS = {
  region: '区域',
  instance_id: '实例ID',
  name: '主机名称',
  private_ip: '内网IP',
  asset_type: '资产类型',
  os: '操作系统',
  cpu: 'CPU核数',
  cpu_arch: 'CPU架构',
  memory: '内存(GB)',
  object_storage: '对象存储(GB)',
  system_disk: '系统盘(GB)',
  data_disk: '数据盘(GB)',
  env_type: '环境类型',
  is_db_server: '数据库服务器',
  status: '状态',
  open_ports: '开放端口',
  tags: '标签',
  person_id: '关联人员',
  apply_unit: '申请单位',
  applicant: '申请人',
  applicant_contact: '联系方式',
  project: '所属项目',
  apply_reason: '申请理由',
  apply_config: '申请配置',
  apply_time: '申请时间',
  object_storage_size: '对象存储',
  remark: '备注',
  account_name: '账户名',
  account_name_zh: '账号姓名',
  contact: '账户联系方式',
  exit_location: '接入地区',
  targets: '接入目标',
  system_name: '系统名称',
  public_ip: '公网IP',
  host_id: '内网主机',
  external_ports: '外网端口',
  external_protocols: '外网协议',
  internal_ports: '内网端口',
  internal_protocols: '内网协议',
  domain: '域名'
}

function actionLabel(action) {
  return ACTION_LABELS[action] || action || '-'
}

function actionType(action) {
  return ACTION_TYPES[action] || 'info'
}

function fieldLabel(key) {
  return FIELD_LABELS[key] || key
}

function parseDetail(item) {
  let detail = item.detail
  if (typeof detail === 'string') {
    try {
      detail = JSON.parse(detail)
    } catch {
      return null
    }
  }
  if (!detail || typeof detail !== 'object') return null
  return detail
}

// 变更行：修改只显示有变化的字段（旧值→新值）；新增/删除显示对应侧全部非空字段
function rowsOf(item) {
  const detail = parseDetail(item)
  if (!detail) return []

  const before = detail.before || null
  const after = detail.after || null

  if (item.action === 'update' || item.action === 'batch_update') {
    const beforeObj = before || {}
    const afterObj = after || {}
    const keys = new Set([...Object.keys(beforeObj), ...Object.keys(afterObj)])
    const rows = []
    for (const key of keys) {
      const b = formatVal(beforeObj[key])
      const a = formatVal(afterObj[key])
      if (b === a) continue
      rows.push({ key, label: fieldLabel(key), before: b || '空', after: a || '空', changed: true })
    }
    return rows
  }

  // create / delete：展示有值的一侧字段
  const src = item.action === 'delete' ? (before || {}) : (after || {})
  const rows = []
  for (const [key, val] of Object.entries(src)) {
    const v = formatVal(val)
    if (!v) continue
    rows.push({ key, label: fieldLabel(key), before: '', after: v, changed: false })
  }
  return rows
}

function formatVal(v) {
  if (v === null || v === undefined || v === '') return ''
  if (typeof v === 'boolean') return v ? '是' : '否'
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

async function load() {
  if (!props.resourceId) {
    items.value = []
    return
  }
  loading.value = true
  try {
    const res = await getResourceHistory(props.resourceType, props.resourceId)
    if (res.code === 0) {
      items.value = res.data || []
    }
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.resourceType, props.resourceId],
  () => load(),
  { immediate: true }
)

defineExpose({ load })
</script>

<style scoped>
.audit-timeline {
  min-height: 120px;
  max-height: 56vh;
  overflow-y: auto;
  padding: 4px 8px 0 4px;
}

.timeline-head {
  display: flex;
  align-items: center;
  gap: 8px;
}

.operator {
  color: #606266;
  font-size: 13px;
}

.diff-block {
  margin-top: 8px;
  background: #f5f7fa;
  border-radius: 4px;
  padding: 8px 10px;
}

.diff-row {
  display: flex;
  align-items: baseline;
  gap: 6px;
  font-size: 12px;
  line-height: 1.9;
}

.diff-key {
  color: #909399;
  min-width: 96px;
  flex-shrink: 0;
}

.diff-before {
  color: #f56c6c;
  text-decoration: line-through;
  word-break: break-all;
}

.diff-arrow {
  color: #c0c4cc;
}

.diff-after {
  color: #67c23a;
  word-break: break-all;
}

.diff-empty {
  margin-top: 8px;
  color: #c0c4cc;
  font-size: 12px;
}
</style>
