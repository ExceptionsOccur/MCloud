import api from './index'

export function importCSV(file) {
  const formData = new FormData()
  formData.append('file', file)
  return api.post('/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function exportCSV(params) {
  return api.get('/export', { params, responseType: 'blob' })
}

export function downloadTemplate() {
  return api.get('/template', { responseType: 'blob' })
}
