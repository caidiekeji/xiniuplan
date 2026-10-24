package appstore

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"cloudpivot/internal/dto"
	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// 社区源缓存根目录（相对于 backend 工作目录）
const sourceCacheDir = "../runtime/app_sources"

var (
	defaultSourceService *SourceService
	sourceOnce           sync.Once
)

// GetSourceService 获取共享的 SourceService 单例
func GetSourceService(db *gorm.DB) *SourceService {
	sourceOnce.Do(func() {
		defaultSourceService = NewSourceService(db)
	})
	return defaultSourceService
}

// SourceService 应用源管理服务
type SourceService struct {
	db         *gorm.DB
	cacheDir   string
	mu         sync.RWMutex
	syncedApps map[uint][]appDefinition // 按源 ID 缓存已同步的应用列表
}

// NewSourceService 创建应用源服务
func NewSourceService(db *gorm.DB) *SourceService {
	return &SourceService{
		db:         db,
		cacheDir:   sourceCacheDir,
		syncedApps: make(map[uint][]appDefinition),
	}
}

// ListSources 列出所有源
func (s *SourceService) ListSources() ([]model.AppSource, error) {
	sources := make([]model.AppSource, 0)
	if err := s.db.Order("id ASC").Find(&sources).Error; err != nil {
		return nil, fmt.Errorf("查询应用源失败: %w", err)
	}
	return sources, nil
}

// CreateSource 创建源（official 类型只能有一个）
func (s *SourceService) CreateSource(req dto.CreateSourceRequest) (*model.AppSource, error) {
	if req.Type == "official" {
		var count int64
		s.db.Model(&model.AppSource{}).Where("type = ?", "official").Count(&count)
		if count > 0 {
			return nil, fmt.Errorf("官方源只能存在一个")
		}
	}

	// 校验必填字段
	if req.Type == "community" && req.URL == "" {
		return nil, fmt.Errorf("社区源必须填写 URL")
	}
	if (req.Type == "official" || req.Type == "private") && req.Path == "" {
		return nil, fmt.Errorf("本地源必须填写路径")
	}

	source := &model.AppSource{
		Name:    req.Name,
		Type:    req.Type,
		URL:     req.URL,
		Path:    req.Path,
		Enabled: true,
	}
	if err := s.db.Create(source).Error; err != nil {
		return nil, fmt.Errorf("创建应用源失败: %w", err)
	}
	return source, nil
}

// UpdateSource 更新源
func (s *SourceService) UpdateSource(id uint, req dto.UpdateSourceRequest) (*model.AppSource, error) {
	var source model.AppSource
	if err := s.db.First(&source, id).Error; err != nil {
		return nil, fmt.Errorf("应用源不存在")
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.URL != nil {
		updates["url"] = *req.URL
	}
	if req.Path != nil {
		updates["path"] = *req.Path
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		if err := s.db.Model(&source).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("更新应用源失败: %w", err)
		}
		// 清除该源的缓存，下次访问时重新扫描
		s.mu.Lock()
		delete(s.syncedApps, id)
		s.mu.Unlock()
	}

	if err := s.db.First(&source, id).Error; err != nil {
		return nil, fmt.Errorf("重新查询应用源失败: %w", err)
	}
	return &source, nil
}

// DeleteSource 删除源（官方源不可删）
func (s *SourceService) DeleteSource(id uint) error {
	var source model.AppSource
	if err := s.db.First(&source, id).Error; err != nil {
		return fmt.Errorf("应用源不存在")
	}
	if source.Type == "official" {
		return fmt.Errorf("官方源不可删除")
	}

	if err := s.db.Delete(&source).Error; err != nil {
		return fmt.Errorf("删除应用源失败: %w", err)
	}

	// 清除缓存
	s.mu.Lock()
	delete(s.syncedApps, id)
	s.mu.Unlock()

	// 清理社区源缓存目录
	cachePath := filepath.Join(s.cacheDir, fmt.Sprintf("%d", id))
	_ = os.RemoveAll(cachePath)

	return nil
}

// ToggleSource 启用/禁用源
func (s *SourceService) ToggleSource(id uint) (*model.AppSource, error) {
	var source model.AppSource
	if err := s.db.First(&source, id).Error; err != nil {
		return nil, fmt.Errorf("应用源不存在")
	}

	newEnabled := !source.Enabled
	if err := s.db.Model(&source).Update("enabled", newEnabled).Error; err != nil {
		return nil, fmt.Errorf("切换源状态失败: %w", err)
	}
	source.Enabled = newEnabled

	// 禁用时清除缓存
	if !newEnabled {
		s.mu.Lock()
		delete(s.syncedApps, id)
		s.mu.Unlock()
	}

	return &source, nil
}

// SyncSource 同步源
//   - official/private: 扫描本地 path 目录的 app.json
//   - community: HTTP GET url 获取应用列表 JSON，缓存到本地
func (s *SourceService) SyncSource(id uint) (*dto.SyncResult, error) {
	var source model.AppSource
	if err := s.db.First(&source, id).Error; err != nil {
		return nil, fmt.Errorf("应用源不存在")
	}

	var apps []appDefinition
	var err error

	switch source.Type {
	case "official", "private":
		apps, err = s.scanLocalDir(source.Path)
	case "community":
		apps, err = s.fetchCommunity(source)
	default:
		return nil, fmt.Errorf("未知源类型: %s", source.Type)
	}
	if err != nil {
		return nil, err
	}

	// 计算与上次同步的差异
	s.mu.Lock()
	oldApps := s.syncedApps[id]
	oldKeys := make(map[string]bool, len(oldApps))
	for _, a := range oldApps {
		oldKeys[a.Key] = true
	}
	newKeys := make(map[string]bool, len(apps))
	for _, a := range apps {
		newKeys[a.Key] = true
	}

	added, updated, removed := 0, 0, 0
	for k := range newKeys {
		if oldKeys[k] {
			updated++
		} else {
			added++
		}
	}
	for k := range oldKeys {
		if !newKeys[k] {
			removed++
		}
	}

	s.syncedApps[id] = apps
	s.mu.Unlock()

	// 更新最后同步时间
	now := time.Now()
	s.db.Model(&source).Update("last_sync_at", now)

	return &dto.SyncResult{
		Added:   added,
		Updated: updated,
		Removed: removed,
		Total:   len(apps),
	}, nil
}

