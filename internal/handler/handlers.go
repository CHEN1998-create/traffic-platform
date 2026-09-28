package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handlers 聚合所有 HTTP handler，便于统一装配。
type Handlers struct {
	Ingest    *IngestHandler
	Dashboard *DashboardHandler
	Alerts    *AlertsHandler
	Admin     *AdminHandler
}

func NewHandlers(ingest *IngestHandler, dashboard *DashboardHandler, alerts *AlertsHandler, admin *AdminHandler) *Handlers {
	return &Handlers{
		Ingest:    ingest,
		Dashboard: dashboard,
		Alerts:    alerts,
		Admin:     admin,
	}
}

// Healthz 健康检查。
func (h *Handlers) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
