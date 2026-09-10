package container

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	dockerpkg "cloudpivot/pkg/docker"

	"github.com/docker/docker/pkg/stdcopy"
)

// Service 容器管理服务
type Service struct {
	client *dockerpkg.DockerClient
}

// NewService 创建容器管理服务
// Docker 不可用时返回 nil client，调用各方法时返回明确错误
func NewService() *Service {
	cli, err := dockerpkg.NewClient()
	if err != nil {
		// Docker 不可用时不 panic，记录错误，后续操作返回明确错误
		return &Service{client: nil}
	}
	return &Service{client: cli}
}

// errDockerUnavailable Docker 不可用错误
var errDockerUnavailable = fmt.Errorf("Docker daemon 不可用，请确认 Docker 已启动并可访问")

// ListContainers 列出所有容器（含停止的）
func (s *Service) ListContainers(all bool) ([]dockerpkg.ContainerInfo, error) {
	if s.client == nil {
		return nil, errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.client.ListContainers(ctx, all)
}

// GetContainer 获取容器详情
func (s *Service) GetContainer(id string) (*dockerpkg.ContainerInfo, error) {
	if s.client == nil {
		return nil, errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.client.GetContainer(ctx, id)
}

// StartContainer 启动容器
func (s *Service) StartContainer(id string) error {
	if s.client == nil {
		return errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.client.StartContainer(ctx, id)
}

// StopContainer 停止容器
func (s *Service) StopContainer(id string) error {
	if s.client == nil {
		return errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.client.StopContainer(ctx, id)
}

// RestartContainer 重启容器
func (s *Service) RestartContainer(id string) error {
	if s.client == nil {
		return errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.client.RestartContainer(ctx, id)
}

// RemoveContainer 删除容器
func (s *Service) RemoveContainer(id string, force bool) error {
	if s.client == nil {
		return errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.client.RemoveContainer(ctx, id, force)
}

// GetContainerLogs 获取容器日志
// follow=true 时返回 io.Reader（持续流），调用方需负责关闭
func (s *Service) GetContainerLogs(id string, tail int, follow bool) (io.ReadCloser, error) {
	if s.client == nil {
		return nil, errDockerUnavailable
	}
	var ctx context.Context
	var cancel context.CancelFunc
	if follow {
		// follow 模式不设超时，由调用方控制生命周期
		ctx = context.Background()
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}
	return s.client.GetContainerLogs(ctx, id, tail, follow)
}

// GetContainerLogsContent 获取容器日志内容（非流式，读取全部）
// Docker 日志是多路复用格式（stdout/stderr 带 8 字节头部），需要用 stdcopy 解析
func (s *Service) GetContainerLogsContent(id string, tail int) (string, error) {
	reader, err := s.GetContainerLogs(id, tail, false)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	// 用 stdcopy 分离 stdout 和 stderr，合并输出
	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, reader); err != nil {
		return "", fmt.Errorf("decode docker logs: %w", err)
	}

	result := stdout.String()
	if stderr.Len() > 0 {
		if result != "" {
			result += "\n"
		}
		result += stderr.String()
	}
	return strings.TrimSpace(result), nil
}

// ListImages 列出镜像
func (s *Service) ListImages() ([]dockerpkg.ImageInfo, error) {
	if s.client == nil {
		return nil, errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.client.ListImages(ctx)
}

// GetContainerStats 获取容器实时 CPU/内存占用
func (s *Service) GetContainerStats(id string) (*dockerpkg.ContainerStats, error) {
	if s.client == nil {
		return nil, errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.client.GetContainerStats(ctx, id)
}

// IsAvailable 检查 Docker 是否可用
func (s *Service) IsAvailable() bool {
	return s.client != nil
}
