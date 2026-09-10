package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders 安全响应头中间件
// 设置 X-Content-Type-Options、X-Frame-Options、Referrer-Policy 等安全头
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止 MIME 类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")
		// 防止点击劫持（iframe 嵌套）
		c.Header("X-Frame-Options", "DENY")
		// 限制 Referer 信息泄露
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// 基础 CSP（允许同源脚本和样式，禁止 inline 脚本）
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' ws: wss:; frame-ancestors 'none'")
		// 禁止浏览器缓存敏感页面（API 响应）
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.Header("Pragma", "no-cache")
		// XSS 保护（旧浏览器兼容）
		c.Header("X-XSS-Protection", "1; mode=block")

		c.Next()
	}
}
