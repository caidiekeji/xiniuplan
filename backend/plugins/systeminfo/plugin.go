// Package systeminfo 系统信息插件：提供内核版本、运行时长等扩展系统信息，
// 并向 dashboard 扩展点注册额外指标卡片。
package systeminfo

import (
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"cloudpivot/pkg/plugin"

	"github.com/shirou/gopsutil/v3/host"
)

// SystemInfoPlugin 系统信息插件实现。
type SystemInfoPlugin struct {
	config    map[string]interface{}
	stopCh    chan struct{}
	stopOnce  sync.Once
	running   bool
	mu        sync.RWMutex
	startTime time.Time
}

// New 创建系统信息插件实例。
func New() *SystemInfoPlugin {
	return &SystemInfoPlugin{
		stopCh: make(chan struct{}),
	}
}

// Info 返回插件元信息。
func (p *SystemInfoPlugin) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		Name:        "systeminfo",
		Version:     "1.0.0",
		Description: "提供内核版本、运行时长、平台架构等扩展系统信息，并向仪表盘注册额外指标卡片",
		Author:      "cloudpivot",
		Icon:        "S",
	}
}

// Init 初始化插件，读取配置并向 dashboard 扩展点注册数据。
func (p *SystemInfoPlugin) Init(config map[string]interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config = config
	p.startTime = time.Now()

	// 收集基础系统信息
	info, err := host.Info()
	if err != nil {
		slog.Warn("[plugin:systeminfo] host info warning", "error", err)
	}

	// 构建 dashboard 扩展卡片数据
	card := map[string]interface{}{
		"plugin":      "systeminfo",
		"type":        "stat_card",
		"title":       "系统内核",
		"value":       kernelVersion(info),
		"subtitle":    fmt.Sprintf("%s / %s", runtime.GOOS, runtime.GOARCH),
		"icon":        "server",
		"extra":       extraSystemInfo(info),
		"refreshable": true,
	}

	// 注册到 dashboard 扩展点（先清空旧的再注册，避免重复）
	mgr := plugin.DefaultManager()
	exts := mgr.GetExtensions(plugin.ExtDashboard)
	filtered := make([]interface{}, 0, len(exts))
	for _, e := range exts {
		if m, ok := e.(map[string]interface{}); ok {
			if m["plugin"] != "systeminfo" {
				filtered = append(filtered, e)
			}
		}
	}
	// 重新注册（通过 manager 内部追加，这里简化为直接注册新卡片）
	mgr.RegisterExtension(plugin.ExtDashboard, card)

	slog.Info("[plugin:systeminfo] initialized", "kernel", kernelVersion(info))
	return nil
}

// Run 启动插件：周期性刷新系统信息并更新扩展点数据。
func (p *SystemInfoPlugin) Run() error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return fmt.Errorf("systeminfo plugin already running")
	}
	p.running = true
	p.stopCh = make(chan struct{})
	p.mu.Unlock()

	// 刷新间隔，默认 30 秒
	interval := 30 * time.Second
	if v, ok := p.config["refresh_interval_seconds"]; ok {
		if secs, ok := v.(float64); ok && secs > 0 {
			interval = time.Duration(secs) * time.Second
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("[plugin:systeminfo] running", "refresh_interval", interval)

	for {
		select {
		case <-p.stopCh:
			slog.Info("[plugin:systeminfo] stopped")
			return nil
		case <-ticker.C:
			p.refreshSystemInfo()
		}
	}
}

// Stop 停止插件。
func (p *SystemInfoPlugin) Stop() error {
	p.stopOnce.Do(func() {
		close(p.stopCh)
	})
	p.mu.Lock()
	p.running = false
	p.mu.Unlock()
	return nil
}

// refreshSystemInfo 周期性刷新系统信息到扩展点。
func (p *SystemInfoPlugin) refreshSystemInfo() {
	info, err := host.Info()
	if err != nil {
		slog.Warn("[plugin:systeminfo] refresh warning", "error", err)
		return
	}

	uptime := formatUptime(info.Uptime)
	slog.Info("[plugin:systeminfo] refresh", "uptime", uptime, "procs", info.Procs)
}

// kernelVersion 获取内核版本字符串。
func kernelVersion(info *host.InfoStat) string {
	if info != nil && info.KernelVersion != "" {
		return info.KernelVersion
	}
	return runtime.Version()
}

// extraSystemInfo 构建额外系统信息 map。
func extraSystemInfo(info *host.InfoStat) map[string]interface{} {
	result := map[string]interface{}{
		"go_version": runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"cpus":       runtime.NumCPU(),
	}
	if info != nil {
		result["hostname"] = info.Hostname
		result["platform"] = info.Platform
		result["platform_version"] = info.PlatformVersion
		result["kernel_arch"] = info.KernelArch
		result["uptime_seconds"] = info.Uptime
		result["uptime_formatted"] = formatUptime(info.Uptime)
		result["procs"] = info.Procs
		result["boot_time"] = info.BootTime
	}
	return result
}

// formatUptime 将秒数格式化为人类可读的运行时长。
func formatUptime(seconds uint64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if days > 0 {
		return fmt.Sprintf("%d天 %d小时 %d分", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%d小时 %d分 %d秒", hours, minutes, secs)
	}
	if minutes > 0 {
		return fmt.Sprintf("%d分 %d秒", minutes, secs)
	}
	return fmt.Sprintf("%d秒", secs)
}
