<template>
  <div class="ip-statistics">
    <el-header class="app-header">
      <div class="header-left">
        <h1>MCloud</h1>
        <span>云平台主机资产管理</span>
      </div>
      <div class="header-center">
        <router-link to="/" class="nav-tab">主机管理</router-link>
        <router-link to="/statistics" class="nav-tab">资源统计</router-link>
        <router-link to="/ip-statistics" class="nav-tab active">IP统计</router-link>
        <router-link to="/business-statistics" class="nav-tab">业务统计</router-link>
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
              <el-dropdown-item command="changePassword">修改密码</el-dropdown-item>
              <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <div class="main-content" v-loading="statsStore.loading">
      <template v-if="!statsStore.loading">
        <div class="top-bar">
          <div class="ws-status" :class="wsConnected ? 'online' : 'offline'">
            <em></em>{{ wsConnected ? '探测服务已连接' : '探测服务连接中...' }}
          </div>
          <el-button size="small" @click="subnetDialog?.open()">管理网段</el-button>
        </div>
        <div v-if="statsStore.ipUsage.length === 0" class="empty-state">
          暂无 IP 网段，请点击右上角「管理网段」添加
        </div>
        <div class="subnet-grid">
          <div v-for="subnet in statsStore.ipUsage" :key="subnet.subnet" class="subnet-card">
            <div class="subnet-header">
              <h3>{{ subnet.subnet }}</h3>
              <el-button
                size="small"
                type="primary"
                :loading="isBatchRunning(subnet)"
                :disabled="isBatchRunning(subnet)"
                @click="handleBatchTest(subnet)"
              >全量测试</el-button>
            </div>
            <div class="subnet-stats">
              <span class="stat"><em class="used-dot"></em>已用 {{ getUsedCount(subnet) }}</span>
              <span class="stat"><em class="empty-dot"></em>空记录 {{ getEmptyCount(subnet) }}</span>
              <span class="stat"><em class="unused-dot"></em>未用 {{ getUnusedCount(subnet) }}</span>
            </div>
            <div v-if="isBatchRunning(subnet)" class="batch-progress">
              <el-progress
                :percentage="getBatchPercent(subnet)"
                :stroke-width="8"
                :show-text="false"
                striped
                striped-flow
              />
              <span class="batch-text">{{ getBatchDone(subnet) }} / 256</span>
            </div>
            <div class="bitmap-wrapper">
              <div class="col-labels">
                <span v-for="c in 10" :key="c">{{ c - 1 }}</span>
              </div>
              <div class="bitmap-body">
                <div v-for="row in 26" :key="'r'+row" class="bitmap-row">
                  <span class="row-label">{{ row - 1 }}</span>
                  <div
                    v-for="col in 10" :key="'c'+col"
                    class="bit-cell"
                    :class="getCellClasses(subnet, (row - 1) * 10 + (col - 1))"
                    :title="getIP(subnet.subnet, (row - 1) * 10 + (col - 1))"
                    @click="handleCellClick(subnet, (row - 1) * 10 + (col - 1))"
                  ></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>

    <ChangePasswordDialog ref="changePasswordDialog" />
    <CloudResourceDialog ref="cloudResourceDialog" />
    <SubnetManageDialog ref="subnetDialog" @changed="onSubnetChanged" />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useStatsStore } from '../stores/stats'
import { probeIP as probeIPHttp } from '../api/stats'
import ChangePasswordDialog from '../components/ChangePasswordDialog.vue'
import CloudResourceDialog from '../components/CloudResourceDialog.vue'
import SubnetManageDialog from '../components/SubnetManageDialog.vue'
import { ElMessage } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()
const statsStore = useStatsStore()
const changePasswordDialog = ref(null)
const cloudResourceDialog = ref(null)
const subnetDialog = ref(null)
const probing = reactive({})
const cellColors = reactive({})
const batchState = reactive({})
const batchTested = reactive({})
const wsConnected = ref(false)
let ws = null
let reconnectTimer = null
let unmounted = false
let suppressMessages = false
const probeResolvers = {}

const subnetCounts = computed(() => {
  const counts = {}
  for (const subnet of statsStore.ipUsage) {
    let red = 0, yellow = 0, green = 0
    for (let i = 0; i < 256; i++) {
      const ip = getIP(subnet.subnet, i)
      const c = cellColors[ip]
      if (c === 'red') red++
      else if (c === 'yellow') yellow++
      else green++
    }
    counts[subnet.subnet] = { red, yellow, green }
  }
  return counts
})

onMounted(async () => {
  authStore.fetchUserInfo()
  connectWS()
  try {
    await statsStore.fetchIPUsage()
    initCellColors()
  } catch (e) {
    console.error('加载IP使用情况失败:', e)
  }
})

