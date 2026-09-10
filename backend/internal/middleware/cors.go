package middleware

import (
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
// 生产环境（release 模式）下仅允许同源请求；开发环境（debug 模式）下允许所有来源
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if gin.Mode() == gin.ReleaseMode {
			// 生产环境：仅允许同源（Origin 为空或与请求 Host 一致）
			if origin != "" {
				originURL, err := url.Parse(origin)
				if err != nil || originURL.Host != c.Request.Host {
					// 非同源请求，不设置 CORS 头，浏览器将拦截
					c.AbortWithStatus(403)
					return
				}
			}
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Disposition")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		} else {
			// 开发环境：允许所有来源（反射 Origin）
			if origin == "" {
				origin = "*"
			}
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Disposition")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isSameOrigin 检查 origin 是否与请求 host 同源
func isSameOrigin(origin, host string) bool {
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Host, host)
}
