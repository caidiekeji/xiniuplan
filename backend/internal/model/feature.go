package model

import "time"

// SiteTamperBaseline 网站防篡改基线（文件哈希快照）
type SiteTamperBaseline struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	SiteID    uint      `gorm:"not null;index:idx_tamper_site" json:"site_id"`
	Path      string    `gorm:"size:512;not null;index:idx_tamper_site" json:"path"` // 站点内相对路径
	Hash      string    `gorm:"size:64;not null" json:"hash"`                        // SHA-256
	UpdatedAt time.Time `json:"updated_at"`
}

// RecycleItem 文件回收站记录（硬删除前的缓冲，7 天后自动清理）
type RecycleItem struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	OriginalPath string    `gorm:"size:512;not null" json:"original_path"` // 原虚拟路径
	RecyclePath  string    `gorm:"size:512;not null" json:"recycle_path"`  // 回收站内绝对路径
	Size         int64     `json:"size"`
	IsDir        bool      `json:"is_dir"`
	DeletedAt    time.Time `gorm:"index" json:"deleted_at"`
}
