<template>
  <el-dialog
    v-model="visible"
    :title="dialogTitle"
    width="800px"
    :close-on-click-modal="false"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="120px" :disabled="mode === 'view'">
      <el-tabs v-model="activeTab">
        <el-tab-pane label="技术信息" name="tech">
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="区域" prop="region">
                <el-select v-model="form.region" placeholder="请选择区域" style="width: 100%">
                  <el-option v-for="item in regionOptions" :key="item" :label="item" :value="item" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="主机名称" prop="name">
                <el-input v-model="form.name" placeholder="请输入主机名称" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="实例ID" prop="instance_id">
                <el-input v-model="form.instance_id" placeholder="请输入实例ID" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="内网IP" prop="private_ip">
                <el-input v-model="form.private_ip" placeholder="请输入内网IP" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="公网IP">
                <el-input v-model="form.public_ip" placeholder="请输入公网IP" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="资产类型">
                <el-select v-model="form.asset_type" placeholder="请选择" style="width: 100%">
                  <el-option v-for="item in assetTypeOptions" :key="item" :label="item" :value="item" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="操作系统">
                <el-input v-model="form.os" placeholder="如: CentOS 7.9" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="CPU核数">
                <el-input v-model="form.cpu" placeholder="请输入CPU核数" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="CPU架构">
                <el-select v-model="form.cpu_arch" placeholder="请选择" style="width: 100%">
                  <el-option v-for="item in cpuArchOptions" :key="item.value" :label="item.label" :value="item.value" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12" />
          </el-row>
          <el-row :gutter="16">
            <el-col :span="8">
              <el-form-item label="内存(GB)">
                <el-input v-model="form.memory" placeholder="请输入内存" />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="系统盘(GB)">
                <el-input v-model="form.system_disk" placeholder="请输入系统盘" />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="数据盘(GB)">
                <el-input v-model="form.data_disk" placeholder="请输入数据盘" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="16">
            <el-col :span="8">
              <el-form-item label="环境类型">
                <el-select v-model="form.env_type" placeholder="请选择" style="width: 100%">
                  <el-option v-for="item in envTypeOptions" :key="item" :label="item" :value="item" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="状态">
                <el-select v-model="form.status" placeholder="请选择" style="width: 100%">
                  <el-option v-for="item in statusOptions" :key="item" :label="item" :value="item" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="数据库服务器">
                <el-switch v-model="form.is_db_server" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-form-item label="开放端口">
            <el-input v-model="form.open_ports" placeholder="如: 80,443,22/tcp" />
          </el-form-item>
          <el-form-item label="标签">
            <el-input v-model="form.tags" placeholder="标签，逗号分隔" />
          </el-form-item>
        </el-tab-pane>

        <el-tab-pane label="申请信息" name="apply">
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="申请单位">
                <el-input v-model="form.apply_unit" placeholder="请输入申请单位" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="申请人">
                <el-select
                  v-model="form.applicant"
                  placeholder="选择已有人员，或直接输入新人员"
                  clearable
                  filterable
                  allow-create
                  default-first-option
                  style="width: 100%"
                  @change="handleApplicantChange"
                >
                  <el-option v-for="p in personOptions" :key="p.id" :label="personLabel(p)" :value="p.name" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="联系方式">
                <el-input v-model="form.applicant_contact" placeholder="请输入联系方式" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="所属项目">
                <el-input v-model="form.project" placeholder="请输入所属项目" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-form-item label="申请理由">
            <el-input v-model="form.apply_reason" type="textarea" :rows="2" placeholder="请输入申请理由" />
          </el-form-item>
          <el-form-item label="申请配置">
            <el-input v-model="form.apply_config" type="textarea" :rows="2" placeholder="请输入申请配置" />
          </el-form-item>
          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="申请时间">
                <el-input v-model="form.apply_time" placeholder="如: 2025-01-01" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="对象存储">
                <el-input v-model="form.object_storage_size" placeholder="如: 500GB" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-form-item label="备注">
            <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="请输入备注" />
          </el-form-item>
        </el-tab-pane>
      </el-tabs>
    </el-form>

    <template #footer v-if="mode !== 'view'">
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ submitting ? '提交中...' : '确定' }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useHostStore } from '../stores/host'
import { useCloudResourceStore } from '../stores/cloudResource'
import { getPersons, createPerson } from '../api/person'
import { ElMessage } from 'element-plus'
import { statusOptions, envTypeOptions, assetTypeOptions, cpuArchOptions } from '../utils'

const hostStore = useHostStore()
const cloudStore = useCloudResourceStore()
const regionOptions = computed(() => cloudStore.allRegions)

const visible = ref(false)
const mode = ref('create')
const submitting = ref(false)
const activeTab = ref('tech')
const editId = ref(null)
const formRef = ref(null)

const form = reactive({
  region: '', name: '', instance_id: '', private_ip: '', public_ip: '',
  asset_type: '', os: '', cpu: '', cpu_arch: '', memory: '', disk: '',
  system_disk: '', data_disk: '', env_type: '', is_db_server: false,
  status: '', open_ports: '', tags: '', person_id: '',
  apply_unit: '', applicant: '', applicant_contact: '', project: '',
  apply_reason: '', apply_config: '', apply_time: '', object_storage_size: '', remark: ''
})

