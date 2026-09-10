package ssh

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"

	"gorm.io/gorm"
)

// 常量定义
const (
	sshdConfigPath    = "/etc/ssh/sshd_config"
	fail2banJailLocal = "/etc/fail2ban/jail.local"
)

// Service SSH 安全加固服务
type Service struct {
	db       *gorm.DB
	executor *common.CommandExecutor
}

// NewService 创建 SSH 服务
func NewService(db *gorm.DB) *Service {
	executor := common.NewCommandExecutor(30 * time.Second)
	registerCommandTemplates(executor)
	return &Service{
		db:       db,
		executor: executor,
	}
}

// registerCommandTemplates 注册系统命令白名单模板
func registerCommandTemplates(e *common.CommandExecutor) {
	// sshd 配置语法校验
	e.Register(&common.CommandTemplate{
		Name: "sshd_test",
		Bin:  "sshd",
		Args: []string{"-t"},
	})
	// 重载 sshd 服务
	e.Register(&common.CommandTemplate{
		Name: "sshd_reload",
		Bin:  "systemctl",
		Args: []string{"reload", "sshd"},
	})
	// fail2ban 状态查询
	e.Register(&common.CommandTemplate{
		Name: "fail2ban_status",
		Bin:  "fail2ban-client",
		Args: []string{"status", "sshd"},
	})
	// fail2ban 解封 IP
	e.Register(&common.CommandTemplate{
		Name: "fail2ban_unban",
		Bin:  "fail2ban-client",
		Args: []string{"set", "sshd", "unbanip", "{{ip}}"},
	})
	// fail2ban 封禁 IP
	e.Register(&common.CommandTemplate{
		Name: "fail2ban_ban",
		Bin:  "fail2ban-client",
		Args: []string{"set", "sshd", "banip", "{{ip}}"},
	})
	// 重载 fail2ban 服务
	e.Register(&common.CommandTemplate{
		Name: "fail2ban_reload",
		Bin:  "systemctl",
		Args: []string{"reload", "fail2ban"},
	})
	// 设置 authorized_keys 权限
	e.Register(&common.CommandTemplate{
		Name: "chmod_authorized_keys",
		Bin:  "chmod",
		Args: []string{"600", "{{path}}"},
	})
	// last 命令查询登录记录
	e.Register(&common.CommandTemplate{
		Name: "last_logins",
		Bin:  "last",
		Args: []string{"-n", "{{count}}"},
	})
	// lastb 命令查询失败登录记录
	e.Register(&common.CommandTemplate{
		Name: "lastb_failed",
		Bin:  "lastb",
		Args: []string{"-n", "{{count}}"},
	})
}

// ============================================================
// SSH 配置管理
// ============================================================

// GetConfig 读取当前 SSH 配置
func (s *Service) GetConfig() (*dto.SSHConfig, error) {
	data, err := os.ReadFile(sshdConfigPath)
	if err != nil {
		return nil, fmt.Errorf("读取 sshd_config 失败: %w", err)
	}

	cfg := &dto.SSHConfig{
		Port:                22,
		PermitRootLogin:     "yes",
		PermitPasswordLogin: "yes",
		MaxAuthTries:        6,
		ClientAliveInterval: 0,
		AllowUsers:          "",
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		key := strings.ToLower(parts[0])
		value := strings.Join(parts[1:], " ")

		switch key {
		case "port":
			if p, err := strconv.Atoi(parts[1]); err == nil {
				cfg.Port = p
			}
		case "permitrootlogin":
			cfg.PermitRootLogin = value
		case "passwordauthentication":
			cfg.PermitPasswordLogin = value
		case "maxauthtries":
			if m, err := strconv.Atoi(parts[1]); err == nil {
				cfg.MaxAuthTries = m
			}
		case "clientaliveinterval":
			if c, err := strconv.Atoi(parts[1]); err == nil {
				cfg.ClientAliveInterval = c
			}
		case "allowusers":
			cfg.AllowUsers = value
		}
	}

	return cfg, nil
}

