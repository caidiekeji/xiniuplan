package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/model"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth 校验 X-API-Key 请求头（格式：keyID.secret）
// 校验通过则注入用户上下文并放行；否则返回 false 由调用方继续 JWT 流程
func APIKeyAuth(c *gin.Context) bool {
	header := c.GetHeader("X-API-Key")
	if header == "" {
		return false
	}
	parts := strings.SplitN(header, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		common.Fail(c, common.CodeUnauthorized, "API Key 格式错误（应为 keyID.secret）")
		c.Abort()
		return true
	}
	keyID, secret := parts[0], parts[1]

	var apiKey model.APIKey
	if err := model.DB.Where("key_id = ?", keyID).First(&apiKey).Error; err != nil {
		common.Fail(c, common.CodeUnauthorized, "API Key 无效")
		c.Abort()
		return true
	}
	if apiKey.Status != "active" {
		common.Fail(c, common.CodeUnauthorized, "API Key 已被禁用")
		c.Abort()
		return true
	}
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		common.Fail(c, common.CodeUnauthorized, "API Key 已过期")
		c.Abort()
		return true
	}
	sum := sha256.Sum256([]byte(secret))
	if hex.EncodeToString(sum[:]) != apiKey.KeyHash {
		common.Fail(c, common.CodeUnauthorized, "API Key 无效")
		c.Abort()
		return true
	}

	// 加载绑定用户
	var user model.User
	if err := model.DB.First(&user, apiKey.UserID).Error; err != nil || user.Status != "active" {
		common.Fail(c, common.CodeUnauthorized, "API Key 绑定的用户不可用")
		c.Abort()
		return true
	}

	// 注入上下文（与 JWT 相同的键）
	c.Set(string(UserIDKey), user.ID)
	c.Set(string(UsernameKey), user.Username)
	c.Set(string(RoleKey), user.Role)
	c.Set(string(TokenTypeKey), "api_key")

	// 权限：key 自带权限优先，否则继承用户角色权限
	if perms := splitPerms(apiKey.Permissions); len(perms) > 0 {
		c.Set(string(PermissionsKey), perms)
	} else if perms := loadRolePermissions(&user); len(perms) > 0 {
		c.Set(string(PermissionsKey), perms)
	}

	// 更新最近使用时间（每分钟最多一次，降低写放大）
	now := time.Now()
	if apiKey.LastUsedAt == nil || now.Sub(*apiKey.LastUsedAt) > time.Minute {
		model.DB.Model(&apiKey).Update("last_used_at", now)
	}

	c.Next()
	return true
}

func splitPerms(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func loadRolePermissions(user *model.User) []string {
	var role model.Role
	if user.RoleID > 0 {
		if err := model.DB.First(&role, user.RoleID).Error; err == nil {
			return role.Permissions
		}
	} else if err := model.DB.Where("code = ?", user.Role).First(&role).Error; err == nil {
		return role.Permissions
	}
	return []string{}
}
