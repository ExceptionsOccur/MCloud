<template>
  <div class="audit-log">
    <el-header class="app-header">
      <div class="header-left">
        <h1>MCloud</h1>
        <span>云平台资产管理</span>
      </div>
      <div class="header-center">
        <router-link
          to="/"
          class="nav-tab"
          :class="{ active: $route.path === '/' }"
        >
          主机管理
        </router-link>
        <router-link
          to="/statistics"
          class="nav-tab"
          :class="{ active: $route.path === '/statistics' }"
        >
          资源统计
        </router-link>
        <router-link
          to="/ip-statistics"
          class="nav-tab"
          :class="{ active: $route.path === '/ip-statistics' }"
        >
          IP统计
        </router-link>
        <router-link
          to="/business-statistics"
          class="nav-tab"
          :class="{ active: $route.path === '/business-statistics' }"
        >
          业务统计
        </router-link>
        <router-link
          to="/zero-trust"
          class="nav-tab"
          :class="{ active: $route.path === '/zero-trust' }"
        >
          零信任
        </router-link>
        <router-link
          to="/mapping-ledger"
          class="nav-tab"
          :class="{ active: $route.path === '/mapping-ledger' }"
        >
          映射
        </router-link>
        <router-link
          to="/audit-logs"
          class="nav-tab"
          :class="{ active: $route.path === '/audit-logs' }"
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
              <el-dropdown-item command="publicIP">
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

    <div class="main-content">
      <div class="toolbar">
        <el-input
          v-model="filters.keyword"
          placeholder="搜索操作人/动作/资源/request_id"
          clearable
          style="width: 240px"
          @keyup.enter="loadList(1)"
          @clear="loadList(1)"
        />
        <el-select
          v-model="filters.action"
          placeholder="动作"
          clearable
          style="width: 140px"
          @change="loadList(1)"
        >
          <el-option
            label="create"
            value="create"
          />
          <el-option
            label="update"
            value="update"
          />
          <el-option
            label="delete"
            value="delete"
          />
          <el-option
            label="batch_update"
            value="batch_update"
          />
          <el-option
            label="import_csv"
            value="import_csv"
          />
          <el-option
            label="import_xlsx"
            value="import_xlsx"
          />
        </el-select>
        <el-select
          v-model="filters.resource_type"
          placeholder="资源类型"
          clearable
          style="width: 160px"
          @change="loadList(1)"
        >
          <el-option
            label="主机"
            value="host"
          />
          <el-option
            label="人员"
            value="person"
          />
          <el-option
            label="公网IP"
            value="public_ip"
          />
          <el-option
            label="零信任"
            value="zero_trust"
          />
          <el-option
            label="端口映射"
            value="port_mapping"
          />
          <el-option
            label="云资源"
            value="cloud_resource"
          />
          <el-option
            label="IP网段"
            value="ip_subnet"
          />
          <el-option
            label="数据交换"
            value="data_exchange"
          />
        </el-select>
        <el-button
          type="primary"
          @click="loadList(1)"
        >
          查询
        </el-button>
      </div>

      <el-table
        v-loading="loading"
        :data="items"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column
          prop="id"
          label="ID"
          width="80"
        />
        <el-table-column
          prop="created_at"
          label="时间"
          width="170"
        />
        <el-table-column
          prop="operator_name"
          label="操作人"
          width="120"
        />
        <el-table-column
          prop="action"
          label="动作"
          width="120"
        />
        <el-table-column
          prop="resource_type"
          label="资源类型"
          width="120"
        />
        <el-table-column
          prop="resource_id"
          label="资源ID"
          width="90"
        />
        <el-table-column
          prop="request_id"
          label="请求ID"
          width="140"
          show-overflow-tooltip
        />
        <el-table-column
          label="详情"
          min-width="200"
        >
          <template #default="{ row }">
            <el-button
              v-if="row.detail"
              text
              type="primary"
              size="small"
              @click="showDetail(row)"
            >
              查看
            </el-button>
            <span v-else>—</span>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="pagination"
        background
        layout="total, prev, pager, next, sizes"
        :total="total"
        :page-size="pageSize"
        :current-page="page"
        :page-sizes="[20, 50, 100]"
        @current-change="handlePageChange"
        @size-change="handleSizeChange"
      />
    </div>

    <el-dialog
      v-model="detailVisible"
      title="审计详情"
      width="640px"
    >
      <pre class="detail-pre">{{ detailText }}</pre>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { Setting } from '@element-plus/icons-vue'
import { getAuditLogs } from '../api/audit'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const items = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = ref({
  keyword: '',
  action: '',
  resource_type: ''
})
const detailVisible = ref(false)
const detailText = ref('')

const loadList = async (p = page.value) => {
  loading.value = true
  try {
    const resp = await getAuditLogs({
      page: p,
      page_size: pageSize.value,
      keyword: filters.value.keyword || undefined,
      action: filters.value.action || undefined,
      resource_type: filters.value.resource_type || undefined
    })
    items.value = resp.items || []
    total.value = resp.total || 0
    page.value = resp.page || p
  } finally {
    loading.value = false
  }
}

const handlePageChange = (p) => loadList(p)
const handleSizeChange = (s) => {
  pageSize.value = s
  loadList(1)
}

const showDetail = (row) => {
  try {
    detailText.value = JSON.stringify(JSON.parse(row.detail), null, 2)
  } catch {
    detailText.value = row.detail
  }
  detailVisible.value = true
}

const handleCommand = (command) => {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
  } else if (command === 'cloudResource') {
    router.push('/statistics')
  } else if (command === 'personnel') {
    router.push('/personnel')
  } else if (command === 'publicIP') {
    router.push('/public-ip')
  } else if (command === 'dataBackup') {
    router.push('/data-backup')
  }
}

onMounted(() => loadList(1))
</script>

<style scoped>
.audit-log {
  min-height: 100vh;
  background: #f5f7fa;
}

.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  padding: 0 24px;
  height: 56px;
}

.header-left {
  display: flex;
  align-items: baseline;
  gap: 12px;
}

.header-left h1 {
  margin: 0;
  font-size: 18px;
  color: #303133;
}

.header-left span {
  font-size: 12px;
  color: #909399;
}

.header-center {
  display: flex;
  gap: 4px;
}

.nav-tab {
  padding: 8px 14px;
  color: #606266;
  text-decoration: none;
  border-radius: 4px;
  font-size: 14px;
}

.nav-tab.active {
  color: #409eff;
  background: #ecf5ff;
  font-weight: 600;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-info {
  color: #606266;
  font-size: 13px;
}

.main-content {
  padding: 20px 24px;
}

.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.pagination {
  margin-top: 16px;
  justify-content: flex-end;
}

.detail-pre {
  margin: 0;
  max-height: 420px;
  overflow: auto;
  background: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  font-size: 12px;
  line-height: 1.6;
}
</style>
