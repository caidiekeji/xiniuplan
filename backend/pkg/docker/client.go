package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// 默认超时
const DefaultTimeout = 30 * time.Second

// DockerClient Docker SDK 封装
type DockerClient struct {
	cli *client.Client
}

// ContainerInfo 容器信息（实时从 Docker 获取，不存 DB）
type ContainerInfo struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Status  string            `json:"status"`
	State   string            `json:"state"`
	Ports   string            `json:"ports"`
	Created int64             `json:"created"`
	Labels  map[string]string `json:"labels"`
}

// ImageInfo 镜像信息
type ImageInfo struct {
	ID      string   `json:"id"`
	Tags    []string `json:"tags"`
	Size    int64    `json:"size"`
	Created int64    `json:"created"`
}

// ContainerStats 容器实时资源占用
type ContainerStats struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryUsage   uint64  `json:"memory_usage"`
	MemoryLimit   uint64  `json:"memory_limit"`
	MemoryPercent float64 `json:"memory_percent"`
}

// NewClient 创建 Docker SDK 客户端
// 环境变量 DOCKER_HOST 指定连接地址，默认 unix:///var/run/docker.sock
func NewClient() (*DockerClient, error) {
	host := os.Getenv("DOCKER_HOST")
	if host == "" {
		host = "unix:///var/run/docker.sock"
	}

	cli, err := client.NewClientWithOpts(
		client.WithHost(host),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}

	// 探测 Docker 是否可用
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := cli.Ping(ctx); err != nil {
		return nil, fmt.Errorf("docker daemon unavailable: %w", err)
	}

	return &DockerClient{cli: cli}, nil
}

// withTimeout 创建带默认超时的 context
func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		return context.WithTimeout(ctx, DefaultTimeout)
	}
	return ctx, func() {}
}

// ListContainers 列出容器
func (d *DockerClient) ListContainers(ctx context.Context, all bool) ([]ContainerInfo, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	containers, err := d.cli.ContainerList(ctx, container.ListOptions{All: all})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	result := make([]ContainerInfo, 0, len(containers))
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}
		result = append(result, ContainerInfo{
			ID:      c.ID,
			Name:    name,
			Image:   c.Image,
			Status:  c.Status,
			State:   c.State,
			Ports:   formatPorts(c.Ports),
			Created: c.Created,
			Labels:  c.Labels,
		})
	}
	return result, nil
}

// GetContainer 获取容器详情
func (d *DockerClient) GetContainer(ctx context.Context, id string) (*ContainerInfo, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	inspect, err := d.cli.ContainerInspect(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("inspect container %s: %w", id, err)
	}

	name := inspect.Name
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}

	state := ""
	if inspect.State != nil {
		state = inspect.State.Status
	}

	created := int64(0)
	if inspect.Created != "" {
		if t, err := time.Parse(time.RFC3339Nano, inspect.Created); err == nil {
			created = t.Unix()
		}
	}

	return &ContainerInfo{
		ID:      inspect.ID,
		Name:    name,
		Image:   inspect.Config.Image,
		Status:  state,
		State:   state,
		Ports:   formatPortBindings(inspect.HostConfig.PortBindings),
		Created: created,
		Labels:  inspect.Config.Labels,
	}, nil
}

// StartContainer 启动容器
func (d *DockerClient) StartContainer(ctx context.Context, id string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	if err := d.cli.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		return fmt.Errorf("start container %s: %w", id, err)
	}
	return nil
}

// StopContainer 停止容器
func (d *DockerClient) StopContainer(ctx context.Context, id string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	timeout := 30
	if err := d.cli.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("stop container %s: %w", id, err)
	}
	return nil
}

// RestartContainer 重启容器
func (d *DockerClient) RestartContainer(ctx context.Context, id string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	timeout := 30
	if err := d.cli.ContainerRestart(ctx, id, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("restart container %s: %w", id, err)
	}
	return nil
}

// RemoveContainer 删除容器
func (d *DockerClient) RemoveContainer(ctx context.Context, id string, force bool) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	if err := d.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: force}); err != nil {
		return fmt.Errorf("remove container %s: %w", id, err)
	}
	return nil
}

// GetContainerLogs 获取容器日志
// follow=true 时返回持续流，调用方需负责关闭
func (d *DockerClient) GetContainerLogs(ctx context.Context, id string, tail int, follow bool) (io.ReadCloser, error) {
	opts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       fmt.Sprintf("%d", tail),
	}

	reader, err := d.cli.ContainerLogs(ctx, id, opts)
	if err != nil {
		return nil, fmt.Errorf("get container logs %s: %w", id, err)
	}
	return reader, nil
}

