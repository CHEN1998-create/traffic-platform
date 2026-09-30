package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"traffic-platform/internal/service"
	"traffic-platform/internal/store"
)

// AdminHandler 处理管理端接口。
type AdminHandler struct {
	ingest *service.IngestService
	agg    *service.AggregationService
	alerts *service.AlertService
}

func NewAdminHandler(ingest *service.IngestService, agg *service.AggregationService, alerts *service.AlertService) *AdminHandler {
	return &AdminHandler{ingest: ingest, agg: agg, alerts: alerts}
}

// ImportJobs 导入任务状态列表。
func (h *AdminHandler) ImportJobs(c *gin.Context) {
	jobs, err := h.ingest.ListImportJobs(c.Request.Context())
	if err != nil {
		Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	OK(c, jobs)
}

// AggregateRun 手动触发聚合任务。
func (h *AdminHandler) AggregateRun(c *gin.Context) {
	var in struct {
		Window  string `json:"window"`
		Minutes int    `json:"minutes"` // 可选：回填最近 N 分钟的所有窗口
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		Fail(c, http.StatusBadRequest, 40001, "invalid request body: "+err.Error())
		return
	}
	if in.Minutes > 0 {
		// 回填：聚合最近 minutes 分钟的所有窗口
		end := time.Now()
		start := end.Add(-time.Duration(in.Minutes) * time.Minute)
		if err := h.agg.RunRange(c.Request.Context(), store.AggWindow(in.Window), start, end); err != nil {
			Fail(c, http.StatusBadRequest, 40006, err.Error())
			return
		}
	} else {
		if err := h.agg.Run(c.Request.Context(), in.Window); err != nil {
			Fail(c, http.StatusBadRequest, 40006, err.Error())
			return
		}
	}
	// 1m 聚合后评估告警（等价于 cron 的完整流程）
	if in.Window == "1m" {
		if err := h.alerts.Evaluate(c.Request.Context()); err != nil {
			Fail(c, http.StatusInternalServerError, 50000, err.Error())
			return
		}
	}
	OK(c, gin.H{"window": in.Window, "status": "completed"})
}
