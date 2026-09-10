package model

import "time"

// DatabaseInstance 数据库实例模型
type DatabaseInstance struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Type         string    `gorm:"size:32;not null" json:"type"` // mysql | postgres | redis
	Version      string    `gorm:"size:32" json:"version"`
	Host         string    `gorm:"size:255;default:127.0.0.1" json:"host"`
	Port         int       `json:"port"`
	Username     string    `gorm:"size:64;default:root" json:"username"`
	ContainerID  string    `gorm:"size:128" json:"container_id"`
	RootPassword string    `gorm:"size:255" json:"-"`
	IsExternal   bool      `gorm:"default:false" json:"is_external"` // 外部导入的实例（非面板创建的Docker）
	Status       string    `gorm:"size:16;default:running" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func (DatabaseInstance) TableName() string { return "database_instances" }