// UpdateConfig 修改 SSH 配置（备份 → 修改 → 语法校验 → reload）
func (s *Service) UpdateConfig(req dto.UpdateSSHConfigRequest) error {
	// 1. 读取原文件
	data, err := os.ReadFile(sshdConfigPath)
	if err != nil {
		return fmt.Errorf("读取 sshd_config 失败: %w", err)
	}

	// 2. 备份原文件
	backupPath := fmt.Sprintf("%s.bak.%s", sshdConfigPath, time.Now().Format("20060102_150405"))
	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return fmt.Errorf("备份 sshd_config 失败: %w", err)
	}

	// 3. 构建需要更新的键值对
	updates := map[string]string{
		"port":                   strconv.Itoa(req.Port),
		"permitrootlogin":        req.PermitRootLogin,
		"passwordauthentication": req.PermitPasswordLogin,
		"maxauthtries":           strconv.Itoa(req.MaxAuthTries),
		"clientaliveinterval":    strconv.Itoa(req.ClientAliveInterval),
		"allowusers":             req.AllowUsers,
	}

	// 4. 修改配置：逐行处理，替换已有键，记录已处理的键
	lines := strings.Split(string(data), "\n")
	processed := make(map[string]bool)
	var newLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			newLines = append(newLines, line)
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 1 {
			newLines = append(newLines, line)
			continue
		}
		key := strings.ToLower(parts[0])
		if newValue, ok := updates[key]; ok {
			processed[key] = true
			if newValue == "" && key == "allowusers" {
				// AllowUsers 为空时注释掉该行
				newLines = append(newLines, "#"+line)
			} else {
				newLines = append(newLines, parts[0]+" "+newValue)
			}
		} else {
			newLines = append(newLines, line)
		}
	}

	// 5. 追加未找到的键
	var appendKeys []string
	for key, value := range updates {
		if !processed[key] {
			if key == "allowusers" && value == "" {
				continue
			}
			// 还原首字母大写的键名
			displayKey := map[string]string{
				"port":                   "Port",
				"permitrootlogin":        "PermitRootLogin",
				"passwordauthentication": "PasswordAuthentication",
				"maxauthtries":           "MaxAuthTries",
				"clientaliveinterval":    "ClientAliveInterval",
				"allowusers":             "AllowUsers",
			}[key]
			appendKeys = append(appendKeys, displayKey+" "+value)
		}
	}
	sort.Strings(appendKeys)
	newLines = append(newLines, appendKeys...)

	// 6. 写入新配置
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(sshdConfigPath, []byte(newContent), 0600); err != nil {
		// 写入失败，恢复备份
		_ = os.WriteFile(sshdConfigPath, data, 0600)
		return fmt.Errorf("写入 sshd_config 失败: %w", err)
	}

	// 7. sshd -t 语法校验
	ctx := context.Background()
	result, err := s.executor.Execute(ctx, "sshd_test", nil)
	if err != nil || result.ExitCode != 0 {
		// 校验失败，恢复备份
		_ = os.WriteFile(sshdConfigPath, data, 0600)
		stderr := ""
		if result != nil {
			stderr = result.Stderr
		}
		return fmt.Errorf("sshd 配置语法校验失败: %s", stderr)
	}

	// 8. systemctl reload sshd
	if _, err := s.executor.Execute(ctx, "sshd_reload", nil); err != nil {
		// reload 失败不回滚配置（语法已通过），但返回警告
		fmt.Printf("[ssh] sshd reload 失败（配置已更新，语法校验通过）: %v\n", err)
	}

	return nil
}

// ============================================================
// SSH 密钥管理
// ============================================================

// authorizedKeysPath 获取 authorized_keys 文件路径
func authorizedKeysPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %w", err)
	}
	return filepath.Join(home, ".ssh", "authorized_keys"), nil
}

