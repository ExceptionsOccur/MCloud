import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getHosts, getHostStats, createHost, updateHost, deleteHost } from '../api/host'
import { usePagedTable } from '../composables/usePagedTable'

const DEFAULT_FILTERS = {
  keyword: '',
  env_type: '',
  asset_type: '',
  cpu_arch: '',
  is_db_server: '',
  is_external: '0',
  status: '',
  region: '',
  applicant_empty: ''
}

export const useHostStore = defineStore('host', () => {
  const totalAll = ref(0)
  const internalCount = ref(0)
  const externalCount = ref(0)

  const {
    items: hosts, total, loading, currentPage, pageSize, filters,
    fetchPage, setFilter, setPage, resetFilters
  } = usePagedTable({
    defaultFilters: DEFAULT_FILTERS,
    fetcher: async (params) => {
      const res = await getHosts(params)
      if (res.code === 0) {
        return {
          items: res.data.hosts || [],
          total: res.data.total || 0
        }
      }
      return null
    }
  })

  async function fetchHosts() {
    return fetchPage()
  }

  // T-086：页脚全库计数走独立接口，与列表筛选彻底解耦
  async function fetchStats() {
    try {
      const res = await getHostStats()
      if (res.code === 0) {
        totalAll.value = res.data.total ?? 0
        internalCount.value = res.data.internal_count ?? 0
        externalCount.value = res.data.external_count ?? 0
      }
    } catch {
      // 错误提示由 axios 拦截器统一弹出；统计失败不影响列表
    }
  }

  async function create(data) {
    const res = await createHost(data)
    if (res.code === 0) {
      await Promise.all([fetchHosts(), fetchStats()])
      return true
    }
    return false
  }

  async function update(id, data) {
    const res = await updateHost(id, data)
    if (res.code === 0) {
      await Promise.all([fetchHosts(), fetchStats()])
      return true
    }
    return false
  }

  async function remove(id) {
    const res = await deleteHost(id)
    if (res.code === 0) {
      await Promise.all([fetchHosts(), fetchStats()])
      return true
    }
    return false
  }

  return {
    hosts, total, totalAll, internalCount, externalCount, loading, currentPage, pageSize, filters,
    fetchHosts, fetchStats, create, update, remove, setFilter, setPage, resetFilters
  }
})
