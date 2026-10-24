package file

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"cloudpivot/internal/dto"
)

const maxReadSize = 10 * 1024 * 1024 // 10MB

// Manager 文件管理器，基于虚拟根目录操作
type Manager struct {
	root string // 虚拟根目录的绝对路径
}

// NewManager 创建文件管理器
// Windows 开发环境使用 ./sandbox 作为虚拟根；Linux 生产环境可通过参数指定
func NewManager(root string) (*Manager, error) {
	if root == "" {
		root = "./sandbox"
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve root path: %w", err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("create root dir: %w", err)
	}
	return &Manager{root: abs}, nil
}

// Root 返回虚拟根目录绝对路径
func (m *Manager) Root() string {
	return m.root
}

// safePath 解析并校验路径，禁止 ../ 穿越，返回绝对路径
func (m *Manager) safePath(rel string) (string, error) {
	// 规范化：去除前导 /，统一使用 filepath.Clean
	rel = strings.TrimPrefix(rel, "/")
	rel = strings.TrimPrefix(rel, "\\")
	if rel == "" || rel == "." {
		return m.root, nil
	}

	// 显式拒绝包含 .. 的路径段
	for _, seg := range strings.Split(filepath.ToSlash(rel), "/") {
		if seg == ".." {
			return "", fmt.Errorf("path traversal not allowed: %s", rel)
		}
	}

	abs := filepath.Join(m.root, rel)
	abs = filepath.Clean(abs)

	// 二次校验：解析后必须仍在 root 下
	relToRoot, err := filepath.Rel(m.root, abs)
	if err != nil {
		return "", fmt.Errorf("path relation error: %w", err)
	}
	if strings.HasPrefix(relToRoot, "..") {
		return "", fmt.Errorf("path escapes root directory: %s", rel)
	}
	return abs, nil
}

// toVirtual 将绝对路径转换为虚拟路径（相对于 root）
func (m *Manager) toVirtual(abs string) string {
	rel, err := filepath.Rel(m.root, abs)
	if err != nil {
		return abs
	}
	if rel == "." {
		return "/"
	}
	return "/" + filepath.ToSlash(rel)
}

// fileInfoToDTO 将 os.FileInfo 转换为 dto.FileInfo
func (m *Manager) fileInfoToDTO(abs string, info os.FileInfo) dto.FileInfo {
	return dto.FileInfo{
		Name:      info.Name(),
		Size:      info.Size(),
		Mode:      info.Mode().String(),
		ModTime:   info.ModTime(),
		IsDir:     info.IsDir(),
		Extension: strings.TrimPrefix(filepath.Ext(info.Name()), "."),
		Path:      m.toVirtual(abs),
	}
}

// ListDir 列出目录内容
func (m *Manager) ListDir(path string) (*dto.ListResponse, error) {
	abs, err := m.safePath(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("stat path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", path)
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	items := make([]dto.FileInfo, 0, len(entries))
	for _, entry := range entries {
		fi, err := entry.Info()
		if err != nil {
			continue
		}
		items = append(items, m.fileInfoToDTO(filepath.Join(abs, entry.Name()), fi))
	}

	// 计算父目录虚拟路径
	parent := ""
	vpath := m.toVirtual(abs)
	if vpath != "/" {
		parent = filepath.ToSlash(filepath.Dir(vpath))
		if parent == "." {
			parent = "/"
		}
		if !strings.HasPrefix(parent, "/") {
			parent = "/" + parent
		}
	}

	return &dto.ListResponse{
		Path:   vpath,
		Parent: parent,
		Items:  items,
	}, nil
}

// CreateFile 创建文件并写入内容
func (m *Manager) CreateFile(path, content string) (*dto.FileInfo, error) {
	abs, err := m.safePath(path)
	if err != nil {
		return nil, err
	}

	// 确保父目录存在
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return nil, fmt.Errorf("create parent dir: %w", err)
	}

	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	fi := m.fileInfoToDTO(abs, info)
	return &fi, nil
}

