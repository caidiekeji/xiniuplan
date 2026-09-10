package model

import "time"

// CronTask 计划任务模型
type CronTask struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Name      string     `gorm:"size:128;not null" json:"name"`
	Type      string     `gorm:"size:32;not null" json:"type"` // shell | backup | url | app_restart
	Schedule  string     `gorm:"size:64;not null" json:"schedule"`
	Command   string     `gorm:"type:text" json:"command"`
	Target    string     `gorm:"size:255" json:"target"`
	Remark    string     `gorm:"size:512" json:"remark"`
	Status    string     `gorm:"size:16;default:enabled" json:"status"` // enabled | disabled
	EntryID   int        `gorm:"default:0" json:"entry_id"`
	LastRunAt *time.Time `json:"last_run_at"`
	NextRunAt *time.Time `json:"next_run_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (CronTask) TableName() string { return "cron_tasks" }

// CronLog 计划任务执行日志模型
type CronLog struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	TaskID    uint       `gorm:"index;not null" json:"task_id"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	Output    string     `gorm:"type:text" json:"output"`
	Result    string     `gorm:"size:16;default:running" json:"result"` // running | success | failed
	CreatedAt time.Time  `json:"created_at"`
}

func (CronLog) TableName() string { return "cron_logs" }
