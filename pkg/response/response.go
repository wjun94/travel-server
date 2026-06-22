// Package response 提供统一的 JSON 响应格式
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 标准响应结构
type Response struct {
	Code int         `json:"code"` // 0 表示成功，其他表示错误码
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: "ok", Data: data})
}

// Fail 返回错误响应
func Fail(c *gin.Context, httpCode int, msg string) {
	c.JSON(httpCode, Response{Code: httpCode, Msg: msg})
}
