<template>
  <div class="host-table">
    <el-table
      v-loading="hostStore.loading"
      :data="hostStore.hosts"
      border
      stripe
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="40" />
      <el-table-column label="#" width="50">
        <template #default="{ $index }">
          {{ (hostStore.currentPage - 1) * hostStore.pageSize + $index + 1 }}
        </template>
      </el-table-column>
      <el-table-column prop="region" label="区域" width="80" />
      <el-table-column label="主机名称" min-width="140">
        <template #default="{ row }">
          <div class="name-cell">
            <span class="name-text">{{ row.name }}</span>
            <span v-if="row.env_type === '生产'" class="env-tag-prod">生产</span>
            <span v-else-if="row.env_type === '测试'" class="env-tag-test">测试</span>
            <span v-if="row.is_db_server" class="db-tag">DB</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="private_ip" label="内网IP" width="140" />
      <el-table-column prop="public_ip" label="公网IP" width="140" show-overflow-tooltip />
      <el-table-column label="资产类型" min-width="70">
        <template #default="{ row }">
          <div class="name-cell">
            <span class="name-text">{{ row.asset_type }}</span>
            <span v-if="row.cpu_arch" :class="'arch-tag arch-' + row.cpu_arch">{{ row.cpu_arch }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="规格" min-width="90">
        <template #default="{ row }">
          {{ row.cpu ? row.cpu + 'vCPU' : '-' }}/{{ row.memory ? row.memory + 'G' : '-' }}/{{ row.system_disk ? row.system_disk + 'G' : '-' }}/{{ row.data_disk ? row.data_disk + 'G' : '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }">
          <span :class="'status-' + row.status">{{ row.status }}</span>
        </template>
      </el-table-column>
      <el-table-column label="项目" min-width="120" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.application?.project || '' }}
        </template>
      </el-table-column>
      <el-table-column label="申请人" min-width="40" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.application?.applicant || '' }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button type="primary" link size="small" @click="handleView(row)">详情</el-button>
          <el-popconfirm title="确定删除？" @confirm="handleDelete(row.id)">
            <template #reference>
              <el-button type="danger" link size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-pagination
        v-model:current-page="hostStore.currentPage"
        :page-size="hostStore.pageSize"
        :total="hostStore.total"
        layout="total, prev, pager, next, sizes"
        :page-sizes="[20, 50, 100]"
        @current-change="hostStore.setPage"
        @size-change="handleSizeChange"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useHostStore } from '../stores/host'
import { ElMessage } from 'element-plus'

const hostStore = useHostStore()
const emit = defineEmits(['selection-change'])

const selectedIds = ref([])

function handleSelectionChange(selection) {
  selectedIds.value = selection.map(row => row.id)
  emit('selection-change', selectedIds.value)
  // Update parent's selectedIds
  window.dispatchEvent(new CustomEvent('selection-change', { detail: { ids: selectedIds.value } }))
}

function handleEdit(row) {
  window.dispatchEvent(new CustomEvent('open-host-form', { detail: { mode: 'edit', data: row } }))
}

function handleView(row) {
  window.dispatchEvent(new CustomEvent('open-host-form', { detail: { mode: 'view', data: row } }))
}

async function handleDelete(id) {
  const success = await hostStore.remove(id)
  if (success) {
    ElMessage.success('删除成功')
  }
}

function handleSizeChange(size) {
  hostStore.pageSize = size
  hostStore.currentPage = 1
  hostStore.fetchHosts()
}

// Listen for selection changes from SearchToolbar
function handleToolbarSelection(e) {
  selectedIds.value = e.detail.ids
}

onMounted(() => {
  window.addEventListener('selection-change', handleToolbarSelection)
})

onUnmounted(() => {
  window.removeEventListener('selection-change', handleToolbarSelection)
})
</script>

<style scoped>
.host-table {
  background: #fff;
  padding: 16px;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.name-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.name-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:deep(.name-cell .db-tag) {
  margin-left: 0;
}

.arch-tag {
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 12px;
  flex-shrink: 0;
}

.arch-C86 {
  color: #409eff;
  border: 1px solid #d9ecff;
  background: #ecf5ff;
}

.arch-ARM {
  color: #67c23a;
  border: 1px solid #e1f3d8;
  background: #f0f9eb;
}

.arch-X86 {
  color: #e6a23c;
  border: 1px solid #faecd8;
  background: #fdf6ec;
}
</style>
