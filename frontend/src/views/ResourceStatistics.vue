<template>
  <div class="resource-statistics" :class="{ 'full-height': regions.length < 3 }">
    <el-header class="app-header">
      <div class="header-left">
        <h1>MCloud</h1>
        <span>云平台主机资产管理</span>
      </div>
      <div class="header-center">
        <router-link to="/" class="nav-tab">主机管理</router-link>
        <router-link to="/statistics" class="nav-tab active">资源统计</router-link>
        <router-link to="/ip-statistics" class="nav-tab">IP统计</router-link>
        <router-link to="/business-statistics" class="nav-tab">业务统计</router-link>
        <router-link to="/zero-trust" class="nav-tab">零信任</router-link>
        <router-link to="/domain-ledger" class="nav-tab">域名</router-link>
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
              <el-dropdown-item command="personnel">人员录入</el-dropdown-item>
              <el-dropdown-item command="zeroTrust">零信任台账</el-dropdown-item>
              <el-dropdown-item command="domain">域名台账</el-dropdown-item>
              <el-dropdown-item command="changePassword">修改密码</el-dropdown-item>
              <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <div class="main-content" v-loading="cloudStore.loading">
      <div v-if="cloudStore.loading" class="loading-placeholder"></div>
      <template v-else>
        <div class="summary-section" :style="{ gridTemplateColumns: gridCols }">
          <div v-for="region in regions" :key="region" class="summary-card">
            <div class="card-header">
              <h3>{{ region }}</h3>
              <span class="card-subtitle">云端总资源</span>
            </div>
            <div class="resource-table">
              <div class="rt-row rt-head">
                <span class="rt-name">资源</span>
                <span class="rt-col">总量</span>
                <span class="rt-col">已用</span>
                <span class="rt-col">剩余</span>
                <span class="rt-col rt-running-col">运行中</span>
                <span class="rt-col rt-stopped-col">已停止</span>
              </div>
              <div v-for="row in resourceRows" :key="row.name" class="rt-row">
                <span class="rt-name">{{ row.name }}<em>({{ row.unit }})</em></span>
                <span class="rt-col rt-total">{{ getTotalField(region, row.totalField) }}</span>
                <span class="rt-col rt-used">{{ row.hasUsage ? getUsedField(region, row.usedField) : '-' }}</span>
                <span class="rt-col rt-remain">{{ getRemainDisplay(region, row) }}</span>
                <span class="rt-col rt-running-col">{{ row.hasUsage ? getRunningField(region, row.usedField) : '-' }}</span>
                <span class="rt-col rt-stopped-col">{{ row.hasUsage ? getStoppedField(region, row.usedField) : '-' }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="chart-columns" :style="{ gridTemplateColumns: gridCols }">
          <div v-for="region in regions" :key="region" class="chart-column">
            <h2 class="section-title">{{ region }}资源使用情况</h2>
            <div class="chart-grid">
              <div class="chart-card">
                <h4>vCPU</h4>
                <v-chart class="chart" :option="getPieOption(region, 'vcpu', 'cpu', 'CPU')" autoresize />
              </div>
              <div class="chart-card">
                <h4>内存</h4>
                <v-chart class="chart" :option="getPieOption(region, 'memory', 'memory', '内存')" autoresize />
              </div>
              <div class="chart-card">
                <h4>存储</h4>
                <v-chart class="chart" :option="getPieOption(region, 'storage', 'storage', '存储')" autoresize />
              </div>
              <div class="chart-card">
                <h4>裸金属</h4>
                <v-chart class="chart" :option="getPieOption(region, 'bare_metal', 'bare_metal', '裸金属')" autoresize />
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>

    <ChangePasswordDialog ref="changePasswordDialog" />
    <CloudResourceDialog ref="cloudResourceDialog" @saved="onResourceSaved" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useCloudResourceStore } from '../stores/cloudResource'
import ChangePasswordDialog from '../components/ChangePasswordDialog.vue'
import CloudResourceDialog from '../components/CloudResourceDialog.vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { PieChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

use([PieChart, TitleComponent, TooltipComponent, LegendComponent, CanvasRenderer])

const router = useRouter()
const authStore = useAuthStore()
const cloudStore = useCloudResourceStore()
const changePasswordDialog = ref(null)
const cloudResourceDialog = ref(null)

const regions = computed(() => cloudStore.allRegions)
const gridCols = computed(() => `repeat(${Math.max(1, Math.min(regions.value.length, 2))}, 1fr)`)

const resourceRows = [
  { name: 'vCPU', unit: '核', totalField: 'vcpu', usedField: 'cpu', hasUsage: true },
  { name: '内存', unit: 'GB', totalField: 'memory', usedField: 'memory', hasUsage: true },
  { name: '存储', unit: 'GB', totalField: 'storage', usedField: 'storage', hasUsage: true },
  { name: '物理CPU', unit: '核', totalField: 'physical_cpu', hasUsage: false },
  { name: '裸金属', unit: '台', totalField: 'bare_metal', usedField: 'bare_metal', hasUsage: true },
  { name: 'GPU卡数', unit: '卡', totalField: 'gpu_card_count', hasUsage: false }
]

onMounted(() => {
  authStore.fetchUserInfo()
  cloudStore.fetchStatistics()
})

