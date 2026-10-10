import { ref } from 'vue'

/**
 * 分页列表通用态：页码/页大小/筛选/加载/总数 + 拉取与筛选重置
 * @param {object} options
 * @param {object} options.defaultFilters 筛选默认值（resetFilters 恢复目标，按键与调用方约定一致）
 * @param {(params: object) => Promise<{items: Array, total: number}|null>} options.fetcher
 *   按组装后的 params 拉取；返回 null 表示本次不覆盖现有数据（对应业务失败）
 */
export function usePagedTable({ defaultFilters = {}, fetcher }) {
  const items = ref([])
  const total = ref(0)
  const loading = ref(false)
  const currentPage = ref(1)
  const pageSize = ref(20)
  const filters = ref({ ...defaultFilters })

  async function fetchPage() {
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

      const data = await fetcher(params)
      if (data) {
        items.value = data.items
        total.value = data.total
      }
    } finally {
      loading.value = false
    }
  }

  function setFilter(key, value) {
    filters.value[key] = value
    currentPage.value = 1
    fetchPage()
  }

  function setPage(page) {
    currentPage.value = page
    fetchPage()
  }

  function resetFilters() {
    filters.value = { ...defaultFilters }
    currentPage.value = 1
    fetchPage()
  }

  return {
    items, total, loading, currentPage, pageSize, filters,
    fetchPage, setFilter, setPage, resetFilters
  }
}
