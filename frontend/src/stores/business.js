import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getBusinessStats } from '../api/stats'

export const useBusinessStore = defineStore('business', () => {
  const summary = ref({ host_count: 0, project_count: 0, company_count: 0, applicant_count: 0 })
  const byProject = ref([])
  const byCompany = ref([])
  const byApplicant = ref([])
  const loading = ref(false)

  async function fetchBusinessStats() {
    loading.value = true
    try {
      const res = await getBusinessStats()
      if (res.code === 0) {
        summary.value = res.data.summary || summary.value
        byProject.value = res.data.by_project || []
        byCompany.value = res.data.by_company || []
        byApplicant.value = res.data.by_applicant || []
      }
    } finally {
      loading.value = false
    }
  }

  return { summary, byProject, byCompany, byApplicant, loading, fetchBusinessStats }
})
