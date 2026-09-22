import api from './index'

export function getIPUsage() {
  return api.get('/stats/ip-usage')
}

export function probeIP(ip, color) {
  return api.post('/stats/probe', { ip, color })
}

export function getBusinessStats() {
  return api.get('/stats/business')
}
