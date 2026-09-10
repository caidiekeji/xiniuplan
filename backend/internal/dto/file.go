package dto

import "time"

// FileInfo 文件/目录信息
type FileInfo struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	Mode      string    `json:"mode"`
	ModTime   time.Time `json:"mod_time"`
	IsDir     bool      `json:"is_dir"`
	Extension string    `json:"extension"`
	Path      string    `json:"path"`
}

// ListResponse 目录列表响应
type ListResponse struct {
	Path   string     `json:"path"`
	Parent string     `json:"parent"`
	Items  []FileInfo `json:"items"`
}

// CreateRequest 创建文件/目录请求
type CreateRequest struct {
	Path    string `json:"path"`
	Type    string `json:"type"` // file | dir
	Content string `json:"content"`
}

// WriteRequest 写入文件请求
type WriteRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// RenameRequest 重命名请求
type RenameRequest struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
}

// CopyRequest 复制请求
type CopyRequest struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

// MoveRequest 移动请求
type MoveRequest struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

// DeleteRequest 批量删除请求
type DeleteRequest struct {
	Paths []string `json:"paths"`
}

// CompressRequest 压缩请求
type CompressRequest struct {
	SrcPaths []string `json:"src_paths"`
	DestPath string   `json:"dest_path"`
	Format   string   `json:"format"` // zip | tar.gz
}

// DecompressRequest 解压请求
type DecompressRequest struct {
	ArchivePath string `json:"archive_path"`
	DestPath    string `json:"dest_path"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Root    string `json:"root"`
	Keyword string `json:"keyword"`
	Regex   bool   `json:"regex"`
}

// ContentResponse 读取文件内容响应
type ContentResponse struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}
