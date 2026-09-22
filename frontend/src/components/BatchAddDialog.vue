<template>
  <el-dialog v-model="visible" title="批量添加主机" width="700px" :close-on-click-modal="false">
    <p style="margin-bottom: 12px; color: #909399;">请输入 JSON 格式的主机数据，每条记录需要 region, name, private_ip 字段。</p>
    <el-input v-model="jsonInput" type="textarea" :rows="12" placeholder='[
  {
    "region": "region-a",
    "name": "web-01",
    "private_ip": "192.168.1.10",
    "asset_type": "虚拟机",
    "env_type": "生产",
    "status": "运行中"
  }
]' />
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">提交</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref } from 'vue'
import { batchCreateHosts } from '../api/host'
import { ElMessage } from 'element-plus'

const visible = ref(false)
const submitting = ref(false)
const jsonInput = ref('')

function open() {
  jsonInput.value = ''
  visible.value = true
}

async function handleSubmit() {
  let hosts
  try {
    hosts = JSON.parse(jsonInput.value)
    if (!Array.isArray(hosts)) {
      ElMessage.error('请输入 JSON 数组格式')
      return
    }
  } catch {
    ElMessage.error('JSON 格式错误')
    return
  }

  submitting.value = true
  try {
    const res = await batchCreateHosts(hosts)
    if (res.code === 0) {
      ElMessage.success(`成功 ${res.data.success} 条，跳过 ${res.data.skipped} 条，失败 ${res.data.errors} 条`)
      visible.value = false
    }
  } finally {
    submitting.value = false
  }
}

function handleOpenBatchAdd() {
  open()
}

import { onMounted, onUnmounted } from 'vue'
onMounted(() => window.addEventListener('open-batch-add', handleOpenBatchAdd))
onUnmounted(() => window.removeEventListener('open-batch-add', handleOpenBatchAdd))
</script>
