<template>
  <el-dialog
    v-model="visible"
    title="批量添加零信任申请"
    width="760px"
    :close-on-click-modal="false"
  >
    <div class="hint">
      <p>每行一条记录，字段用逗号分隔，列顺序如下（至少填写前 4 列）：</p>
      <p class="columns">
        {{ columnOrder }}
      </p>
      <p>主机按「内网IP」定位，不存在则该行失败；合法行全部插入。支持双引号包裹含逗号字段；# 开头为注释；首行可粘贴表头。</p>
    </div>
    <el-input
      v-model="textInput"
      type="textarea"
      :rows="12"
      placeholder="申请单位,账户名,联系方式,内网IP,申请端口,系统名称,申请时间,备注
某某研究院,zhangsan,13800000000,192.168.1.10,22,统一门户,2026-10-08T10:00:00,临时开通
某某研究院,lisi,13900000000,192.168.1.11,3389,运维通道"
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
import { batchCreateZeroTrustsText } from '../api/zero_trust'
import { ElMessage } from 'element-plus'

const columnOrder = [
  '申请单位', '账户名', '联系方式', '内网IP', '申请端口',
  '系统名称', '申请时间', '备注'
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
    ElMessage.warning('请输入零信任数据')
    return
  }

  submitting.value = true
  try {
    const res = await batchCreateZeroTrustsText(text)
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
onMounted(() => window.addEventListener('open-zero-trust-batch', handleOpen))
onUnmounted(() => window.removeEventListener('open-zero-trust-batch', handleOpen))

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
