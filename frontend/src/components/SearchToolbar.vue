<template>
  <div class="search-toolbar">
    <div class="toolbar-row">
      <div class="search-box">
        <el-input
          v-model="keyword"
          placeholder="搜索主机名称、IP、区域等..."
          prefix-icon="Search"
          clearable
          style="width: 320px"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
        <el-button type="primary" @click="handleSearch">搜索</el-button>
      </div>
      <div class="action-buttons">
        <el-button type="primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon> 新增主机
        </el-button>
        <el-button @click="openBatchAdd">
          <el-icon><Upload /></el-icon> 批量添加
        </el-button>
        <el-button :disabled="!selectedIds.length" @click="openBatchEdit">
          <el-icon><Edit /></el-icon> 批量编辑
        </el-button>
        <el-button @click="handleExport">
          <el-icon><Download /></el-icon> 导出CSV
        </el-button>
        <el-button @click="openImport">
          <el-icon><Upload /></el-icon> 导入CSV
        </el-button>
        <el-button @click="handleDownloadTemplate">
          <el-icon><Document /></el-icon> 下载模板
        </el-button>
      </div>
    </div>
    <div class="filter-row">
      <el-select v-model="filters.region" placeholder="区域" clearable @change="v => hostStore.setFilter('region', v)" style="width: 120px">
        <el-option v-for="item in regionOptions" :key="item" :label="item" :value="item" />
      </el-select>
      <el-select v-model="filters.env_type" placeholder="环境类型" clearable @change="v => hostStore.setFilter('env_type', v)" style="width: 120px">
        <el-option v-for="item in envTypeOptions" :key="item" :label="item" :value="item" />
      </el-select>
      <el-select v-model="filters.asset_type" placeholder="资产类型" clearable @change="v => hostStore.setFilter('asset_type', v)" style="width: 150px">
        <el-option v-for="item in assetTypeOptions" :key="item" :label="item" :value="item" />
      </el-select>
      <el-select v-model="filters.cpu_arch" placeholder="CPU架构" clearable @change="v => hostStore.setFilter('cpu_arch', v)" style="width: 130px">
        <el-option label="C86" value="C86" />
        <el-option label="ARM" value="ARM" />
      </el-select>
      <el-select v-model="filters.is_db_server" placeholder="数据库服务器" clearable @change="v => hostStore.setFilter('is_db_server', v)" style="width: 140px">
        <el-option label="是" value="1" />
        <el-option label="否" value="0" />
      </el-select>
      <el-select v-model="filters.status" placeholder="状态" clearable @change="v => hostStore.setFilter('status', v)" style="width: 120px">
        <el-option v-for="item in statusOptions" :key="item" :label="item" :value="item" />
      </el-select>
      <el-select v-model="filters.applicant_empty" placeholder="申请人" clearable @change="v => hostStore.setFilter('applicant_empty', v)" style="width: 130px">
        <el-option label="未填写" value="1" />
        <el-option label="已填写" value="0" />
      </el-select>
      <el-button @click="hostStore.resetFilters(); resetLocalFilters()">重置筛选</el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useHostStore } from '../stores/host'
import { useCloudResourceStore } from '../stores/cloudResource'
import { exportCSV, downloadTemplate } from '../api/csv'
import { downloadBlob } from '../utils'
import { statusOptions, envTypeOptions, assetTypeOptions } from '../utils'

const hostStore = useHostStore()
const cloudStore = useCloudResourceStore()
const regionOptions = computed(() => cloudStore.allRegions)

const keyword = ref('')
const selectedIds = ref([])

const filters = reactive({
  region: '',
  env_type: '',
  asset_type: '',
  cpu_arch: '',
  is_db_server: '',
  status: '',
  applicant_empty: ''
})

function handleSearch() {
  hostStore.setFilter('keyword', keyword.value)
}

function resetLocalFilters() {
  keyword.value = ''
  Object.keys(filters).forEach(k => filters[k] = '')
}

function openCreateDialog() {
  window.dispatchEvent(new CustomEvent('open-host-form', { detail: { mode: 'create' } }))
}

function openBatchAdd() {
  window.dispatchEvent(new CustomEvent('open-batch-add'))
}

function openBatchEdit() {
  window.dispatchEvent(new CustomEvent('open-batch-edit', { detail: { ids: selectedIds.value } }))
}

function openImport() {
  window.dispatchEvent(new CustomEvent('open-import'))
}

async function handleExport() {
  try {
    const res = await exportCSV()
    downloadBlob(res, 'hosts_export.csv')
  } catch {}
}

async function handleDownloadTemplate() {
  try {
    const res = await downloadTemplate()
    downloadBlob(res, 'import_template.csv')
  } catch {}
}

import { onMounted, onUnmounted } from 'vue'

function handleToolbarSelection(e) {
  selectedIds.value = e.detail.ids
}

onMounted(() => {
  window.addEventListener('selection-change', handleToolbarSelection)
})

onUnmounted(() => {
  window.removeEventListener('selection-change', handleToolbarSelection)
})

defineExpose({ selectedIds })
</script>

<style scoped>
.search-toolbar {
  background: #fff;
  padding: 16px;
  border-radius: 8px;
  margin-bottom: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.toolbar-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.search-box {
  display: flex;
  gap: 8px;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.filter-row {
  display: flex;
  gap: 12px;
  align-items: center;
}
</style>
