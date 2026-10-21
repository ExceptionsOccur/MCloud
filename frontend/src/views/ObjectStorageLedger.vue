<template>
  <div class="object-storage-ledger">
    <el-header class="app-header">
      <div class="header-left">
        <h1>MCloud</h1>
        <span>云平台资产管理</span>
      </div>
      <AppNav />
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
              <el-dropdown-item
                command="publicIP"
                :disabled="$route.path === '/public-ip'"
              >
                公网IP录入
              </el-dropdown-item>
              <el-dropdown-item command="dataBackup">
                数据备份
              </el-dropdown-item>
              <el-dropdown-item
                command="aiSettings"
                :disabled="$route.path === '/ai-settings'"
              >
                AI 接入设置
              </el-dropdown-item>
              <el-dropdown-item
                command="aiImport"
                :disabled="$route.path === '/ai-import'"
              >
                AI 申请入账
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
          placeholder="搜索 存储桶 / 用途 / 备注"
          clearable
          style="width: 260px"
          @keyup.enter="loadList"
          @clear="loadList"
        />
        <el-button
          type="primary"
          @click="loadList"
        >
          查询
        </el-button>
        <el-button
          type="primary"
          @click="openDialog()"
        >
          新增对象存储
        </el-button>
        <span class="summary">共 {{ records.length }} 个存储桶，合计分配 {{ totalSize }} GB</span>
      </div>

      <el-table
        v-loading="loading"
        :data="records"
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
          prop="bucket"
          label="存储桶名称"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column
          label="AccessKey"
          min-width="150"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.access_key || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          label="SecretKey"
          width="130"
        >
          <template #default="{ row }">
            {{ row.secret_key_mask || (row.secret_key_set ? '已设置' : '-') }}
          </template>
        </el-table-column>
        <el-table-column
          label="分配大小"
          width="120"
          align="center"
        >
          <template #default="{ row }">
            {{ row.size_gb ? row.size_gb + ' GB' : '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="purpose"
          label="用途"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column
          label="主机"
          min-width="240"
        >
          <template #default="{ row }">
            <template v-if="row.hosts && row.hosts.length">
              <el-tag
                v-for="pair in row.hosts"
                :key="pair.id"
                size="small"
                class="host-tag"
              >
                {{ pair.host ? pair.host.name + '（' + pair.host.private_ip + '）' : pair.host_id }}
              </el-tag>
            </template>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="remark"
          label="备注"
          min-width="180"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.remark || '-' }}
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
              :title="`确定删除存储桶 ${row.bucket}？`"
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
      width="560px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item
          label="存储桶名称"
          prop="bucket"
        >
          <el-input
            v-model="form.bucket"
            placeholder="请输入存储桶名称（全局唯一）"
            maxlength="128"
          />
        </el-form-item>
        <el-form-item
          label="用途"
          prop="purpose"
        >
          <el-input
            v-model="form.purpose"
            placeholder="请输入用途"
            maxlength="200"
          />
        </el-form-item>
        <el-form-item
          label="分配大小(GB)"
          prop="size_gb"
        >
          <el-input
            v-model="form.size_gb"
            placeholder="请输入分配大小（整数，单位 GB）"
          />
        </el-form-item>
        <el-form-item
          label="AccessKey"
          prop="access_key"
        >
          <el-input
            v-model="form.access_key"
            placeholder="请输入访问存储桶的用户名（AccessKey）"
            maxlength="128"
          />
        </el-form-item>
        <el-form-item
          label="SecretKey"
          prop="secret_key"
        >
          <el-input
            v-model="form.secret_key"
            type="password"
            show-password
            :placeholder="secretKeyPlaceholder"
            maxlength="256"
          />
        </el-form-item>
        <el-form-item
          label="主机"
          prop="host_ids"
        >
          <el-select
            v-model="form.host_ids"
            multiple
            filterable
            collapse-tags
            collapse-tags-tooltip
            placeholder="可多选：对象存储为外挂存储，可被多台主机共享访问"
            style="width: 100%"
          >
            <el-option
              v-for="h in hostOptions"
              :key="h.id"
              :label="`${h.name}（${h.private_ip}）`"
              :value="h.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          label="备注"
          prop="remark"
        >
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入备注"
            maxlength="500"
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
import { getObjectStorages, createObjectStorage, updateObjectStorage, deleteObjectStorage } from '../api/object_storage'
import { getHosts } from '../api/host'
import { formatTime } from '../utils'
import ChangePasswordDialog from '../components/ChangePasswordDialog.vue'
import CloudResourceDialog from '../components/CloudResourceDialog.vue'
import AppNav from '../components/AppNav.vue'

