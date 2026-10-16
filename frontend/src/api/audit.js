import api from './index'

// 查询某资产的变更记录（资产详情「变更记录」时间线，T-066）
export function getResourceHistory(resourceType, resourceId) {
  return api.get('/audit-logs/resource', {
    params: { resource_type: resourceType, resource_id: resourceId }
  })
}