function handleCommand(command) {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
  } else if (command === 'changePassword') {
    changePasswordDialog.value?.open()
  } else if (command === 'cloudResource') {
    cloudResourceDialog.value?.open()
  } else if (command === 'personnel') {
    router.push('/personnel')
  } else if (command === 'zeroTrust') {
    router.push('/zero-trust')
  } else if (command === 'domain') {
    router.push('/domain-ledger')
  }
}

function onResourceSaved() {
  cloudStore.fetchStatistics()
}

function getTotalField(region, field) {
  const res = cloudStore.getRegionResource(region)
  return res ? (res[field] || 0) : 0
}

function getUsedField(region, field) {
  const used = cloudStore.getRegionUsed(region)
  return used[field] || 0
}

function getRunningField(region, field) {
  const running = cloudStore.getRegionRunning(region)
  return running[field] || 0
}

function getStoppedField(region, field) {
  const stopped = cloudStore.getRegionStopped(region)
  return stopped[field] || 0
}

function getRemainField(region, totalField, usedField) {
  const total = getTotalField(region, totalField)
  const used = getUsedField(region, usedField)
  return Math.max(0, total - used)
}

function getRemainDisplay(region, row) {
  if (!row.hasUsage) return '-'
  const total = getTotalField(region, row.totalField)
  const remain = getRemainField(region, row.totalField, row.usedField)
  const pct = total > 0 ? ((remain / total) * 100).toFixed(1) : '0.0'
  return `${remain} (${pct}%)`
}

function getPieOption(region, totalField, field, label) {
  const total = getTotalField(region, totalField)
  const runningValue = getRunningField(region, field)
  const stoppedValue = getStoppedField(region, field)
  const unusedValue = Math.max(0, total - runningValue - stoppedValue)

  if (total === 0) {
    return {
      graphic: {
        type: 'text',
        left: 'center',
        top: 'middle',
        style: { text: '暂无数据', fontSize: 14, fill: '#999' }
      }
    }
  }

  return {
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} {d}%'
    },
    legend: {
      bottom: 0,
      data: ['运行中', '已停止', '未使用']
    },
    series: [{
      name: label,
      type: 'pie',
      radius: ['40%', '65%'],
      center: ['50%', '45%'],
      avoidLabelOverlap: true,
      itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
      label: {
        show: true,
        formatter: '{b}\n{c} ({d}%)'
      },
      data: [
        { value: runningValue, name: '运行中', itemStyle: { color: '#409eff' } },
        { value: stoppedValue, name: '已停止', itemStyle: { color: '#e6a23c' } },
        { value: unusedValue, name: '未使用', itemStyle: { color: '#e4e7ed' } }
      ]
    }]
  }
}
</script>

<style scoped>
.resource-statistics {
  display: flex;
  flex-direction: column;
}

.resource-statistics.full-height {
  height: 100vh;
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
  width: 100%;
  min-width: 0;
}

.loading-placeholder {
  min-height: 400px;
}

.summary-section {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-bottom: 12px;
  flex: none;
}

.summary-card {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
  padding: 16px;
  min-width: 0;
}

.card-header {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 8px;
  padding-bottom: 6px;
  border-bottom: 1px solid #ebeef5;
}

.card-header h3 {
  font-size: 18px;
  color: #303133;
  font-weight: 600;
}

.card-subtitle {
  font-size: 13px;
  color: #909399;
}

.resource-table {
  display: flex;
  flex-direction: column;
}

.rt-row {
  display: grid;
  grid-template-columns: 1.2fr 1fr 1fr 1.5fr 1fr 1fr;
  align-items: center;
  padding: 5px 0;
  border-bottom: 1px solid #f2f6fc;
  font-size: 12px;
}

.rt-row:last-child {
  border-bottom: none;
}

.rt-head {
  font-size: 12px;
  color: #909399;
  border-bottom: 1px solid #ebeef5;
}

.rt-name {
  color: #606266;
  text-align: center;
}

.rt-name em {
  font-style: normal;
  font-size: 11px;
  color: #c0c4cc;
  margin-left: 2px;
}

.rt-col {
  text-align: center;
  font-variant-numeric: tabular-nums;
}

.rt-total {
  color: #303133;
  font-weight: 600;
}

.rt-used {
  color: #606266;
}

.rt-remain {
  color: #67c23a;
  font-weight: 500;
}

.rt-running-col {
  color: #409eff;
}

.rt-stopped-col {
  color: #e6a23c;
}

.chart-columns {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  flex: 1;
  min-height: 0;
  min-width: 0;
}

.chart-column {
  display: flex;
  flex-direction: column;
  min-height: 0;
  min-width: 0;
}

.section-title {
  font-size: 15px;
  color: #303133;
  font-weight: 600;
  margin-bottom: 8px;
  padding-left: 12px;
  border-left: 3px solid #409eff;
  flex: none;
}

.chart-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  grid-template-rows: repeat(2, 1fr);
  gap: 12px;
  flex: 1;
  min-height: 0;
  min-width: 0;
}

.chart-card {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
  padding: 8px;
  text-align: center;
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.chart-card h4 {
  font-size: 12px;
  color: #606266;
  margin-bottom: 2px;
  font-weight: 500;
  flex: none;
}

.chart {
  width: 100%;
  flex: 1;
  min-height: 0;
}

@media (max-width: 900px) {
  .chart-columns {
    grid-template-columns: 1fr;
  }
  .summary-section {
    grid-template-columns: 1fr;
  }
}
</style>
