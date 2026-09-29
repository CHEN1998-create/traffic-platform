package store

import (
	"context"
	"time"

	"traffic-platform/internal/model"
)

// AggWindow 聚合窗口类型，用于区分 1m / 5m 聚合表。
type AggWindow string

const (
	AggWindow1m AggWindow = "1m"
	AggWindow5m AggWindow = "5m"
)

// Store 定义数据访问接口。当前实现为 PostgreSQL（postgres.go）。
type Store interface {
	// Ping 用于健康检查。
	Ping(ctx context.Context) error
	Close()

	// ---- 原始事件 ----
	InsertEvent(ctx context.Context, e *model.RawTrafficEvent) error
	// InsertEvents 批量写入，返回成功写入条数；单条失败不会中断整体。
	InsertEvents(ctx context.Context, events []*model.RawTrafficEvent) (int, error)
	EventsBetween(ctx context.Context, start, end time.Time) ([]*model.RawTrafficEvent, error)
	// SumVehiclesBetween 返回时间范围内的车流总量（原始事件求和）。
	SumVehiclesBetween(ctx context.Context, start, end time.Time) (int, error)

	// ---- 聚合 ----
	InsertAgg(ctx context.Context, window AggWindow, a *model.TrafficAgg) error
	// LatestAgg 返回最新一个窗口的所有路口聚合。
	LatestAgg(ctx context.Context, window AggWindow) ([]*model.TrafficAgg, error)
	// AggTrend 返回时间范围内的时序聚合（按窗口汇总所有路口）。
	AggTrend(ctx context.Context, window AggWindow, start, end time.Time) ([]*model.TrafficAgg, error)
	// AggRecentForIntersection 返回某路口在 since 之后的聚合记录（按窗口降序）。
	AggRecentForIntersection(ctx context.Context, window AggWindow, intersectionID string, since time.Time) ([]*model.TrafficAgg, error)
	// TopIntersections 返回拥堵指数最高的路口排行。
	TopIntersections(ctx context.Context, window AggWindow, limit int) ([]*model.TrafficAgg, error)

	// ---- 告警 ----
	InsertAlert(ctx context.Context, a *model.Alert) error
	ListAlerts(ctx context.Context, level, status string) ([]*model.Alert, error)
	ResolveAlert(ctx context.Context, id int64) error
	// AckAlert 将告警标记为已确认（status=acked）。
	AckAlert(ctx context.Context, id int64) error
	HasActiveAlert(ctx context.Context, intersectionID, ruleCode string) (bool, error)

	// ---- 导入任务 ----
	InsertImportJob(ctx context.Context, j *model.ImportJob) error
	ListImportJobs(ctx context.Context) ([]*model.ImportJob, error)
}
