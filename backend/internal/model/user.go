package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	Username           string     `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash       string     `gorm:"size:255;not null" json:"-"`
	Email              string     `gorm:"size:128;index" json:"email"`      // 用于忘记密码邮件重置
	Role               string     `gorm:"size:32;default:user" json:"role"` // 角色 code，兼容旧字段
	RoleID             uint       `gorm:"default:0" json:"role_id"`         // 关联 roles.id
	TOTPSecret         string     `gorm:"size:128" json:"-"`
	Status             string     `gorm:"size:16;default:active" json:"status"` // active | disabled
	MustChangePassword bool       `gorm:"default:false" json:"must_change_password"`
	LastLoginAt        *time.Time `json:"last_login_at"`
	CreatedAt          time.Time  `json:"created_at"`
}

func (User) TableName() string { return "users" }

// PasswordResetToken 密码重置令牌（忘记密码邮件流程）
type PasswordResetToken struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Username  string     `gorm:"size:64;index;not null" json:"username"`
	TokenHash string     `gorm:"size:64;uniqueIndex;not null" json:"-"` // SHA-256 hex，不存明文
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}

func (PasswordResetToken) TableName() string { return "password_reset_tokens" }