onUnmounted(() => {
  unmounted = true
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  if (ws) {
    ws.onclose = null
    ws.close()
    ws = null
  }
  Object.keys(probeResolvers).forEach(ip => {
    probeResolvers[ip]()
    delete probeResolvers[ip]
  })
})

function connectWS() {
  if (unmounted) return
  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const token = authStore.token
  const url = `${protocol}//${location.host}/api/ws/probe?token=${encodeURIComponent(token)}`

  try {
    ws = new WebSocket(url)
  } catch (e) {
    console.error('创建WebSocket失败:', e)
    scheduleReconnect()
    return
  }

  ws.onopen = () => {
    wsConnected.value = true
    console.log('WebSocket已连接')
  }

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      applyProbeResult(data)
    } catch (e) {
      console.error('解析WebSocket消息失败:', e)
    }
  }

  ws.onclose = () => {
    wsConnected.value = false
    console.log('WebSocket已断开，3秒后重连')
    scheduleReconnect()
  }

  ws.onerror = (err) => {
    console.error('WebSocket错误:', err)
    if (ws) ws.close()
  }
}

function scheduleReconnect() {
  if (unmounted || reconnectTimer) return
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null
    connectWS()
  }, 3000)
}

function applyProbeResult(data) {
  if (!data || !data.ip || !data.color) return
  cellColors[data.ip] = data.color
  probing[data.ip] = false

  if (!suppressMessages) {
    if (data.color === 'red') {
      ElMessage.error(`${data.ip} 已确认使用`)
    } else if (data.color === 'yellow') {
      ElMessage.warning(`${data.ip} 已标记为空记录`)
    } else {
      ElMessage.success(`${data.ip} 探测完成，IP未使用`)
    }
  }

  const resolver = probeResolvers[data.ip]
  if (resolver) {
    delete probeResolvers[data.ip]
    resolver()
  }
}

function sendProbe(payload) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(payload))
    return true
  }
  return false
}

function initCellColors() {
  for (const subnet of statsStore.ipUsage) {
    for (let i = 0; i < 256; i++) {
      const ip = getIP(subnet.subnet, i)
      if (subnet.used_ips && subnet.used_ips.includes(ip)) {
        cellColors[ip] = 'red'
      } else if (subnet.empty_ips && subnet.empty_ips.includes(ip)) {
        cellColors[ip] = 'yellow'
      } else {
        cellColors[ip] = 'green'
      }
    }
  }
}

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
  }
}

async function onSubnetChanged() {
  await statsStore.fetchIPUsage()
  initCellColors()
}

function getIP(subnet, offset) {
  const parts = subnet.split('.')
  const third = parseInt(parts[2])
  const fourth = offset
  return `${parts[0]}.${parts[1]}.${third}.${fourth}`
}

function getCellClasses(subnet, offset) {
  if (offset > 255) return { empty: true }
  const ip = getIP(subnet.subnet, offset)
  const color = cellColors[ip]
  return {
    used: color === 'red',
    unused: color === 'green',
    alive: color === 'yellow',
    probing: !!probing[ip],
    tested: !!batchTested[ip]
  }
}

function getUsedCount(subnet) {
  return subnetCounts.value[subnet.subnet]?.red ?? 0
}

function getEmptyCount(subnet) {
  return subnetCounts.value[subnet.subnet]?.yellow ?? 0
}

function getUnusedCount(subnet) {
  return subnetCounts.value[subnet.subnet]?.green ?? 0
}

function isBatchRunning(subnet) {
  return !!batchState[subnet.subnet]?.running
}

function getBatchDone(subnet) {
  return batchState[subnet.subnet]?.done ?? 0
}

function getBatchPercent(subnet) {
  const state = batchState[subnet.subnet]
  if (!state) return 0
  return Math.min(100, Math.round((state.done / state.total) * 100))
}

function probeAndWait(ip) {
  return new Promise((resolve) => {
    let settled = false
    const finish = () => {
      if (settled) return
      settled = true
      clearTimeout(timer)
      delete probeResolvers[ip]
      resolve()
    }
    const timer = setTimeout(() => {
      probing[ip] = false
      finish()
    }, 10000)

    probeResolvers[ip] = finish

    probing[ip] = true
    const color = cellColors[ip] || 'green'
    const payload = { ip, color }

    if (sendProbe(payload)) return

    probeIPHttp(ip, color)
      .then(res => {
        if (res.code === 0) {
          applyProbeResult({ ip, color: res.data.color })
        } else {
          probing[ip] = false
        }
        finish()
      })
      .catch(() => {
        probing[ip] = false
        finish()
      })
  })
}

