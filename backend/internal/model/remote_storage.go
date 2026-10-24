package model

import "time"

// RemoteStorage 远程备份存储配置
// type: s3 | webdav | sftp | panel（另一台云枢面板实例）
type RemoteStorage struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"size:64;not null" json:"name"`
	Type string `gorm:"size:16;not null" json:"type"` // s3 | webdav | sftp | panel
	// 通用连接配置（JSON，敏感字段 base64 编码存储，API 输出脱敏）
	// s3: endpoint,region,bucket,access_key,secret_key,prefix,path_style
	// webdav: url,username,password,prefix
	// sftp: host,port,username,password/private_key,remote_dir
	// panel: url,token,remote_dir
	Config    string    `gorm:"type:text" json:"-"`
	Enabled   bool      `gorm:"default:false" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (RemoteStorage) TableName() string { return "remote_storages" }