// ListKeys 列出 authorized_keys 中的公钥
func (s *Service) ListKeys() ([]dto.SSHKey, error) {
	keyPath, err := authorizedKeysPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(keyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []dto.SSHKey{}, nil
		}
		return nil, fmt.Errorf("读取 authorized_keys 失败: %w", err)
	}

	// 获取文件修改时间作为 created_at 近似值
	fileInfo, err := os.Stat(keyPath)
	var modTime string
	if err == nil {
		modTime = fileInfo.ModTime().Format("2006-01-02 15:04:05")
	}

	keys := make([]dto.SSHKey, 0)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	id := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		keyType := parts[0]
		keyData := parts[1]
		comment := ""
		if len(parts) >= 3 {
			comment = strings.Join(parts[2:], " ")
		}

		// 校验密钥类型前缀
		if !strings.HasPrefix(keyType, "ssh-") && !strings.HasPrefix(keyType, "ecdsa-") {
			continue
		}

		id++
		fingerprint := computeFingerprint(keyData)
		keys = append(keys, dto.SSHKey{
			ID:          id,
			Type:        normalizeKeyType(keyType),
			Comment:     comment,
			Fingerprint: fingerprint,
			CreatedAt:   modTime,
		})
	}

	return keys, nil
}

// AddKey 追加公钥到 authorized_keys
func (s *Service) AddKey(publicKey string) (*dto.SSHKey, error) {
	publicKey = strings.TrimSpace(publicKey)
	if publicKey == "" {
		return nil, fmt.Errorf("公钥内容不能为空")
	}

	parts := strings.Fields(publicKey)
	if len(parts) < 2 {
		return nil, fmt.Errorf("公钥格式无效，需要 type 和 key 两部分")
	}

	keyType := parts[0]
	// 校验格式前缀
	if !strings.HasPrefix(keyType, "ssh-") && !strings.HasPrefix(keyType, "ecdsa-") {
		return nil, fmt.Errorf("不支持的密钥类型: %s，仅支持 ssh-rsa / ssh-ed25519 / ecdsa-*", keyType)
	}

	keyData := parts[1]
	comment := ""
	if len(parts) >= 3 {
		comment = strings.Join(parts[2:], " ")
	}

	// 校验 base64 格式
	if _, err := base64.StdEncoding.DecodeString(keyData); err != nil {
		return nil, fmt.Errorf("公钥 base64 解码失败: %w", err)
	}

	keyPath, err := authorizedKeysPath()
	if err != nil {
		return nil, err
	}

	// 确保 .ssh 目录存在
	sshDir := filepath.Dir(keyPath)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return nil, fmt.Errorf("创建 .ssh 目录失败: %w", err)
	}

	// 检查是否已存在（按 fingerprint 去重）
	newFingerprint := computeFingerprint(keyData)
	existingKeys, err := s.ListKeys()
	if err == nil {
		for _, k := range existingKeys {
			if k.Fingerprint == newFingerprint {
				return nil, fmt.Errorf("该公钥已存在（fingerprint: %s）", newFingerprint)
			}
		}
	}

	// 追加公钥
	f, err := os.OpenFile(keyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("打开 authorized_keys 失败: %w", err)
	}
	defer f.Close()

	// 确保文件末尾有换行
	if fileInfo, statErr := os.Stat(keyPath); statErr == nil && fileInfo.Size() > 0 {
		// 读取最后一个字节判断是否有换行
		if data, readErr := os.ReadFile(keyPath); readErr == nil && len(data) > 0 && data[len(data)-1] != '\n' {
			if _, err := f.WriteString("\n"); err != nil {
				return nil, fmt.Errorf("写入换行失败: %w", err)
			}
		}
	}

	line := keyType + " " + keyData
	if comment != "" {
		line += " " + comment
	}
	if _, err := f.WriteString(line + "\n"); err != nil {
		return nil, fmt.Errorf("写入公钥失败: %w", err)
	}

	// chmod 600
	ctx := context.Background()
	if _, err := s.executor.Execute(ctx, "chmod_authorized_keys", map[string]string{
		"path": keyPath,
	}); err != nil {
		fmt.Printf("[ssh] chmod authorized_keys 失败: %v\n", err)
	}

	// 获取文件修改时间
	fileInfo, err := os.Stat(keyPath)
	modTime := ""
	if err == nil {
		modTime = fileInfo.ModTime().Format("2006-01-02 15:04:05")
	}

	// 计算新 ID
	newID := len(existingKeys) + 1

	return &dto.SSHKey{
		ID:          newID,
		Type:        normalizeKeyType(keyType),
		Comment:     comment,
		Fingerprint: newFingerprint,
		CreatedAt:   modTime,
	}, nil
}

