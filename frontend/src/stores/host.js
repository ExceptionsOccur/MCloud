import { defineStore } from 'pinia'
import { getHosts, createHost, updateHost, deleteHost } from '../api/host'
import { usePagedTable } from '../composables/usePagedTable'

const DEFAULT_FILTERS = {
  keyword: '',
  env_type: '',
  asset_type: '',
  cpu_arch: '',
  is_db_server: '',
  status: '',
  region: '',
  applicant_empty: ''
}

export const useHostStore = defineStore('host', () => {
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

  async function create(data) {
    const res = await createHost(data)
    if (res.code === 0) {
      await fetchHosts()
      return true
    }
    return false
  }

  async function update(id, data) {
    const res = await updateHost(id, data)
    if (res.code === 0) {
      await fetchHosts()
      return true
    }
    return false
  }

  async function remove(id) {
    const res = await deleteHost(id)
    if (res.code === 0) {
      await fetchHosts()
      return true
    }
    return false
  }

  return {
    hosts, total, loading, currentPage, pageSize, filters,
    fetchHosts, create, update, remove, setFilter, setPage, resetFilters
  }
})
