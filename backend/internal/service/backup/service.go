package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

const (
	backupRootDir = "./backup"
)

// Service 备份管理服务
type Service struct {
	db       *gorm.DB
	executor *common.CommandExecutor
}

// NewService 创建备份管理服务
func NewService(db *gorm.DB) *Service {
	exec := common.NewCommandExecutor(300 * time.Second)
	registerBackupTemplates(exec)
	return &Service{
		db:       db,
		executor: exec,
	}
}

// registerBackupTemplates 注册备份相关命令白名单模板
func registerBackupTemplates(exec *common.CommandExecutor) {
	// mysqldump 备份（密码通过 MYSQL_PWD 环境变量传递，避免出现在进程参数中）
	exec.Register(&common.CommandTemplate{
		Name:        "mysqldump",
		Description: "MySQL database dump",
		Bin:         "mysqldump",
		Args: []string{
			"-h", "{{host}}",
			"-P", "{{port}}",
			"-u", "{{user}}",
			"--single-transaction",
			"--routines",
			"--triggers",
			"--result-file={{outfile}}",
			"{{db}}",
		},
		AllowedArgs: map[string]bool{
			"host": true, "port": true, "user": true,
			"outfile": true, "db": true,
		},
	})
	// mysql 恢复（密码通过 MYSQL_PWD 环境变量传递）
	exec.Register(&common.CommandTemplate{
		Name:        "mysql_restore",
		Description: "MySQL restore from stdin",
		Bin:         "mysql",
		Args: []string{
			"-h", "{{host}}",
			"-P", "{{port}}",
			"-u", "{{user}}",
			"{{db}}",
		},
		AllowedArgs: map[string]bool{
			"host": true, "port": true, "user": true, "db": true,
		},
	})
}

// --- 路径辅助 ---

func backupDir(backupType string) string {
	return filepath.Join(backupRootDir, backupType)
}

func timestamp() string {
	return time.Now().Format("20060102_150405")
}

// ensureBackupDir 确保备份目录存在
func ensureBackupDir(backupType string) (string, error) {
	dir := backupDir(backupType)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create backup dir: %w", err)
	}
	return dir, nil
}

// --- CRUD ---

// CreateBackup 创建备份任务，大文件异步执行，立即返回任务记录
func (s *Service) CreateBackup(req dto.CreateBackupRequest) (*model.Backup, error) {
	// 校验类型
	switch req.Type {
	case "site", "database", "system":
	default:
		return nil, fmt.Errorf("unsupported backup type: %s", req.Type)
	}

	// 创建备份记录
	backup := &model.Backup{
		Type:   req.Type,
		Target: req.Target,
		Status: "pending",
		Remark: req.Remark,
	}
	if err := s.db.Create(backup).Error; err != nil {
		return nil, fmt.Errorf("create backup record: %w", err)
	}

	// 异步执行备份
	go s.executeBackup(backup.ID)

	return backup, nil
}

// executeBackup 异步执行备份
func (s *Service) executeBackup(backupID uint) {
	backup, err := s.GetBackup(backupID)
	if err != nil {
		return
	}

	// 更新状态为 running
	backup.Status = "running"
	s.db.Save(backup)

	var filePath string
	var execErr error

	switch backup.Type {
	case "site":
		filePath, execErr = s.backupSite(backup.Target)
	case "database":
		filePath, execErr = s.backupDatabase(backup.Target)
	case "system":
		filePath, execErr = s.backupSystem()
	default:
		execErr = fmt.Errorf("unknown backup type: %s", backup.Type)
	}

	if execErr != nil {
		backup.Status = "failed"
		backup.ErrorMsg = execErr.Error()
		s.db.Save(backup)
		return
	}

	// 获取文件大小
	var size int64
	if info, err := os.Stat(filePath); err == nil {
		size = info.Size()
	}

	backup.FilePath = filePath
	backup.Size = size
	backup.Status = "success"
	backup.ErrorMsg = ""
	s.db.Save(backup)
}

