package middleware

import (
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin"
)

// tokenInQueryRe 匹配 URL query 中的 token 参数（WS 认证回退使用）
var tokenInQueryRe = regexp.MustCompile(`([?&])token=[^&\s]+`)

// SanitizedLogger 安全日志中间件：基于 gin.Logger，但对 query 中的 token 脱敏。
// 背景：浏览器 WebSocket 无法自定义请求头，认证 token 只能走 query 回退，
// 若按默认 gin.Logger 原样打印 path，会把有效 token 泄漏到访问日志/日志采集系统。
func SanitizedLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		if param.Path != "" {
			param.Path = tokenInQueryRe.ReplaceAllString(param.Path, "${1}token=***")
		}
		return fmt.Sprintf("[GIN] %3d | %13v | %15s | %-7s %s\n",
			param.StatusCode, param.Latency, param.ClientIP, param.Method, param.Path)
	})
}
