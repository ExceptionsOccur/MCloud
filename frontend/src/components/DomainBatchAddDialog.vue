<template>
  <el-dialog
    v-model="visible"
    title="批量添加域名"
    width="760px"
    :close-on-click-modal="false"
  >
    <div class="hint">
      <p>每行一条记录，字段用逗号分隔，列顺序如下（至少填写第 1 列域名）：</p>
      <p class="columns">
        {{ columnOrder }}
      </p>
      <p>域名已存在则跳过；内网IP为空表示不关联主机，填了则按内网IP定位主机。支持双引号包裹含逗号字段；# 开头为注释；首行可粘贴表头。</p>
    </div>
    <el-input
      v-model="textInput"
      type="textarea"
      :rows="12"
      placeholder="域名,解析公网IP,运营商,出口位置,内网IP,主机端口,备注
www.example.com,203.0.113.10,电信,上海,,,业务主站
api.example.com,203.0.113.11,联通,北京,192.168.1.10,443,接口"
    />
    <template #footer>
      <el-button @click="visible = false">
        取消
      </el-button>
      <el-button
        type="primary"
        :loading="submitting"
        @click="handleSubmit"
      >
        提交
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref } from 'vue'
import { batchCreateDomainsText } from '../api/domain'
import { ElMessage } from 'element-plus'

const columnOrder = [
  '域名', '解析公网IP', '运营商', '出口位置', '内网IP', '主机端口', '备注'
].join(' | ')

const visible = ref(false)
const submitting = ref(false)
const textInput = ref('')

function open() {
  textInput.value = ''
  visible.value = true
}

async function handleSubmit() {
  const text = textInput.value.trim()
  if (!text) {
    ElMessage.warning('请输入域名数据')
    return
  }

  submitting.value = true
  try {
    const res = await batchCreateDomainsText(text)
    if (res.code === 0) {
      const d = res.data
      const summary = `成功 ${d.success} 条，跳过 ${d.skipped} 条，失败 ${d.errors} 条`
      if (d.errors > 0) {
        const detail = (d.line_errors || []).join('；')
        ElMessage.warning(detail ? `${summary}：${detail}` : summary)
      } else {
        ElMessage.success(summary)
      }
      visible.value = false
      window.dispatchEvent(new CustomEvent('ledger-batch-done'))
    }
  } finally {
    submitting.value = false
  }
}

function handleOpen() {
  open()
}

import { onMounted, onUnmounted } from 'vue'
onMounted(() => window.addEventListener('open-domain-batch', handleOpen))
onUnmounted(() => window.removeEventListener('open-domain-batch', handleOpen))

defineExpose({ open })
</script>

<style scoped>
.hint {
  margin-bottom: 12px;
  font-size: 13px;
  color: #909399;
  line-height: 1.6;
}

.hint p {
  margin: 0 0 4px;
}

.columns {
  color: #606266;
  word-break: break-all;
}
</style>
