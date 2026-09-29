package service

import (
	"context"
	"time"

	"traffic-platform/internal/model"
	"traffic-platform/internal/store"
)

// Overview 总览卡片数据。
type Overview struct {
	TotalVehicles int           `json:"totalVehicles"`
	ActiveAlerts  int           `json:"activeAlerts"`
	TopCongested  *Intersection `json:"topCongested"`
}

// Intersection 路口指标。
type Intersection struct {
	IntersectionID  string  `json:"intersectionId"`
	TotalVehicles   int     `json:"totalVehicles"`
	AvgSpeed        float64 `json:"avgSpeed"`
	CongestionIndex float64 `json:"congestionIndex"`
}

// TrendPoint 趋势图时间序列点。
type TrendPoint struct {
	WindowStart     time.Time `json:"windowStart"`
	TotalVehicles   int       `json:"totalVehicles"`
	AvgSpeed        float64   `json:"avgSpeed"`
	CongestionIndex float64   `json:"congestionIndex"`
}

// DashboardService 提供看板查询逻辑。
type DashboardService struct {
	store store.Store
}

func NewDashboardService(s store.Store) *DashboardService {
	return &DashboardService{store: s}
}

// Overview 汇总：今日总车流、未处理告警数、最拥堵路口。
func (s *DashboardService) Overview(ctx context.Context) (*Overview, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	total, err := s.store.SumVehiclesBetween(ctx, startOfDay, now)
	if err != nil {
		return nil, err
	}

	alerts, err := s.store.ListAlerts(ctx, "", "")
	if err != nil {
		return nil, err
	}
	active := 0
	for _, a := range alerts {
		if a.Status != model.AlertStatusResolved {
			active++
		}
	}

	var topCongested *Intersection
	top, err := s.store.TopIntersections(ctx, store.AggWindow5m, 1)
	if err == nil && len(top) > 0 {
		i := toIntersection(top[0])
		topCongested = &i
	}

	return &Overview{
		TotalVehicles: total,
		ActiveAlerts:  active,
		TopCongested:  topCongested,
	}, nil
}

// Trend 返回时间范围内的时序数据，默认最近 1 小时。
func (s *DashboardService) Trend(ctx context.Context, start, end time.Time) ([]TrendPoint, error) {
	if start.IsZero() {
		start = time.Now().Add(-time.Hour)
	}
	if end.IsZero() {
		end = time.Now()
	}
	aggs, err := s.store.AggTrend(ctx, store.AggWindow1m, start, end)
	if err != nil {
		return nil, err
	}
	out := make([]TrendPoint, 0, len(aggs))
	for _, a := range aggs {
		out = append(out, TrendPoint{
			WindowStart:     a.WindowStart,
			TotalVehicles:   a.TotalVehicles,
			AvgSpeed:        a.AvgSpeed,
			CongestionIndex: a.CongestionIndex,
		})
	}
	return out, nil
}

// TopIntersections 返回拥堵路口排行，默认 Top10。
func (s *DashboardService) TopIntersections(ctx context.Context, limit int) ([]Intersection, error) {
	if limit <= 0 {
		limit = 10
	}
	aggs, err := s.store.TopIntersections(ctx, store.AggWindow5m, limit)
	if err != nil {
		return nil, err
	}
	out := make([]Intersection, 0, len(aggs))
	for _, a := range aggs {
		out = append(out, toIntersection(a))
	}
	return out, nil
}

func toIntersection(a *model.TrafficAgg) Intersection {
	return Intersection{
		IntersectionID:  a.IntersectionID,
		TotalVehicles:   a.TotalVehicles,
		AvgSpeed:        a.AvgSpeed,
		CongestionIndex: a.CongestionIndex,
	}
}
