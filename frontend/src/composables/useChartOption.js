import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { GraphicComponent } from 'echarts/components'

/**
 * ECharts 按需注册与图表组件获取
 * @param {Array} charts echarts/charts 与 echarts/components 的注册项（CanvasRenderer / GraphicComponent 已内置）
 * @returns {{ VChart: object }} 模板中以 <v-chart> 使用
 */
export function useChartOption(charts = []) {
  use([CanvasRenderer, GraphicComponent, ...charts])
  return { VChart }
}
