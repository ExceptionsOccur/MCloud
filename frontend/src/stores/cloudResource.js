import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getCloudResources } from '../api/cloud_resource'
import { getHosts, getRegions } from '../api/host'

const EMPTY_STAT = { cpu: 0, memory: 0, storage: 0, bare_metal: 0 }

function sumField(hosts, field) {
  return hosts.reduce((sum, h) => sum + (h[field] || 0), 0)
}

function sumStorage(hosts) {
  return hosts.reduce((sum, h) => sum + (h.system_disk || 0) + (h.data_disk || 0), 0)
}

async function fetchAllHostsByRegion(region) {
  const allHosts = []
  let page = 1
  const pageSize = 100
  while (true) {
    const res = await getHosts({ region, page, page_size: pageSize })
    if (res.code !== 0) break
    const hosts = res.data.hosts || []
    allHosts.push(...hosts)
    if (allHosts.length >= res.data.total || hosts.length < pageSize) break
    page++
  }
  return allHosts
}

function computeRegionStats(hosts) {
  const runningHosts = hosts.filter(h => h.status === '运行中')
  const stoppedHosts = hosts.filter(h => h.status === '已停止' || h.status === '已关机')
  const runningVm = runningHosts.filter(h => h.asset_type !== '裸金属服务器')
  const stoppedVm = stoppedHosts.filter(h => h.asset_type !== '裸金属服务器')
  return {
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
}

export const useCloudResourceStore = defineStore('cloudResource', () => {
  const resources = ref([])
  const loading = ref(false)
  const regionHostStats = ref({})
  const hostRegions = ref([])

  const regionList = computed(() => {
    return resources.value.map(r => r.region).filter(Boolean)
  })

  const allRegions = computed(() => {
    const set = new Set([...regionList.value, ...hostRegions.value].filter(Boolean))
    return [...set].sort()
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
      const [cloudRes, regionsRes] = await Promise.all([
        getCloudResources(),
        getRegions()
      ])

      if (cloudRes.code === 0) {
        resources.value = cloudRes.data || []
      }

      const dbRegions = regionsRes.code === 0 ? (regionsRes.data || []) : []
      hostRegions.value = dbRegions

      const stats = {}
      for (const region of dbRegions) {
        try {
          const hosts = await fetchAllHostsByRegion(region)
          stats[region] = computeRegionStats(hosts)
        } catch {
          stats[region] = { running: { ...EMPTY_STAT }, stopped: { ...EMPTY_STAT } }
        }
      }
      regionHostStats.value = stats
    } finally {
      loading.value = false
    }
  }

  async function fetchRegions() {
    const [cloudRes, regionsRes] = await Promise.all([
      getCloudResources(),
      getRegions()
    ])
    if (cloudRes.code === 0) {
      resources.value = cloudRes.data || []
    }
    if (regionsRes.code === 0) {
      hostRegions.value = (regionsRes.data || []).filter(Boolean)
    }
  }

  return {
    resources, loading, regionHostStats, regionList, allRegions,
    getRegionResource, getRegionRunning, getRegionStopped, getRegionUsed, getRegionUnused,
    fetchStatistics, fetchRegions
  }
})
