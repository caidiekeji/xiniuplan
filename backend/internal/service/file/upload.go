package file

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"cloudpivot/internal/dto"
)

const maxUploadSize = 512 * 1024 * 1024 // 512MB

// UploadFile 解析 multipart 请求，保存文件到目标目录
func (m *Manager) UploadFile(r *http.Request, destDir string) (*dto.FileInfo, error) {
	// 限制上传大小
	r.Body = http.MaxBytesReader(nil, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, fmt.Errorf("parse multipart: %w", err)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("get file from form: %w", err)
	}
	defer file.Close()

	absDir, err := m.safePath(destDir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return nil, fmt.Errorf("create dest dir: %w", err)
	}

	// 使用上传文件名，禁止路径穿越
	filename := filepath.Base(header.Filename)
	if filename == "" || filename == "." || filename == ".." {
		return nil, fmt.Errorf("invalid filename")
	}
	absPath := filepath.Join(absDir, filename)

	// 二次校验目标路径在 root 内
	relToRoot, err := filepath.Rel(m.root, absPath)
	if err != nil || (len(relToRoot) >= 2 && relToRoot[:2] == "..") {
		return nil, fmt.Errorf("invalid upload destination")
	}

	out, err := os.Create(absPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return nil, fmt.Errorf("save file: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}
	fi := m.fileInfoToDTO(absPath, info)
	return &fi, nil
}

// DownloadFile 流式写入文件到 ResponseWriter，设置 Content-Disposition
func (m *Manager) DownloadFile(w http.ResponseWriter, path string) error {
	abs, err := m.safePath(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("cannot download a directory: %s", path)
	}

	f, err := os.Open(abs)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", info.Name()))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("stream file: %w", err)
	}
	return nil
}