async function handleBatchTest(subnet) {
  const name = subnet.subnet
  if (batchState[name]?.running) return

  const ips = []
  for (let i = 0; i < 256; i++) {
    ips.push(getIP(name, i))
  }

  ips.forEach(ip => { delete batchTested[ip] })
  batchState[name] = { total: 256, done: 0, running: true }
  suppressMessages = true

  const concurrency = 32
  let index = 0
  const worker = async () => {
    while (index < ips.length) {
      if (unmounted) return
      const ip = ips[index++]
      try {
        await probeAndWait(ip)
      } catch (e) {
        // ignore single failure
      }
      batchState[name].done++
      batchTested[ip] = true
    }
  }

  try {
    await Promise.all(Array.from({ length: concurrency }, worker))
    ElMessage.success(`${name} 全量测试完成`)
  } finally {
    if (batchState[name]) {
      batchState[name].running = false
    }
    suppressMessages = false
    setTimeout(() => {
      ips.forEach(ip => { delete batchTested[ip] })
    }, 2000)
  }
}

async function handleCellClick(subnet, offset) {
  if (offset > 255) return
  const ip = getIP(subnet.subnet, offset)
  if (probing[ip]) return

  probing[ip] = true
  const currentColor = cellColors[ip] || 'green'
  const payload = { ip, color: currentColor }

  if (sendProbe(payload)) return

  try {
    const res = await probeIPHttp(ip, currentColor)
    if (res.code === 0) {
      applyProbeResult({ ip, color: res.data.color })
    } else {
      probing[ip] = false
    }
  } catch (e) {
    probing[ip] = false
    ElMessage.error(`${ip} 探测失败`)
  }
}
</script>

<style scoped>
.ip-statistics {
  min-height: 100vh;
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
  padding: 16px;
  overflow: auto;
  width: 100%;
  min-width: 0;
}

.top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.ws-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 4px;
}

.ws-status em {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  font-style: normal;
}

.ws-status.online {
  color: #67c23a;
  background: #f0f9eb;
}

.ws-status.online em {
  background: #67c23a;
}

.ws-status.offline {
  color: #e6a23c;
  background: #fdf6ec;
}

.ws-status.offline em {
  background: #e6a23c;
}

.subnet-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 16px;
}

.empty-state {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
  padding: 60px 20px;
  text-align: center;
  color: #909399;
  font-size: 14px;
}

.subnet-card {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
  padding: 16px;
  min-width: 0;
}

.subnet-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.subnet-header h3 {
  font-size: 15px;
  color: #303133;
  font-weight: 600;
  white-space: nowrap;
}

.batch-progress {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}

.batch-progress :deep(.el-progress) {
  flex: 1;
}

.batch-text {
  font-size: 12px;
  color: #909399;
  white-space: nowrap;
}

.subnet-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  font-size: 12px;
  color: #909399;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #ebeef5;
}

.stat {
  display: flex;
  align-items: center;
  gap: 4px;
}

.used-dot, .unused-dot, .empty-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 2px;
  font-style: normal;
}

.used-dot {
  background: #f56c6c;
}

.unused-dot {
  background: #67c23a;
}

.empty-dot {
  background: #e6a23c;
}

.bitmap-wrapper {
  display: flex;
  flex-direction: column;
}

.col-labels {
  display: flex;
  gap: 1px;
  padding-left: 20px;
}

.col-labels span {
  width: 20px;
  text-align: center;
  font-size: 10px;
  color: #909399;
}

.bitmap-body {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.bitmap-row {
  display: flex;
  align-items: center;
  gap: 1px;
}

.row-label {
  width: 18px;
  text-align: right;
  font-size: 10px;
  color: #909399;
  padding-right: 2px;
}

.bit-cell {
  width: 20px;
  height: 20px;
  border-radius: 2px;
  cursor: pointer;
  transition: opacity 0.15s;
}

.bit-cell:hover {
  opacity: 0.7;
}

.bit-cell.used {
  background: #f56c6c;
}

.bit-cell.unused {
  background: #67c23a;
}

.bit-cell.alive {
  background: #e6a23c;
}

.bit-cell.tested {
  box-shadow: inset 0 0 0 2px #409eff;
}

.bit-cell.probing {
  animation: probing-pulse 0.8s ease-in-out infinite;
  box-shadow: inset 0 0 0 2px #409eff;
}

@keyframes probing-pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.35;
    transform: scale(0.85);
  }
}

.bit-cell.empty {
  background: transparent;
}

@media (max-width: 1200px) {
  .subnet-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .subnet-grid {
    grid-template-columns: 1fr;
  }
}
</style>