// DeleteKey 按 fingerprint 删除公钥
func (s *Service) DeleteKey(fingerprint string) error {
	fingerprint = strings.TrimSpace(fingerprint)
	if fingerprint == "" {
		return fmt.Errorf("fingerprint 不能为空")
	}

	keyPath, err := authorizedKeysPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("读取 authorized_keys 失败: %w", err)
	}

	var newLines []string
	found := false
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			newLines = append(newLines, line)
			continue
		}
		parts := strings.Fields(trimmed)
		if len(parts) < 2 {
			newLines = append(newLines, line)
			continue
		}
		fp := computeFingerprint(parts[1])
		if fp == fingerprint {
			found = true
			continue // 跳过该行即删除
		}
		newLines = append(newLines, line)
	}

	if !found {
		return fmt.Errorf("未找到 fingerprint 为 %s 的公钥", fingerprint)
	}

	// 写回文件
	newContent := strings.Join(newLines, "\n")
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	if err := os.WriteFile(keyPath, []byte(newContent), 0600); err != nil {
		return fmt.Errorf("写入 authorized_keys 失败: %w", err)
	}

	return nil
}

// computeFingerprint 计算公钥的 SHA256 fingerprint
func computeFingerprint(keyBase64 string) string {
	decoded, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(decoded)
	return "SHA256:" + base64.RawStdEncoding.EncodeToString(hash[:])
}

// normalizeKeyType 将 ssh-rsa / ssh-ed25519 / ecdsa-sha2-nistp256 归一化为 rsa/ed25519/ecdsa
func normalizeKeyType(keyType string) string {
	switch {
	case strings.HasPrefix(keyType, "ssh-rsa"):
		return "rsa"
	case strings.HasPrefix(keyType, "ssh-ed25519"):
		return "ed25519"
	case strings.HasPrefix(keyType, "ecdsa-"):
		return "ecdsa"
	default:
		return keyType
	}
}

// ============================================================
// SSH 登录日志
// ============================================================

// authLogRe 匹配 auth.log 中的 SSH 登录记录
// 示例: Sep  8 12:00:00 host sshd[1234]: Accepted password for user from 1.2.3.4 port 12345 ssh2
var authLogRe = regexp.MustCompile(
	`^(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+\S+\s+sshd\[\d+\]:\s+(Accepted|Failed)\s+(\S+)\s+for\s+(?:invalid user\s+)?(\S+)\s+from\s+(\S+)`,
)

// GetLoginLogs 解析 SSH 登录日志（分页）
func (s *Service) GetLoginLogs(page, pageSize int) ([]dto.SSHLoginLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	allLogs := s.parseAuthLogs()

	// 按时间倒序（日志文件通常是正序，反转）
	for i, j := 0, len(allLogs)-1; i < j; i, j = i+1, j-1 {
		allLogs[i], allLogs[j] = allLogs[j], allLogs[i]
	}

	total := int64(len(allLogs))
	start := (page - 1) * pageSize
	if start >= len(allLogs) {
		return []dto.SSHLoginLog{}, total, nil
	}
	end := start + pageSize
	if end > len(allLogs) {
		end = len(allLogs)
	}

	return allLogs[start:end], total, nil
}

