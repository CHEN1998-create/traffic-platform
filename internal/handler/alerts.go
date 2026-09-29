package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"traffic-platform/internal/service"
)

// AlertsHandler 处理告警接口。
type AlertsHandler struct {
	svc *service.AlertService
}

func NewAlertsHandler(svc *service.AlertService) *AlertsHandler {
	return &AlertsHandler{svc: svc}
}

// List 告警列表，支持 level / status 筛选。
func (h *AlertsHandler) List(c *gin.Context) {
	level := c.Query("level")
	status := c.Query("status")
	data, err := h.svc.List(c.Request.Context(), level, status)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	OK(c, data)
}

// Resolve 处理告警，标记为已处理。
func (h *AlertsHandler) Resolve(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, 40005, "invalid alert id")
		return
	}
	if err := h.svc.Resolve(c.Request.Context(), id); err != nil {
		Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	OK(c, gin.H{"id": id, "status": "resolved"})
}

// Ack 确认告警，标记为已确认。
func (h *AlertsHandler) Ack(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, 40005, "invalid alert id")
		return
	}
	if err := h.svc.Ack(c.Request.Context(), id); err != nil {
		Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	OK(c, gin.H{"id": id, "status": "acked"})
}
