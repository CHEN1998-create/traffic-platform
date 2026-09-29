import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { Alert, ImportJob } from '../api/types'

export default function Operations() {
  const [jobs, setJobs] = useState<ImportJob[]>([])
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [error, setError] = useState('')
  const [aggregating, setAggregating] = useState(false)

  const loadAll = () => {
    Promise.all([api.importJobs(), api.alerts()])
      .then(([j, a]) => {
        setJobs(j)
        setAlerts(a)
      })
      .catch((e: Error) => setError(e.message))
  }

  useEffect(() => {
    loadAll()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const onAggregate = async (window: string) => {
    setAggregating(true)
    try {
      await api.aggregateRun(window)
      await loadAll()
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setAggregating(false)
    }
  }

  return (
    <div>
      <h2>任务与告警管理</h2>
      {error ? <div className="error">加载失败：{error}</div> : null}

      <h3>聚合任务</h3>
      <div style={{ marginBottom: 12, display: 'flex', gap: 8 }}>
        <button onClick={() => onAggregate('1m')} disabled={aggregating}>
          {aggregating ? '执行中...' : '手动触发 1m 聚合'}
        </button>
        <button onClick={() => onAggregate('5m')} disabled={aggregating}>
          {aggregating ? '执行中...' : '手动触发 5m 聚合'}
        </button>
      </div>

      <h3>导入任务</h3>
      <table className="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>文件名</th>
            <th>状态</th>
            <th>总行数</th>
            <th>成功</th>
            <th>失败</th>
          </tr>
        </thead>
        <tbody>
          {jobs.map((j) => (
            <tr key={j.id}>
              <td>{j.id}</td>
              <td>{j.filename}</td>
              <td>{j.status}</td>
              <td>{j.totalRows}</td>
              <td>{j.successRows}</td>
              <td>{j.failedRows}</td>
            </tr>
          ))}
        </tbody>
      </table>

      <h3>告警处理记录</h3>
      <table className="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>路口</th>
            <th>级别</th>
            <th>规则</th>
            <th>状态</th>
            <th>处理时间</th>
          </tr>
        </thead>
        <tbody>
          {alerts.map((a) => (
            <tr key={a.id}>
              <td>{a.id}</td>
              <td>{a.intersectionId}</td>
              <td>{a.level}</td>
              <td>{a.ruleCode}</td>
              <td>{a.status}</td>
              <td>{a.resolvedAt ? new Date(a.resolvedAt).toLocaleString() : '—'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
