package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"traffic-platform/internal/aggregation"
	"traffic-platform/internal/config"
	"traffic-platform/internal/handler"
	"traffic-platform/internal/router"
	"traffic-platform/internal/service"
	"traffic-platform/internal/store"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	// 1. 建立数据库连接
	st, err := store.NewPostgresStore(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer st.Close()
	log.Println("database connected")

	// 2. 装配 service
	ingestSvc := service.NewIngestService(st)
	aggSvc := service.NewAggregationService(st, cfg)
	alertSvc := service.NewAlertService(st, cfg)
	dashSvc := service.NewDashboardService(st)

	// 3. 装配 handler 与路由
	h := handler.NewHandlers(
		handler.NewIngestHandler(ingestSvc),
		handler.NewDashboardHandler(dashSvc),
		handler.NewAlertsHandler(alertSvc),
		handler.NewAdminHandler(ingestSvc),
	)
	r := router.New(h)

	// 4. 启动定时聚合与告警
	sched := aggregation.NewScheduler(aggSvc, alertSvc)
	if err := sched.Start(); err != nil {
		log.Fatalf("start scheduler: %v", err)
	}

	// 5. 启动 HTTP 服务
	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}
	go func() {
		log.Printf("server listening on %s", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// 6. 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	sched.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("server stopped")
}
