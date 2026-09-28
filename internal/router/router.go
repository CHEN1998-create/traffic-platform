package router

import (
	"github.com/gin-gonic/gin"

	"traffic-platform/internal/handler"
)

// New 注册全部路由（对齐 PRD 第 8 节接口草案）。
func New(h *handler.Handlers) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/healthz", h.Healthz)

	api := r.Group("/api")
	{
		traffic := api.Group("/traffic")
		traffic.POST("/events", h.Ingest.PostEvent)
		traffic.POST("/import", h.Ingest.Import)

		dashboard := api.Group("/dashboard")
		dashboard.GET("/overview", h.Dashboard.Overview)
		dashboard.GET("/trend", h.Dashboard.Trend)
		dashboard.GET("/intersections/top", h.Dashboard.TopIntersections)

		alerts := api.Group("/alerts")
		alerts.GET("", h.Alerts.List)
		alerts.PATCH("/:id/resolve", h.Alerts.Resolve)

		admin := api.Group("/admin")
		admin.GET("/import-jobs", h.Admin.ImportJobs)
	}

	return r
}
