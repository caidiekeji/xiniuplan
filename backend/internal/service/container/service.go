package container

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	dockerpkg "cloudpivot/pkg/docker"

	"github.com/docker/docker/api/types"
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
// 返回 cancel 由调用方在读取完成后调用——context 生命周期必须覆盖 reader 的整个读取过程，
// 不能在本函数内 defer cancel（否则函数返回即取消，读取时报 context canceled，2026-10-21 线上复现）
func (s *Service) GetContainerLogs(id string, tail int, follow bool) (io.ReadCloser, context.CancelFunc, error) {
	if s.client == nil {
		return nil, nil, errDockerUnavailable
	}
	var ctx context.Context
	var cancel context.CancelFunc
	if follow {
		// follow 模式不设超时，由调用方控制生命周期
		ctx = context.Background()
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	}
	reader, err := s.client.GetContainerLogs(ctx, id, tail, follow)
	if err != nil && cancel != nil {
		cancel()
	}
	return reader, cancel, err
}

// GetContainerLogsContent 获取容器日志内容（非流式，读取全部）
// Docker 日志是多路复用格式（stdout/stderr 带 8 字节头部），需要用 stdcopy 解析
func (s *Service) GetContainerLogsContent(id string, tail int) (string, error) {
	reader, cancel, err := s.GetContainerLogs(id, tail, false)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	if cancel != nil {
		defer cancel()
	}

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

// --- 容器终端 exec（C-004） ---

// ExecCreate 创建 exec 实例
func (s *Service) ExecCreate(containerID string, cmd []string) (string, error) {
	if s.client == nil {
		return "", errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.client.ExecCreate(ctx, containerID, cmd)
}

// ExecAttach 附加到 exec（长连接，无超时）
func (s *Service) ExecAttach(execID string) (types.HijackedResponse, error) {
	if s.client == nil {
		return types.HijackedResponse{}, errDockerUnavailable
	}
	return s.client.ExecAttach(context.Background(), execID)
}

// ExecResize 调整终端尺寸
func (s *Service) ExecResize(execID string, height, width uint) error {
	if s.client == nil {
		return errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.client.ExecResize(ctx, execID, height, width)
}

// --- 网络管理（C-008） ---

// ListNetworks 列出 Docker 网络
func (s *Service) ListNetworks() ([]dockerpkg.NetworkInfo, error) {
	if s.client == nil {
		return nil, errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.client.ListNetworks(ctx)
}

// CreateNetwork 创建 Docker 网络
func (s *Service) CreateNetwork(name, driver, subnet, gateway string) error {
	if s.client == nil {
		return errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.client.CreateNetwork(ctx, name, driver, subnet, gateway)
}

// RemoveNetwork 删除 Docker 网络
func (s *Service) RemoveNetwork(id string) error {
	if s.client == nil {
		return errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.client.RemoveNetwork(ctx, id)
}

// --- 存储卷管理（C-009） ---

// ListVolumes 列出存储卷
func (s *Service) ListVolumes() ([]dockerpkg.VolumeInfo, error) {
	if s.client == nil {
		return nil, errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.client.ListVolumes(ctx)
}

// CreateVolume 创建存储卷
func (s *Service) CreateVolume(name, driver string) error {
	if s.client == nil {
		return errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.client.CreateVolume(ctx, name, driver)
}

// RemoveVolume 删除存储卷
func (s *Service) RemoveVolume(name string, force bool) error {
	if s.client == nil {
		return errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.client.RemoveVolume(ctx, name, force)
}

// --- 镜像导入导出/删除（C-006 补全） ---

// SaveImages 导出镜像 tar 流
func (s *Service) SaveImages(refs []string) (io.ReadCloser, error) {
	if s.client == nil {
		return nil, errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	return s.client.SaveImages(ctx, refs)
}

// LoadImages 导入镜像
func (s *Service) LoadImages(r io.Reader) (string, error) {
	if s.client == nil {
		return "", errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	return s.client.LoadImages(ctx, r)
}

// RemoveImage 删除镜像
func (s *Service) RemoveImage(imageID string, force bool) error {
	if s.client == nil {
		return errDockerUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	return s.client.RemoveImage(ctx, imageID, force)
}
