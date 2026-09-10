package middleware

import (
	"cloudpivot/internal/common"

	"github.com/gin-gonic/gin"
)

// RequirePermission RBAC 权限校验中间件
// 从 JWT context 读取 permissions，超级管理员（role=super_admin 或 permissions 含 *）直接放行
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetRole(c)
		// 超级管理员直接放行
		if role == "super_admin" {
			c.Next()
			return
		}

		perms := GetPermissions(c)
		for _, p := range perms {
			if p == "*" || p == permission {
				c.Next()
				return
			}
		}

		common.Fail(c, common.CodeForbidden, "无权限执行此操作")
		c.Abort()
	}
}
