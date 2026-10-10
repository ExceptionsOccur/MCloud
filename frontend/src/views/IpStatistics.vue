<template>
  <div class="ip-statistics">
    <el-header class="app-header">
      <div class="header-left">
        <h1>MCloud</h1>
        <span>云平台主机资产管理</span>
      </div>
      <div class="header-center">
        <router-link
          to="/"
          class="nav-tab"
        >
          主机管理
        </router-link>
        <router-link
          to="/statistics"
          class="nav-tab"
        >
          资源统计
        </router-link>
        <router-link
          to="/ip-statistics"
          class="nav-tab active"
        >
          IP统计
        </router-link>
        <router-link
          to="/business-statistics"
          class="nav-tab"
        >
          业务统计
        </router-link>
        <router-link
          to="/zero-trust"
          class="nav-tab"
        >
          零信任
        </router-link>
        <router-link
          to="/mapping-ledger"
          class="nav-tab"
        >
          映射
        </router-link>
        <router-link
          to="/audit-logs"
          class="nav-tab"
        >
          审计日志
        </router-link>
      </div>
      <div class="header-right">
        <span
          v-if="authStore.user"
          class="user-info"
        >{{ authStore.user.username }}</span>
        <el-dropdown @command="handleCommand">
          <el-button text>
            <el-icon><Setting /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="cloudResource">
                云资源录入
              </el-dropdown-item>
              <el-dropdown-item command="personnel">
                人员录入
              </el-dropdown-item>
              <el-dropdown-item
                command="publicIP"
                :disabled="$route.path === '/public-ip'"
              >
                公网IP录入
              </el-dropdown-item>
              <el-dropdown-item command="dataBackup">
                数据备份
              </el-dropdown-item>
              <el-dropdown-item command="changePassword">
                修改密码
              </el-dropdown-item>
              <el-dropdown-item
                command="logout"
                divided
              >
                退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <div
      v-loading="statsStore.loading"
      class="main-content"
    >
      <template v-if="!statsStore.loading">
        <ProbePanel
          :ws-connected="wsConnected"
          @manage="subnetDialog?.open()"
        />
        <div
          v-if="statsStore.ipUsage.length === 0"
          class="empty-state"
        >
          暂无 IP 网段，请点击右上角「管理网段」添加
        </div>
        <div class="subnet-grid">
          <SubnetCard
            v-for="subnet in statsStore.ipUsage"
            :key="subnet.subnet"
            :subnet="subnet"
            :cell-colors="cellColors"
            :probing="probing"
            :batch-state="batchState"
            :batch-tested="batchTested"
            :counts="subnetCounts[subnet.subnet]"
            @cell-click="handleCellClick"
            @batch-test="handleBatchTest"
          />
        </div>
      </template>
    </div>

    <ChangePasswordDialog ref="changePasswordDialog" />
    <CloudResourceDialog ref="cloudResourceDialog" />
    <SubnetManageDialog
      ref="subnetDialog"
      @changed="onSubnetChanged"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useStatsStore } from '../stores/stats'
import { useProbeWebSocket } from '../composables/useProbeWebSocket'
import { useProbeGrid } from '../composables/useProbeGrid'
import ProbePanel from '../components/ip/ProbePanel.vue'
import SubnetCard from '../components/ip/SubnetCard.vue'
import SubnetManageDialog from '../components/ip/SubnetManageDialog.vue'
import ChangePasswordDialog from '../components/ChangePasswordDialog.vue'
import CloudResourceDialog from '../components/CloudResourceDialog.vue'

const router = useRouter()
const authStore = useAuthStore()
const statsStore = useStatsStore()
const changePasswordDialog = ref(null)
const cloudResourceDialog = ref(null)
const subnetDialog = ref(null)

let grid = null
const { wsConnected, connect: connectWS, send: sendProbe, isUnmounted } = useProbeWebSocket(data => grid?.applyProbeResult(data))
grid = useProbeGrid({ getSend: () => sendProbe, isUnmounted, getSubnets: () => statsStore.ipUsage })
const { probing, cellColors, batchState, batchTested, subnetCounts, initCellColors, handleBatchTest, handleCellClick } = grid

onMounted(async () => {
  authStore.fetchUserInfo()
  connectWS()
  try {
    await statsStore.fetchIPUsage()
    initCellColors()
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  }
})

function handleCommand(command) {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
  } else if (command === 'changePassword') {
    changePasswordDialog.value?.open()
  } else if (command === 'dataBackup') {
    router.push('/data-backup')
  } else if (command === 'cloudResource') {
    cloudResourceDialog.value?.open()
  } else if (command === 'personnel') {
    router.push('/personnel')
  } else if (command === 'publicIP') {
    router.push('/public-ip')
  }
}

async function onSubnetChanged() {
  await statsStore.fetchIPUsage()
  initCellColors()
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
