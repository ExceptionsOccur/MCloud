import api from './index'

// 导出 8 张业务表为单个加密备份包（GET /api/export/all?password=xxx，T-072）
export function exportAll(password) {
  return api.get('/export/all', { params: { password }, responseType: 'blob' })
}

// 导入备份并按自然键 upsert（POST /api/import/all，multipart 字段 file + 可选 password）
export function importAll(file, password) {
  const formData = new FormData()
  formData.append('file', file)
  if (password) formData.append('password', password)
  return api.post('/import/all', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 120000
  })
}