// CreateDir 创建目录（递归）
func (m *Manager) CreateDir(path string) (*dto.FileInfo, error) {
	abs, err := m.safePath(path)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	fi := m.fileInfoToDTO(abs, info)
	return &fi, nil
}

// ReadFile 读取文件内容（限制 10MB）
func (m *Manager) ReadFile(path string) (string, error) {
	abs, err := m.safePath(path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("stat file: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory: %s", path)
	}
	if info.Size() > maxReadSize {
		return "", fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), maxReadSize)
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(data), nil
}

// WriteFile 写入文件内容（覆盖）
func (m *Manager) WriteFile(path, content string) error {
	abs, err := m.safePath(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory: %s", path)
	}

	if err := os.WriteFile(abs, []byte(content), info.Mode()); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

// Rename 重命名文件或目录
func (m *Manager) Rename(oldPath, newPath string) error {
	absOld, err := m.safePath(oldPath)
	if err != nil {
		return err
	}
	absNew, err := m.safePath(newPath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(absOld); err != nil {
		return fmt.Errorf("source not found: %w", err)
	}
	// 确保目标父目录存在
	if err := os.MkdirAll(filepath.Dir(absNew), 0o755); err != nil {
		return fmt.Errorf("create target parent dir: %w", err)
	}

	if err := os.Rename(absOld, absNew); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

// Copy 复制文件或目录（递归）
func (m *Manager) Copy(src, dst string) error {
	absSrc, err := m.safePath(src)
	if err != nil {
		return err
	}
	absDst, err := m.safePath(dst)
	if err != nil {
		return err
	}

	srcInfo, err := os.Stat(absSrc)
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	if srcInfo.IsDir() {
		return m.copyDir(absSrc, absDst)
	}
	return m.copyFile(absSrc, absDst)
}

func (m *Manager) copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dest: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy content: %w", err)
	}

	// 保留权限
	if srcInfo, err := os.Stat(src); err == nil {
		_ = os.Chmod(dst, srcInfo.Mode())
	}
	return nil
}

func (m *Manager) copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return m.copyFile(path, target)
	})
}

// Move 移动文件或目录
func (m *Manager) Move(src, dst string) error {
	absSrc, err := m.safePath(src)
	if err != nil {
		return err
	}
	absDst, err := m.safePath(dst)
	if err != nil {
		return err
	}

	if _, err := os.Stat(absSrc); err != nil {
		return fmt.Errorf("source not found: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(absDst), 0o755); err != nil {
		return fmt.Errorf("create target parent dir: %w", err)
	}

	if err := os.Rename(absSrc, absDst); err != nil {
		// 跨设备移动失败时回退到复制+删除
		if copyErr := m.Copy(src, dst); copyErr != nil {
			return fmt.Errorf("move (copy fallback): %w", copyErr)
		}
		if delErr := os.RemoveAll(absSrc); delErr != nil {
			return fmt.Errorf("move (remove source): %w", delErr)
		}
	}
	return nil
}

// Delete 批量删除，目录递归删除
// Delete 删除文件/目录——安全版：移入回收站（可恢复），永久删除走 PurgeRecycle
func (m *Manager) Delete(paths []string) error {
	return m.ToRecycle(paths)
}

// DeleteForever 永久删除（跳过回收站）
func (m *Manager) DeleteForever(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("no paths provided")
	}
	for _, p := range paths {
		abs, err := m.safePath(p)
		if err != nil {
			return err
		}
		if abs == m.root || abs == filepath.Join(m.root, recycleDirName) {
			return fmt.Errorf("cannot delete this directory")
		}
		if err := os.RemoveAll(abs); err != nil {
			return fmt.Errorf("delete %s: %w", p, err)
		}
	}
	return nil
}

// Compress 压缩多个源路径到目标归档文件
// format: zip | tar.gz
func (m *Manager) Compress(srcPaths []string, destPath, format string) error {
	if len(srcPaths) == 0 {
		return fmt.Errorf("no source paths provided")
	}

	absDest, err := m.safePath(destPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(absDest), 0o755); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}

	switch format {
	case "zip":
		return m.compressZip(srcPaths, absDest)
	case "tar.gz", "tgz":
		return m.compressTarGz(srcPaths, absDest)
	default:
		return fmt.Errorf("unsupported format: %s (use zip or tar.gz)", format)
	}
}

