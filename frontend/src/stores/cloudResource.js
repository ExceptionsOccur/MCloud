import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getCloudResources } from '../api/cloud_resource'
import { getRegions } from '../api/host'
import { getRegionUsage } from '../api/stats'


// T-087：区域用量由后端 GET /api/stats/region-usage 一次聚合返回，
// 前端不再按区域分页拉取全量主机明细（原实现为「区域数 × 页数」次请求）。
function emptyStats() {
  return {
    running: { cpu: 0, memory: 0, storage: 0, bare_metal: 0 },
    stopped: { cpu: 0, memory: 0, storage: 0, bare_metal: 0 }
  }
}

function toRegionStats(u) {
  return {
    running: {
      cpu: u?.running_cpu || 0,
      memory: u?.running_memory || 0,
      storage: u?.running_storage || 0,
      bare_metal: u?.running_bare_metal || 0
    },
    stopped: {
      cpu: u?.stopped_cpu || 0,
      memory: u?.stopped_memory || 0,
      storage: u?.stopped_storage || 0,
      bare_metal: u?.stopped_bare_metal || 0
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
    return regionHostStats.value[region]?.running || { cpu: 0, memory: 0, storage: 0, bare_metal: 0 }
  }

  function getRegionStopped(region) {
    return regionHostStats.value[region]?.stopped || { cpu: 0, memory: 0, storage: 0, bare_metal: 0 }
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
      const [cloudRes, regionsRes, usageRes] = await Promise.all([
        getCloudResources(),
        getRegions(),
        getRegionUsage()
      ])

      if (cloudRes.code === 0) {
        resources.value = cloudRes.data || []
      }

      const dbRegions = regionsRes.code === 0 ? (regionsRes.data || []) : []
      hostRegions.value = dbRegions

      const usage = usageRes.code === 0 ? (usageRes.data || {}) : {}
      const stats = {}
      for (const region of dbRegions) {
        stats[region] = usage[region] ? toRegionStats(usage[region]) : emptyStats()
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
