<template>
  <div class="host-management">
    <el-header class="app-header">
      <div class="header-left">
        <h1>MCloud</h1>
        <span>云平台资产管理</span>
      </div>
      <div class="header-center">
        <router-link to="/" class="nav-tab" :class="{ active: $route.path === '/' }">主机管理</router-link>
        <router-link to="/statistics" class="nav-tab" :class="{ active: $route.path === '/statistics' }">资源统计</router-link>
        <router-link to="/ip-statistics" class="nav-tab" :class="{ active: $route.path === '/ip-statistics' }">IP统计</router-link>
        <router-link to="/business-statistics" class="nav-tab" :class="{ active: $route.path === '/business-statistics' }">业务统计</router-link>
        <router-link to="/zero-trust" class="nav-tab" :class="{ active: $route.path === '/zero-trust' }">零信任</router-link>
        <router-link to="/domain-ledger" class="nav-tab" :class="{ active: $route.path === '/domain-ledger' }">域名</router-link>
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

    <div class="main-content">
      <SearchToolbar />
      <HostTable />
    </div>

    <HostFormDialog ref="hostFormDialog" />
    <BatchAddDialog ref="batchAddDialog" />
    <BatchEditDialog ref="batchEditDialog" />
    <ImportDialog ref="importDialog" />
    <ChangePasswordDialog ref="changePasswordDialog" />
    <CloudResourceDialog ref="cloudResourceDialog" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useHostStore } from '../stores/host'
import { useCloudResourceStore } from '../stores/cloudResource'
import SearchToolbar from '../components/SearchToolbar.vue'
import HostTable from '../components/HostTable.vue'
import HostFormDialog from '../components/HostFormDialog.vue'
import BatchAddDialog from '../components/BatchAddDialog.vue'
import BatchEditDialog from '../components/BatchEditDialog.vue'
import ImportDialog from '../components/ImportDialog.vue'
import ChangePasswordDialog from '../components/ChangePasswordDialog.vue'
import CloudResourceDialog from '../components/CloudResourceDialog.vue'

const router = useRouter()
const authStore = useAuthStore()
const hostStore = useHostStore()
const cloudStore = useCloudResourceStore()

const hostFormDialog = ref(null)
const batchAddDialog = ref(null)
const batchEditDialog = ref(null)
const importDialog = ref(null)
const changePasswordDialog = ref(null)
const cloudResourceDialog = ref(null)

onMounted(() => {
  authStore.fetchUserInfo()
  hostStore.fetchHosts()
  cloudStore.fetchRegions()
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
</script>

<style scoped>
.host-management {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
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
  padding: 20px;
  overflow: auto;
}
</style>
