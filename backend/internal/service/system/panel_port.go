package system

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

// --- 面板端口真生效（方案 A）：写回 config.yaml，重启后监听新端口 ---

var (
	portLineRe = regexp.MustCompile(`(?m)^(\s*port:\s*)(\d+)`)
	configPath = "config.yaml" // 与 main.go Load("config.yaml") 一致，工作目录 = backend/
)

// GetCurrentPort 从 config.yaml 读取当前配置的端口
func (s *Service) GetCurrentPort() (int, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return 0, fmt.Errorf("读取 config.yaml 失败: %w", err)
	}
	m := portLineRe.FindSubmatch(data)
	if m == nil {
		return 0, fmt.Errorf("config.yaml 中未找到 server.port 配置")
	}
	return strconv.Atoi(string(m[2]))
}

// UpdatePanelPort 把新端口写回 config.yaml。返回 (旧端口, 是否变更, 错误)。
// 运行中的进程仍监听旧端口——需要重启面板（systemctl restart rhinopanel）后才生效。
func (s *Service) UpdatePanelPort(newPort int) (int, bool, error) {
	if newPort < 1024 || newPort > 65535 {
		return 0, false, fmt.Errorf("端口必须在 1024-65535 之间（避免与系统端口冲突）")
	}

	current, err := s.GetCurrentPort()
	if err != nil {
		return 0, false, err
	}
	if current == newPort {
		return current, false, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return current, false, fmt.Errorf("读取 config.yaml 失败: %w", err)
	}
	updated := portLineRe.ReplaceAll(data, []byte("${1}"+strconv.Itoa(newPort)))
	if err := os.WriteFile(configPath, updated, 0644); err != nil {
		return current, false, fmt.Errorf("写入 config.yaml 失败: %w", err)
	}
	return current, true, nil
}