// backupSite 打包网站根目录
func (s *Service) backupSite(domain string) (string, error) {
	if domain == "" {
		return "", errors.New("site domain is required")
	}

	// 网站根目录
	siteRoot := siteRootDir(domain)
	if _, err := os.Stat(siteRoot); os.IsNotExist(err) {
		return "", fmt.Errorf("site root not found: %s", siteRoot)
	}

	dir, err := ensureBackupDir("site")
	if err != nil {
		return "", err
	}

	outFile := filepath.Join(dir, fmt.Sprintf("%s_%s.tar.gz", sanitizeName(domain), timestamp()))
	if err := tarGzDir(siteRoot, outFile); err != nil {
		return "", fmt.Errorf("tar site: %w", err)
	}

	return outFile, nil
}

// backupDatabase 执行 mysqldump
func (s *Service) backupDatabase(dbName string) (string, error) {
	if dbName == "" {
		return "", errors.New("database name is required")
	}

	dir, err := ensureBackupDir("database")
	if err != nil {
		return "", err
	}

	outFile := filepath.Join(dir, fmt.Sprintf("%s_%s.sql", sanitizeName(dbName), timestamp()))

	// 获取 MySQL 连接信息（优先从已注册实例获取）
	host, port, user, password := s.getMySQLConnInfo()

	params := map[string]string{
		"host":    host,
		"port":    fmt.Sprintf("%d", port),
		"user":    user,
		"outfile": outFile,
		"db":      dbName,
	}
	// 密码通过环境变量传递，避免出现在进程参数列表中
	env := map[string]string{}
	if password != "" {
		env["MYSQL_PWD"] = password
	}

	result, err := s.executor.ExecuteWithEnv(context.Background(), "mysqldump", params, env)
	if err != nil {
		return "", fmt.Errorf("mysqldump failed: %w", err)
	}
	if result.ExitCode != 0 {
		return "", fmt.Errorf("mysqldump exit code %d: %s", result.ExitCode, result.Stderr)
	}

	return outFile, nil
}

// backupSystem 打包配置文件
func (s *Service) backupSystem() (string, error) {
	dir, err := ensureBackupDir("system")
	if err != nil {
		return "", err
	}

	outFile := filepath.Join(dir, fmt.Sprintf("config_%s.tar.gz", timestamp()))

	// 收集需要备份的配置文件/目录
	configPaths := []string{
		"config.yaml",
		"./runtime/nginx/conf.d",
		"./data",
	}

	var existingPaths []string
	for _, p := range configPaths {
		if _, err := os.Stat(p); err == nil {
			existingPaths = append(existingPaths, p)
		}
	}

	if len(existingPaths) == 0 {
		// 创建空 tar.gz
		if err := createEmptyTarGz(outFile); err != nil {
			return "", err
		}
		return outFile, nil
	}

	if err := tarGzPaths(existingPaths, outFile); err != nil {
		return "", fmt.Errorf("tar system config: %w", err)
	}

	return outFile, nil
}

// getMySQLConnInfo 从数据库实例获取 MySQL 连接信息，无实例则用默认值
func (s *Service) getMySQLConnInfo() (host string, port int, user, password string) {
	host = "127.0.0.1"
	port = 3306
	user = "root"
	password = ""

	var instance model.DatabaseInstance
	if err := s.db.Where("type = ?", "mysql").Order("id ASC").First(&instance).Error; err == nil {
		if instance.Port > 0 {
			port = instance.Port
		}
		if instance.RootPassword != "" {
			password = instance.RootPassword
		}
	}
	return
}

// ListBackups 分页查询备份列表
func (s *Service) ListBackups(page, pageSize int, backupType string) ([]model.Backup, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	query := s.db.Model(&model.Backup{})
	if backupType != "" {
		query = query.Where("type = ?", backupType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count backups: %w", err)
	}

	backups := make([]model.Backup, 0)
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&backups).Error; err != nil {
		return nil, 0, fmt.Errorf("query backups: %w", err)
	}

	return backups, total, nil
}

// GetBackup 获取备份详情
func (s *Service) GetBackup(id uint) (*model.Backup, error) {
	var backup model.Backup
	if err := s.db.First(&backup, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("backup not found")
		}
		return nil, fmt.Errorf("query backup: %w", err)
	}
	return &backup, nil
}

