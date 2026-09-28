package service

import (
	"context"
	"log"
	"time"

	"traffic-platform/internal/config"
	"traffic-platform/internal/model"
	"traffic-platform/internal/store"
)

// AggregationService 负责把原始事件聚合成 1m / 5m 窗口指标。
// 骨架阶段仅做最简聚合：车流量求和、平均速度、拥堵指数。
type AggregationService struct {
	store  store.Store
	config *config.Config
}

func NewAggregationService(s store.Store, cfg *config.Config) *AggregationService {
	return &AggregationService{store: s, config: cfg}
}

// Run1m 聚合上一个完整 1 分钟窗口。
func (s *AggregationService) Run1m(ctx context.Context) error {
	end := time.Now().Truncate(time.Minute)
	start := end.Add(-time.Minute)
	return s.runWindow(ctx, store.AggWindow1m, start, end)
}

// Run5m 聚合上一个完整 5 分钟窗口。
func (s *AggregationService) Run5m(ctx context.Context) error {
	end := time.Now().Truncate(5 * time.Minute)
	start := end.Add(-5 * time.Minute)
	return s.runWindow(ctx, store.AggWindow5m, start, end)
}

// runWindow 读取 [start, end) 内的事件，按路口聚合后落库。
func (s *AggregationService) runWindow(ctx context.Context, window store.AggWindow, start, end time.Time) error {
	events, err := s.store.EventsBetween(ctx, start, end)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}

	grouped := groupByIntersection(events)
	for id, list := range grouped {
		agg := aggregateList(id, start, list, s.config.FreeFlowSpeed)
		if err := s.store.InsertAgg(ctx, window, agg); err != nil {
			return err
		}
	}
	log.Printf("[aggregation] window=%s start=%s intersections=%d", window, start.Format(time.RFC3339), len(grouped))
	return nil
}

func groupByIntersection(events []*model.RawTrafficEvent) map[string][]*model.RawTrafficEvent {
	grouped := make(map[string][]*model.RawTrafficEvent)
	for _, e := range events {
		grouped[e.IntersectionID] = append(grouped[e.IntersectionID], e)
	}
	return grouped
}

func aggregateList(intersectionID string, windowStart time.Time, list []*model.RawTrafficEvent, freeFlowSpeed float64) *model.TrafficAgg {
	agg := &model.TrafficAgg{
		IntersectionID: intersectionID,
		WindowStart:    windowStart,
	}
	for _, e := range list {
		agg.TotalVehicles += e.VehicleCount
		agg.AvgSpeed += e.AvgSpeed
	}
	if len(list) > 0 {
		agg.AvgSpeed /= float64(len(list))
	}
	agg.CongestionIndex = congestionIndex(agg.AvgSpeed, freeFlowSpeed)
	return agg
}

// congestionIndex 最简拥堵指数：速度越低指数越高（0~100）。
func congestionIndex(avgSpeed, freeFlowSpeed float64) float64 {
	if freeFlowSpeed <= 0 {
		return 0
	}
	idx := (1 - avgSpeed/freeFlowSpeed) * 100
	if idx < 0 {
		return 0
	}
	if idx > 100 {
		return 100
	}
	return idx
}
