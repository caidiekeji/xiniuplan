package model

import "time"

// Setting 系统设置模型
type Setting struct {
	Key         string    `gorm:"primaryKey;size:128" json:"key"`
	Value       string    `gorm:"size:1024" json:"value"`
	Description string    `gorm:"size:255" json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Setting) TableName() string { return "settings" }

// DefaultSettings 默认设置项
var DefaultSettings = []Setting{
	{Key: "panel_port", Value: "8888", Description: "面板端口"},
	{Key: "panel_domain", Value: "", Description: "面板域名"},
	{Key: "language", Value: "zh-CN", Description: "语言"},
	{Key: "theme", Value: "dark", Description: "主题"},
	{Key: "security_entry", Value: "", Description: "安全入口"},
	// 会话超时（分钟）：JWT 签发时以该值为准，缺省 120 分钟（与
	// config.yaml 的 jwt.access_token_ttl=7200s 对齐）
	{Key: "session_timeout", Value: "120", Description: "会话超时（分钟）"},
	{Key: "smtp_host", Value: "", Description: "SMTP 主机"},
	{Key: "smtp_port", Value: "587", Description: "SMTP 端口"},
	{Key: "smtp_user", Value: "", Description: "SMTP 用户名"},
	{Key: "smtp_pass", Value: "", Description: "SMTP 密码"},
	{Key: "smtp_from", Value: "", Description: "SMTP 发件人"},
}
