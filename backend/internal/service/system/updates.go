package system

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"cloudpivot/internal/common"
)

// --- 系统安全更新（SY-005）与 NTP（SY-006） ---

// PackageUpdate 可升级的系统包
type PackageUpdate struct {
	Name    string `json:"name"`
	Current string `json:"current"`
	New     string `json:"new"`
	Source  string `json:"source"`
}

// detectPackageManager 探测包管理器
func (s *Service) detectPackageManager() (string, string, error) {
	if runtime.GOOS != "linux" {
		return "", "", fmt.Errorf("系统更新仅支持 Linux")
	}
	for _, pm := range []struct {
		bin  string
		kind string
	}{
		{"apt-get", "apt"},
		{"dnf", "dnf"},
		{"yum", "yum"},
	} {
		if _, err := exec.LookPath(pm.bin); err == nil {
			return pm.kind, pm.bin, nil
		}
	}

	return "", "", fmt.Errorf("未检测到 apt-get/dnf/yum 包管理器")
}

// CheckSystemUpdates 检查系统可用更新
func (s *Service) CheckSystemUpdates() ([]PackageUpdate, string, error) {
	kind, _, err := s.detectPackageManager()
	if err != nil {
		return nil, kind, err
	}

	var out string
	switch kind {
	case "apt":
		// 先刷新索引（只读操作，不升级）
		_, _ = s.updExec.Execute(context.Background(), "apt_update_index", nil)
		result, err := s.updExec.Execute(context.Background(), "apt_list_upgradable", nil)
		if err != nil {
			return nil, kind, fmt.Errorf("检查更新失败: %w", err)
		}
		if result.ExitCode != 0 {
			return nil, kind, fmt.Errorf("检查更新失败: %s", result.Stderr)
		}
		out = result.Stdout
	case "yum":
		result, err := s.updExec.Execute(context.Background(), "yum_check_update", nil)
		if err != nil {
			return nil, kind, fmt.Errorf("检查更新失败: %w", err)
		}
		out = result.Stdout
		// yum check-update 有更新时 exit code 100，stdout 仍是列表
	case "dnf":
		result, err := s.updExec.Execute(context.Background(), "dnf_check_update", nil)
		if err != nil {
			return nil, kind, fmt.Errorf("检查更新失败: %w", err)
		}
		out = result.Stdout
	}

	updates := parseUpdateList(kind, out)
	return updates, kind, nil
}

// parseUpdateList 解析包管理器输出
func parseUpdateList(kind, output string) []PackageUpdate {
	updates := make([]PackageUpdate, 0)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.Contains(line, "Listing") {
			continue
		}
		switch kind {
		case "apt":
			// 格式: name/repo version arch [upgradable from: x]
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				name := fields[0]
				if idx := strings.Index(name, "/"); idx > 0 {
					name = name[:idx]
				}
				updates = append(updates, PackageUpdate{
					Name:   name,
					New:    fields[1],
					Source: fields[0][len(name)+1:],
				})
			}
		case "yum", "dnf":
			// 格式: name.arch version repo
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				name := fields[0]
				if idx := strings.LastIndex(name, "."); idx > 0 {
					name = name[:idx]
				}
				updates = append(updates, PackageUpdate{
					Name:   name,
					New:    fields[1],
					Source: fields[2],
				})
			}
		}
	}
	return updates
}

// InstallSystemUpdates 一键安装系统更新（需要显式确认）
func (s *Service) InstallSystemUpdates(confirm bool) (string, error) {
	if !confirm {
		return "", fmt.Errorf("安装系统更新为高危操作，请传入 confirm=true 确认")
	}
	kind, _, err := s.detectPackageManager()
	if err != nil {
		return "", err
	}

	var result *common.CommandResult

	switch kind {
	case "apt":
		result, err = s.updExec.Execute(context.Background(), "apt_upgrade", nil)
	case "yum":
		result, err = s.updExec.Execute(context.Background(), "yum_update", nil)
	case "dnf":
		result, err = s.updExec.Execute(context.Background(), "dnf_update", nil)
	}
	if err != nil {
		return "", fmt.Errorf("系统更新失败: %w", err)
	}
	if result.ExitCode != 0 {
		return "", fmt.Errorf("系统更新失败 (exit %d): %s", result.ExitCode, result.Stderr)
	}
	return result.Stdout, nil
}