// DeleteBackup 删除备份记录和本地文件
func (s *Service) DeleteBackup(id uint) error {
	backup, err := s.GetBackup(id)
	if err != nil {
		return err
	}

	// 删除本地文件
	if backup.FilePath != "" {
		os.Remove(backup.FilePath)
	}

	// 删除数据库记录
	if err := s.db.Delete(backup).Error; err != nil {
		return fmt.Errorf("delete backup from db: %w", err)
	}

	return nil
}

// RestoreBackup 恢复备份，恢复前自动创建当前状态快照
func (s *Service) RestoreBackup(id uint) error {
	backup, err := s.GetBackup(id)
	if err != nil {
		return err
	}

	if backup.Status != "success" {
		return fmt.Errorf("backup is not ready, current status: %s", backup.Status)
	}

	if _, err := os.Stat(backup.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backup.FilePath)
	}

	// 恢复前自动创建当前状态快照
	if err := s.createPreRestoreSnapshot(backup); err != nil {
		fmt.Printf("[backup-service] pre-restore snapshot failed (non-fatal): %v\n", err)
	}

	switch backup.Type {
	case "site":
		return s.restoreSite(backup)
	case "database":
		return s.restoreDatabase(backup)
	case "system":
		return s.restoreSystem(backup)
	default:
		return fmt.Errorf("unsupported restore type: %s", backup.Type)
	}
}

// createPreRestoreSnapshot 恢复前创建当前状态快照
func (s *Service) createPreRestoreSnapshot(backup *model.Backup) error {
	snapshot := &model.Backup{
		Type:   backup.Type,
		Target: backup.Target,
		Status: "pending",
		Remark: fmt.Sprintf("pre-restore snapshot for backup #%d", backup.ID),
	}
	if err := s.db.Create(snapshot).Error; err != nil {
		return err
	}
	go s.executeBackup(snapshot.ID)
	return nil
}

// restoreSite 解压网站备份到网站目录
func (s *Service) restoreSite(backup *model.Backup) error {
	siteRoot := siteRootDir(backup.Target)

	// 清空现有目录（先备份式重命名）
	if _, err := os.Stat(siteRoot); err == nil {
		tmpDir := siteRoot + ".restore_tmp"
		os.RemoveAll(tmpDir)
		if err := os.Rename(siteRoot, tmpDir); err != nil {
			return fmt.Errorf("rename existing site dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)
	}

	if err := os.MkdirAll(siteRoot, 0755); err != nil {
		return fmt.Errorf("create site root: %w", err)
	}

	if err := untarGz(backup.FilePath, siteRoot); err != nil {
		return fmt.Errorf("extract site backup: %w", err)
	}

	return nil
}

// restoreDatabase 执行 mysql < dump.sql（通过 MYSQL_PWD 环境变量传递密码）
func (s *Service) restoreDatabase(backup *model.Backup) error {
	host, port, user, password := s.getMySQLConnInfo()

	args := []string{
		"-h", host,
		"-P", fmt.Sprintf("%d", port),
		"-u", user,
		backup.Target,
	}

	cmd := exec.Command("mysql", args...)
	// 密码通过环境变量传递，避免出现在进程参数中
	if password != "" {
		cmd.Env = append(os.Environ(), "MYSQL_PWD="+password)
	}
	sqlFile, err := os.Open(backup.FilePath)
	if err != nil {
		return fmt.Errorf("open dump file: %w", err)
	}
	defer sqlFile.Close()
	cmd.Stdin = sqlFile

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysql restore failed: %w, stderr: %s", err, stderr.String())
	}

	return nil
}

// restoreSystem 解压系统配置
func (s *Service) restoreSystem(backup *model.Backup) error {
	// 解压到临时目录，避免直接覆盖
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("cloudpivot_restore_%d", time.Now().Unix()))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := untarGz(backup.FilePath, tmpDir); err != nil {
		return fmt.Errorf("extract system backup: %w", err)
	}

	// 将解压出的配置文件复制回原位置
	// config.yaml
	srcConfig := filepath.Join(tmpDir, "config.yaml")
	if _, err := os.Stat(srcConfig); err == nil {
		if err := copyFile(srcConfig, "config.yaml"); err != nil {
			fmt.Printf("[backup-service] restore config.yaml failed: %v\n", err)
		}
	}

	// runtime/nginx/conf.d
	srcNginx := filepath.Join(tmpDir, "runtime", "nginx", "conf.d")
	if _, err := os.Stat(srcNginx); err == nil {
		dstNginx := "./runtime/nginx/conf.d"
		os.MkdirAll(dstNginx, 0755)
		copyDir(srcNginx, dstNginx)
	}

	return nil
}

