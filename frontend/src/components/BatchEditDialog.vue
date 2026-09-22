<template>
  <el-dialog v-model="visible" title="批量编辑主机" width="700px" :close-on-click-modal="false">
    <p style="margin-bottom: 12px; color: #909399;">已选择 {{ selectedIds.length }} 台主机，填写需要统一修改的字段（留空的字段不会修改）。</p>
    <el-form label-width="120px">
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="区域">
            <el-select v-model="data.region" clearable placeholder="不修改" style="width: 100%">
              <el-option v-for="item in regionOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="环境类型">
            <el-select v-model="data.env_type" clearable placeholder="不修改" style="width: 100%">
              <el-option v-for="item in envTypeOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="资产类型">
            <el-select v-model="data.asset_type" clearable placeholder="不修改" style="width: 100%">
              <el-option v-for="item in assetTypeOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="状态">
            <el-select v-model="data.status" clearable placeholder="不修改" style="width: 100%">
              <el-option v-for="item in statusOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="申请人">
            <el-input v-model="data.applicant" placeholder="不修改请留空" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="所属项目">
            <el-input v-model="data.project" placeholder="不修改请留空" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="申请单位">
            <el-input v-model="data.apply_unit" placeholder="不修改请留空" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="标签">
            <el-input v-model="data.tags" placeholder="不修改请留空" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="备注">
        <el-input v-model="data.remark" type="textarea" :rows="2" placeholder="不修改请留空" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">提交</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { batchUpdateHosts } from '../api/host'
import { useHostStore } from '../stores/host'
import { ElMessage } from 'element-plus'
import { statusOptions, envTypeOptions, assetTypeOptions, regionOptions } from '../utils'

const hostStore = useHostStore()
const visible = ref(false)
const submitting = ref(false)
const selectedIds = ref([])

const data = reactive({
  region: '', env_type: '', asset_type: '', status: '',
  applicant: '', project: '', apply_unit: '', tags: '', remark: ''
})

function open(ids) {
  selectedIds.value = ids || []
  Object.keys(data).forEach(k => data[k] = '')
  visible.value = true
}

async function handleSubmit() {
  if (!selectedIds.value.length) {
    ElMessage.warning('未选择主机')
    return
  }

  const payload = {}
  Object.keys(data).forEach(k => {
    if (data[k] !== '' && data[k] !== null && data[k] !== undefined) {
      payload[k] = data[k]
    }
  })

  if (Object.keys(payload).length === 0) {
    ElMessage.warning('请至少填写一个字段')
    return
  }

  submitting.value = true
  try {
    const res = await batchUpdateHosts(selectedIds.value, payload)
    if (res.code === 0) {
      ElMessage.success('批量更新成功')
      hostStore.fetchHosts()
      visible.value = false
    }
  } finally {
    submitting.value = false
  }
}

function handleOpenBatchEdit(e) {
  open(e.detail.ids)
}

import { onMounted, onUnmounted } from 'vue'
onMounted(() => window.addEventListener('open-batch-edit', handleOpenBatchEdit))
onUnmounted(() => window.removeEventListener('open-batch-edit', handleOpenBatchEdit))
</script>
