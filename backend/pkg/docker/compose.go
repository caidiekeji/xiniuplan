package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloudpivot/internal/common"
)

// ComposeProject docker compose 项目
type ComposeProject struct {
	Name     string   `json:"name"`
	Path     string   `json:"path"`
	Services []string `json:"services"`
}

// composeExecutor 全局命令执行器（懒加载）
var composeExecutor *common.CommandExecutor

// getComposeExecutor 获取或初始化 compose 命令执行器
func getComposeExecutor() *common.CommandExecutor {
	if composeExecutor != nil {
		return composeExecutor
	}
	exec := common.NewCommandExecutor(120 * time.Second) // 120s

	// docker compose up -d（工作目录通过参数传入）
	exec.Register(&common.CommandTemplate{
		Name:        "compose_up",
		Description: "docker compose up -d",
		Bin:         "docker",
		Args:        []string{"compose", "-f", "{{compose_file}}", "up", "-d"},
		AllowedArgs: map[string]bool{"compose_file": true},
	})

	// docker compose down
	exec.Register(&common.CommandTemplate{
		Name:        "compose_down",
		Description: "docker compose down",
		Bin:         "docker",
		Args:        []string{"compose", "-f", "{{compose_file}}", "down"},
		AllowedArgs: map[string]bool{"compose_file": true},
	})

	// docker compose logs
	exec.Register(&common.CommandTemplate{
		Name:        "compose_logs",
		Description: "docker compose logs",
		Bin:         "docker",
		Args:        []string{"compose", "-f", "{{compose_file}}", "logs", "--tail", "{{tail}}", "{{service}}"},
		AllowedArgs: map[string]bool{"compose_file": true, "tail": true, "service": true},
	})

	// docker compose ps
	exec.Register(&common.CommandTemplate{
		Name:        "compose_ps",
		Description: "docker compose ps",
		Bin:         "docker",
		Args:        []string{"compose", "-f", "{{compose_file}}", "ps", "--format", "json"},
		AllowedArgs: map[string]bool{"compose_file": true},
	})

	composeExecutor = exec
	return exec
}

// Up 启动 compose 项目
func Up(ctx context.Context, projectPath string) error {
	composeFile := filepath.Join(projectPath, "docker-compose.yml")
	if _, err := os.Stat(composeFile); err != nil {
		return fmt.Errorf("compose file not found: %s", composeFile)
	}

	exec := getComposeExecutor()
	result, err := exec.Execute(ctx, "compose_up", map[string]string{
		"compose_file": composeFile,
	})
	if err != nil {
		return fmt.Errorf("compose up failed: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("compose up exit %d: %s", result.ExitCode, result.Stderr)
	}
	return nil
}

// Down 停止并移除 compose 项目
func Down(ctx context.Context, projectPath string) error {
	composeFile := filepath.Join(projectPath, "docker-compose.yml")
	if _, err := os.Stat(composeFile); err != nil {
		// 文件不存在视为已停止
		return nil
	}

	exec := getComposeExecutor()
	result, err := exec.Execute(ctx, "compose_down", map[string]string{
		"compose_file": composeFile,
	})
	if err != nil {
		return fmt.Errorf("compose down failed: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("compose down exit %d: %s", result.ExitCode, result.Stderr)
	}
	return nil
}

// Logs 获取 compose 项目服务日志
func Logs(ctx context.Context, projectPath, service string, tail int) (string, error) {
	composeFile := filepath.Join(projectPath, "docker-compose.yml")
	if _, err := os.Stat(composeFile); err != nil {
		return "", fmt.Errorf("compose file not found: %s", composeFile)
	}

	if tail <= 0 {
		tail = 100
	}

	exec := getComposeExecutor()
	result, err := exec.Execute(ctx, "compose_logs", map[string]string{
		"compose_file": composeFile,
		"tail":         fmt.Sprintf("%d", tail),
		"service":      service,
	})
	if err != nil {
		return "", fmt.Errorf("compose logs failed: %w", err)
	}

	output := result.Stdout
	if result.Stderr != "" {
		output += "\n" + result.Stderr
	}
	return strings.TrimSpace(output), nil
}

// Ps 获取 compose 项目容器状态（返回原始 JSON 输出）
func Ps(ctx context.Context, projectPath string) (string, error) {
	composeFile := filepath.Join(projectPath, "docker-compose.yml")
	if _, err := os.Stat(composeFile); err != nil {
		return "", fmt.Errorf("compose file not found: %s", composeFile)
	}

	exec := getComposeExecutor()
	result, err := exec.Execute(ctx, "compose_ps", map[string]string{
		"compose_file": composeFile,
	})
	if err != nil {
		return "", fmt.Errorf("compose ps failed: %w", err)
	}
	return result.Stdout, nil
}
