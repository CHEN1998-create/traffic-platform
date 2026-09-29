import { useEffect, useRef, useState } from 'react'
import { api } from '../api/client'
import type { ImportJob } from '../api/types'

export default function Imports() {
  const [jobs, setJobs] = useState<ImportJob[]>([])
  const [error, setError] = useState('')
  const [uploading, setUploading] = useState(false)
  const [simulating, setSimulating] = useState(false)
  const fileRef = useRef<HTMLInputElement>(null)

  const load = () =>
    api
      .importJobs()
      .then(setJobs)
      .catch((e: Error) => setError(e.message))

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const onUpload = async () => {
    const file = fileRef.current?.files?.[0]
    if (!file) return
    setUploading(true)
    try {
      await api.importCsv(file)
      await load()
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setUploading(false)
    }
  }

  const onSimulate = async () => {
    setSimulating(true)
    try {
      await api.simulate({ intersections: 5, minutes: 10, eventsPerMinute: 3 })
      await load()
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setSimulating(false)
    }
  }

  const renderErrors = (j: ImportJob) => {
    if (!j.errorDetails) return '—'
    try {
      const errs = JSON.parse(j.errorDetails) as { row: number; error: string }[]
      return errs
        .slice(0, 3)
        .map((e) => (e.row > 0 ? `行${e.row}: ${e.error}` : e.error))
        .join('; ')
    } catch {
      return j.errorDetails
    }
  }

  return (
    <div>
      <h2>数据导入</h2>
      <div className="upload">
        <input ref={fileRef} type="file" accept=".csv" />
        <button onClick={onUpload} disabled={uploading}>
          {uploading ? '导入中...' : '上传 CSV'}
        </button>
        <button onClick={onSimulate} disabled={simulating}>
          {simulating ? '生成中...' : '生成模拟数据'}
        </button>
      </div>
      {error ? <div className="error">错误：{error}</div> : null}
      <h3>导入任务状态</h3>
      <table className="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>文件名</th>
            <th>状态</th>
            <th>总行数</th>
            <th>成功</th>
            <th>失败</th>
            <th>错误详情</th>
            <th>时间</th>
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
              <td>{renderErrors(j)}</td>
              <td>{new Date(j.createdAt).toLocaleString()}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
