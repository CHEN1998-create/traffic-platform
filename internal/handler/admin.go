package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"traffic-platform/internal/service"
)

// AdminHandler 处理管理端接口。
type AdminHandler struct {
	ingest *service.IngestService
	agg    *service.AggregationService
}

func NewAdminHandler(ingest *service.IngestService, agg *service.AggregationService) *AdminHandler {
	return &AdminHandler{ingest: ingest, agg: agg}
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
		Window string `json:"window"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		Fail(c, http.StatusBadRequest, 40001, "invalid request body: "+err.Error())
		return
	}
	if err := h.agg.Run(c.Request.Context(), in.Window); err != nil {
		Fail(c, http.StatusBadRequest, 40006, err.Error())
		return
	}
	OK(c, gin.H{"window": in.Window, "status": "completed"})
}