const personOptions = ref([])

async function loadPersons() {
  try {
    const res = await getPersons()
    if (res.code === 0) {
      personOptions.value = res.data || []
    }
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  }
}

function personLabel(p) {
  return p.unit ? `${p.name} · ${p.unit}` : p.name
}

function handleApplicantChange(val) {
  const name = (val || '').trim()
  const p = personOptions.value.find(item => item.name === name)
  if (p) {
    form.person_id = p.id
    form.applicant = p.name
    form.apply_unit = p.unit || ''
    form.applicant_contact = p.contact || ''
  } else {
    // 手动输入的申请人：提交时新增至人员库
    form.person_id = ''
  }
}

// 确定提交时关联的人员ID：已有人员直接关联，新输入的先写入人员库
async function resolvePersonId() {
  const name = (form.applicant || '').trim()
  if (!name) return ''

  if (form.person_id) {
    const linked = personOptions.value.find(p => p.id === form.person_id)
    if (linked && linked.name === name) return form.person_id
  }

  const exist = personOptions.value.find(p => p.name === name)
  if (exist) return exist.id

  try {
    const res = await createPerson({
      name,
      contact: (form.applicant_contact || '').trim(),
      unit: (form.apply_unit || '').trim()
    })
    if (res.code === 0) return res.data.id
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  }
  return false
}

const rules = {
  region: [{ required: true, message: '请选择区域', trigger: 'change' }],
  name: [{ required: true, message: '请输入主机名称', trigger: 'blur' }],
  private_ip: [{ required: true, message: '请输入内网IP', trigger: 'blur' }],
  cpu: [{ pattern: /^\d*$/, message: '请输入数字', trigger: 'blur' }],
  memory: [{ pattern: /^\d*$/, message: '请输入数字', trigger: 'blur' }],
  system_disk: [{ pattern: /^\d*$/, message: '请输入数字', trigger: 'blur' }],
  data_disk: [{ pattern: /^\d*$/, message: '请输入数字', trigger: 'blur' }]
}

const dialogTitle = computed(() => {
  if (mode.value === 'create') return '新增主机'
  if (mode.value === 'edit') return '编辑主机'
  return '主机详情'
})

function resetForm() {
  Object.keys(form).forEach(key => {
    if (typeof form[key] === 'boolean') form[key] = false
    else form[key] = ''
  })
}

async function open(data) {
  resetForm()
  mode.value = data?.mode || 'create'
  editId.value = null
  activeTab.value = 'tech'
  await loadPersons()

  if (data?.data) {
    const h = data.data
    form.region = h.region || ''
    form.name = h.name || ''
    form.instance_id = h.instance_id || ''
    form.private_ip = h.private_ip || ''
    form.public_ip = h.public_ip || ''
    form.asset_type = h.asset_type || ''
    form.os = h.os || ''
    form.cpu = h.cpu ?? ''
    form.cpu_arch = h.cpu_arch || ''
    form.memory = h.memory ?? ''
    form.disk = h.disk ?? ''
    form.system_disk = h.system_disk ?? ''
    form.data_disk = h.data_disk ?? ''
    form.env_type = h.env_type || ''
    form.is_db_server = h.is_db_server || false
    form.status = h.status || ''
    form.open_ports = h.open_ports || ''
    form.tags = h.tags || ''
    form.person_id = h.person_id ?? ''
    editId.value = h.id

    if (h.application) {
      form.apply_unit = h.application.apply_unit || ''
      form.applicant = h.application.applicant || ''
      form.applicant_contact = h.application.applicant_contact || ''
      form.project = h.application.project || ''
      form.apply_reason = h.application.apply_reason || ''
      form.apply_config = h.application.apply_config || ''
      form.apply_time = h.application.apply_time || ''
      form.object_storage_size = h.application.object_storage_size || ''
      form.remark = h.application.remark || ''
    }

    // 已关联人员但申请信息为空时，从人员库回填
    if (form.person_id) {
      const p = personOptions.value.find(item => item.id === form.person_id)
      if (p) {
        if (!form.applicant) form.applicant = p.name
        if (!form.apply_unit) form.apply_unit = p.unit
        if (!form.applicant_contact) form.applicant_contact = p.contact
      }
    }
  }

  visible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const personId = await resolvePersonId()
    if (personId === false) return

    const payload = { ...form }
    payload.person_id = personId || null
    ;['cpu', 'memory', 'disk', 'system_disk', 'data_disk'].forEach(key => {
      payload[key] = Number(payload[key]) || 0
    })
    let success
    if (mode.value === 'create') {
      success = await hostStore.create(payload)
    } else {
      success = await hostStore.update(editId.value, payload)
    }
    if (success) {
      ElMessage.success(mode.value === 'create' ? '创建成功' : '更新成功')
      visible.value = false
    }
  } finally {
    submitting.value = false
  }
}

// Listen for open events
function handleOpenHostForm(e) {
  open(e.detail)
}

import { onMounted, onUnmounted } from 'vue'
onMounted(() => window.addEventListener('open-host-form', handleOpenHostForm))
onUnmounted(() => window.removeEventListener('open-host-form', handleOpenHostForm))
</script>
