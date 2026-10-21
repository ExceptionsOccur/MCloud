import api from './index'

export function getHosts(params) {
  return api.get('/hosts', { params })
}

// 全库主机计数（不受筛选影响，T-086）
export function getHostStats() {
  return api.get('/hosts/stats')
}

export function getRegions() {
  return api.get('/hosts/regions')
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

export function batchCreateHostsText(text) {
  return api.post('/batch/hosts/text', { text })
}

export function batchUpdateHosts(ids, data) {
  return api.put('/batch/hosts', { ids, data })
}
