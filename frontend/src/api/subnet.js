import api from './index'

export function getSubnets() {
  return api.get('/ip-subnets')
}

export function createSubnet(cidr) {
  return api.post('/ip-subnets', { cidr })
}

export function updateSubnet(id, cidr) {
  return api.put(`/ip-subnets/${id}`, { cidr })
}

export function deleteSubnet(id) {
  return api.delete(`/ip-subnets/${id}`)
}