// GetAllApps 从所有 enabled 源合并应用列表
func (s *SourceService) GetAllApps() ([]dto.AppInfo, error) {
	var sources []model.AppSource
	if err := s.db.Where("enabled = ?", true).Order("id ASC").Find(&sources).Error; err != nil {
		return nil, fmt.Errorf("查询启用源失败: %w", err)
	}
	allApps := make([]dto.AppInfo, 0)
	for _, source := range sources {
		apps := s.getSourceApps(source)
		for _, app := range apps {
			info := toAppInfo(&app)
			info.SourceID = source.ID
			info.SourceName = source.Name
			allApps = append(allApps, info)
		}
	}
	return allApps, nil
}

// GetAppDir 获取指定源中某个应用的本地目录路径
func (s *SourceService) GetAppDir(sourceID uint, key string) (string, error) {
	var source model.AppSource
	if err := s.db.First(&source, sourceID).Error; err != nil {
		return "", fmt.Errorf("应用源不存在")
	}

	if source.Type == "community" {
		return filepath.Join(s.cacheDir, fmt.Sprintf("%d", source.ID), key), nil
	}
	return filepath.Join(source.Path, key), nil
}

// FindApp 在所有启用源中查找应用，返回应用定义和所属源
func (s *SourceService) FindApp(key string) (*appDefinition, *model.AppSource, error) {
	apps, err := s.GetAllApps()
	if err != nil {
		return nil, nil, err
	}
	for _, app := range apps {
		if app.Key == key {
			def := appDefinition{
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
			var source model.AppSource
			s.db.First(&source, app.SourceID)
			return &def, &source, nil
		}
	}
	return nil, nil, fmt.Errorf("应用不存在: %s", key)
}

// FindAppDir 在所有启用源中查找应用目录（升级流程用）
func (s *SourceService) FindAppDir(key string) (string, error) {
	_, source, err := s.FindApp(key)
	if err != nil {
		return "", err
	}
	return s.GetAppDir(source.ID, key)
}

// getSourceApps 获取某个源的应用列表（优先读缓存，官方/私有源支持懒加载）
func (s *SourceService) getSourceApps(source model.AppSource) []appDefinition {
	s.mu.RLock()
	if apps, ok := s.syncedApps[source.ID]; ok && len(apps) > 0 {
		s.mu.RUnlock()
		return apps
	}
	s.mu.RUnlock()

	// 懒加载：官方源和私有源直接扫描本地目录
	if source.Type == "official" || source.Type == "private" {
		apps, err := s.scanLocalDir(source.Path)
		if err == nil {
			s.mu.Lock()
			s.syncedApps[source.ID] = apps
			s.mu.Unlock()
			return apps
		}
	}
	return nil
}

// scanLocalDir 扫描本地目录下的应用定义
func (s *SourceService) scanLocalDir(dir string) ([]appDefinition, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("读取源目录失败: %w", err)
	}
	apps := make([]appDefinition, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		appPath := filepath.Join(dir, entry.Name(), "app.json")
		data, err := os.ReadFile(appPath)
		if err != nil {
			continue // 跳过无 app.json 的目录
		}
		var app appDefinition
		if err := json.Unmarshal(data, &app); err != nil {
			continue // 跳过无效定义
		}
		if app.Key == "" {
			app.Key = entry.Name()
		}
		apps = append(apps, app)
	}
	return apps, nil
}

// fetchCommunity 从社区源 URL 获取应用列表并缓存到本地
func (s *SourceService) fetchCommunity(source model.AppSource) ([]appDefinition, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(source.URL)
	if err != nil {
		return nil, fmt.Errorf("请求社区源失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("社区源返回状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取社区源响应失败: %w", err)
	}

	apps := make([]appDefinition, 0)
	if err := json.Unmarshal(body, &apps); err != nil {
		// 尝试解析为 {"apps": [...]} 包装格式
		var wrapper struct {
			Apps []appDefinition `json:"apps"`
		}
		if err2 := json.Unmarshal(body, &wrapper); err2 != nil {
			return nil, fmt.Errorf("解析应用列表 JSON 失败: %w", err)
		}
		apps = wrapper.Apps
	}

	// 缓存到本地目录
	cachePath := filepath.Join(s.cacheDir, fmt.Sprintf("%d", source.ID))
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return nil, fmt.Errorf("创建缓存目录失败: %w", err)
	}
	for _, app := range apps {
		if app.Key == "" {
			continue
		}
		appDir := filepath.Join(cachePath, app.Key)
		if err := os.MkdirAll(appDir, 0755); err != nil {
			continue
		}
		data, _ := json.MarshalIndent(app, "", "  ")
		if err := os.WriteFile(filepath.Join(appDir, "app.json"), data, 0644); err != nil {
			slog.Warn("[appstore] failed to cache app", "key", app.Key, "error", err)
		}
	}

	return apps, nil
}