const router = useRouter()
const authStore = useAuthStore()

const changePasswordDialog = ref(null)
const cloudResourceDialog = ref(null)

const records = ref([])
const hostOptions = ref([])
const loading = ref(false)
const keyword = ref('')
const dialogVisible = ref(false)
const submitting = ref(false)
const editId = ref(null)
const formRef = ref(null)

const form = reactive({ bucket: '', purpose: '', size_gb: '', access_key: '', secret_key: '', host_ids: [], remark: '' })
// 编辑时 SecretKey 只回显末四位；留空表示保留原值
const secretKeyMask = ref('')
const secretKeyPlaceholder = computed(() => (secretKeyMask.value
  ? `已设置（${secretKeyMask.value}），留空则保留原值`
  : '请输入访问存储桶的密钥（SecretKey）'))

const rules = {
  bucket: [{ required: true, message: '请输入存储桶名称', trigger: 'blur' }],
  purpose: [{ required: true, message: '请输入用途', trigger: 'blur' }],
  size_gb: [{ pattern: /^\d*$/, message: '请输入非负整数', trigger: 'blur' }],
  host_ids: [{ type: 'array', required: true, min: 1, message: '请至少选择一台主机', trigger: 'change' }]
}

const dialogTitle = computed(() => (editId.value ? '编辑对象存储' : '新增对象存储'))
const totalSize = computed(() => records.value.reduce((sum, r) => sum + (Number(r.size_gb) || 0), 0))

onMounted(() => {
  authStore.fetchUserInfo()
  loadHostOptions()
  loadList()
})

// 主机候选：分页拉全量（单页上限 100），避免主机较多时下拉缺项
async function loadHostOptions() {
  const all = []
  const size = 100
  for (let page = 1; page <= 50; page++) {
    const res = await getHosts({ page, page_size: size })
    if (res.code !== 0) break
    const batch = res.data?.hosts || []
    all.push(...batch)
    if (batch.length < size || all.length >= (res.data?.total || 0)) break
  }
  hostOptions.value = all
}

async function loadList() {
  loading.value = true
  try {
    const res = await getObjectStorages(keyword.value)
    if (res.code === 0) {
      records.value = res.data || []
    }
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  } finally {
    loading.value = false
  }
}

function openDialog(row) {
  editId.value = row?.id || null
  form.bucket = row?.bucket || ''
  form.purpose = row?.purpose || ''
  form.size_gb = row?.size_gb ?? ''
  form.access_key = row?.access_key || ''
  form.secret_key = ''
  secretKeyMask.value = row?.secret_key_mask || ''
  form.host_ids = row?.hosts ? row.hosts.map((p) => p.host_id).filter(Boolean) : []
  form.remark = row?.remark || ''
  dialogVisible.value = true
  formRef.value?.clearValidate()
}

async function handleSubmit() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const payload = {
      bucket: form.bucket.trim(),
      purpose: form.purpose.trim(),
      size_gb: Number(form.size_gb) || 0,
      access_key: form.access_key.trim(),
      secret_key: form.secret_key.trim(),
      host_ids: form.host_ids,
      remark: form.remark.trim()
    }
    const res = editId.value
      ? await updateObjectStorage(editId.value, payload)
      : await createObjectStorage(payload)
    if (res.code === 0) {
      ElMessage.success(editId.value ? '修改成功' : '新增成功')
      dialogVisible.value = false
      await loadList()
    }
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id) {
  try {
    const res = await deleteObjectStorage(id)
    if (res.code === 0) {
      ElMessage.success('删除成功')
      await loadList()
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
  } else if (command === 'dataBackup') {
    router.push('/data-backup')
  } else if (command === 'cloudResource') {
    cloudResourceDialog.value?.open()
  } else if (command === 'personnel') {
    router.push('/personnel')
  } else if (command === 'publicIP') {
    router.push('/public-ip')
  } else if (command === 'aiSettings') {
    router.push('/ai-settings')
  } else if (command === 'aiImport') {
    router.push('/ai-import')
  }
}
</script>

<style scoped>
.object-storage-ledger {
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
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}

.summary {
  margin-left: auto;
  font-size: 14px;
  color: #909399;
}

.host-tag {
  margin: 2px 6px 2px 0;
}
</style>
