package container

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// execCommandContext 执行外部命令并合并输出
func execCommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd
}

// ===== Compose 编排（roadmap #7） + 镜像构建（roadmap #8） =====

var stackNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

// ComposeStack 一个 Compose 栈（文件型，无 DB model）：runtime/compose/<name>/docker-compose.yml

// composeDir 返回栈根目录（与面板 nginx runtime 同级）
func composeDir() string {
	dir := os.Getenv("COMPOSE_DIR")
	if dir == "" {
		dir = "./runtime/compose"
	}
	_ = os.MkdirAll(dir, 0755)
	return dir
}

// ComposeStackInfo 栈信息
type ComposeStackInfo struct {
	Name     string `json:"name"`
	Running  int    `json:"running"`
	Total    int    `json:"total"`
	FilePath string `json:"file_path"`
	Content  string `json:"content,omitempty"`
}

// ListComposeStacks 列出所有栈（扫描目录 + docker compose ps 统计）
func (s *Service) ListComposeStacks() ([]ComposeStackInfo, error) {
	list := make([]ComposeStackInfo, 0)
	entries, err := os.ReadDir(composeDir())
	if err != nil {
		return list, nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info := ComposeStackInfo{Name: e.Name(), FilePath: filepath.Join(composeDir(), e.Name(), "docker-compose.yml")}
		if content, err := os.ReadFile(info.FilePath); err == nil {
			info.Content = string(content)
		} else {
			continue // 无 compose 文件的目录跳过
		}
		out, _ := s.composeCmd(e.Name(), "ps", "--format", "json")
		info.Running, info.Total = countComposePS(out)
		list = append(list, info)
	}
	return list, nil
}

// countComposePS 解析 docker compose ps --format json 输出（可能为多行 JSON）
func countComposePS(out string) (running, total int) {
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		total++
		if strings.Contains(line, `"running"`) || strings.Contains(line, `"Running"`) ||
			strings.Contains(line, `"State":"running"`) {
			running++
		}
	}
	return
}

// GetComposeStack 读取单个栈的 compose 内容
func (s *Service) GetComposeStack(name string) (*ComposeStackInfo, error) {
	if !stackNameRe.MatchString(name) {
		return nil, fmt.Errorf("栈名只能包含小写字母、数字、下划线和连字符")
	}
	file := filepath.Join(composeDir(), name, "docker-compose.yml")
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("栈不存在")
	}
	return &ComposeStackInfo{Name: name, FilePath: file, Content: string(content)}, nil
}

// SaveComposeStack 保存（新建或更新）栈
func (s *Service) SaveComposeStack(name, content string) error {
	if !stackNameRe.MatchString(name) {
		return fmt.Errorf("栈名只能包含小写字母、数字、下划线和连字符")
	}
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("compose 内容不能为空")
	}
	dir := filepath.Join(composeDir(), name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(content), 0644)
}

// DeleteComposeStack 删除栈（先 down 再移除文件）
func (s *Service) DeleteComposeStack(name string) (string, error) {
	if !stackNameRe.MatchString(name) {
		return "", fmt.Errorf("非法栈名")
	}
	out, _ := s.composeCmd(name, "down")
	out2, err := s.composeCmd(name, "down", "--rmi", "local", "-v")
	if err == nil {
		out += out2
	}
	_ = os.RemoveAll(filepath.Join(composeDir(), name))
	return out, nil
}

// ComposeAction 对栈执行操作：up | down | restart | logs
func (s *Service) ComposeAction(name, action string) (string, error) {
	if !stackNameRe.MatchString(name) {
		return "", fmt.Errorf("非法栈名")
	}
	switch action {
	case "up":
		return s.composeCmd(name, "up", "-d", "--build")
	case "down":
		return s.composeCmd(name, "down")
	case "restart":
		out1, _ := s.composeCmd(name, "down")
		out2, err := s.composeCmd(name, "up", "-d")
		return out1 + out2, err
	case "logs":
		return s.composeCmd(name, "logs", "--tail", "200")
	default:
		return "", fmt.Errorf("不支持的操作: %s", action)
	}
}

// composeCmd 执行 docker compose 命令（栈文件路径由服务端生成，无用户可控路径注入）
func (s *Service) composeCmd(name, action string, extra ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	args := []string{"compose", "-f", filepath.Join(composeDir(), name, "docker-compose.yml"), action}
	args = append(args, extra...)
	cmd := execCommandContext(ctx, "docker", args...)
	b, err := cmd.CombinedOutput()
	return string(b), err
}

// BuildImage 镜像构建（roadmap #8）：Dockerfile 内容 + tag → 构建流式日志回调
// onLine 回调每收到一行日志触发一次；返回最终错误
func (s *Service) BuildImage(tag, dockerfile string, onLine func(string)) error {
	if strings.TrimSpace(dockerfile) == "" {
		return fmt.Errorf("Dockerfile 内容不能为空")
	}
	if !strings.Contains(tag, ":") {
		tag += ":latest"
	}
	// tag 合法性（防任意参数注入 build API）
	tagRe := regexp.MustCompile(`^[a-z0-9][a-z0-9_/.:-]{0,127}$`)
	if !tagRe.MatchString(tag) {
		return fmt.Errorf("镜像 tag 不合法")
	}

	// 构建上下文：仅包含 Dockerfile 的最小 tar
	ctxDir, err := os.MkdirTemp("", "buildctx-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(ctxDir)
	if err := os.WriteFile(filepath.Join(ctxDir, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
		return err
	}

	tarFile, err := os.CreateTemp("", "buildctx-*.tar")
	if err != nil {
		return err
	}
	defer os.Remove(tarFile.Name())
	tw := tar.NewWriter(tarFile)
	df, _ := os.Open(filepath.Join(ctxDir, "Dockerfile"))
	info, _ := df.Stat()
	hdr, _ := tar.FileInfoHeader(info, "")
	hdr.Name = "Dockerfile"
	tw.WriteHeader(hdr)
	io.Copy(tw, df)
	df.Close()
	tw.Close()
	tarFile.Close()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("连接 Docker 失败: %w", err)
	}
	defer cli.Close()

	tarReader, err := os.Open(tarFile.Name())
	if err != nil {
		return err
	}
	defer tarReader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	resp, err := cli.ImageBuild(ctx, tarReader, types.ImageBuildOptions{
		Tags:        []string{tag},
		Remove:      true,
		ForceRemove: true,
	})
	if err != nil {
		return fmt.Errorf("构建请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 流式读取构建输出（JSON 行），提取 stream 字段回调
	decoder := json.NewDecoder(resp.Body)
	for {
		var msg struct {
			Stream string `json:"stream"`
			Error  string `json:"errorDetail"`
			ErrStr string `json:"error"`
		}
		if err := decoder.Decode(&msg); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("读取构建输出失败: %w", err)
		}
		if msg.ErrStr != "" {
			return fmt.Errorf("构建失败: %s", msg.ErrStr)
		}
		if msg.Stream != "" && onLine != nil {
			onLine(strings.TrimRight(msg.Stream, "\n"))
		}
	}
	return nil
}
