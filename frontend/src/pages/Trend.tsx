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
    title: { text: '交通趋势' },
    tooltip: { trigger: 'axis' },
    legend: { data: ['车流量', '平均速度', '拥堵指数'] },
    xAxis: {
      type: 'category',
      data: points.map((p) => new Date(p.windowStart).toLocaleTimeString()),
    },
    yAxis: [
      { type: 'value', name: '车流量' },
      { type: 'value', name: '速度/指数' },
    ],
    series: [
      {
        name: '车流量',
        type: 'line',
        data: points.map((p) => p.totalVehicles),
        smooth: true,
        areaStyle: {},
      },
      {
        name: '平均速度',
        type: 'line',
        yAxisIndex: 1,
        data: points.map((p) => Math.round(p.avgSpeed * 10) / 10),
        smooth: true,
      },
      {
        name: '拥堵指数',
        type: 'line',
        yAxisIndex: 1,
        data: points.map((p) => Math.round(p.congestionIndex * 10) / 10),
        smooth: true,
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
      {error ? (
        <div className="error">加载失败：{error}</div>
      ) : points.length === 0 ? (
        <div className="empty">暂无趋势数据（请先生成模拟数据并触发聚合）</div>
      ) : (
        <Chart option={option} height={360} />
      )}
    </div>
  )
}
