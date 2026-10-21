<template>
  <el-header
    class="app-header"
    :class="{ 'app-header--backup': navVariant === 'backup' }"
  >
    <div class="header-left">
      <h1>MCloud</h1>
      <span>云平台资产管理</span>
    </div>
    <AppNav :variant="navVariant" />
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
            <el-dropdown-item
              v-for="item in settingItems"
              :key="item.command"
              :command="item.command"
              :disabled="$route.path === item.path"
            >
              {{ item.label }}
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

    <ChangePasswordDialog ref="changePasswordDialog" />
    <CloudResourceDialog ref="cloudResourceDialog" />
  </el-header>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import AppNav from './AppNav.vue'
import ChangePasswordDialog from './ChangePasswordDialog.vue'
import CloudResourceDialog from './CloudResourceDialog.vue'

defineProps({
  // 后端数据备份页历史导航样式与其他页存在差异，透传给 AppNav 保证视觉零变化（T-019/T-089）
  navVariant: {
    type: String,
    default: 'default',
    validator: (v) => ['default', 'backup'].includes(v)
  }
})

const router = useRouter()
const authStore = useAuthStore()

const changePasswordDialog = ref(null)
const cloudResourceDialog = ref(null)

// 设置菜单导航项：禁用态统一按当前路由自动判断，无需各页手写（T-089）
const settingItems = [
  { command: 'personnel', label: '人员录入', path: '/personnel' },
  { command: 'publicIP', label: '公网IP录入', path: '/public-ip' },
  { command: 'dataBackup', label: '数据备份', path: '/data-backup' },
  { command: 'aiSettings', label: 'AI 接入设置', path: '/ai-settings' },
  { command: 'aiImport', label: 'AI 申请入账', path: '/ai-import' }
]

onMounted(() => {
  authStore.fetchUserInfo()
})

function handleCommand(command) {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
    return
  }
  if (command === 'changePassword') {
    changePasswordDialog.value?.open()
    return
  }
  if (command === 'cloudResource') {
    cloudResourceDialog.value?.open()
    return
  }
  const item = settingItems.find((i) => i.command === command)
  if (item) router.push(item.path)
}
</script>

<style scoped>
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

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-info {
  font-size: 14px;
  color: #606266;
}

/* 数据备份页历史样式变体（T-019/T-089）：边框替代阴影、标题字号与间距不同 */
.app-header--backup {
  background: #fff;
  box-shadow: none;
  border-bottom: 1px solid #e4e7ed;
}

.app-header--backup .header-left h1 {
  display: inline;
  font-size: 18px;
  margin: 0 8px 0 0;
}

.app-header--backup .header-left span {
  color: #909399;
  font-size: 13px;
}

.app-header--backup .user-info {
  color: #606266;
  font-size: 14px;
}
</style>
