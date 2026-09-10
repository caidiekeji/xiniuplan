package model

import "time"

// AuditLog 审计日志模型
type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Action    string    `gorm:"size:64;not null" json:"action"`
	Resource  string    `gorm:"size:255" json:"resource"`
	IP        string    `gorm:"size:64" json:"ip"`
	UserAgent string    `gorm:"size:512" json:"user_agent"`
	Result    string    `gorm:"size:16;default:success" json:"result"` // success | fail
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_logs" }
