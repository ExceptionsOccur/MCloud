<template>
  <el-dialog
    v-model="visible"
    title="零信任申请详情"
    width="620px"
  >
    <el-tabs v-model="activeTab">
      <el-tab-pane
        label="申请信息"
        name="info"
      >
        <el-descriptions
          :column="2"
          border
        >
          <el-descriptions-item label="申请单位">
            {{ record.apply_unit || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="账户名">
            {{ record.account_name || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="账号姓名">
            {{ record.account_name_zh || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="申请人">
            {{ record.applicant || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="账户联系方式">
            {{ record.contact || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="系统名称">
            {{ record.system_name || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="申请时间">
            {{ record.apply_time || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="接入地区">
            <template v-if="locations.length">
              <el-tag
                v-for="(loc, i) in locations"
                :key="i"
                size="small"
                type="info"
                class="loc-tag"
              >
                {{ loc }}
              </el-tag>
            </template>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item
            label="备注"
            :span="2"
          >
            {{ record.remark || '-' }}
          </el-descriptions-item>
        </el-descriptions>

        <div class="targets-block">
          <h4>接入目标</h4>
          <el-table
            v-if="targetGroups.length"
            :data="targetGroups"
            size="small"
            border
          >
            <el-table-column
              prop="label"
              label="主机"
              min-width="160"
            />
            <el-table-column
              label="端口 / 协议"
              min-width="200"
            >
              <template #default="{ row }">
                <el-tag
                  v-for="(p, pi) in row.ports"
                  :key="pi"
                  size="small"
                  class="port-tag"
                >
                  {{ p.ports }}/{{ p.protocol }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
          <el-empty
            v-else
            description="无接入目标"
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
          resource-type="zero_trust"
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
  apply_unit: '',
  account_name: '',
  account_name_zh: '',
  applicant: '',
  contact: '',
  system_name: '',
  apply_time: '',
  exit_location: '',
  remark: '',
  targets: []
})

const locations = computed(() =>
  String(record.exit_location || '')
    .split(',')
    .map(s => s.trim())
    .filter(Boolean)
)

const targetGroups = computed(() => {
  const map = new Map()
  for (const t of record.targets || []) {
    if (!map.has(t.host_id)) {
      map.set(t.host_id, {
        host_id: t.host_id,
        label: t.host_name ? (t.private_ip ? `${t.host_name}(${t.private_ip})` : t.host_name) : `#${t.host_id}`,
        ports: []
      })
    }
    map.get(t.host_id).ports.push({ ports: t.ports || '', protocol: t.protocol || 'tcp' })
  }
  return [...map.values()]
})

function open(row) {
  Object.assign(record, {
    id: row.id,
    apply_unit: row.apply_unit || '',
    account_name: row.account_name || '',
    account_name_zh: row.account_name_zh || '',
    applicant: row.applicant || '',
    contact: row.contact || '',
    system_name: row.system_name || '',
    apply_time: row.apply_time || '',
    exit_location: row.exit_location || '',
    remark: row.remark || '',
    targets: row.targets || []
  })
  activeTab.value = 'info'
  visible.value = true
}

defineExpose({ open })
</script>

<style scoped>
.loc-tag,
.port-tag {
  margin: 2px 4px 2px 0;
}

.targets-block {
  margin-top: 16px;
}

.targets-block h4 {
  margin: 0 0 8px;
  font-size: 14px;
  color: #303133;
}
</style>
