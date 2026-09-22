import api from './index'

export function importCSV(file) {
  const formData = new FormData()
  formData.append('file', file)
  return api.post('/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function exportCSV() {
  return api.get('/export', { responseType: 'blob' })
}

export function downloadTemplate() {
  return api.get('/template', { responseType: 'blob' })
}
