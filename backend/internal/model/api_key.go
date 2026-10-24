package model

import "time"

// APIKey 程序化调用面板接口的密钥（U-007）
type APIKey struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"size:64;not null" json:"name"`
	KeyID       string     `gorm:"size:32;uniqueIndex;not null" json:"key_id"` // 明文 key id，随请求头传递
	KeyHash     string     `gorm:"size:64;uniqueIndex;not null" json:"-"`      // SHA-256(secret)，不存明文 secret
	Secret      string     `gorm:"-" json:"secret,omitempty"`                  // 仅创建时返回一次
	UserID      uint       `gorm:"index;not null" json:"user_id"`              // 绑定用户，权限随用户
	Permissions string     `gorm:"type:text" json:"permissions"`               // 逗号分隔权限点，空=继承用户角色
	Status      string     `gorm:"size:16;default:active" json:"status"`       // active | disabled
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (APIKey) TableName() string { return "api_keys" }
