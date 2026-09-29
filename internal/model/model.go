package model

import "time"

// 告警状态流转：new -> acked -> resolved
const (
	AlertStatusNew      = "new"
	AlertStatusAcked    = "acked"
	AlertStatusResolved = "resolved"
)

// RawTrafficEvent 原始交通事件，对应 raw_traffic_events 表。
type RawTrafficEvent struct {
	ID             int64     `json:"id" db:"id"`
	IntersectionID string    `json:"intersectionId" db:"intersection_id"`
	EventTime      time.Time `json:"timestamp" db:"event_time"`
	VehicleCount   int       `json:"vehicleCount" db:"vehicle_count"`
	AvgSpeed       float64   `json:"avgSpeed" db:"avg_speed"`
	Source         string    `json:"source" db:"source"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
}

// TrafficAgg 聚合结果，对应 traffic_agg_1m / traffic_agg_5m 表。
type TrafficAgg struct {
	ID              int64     `json:"id" db:"id"`
	IntersectionID  string    `json:"intersectionId" db:"intersection_id"`
	WindowStart     time.Time `json:"windowStart" db:"window_start"`
	TotalVehicles   int       `json:"totalVehicles" db:"total_vehicles"`
	AvgSpeed        float64   `json:"avgSpeed" db:"avg_speed"`
	CongestionIndex float64   `json:"congestionIndex" db:"congestion_index"`
}

// Alert 告警，对应 alerts 表。
type Alert struct {
	ID             int64      `json:"id" db:"id"`
	IntersectionID string     `json:"intersectionId" db:"intersection_id"`
	Level          string     `json:"level" db:"level"`
	RuleCode       string     `json:"ruleCode" db:"rule_code"`
	Status         string     `json:"status" db:"status"`
	Message        string     `json:"message" db:"message"`
	CreatedAt      time.Time  `json:"createdAt" db:"created_at"`
	ResolvedAt     *time.Time `json:"resolvedAt" db:"resolved_at"`
}

// ImportJob 导入任务，对应 import_jobs 表。
type ImportJob struct {
	ID           int64     `json:"id" db:"id"`
	Filename     string    `json:"filename" db:"filename"`
	Status       string    `json:"status" db:"status"`
	TotalRows    int       `json:"totalRows" db:"total_rows"`
	SuccessRows  int       `json:"successRows" db:"success_rows"`
	FailedRows   int       `json:"failedRows" db:"failed_rows"`
	ErrorDetails string    `json:"errorDetails" db:"error_details"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}
