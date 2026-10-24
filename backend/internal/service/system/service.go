package system

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
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
	updExec  *common.CommandExecutor // 系统更新专用：长超时
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
	// 系统更新命令需要远超 60s 的执行窗口（SY-005）
	updExec := common.NewCommandExecutor(30 * time.Minute)
	registerUpdateTemplates(updExec)
	return &Service{
		db:       db,
		executor: executor,
		updExec:  updExec,
	}
}

// GetSettings 读取所有设置为 map
//
// 补齐缺失的默认设置项：旧库（如线上库）是在 DefaultSettings 增加
// session_timeout 之前初始化的，缺这个键会让「会话超时」在界面上显示为 0，
// 也让 auth service 读不到值而回退到配置文件（2026-10-23 闭环审计）。
func (s *Service) GetSettings() (map[string]string, error) {
	var settings []model.Setting
	if err := s.db.Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("query settings: %w", err)
	}

	result := make(map[string]string, len(settings))
	for _, s := range settings {
		result[s.Key] = s.Value
	}

	// 缺失的默认项用默认值补上（下次保存设置时随整包落库）
	for _, def := range model.DefaultSettings {
		if _, ok := result[def.Key]; !ok {
			result[def.Key] = def.Value
		}
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

// getPublicIP 获取公网 IP（走公共缓存：后台定时刷新，接口路径零阻塞）
func getPublicIP() string {
	return common.GetPublicIP()
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

// SendTestEmail 发送测试邮件（真实 SMTP 发送，读取系统设置中的 smtp_* 配置）
func (s *Service) SendTestEmail(to string) error {
	if to == "" {
		return fmt.Errorf("收件人不能为空")
	}
	cfg, err := s.GetSmtpConfig()
	if err != nil {
		return fmt.Errorf("读取 SMTP 配置失败: %w", err)
	}
	host := cfg["smtp_host"]
	user := cfg["smtp_user"]
	pass := cfg["smtp_pass"]
	if host == "" || user == "" {
		return fmt.Errorf("SMTP 配置不完整，请先填写服务器地址和账号")
	}
	port := cfg["smtp_port"]
	if port == "" {
		port = "587"
	}
	from := cfg["smtp_from"]
	if from == "" {
		from = user
	}
	addr := host + ":" + port
	subject := "[云枢面板] SMTP 测试邮件"
	body := "这是一封来自云枢面板的测试邮件。\n\n如果您收到了这封邮件，说明 SMTP 配置正确，告警通知和密码重置邮件均可正常发送。\n\n时间: " + time.Now().Format("2006-01-02 15:04:05")
	msg := buildPlainEmail(from, []string{to}, subject, body)
	auth := smtp.PlainAuth("", user, pass, host)
	if err := smtp.SendMail(addr, auth, from, []string{to}, msg); err != nil {
		return fmt.Errorf("发送失败: %w", err)
	}
	return nil
}

// buildPlainEmail 构造纯文本邮件报文
func buildPlainEmail(from string, to []string, subject, body string) []byte {
	var msg strings.Builder
	msg.WriteString("From: " + from + "\r\n")
	msg.WriteString("To: " + strings.Join(to, ", ") + "\r\n")
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)
	return []byte(msg.String())
}

// CheckUpdate 检查面板新版本
// 优先读取设置 update_check_url（JSON: {"version":"x.y.z"}），
// 未配置时回退到 GitHub Latest Release API（update_github_repo 设置，如 "cloudpivot/cloudpivot"）。
// 任一来源不可达时优雅降级：返回 hasUpdate=false 并在 latest_version 中说明。
func (s *Service) CheckUpdate() (*dto.VersionResponse, error) {
	resp := &dto.VersionResponse{
		PanelVersion:  panelVersion,
		GoVersion:     runtime.Version(),
		BuildTime:     buildTime,
		LatestVersion: panelVersion,
		HasUpdate:     false,
	}

	settings, _ := s.GetSettings()
	checkURL := settings["update_check_url"]
	if checkURL == "" {
		if repo := settings["update_github_repo"]; repo != "" {
			checkURL = "https://api.github.com/repos/" + repo + "/releases/latest"
		}
	}
	if checkURL == "" {
		resp.LatestVersion = "未配置更新源（设置 update_check_url 或 update_github_repo）"
		return resp, nil
	}

	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequest(http.MethodGet, checkURL, nil)
	if err != nil {
		resp.LatestVersion = "检查失败: " + err.Error()
		return resp, nil
	}
	req.Header.Set("Accept", "application/json")
	if ghToken := settings["update_github_token"]; ghToken != "" && strings.Contains(checkURL, "api.github.com") {
		req.Header.Set("Authorization", "Bearer "+ghToken)
	}
	httpResp, err := client.Do(req)
	if err != nil {
		resp.LatestVersion = "检查失败: " + err.Error()
		return resp, nil
	}
	defer httpResp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(httpResp.Body, 1<<20))
	if err != nil || httpResp.StatusCode != http.StatusOK {
		resp.LatestVersion = fmt.Sprintf("检查失败: HTTP %d", httpResp.StatusCode)
		return resp, nil
	}

	// 兼容两种格式：{"version":"1.0.1"} 与 GitHub release {"tag_name":"v1.0.1"}
	var payload struct {
		Version string `json:"version"`
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		resp.LatestVersion = "检查失败: 响应格式错误"
		return resp, nil
	}
	latest := payload.Version
	if latest == "" {
		latest = strings.TrimPrefix(payload.TagName, "v")
	}
	if latest == "" {
		resp.LatestVersion = "检查失败: 更新源无版本信息"
		return resp, nil
	}
	resp.LatestVersion = latest
	resp.HasUpdate = compareVersions(latest, panelVersion) > 0
	return resp, nil
}

// compareVersions 语义化版本比较：a>b 返回 1，a<b 返回 -1，相等返回 0
func compareVersions(a, b string) int {
	parse := func(v string) []int {
		v = strings.TrimPrefix(strings.TrimSpace(v), "v")
		parts := strings.SplitN(v, "-", 2) // 去掉 -beta 等后缀参与主比较
		segs := strings.Split(parts[0], ".")
		nums := make([]int, 0, 3)
		for _, seg := range segs {
			n, err := strconv.Atoi(strings.TrimSpace(seg))
			if err != nil {
				n = 0
			}
			nums = append(nums, n)
		}
		for len(nums) < 3 {
			nums = append(nums, 0)
		}
		return nums
	}
	pa, pb := parse(a), parse(b)
	for i := 0; i < 3; i++ {
		if pa[i] > pb[i] {
			return 1
		}
		if pa[i] < pb[i] {
			return -1
		}
	}
	return 0
}
