import api from './index'

export function getObjectStorages(keyword = '') {
  return api.get('/object-storages', { params: { keyword } })
}

export function createObjectStorage(data) {
  return api.post('/object-storages', data)
}

export function updateObjectStorage(id, data) {
  return api.put(`/object-storages/${id}`, data)
}

export function deleteObjectStorage(id) {
  return api.delete(`/object-storages/${id}`)
}
