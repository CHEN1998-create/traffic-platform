package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"traffic-platform/internal/service"
)

// DashboardHandler 处理看板查询接口。
type DashboardHandler struct {
	svc *service.DashboardService
}

func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// Overview 概览卡片数据。
func (h *DashboardHandler) Overview(c *gin.Context) {
	data, err := h.svc.Overview(c.Request.Context())
	if err != nil {
		Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	OK(c, data)
}

// Trend 趋势数据，支持 start / end 时间范围参数（RFC3339）。
func (h *DashboardHandler) Trend(c *gin.Context) {
	start := parseTime(c.Query("start"))
	end := parseTime(c.Query("end"))
	data, err := h.svc.Trend(c.Request.Context(), start, end)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	OK(c, data)
}

// TopIntersections 拥堵路口排行，支持 limit 参数。
func (h *DashboardHandler) TopIntersections(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	data, err := h.svc.TopIntersections(c.Request.Context(), limit)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	OK(c, data)
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