// DownloadBackup 返回备份文件路径供流式下载
func (s *Service) DownloadBackup(id uint) (string, error) {
	backup, err := s.GetBackup(id)
	if err != nil {
		return "", err
	}

	if backup.Status != "success" {
		return "", fmt.Errorf("backup is not ready, current status: %s", backup.Status)
	}

	if backup.FilePath == "" {
		return "", errors.New("backup file path is empty")
	}

	if _, err := os.Stat(backup.FilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("backup file not found: %s", backup.FilePath)
	}

	return backup.FilePath, nil
}

// CleanupOldBackups 自动清理超过保留份数的旧备份
func (s *Service) CleanupOldBackups(keepCount int) error {
	if keepCount <= 0 {
		keepCount = 10
	}

	// 按类型分组清理
	types := []string{"site", "database", "system"}
	for _, t := range types {
		var backups []model.Backup
		if err := s.db.Where("type = ? AND status = ?", t, "success").
			Order("created_at DESC").Find(&backups).Error; err != nil {
			return fmt.Errorf("query backups for cleanup: %w", err)
		}

		if len(backups) <= keepCount {
			continue
		}

		// 删除超出保留份数的旧备份
		for i := keepCount; i < len(backups); i++ {
			b := backups[i]
			if b.FilePath != "" {
				os.Remove(b.FilePath)
			}
			s.db.Delete(&b)
		}
	}

	return nil
}

// --- 工具函数 ---

// siteRootDir 网站根目录（与 site service 保持一致）
func siteRootDir(domain string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join("./runtime/www", domain)
	}
	return filepath.Join("/www", domain)
}

// sanitizeName 清理文件名中的不安全字符
func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")
	return name
}

// tarGzDir 将目录打包为 tar.gz
func tarGzDir(srcDir, outFile string) error {
	return tarGzPaths([]string{srcDir}, outFile)
}

// tarGzPaths 将多个路径打包为 tar.gz
func tarGzPaths(paths []string, outFile string) error {
	out, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()

	gzWriter := gzip.NewWriter(out)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return err
		}
		baseDir := filepath.Dir(abs)

		err = filepath.Walk(abs, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(baseDir, path)
			if err != nil {
				return err
			}
			// 使用正斜杠，确保跨平台一致性
			relPath = filepath.ToSlash(relPath)

			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = relPath

			if err := tarWriter.WriteHeader(header); err != nil {
				return err
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			return err
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// createEmptyTarGz 创建空的 tar.gz 文件
func createEmptyTarGz(outFile string) error {
	out, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()

	gzWriter := gzip.NewWriter(out)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	return nil
}

// untarGz 解压 tar.gz 到目标目录
func untarGz(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// 安全检查：禁止路径穿越
		target := filepath.Join(destDir, header.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)+string(os.PathSeparator)) &&
			filepath.Clean(target) != filepath.Clean(destDir) {
			return fmt.Errorf("tar entry path traversal: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// copyDir 复制目录
func copyDir(src, dst string) error {
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
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}

// ToResponse 将 model.Backup 转换为 dto.BackupResponse
func ToResponse(b *model.Backup) dto.BackupResponse {
	return dto.BackupResponse{
		ID:            b.ID,
		Type:          b.Type,
		Target:        b.Target,
		FilePath:      b.FilePath,
		FileName:      filepath.Base(b.FilePath),
		Size:          b.Size,
		FormattedSize: formatSize(b.Size),
		Status:        b.Status,
		Remark:        b.Remark,
		ErrorMsg:      b.ErrorMsg,
		CreatedAt:     b.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// formatSize 格式化文件大小
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
