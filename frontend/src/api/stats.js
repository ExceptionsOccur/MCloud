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

// 各区域资源用量聚合（T-087，后端单条 SQL）
export function getRegionUsage() {
  return api.get('/stats/region-usage')
}
