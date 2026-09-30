<template>
  <div class="personnel-management">
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
              <el-dropdown-item
                command="personnel"
                :disabled="$route.path === '/personnel'"
              >
                人员录入
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
          v-model="keyword"
          placeholder="搜索姓名 / 联系方式 / 单位名称"
          clearable
          style="width: 260px"
          @keyup.enter="loadPersons"
          @clear="loadPersons"
        />
        <el-button
          type="primary"
          @click="loadPersons"
        >
          查询
        </el-button>
        <el-button
          type="primary"
          @click="openDialog()"
        >
          新增人员
        </el-button>
      </div>

      <el-table
        v-loading="loading"
        :data="persons"
        border
        stripe
      >
        <el-table-column
          label="#"
          width="60"
        >
          <template #default="{ $index }">
            {{ $index + 1 }}
          </template>
        </el-table-column>
        <el-table-column
          prop="name"
          label="姓名"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column
          prop="contact"
          label="联系方式"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.contact || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="unit"
          label="单位名称"
          min-width="180"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.unit || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          label="关联主机"
          width="90"
          align="center"
        >
          <template #default="{ row }">
            {{ row.host_count || 0 }}
          </template>
        </el-table-column>
        <el-table-column
          label="创建时间"
          width="170"
        >
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column
          label="操作"
          width="140"
          align="center"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              type="primary"
              link
              size="small"
              @click="openDialog(row)"
            >
              编辑
            </el-button>
            <el-popconfirm
              :title="deleteHint(row)"
              @confirm="handleDelete(row.id)"
            >
              <template #reference>
                <el-button
                  type="danger"
                  link
                  size="small"
                >
                  删除
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="480px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="90px"
      >
        <el-form-item
          label="姓名"
          prop="name"
        >
          <el-input
            v-model="form.name"
            placeholder="请输入姓名"
            maxlength="64"
          />
        </el-form-item>
        <el-form-item
          label="联系方式"
          prop="contact"
        >
          <el-input
            v-model="form.contact"
            placeholder="请输入联系方式"
            maxlength="64"
          />
        </el-form-item>
        <el-form-item
          label="单位名称"
          prop="unit"
        >
          <el-input
            v-model="form.unit"
            placeholder="请输入单位名称"
            maxlength="128"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">
          取消
        </el-button>
        <el-button
          type="primary"
          :loading="submitting"
          @click="handleSubmit"
        >
          {{ submitting ? '提交中...' : '确定' }}
        </el-button>
      </template>
    </el-dialog>

    <ChangePasswordDialog ref="changePasswordDialog" />
    <CloudResourceDialog ref="cloudResourceDialog" />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import { getPersons, createPerson, updatePerson, deletePerson } from '../api/person'
import { formatTime } from '../utils'
import ChangePasswordDialog from '../components/ChangePasswordDialog.vue'
import CloudResourceDialog from '../components/CloudResourceDialog.vue'

const router = useRouter()
const authStore = useAuthStore()

const changePasswordDialog = ref(null)
const cloudResourceDialog = ref(null)

const persons = ref([])
const loading = ref(false)
const keyword = ref('')
const dialogVisible = ref(false)
const submitting = ref(false)
const editId = ref(null)
const formRef = ref(null)

const form = reactive({ name: '', contact: '', unit: '' })

const rules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }]
}

const dialogTitle = computed(() => (editId.value ? '编辑人员' : '新增人员'))

onMounted(() => {
  authStore.fetchUserInfo()
  loadPersons()
})

async function loadPersons() {
  loading.value = true
  try {
    const res = await getPersons(keyword.value)
    if (res.code === 0) {
      persons.value = res.data || []
    }
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  } finally {
    loading.value = false
  }
}

function openDialog(row) {
  editId.value = row?.id || null
  form.name = row?.name || ''
  form.contact = row?.contact || ''
  form.unit = row?.unit || ''
  dialogVisible.value = true
  formRef.value?.clearValidate()
}

async function handleSubmit() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const payload = { name: form.name.trim(), contact: form.contact.trim(), unit: form.unit.trim() }
    const res = editId.value ? await updatePerson(editId.value, payload) : await createPerson(payload)
    if (res.code === 0) {
      ElMessage.success(editId.value ? '修改成功' : '新增成功')
      dialogVisible.value = false
      await loadPersons()
    }
  } finally {
    submitting.value = false
  }
}

function deleteHint(row) {
  return row.host_count > 0
    ? `该人员关联 ${row.host_count} 台主机，删除将被拒绝，仍要尝试删除？`
    : `确定删除人员 ${row.name}？`
}

async function handleDelete(id) {
  try {
    const res = await deletePerson(id)
    if (res.code === 0) {
      ElMessage.success('删除成功')
      await loadPersons()
    }
  } catch {
    // 错误提示由 axios 拦截器统一弹出
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
</script>

<style scoped>
.personnel-management {
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

.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}
</style>
