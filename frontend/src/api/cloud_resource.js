import api from './index'

export function getCloudResources() {
  return api.get('/cloud-resources')
}

export function updateCloudResource(data) {
  return api.put('/cloud-resources', data)
}