// parseAuthLogs 从 auth.log / secure 解析 SSH 登录记录，失败则回退到 last/lastb
func (s *Service) parseAuthLogs() []dto.SSHLoginLog {
	logs := make([]dto.SSHLoginLog, 0)

	// 尝试常见日志文件路径
	logPaths := []string{
		"/var/log/auth.log",
		"/var/log/secure",
	}

	for _, logPath := range logPaths {
		data, err := os.ReadFile(logPath)
		if err != nil {
			continue
		}
		logs = append(logs, parseAuthLogContent(string(data))...)
	}

	if len(logs) > 0 {
		return logs
	}

	// 回退：使用 last / lastb 命令
	return s.parseLastCommands()
}

// parseAuthLogContent 解析 auth.log 内容
func parseAuthLogContent(content string) []dto.SSHLoginLog {
	logs := make([]dto.SSHLoginLog, 0)
	currentYear := time.Now().Year()

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		match := authLogRe.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		timeStr := match[1]
		status := "success"
		if match[2] == "Failed" {
			status = "failed"
		}
		method := match[3]
		user := match[4]
		ip := match[5]

		// 解析时间（补全年份）
		fullTimeStr := fmt.Sprintf("%d %s", currentYear, timeStr)
		parsedTime, err := time.Parse("2006 Jan 2 15:04:05", fullTimeStr)
		if err != nil {
			// 解析失败，使用原始字符串
			logs = append(logs, dto.SSHLoginLog{
				Time:   timeStr,
				User:   user,
				IP:     ip,
				Status: status,
				Method: method,
			})
			continue
		}

		// 如果解析出的时间在未来，说明是去年的日志
		if parsedTime.After(time.Now()) {
			parsedTime = parsedTime.AddDate(-1, 0, 0)
		}

		logs = append(logs, dto.SSHLoginLog{
			Time:   parsedTime.Format("2006-01-02 15:04:05"),
			User:   user,
			IP:     ip,
			Status: status,
			Method: method,
		})
	}

	return logs
}

// parseLastCommands 使用 last/lastb 命令回退解析登录记录
func (s *Service) parseLastCommands() []dto.SSHLoginLog {
	logs := make([]dto.SSHLoginLog, 0)
	ctx := context.Background()

	// 成功登录
	if result, err := s.executor.Execute(ctx, "last_logins", map[string]string{"count": "50"}); err == nil {
		logs = append(logs, parseLastOutput(result.Stdout, "success")...)
	}

	// 失败登录
	if result, err := s.executor.Execute(ctx, "lastb_failed", map[string]string{"count": "50"}); err == nil {
		logs = append(logs, parseLastOutput(result.Stdout, "failed")...)
	}

	return logs
}

// parseLastOutput 解析 last/lastb 命令输出
// 格式: user     pts/0        192.168.1.1     Mon Sep  8 10:00   still logged in
func parseLastOutput(output string, status string) []dto.SSHLoginLog {
	logs := make([]dto.SSHLoginLog, 0)
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "wtmp") || strings.HasPrefix(line, "btmp") || strings.HasPrefix(line, "reboot") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		user := parts[0]
		// IP 通常在第三列（pts/0 之后）
		ip := ""
		timeStr := ""
		for i, p := range parts {
			if isIPAddress(p) {
				ip = p
				if i+1 < len(parts) {
					timeStr = strings.Join(parts[i+1:], " ")
				}
				break
			}
		}
		if ip == "" {
			continue
		}
		logs = append(logs, dto.SSHLoginLog{
			Time:   timeStr,
			User:   user,
			IP:     ip,
			Status: status,
			Method: "unknown",
		})
	}
	return logs
}

