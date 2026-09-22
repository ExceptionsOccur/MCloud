import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getHosts, createHost, updateHost, deleteHost } from '../api/host'

export const useHostStore = defineStore('host', () => {
  const hosts = ref([])
  const total = ref(0)
  const loading = ref(false)
  const currentPage = ref(1)
  const pageSize = ref(20)
  const filters = ref({
    keyword: '',
    env_type: '',
    asset_type: '',
    cpu_arch: '',
    is_db_server: '',
    status: '',
    region: '',
    applicant_empty: ''
  })

  async function fetchHosts() {
    loading.value = true
    try {
      const params = {
        page: currentPage.value,
        page_size: pageSize.value
      }
      Object.keys(filters.value).forEach(key => {
        if (filters.value[key] !== '' && filters.value[key] !== null && filters.value[key] !== undefined) {
          params[key] = filters.value[key]
        }
      })

      const res = await getHosts(params)

      if (res.code === 0) {
        hosts.value = res.data.hosts || []
        total.value = res.data.total || 0
      }
    } finally {
      loading.value = false
    }
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

  function setFilter(key, value) {
    filters.value[key] = value
    currentPage.value = 1
    fetchHosts()
  }

  function setPage(page) {
    currentPage.value = page
    fetchHosts()
  }

  function resetFilters() {
    filters.value = {
      keyword: '',
      env_type: '',
      asset_type: '',
      cpu_arch: '',
      is_db_server: '',
      status: '',
      region: '',
      applicant_empty: ''
    }
    currentPage.value = 1
    fetchHosts()
  }

  return {
    hosts, total, loading, currentPage, pageSize, filters,
    fetchHosts, create, update, remove, setFilter, setPage, resetFilters
  }
})
