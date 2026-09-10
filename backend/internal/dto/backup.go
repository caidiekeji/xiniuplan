package dto

// CreateBackupRequest 创建备份请求
type CreateBackupRequest struct {
	Type   string `json:"type" binding:"required,oneof=site database system"`
	Target string `json:"target"` // 站点域名 / 数据库名
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
	CreatedAt     string `json:"created_at"`
}

// RestoreRequest 恢复备份请求
type RestoreRequest struct {
	BackupID uint `json:"backup_id" binding:"required"`
}
