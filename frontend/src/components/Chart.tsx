import { useEffect, useRef } from 'react'
import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'

interface ChartProps {
  option: EChartsOption
  height?: number | string
}

export default function Chart({ option, height = 320 }: ChartProps) {
  const ref = useRef<HTMLDivElement>(null)
  const chartRef = useRef<echarts.ECharts>()

  useEffect(() => {
    if (!ref.current) return
    const chart = echarts.init(ref.current)
    chartRef.current = chart
    const onResize = () => chart.resize()
    window.addEventListener('resize', onResize)
    return () => {
      window.removeEventListener('resize', onResize)
      chart.dispose()
    }
  }, [])

  useEffect(() => {
    chartRef.current?.setOption(option)
  }, [option])

  return <div ref={ref} style={{ width: '100%', height }} />
}
