package appstore

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloudpivot/internal/dto"
	"cloudpivot/internal/model"
	dockerpkg "cloudpivot/pkg/docker"

	"gorm.io/gorm"
)

// 默认路径（相对于 backend 工作目录，指向项目根目录的 store）
const (
	StoreAppsDir   = "../store/apps"
	RuntimeAppsDir = "../runtime/apps"
)

// Service 应用商店服务
type Service struct {
	db         *gorm.DB
	storeDir   string
	runtimeDir string
	sourceSvc  *SourceService
}

// NewService 创建应用商店服务
func NewService(db *gorm.DB) *Service {
	return &Service{
		db:         db,
		storeDir:   StoreAppsDir,
		runtimeDir: RuntimeAppsDir,
		sourceSvc:  GetSourceService(db),
	}
}

// appDefinition 应用定义文件结构（对应 app.json）
type appDefinition struct {
	Key           string         `json:"key"`
	Name          string         `json:"name"`
	Version       string         `json:"version"`
	Description   string         `json:"description"`
	Icon          string         `json:"icon"`
	Category      string         `json:"category"`
	Tags          []string       `json:"tags"`
	RequiredPorts []int          `json:"required_ports"`
	EnvParams     []dto.EnvParam `json:"env_params"`
	ComposeFile   string         `json:"compose_file"`
	MinMemoryMB   int            `json:"min_memory_mb"`
	MinDiskGB     int            `json:"min_disk_gb"`
}

// ListApps 从所有启用源合并应用列表，支持分类、关键词和源筛选
func (s *Service) ListApps(category, keyword, sourceID string) ([]dto.AppInfo, error) {
	apps, err := s.sourceSvc.GetAllApps()
	if err != nil {
		return nil, fmt.Errorf("获取应用列表失败: %w", err)
	}

	filtered := make([]dto.AppInfo, 0)
	for _, app := range apps {
		// 按源筛选
		if sourceID != "" {
			var sid uint
			if _, err := fmt.Sscanf(sourceID, "%d", &sid); err == nil {
				if app.SourceID != sid {
					continue
				}
			}
		}

		// 分类过滤
		if category != "" && !strings.EqualFold(app.Category, category) {
			continue
		}

		// 关键词搜索（名称、描述、标签、key）
		if keyword != "" {
			kw := strings.ToLower(keyword)
			matched := strings.Contains(strings.ToLower(app.Name), kw) ||
				strings.Contains(strings.ToLower(app.Description), kw) ||
				strings.Contains(strings.ToLower(app.Key), kw)
			if !matched {
				for _, tag := range app.Tags {
					if strings.Contains(strings.ToLower(tag), kw) {
						matched = true
						break
					}
				}
			}
			if !matched {
				continue
			}
		}

		filtered = append(filtered, app)
	}
	return filtered, nil
}

// GetApp 获取应用详情（从所有启用源中查找）
func (s *Service) GetApp(key string) (*dto.AppInfo, error) {
	def, source, err := s.sourceSvc.FindApp(key)
	if err != nil {
		return nil, err
	}
	info := toAppInfo(def)
	info.SourceID = source.ID
	info.SourceName = source.Name
	return &info, nil
}

// loadAppDefinition 从指定目录加载应用定义
func loadAppDefinition(dir, key string) (*appDefinition, error) {
	appPath := filepath.Join(dir, key, "app.json")
	data, err := os.ReadFile(appPath)
	if err != nil {
		return nil, fmt.Errorf("read app.json: %w", err)
	}

	var app appDefinition
	if err := json.Unmarshal(data, &app); err != nil {
		return nil, fmt.Errorf("parse app.json: %w", err)
	}
	if app.Key == "" {
		app.Key = key
	}
	return &app, nil
}

// toAppInfo 转换为 DTO
func toAppInfo(app *appDefinition) dto.AppInfo {
	return dto.AppInfo{
		Key:           app.Key,
		Name:          app.Name,
		Version:       app.Version,
		Description:   app.Description,
		Icon:          app.Icon,
		Category:      app.Category,
		Tags:          app.Tags,
		RequiredPorts: app.RequiredPorts,
		EnvParams:     app.EnvParams,
		ComposeFile:   app.ComposeFile,
		MinMemoryMB:   app.MinMemoryMB,
		MinDiskGB:     app.MinDiskGB,
	}
}

