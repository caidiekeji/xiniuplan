package dto

import "time"

// CreateSourceRequest 创建应用源请求
type CreateSourceRequest struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required,oneof=official community private"`
	URL  string `json:"url"`
	Path string `json:"path"`
}

// UpdateSourceRequest 更新应用源请求（指针字段支持部分更新）
type UpdateSourceRequest struct {
	Name    *string `json:"name"`
	URL     *string `json:"url"`
	Path    *string `json:"path"`
	Enabled *bool   `json:"enabled"`
}

// SourceResponse 应用源响应
type SourceResponse struct {
	ID         uint       `json:"id"`
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	URL        string     `json:"url"`
	Path       string     `json:"path"`
	Enabled    bool       `json:"enabled"`
	LastSyncAt *time.Time `json:"last_sync_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// SyncResult 同步结果
type SyncResult struct {
	Added   int `json:"added"`
	Updated int `json:"updated"`
	Removed int `json:"removed"`
	Total   int `json:"total"`
}