func (m *Manager) compressZip(srcPaths []string, dest string) error {
	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	defer zw.Close()

	for _, sp := range srcPaths {
		abs, err := m.safePath(sp)
		if err != nil {
			return err
		}
		base := filepath.Dir(abs)
		if err := filepath.Walk(abs, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(base, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			header.Name = rel
			if info.IsDir() {
				header.Name += "/"
			} else {
				header.Method = zip.Deflate
			}

			writer, err := zw.CreateHeader(header)
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(writer, f)
			return err
		}); err != nil {
			return fmt.Errorf("add %s to zip: %w", sp, err)
		}
	}
	return nil
}

func (m *Manager) compressTarGz(srcPaths []string, dest string) error {
	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, sp := range srcPaths {
		abs, err := m.safePath(sp)
		if err != nil {
			return err
		}
		base := filepath.Dir(abs)
		if err := filepath.Walk(abs, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(base, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)

			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = rel

			if err := tw.WriteHeader(header); err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(tw, f)
			return err
		}); err != nil {
			return fmt.Errorf("add %s to tar.gz: %w", sp, err)
		}
	}
	return nil
}

// Decompress 根据扩展名自动识别格式并解压
func (m *Manager) Decompress(archivePath, destPath string) error {
	absArchive, err := m.safePath(archivePath)
	if err != nil {
		return err
	}
	absDest, err := m.safePath(destPath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(absArchive); err != nil {
		return fmt.Errorf("archive not found: %w", err)
	}
	if err := os.MkdirAll(absDest, 0o755); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(absArchive))
	name := strings.ToLower(filepath.Base(absArchive))

	switch {
	case ext == ".zip":
		return m.decompressZip(absArchive, absDest)
	case strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz") || ext == ".gz":
		return m.decompressTarGz(absArchive, absDest)
	default:
		return fmt.Errorf("unsupported archive format: %s", archivePath)
	}
}

func (m *Manager) decompressZip(archive, dest string) error {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		target := filepath.Join(dest, f.Name)
		// 防穿越
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) && target != filepath.Clean(dest) {
			return fmt.Errorf("zip entry escapes dest: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, f.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(target)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return err
		}
		_ = os.Chmod(target, f.Mode())
	}
	return nil
}

func (m *Manager) decompressTarGz(archive, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}

		target := filepath.Join(dest, header.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) && target != filepath.Clean(dest) {
			return fmt.Errorf("tar entry escapes dest: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.Create(target)
			if err != nil {
				return err
			}
			_, err = io.Copy(out, tr)
			out.Close()
			if err != nil {
				return err
			}
			_ = os.Chmod(target, os.FileMode(header.Mode))
		}
	}
	return nil
}

// Search 按文件名搜索（支持正则）
func (m *Manager) Search(rootPath, keyword string, useRegex bool) ([]dto.FileInfo, error) {
	absRoot, err := m.safePath(rootPath)
	if err != nil {
		return nil, err
	}

	var matcher func(string) bool
	if useRegex {
		re, err := regexp.Compile(keyword)
		if err != nil {
			return nil, fmt.Errorf("invalid regex: %w", err)
		}
		matcher = func(name string) bool { return re.MatchString(name) }
	} else {
		matcher = func(name string) bool { return strings.Contains(strings.ToLower(name), strings.ToLower(keyword)) }
	}

	results := make([]dto.FileInfo, 0)
	err = filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 跳过无权限的文件
		}
		if matcher(info.Name()) {
			results = append(results, m.fileInfoToDTO(path, info))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk: %w", err)
	}
	if results == nil {
		results = []dto.FileInfo{}
	}
	return results, nil
}

// Stat 获取单个文件/目录信息
func (m *Manager) Stat(path string) (*dto.FileInfo, error) {
	abs, err := m.safePath(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}
	fi := m.fileInfoToDTO(abs, info)
	return &fi, nil
}
