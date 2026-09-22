import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getCloudResources } from '../api/cloud_resource'
import { getHosts } from '../api/host'

const EMPTY_STAT = { cpu: 0, memory: 0, storage: 0, bare_metal: 0 }

function sumField(hosts, field) {
  return hosts.reduce((sum, h) => sum + (h[field] || 0), 0)
}

function sumStorage(hosts) {
  return hosts.reduce((sum, h) => sum + (h.system_disk || 0) + (h.data_disk || 0), 0)
}

export const useCloudResourceStore = defineStore('cloudResource', () => {
  const resources = ref([])
  const loading = ref(false)
  const regionHostStats = ref({})

  const regionList = computed(() => {
    return resources.value.map(r => r.region)
  })

  function getRegionResource(region) {
    return resources.value.find(r => r.region === region) || null
  }

  function getRegionRunning(region) {
    return regionHostStats.value[region]?.running || { ...EMPTY_STAT }
  }

  function getRegionStopped(region) {
    return regionHostStats.value[region]?.stopped || { ...EMPTY_STAT }
  }

  function getRegionUsed(region) {
    const running = getRegionRunning(region)
    const stopped = getRegionStopped(region)
    return {
      cpu: running.cpu + stopped.cpu,
      memory: running.memory + stopped.memory,
      storage: running.storage + stopped.storage,
      bare_metal: running.bare_metal + stopped.bare_metal
    }
  }

  function getRegionUnused(region) {
    const res = getRegionResource(region)
    if (!res) return { vcpu: 0, memory: 0, storage: 0 }
    const used = getRegionUsed(region)
    return {
      vcpu: Math.max(0, res.vcpu - used.cpu),
      memory: Math.max(0, res.memory - used.memory),
      storage: Math.max(0, res.storage - used.storage)
    }
  }

  async function fetchStatistics() {
    loading.value = true
    try {
      const res = await getCloudResources()
      if (res.code === 0) {
        resources.value = res.data || []
      }

      const regions = ['region-a', 'region-b']
      const stats = {}
      for (const region of regions) {
        try {
          let allHosts = []
          let page = 1
          const pageSize = 100
          while (true) {
            const hostRes = await getHosts({ region, page, page_size: pageSize })
            if (hostRes.code !== 0) break
            const hosts = hostRes.data.hosts || []
            allHosts = allHosts.concat(hosts)
            if (allHosts.length >= hostRes.data.total || hosts.length < pageSize) break
            page++
          }

          const runningHosts = allHosts.filter(h => h.status === '运行中')
          const stoppedHosts = allHosts.filter(h => h.status === '已停止' || h.status === '已关机')
          const runningVm = runningHosts.filter(h => h.asset_type !== '裸金属服务器')
          const stoppedVm = stoppedHosts.filter(h => h.asset_type !== '裸金属服务器')

          stats[region] = {
            running: {
              cpu: sumField(runningVm, 'cpu'),
              memory: sumField(runningVm, 'memory'),
              storage: sumStorage(runningVm),
              bare_metal: runningHosts.filter(h => h.asset_type === '裸金属服务器').length
            },
            stopped: {
              cpu: sumField(stoppedVm, 'cpu'),
              memory: sumField(stoppedVm, 'memory'),
              storage: sumStorage(stoppedVm),
              bare_metal: stoppedHosts.filter(h => h.asset_type === '裸金属服务器').length
            }
          }
        } catch {
          stats[region] = { running: { ...EMPTY_STAT }, stopped: { ...EMPTY_STAT } }
        }
      }
      regionHostStats.value = stats
    } finally {
      loading.value = false
    }
  }

  return {
    resources, loading, regionHostStats, regionList,
    getRegionResource, getRegionRunning, getRegionStopped, getRegionUsed, getRegionUnused,
    fetchStatistics
  }
})
