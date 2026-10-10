<template>
  <div
    class="header-center"
    :class="`header-center--${variant}`"
  >
    <router-link
      v-for="item in navItems"
      :key="item.to"
      :to="item.to"
      class="nav-tab"
      :class="{ active: $route.path === item.to }"
    >
      {{ item.label }}
    </router-link>
  </div>
</template>

<script setup>
defineProps({
  // audit/backup 两页历史样式与其他页存在差异，保留 variant 以保证抽取后视觉零变化（T-019）
  variant: {
    type: String,
    default: 'default',
    validator: (v) => ['default', 'audit', 'backup'].includes(v)
  }
})

const navItems = [
  { to: '/', label: '主机管理' },
  { to: '/statistics', label: '资源统计' },
  { to: '/ip-statistics', label: 'IP统计' },
  { to: '/business-statistics', label: '业务统计' },
  { to: '/zero-trust', label: '零信任' },
  { to: '/mapping-ledger', label: '映射' },
  { to: '/audit-logs', label: '审计日志' }
]
</script>

<style scoped>
.header-center {
  display: flex;
  gap: 4px;
}

/* default：与原 8 个页面样式一致 */
.header-center--default .nav-tab {
  padding: 6px 16px;
  font-size: 14px;
  color: #606266;
  text-decoration: none;
  border-radius: 4px;
  transition: all 0.2s;
}

.header-center--default .nav-tab:hover {
  color: #409eff;
  background: #ecf5ff;
}

.header-center--default .nav-tab.active {
  color: #409eff;
  font-weight: 600;
  background: #ecf5ff;
}

/* audit：AuditLog 页原样式（无 hover、padding 8/14） */
.header-center--audit .nav-tab {
  padding: 8px 14px;
  color: #606266;
  text-decoration: none;
  border-radius: 4px;
  font-size: 14px;
}

.header-center--audit .nav-tab.active {
  color: #409eff;
  background: #ecf5ff;
  font-weight: 600;
}

/* backup：DataBackup 页原样式（gap 20、无 padding、active 无背景、无 hover） */
.header-center--backup {
  gap: 20px;
}

.header-center--backup .nav-tab {
  color: #606266;
  text-decoration: none;
  font-size: 14px;
}

.header-center--backup .nav-tab.active {
  color: #409eff;
  font-weight: 600;
}
</style>
