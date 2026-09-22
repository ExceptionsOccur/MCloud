<template>
  <div class="business-statistics">
    <el-header class="app-header">
      <div class="header-left">
        <h1>MCloud</h1>
        <span>云平台主机资产管理</span>
      </div>
      <div class="header-center">
        <router-link to="/" class="nav-tab">主机管理</router-link>
        <router-link to="/statistics" class="nav-tab">资源统计</router-link>
        <router-link to="/ip-statistics" class="nav-tab">IP统计</router-link>
        <router-link to="/business-statistics" class="nav-tab active">业务统计</router-link>
      </div>
      <div class="header-right">
        <span v-if="authStore.user" class="user-info">{{ authStore.user.username }}</span>
        <el-dropdown @command="handleCommand">
          <el-button text>
            <el-icon><Setting /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="cloudResource">云资源录入</el-dropdown-item>
              <el-dropdown-item command="changePassword">修改密码</el-dropdown-item>
              <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <div class="main-content" v-loading="businessStore.loading">
      <div class="summary-section">
        <div class="summary-card">
          <span class="sc-value">{{ businessStore.summary.project_count }}</span>
          <span class="sc-label">项目数</span>
        </div>
        <div class="summary-card">
          <span class="sc-value">{{ businessStore.summary.company_count }}</span>
          <span class="sc-label">公司数</span>
        </div>
        <div class="summary-card">
          <span class="sc-value">{{ businessStore.summary.applicant_count }}</span>
          <span class="sc-label">人员数</span>
        </div>
        <div class="summary-card">
          <span class="sc-value">{{ businessStore.summary.host_count }}</span>
          <span class="sc-label">主机总数</span>
        </div>
      </div>

      <div class="control-bar">
        <el-radio-group v-model="dimension" size="small">
          <el-radio-button value="project">按项目</el-radio-button>
          <el-radio-button value="company">按公司</el-radio-button>
          <el-radio-button value="applicant">按人员</el-radio-button>
        </el-radio-group>
        <el-radio-group v-model="metric" size="small">
          <el-radio-button value="host_count">主机数</el-radio-button>
          <el-radio-button value="cpu">vCPU</el-radio-button>
          <el-radio-button value="memory">内存</el-radio-button>
          <el-radio-button value="storage">存储</el-radio-button>
        </el-radio-group>
      </div>

      <div class="chart-panel">
        <v-chart class="chart" :option="chartOption" autoresize @click="handleChartClick" />
      </div>

      <div class="table-panel">
        <el-table :data="tableData" height="100%" size="small" stripe>
          <el-table-column label="#" type="index" width="50" align="center" />
          <el-table-column :label="keyLabel" prop="key" min-width="160" align="left" show-overflow-tooltip />
          <template v-if="dimension === 'project'">
            <el-table-column label="所属公司" min-width="180" align="left" show-overflow-tooltip>
              <template #default="{ row }">{{ row.companies.join('、') || '-' }}</template>
            </el-table-column>
            <el-table-column label="申请人" min-width="100" align="left" show-overflow-tooltip>
              <template #default="{ row }">{{ row.applicants.join('、') || '-' }}</template>
            </el-table-column>
          </template>
          <template v-else-if="dimension === 'company'">
            <el-table-column label="项目数" prop="projectCount" width="80" align="center" sortable />
            <el-table-column label="涉及项目" min-width="200" align="left" show-overflow-tooltip>
              <template #default="{ row }">{{ row.projects.join('、') || '-' }}</template>
            </el-table-column>
            <el-table-column label="人员数" prop="applicantCount" width="80" align="center" />
          </template>
          <template v-else>
            <el-table-column label="所属公司" min-width="180" align="left" show-overflow-tooltip>
              <template #default="{ row }">{{ row.companies.join('、') || '-' }}</template>
            </el-table-column>
            <el-table-column label="项目数" prop="projectCount" width="80" align="center" />
            <el-table-column label="涉及项目" min-width="200" align="left" show-overflow-tooltip>
              <template #default="{ row }">{{ row.projects.join('、') || '-' }}</template>
            </el-table-column>
          </template>
          <el-table-column label="主机数" prop="host_count" width="80" align="center" sortable />
          <el-table-column label="vCPU" prop="cpu" width="90" align="center" sortable />
          <el-table-column label="内存(GB)" prop="memory" width="100" align="center" sortable />
          <el-table-column label="存储(GB)" prop="storage" width="110" align="center" sortable />
          <el-table-column label="运行中" prop="running" width="80" align="center" />
          <el-table-column label="已停止" prop="stopped" width="80" align="center" />
        </el-table>
      </div>
    </div>

    <el-dialog v-model="hostDialogVisible" :title="hostDialogTitle" width="640px" top="8vh">
      <el-table :data="dialogHosts" height="420px" size="small" stripe>
        <el-table-column type="index" label="#" width="50" align="center" />
        <el-table-column prop="name" label="主机名" min-width="180" align="left" show-overflow-tooltip />
        <el-table-column prop="ip" label="内网IP" width="150" align="center" />
        <el-table-column prop="project" label="对应项目" min-width="180" align="left" show-overflow-tooltip />
      </el-table>
      <template #footer>
        <span class="dialog-count">共 {{ dialogHosts.length }} 台主机</span>
      </template>
    </el-dialog>

    <ChangePasswordDialog ref="changePasswordDialog" />
    <CloudResourceDialog ref="cloudResourceDialog" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useBusinessStore } from '../stores/business'
