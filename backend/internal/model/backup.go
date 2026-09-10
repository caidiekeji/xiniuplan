package model

import "time"

// Backup 备份模型
type Backup struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Type      string    `gorm:"size:32;not null" json:"type"` // site | database | app | system
	Target    string    `gorm:"size:255" json:"target"`
	FilePath  string    `gorm:"size:512" json:"file_path"`
	Size      int64     `json:"size"`
	Status    string    `gorm:"size:16;default:pending" json:"status"` // pending | running | success | failed
	Remark    string    `gorm:"size:512" json:"remark"`
	ErrorMsg  string    `gorm:"type:text" json:"error_msg"`
	CreatedAt time.Time `json:"created_at"`
}

func (Backup) TableName() string { return "backups" }
