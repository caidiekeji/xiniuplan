package middleware

import (
	"time"

	"cloudpivot/internal/model"

	"github.com/gin-gonic/gin"
)

// Audit 操作审计中间件，记录请求到 audit_logs 表
func Audit() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 只记录写操作和需要审计的操作
		method := c.Request.Method
		path := c.Request.URL.Path

		// 跳过 GET 请求的审计（除登录等关键操作），减少日志量
		if method == "GET" && path != "/api/v1/auth/login" {
			c.Next()
			return
		}

		c.Next()

		// 记录审计日志
		userID := GetUserID(c)
		result := "success"
		if c.Writer.Status() >= 400 {
			result = "fail"
		}

		log := model.AuditLog{
			UserID:    userID,
			Action:    method + " " + path,
			Resource:  path,
			IP:        c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			Result:    result,
			CreatedAt: start,
		}

		if model.DB != nil {
			_ = model.DB.Create(&log).Error
		}
	}
}
