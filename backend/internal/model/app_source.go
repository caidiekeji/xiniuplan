package model

import "time"

// AppSource 应用源模型
type AppSource struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	Name       string     `gorm:"size:128;not null" json:"name"`
	Type       string     `gorm:"size:16;not null;index" json:"type"` // official | community | private
	URL        string     `gorm:"size:512" json:"url"`                // 社区源 URL
	Path       string     `gorm:"size:512" json:"path"`               // 本地路径（官方/私有源）
	Enabled    bool       `gorm:"default:true" json:"enabled"`
	LastSyncAt *time.Time `json:"last_sync_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (AppSource) TableName() string { return "app_sources" }
