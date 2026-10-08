import api from './index'

export function getPortMappings(keyword = '') {
  return api.get('/port-mappings', { params: { keyword } })
}

export function createPortMapping(data) {
  return api.post('/port-mappings', data)
}

export function updatePortMapping(id, data) {
  return api.put(`/port-mappings/${id}`, data)
}

export function deletePortMapping(id) {
  return api.delete(`/port-mappings/${id}`)
}

export function batchCreatePortMappingsText(text) {
  return api.post('/port-mappings/batch', { text })
}
