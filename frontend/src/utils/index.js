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

function pad(n) {
  return String(n).padStart(2, '0')
}

// 展示口径统一 YYYY-MM-DD（浏览器本地日期）
export function formatDate(d = new Date()) {
  if (!(d instanceof Date) || Number.isNaN(d.getTime())) return ''
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

// 展示口径统一 YYYY-MM-DD HH:mm:ss（浏览器本地时区）
export function formatTime(time) {
  if (!time) return '-'
  const d = new Date(time)
  if (Number.isNaN(d.getTime())) return '-'
  return `${formatDate(d)} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// RFC3339（带时区）→ 本地无时区 YYYY-MM-DDTHH:mm:ss（date-picker 回填，与时区安全）
export function toLocalPickerValue(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return `${formatDate(d)}T${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// 本地无时区值（date-picker value-format）→ RFC3339 带时区提交
export function toRFC3339(localNaive) {
  if (!localNaive) return undefined
  const d = new Date(localNaive)
  if (Number.isNaN(d.getTime())) return undefined
  return d.toISOString()
}

export const statusOptions = ['运行中', '已停止', '已关机', '待确认']
export const envTypeOptions = ['测试', '生产']
export const assetTypeOptions = ['虚拟机', '裸金属服务器']
export const cpuArchOptions = [
  { label: 'C86', value: 'C86' },
  { label: 'X86', value: 'X86' },
  { label: 'ARM', value: 'ARM' }
]
