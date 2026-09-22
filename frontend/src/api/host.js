import api from './index'

export function getHosts(params) {
  return api.get('/hosts', { params })
}

export function getHost(id) {
  return api.get(`/hosts/${id}`)
}

export function createHost(data) {
  return api.post('/hosts', data)
}

export function updateHost(id, data) {
  return api.put(`/hosts/${id}`, data)
}

export function deleteHost(id) {
  return api.delete(`/hosts/${id}`)
}

export function batchCreateHosts(hosts) {
  return api.post('/batch/hosts', { hosts })
}

export function batchUpdateHosts(ids, data) {
  return api.put('/batch/hosts', { ids, data })
}
