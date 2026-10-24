package dto

// CreateBackupRequest 创建备份请求
type CreateBackupRequest struct {
	Type   string `json:"type" binding:"required,oneof=site database system app"`
	Target string `json:"target"` // 站点域名 / 数据库名 / 应用 key
	Remark string `json:"remark"`
}

// BackupResponse 备份响应
type BackupResponse struct {
	ID            uint   `json:"id"`
	Type          string `json:"type"`
	Target        string `json:"target"`
	FilePath      string `json:"file_path"`
	FileName      string `json:"file_name"`
	Size          int64  `json:"size"`
	FormattedSize string `json:"formatted_size"`
	Status        string `json:"status"`
	Remark        string `json:"remark"`
	ErrorMsg      string `json:"error_msg"`
	RemoteID      uint   `json:"remote_id"`
	RemoteStatus  string `json:"remote_status"`
	RemoteMsg     string `json:"remote_msg"`
	CreatedAt     string `json:"created_at"`
}

// RestoreRequest 恢复备份请求
type RestoreRequest struct {
	BackupID uint `json:"backup_id" binding:"required"`
}

// PushBackupRequest 推送备份到远程存储
type PushBackupRequest struct {
	RemoteID uint `json:"remote_id" binding:"required"`
}

// RemoteStorageConfigRequest 远程存储创建/更新请求
type RemoteStorageConfigRequest struct {
	Name    string            `json:"name"`
	Type    string            `json:"type" binding:"omitempty,oneof=s3 webdav sftp panel"`
	Config  map[string]string `json:"config"`
	Enabled *bool             `json:"enabled"`
}

// RemoteStorageResponse 远程存储响应（配置脱敏）
type RemoteStorageResponse struct {
	ID        uint              `json:"id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Config    map[string]string `json:"config"`
	Enabled   bool              `json:"enabled"`
	CreatedAt string            `json:"created_at"`
}
