import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { Alert } from '../api/types'

export default function Alerts() {
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [level, setLevel] = useState('')
  const [status, setStatus] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    api
      .alerts(level, status)
      .then(setAlerts)
      .catch((e: Error) => setError(e.message))
  }, [level, status])

  const resolve = async (id: number) => {
    await api.resolveAlert(id)
    const list = await api.alerts(level, status)
    setAlerts(list)
  }

  const ack = async (id: number) => {
    await api.ackAlert(id)
    const list = await api.alerts(level, status)
    setAlerts(list)
  }

  const levelBadge = (level: string) =>
    level === 'critical' ? 'badge badge-critical' : level === 'warning' ? 'badge badge-warning' : 'badge'

  const statusLabel = (status: string) => {
    if (status === 'new') return '新建'
    if (status === 'acked') return '已确认'
    if (status === 'resolved') return '已处理'
    return status
  }

  const statusBadge = (status: string) =>
    status === 'new' ? 'badge badge-new' : status === 'acked' ? 'badge badge-acked' : 'badge badge-resolved'

  return (
    <div>
      <h2>告警</h2>
      <div style={{ marginBottom: 16, display: 'flex', gap: 12 }}>
        <select value={level} onChange={(e) => setLevel(e.target.value)}>
          <option value="">全部级别</option>
          <option value="critical">critical</option>
          <option value="warning">warning</option>
        </select>
        <select value={status} onChange={(e) => setStatus(e.target.value)}>
          <option value="">全部状态</option>
          <option value="new">new</option>
          <option value="acked">acked</option>
          <option value="resolved">resolved</option>
        </select>
      </div>
      {error ? (
        <div className="error">加载失败：{error}</div>
      ) : (
        <table className="table">
          <thead>
            <tr>
              <th>ID</th>
              <th>路口</th>
              <th>级别</th>
              <th>规则</th>
              <th>状态</th>
              <th>消息</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {alerts.length === 0 && (
              <tr>
                <td colSpan={7} className="empty">暂无告警</td>
              </tr>
            )}
            {alerts.map((a) => (
              <tr key={a.id}>
                <td>{a.id}</td>
                <td>{a.intersectionId}</td>
                <td>
                  <span className={levelBadge(a.level)}>{a.level}</span>
                </td>
                <td>{a.ruleCode}</td>
                <td>
                  <span className={statusBadge(a.status)}>{statusLabel(a.status)}</span>
                </td>
                <td>{a.message}</td>
                <td>
                  {a.status === 'new' && <button onClick={() => ack(a.id)}>确认</button>}
                  {a.status !== 'resolved' && (
                    <button onClick={() => resolve(a.id)}>标记已处理</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
