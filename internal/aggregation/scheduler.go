package aggregation

import (
	"context"
	"log"

	"github.com/robfig/cron/v3"

	"traffic-platform/internal/service"
)

// Scheduler 基于 robfig/cron 的定时聚合与告警调度骨架。
type Scheduler struct {
	cron   *cron.Cron
	agg    *service.AggregationService
	alerts *service.AlertService
}

func NewScheduler(agg *service.AggregationService, alerts *service.AlertService) *Scheduler {
	return &Scheduler{
		cron:   cron.New(cron.WithSeconds()),
		agg:    agg,
		alerts: alerts,
	}
}

// Start 注册并启动定时任务：
//   - 每分钟：先执行 1m 聚合，再评估告警规则
//   - 每 5 分钟：执行 5m 聚合
func (s *Scheduler) Start() error {
	if _, err := s.cron.AddFunc("0 * * * * *", func() {
		ctx := context.Background()
		if err := s.agg.Run1m(ctx); err != nil {
			log.Printf("[scheduler] 1m aggregation failed: %v", err)
			return
		}
		if err := s.alerts.Evaluate(ctx); err != nil {
			log.Printf("[scheduler] alert evaluate failed: %v", err)
		}
	}); err != nil {
		return err
	}

	if _, err := s.cron.AddFunc("0 */5 * * * *", func() {
		if err := s.agg.Run5m(context.Background()); err != nil {
			log.Printf("[scheduler] 5m aggregation failed: %v", err)
		}
	}); err != nil {
		return err
	}

	s.cron.Start()
	log.Println("[scheduler] started")
	return nil
}

// Stop 停止调度器并等待正在执行的任务结束。
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("[scheduler] stopped")
}
