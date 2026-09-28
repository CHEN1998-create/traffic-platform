package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"traffic-platform/internal/service"
)

// AdminHandler 处理管理端接口。
type AdminHandler struct {
	ingest *service.IngestService
}

func NewAdminHandler(ingest *service.IngestService) *AdminHandler {
	return &AdminHandler{ingest: ingest}
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
