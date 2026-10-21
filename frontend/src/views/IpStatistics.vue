<template>
  <div class="ip-statistics">
    <AppHeader />
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

    <SubnetManageDialog
      ref="subnetDialog"
      @changed="onSubnetChanged"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useStatsStore } from '../stores/stats'
import { useProbeWebSocket } from '../composables/useProbeWebSocket'
import { useProbeGrid } from '../composables/useProbeGrid'
import ProbePanel from '../components/ip/ProbePanel.vue'
import SubnetCard from '../components/ip/SubnetCard.vue'
import SubnetManageDialog from '../components/ip/SubnetManageDialog.vue'
import AppHeader from '../components/AppHeader.vue'

const statsStore = useStatsStore()
const subnetDialog = ref(null)

let grid = null
const { wsConnected, connect: connectWS, send: sendProbe, isUnmounted } = useProbeWebSocket(data => grid?.applyProbeResult(data))
grid = useProbeGrid({ getSend: () => sendProbe, isUnmounted, getSubnets: () => statsStore.ipUsage })
const { probing, cellColors, batchState, batchTested, subnetCounts, initCellColors, handleBatchTest, handleCellClick } = grid

onMounted(async () => {
  connectWS()
  try {
    await statsStore.fetchIPUsage()
    initCellColors()
  } catch {
    // 错误提示由 axios 拦截器统一弹出
  }
})

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
