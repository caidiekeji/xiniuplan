package model

import "time"

// Node 纳管的远程节点（roadmap #13 多节点管理）
// 节点为另一台部署了本面板的服务器：主控端凭节点签发的 API Key 拉取只读概览。
type Node struct {
	ID       uint   `gorm:"primarykey" json:"id"`
	Name     string `gorm:"size:64;not null" json:"name"`
	BaseURL  string `gorm:"size:255;not null;uniqueIndex" json:"base_url"` // 如 https://192.168.1.20:8888
	APIKey   string `gorm:"size:255;not null" json:"api_key"`              // 节点面板签发的 API Key（keyID.secret）
	Hostname string `gorm:"size:128" json:"hostname"`
	Version  string `gorm:"size:32" json:"version"`
	Enabled  bool   `gorm:"not null;default:true" json:"enabled"`

	LastStatus  string     `gorm:"size:16" json:"last_status"` // online | offline
	CPUPercent  float64    `json:"cpu_percent"`
	MemPercent  float64    `json:"mem_percent"`
	DiskPercent float64    `json:"disk_percent"`
	LastError   string     `gorm:"size:256" json:"last_error"`
	LastSeenAt  *time.Time `json:"last_seen_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
