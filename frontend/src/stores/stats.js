import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getIPUsage } from '../api/stats'

export const useStatsStore = defineStore('stats', () => {
  const ipUsage = ref([])
  const loading = ref(false)

  async function fetchIPUsage() {
    loading.value = true
    try {
      const res = await getIPUsage()
      if (res.code === 0) {
        ipUsage.value = res.data || []
      }
    } finally {
      loading.value = false
    }
  }

  return { ipUsage, loading, fetchIPUsage }
})
