import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { Overview as OverviewData } from '../api/types'
import MetricCard from '../components/MetricCard'

export default function Overview() {
  const [data, setData] = useState<OverviewData | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    api
      .overview()
      .then(setData)
      .catch((e: Error) => setError(e.message))
  }, [])

  if (error) return <div className="error">加载失败：{error}（请确认后端已启动）</div>
  if (!data) return <div className="loading">加载中...</div>

  const top = data.topCongested

  return (
    <div>
      <h2>总览看板</h2>
      <div className="metric-grid">
        <MetricCard title="总车流" value={data.totalVehicles} unit="辆" />
        <MetricCard title="当前告警数" value={data.activeAlerts} unit="条" />
        <MetricCard
          title="最拥堵路口"
          value={top ? top.intersectionId : '—'}
          unit={top ? `指数 ${top.congestionIndex.toFixed(1)}` : undefined}
        />
      </div>
    </div>
  )
}
