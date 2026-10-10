import { reactive, computed, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { probeIP as probeIPHttp } from '../api/stats'

/**
 * IP 位图探测状态机：格子颜色/单点探测/批量测试/网段计数
 * @param {object} options
 * @param {(payload: object) => boolean} options.getSend 取 WS 发送函数（true=已发送走 WS）
 * @param {() => boolean} options.isUnmounted 组件是否已卸载（批量循环中止）
 * @param {() => Array} options.getSubnets 取当前网段列表
 */
export function useProbeGrid({ getSend, isUnmounted, getSubnets }) {
  const probing = reactive({})
  const cellColors = reactive({})
  const batchState = reactive({})
  const batchTested = reactive({})
  let suppressMessages = false
  const probeResolvers = {}

  function getIP(subnet, offset) {
    const parts = subnet.split('.')
    const third = parseInt(parts[2])
    const fourth = offset
    return `${parts[0]}.${parts[1]}.${third}.${fourth}`
  }

  const subnetCounts = computed(() => {
    const counts = {}
    for (const subnet of getSubnets()) {
      let red = 0, yellow = 0, green = 0
      for (let i = 0; i < 256; i++) {
        const ip = getIP(subnet.subnet, i)
        const c = cellColors[ip]
        if (c === 'red') red++
        else if (c === 'yellow') yellow++
        else green++
      }
      counts[subnet.subnet] = { red, yellow, green }
    }
    return counts
  })

  onUnmounted(() => {
    Object.keys(probeResolvers).forEach(ip => {
      probeResolvers[ip]()
      delete probeResolvers[ip]
    })
  })

  function applyProbeResult(data) {
    if (!data || !data.ip || !data.color) return
    cellColors[data.ip] = data.color
    probing[data.ip] = false

    if (!suppressMessages) {
      if (data.color === 'red') {
        ElMessage.error(`${data.ip} 已确认使用`)
      } else if (data.color === 'yellow') {
        ElMessage.warning(`${data.ip} 探测完成，状态待确认`)
      } else {
        ElMessage.success(`${data.ip} 探测完成，IP未使用`)
      }
    }

    const resolver = probeResolvers[data.ip]
    if (resolver) {
      delete probeResolvers[data.ip]
      resolver()
    }
  }

  function initCellColors() {
    for (const subnet of getSubnets()) {
      for (let i = 0; i < 256; i++) {
        const ip = getIP(subnet.subnet, i)
        if (subnet.used_ips && subnet.used_ips.includes(ip)) {
          cellColors[ip] = 'red'
        } else if (subnet.empty_ips && subnet.empty_ips.includes(ip)) {
          cellColors[ip] = 'yellow'
        } else {
          cellColors[ip] = 'green'
        }
      }
    }
  }

  function probeAndWait(ip) {
    return new Promise((resolve) => {
      let settled = false
      const finish = () => {
        if (settled) return
        settled = true
        clearTimeout(timer)
        delete probeResolvers[ip]
        resolve()
      }
      const timer = setTimeout(() => {
        probing[ip] = false
        finish()
      }, 10000)

      probeResolvers[ip] = finish

      probing[ip] = true
      const color = cellColors[ip] || 'green'
      const payload = { ip, color }

      if (getSend()(payload)) return

      probeIPHttp(ip, color)
        .then(res => {
          if (res.code === 0) {
            applyProbeResult({ ip, color: res.data.color })
          } else {
            probing[ip] = false
          }
          finish()
        })
        .catch(() => {
          probing[ip] = false
          finish()
        })
    })
  }

  async function handleBatchTest(subnet) {
    const name = subnet.subnet
    if (batchState[name]?.running) return

    const ips = []
    for (let i = 0; i < 256; i++) {
      ips.push(getIP(name, i))
    }

    ips.forEach(ip => { delete batchTested[ip] })
    batchState[name] = { total: 256, done: 0, running: true }
    suppressMessages = true

    const concurrency = 32
    let index = 0
    const worker = async () => {
      while (index < ips.length) {
        if (isUnmounted()) return
        const ip = ips[index++]
        try {
          await probeAndWait(ip)
        } catch (e) {
          // ignore single failure
        }
        batchState[name].done++
        batchTested[ip] = true
      }
    }

    try {
      await Promise.all(Array.from({ length: concurrency }, worker))
      ElMessage.success(`${name} 全量测试完成`)
    } finally {
      if (batchState[name]) {
        batchState[name].running = false
      }
      suppressMessages = false
      setTimeout(() => {
        ips.forEach(ip => { delete batchTested[ip] })
      }, 2000)
    }
  }

  async function handleCellClick(subnet, offset) {
    if (offset > 255) return
    const ip = getIP(subnet.subnet, offset)
    if (probing[ip]) return

    probing[ip] = true
    const currentColor = cellColors[ip] || 'green'
    const payload = { ip, color: currentColor }

    if (getSend()(payload)) return

    try {
      const res = await probeIPHttp(ip, currentColor)
      if (res.code === 0) {
        applyProbeResult({ ip, color: res.data.color })
      } else {
        probing[ip] = false
      }
    } catch (e) {
      probing[ip] = false
      ElMessage.error(`${ip} 探测失败`)
    }
  }

  return {
    probing, cellColors, batchState, batchTested, subnetCounts,
    initCellColors, applyProbeResult, handleBatchTest, handleCellClick
  }
}
