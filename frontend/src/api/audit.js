import api from './index'

export function getAuditLogs(params = {}) {
  return api.get('/audit-logs', { params })
}
