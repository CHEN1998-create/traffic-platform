import { useEffect, useState } from 'react'
import type { EChartsOption } from 'echarts'
import { api } from '../api/client'
import type { TrendPoint } from '../api/types'
import Chart from '../components/Chart'

const ranges = [
  { label: '最近 1 小时', hours: 1 },
  { label: '最近 6 小时', hours: 6 },
  { label: '最近 24 小时', hours: 24 },
]

export default function Trend() {
  const [hours, setHours] = useState(1)
  const [points, setPoints] = useState<TrendPoint[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    const end = new Date()
    const start = new Date(end.getTime() - hours * 3600 * 1000)
    api
      .trend(start.toISOString(), end.toISOString())
      .then(setPoints)
      .catch((e: Error) => setError(e.message))
  }, [hours])

  const option: EChartsOption = {
    title: { text: '车流趋势' },
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      data: points.map((p) => new Date(p.windowStart).toLocaleTimeString()),
    },
    yAxis: { type: 'value' },
    series: [
      {
        name: '车流量',
        type: 'line',
        data: points.map((p) => p.totalVehicles),
        smooth: true,
        areaStyle: {},
      },
    ],
  }

  return (
    <div>
      <h2>趋势分析</h2>
      <div className="range-switch">
        {ranges.map((r) => (
          <button
            key={r.hours}
            className={hours === r.hours ? 'active' : ''}
            onClick={() => setHours(r.hours)}
          >
            {r.label}
          </button>
        ))}
      </div>
      {error ? <div className="error">加载失败：{error}</div> : <Chart option={option} height={360} />}
    </div>
  )
}
