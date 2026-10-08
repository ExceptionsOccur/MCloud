<template>
  <el-dialog
    v-model="visible"
    title="批量添加主机"
    width="760px"
    :close-on-click-modal="false"
  >
    <div class="hint">
      <p>每行一条记录，字段用逗号分隔，列顺序如下（至少填写前 3 列，其余列可省略）：</p>
      <p class="columns">
        {{ columnOrder }}
      </p>
      <p>支持用双引号包裹含逗号的字段；以 # 开头的行视为注释；首行可直接粘贴表头。</p>
    </div>
    <el-input
      v-model="textInput"
      type="textarea"
      :rows="12"
      placeholder="区域,实例ID,主机名称,内网IP,是否映射公网,资产类型,操作系统,CPU核数,CPU架构,内存(GB),系统盘(GB),数据盘(GB),环境类型,是否数据库服务器,状态,开放端口,标签,申请单位,申请人,申请人联系方式,所属项目,申请理由,申请配置,申请时间,对象存储大小,备注
region-a,ins-001,web-01,192.168.1.10,否,虚拟机,CentOS 7.9,4,X86,8,50,100,生产,否,运行中
region-a,ins-002,db-01,192.168.1.11,是,虚拟机,CentOS 7.9,8,X86,16,100,200,生产,是,运行中"
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
import { batchCreateHostsText } from '../api/host'
import { ElMessage } from 'element-plus'

const columnOrder = [
  '区域', '实例ID', '主机名称', '内网IP', '是否映射公网',
  '资产类型', '操作系统', 'CPU核数', 'CPU架构', '内存(GB)',
  '系统盘(GB)', '数据盘(GB)', '环境类型', '是否数据库服务器',
  '状态', '开放端口', '标签', '申请单位', '申请人', '申请人联系方式',
  '所属项目', '申请理由', '申请配置', '申请时间', '对象存储大小', '备注'
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
    ElMessage.warning('请输入主机数据')
    return
  }

  submitting.value = true
  try {
    const res = await batchCreateHostsText(text)
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