// isIPAddress 简单判断是否为 IP 地址
func isIPAddress(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || n > 255 {
			return false
		}
	}
	return true
}

// ============================================================
// Fail2ban 管理
// ============================================================

// fail2banInstalled 检测 fail2ban 是否已安装
func (s *Service) fail2banInstalled() bool {
	ctx := context.Background()
	_, err := s.executor.Execute(ctx, "fail2ban_status", nil)
	return err == nil
}

// GetFail2banStatus 获取 Fail2ban 状态和封禁列表
func (s *Service) GetFail2banStatus() (*dto.Fail2banStatus, error) {
	status := &dto.Fail2banStatus{
		Installed: false,
		BannedIPs: make([]dto.BannedIP, 0),
	}

	// 检测是否安装
	if !s.fail2banInstalled() {
		return status, nil
	}
	status.Installed = true

	// 读取配置判断是否启用
	cfg, _ := s.GetFail2banConfig()
	if cfg != nil {
		status.Enabled = cfg.Enabled
	}

	// 执行 fail2ban-client status sshd
	ctx := context.Background()
	result, err := s.executor.Execute(ctx, "fail2ban_status", nil)
	if err != nil {
		return status, nil
	}

	// 解析输出
	status.CurrentlyBanned = parseFail2banInt(result.Stdout, "Currently banned")
	status.TotalBanned = parseFail2banInt(result.Stdout, "Total banned")
	status.CurrentlyFailed = parseFail2banInt(result.Stdout, "Currently failed")
	status.TotalFailed = parseFail2banInt(result.Stdout, "Total failed")

	// 解析封禁 IP 列表
	bannedIPs := parseFail2banBannedIPs(result.Stdout)
	now := time.Now().Format("2006-01-02 15:04:05")
	for _, ip := range bannedIPs {
		status.BannedIPs = append(status.BannedIPs, dto.BannedIP{
			IP:       ip,
			Jail:     "sshd",
			BannedAt: now,
			Reason:   "SSH 暴力破解",
		})
	}

	return status, nil
}

// parseFail2banInt 从 fail2ban 状态输出中解析整数值
func parseFail2banInt(output, key string) int {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, key) {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				val := strings.TrimSpace(parts[len(parts)-1])
				if n, err := strconv.Atoi(val); err == nil {
					return n
				}
			}
		}
	}
	return 0
}

// parseFail2banBannedIPs 解析 Banned IP list
func parseFail2banBannedIPs(output string) []string {
	ips := make([]string, 0)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Banned IP list") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				ipStr := strings.TrimSpace(parts[len(parts)-1])
				if ipStr != "" {
					for _, ip := range strings.Fields(ipStr) {
						ips = append(ips, ip)
					}
				}
			}
			break
		}
	}
	return ips
}

// GetFail2banConfig 读取 Fail2ban 配置
func (s *Service) GetFail2banConfig() (*dto.Fail2banConfig, error) {
	cfg := &dto.Fail2banConfig{
		Enabled:  false,
		MaxRetry: 3,
		BanTime:  3600,
		FindTime: 600,
	}

	data, err := os.ReadFile(fail2banJailLocal)
	if err != nil {
		// 文件不存在，返回默认值
		return cfg, nil
	}

	inSSHDSection := false
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 段标记
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inSSHDSection = (line == "[sshd]")
			continue
		}
		if !inSSHDSection {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "enabled":
			cfg.Enabled = (value == "true" || value == "1")
		case "maxretry":
			if n, err := strconv.Atoi(value); err == nil {
				cfg.MaxRetry = n
			}
		case "bantime":
			if n, err := strconv.Atoi(value); err == nil {
				cfg.BanTime = n
			}
		case "findtime":
			if n, err := strconv.Atoi(value); err == nil {
				cfg.FindTime = n
			}
		}
	}

	return cfg, nil
}

