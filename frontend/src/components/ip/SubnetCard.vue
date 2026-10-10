<template>
  <div class="subnet-card">
    <div class="subnet-header">
      <h3>{{ subnet.subnet }}</h3>
      <el-button
        size="small"
        type="primary"
        :loading="isBatchRunning"
        :disabled="isBatchRunning"
        @click="emit('batch-test', subnet)"
      >
        全量测试
      </el-button>
    </div>
    <div class="subnet-stats">
      <span class="stat"><em class="used-dot"></em>已用 {{ counts.red ?? 0 }}</span>
      <span class="stat"><em class="empty-dot"></em>空记录 {{ counts.yellow ?? 0 }}</span>
      <span class="stat"><em class="unused-dot"></em>未用 {{ counts.green ?? 0 }}</span>
    </div>
    <div
      v-if="isBatchRunning"
      class="batch-progress"
    >
      <el-progress
        :percentage="batchPercent"
        :stroke-width="8"
        :show-text="false"
        striped
        striped-flow
      />
      <span class="batch-text">{{ batchDone }} / 256</span>
    </div>
    <div class="bitmap-wrapper">
      <div class="col-labels">
        <span
          v-for="c in 10"
          :key="c"
        >{{ c - 1 }}</span>
      </div>
      <div class="bitmap-body">
        <div
          v-for="row in 26"
          :key="'r'+row"
          class="bitmap-row"
        >
          <span class="row-label">{{ row - 1 }}</span>
          <div
            v-for="col in 10"
            :key="'c'+col"
            class="bit-cell"
            :class="getCellClasses((row - 1) * 10 + (col - 1))"
            :title="getIP(subnet.subnet, (row - 1) * 10 + (col - 1))"
            @click="emit('cell-click', subnet, (row - 1) * 10 + (col - 1))"
          ></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  subnet: { type: Object, required: true },
  cellColors: { type: Object, required: true },
  probing: { type: Object, required: true },
  batchState: { type: Object, required: true },
  batchTested: { type: Object, required: true },
  counts: { type: Object, default: () => ({}) }
})
const emit = defineEmits(['cell-click', 'batch-test'])

const isBatchRunning = computed(() => !!props.batchState[props.subnet.subnet]?.running)
const batchDone = computed(() => props.batchState[props.subnet.subnet]?.done ?? 0)
const batchPercent = computed(() => {
  const state = props.batchState[props.subnet.subnet]
  if (!state) return 0
  return Math.min(100, Math.round((state.done / state.total) * 100))
})

function getIP(subnet, offset) {
  const parts = subnet.split('.')
  const third = parseInt(parts[2])
  const fourth = offset
  return `${parts[0]}.${parts[1]}.${third}.${fourth}`
}

function getCellClasses(offset) {
  if (offset > 255) return { empty: true }
  const ip = getIP(props.subnet.subnet, offset)
  const color = props.cellColors[ip]
  return {
    used: color === 'red',
    unused: color === 'green',
    alive: color === 'yellow',
    probing: !!props.probing[ip],
    tested: !!props.batchTested[ip]
  }
}
</script>

<style scoped>
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
</style>
