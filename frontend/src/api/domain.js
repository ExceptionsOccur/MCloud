import api from './index'

export function getDomains(keyword = '') {
  return api.get('/domains', { params: { keyword } })
}

export function createDomain(data) {
  return api.post('/domains', data)
}

export function updateDomain(id, data) {
  return api.put(`/domains/${id}`, data)
}

export function deleteDomain(id) {
  return api.delete(`/domains/${id}`)
}
