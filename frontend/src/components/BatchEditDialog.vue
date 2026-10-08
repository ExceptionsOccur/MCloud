<template>
  <el-dialog
    v-model="visible"
    title="批量编辑主机"
    width="700px"
    :close-on-click-modal="false"
  >
    <p style="margin-bottom: 12px; color: #909399;">
      已选择 {{ selectedIds.length }} 台主机，填写需要统一修改的字段（留空的字段不会修改）。
    </p>
    <el-form label-width="120px">
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="区域">
            <el-select
              v-model="data.region"
              clearable
              placeholder="不修改"
              style="width: 100%"
            >
              <el-option
                v-for="item in regionOptions"
                :key="item"
                :label="item"
                :value="item"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="环境类型">
            <el-select
              v-model="data.env_type"
              clearable
              placeholder="不修改"
              style="width: 100%"
            >
              <el-option
                v-for="item in envTypeOptions"
                :key="item"
                :label="item"
                :value="item"
              />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="资产类型">
            <el-select
              v-model="data.asset_type"
              clearable
              placeholder="不修改"
              style="width: 100%"
            >
              <el-option
                v-for="item in assetTypeOptions"
                :key="item"
                :label="item"
                :value="item"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="状态">
            <el-select
              v-model="data.status"
              clearable
              placeholder="不修改"
              style="width: 100%"
            >
              <el-option
                v-for="item in statusOptions"
                :key="item"
                :label="item"
                :value="item"
              />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="申请人">
            <el-select
              v-model="data.applicant"
              placeholder="选择已有人员，或直接输入新人员"
              clearable
              filterable
              allow-create
              default-first-option
              style="width: 100%"
              :filter-method="handleApplicantFilter"
              @change="handleApplicantChange"
              @visible-change="handleApplicantVisibleChange"
            >
              <el-option
                v-for="p in filteredPersons"
                :key="p.id"
                :label="personLabel(p)"
                :value="p.name"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="所属项目">
            <el-input
              v-model="data.project"
              placeholder="不修改请留空"
            />
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="申请单位">
            <el-input
              v-model="data.apply_unit"
              placeholder="不修改请留空"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="联系方式">
            <el-input
              v-model="data.applicant_contact"
              placeholder="不修改请留空"
            />
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="标签">
        <el-input
          v-model="data.tags"
          placeholder="不修改请留空"
        />
      </el-form-item>
      <el-form-item label="备注">
        <el-input
          v-model="data.remark"
          type="textarea"
          :rows="2"
          placeholder="不修改请留空"
        />
      </el-form-item>
    </el-form>
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
import { ref, reactive, computed } from 'vue'
import { batchUpdateHosts } from '../api/host'
import { getPersons, createPerson } from '../api/person'
import { useHostStore } from '../stores/host'
import { useCloudResourceStore } from '../stores/cloudResource'
import { ElMessage } from 'element-plus'
import { statusOptions, envTypeOptions, assetTypeOptions } from '../utils'

const hostStore = useHostStore()
const cloudStore = useCloudResourceStore()
const regionOptions = computed(() => cloudStore.allRegions)
const visible = ref(false)
const submitting = ref(false)
const selectedIds = ref([])

const data = reactive({
  region: '', env_type: '', asset_type: '', status: '',
  applicant: '', project: '', apply_unit: '', applicant_contact: '',
  tags: '', remark: '', person_id: ''
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

const applicantQuery = ref('')

function handleApplicantFilter(q) {
  applicantQuery.value = (q || '').trim()
}

const filteredPersons = computed(() => {
  const q = applicantQuery.value.toLowerCase()
  if (!q) return personOptions.value
  return personOptions.value.filter(p =>
    personLabel(p).toLowerCase().includes(q) || p.name.toLowerCase().includes(q)
  )
})

function applyPerson(p) {
  data.person_id = p.id
  data.applicant = p.name
  data.apply_unit = p.unit || ''
  data.applicant_contact = p.contact || ''
}

function handleApplicantChange(val) {
  const name = (val || '').trim()
  applicantQuery.value = ''
  const p = personOptions.value.find(item => item.name === name)
  if (p) applyPerson(p)
  else data.person_id = ''
}

function handleApplicantVisibleChange(open) {
  if (open) return
  commitPendingApplicant()
}

// 下拉关闭时把未回车提交的手动输入落到表单，保证提交时会新增人员
function commitPendingApplicant() {
  const q = applicantQuery.value
  applicantQuery.value = ''
  if (!q || q === data.applicant) return

  const exact = personOptions.value.find(p => p.name === q || personLabel(p) === q)
  if (exact) {
    if (exact.id !== data.person_id) applyPerson(exact)
    return
  }

  // 只输了筛选关键字（仍有前缀匹配项）则视为未敲定，保持原值
  const lower = q.toLowerCase()
  const similar = personOptions.value.some(p =>
    p.name.toLowerCase().startsWith(lower) || personLabel(p).toLowerCase().startsWith(lower)
  )
  if (similar) return

  data.applicant = q
  data.person_id = ''
}

// 返回 undefined=不修改关联；false=新增人员失败，需中止提交
async function resolvePersonId() {
  commitPendingApplicant()
  const name = (data.applicant || '').trim()
  if (!name) return undefined

  if (data.person_id) {
    const linked = personOptions.value.find(p => p.id === data.person_id)
    if (linked && linked.name === name) return data.person_id
  }

  const exist = personOptions.value.find(p => p.name === name)
  if (exist) return exist.id

  try {
    const res = await createPerson({
      name,
      contact: (data.applicant_contact || '').trim(),
      unit: (data.apply_unit || '').trim()
    })
    if (res.code === 0) return res.data.id
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  }
  return false
}

async function open(ids) {
  selectedIds.value = ids || []
  Object.keys(data).forEach(k => { data[k] = '' })
  applicantQuery.value = ''
  await loadPersons()
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
    const personId = await resolvePersonId()
    if (personId === false) return
    if (personId !== undefined) {
      payload.person_id = personId
    }

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
