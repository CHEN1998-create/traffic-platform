package service

import (
	"context"
	"fmt"
	"time"

	"traffic-platform/internal/config"
	"traffic-platform/internal/model"
	"traffic-platform/internal/store"
)

// 告警规则码（对齐 PRD 第 7 节）。
const (
	RuleFlowSpike = "flow_spike" // 流量高于近 5 分钟均值阈值
	RuleLowSpeed  = "low_speed"  // 速度连续低于阈值
)

// AlertService 负责告警规则评估与告警处理。
type AlertService struct {
	store  store.Store
	config *config.Config
}

func NewAlertService(s store.Store, cfg *config.Config) *AlertService {
	return &AlertService{store: s, config: cfg}
}

// Evaluate 基于最新 1 分钟聚合评估告警规则。
func (s *AlertService) Evaluate(ctx context.Context) error {
	latest, err := s.store.LatestAgg(ctx, store.AggWindow1m)
	if err != nil {
		return err
	}
	for _, a := range latest {
		if err := s.evaluateIntersection(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

func (s *AlertService) evaluateIntersection(ctx context.Context, a *model.TrafficAgg) error {
	// 规则 1：平均速度连续 N 个窗口低于阈值
	if err := s.evaluateLowSpeed(ctx, a); err != nil {
		return err
	}

	// 规则 2：当前流量高于近 5 分钟均值阈值
	since := time.Now().Add(-5 * time.Minute)
	recent, err := s.store.AggRecentForIntersection(ctx, store.AggWindow1m, a.IntersectionID, since)
	if err != nil {
		return err
	}
	var sum float64
	var n int
	for _, r := range recent {
		if r.ID == a.ID {
			continue
		}
		sum += float64(r.TotalVehicles)
		n++
	}
	if n > 0 {
		avg := sum / float64(n)
		if avg > 0 && float64(a.TotalVehicles) > avg*s.config.FlowSpikeFactor {
			if err := s.ensureAlert(ctx, a.IntersectionID, RuleFlowSpike, "warning",
				fmt.Sprintf("vehicle count %d exceeds %.1fx of 5m average %.1f", a.TotalVehicles, s.config.FlowSpikeFactor, avg)); err != nil {
				return err
			}
		}
	}
	return nil
}

// evaluateLowSpeed 判断速度是否连续 N 个窗口低于阈值（PRD 第 7 节「连续低于阈值」）。
func (s *AlertService) evaluateLowSpeed(ctx context.Context, a *model.TrafficAgg) error {
	consecutive := s.config.LowSpeedConsecutiveWindows
	if consecutive <= 0 {
		consecutive = 3
	}
	if a.AvgSpeed >= s.config.LowSpeedThreshold {
		return nil
	}

	// 查询最近 consecutive 个窗口（含当前）的聚合
	since := a.WindowStart.Add(-time.Duration(consecutive-1) * time.Minute)
	recent, err := s.store.AggRecentForIntersection(ctx, store.AggWindow1m, a.IntersectionID, since)
	if err != nil {
		return err
	}

	byWindow := make(map[int64]float64, len(recent))
	for _, r := range recent {
		byWindow[r.WindowStart.Unix()] = r.AvgSpeed
	}

	// 从当前窗口往前检查 consecutive 个连续窗口是否都低于阈值
	for i := 0; i < consecutive; i++ {
		ws := a.WindowStart.Add(-time.Duration(i) * time.Minute)
		speed, ok := byWindow[ws.Unix()]
		if !ok || speed >= s.config.LowSpeedThreshold {
			return nil // 存在窗口缺口或某窗口速度不低于阈值
		}
	}

	return s.ensureAlert(ctx, a.IntersectionID, RuleLowSpeed, "critical",
		fmt.Sprintf("avg speed below %.1f km/h for %d consecutive minutes", s.config.LowSpeedThreshold, consecutive))
}

// ensureAlert 避免重复生成同一路口同一规则的未处理告警。
func (s *AlertService) ensureAlert(ctx context.Context, intersectionID, ruleCode, level, message string) error {
	exists, err := s.store.HasActiveAlert(ctx, intersectionID, ruleCode)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	a := &model.Alert{
		IntersectionID: intersectionID,
		Level:          level,
		RuleCode:       ruleCode,
		Status:         model.AlertStatusNew,
		Message:        message,
	}
	return s.store.InsertAlert(ctx, a)
}

// List 返回告警列表，支持按级别、状态筛选。
func (s *AlertService) List(ctx context.Context, level, status string) ([]*model.Alert, error) {
	return s.store.ListAlerts(ctx, level, status)
}

// Resolve 将告警标记为已处理。
func (s *AlertService) Resolve(ctx context.Context, id int64) error {
	return s.store.ResolveAlert(ctx, id)
}
