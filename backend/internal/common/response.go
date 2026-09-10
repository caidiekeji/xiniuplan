package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 错误码定义
const (
	CodeSuccess      = 200
	CodeParamError   = 40001
	CodeUnauthorized = 40101
	CodeForbidden    = 40301
	// CodePasswordChangeRequired 首次登录/密码被重置后必须先修改密码
	CodePasswordChangeRequired = 40302
	CodeNotFound               = 40401
	CodeServerError            = 50001
)

// Response 统一响应体
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PageData 分页数据
type PageData struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// Fail 失败响应
func Fail(c *gin.Context, code int, message string) {
	status := http.StatusOK
	switch code {
	case CodeParamError:
		status = http.StatusBadRequest
	case CodeUnauthorized:
		status = http.StatusUnauthorized
	case CodeForbidden:
		status = http.StatusForbidden
	case CodePasswordChangeRequired:
		status = http.StatusForbidden
	case CodeNotFound:
		status = http.StatusNotFound
	case CodeServerError:
		status = http.StatusInternalServerError
	}
	c.JSON(status, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// Page 分页响应
func Page(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	Success(c, PageData{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
