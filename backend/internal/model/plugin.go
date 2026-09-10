package model

import "time"

// PluginConfig 插件配置模型，持久化每个插件的启用状态与 JSON 配置。
type PluginConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:128;not null" json:"name"`
	Enabled   bool      `gorm:"default:false" json:"enabled"`
	Config    string    `gorm:"type:text" json:"config"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (PluginConfig) TableName() string { return "plugin_configs" }
