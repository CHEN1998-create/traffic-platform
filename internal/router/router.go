package router

import (
	"net/http"
	"os"
	"strings"

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
		traffic.POST("/simulate", h.Ingest.Simulate)

		dashboard := api.Group("/dashboard")
		dashboard.GET("/overview", h.Dashboard.Overview)
		dashboard.GET("/trend", h.Dashboard.Trend)
		dashboard.GET("/intersections/top", h.Dashboard.TopIntersections)

		alerts := api.Group("/alerts")
		alerts.GET("", h.Alerts.List)
		alerts.PATCH("/:id/resolve", h.Alerts.Resolve)
		alerts.PATCH("/:id/ack", h.Alerts.Ack)

		admin := api.Group("/admin")
		admin.GET("/import-jobs", h.Admin.ImportJobs)
		admin.POST("/aggregate/run", h.Admin.AggregateRun)
	}

	// 托管前端静态文件（生产环境）。本地开发时 frontend/dist 不存在，仅提供 API。
	dist := "./frontend/dist"
	if isDir(dist) {
		r.Static("/assets", dist+"/assets")
	}
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"code": 40400, "message": "not found"})
			return
		}
		if isDir(dist) {
			c.File(dist + "/index.html")
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"code": 40400, "message": "not found"})
	})

	return r
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
