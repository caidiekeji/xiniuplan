package model

import (
	"time"
)

// Site 网站模型
type Site struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	PrimaryDomain string      `gorm:"size:255;not null;index" json:"primary_domain"`
	Domains       StringSlice `gorm:"type:text" json:"domains"`
	Type          string      `gorm:"size:32;not null" json:"type"` // static | runtime | reverse_proxy
	Runtime       string      `gorm:"size:64" json:"runtime"`
	RootDir       string      `gorm:"size:512" json:"root_dir"`
	NginxConf     string      `gorm:"size:512" json:"nginx_conf"`
	SSLEnabled    bool        `gorm:"default:false" json:"ssl_enabled"`
	SSLProvider   string      `gorm:"size:32" json:"ssl_provider"`
	SSLCertPath   string      `gorm:"size:512" json:"ssl_cert_path"`
	SSLKeyPath    string      `gorm:"size:512" json:"ssl_key_path"`
	SSLExpireAt   *time.Time  `json:"ssl_expire_at"`
	SSLAutoRenew  bool        `gorm:"default:false" json:"ssl_auto_renew"`
	Status        string      `gorm:"size:16;default:running" json:"status"` // running | stopped | error
	ContainerID   string      `gorm:"size:128" json:"container_id"`
	Remark        string      `gorm:"size:512" json:"remark"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

func (Site) TableName() string { return "sites" }
