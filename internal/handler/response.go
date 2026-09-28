package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一 API 返回结构（对齐 PRD 第 9 节"API 返回结构统一"）。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// OK 返回成功（200）。
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "ok", Data: data})
}

// Created 返回创建成功（201）。
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Code: 0, Message: "ok", Data: data})
}

// Fail 返回错误，code 为业务错误码。
func Fail(c *gin.Context, status, code int, msg string) {
	c.JSON(status, Response{Code: code, Message: msg})
}
