<template>
  <el-dialog v-model="visible" title="云资源录入" width="560px" :close-on-click-modal="false">
    <el-tabs v-model="activeRegion">
      <el-tab-pane v-for="region in regions" :key="region" :label="region" :name="region">
        <el-form :model="forms[region]" label-width="120px" size="default">
          <el-form-item v-for="f in fields" :key="f.key" :label="f.label">
            <el-input
              v-model="forms[region][f.key]"
              placeholder="请输入"
              @input="v => onInput(region, f.key, v)"
            />
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSave">保存{{ activeRegion }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { getCloudResources, updateCloudResource } from '../api/cloud_resource'
import { ElMessage } from 'element-plus'

const emit = defineEmits(['saved'])

const regions = ['region-a', 'region-b']
const fields = [
  { key: 'physical_cpu', label: '物理CPU(核)' },
  { key: 'vcpu', label: 'vCPU(核)' },
  { key: 'memory', label: '内存(G)' },
  { key: 'storage', label: '存储(G)' },
  { key: 'bare_metal', label: '裸金属(台)' },
  { key: 'gpu_card_count', label: 'GPU卡数' },
  { key: 'object_storage', label: '对象存储(G)' }
]

const visible = ref(false)
const saving = ref(false)
const activeRegion = ref(regions[0])
const forms = reactive({})

function emptyForm() {
  const f = {}
  fields.forEach(item => { f[item.key] = '0' })
  return f
}

function resetForms() {
  regions.forEach(r => { forms[r] = emptyForm() })
}

function onInput(region, key, value) {
  forms[region][key] = String(value).replace(/\D/g, '')
}

async function open() {
  resetForms()
  visible.value = true
  try {
    const res = await getCloudResources()
    if (res.code === 0) {
      ;(res.data || []).forEach(item => {
        if (forms[item.region]) {
          fields.forEach(f => {
            forms[item.region][f.key] = String(item[f.key] ?? 0)
          })
        }
      })
    }
  } catch (e) {
    // ignore
  }
}

async function handleSave() {
  const payload = { region: activeRegion.value }
  fields.forEach(f => {
    payload[f.key] = Number(forms[activeRegion.value][f.key]) || 0
  })

  saving.value = true
  try {
    const res = await updateCloudResource(payload)
    if (res.code === 0) {
      ElMessage.success(`${activeRegion.value} 云端总资源已保存`)
      emit('saved')
    }
  } finally {
    saving.value = false
  }
}

defineExpose({ open })
</script>
