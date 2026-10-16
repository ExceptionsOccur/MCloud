<template>
  <el-dialog
    v-model="visible"
    title="端口映射详情"
    width="620px"
  >
    <el-tabs v-model="activeTab">
      <el-tab-pane
        label="映射信息"
        name="info"
      >
        <el-descriptions
          :column="2"
          border
        >
          <el-descriptions-item label="公网IP">
            {{ record.public_ip || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="内网主机">
            <span v-if="record.host">
              {{ record.host.name }}
              <span class="host-ip">({{ record.host.private_ip }})</span>
            </span>
            <span v-else>#{{ record.host_id }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="运营商">
            {{ record.isp || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="出口位置">
            {{ record.exit_location || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="域名">
            {{ record.domain || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="备注">
            {{ record.remark || '-' }}
          </el-descriptions-item>
        </el-descriptions>

        <div class="ports-block">
          <h4>端口映射</h4>
          <el-table
            v-if="portPairs.length"
            :data="portPairs"
            size="small"
            border
          >
            <el-table-column
              prop="external"
              label="外网端口"
              min-width="120"
            />
            <el-table-column
              prop="internal"
              label="内网端口"
              min-width="120"
            />
          </el-table>
          <el-empty
            v-else
            description="无端口映射"
            :image-size="48"
          />
        </div>
      </el-tab-pane>
      <el-tab-pane
        label="变更记录"
        name="history"
      >
        <AuditTimeline
          v-if="visible"
          :key="record.id"
          resource-type="port_mapping"
          :resource-id="record.id"
        />
      </el-tab-pane>
    </el-tabs>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import AuditTimeline from './AuditTimeline.vue'

const visible = ref(false)
const activeTab = ref('info')
const record = reactive({
  id: null,
  public_ip: '',
  host_id: null,
  host: null,
  external_ports: '',
  external_protocols: '',
  internal_ports: '',
  internal_protocols: '',
  isp: '',
  exit_location: '',
  domain: '',
  remark: ''
})

function splitPorts(val) {
  if (!val) return []
  return String(val).split(',').map(s => s.trim()).filter(Boolean)
}

const portPairs = computed(() => {
  const ext = splitPorts(record.external_ports)
  const extPro = splitPorts(record.external_protocols)
  const intl = splitPorts(record.internal_ports)
  const intlPro = splitPorts(record.internal_protocols)
  const n = Math.max(ext.length, intl.length)
  const pairs = []
  for (let i = 0; i < n; i++) {
    pairs.push({
      external: ext[i] ? `${ext[i]}/${extPro[i] || 'tcp'}` : '-',
      internal: intl[i] ? `${intl[i]}/${intlPro[i] || 'tcp'}` : '-'
    })
  }
  return pairs
})

function open(row) {
  Object.assign(record, {
    id: row.id,
    public_ip: row.public_ip || '',
    host_id: row.host_id,
    host: row.host || null,
    external_ports: row.external_ports || '',
    external_protocols: row.external_protocols || '',
    internal_ports: row.internal_ports || '',
    internal_protocols: row.internal_protocols || '',
    isp: row.isp || '',
    exit_location: row.exit_location || '',
    domain: row.domain || '',
    remark: row.remark || ''
  })
  activeTab.value = 'info'
  visible.value = true
}

defineExpose({ open })
</script>

<style scoped>
.host-ip {
  color: #909399;
  font-size: 12px;
}

.ports-block {
  margin-top: 16px;
}

.ports-block h4 {
  margin: 0 0 8px;
  font-size: 14px;
  color: #303133;
}
</style>
