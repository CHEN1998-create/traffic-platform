package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"traffic-platform/internal/service"
)

// IngestHandler 处理数据接入接口。
type IngestHandler struct {
	svc *service.IngestService
}

func NewIngestHandler(svc *service.IngestService) *IngestHandler {
	return &IngestHandler{svc: svc}
}

// PostEvent 写入单条交通事件。
func (h *IngestHandler) PostEvent(c *gin.Context) {
	var in service.EventInput
	if err := c.ShouldBindJSON(&in); err != nil {
		Fail(c, http.StatusBadRequest, 40001, "invalid request body: "+err.Error())
		return
	}
	ev, err := h.svc.IngestEvent(c.Request.Context(), in)
	if err != nil {
		Fail(c, http.StatusBadRequest, 40002, err.Error())
		return
	}
	Created(c, ev)
}

// Simulate 生成模拟交通事件。
func (h *IngestHandler) Simulate(c *gin.Context) {
	var in service.SimulateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		Fail(c, http.StatusBadRequest, 40001, "invalid request body: "+err.Error())
		return
	}
	job, err := h.svc.Simulate(c.Request.Context(), in)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	Created(c, job)
}

// Import 批量导入：支持 JSON 数组与 multipart CSV 两种方式。
func (h *IngestHandler) Import(c *gin.Context) {
	ct := c.GetHeader("Content-Type")
	if isMultipart(ct) {
		h.importCSV(c)
		return
	}
	h.importJSON(c)
}

func (h *IngestHandler) importJSON(c *gin.Context) {
	var inputs []service.EventInput
	if err := c.ShouldBindJSON(&inputs); err != nil {
		Fail(c, http.StatusBadRequest, 40001, "invalid request body: "+err.Error())
		return
	}
	job, err := h.svc.ImportJSON(c.Request.Context(), "json_batch", inputs)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	Created(c, job)
}

func (h *IngestHandler) importCSV(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		Fail(c, http.StatusBadRequest, 40003, "missing multipart field 'file'")
		return
	}
	f, err := file.Open()
	if err != nil {
		Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	defer f.Close()

	job, err := h.svc.ImportCSV(c.Request.Context(), file.Filename, f)
	if err != nil {
		Fail(c, http.StatusBadRequest, 40004, err.Error())
		return
	}
	Created(c, job)
}

func isMultipart(contentType string) bool {
	return strings.HasPrefix(contentType, "multipart/form-data")
}