import ChangePasswordDialog from '../components/ChangePasswordDialog.vue'
import CloudResourceDialog from '../components/CloudResourceDialog.vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

use([BarChart, GridComponent, TooltipComponent, CanvasRenderer])

const router = useRouter()
const authStore = useAuthStore()
const businessStore = useBusinessStore()
const changePasswordDialog = ref(null)
const cloudResourceDialog = ref(null)

const dimension = ref('project')
const metric = ref('host_count')
const hostDialogVisible = ref(false)
const hostDialogTitle = ref('')
const dialogHosts = ref([])

const metricLabels = {
  host_count: '主机数',
  cpu: 'vCPU',
  memory: '内存(GB)',
  storage: '存储(GB)'
}

const keyLabel = computed(() => {
  if (dimension.value === 'project') return '项目'
  if (dimension.value === 'company') return '公司'
  return '申请人'
})

const rawData = computed(() => {
  if (dimension.value === 'project') return businessStore.byProject
  if (dimension.value === 'company') return businessStore.byCompany
  return businessStore.byApplicant
})

const tableData = computed(() => {
  return rawData.value.map(item => ({
    ...item,
    projectCount: item.projects ? item.projects.length : 0,
    applicantCount: item.applicants ? item.applicants.length : 0
  }))
})

const chartOption = computed(() => {
  const top = rawData.value.slice(0, 10)
  const names = top.map(d => d.key).reverse()
  const values = top.map(d => d[metric.value] || 0).reverse()

  return {
    grid: { left: 10, right: 60, top: 10, bottom: 10, containLabel: true },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: params => {
        const p = params[0]
        return `${p.name}<br/>${metricLabels[metric.value]}: ${p.value}`
      }
    },
    xAxis: { type: 'value', splitLine: { lineStyle: { type: 'dashed' } } },
    yAxis: {
      type: 'category',
      data: names,
      axisLabel: { fontSize: 11, formatter: v => (v.length > 14 ? v.slice(0, 14) + '…' : v) },
      axisTick: { show: false }
    },
    series: [{
      name: metricLabels[metric.value],
      type: 'bar',
      data: values,
      barMaxWidth: 18,
      itemStyle: { color: '#409eff', borderRadius: [0, 4, 4, 0], cursor: 'pointer' },
      label: { show: true, position: 'right', fontSize: 11, color: '#606266' }
    }]
  }
})

onMounted(() => {
  authStore.fetchUserInfo()
  businessStore.fetchBusinessStats()
})

function handleChartClick(params) {
  const key = params.name
  const group = rawData.value.find(d => d.key === key)
  if (!group) return
  dialogHosts.value = group.hosts || []
  hostDialogTitle.value = `${keyLabel.value}：${key}（${dialogHosts.value.length} 台主机）`
  hostDialogVisible.value = true
}

function handleCommand(command) {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
  } else if (command === 'changePassword') {
    changePasswordDialog.value?.open()
  } else if (command === 'cloudResource') {
    cloudResourceDialog.value?.open()
  }
}
</script>

<style scoped>
.business-statistics {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  height: 60px;
  flex: none;
  background: #fff;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  z-index: 10;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-left h1 {
  font-size: 20px;
  color: #409eff;
  font-weight: 700;
}

.header-left span {
  font-size: 14px;
  color: #909399;
}

.header-center {
  display: flex;
  gap: 4px;
}

.nav-tab {
  padding: 6px 16px;
  font-size: 14px;
  color: #606266;
  text-decoration: none;
  border-radius: 4px;
  transition: all 0.2s;
}

.nav-tab:hover {
  color: #409eff;
  background: #ecf5ff;
}

.nav-tab.active {
  color: #409eff;
  font-weight: 600;
  background: #ecf5ff;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-info {
  font-size: 14px;
  color: #606266;
}

.main-content {
  flex: 1;
  min-height: 0;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.summary-section {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  flex: none;
}

.summary-card {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
  padding: 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.sc-value {
  font-size: 24px;
  font-weight: 700;
  color: #409eff;
  line-height: 1.2;
}

.sc-label {
  font-size: 12px;
  color: #909399;
}

.control-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  flex: none;
  flex-wrap: wrap;
}

.chart-panel {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
  padding: 8px;
  height: 260px;
  flex: none;
}

.chart {
  width: 100%;
  height: 100%;
}

.table-panel {
  flex: 1;
  min-height: 0;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
  padding: 8px;
}

.dialog-count {
  font-size: 13px;
  color: #909399;
}

@media (max-width: 900px) {
  .summary-section {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
