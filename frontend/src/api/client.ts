import axios, { type AxiosResponse } from 'axios'
import type {
  ApiResponse,
  Overview,
  TrendPoint,
  Intersection,
  Alert,
  ImportJob,
} from './types'

const http = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

// 统一解包后端响应 { code, message, data }
async function unwrap<T>(p: Promise<AxiosResponse<ApiResponse<T>>>): Promise<T> {
  const res = await p
  if (res.data.code !== 0) {
    throw new Error(res.data.message || '请求失败')
  }
  return res.data.data as T
}

export const api = {
  // 看板
  overview: () => unwrap<Overview>(http.get<ApiResponse<Overview>>('/dashboard/overview')),
  trend: (start?: string, end?: string) =>
    unwrap<TrendPoint[]>(
      http.get<ApiResponse<TrendPoint[]>>('/dashboard/trend', { params: { start, end } }),
    ),
  topIntersections: (limit = 10) =>
    unwrap<Intersection[]>(
      http.get<ApiResponse<Intersection[]>>('/dashboard/intersections/top', {
        params: { limit },
      }),
    ),

  // 告警
  alerts: (level = '', status = '') =>
    unwrap<Alert[]>(http.get<ApiResponse<Alert[]>>('/alerts', { params: { level, status } })),
  resolveAlert: (id: number) =>
    unwrap<{ id: number; status: string }>(
      http.patch<ApiResponse<{ id: number; status: string }>>(`/alerts/${id}/resolve`),
    ),
  ackAlert: (id: number) =>
    unwrap<{ id: number; status: string }>(
      http.patch<ApiResponse<{ id: number; status: string }>>(`/alerts/${id}/ack`),
    ),

  // 管理端
  importJobs: () => unwrap<ImportJob[]>(http.get<ApiResponse<ImportJob[]>>('/admin/import-jobs')),
  importJson: (events: unknown[]) =>
    unwrap<ImportJob>(http.post<ApiResponse<ImportJob>>('/traffic/import', events)),
  importCsv: (file: File) => {
    const form = new FormData()
    form.append('file', file)
    return unwrap<ImportJob>(http.post<ApiResponse<ImportJob>>('/traffic/import', form))
  },
  simulate: (params: { intersections?: number; minutes?: number; eventsPerMinute?: number }) =>
    unwrap<ImportJob>(http.post<ApiResponse<ImportJob>>('/traffic/simulate', params)),
  aggregateRun: (window: string) =>
    unwrap<{ window: string; status: string }>(
      http.post<ApiResponse<{ window: string; status: string }>>('/admin/aggregate/run', {
        window,
      }),
    ),
}
