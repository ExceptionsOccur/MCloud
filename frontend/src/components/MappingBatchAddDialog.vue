<template>
  <el-dialog
    v-model="visible"
    title="批量添加端口映射"
    width="760px"
    :close-on-click-modal="false"
  >
    <div class="hint">
      <p>每行一条记录，字段用逗号分隔，列顺序如下（至少填写前 4 列）：</p>
      <p class="columns">
        {{ columnOrder }}
      </p>
      <p>外网端口/内网端口各自用逗号分隔且数量一致；内网IP定位主机；域名可选。# 开头为注释；首行可粘贴表头。</p>
    </div>
    <el-input
      v-model="textInput"
      type="textarea"
      :rows="12"
      placeholder="公网IP,内网IP,外网端口,内网端口,域名,运营商,出口位置,备注
203.0.113.10,192.168.1.10,80,8080,www.example.com,电信,上海,业务
203.0.113.10,192.168.1.10,443,8443,,,,"
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
import { batchCreatePortMappingsText } from '../api/port_mapping'
import { ElMessage } from 'element-plus'

const columnOrder = [
  '公网IP', '内网IP', '外网端口', '内网端口',
  '域名', '运营商', '出口位置', '备注'
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
    ElMessage.warning('请输入映射数据')
    return
  }

  submitting.value = true
  try {
    const res = await batchCreatePortMappingsText(text)
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
onMounted(() => window.addEventListener('open-mapping-batch', handleOpen))
onUnmounted(() => window.removeEventListener('open-mapping-batch', handleOpen))

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