// ListImages 列出镜像
func (d *DockerClient) ListImages(ctx context.Context) ([]ImageInfo, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	images, err := d.cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}

	result := make([]ImageInfo, 0, len(images))
	for _, img := range images {
		result = append(result, ImageInfo{
			ID:      img.ID,
			Tags:    img.RepoTags,
			Size:    img.Size,
			Created: img.Created,
		})
	}
	return result, nil
}

// PullImage 拉取镜像
func (d *DockerClient) PullImage(ctx context.Context, ref string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	reader, err := d.cli.ImagePull(ctx, ref, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("pull image %s: %w", ref, err)
	}
	defer reader.Close()

	// 读取完拉取输出（丢弃），确保拉取完成
	if _, err := io.Copy(io.Discard, reader); err != nil {
		return fmt.Errorf("wait pull image %s: %w", ref, err)
	}
	return nil
}

// GetContainerStats 获取容器实时 CPU/内存占用
func (d *DockerClient) GetContainerStats(ctx context.Context, id string) (*ContainerStats, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	stats, err := d.cli.ContainerStatsOneShot(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get container stats %s: %w", id, err)
	}
	defer stats.Body.Close()

	// 解析 stats JSON
	var v struct {
		CPUStats struct {
			CPUUsage struct {
				TotalUsage  uint64   `json:"total_usage"`
				PercpuUsage []uint64 `json:"percpu_usage"`
			} `json:"cpu_usage"`
			SystemCPUUsage uint64 `json:"system_cpu_usage"`
			OnlineCPUs     uint32 `json:"online_cpus"`
		} `json:"cpu_stats"`
		PreCPUStats struct {
			CPUUsage struct {
				TotalUsage uint64 `json:"total_usage"`
			} `json:"cpu_usage"`
			SystemCPUUsage uint64 `json:"system_cpu_usage"`
		} `json:"precpu_stats"`
		MemoryStats struct {
			Usage uint64 `json:"usage"`
			Limit uint64 `json:"limit"`
		} `json:"memory_stats"`
	}

	// 使用 json decoder 从 body 读取
	dec := json.NewDecoder(stats.Body)
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("decode stats: %w", err)
	}

	cpuPercent := calculateCPUPercent(v.PreCPUStats.CPUUsage.TotalUsage, v.PreCPUStats.SystemCPUUsage,
		v.CPUStats.CPUUsage.TotalUsage, v.CPUStats.SystemCPUUsage, v.CPUStats.OnlineCPUs)

	memPercent := 0.0
	if v.MemoryStats.Limit > 0 {
		memPercent = float64(v.MemoryStats.Usage) / float64(v.MemoryStats.Limit) * 100.0
	}

	return &ContainerStats{
		CPUPercent:    cpuPercent,
		MemoryUsage:   v.MemoryStats.Usage,
		MemoryLimit:   v.MemoryStats.Limit,
		MemoryPercent: memPercent,
	}, nil
}

// calculateCPUPercent 计算 CPU 使用率百分比
func calculateCPUPercent(previousCPU, previousSystem, currentCPU, currentSystem uint64, onlineCPUs uint32) float64 {
	cpuDelta := float64(currentCPU - previousCPU)
	systemDelta := float64(currentSystem - previousSystem)
	cpuPercent := 0.0
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(onlineCPUs) * 100.0
	}
	return cpuPercent
}

// formatPorts 格式化端口映射列表为字符串
func formatPorts(ports []types.Port) string {
	if len(ports) == 0 {
		return ""
	}
	parts := make([]string, 0, len(ports))
	for _, p := range ports {
		part := fmt.Sprintf("%d/%s", p.PrivatePort, p.Type)
		if p.PublicPort > 0 {
			part = fmt.Sprintf("%d->%s", p.PublicPort, part)
		}
		if p.IP != "" {
			part = p.IP + ":" + part
		}
		parts = append(parts, part)
	}
	return joinStrings(parts, ", ")
}

// formatPortBindings 格式化端口绑定为字符串
func formatPortBindings(bindings nat.PortMap) string {
	if len(bindings) == 0 {
		return ""
	}
	parts := make([]string, 0, len(bindings))
	for port, bindingList := range bindings {
		for _, b := range bindingList {
			part := fmt.Sprintf("%s:%s->%s", b.HostIP, b.HostPort, port.Port())
			parts = append(parts, part)
		}
	}
	return joinStrings(parts, ", ")
}

// joinStrings 简单字符串拼接（避免引入 strings 包重复）
func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}
