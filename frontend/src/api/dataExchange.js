import api from './index'

// 导出 8 张业务表为单个 xlsx（GET /api/export/all）
export function exportAll() {
  return api.get('/export/all', { responseType: 'blob' })
}

// 导入 xlsx 并按自然键 upsert（POST /api/import/all，multipart 字段 file）
export function importAll(file) {
  const formData = new FormData()
  formData.append('file', file)
  return api.post('/import/all', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 120000
  })
}
