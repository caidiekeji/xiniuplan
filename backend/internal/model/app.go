package model

import "time"

// InstalledApp 已安装应用模型
type InstalledApp struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Key         string    `gorm:"size:64;not null;uniqueIndex" json:"key"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Version     string    `gorm:"size:32" json:"version"`
	InstallPath string    `gorm:"size:512" json:"install_path"`
	Status      string    `gorm:"size:16;default:installed" json:"status"` // installed | running | stopped | error
	EnvParams   string    `gorm:"type:text" json:"env_params"`             // JSON 存储
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (InstalledApp) TableName() string { return "installed_apps" }
