export function downloadBlob(blob, filename) {
  const url = window.URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  window.URL.revokeObjectURL(url)
  document.body.removeChild(a)
}

export function formatTime(time) {
  if (!time) return '-'
  const d = new Date(time)
  return d.toLocaleString('zh-CN')
}

export const statusOptions = ['运行中', '已停止', '已关机', '待确认']
export const envTypeOptions = ['测试', '生产']
export const assetTypeOptions = ['虚拟机', '裸金属服务器']
export const cpuArchOptions = [
  { label: 'C86', value: 'C86' },
  { label: 'X86', value: 'X86' },
  { label: 'ARM', value: 'ARM' }
]
