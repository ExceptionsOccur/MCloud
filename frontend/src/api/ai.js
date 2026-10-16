import api from './index'

export function getAISettings() {
  return api.get('/ai/settings')
}

export function saveAISettings(data) {
  return api.put('/ai/settings', data)
}

export function extractApplication(file, onUploadProgress) {
  const formData = new FormData()
  formData.append('file', file)
  return api.post('/ai/extract', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 120000,
    onUploadProgress
  })
}

export function archiveApplication(fileToken) {
  return api.post('/ai/archive', { file_token: fileToken })
}
