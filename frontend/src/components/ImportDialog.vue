<template>
  <el-dialog
    v-model="visible"
    title="导入CSV"
    width="500px"
    :close-on-click-modal="false"
  >
    <el-upload
      ref="uploadRef"
      drag
      :auto-upload="false"
      :limit="1"
      accept=".csv"
      :on-change="handleFileChange"
    >
      <el-icon
        class="el-icon--upload"
        :size="40"
      >
        <Upload />
      </el-icon>
      <div class="el-upload__text">
        将文件拖到此处，或<em>点击上传</em>
      </div>
      <template #tip>
        <div class="el-upload__tip">
          仅支持 .csv 文件，最大 16MB
        </div>
      </template>
    </el-upload>
    <div
      v-if="uploadResult"
      style="margin-top: 16px;"
    >
      <el-alert
        :title="resultMessage"
        :type="resultType"
        show-icon
      />
    </div>
    <template #footer>
      <el-button @click="visible = false">
        关闭
      </el-button>
      <el-button
        type="primary"
        :loading="submitting"
        :disabled="!file"
        @click="handleSubmit"
      >
        开始导入
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref } from 'vue'
import { importCSV } from '../api/csv'
import { useHostStore } from '../stores/host'
import { ElMessage } from 'element-plus'

const hostStore = useHostStore()
const visible = ref(false)
const submitting = ref(false)
const file = ref(null)
const uploadResult = ref(null)

const resultMessage = ref('')
const resultType = ref('success')

function open() {
  file.value = null
  uploadResult.value = null
  visible.value = true
}

function handleFileChange(f) {
  file.value = f.raw
}

async function handleSubmit() {
  if (!file.value) {
    ElMessage.warning('请选择文件')
    return
  }

  submitting.value = true
  try {
    const res = await importCSV(file.value)
    if (res.code === 0) {
      uploadResult.value = true
      resultMessage.value = `导入完成：成功 ${res.data.success} 条，跳过 ${res.data.skipped} 条，失败 ${res.data.errors} 条`
      resultType.value = res.data.errors > 0 ? 'warning' : 'success'
      hostStore.fetchHosts()
    }
  } catch {
    uploadResult.value = true
    resultMessage.value = '导入失败'
    resultType.value = 'error'
  } finally {
    submitting.value = false
  }
}

function handleOpenImport() {
  open()
}

import { onMounted, onUnmounted } from 'vue'
onMounted(() => window.addEventListener('open-import', handleOpenImport))
onUnmounted(() => window.removeEventListener('open-import', handleOpenImport))
</script>
