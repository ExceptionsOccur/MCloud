<template>
  <el-dialog
    v-model="visible"
    title="IP网段管理"
    width="540px"
  >
    <div class="add-row">
      <el-input
        v-model="newCidr"
        placeholder="输入网段，如 172.17.100.0/24"
        @keyup.enter="handleCreate"
      />
      <el-button
        type="primary"
        :loading="creating"
        @click="handleCreate"
      >
        新增
      </el-button>
    </div>
    <el-table
      v-loading="loading"
      :data="subnets"
      size="small"
      max-height="360px"
      style="margin-top: 12px"
    >
      <el-table-column
        type="index"
        label="#"
        width="50"
        align="center"
      />
      <el-table-column
        prop="cidr"
        label="网段"
        align="left"
      />
      <el-table-column
        label="操作"
        width="140"
        align="center"
      >
        <template #default="{ row }">
          <el-button
            link
            type="primary"
            size="small"
            @click="handleEdit(row)"
          >
            编辑
          </el-button>
          <el-button
            link
            type="danger"
            size="small"
            @click="handleDelete(row)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="hint">
      仅支持 /24 掩码的 IPv4 网段，系统会自动规范化为网络地址（末位归 0）
    </div>
  </el-dialog>
</template>

<script setup>
import { ref } from 'vue'
import { getSubnets, createSubnet, updateSubnet, deleteSubnet } from '../api/subnet'
import { ElMessage, ElMessageBox } from 'element-plus'

const emit = defineEmits(['changed'])

const visible = ref(false)
const subnets = ref([])
const newCidr = ref('')
const creating = ref(false)
const loading = ref(false)

async function loadSubnets() {
  loading.value = true
  try {
    const res = await getSubnets()
    if (res.code === 0) {
      subnets.value = res.data || []
    }
  } finally {
    loading.value = false
  }
}

function open() {
  visible.value = true
  newCidr.value = ''
  loadSubnets()
}

async function handleCreate() {
  const cidr = newCidr.value.trim()
  if (!cidr) {
    ElMessage.warning('请输入网段')
    return
  }
  creating.value = true
  try {
    const res = await createSubnet(cidr)
    if (res.code === 0) {
      ElMessage.success('新增成功')
      newCidr.value = ''
      await loadSubnets()
      emit('changed')
    }
  } finally {
    creating.value = false
  }
}

async function handleEdit(row) {
  let value
  try {
    const ret = await ElMessageBox.prompt('修改网段', '编辑', {
      inputValue: row.cidr,
      inputPlaceholder: '如 172.17.100.0/24',
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputValidator: v => (v && v.trim() !== '' ? true : '网段不能为空')
    })
    value = ret.value.trim()
  } catch {
    return
  }

  const res = await updateSubnet(row.id, value)
  if (res.code === 0) {
    ElMessage.success('修改成功')
    await loadSubnets()
    emit('changed')
  }
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确定删除网段 ${row.cidr}？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  const res = await deleteSubnet(row.id)
  if (res.code === 0) {
    ElMessage.success('删除成功')
    await loadSubnets()
    emit('changed')
  }
}

defineExpose({ open })
</script>

<style scoped>
.add-row {
  display: flex;
  gap: 8px;
}

.hint {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}
</style>
