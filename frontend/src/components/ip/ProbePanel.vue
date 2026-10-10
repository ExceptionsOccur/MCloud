<template>
  <div class="top-bar">
    <div
      class="ws-status"
      :class="wsConnected ? 'online' : 'offline'"
    >
      <em></em>{{ wsConnected ? '探测服务已连接' : '探测服务连接中...' }}
    </div>
    <el-button
      size="small"
      @click="emit('manage')"
    >
      管理网段
    </el-button>
  </div>
</template>

<script setup>
defineProps({
  wsConnected: { type: Boolean, default: false }
})
const emit = defineEmits(['manage'])
</script>

<style scoped>
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
</style>
