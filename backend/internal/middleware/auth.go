package middleware

import (
	"strings"

	"cloudpivot/internal/common"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ContextKey 上下文键类型
type ContextKey string

const (
	UserIDKey             ContextKey = "user_id"
	UsernameKey           ContextKey = "username"
	RoleKey               ContextKey = "role"
	PermissionsKey        ContextKey = "permissions"
	TokenTypeKey          ContextKey = "token_type"
	MustChangePasswordKey ContextKey = "must_change_password"
)

// passwordChangeExemptPaths 强制改密期间仍可访问的路径（按路由 FullPath 匹配）
var passwordChangeExemptPaths = map[string]bool{
	"/api/v1/auth/change-password": true,
	"/api/v1/auth/info":            true,
	"/api/v1/auth/logout":          true,
}

// JWTAuth JWT RS256 认证中间件
// 优先读 Authorization 请求头；header 缺失时回退读 query 参数 token
// （WebSocket 握手无法携带自定义请求头，终端等 WS 路由依赖此回退）
func JWTAuth(verifyKey interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string
		if authHeader := c.GetHeader("Authorization"); authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				common.Fail(c, common.CodeUnauthorized, "认证令牌格式错误")
				c.Abort()
				return
			}
			tokenStr = parts[1]
		} else if qsToken := c.Query("token"); qsToken != "" {
			// WebSocket 回退通道
			tokenStr = qsToken
		} else {
			common.Fail(c, common.CodeUnauthorized, "未提供认证令牌")
			c.Abort()
			return
		}

		// 检查 token 是否在黑名单
		if IsTokenBlacklisted(tokenStr) {
			common.Fail(c, common.CodeUnauthorized, "令牌已失效")
			c.Abort()
			return
		}

		// 解析并验证 token
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return verifyKey, nil
		})
		if err != nil || !token.Valid {
			common.Fail(c, common.CodeUnauthorized, "令牌无效或已过期")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			common.Fail(c, common.CodeUnauthorized, "令牌声明解析失败")
			c.Abort()
			return
		}

		// 提取用户信息注入 context
		if userID, ok := claims["user_id"].(float64); ok {
			c.Set(string(UserIDKey), uint(userID))
		}
		if username, ok := claims["username"].(string); ok {
			c.Set(string(UsernameKey), username)
		}
		if role, ok := claims["role"].(string); ok {
			c.Set(string(RoleKey), role)
		}
		if tokenType, ok := claims["type"].(string); ok {
			c.Set(string(TokenTypeKey), tokenType)
		}
		if mustChange, ok := claims["must_change_password"].(bool); ok {
			c.Set(string(MustChangePasswordKey), mustChange)
		}
		// 提取权限列表
		if permsRaw, ok := claims["permissions"].([]interface{}); ok {
			perms := make([]string, 0, len(permsRaw))
			for _, p := range permsRaw {
				if ps, ok := p.(string); ok {
					perms = append(perms, ps)
				}
			}
			c.Set(string(PermissionsKey), perms)
		}

		c.Next()
	}
}

// GetUserID 从 context 获取用户 ID
func GetUserID(c *gin.Context) uint {
	if v, exists := c.Get(string(UserIDKey)); exists {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// GetUsername 从 context 获取用户名
func GetUsername(c *gin.Context) string {
	if v, exists := c.Get(string(UsernameKey)); exists {
		if name, ok := v.(string); ok {
			return name
		}
	}
	return ""
}

// GetRole 从 context 获取角色 code
func GetRole(c *gin.Context) string {
	if v, exists := c.Get(string(RoleKey)); exists {
		if role, ok := v.(string); ok {
			return role
		}
	}
	return ""
}

// GetPermissions 从 context 获取权限列表
func GetPermissions(c *gin.Context) []string {
	if v, exists := c.Get(string(PermissionsKey)); exists {
		if perms, ok := v.([]string); ok {
			return perms
		}
	}
	return nil
}

// GetTokenType 从 context 获取 token 类型
func GetTokenType(c *gin.Context) string {
	if v, exists := c.Get(string(TokenTypeKey)); exists {
		if t, ok := v.(string); ok {
			return t
		}
	}
	return ""
}

// MustChangePassword 从 context 获取是否需要强制修改密码
func MustChangePassword(c *gin.Context) bool {
	if v, exists := c.Get(string(MustChangePasswordKey)); exists {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// EnsurePasswordChanged 强制改密中间件：
// 标志位为 true 时仅放行改密/用户信息/登出三个接口，其余一律 40302 拒绝。
// 必须挂在 JWTAuth 之后使用。
func EnsurePasswordChanged() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !MustChangePassword(c) {
			c.Next()
			return
		}
		if passwordChangeExemptPaths[c.FullPath()] {
			c.Next()
			return
		}
		common.Fail(c, common.CodePasswordChangeRequired, "首次登录请先修改密码")
		c.Abort()
	}
}
