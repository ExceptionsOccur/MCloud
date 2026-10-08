import api from './index'

export function getZeroTrusts(keyword = '') {
  return api.get('/zero-trusts', { params: { keyword } })
}

export function createZeroTrust(data) {
  return api.post('/zero-trusts', data)
}

export function updateZeroTrust(id, data) {
  return api.put(`/zero-trusts/${id}`, data)
}

export function deleteZeroTrust(id) {
  return api.delete(`/zero-trusts/${id}`)
}

export function batchCreateZeroTrustsText(text) {
  return api.post('/zero-trusts/batch', { text })
}
