package docker

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/pkg/stdcopy"
)

// --- 容器终端 exec（C-004） ---

// ExecCreate 创建 exec 实例
func (d *DockerClient) ExecCreate(ctx context.Context, containerID string, cmd []string) (string, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	resp, err := d.cli.ContainerExecCreate(ctx, containerID, types.ExecConfig{
		Tty:          true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd, // 通常 ["/bin/sh"] 或 ["/bin/bash"]
		Env:          []string{"TERM=xterm-256color"},
	})
	if err != nil {
		return "", fmt.Errorf("exec create: %w", err)
	}
	return resp.ID, nil
}

// ExecAttach 附加到 exec 实例，返回可读写的连接
func (d *DockerClient) ExecAttach(ctx context.Context, execID string) (types.HijackedResponse, error) {
	resp, err := d.cli.ContainerExecAttach(ctx, execID, types.ExecStartCheck{Tty: true})
	if err != nil {
		return types.HijackedResponse{}, fmt.Errorf("exec attach: %w", err)
	}
	return resp, nil
}

// ExecResize 调整 exec 终端尺寸
func (d *DockerClient) ExecResize(ctx context.Context, execID string, height, width uint) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	return d.cli.ContainerExecResize(ctx, execID, container.ResizeOptions{
		Height: height,
		Width:  width,
	})
}

// ExecInspectRaw 检查 exec 状态（是否运行中）
func (d *DockerClient) ExecInspectRaw(ctx context.Context, execID string) (bool, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	inspect, err := d.cli.ContainerExecInspect(ctx, execID)
	if err != nil {
		return false, err
	}
	return inspect.Running, nil
}

// demuxStream 将 docker 多路复用流拆为 stdout/stderr（tty 模式无需 demux，这里供非 tty 用途）
func demuxStream(reader io.Reader, stdout, stderr io.Writer) error {
	_, err := stdcopy.StdCopy(stdout, stderr, reader)
	return err
}

// --- 网络管理（C-008） ---

// NetworkInfo Docker 网络信息
type NetworkInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Driver     string   `json:"driver"`
	Scope      string   `json:"scope"`
	Internal   bool     `json:"internal"`
	Subnet     string   `json:"subnet"`
	Gateway    string   `json:"gateway"`
	Containers []string `json:"containers"` // 容器名列表
	BuiltIn    bool     `json:"built_in"`   // bridge/host/none 等内置网络不可删除
}

// ListNetworks 列出网络
func (d *DockerClient) ListNetworks(ctx context.Context) ([]NetworkInfo, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	networks, err := d.cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list networks: %w", err)
	}

	result := make([]NetworkInfo, 0, len(networks))
	for _, n := range networks {
		info := NetworkInfo{
			ID:       n.ID,
			Name:     n.Name,
			Driver:   n.Driver,
			Scope:    n.Scope,
			Internal: n.Internal,
			BuiltIn:  n.Name == "bridge" || n.Name == "host" || n.Name == "none",
		}
		if n.IPAM.Config != nil {
			for _, c := range n.IPAM.Config {
				if info.Subnet == "" {
					info.Subnet = c.Subnet
					info.Gateway = c.Gateway
				}
			}
		}
		// 附带容器列表
		inspect, err := d.cli.NetworkInspect(ctx, n.ID, types.NetworkInspectOptions{})
		if err == nil {
			names := make([]string, 0, len(inspect.Containers))
			for _, cn := range inspect.Containers {
				name := cn.Name
				if name == "" && len(cn.EndpointID) > 8 {
					name = cn.EndpointID[:8]
				}
				names = append(names, name)
			}
			info.Containers = names
		}
		result = append(result, info)
	}
	return result, nil
}