// InstallApp 安装应用（根据 source_id 找到对应源的应用目录）
func (s *Service) InstallApp(key string, sourceID uint, envParams map[string]string) (*model.InstalledApp, error) {
	// 1. 检查应用是否已安装
	var existing model.InstalledApp
	if err := s.db.Where("key = ?", key).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("应用 %s 已安装", key)
	}

	// 2. 根据 source_id 找到应用目录
	appDir, err := s.sourceSvc.GetAppDir(sourceID, key)
	if err != nil {
		return nil, fmt.Errorf("获取应用目录失败: %w", err)
	}

	// 3. 加载应用定义
	appDef, err := loadAppDefinition(filepath.Dir(appDir), key)
	if err != nil {
		return nil, fmt.Errorf("加载应用定义失败: %w", err)
	}

	// 3. 校验必填环境变量
	for _, p := range appDef.EnvParams {
		if p.Required {
			if val, ok := envParams[p.Key]; !ok || val == "" {
				if p.Default != "" {
					envParams[p.Key] = p.Default
				} else {
					return nil, fmt.Errorf("缺少必填环境变量: %s (%s)", p.Key, p.Label)
				}
			}
		}
	}

	// 4. 复制应用目录到 runtime/apps/{key}/
	installPath := filepath.Join(s.runtimeDir, key)
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return nil, fmt.Errorf("create install dir: %w", err)
	}

	srcDir := appDir
	if err := copyDir(srcDir, installPath); err != nil {
		return nil, fmt.Errorf("copy app files: %w", err)
	}

	// 5. 渲染 docker-compose.yml（替换环境变量占位符）
	composePath := filepath.Join(installPath, "docker-compose.yml")
	if err := renderComposeFile(composePath, envParams); err != nil {
		return nil, fmt.Errorf("render compose file: %w", err)
	}

	// 6. 执行 docker compose up -d
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	status := "installed"
	if err := dockerpkg.Up(ctx, installPath); err != nil {
		// compose 启动失败不回滚文件，标记为 error 状态
		status = "error"
	} else {
		status = "running"
	}

	// 7. 记录安装状态到 DB
	envParamsJSON, _ := json.Marshal(envParams)
	installed := &model.InstalledApp{
		Key:         key,
		Name:        appDef.Name,
		Version:     appDef.Version,
		InstallPath: installPath,
		Status:      status,
		EnvParams:   string(envParamsJSON),
	}
	if err := s.db.Create(installed).Error; err != nil {
		return nil, fmt.Errorf("save install record: %w", err)
	}

	if status == "error" {
		return installed, fmt.Errorf("应用安装完成但容器启动失败，请检查日志")
	}
	return installed, nil
}

// UninstallApp 卸载应用
func (s *Service) UninstallApp(key string) error {
	// 1. 查找已安装记录
	var installed model.InstalledApp
	if err := s.db.Where("key = ?", key).First(&installed).Error; err != nil {
		return fmt.Errorf("应用 %s 未安装", key)
	}

	// 2. 执行 docker compose down
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_ = dockerpkg.Down(ctx, installed.InstallPath) // down 失败不阻断卸载

	// 3. 清理安装目录
	if installed.InstallPath != "" {
		_ = os.RemoveAll(installed.InstallPath)
	}

	// 4. 删除 DB 记录
	if err := s.db.Delete(&installed).Error; err != nil {
		return fmt.Errorf("delete install record: %w", err)
	}
	return nil
}

// ListInstalled 列出已安装应用
func (s *Service) ListInstalled() ([]dto.InstalledAppResponse, error) {
	apps := make([]model.InstalledApp, 0)
	if err := s.db.Order("created_at DESC").Find(&apps).Error; err != nil {
		return nil, fmt.Errorf("query installed apps: %w", err)
	}

	responses := make([]dto.InstalledAppResponse, 0, len(apps))
	for _, app := range apps {
		containerStatus := s.getContainerStatus(app.InstallPath)
		responses = append(responses, toInstalledResponse(&app, containerStatus))
	}
	return responses, nil
}

// GetInstallStatus 检查应用安装状态
func (s *Service) GetInstallStatus(key string) (*dto.AppStatusResponse, error) {
	var installed model.InstalledApp
	if err := s.db.Where("key = ?", key).First(&installed).Error; err != nil {
		return nil, fmt.Errorf("应用 %s 未安装", key)
	}

	containerStatus := s.getContainerStatus(installed.InstallPath)
	return &dto.AppStatusResponse{
		Key:             installed.Key,
		Name:            installed.Name,
		Status:          installed.Status,
		ContainerStatus: containerStatus,
	}, nil
}

// getContainerStatus 通过 compose ps 检查容器运行状态
func (s *Service) getContainerStatus(installPath string) string {
	if installPath == "" {
		return "unknown"
	}
	if _, err := os.Stat(filepath.Join(installPath, "docker-compose.yml")); err != nil {
		return "not_found"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	output, err := dockerpkg.Ps(ctx, installPath)
	if err != nil {
		return "unknown"
	}

	// 简单解析：检查是否有 running 状态的容器
	if strings.Contains(output, "\"running\"") || strings.Contains(output, "running") {
		if strings.Contains(output, "exited") || strings.Contains(output, "\"exited\"") {
			return "partial"
		}
		return "running"
	}
	if strings.Contains(output, "exited") || strings.Contains(output, "\"exited\"") {
		return "stopped"
	}
	return "unknown"
}

// toInstalledResponse 转换为响应 DTO
func toInstalledResponse(app *model.InstalledApp, containerStatus string) dto.InstalledAppResponse {
	var envParams map[string]string
	if app.EnvParams != "" {
		_ = json.Unmarshal([]byte(app.EnvParams), &envParams)
	}
	if envParams == nil {
		envParams = make(map[string]string)
	}
	return dto.InstalledAppResponse{
		ID:              app.ID,
		Key:             app.Key,
		Name:            app.Name,
		Version:         app.Version,
		InstallPath:     app.InstallPath,
		Status:          app.Status,
		EnvParams:       envParams,
		ContainerStatus: containerStatus,
		CreatedAt:       app.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       app.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// copyDir 复制目录
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}
		return copyFile(path, targetPath)
	})
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

// renderComposeFile 渲染 docker-compose.yml，替换 {{ENV_VAR}} 占位符
func renderComposeFile(path string, envParams map[string]string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(data)
	for key, val := range envParams {
		placeholder := "{{" + key + "}}"
		content = strings.ReplaceAll(content, placeholder, val)
	}

	return os.WriteFile(path, []byte(content), 0644)
}
