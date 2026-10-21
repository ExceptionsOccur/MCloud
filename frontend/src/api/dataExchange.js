import api from './index'

// 导出 8 张业务表为普通 xlsx 字节（GET /api/export/all）；加密在前端完成（T-074）
export function exportAll() {
  return api.get('/export/all', { responseType: 'blob' })
}

// 导入备份并按自然键 upsert（POST /api/import/all，multipart 字段 file；加密文件由前端先解密）
export function importAll(file) {
  const formData = new FormData()
  formData.append('file', file)
  return api.post('/import/all', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 120000
  })
}