// CreateNetwork 创建自定义网络
func (d *DockerClient) CreateNetwork(ctx context.Context, name, driver, subnet, gateway string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	if driver == "" {
		driver = "bridge"
	}
	opts := types.NetworkCreate{
		Driver: driver,
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{},
		},
	}
	if subnet != "" {
		opts.IPAM.Config = append(opts.IPAM.Config, network.IPAMConfig{Subnet: subnet, Gateway: gateway})
	}
	if _, err := d.cli.NetworkCreate(ctx, name, opts); err != nil {
		return fmt.Errorf("create network %s: %w", name, err)
	}
	return nil
}

// RemoveNetwork 删除网络
func (d *DockerClient) RemoveNetwork(ctx context.Context, id string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	return d.cli.NetworkRemove(ctx, id)
}

// --- 存储卷管理（C-009） ---

// VolumeInfo Docker 存储卷信息
type VolumeInfo struct {
	Name       string `json:"name"`
	Driver     string `json:"driver"`
	Mountpoint string `json:"mountpoint"`
	Size       int64  `json:"size"`
	CreatedAt  string `json:"created_at"`
	InUse      bool   `json:"in_use"`
	UsedBy     string `json:"used_by"`
}

// ListVolumes 列出存储卷及占用情况
func (d *DockerClient) ListVolumes(ctx context.Context) ([]VolumeInfo, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	volList, err := d.cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list volumes: %w", err)
	}

	// 建立容器对卷的占用索引
	containers, err := d.cli.ContainerList(ctx, container.ListOptions{All: true})
	usedBy := map[string][]string{}
	if err == nil {
		for _, c := range containers {
			name := c.ID
			if len(c.Names) > 0 {
				name = strings.TrimPrefix(c.Names[0], "/")
			}
			for _, m := range c.Mounts {
				if m.Type == "volume" {
					usedBy[m.Name] = append(usedBy[m.Name], name)
				}
			}
		}
	}

	result := make([]VolumeInfo, 0, len(volList.Volumes))
	for _, v := range volList.Volumes {
		info := VolumeInfo{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			CreatedAt:  v.CreatedAt,
			UsedBy:     strings.Join(usedBy[v.Name], ", "),
		}
		info.InUse = len(usedBy[v.Name]) > 0
		// 尝试获取用量（需要 linux 内核支持，失败则忽略）
		if vol, _, err := d.cli.VolumeInspectWithRaw(ctx, v.Name); err == nil {
			if vol.UsageData != nil {
				info.Size = vol.UsageData.Size
			}
		}
		result = append(result, info)
	}
	return result, nil
}

// CreateVolume 创建存储卷
func (d *DockerClient) CreateVolume(ctx context.Context, name, driver string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	if driver == "" {
		driver = "local"
	}
	_, err := d.cli.VolumeCreate(ctx, volume.CreateOptions{
		Name:   name,
		Driver: driver,
	})
	if err != nil {
		return fmt.Errorf("create volume %s: %w", name, err)
	}
	return nil
}

// RemoveVolume 删除存储卷
func (d *DockerClient) RemoveVolume(ctx context.Context, name string, force bool) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	return d.cli.VolumeRemove(ctx, name, force)
}

// --- 镜像导入导出（C-006 补全） ---

// SaveImages 导出镜像为 tar 流（docker save）
func (d *DockerClient) SaveImages(ctx context.Context, imageRefs []string) (io.ReadCloser, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	reader, err := d.cli.ImageSave(ctx, imageRefs)
	if err != nil {
		return nil, fmt.Errorf("image save: %w", err)
	}
	return reader, nil
}

// LoadImages 从 tar 流导入镜像（docker load）
func (d *DockerClient) LoadImages(ctx context.Context, tarStream io.Reader) (string, error) {
	resp, err := d.cli.ImageLoad(ctx, tarStream, false)
	if err != nil {
		return "", fmt.Errorf("image load: %w", err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return string(out), nil
}

// RemoveImage 删除镜像
func (d *DockerClient) RemoveImage(ctx context.Context, imageID string, force bool) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	_, err := d.cli.ImageRemove(ctx, imageID, types.ImageRemoveOptions{Force: force})
	return err
}
