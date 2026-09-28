// 与后端统一响应结构对应
export interface ApiResponse<T> {
  code: number
  message: string
  data?: T
}

// 原始交通事件
export interface TrafficEvent {
  id: number
  intersectionId: string
  timestamp: string
  vehicleCount: number
  avgSpeed: number
  source: string
  createdAt: string
}

// 路口指标
export interface Intersection {
  intersectionId: string
  totalVehicles: number
  avgSpeed: number
  congestionIndex: number
}

// 总览卡片
export interface Overview {
  totalVehicles: number
  activeAlerts: number
  topCongested: Intersection | null
}

// 趋势时间序列点
export interface TrendPoint {
  windowStart: string
  totalVehicles: number
  avgSpeed: number
  congestionIndex: number
}

// 告警
export interface Alert {
  id: number
  intersectionId: string
  level: string
  ruleCode: string
  status: string
  message: string
  createdAt: string
  resolvedAt: string | null
}

// 导入任务
export interface ImportJob {
  id: number
  filename: string
  status: string
  totalRows: number
  successRows: number
  failedRows: number
  createdAt: string
}
