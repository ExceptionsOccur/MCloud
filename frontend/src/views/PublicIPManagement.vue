<template>
  <div class="public-ip-management">
    <AppHeader />
    <div class="main-content">
      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="搜索 IP / 运营商 / 出口位置 / 备注"
          clearable
          style="width: 260px"
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
          新增公网IP
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
          prop="ip"
          label="公网IP"
          min-width="160"
          show-overflow-tooltip
        />
        <el-table-column
          prop="isp"
          label="运营商"
          min-width="120"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.isp || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="exit_location"
          label="出口位置"
          min-width="120"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.exit_location || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="remark"
          label="备注"
          min-width="180"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.remark || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          label="创建时间"
          width="170"
        >
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
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
              :title="`确定删除公网IP ${row.ip}？`"
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
        label-width="90px"
      >
        <el-form-item
          label="公网IP"
          prop="ip"
        >
          <el-input
            v-model="form.ip"
            placeholder="请输入公网IP"
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
          label="出口位置"
          prop="exit_location"
        >
          <el-input
            v-model="form.exit_location"
            placeholder="请输入出口位置（IP所在地，选填）"
            maxlength="128"
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
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getPublicIPs, createPublicIP, updatePublicIP, deletePublicIP } from '../api/public_ip'
import { formatTime } from '../utils'
import AppHeader from '../components/AppHeader.vue'

const records = ref([])
const loading = ref(false)
const keyword = ref('')
const dialogVisible = ref(false)
const submitting = ref(false)
const editId = ref(null)
const formRef = ref(null)

const form = reactive({ ip: '', isp: '', exit_location: '', remark: '' })

const rules = {
  ip: [{ required: true, message: '请输入公网IP', trigger: 'blur' }]
}

const dialogTitle = computed(() => (editId.value ? '编辑公网IP' : '新增公网IP'))

onMounted(() => {
  loadList()
})

async function loadList() {
  loading.value = true
  try {
    const res = await getPublicIPs(keyword.value)
    if (res.code === 0) {
      records.value = res.data || []
    }
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  } finally {
    loading.value = false
  }
}

function openDialog(row) {
  editId.value = row?.id || null
  form.ip = row?.ip || ''
  form.isp = row?.isp || ''
  form.exit_location = row?.exit_location || ''
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
      ip: form.ip.trim(),
      isp: form.isp.trim(),
      exit_location: form.exit_location.trim(),
      remark: form.remark.trim()
    }
    const res = editId.value
      ? await updatePublicIP(editId.value, payload)
      : await createPublicIP(payload)
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
    const res = await deletePublicIP(id)
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
.public-ip-management {
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
</style>