// UpdateFail2banConfig 修改 Fail2ban 配置
func (s *Service) UpdateFail2banConfig(req dto.UpdateFail2banConfigRequest) error {
	// 检测 fail2ban 是否安装
	if !s.fail2banInstalled() {
		return fmt.Errorf("fail2ban 未安装，请先安装 fail2ban（如: apt install fail2ban 或 yum install fail2ban）")
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(fail2banJailLocal), 0755); err != nil {
		return fmt.Errorf("创建 fail2ban 配置目录失败: %w", err)
	}

	// 读取现有配置
	data, err := os.ReadFile(fail2banJailLocal)
	var lines []string
	if err == nil {
		lines = strings.Split(string(data), "\n")
	} else {
		lines = []string{}
	}

	// 构建新的 [sshd] 段内容
	enabledStr := "false"
	if req.Enabled {
		enabledStr = "true"
	}
	newSSHDSection := []string{
		"[sshd]",
		fmt.Sprintf("enabled = %s", enabledStr),
		fmt.Sprintf("maxretry = %d", req.MaxRetry),
		fmt.Sprintf("bantime = %d", req.BanTime),
		fmt.Sprintf("findtime = %d", req.FindTime),
	}

	// 查找并替换 [sshd] 段
	var newLines []string
	inSSHDSection := false
	foundSSHD := false
	sectionReplaced := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			if inSSHDSection {
				// 离开 sshd 段，插入新内容
				if !sectionReplaced {
					newLines = append(newLines, newSSHDSection...)
					sectionReplaced = true
				}
				inSSHDSection = false
			}
			if trimmed == "[sshd]" {
				inSSHDSection = true
				foundSSHD = true
				continue // 跳过旧的段头，稍后统一插入
			}
		}
		if inSSHDSection {
			continue // 跳过旧段内容
		}
		newLines = append(newLines, line)
	}

	// 如果在文件末尾的 sshd 段
	if inSSHDSection && !sectionReplaced {
		newLines = append(newLines, newSSHDSection...)
		sectionReplaced = true
	}

	// 如果没有找到 [sshd] 段，追加
	if !foundSSHD {
		if len(newLines) > 0 && newLines[len(newLines)-1] != "" {
			newLines = append(newLines, "")
		}
		newLines = append(newLines, newSSHDSection...)
	}

	// 写入配置
	newContent := strings.Join(newLines, "\n")
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	if err := os.WriteFile(fail2banJailLocal, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("写入 jail.local 失败: %w", err)
	}

	// 重载 fail2ban
	ctx := context.Background()
	if _, err := s.executor.Execute(ctx, "fail2ban_reload", nil); err != nil {
		fmt.Printf("[ssh] fail2ban reload 失败: %v\n", err)
	}

	return nil
}

// UnbanIP 解封 IP
func (s *Service) UnbanIP(ip string) error {
	if !s.fail2banInstalled() {
		return fmt.Errorf("fail2ban 未安装，请先安装 fail2ban")
	}
	if !isIPAddress(ip) {
		return fmt.Errorf("无效的 IP 地址: %s", ip)
	}

	ctx := context.Background()
	result, err := s.executor.Execute(ctx, "fail2ban_unban", map[string]string{"ip": ip})
	if err != nil {
		return fmt.Errorf("解封 IP 失败: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("解封 IP 失败: %s", result.Stderr)
	}
	return nil
}

// BanIP 手动封禁 IP
func (s *Service) BanIP(ip string) error {
	if !s.fail2banInstalled() {
		return fmt.Errorf("fail2ban 未安装，请先安装 fail2ban")
	}
	if !isIPAddress(ip) {
		return fmt.Errorf("无效的 IP 地址: %s", ip)
	}

	ctx := context.Background()
	result, err := s.executor.Execute(ctx, "fail2ban_ban", map[string]string{"ip": ip})
	if err != nil {
		return fmt.Errorf("封禁 IP 失败: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("封禁 IP 失败: %s", result.Stderr)
	}
	return nil
}
