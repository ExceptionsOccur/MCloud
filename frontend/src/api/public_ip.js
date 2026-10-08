import api from './index'

export function getPublicIPs(keyword = '') {
  return api.get('/public-ips', { params: { keyword } })
}

export function createPublicIP(data) {
  return api.post('/public-ips', data)
}

export function updatePublicIP(id, data) {
  return api.put(`/public-ips/${id}`, data)
}

export function deletePublicIP(id) {
  return api.delete(`/public-ips/${id}`)
}