// NTPStatus NTP 时间同步状态（SY-006）
type NTPStatus struct {
	NTPEnabled bool   `json:"ntp_enabled"`
	Synced     bool   `json:"synced"`
	Timezone   string `json:"timezone"`
	LocalTime  string `json:"local_time"`
}

// GetNTPStatus 读取时间同步状态
func (s *Service) GetNTPStatus() (*NTPStatus, error) {
	result, err := s.executor.Execute(context.Background(), "timedatectl_status", nil)
	if err != nil || result.ExitCode != 0 {
		// timedatectl 不可用（如容器/Windows），返回本机时间兜底
		now := time.Now()
		zone, _ := now.Zone()
		return &NTPStatus{
			Timezone:  zone,
			LocalTime: now.Format("2006-01-02 15:04:05 MST"),
		}, nil
	}
	status := &NTPStatus{}
	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "NTP service:"):
			status.NTPEnabled = strings.Contains(line, "active")
		case strings.HasPrefix(line, "System clock synchronized:"):
			status.Synced = strings.Contains(line, "yes")
		case strings.HasPrefix(line, "Time zone:"):
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				status.Timezone = fields[2]
			}
		case strings.HasPrefix(line, "Local time:"):
			status.LocalTime = strings.TrimPrefix(line, "Local time: ")
		}
	}
	return status, nil
}

// SetNTPEnabled 开启/关闭 NTP 同步
func (s *Service) SetNTPEnabled(enabled bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if enabled {
		_, err := s.executor.Execute(ctx, "timedatectl_ntp_on", nil)
		if err != nil {
			return fmt.Errorf("开启 NTP 失败: %w", err)
		}
		return nil
	}
	_, err := s.executor.Execute(ctx, "timedatectl_ntp_off", nil)
	if err != nil {
		return fmt.Errorf("关闭 NTP 失败: %w", err)
	}
	return nil
}

// registerUpdateTemplates 注册系统更新/NTP 命令白名单
func registerUpdateTemplates(e *common.CommandExecutor) {
	e.Register(&common.CommandTemplate{
		Name: "apt_update_index",
		Bin:  "apt-get",
		Args: []string{"update"},
	})
	e.Register(&common.CommandTemplate{
		Name: "apt_list_upgradable",
		Bin:  "apt",
		Args: []string{"list", "--upgradable"},
	})
	e.Register(&common.CommandTemplate{
		Name: "apt_upgrade",
		Bin:  "apt-get",
		Args: []string{"upgrade", "-y"},
	})
	e.Register(&common.CommandTemplate{
		Name: "yum_check_update",
		Bin:  "yum",
		Args: []string{"check-update", "-q"},
	})
	e.Register(&common.CommandTemplate{
		Name: "dnf_check_update",
		Bin:  "dnf",
		Args: []string{"check-update", "-q"},
	})
	e.Register(&common.CommandTemplate{
		Name: "yum_update",
		Bin:  "yum",
		Args: []string{"update", "-y"},
	})
	e.Register(&common.CommandTemplate{
		Name: "dnf_update",
		Bin:  "dnf",
		Args: []string{"update", "-y"},
	})
	e.Register(&common.CommandTemplate{
		Name: "timedatectl_status",
		Bin:  "timedatectl",
		Args: []string{"status"},
	})
	e.Register(&common.CommandTemplate{
		Name: "timedatectl_ntp_on",
		Bin:  "timedatectl",
		Args: []string{"set-ntp", "true"},
	})
	e.Register(&common.CommandTemplate{
		Name: "timedatectl_ntp_off",
		Bin:  "timedatectl",
		Args: []string{"set-ntp", "false"},
	})
}
