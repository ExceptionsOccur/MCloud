<template>
  <el-dialog
    v-model="visible"
    title="修改密码"
    width="400px"
    :close-on-click-modal="false"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="80px"
    >
      <el-form-item
        label="原密码"
        prop="old_password"
      >
        <el-input
          v-model="form.old_password"
          type="password"
          show-password
          placeholder="请输入原密码"
        />
      </el-form-item>
      <el-form-item
        label="新密码"
        prop="new_password"
      >
        <el-input
          v-model="form.new_password"
          type="password"
          show-password
          placeholder="请输入新密码"
        />
      </el-form-item>
      <el-form-item
        label="确认密码"
        prop="confirm_password"
      >
        <el-input
          v-model="form.confirm_password"
          type="password"
          show-password
          placeholder="请再次输入新密码"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">
        取消
      </el-button>
      <el-button
        type="primary"
        :loading="submitting"
        @click="handleSubmit"
      >
        确定
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { changePassword } from '../api/auth'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

const authStore = useAuthStore()
const router = useRouter()
const visible = ref(false)
const submitting = ref(false)
const formRef = ref(null)

const form = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const rules = {
  old_password: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== form.new_password) {
          callback(new Error('两次输入密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

function open() {
  form.old_password = ''
  form.new_password = ''
  form.confirm_password = ''
  visible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const res = await changePassword(form.old_password, form.new_password)
    if (res.code === 0) {
      ElMessage.success('密码修改成功，请重新登录')
      authStore.logout()
      router.push('/login')
      visible.value = false
    }
  } finally {
    submitting.value = false
  }
}

defineExpose({ open })
</script>
