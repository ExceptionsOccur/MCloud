<template>
  <div class="host-management">
    <AppHeader />
    <div class="main-content">
      <SearchToolbar />
      <HostTable />
    </div>

    <HostFormDialog ref="hostFormDialog" />
    <BatchAddDialog ref="batchAddDialog" />
    <BatchEditDialog ref="batchEditDialog" />
    <ImportDialog ref="importDialog" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useHostStore } from '../stores/host'
import { useCloudResourceStore } from '../stores/cloudResource'
import SearchToolbar from '../components/SearchToolbar.vue'
import HostTable from '../components/HostTable.vue'
import HostFormDialog from '../components/HostFormDialog.vue'
import BatchAddDialog from '../components/BatchAddDialog.vue'
import BatchEditDialog from '../components/BatchEditDialog.vue'
import ImportDialog from '../components/ImportDialog.vue'
import AppHeader from '../components/AppHeader.vue'

const hostStore = useHostStore()
const cloudStore = useCloudResourceStore()

const hostFormDialog = ref(null)
const batchAddDialog = ref(null)
const batchEditDialog = ref(null)
const importDialog = ref(null)

onMounted(() => {
  hostStore.fetchHosts()
  hostStore.fetchStats()
  cloudStore.fetchRegions()
})

</script>

<style scoped>
.host-management {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}


.main-content {
  flex: 1;
  padding: 20px;
  overflow: auto;
}
</style>
