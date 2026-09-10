package system

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"gorm.io/gorm"
)

const (
	panelVersion = "1.0.0"
	buildTime    = "2026-01-01T00:00:00Z"
)

// Service 系统设置服务
type Service struct {
	db       *gorm.DB
	executor *common.CommandExecutor
}

// NewService 创建系统设置服务
func NewService(db *gorm.DB) *Service {
	executor := common.NewCommandExecutor(60 * time.Second)
	// 注册 docker prune 白名单命令
	executor.Register(&common.CommandTemplate{
		Name: "docker_system_prune",
		Bin:  "docker",
		Args: []string{"system", "prune", "-f"},
	})
	return &Service{
		db:       db,
		executor: executor,
	}
}

// GetSettings 读取所有设置为 map
func (s *Service) GetSettings() (map[string]string, error) {
	var settings []model.Setting
	if err := s.db.Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("query settings: %w", err)
	}

	result := make(map[string]string, len(settings))
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}

// UpdateSettings 批量更新设置
func (s *Service) UpdateSettings(settings map[string]string) error {
	if len(settings) == 0 {
		return fmt.Errorf("no settings to update")
	}

	for key, value := range settings {
		var setting model.Setting
		err := s.db.First(&setting, "`key` = ?", key).Error
		if err == gorm.ErrRecordNotFound {
			// 不存在则创建
			setting = model.Setting{
				Key:       key,
				Value:     value,
				UpdatedAt: time.Now(),
			}
			if err := s.db.Create(&setting).Error; err != nil {
				return fmt.Errorf("create setting %s: %w", key, err)
			}
		} else if err != nil {
			return fmt.Errorf("query setting %s: %w", key, err)
		} else {
			// 存在则更新
			setting.Value = value
			setting.UpdatedAt = time.Now()
			if err := s.db.Save(&setting).Error; err != nil {
				return fmt.Errorf("update setting %s: %w", key, err)
			}
		}
	}

	// 安全入口变更后立即失效中间件缓存，保证即时生效
	if _, ok := settings["security_entry"]; ok {
		middleware.InvalidateSecurityEntryCache()
	}
	return nil
}

// GetVersion 返回面板版本信息
func (s *Service) GetVersion() (*dto.VersionResponse, error) {
	var uptime uint64
	if h, err := host.Info(); err == nil {
		uptime = h.Uptime
	}

	return &dto.VersionResponse{
		PanelVersion: panelVersion,
		GoVersion:    runtime.Version(),
		BuildTime:    buildTime,
		Uptime:       uptime,
	}, nil
}

// CleanSystem 清理系统
// cleanType: cache | docker | logs
func (s *Service) CleanSystem(cleanType string) (string, error) {
	switch cleanType {
	case "cache":
		return s.cleanCache()
	case "docker":
		return s.cleanDocker()
	case "logs":
		return s.cleanLogs()
	default:
		return "", fmt.Errorf("unsupported clean type: %s", cleanType)
	}
}

// cleanCache 清理 /tmp 和系统缓存
func (s *Service) cleanCache() (string, error) {
	cleaned := int64(0)

	// 清理 /tmp 目录下超过 7 天的文件
	tmpDir := os.TempDir()
	filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		// 只清理超过 7 天的文件
		if time.Since(info.ModTime()) > 7*24*time.Hour {
			size := info.Size()
			if os.Remove(path) == nil {
				cleaned += size
			}
		}
		return nil
	})

	return fmt.Sprintf("已清理临时文件，释放约 %d 字节", cleaned), nil
}

// cleanDocker 执行 docker system prune -f
func (s *Service) cleanDocker() (string, error) {
	result, err := s.executor.Execute(context.Background(), "docker_system_prune", nil)
	if err != nil {
		return "", fmt.Errorf("docker prune failed: %w (当前环境可能未安装 docker)", err)
	}
	if result.ExitCode != 0 {
		return "", fmt.Errorf("docker prune exit code %d: %s", result.ExitCode, result.Stderr)
	}
	return result.Stdout, nil
}

// cleanLogs 清理旧日志文件
func (s *Service) cleanLogs() (string, error) {
	cleaned := int64(0)
	logDirs := []string{
		"/var/log",
		"./runtime/logs",
		"./logs",
	}

	for _, dir := range logDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			// 只清理 .log 文件且超过 30 天
			ext := filepath.Ext(path)
			if (ext == ".log" || ext == ".gz") && time.Since(info.ModTime()) > 30*24*time.Hour {
				size := info.Size()
				if os.Remove(path) == nil {
					cleaned += size
				}
			}
			return nil
		})
	}

	return fmt.Sprintf("已清理旧日志文件，释放约 %d 字节", cleaned), nil
}

// GetSystemInfo 获取详细系统信息
func (s *Service) GetSystemInfo() (*dto.SystemInfoResponse, error) {
	info := &dto.SystemInfoResponse{
		PanelVersion: panelVersion,
	}

	// 主机信息
	if h, err := host.Info(); err == nil {
		info.Hostname = h.Hostname
		info.OS = h.OS + " " + h.PlatformVersion
		info.Kernel = h.KernelVersion
	}

	// CPU 信息
	if cpuInfo, err := cpu.Info(); err == nil && len(cpuInfo) > 0 {
		info.CPUModel = cpuInfo[0].ModelName
	}
	if cores, err := cpu.Counts(false); err == nil {
		info.CPUCores = cores
	}

	// 内存信息
	if vm, err := mem.VirtualMemory(); err == nil {
		info.MemoryTotal = vm.Total
	}

	// 磁盘总数（所有挂载点总和）
	if parts, err := disk.Partitions(false); err == nil {
		var total uint64
		for _, part := range parts {
			if usage, err := disk.Usage(part.Mountpoint); err == nil {
				total += usage.Total
			}
		}
		info.DiskTotal = total
	}

	// 公网 IP（复用 monitor 中的获取方式）
	info.PublicIP = getPublicIP()

	return info, nil
}

// getPublicIP 尝试获取公网 IP（超时3秒）
func getPublicIP() string {
	// 避免循环依赖，直接内联实现
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://api.ipify.org?format=text")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

// GetSmtpConfig 读取 SMTP 配置
func (s *Service) GetSmtpConfig() (map[string]string, error) {
	settings, err := s.GetSettings()
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for k, v := range settings {
		if len(k) > 5 && k[:5] == "smtp_" {
			result[k] = v
		}
	}
	return result, nil
}

// UpdateSmtpConfig 更新 SMTP 配置
func (s *Service) UpdateSmtpConfig(config map[string]string) error {
	return s.UpdateSettings(config)
}

// SendTestEmail 发送测试邮件（占位实现，返回成功）
func (s *Service) SendTestEmail(to string) error {
	// TODO: 实际 SMTP 发送实现
	return nil
}

// CheckUpdate 检查版本更新（占位实现，返回当前版本为最新）
func (s *Service) CheckUpdate() (*dto.VersionResponse, error) {
	return &dto.VersionResponse{
		PanelVersion:  panelVersion,
		GoVersion:     runtime.Version(),
		BuildTime:     buildTime,
		LatestVersion: panelVersion,
		HasUpdate:     false,
	}, nil
}
