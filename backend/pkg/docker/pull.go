package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
)

// PullProgress 拉取进度
type PullProgress struct {
	// LayerID 正在下载的层 ID
	LayerID string `json:"layer_id"`
	// Status 当前状态描述，如 "Downloading"、"Extracting"、"Pull complete"
	Status string `json:"status"`
	// Current / Total 当前与总字节数
	Current int64 `json:"current"`
	Total   int64 `json:"total"`
	// Percent 百分比 0-100；Total 为 0 时无法计算，返回 -1
	Percent float64 `json:"percent"`
}

// PullResult 拉取结果
type PullResult struct {
	// ImageRef 最终拉取的镜像引用（可能已改写为加速源地址）
	ImageRef string `json:"image_ref"`
	// Status 完成状态文案
	Status string `json:"status"`
}

// PullImageWithProgress 拉取镜像，逐层上报进度
//
// onProgress 每收到一条进度消息回调一次（可能非常频繁，调用方应自行节流）；
// 拉取完成或出错时返回。
func (d *DockerClient) PullImageWithProgress(
	ctx context.Context,
	imageRef string,
	registryAuth string,
	onProgress func(PullProgress),
) (*PullResult, error) {
	// 注意：这里故意不加默认超时（withTimeout）。镜像拉取是长耗时操作，
	// 时长由调用方 ctx 控制：HTTP 场景下传 c.Request.Context()，客户端断开即取消；
	// 若调用方传 context.Background() 则无超时（service 层注释已声明「不设整体超时」）。
	// 一旦在此套上 30s 默认超时，任何下载超过 30s 的镜像都会以
	// context deadline exceeded 失败——这正是此前拉取功能坏死的根因。

	reader, err := d.cli.ImagePull(ctx, imageRef, ImagePullOptionsWithAuth(registryAuth))
	if err != nil {
		return nil, fmt.Errorf("pull image %s: %w", imageRef, err)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lastStatus := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var msg struct {
			Status         string `json:"status"`
			ProgressDetail *struct {
				Current int64 `json:"current"`
				Total   int64 `json:"total"`
			} `json:"progressDetail"`
			ID    string `json:"id"`
			Error string `json:"error"`
		}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// 非 JSON 行（Docker 偶尔输出纯文本），忽略而非中断
			continue
		}
		if msg.Error != "" {
			return nil, fmt.Errorf("拉取镜像失败: %s", msg.Error)
		}
		if msg.Status != "" {
			lastStatus = msg.Status
		}
		if onProgress == nil {
			continue
		}

		p := PullProgress{
			LayerID: msg.ID,
			Status:  msg.Status,
			Percent: -1,
		}
		if msg.ProgressDetail != nil {
			p.Current = msg.ProgressDetail.Current
			p.Total = msg.ProgressDetail.Total
			if p.Total > 0 {
				p.Percent = float64(p.Current) / float64(p.Total) * 100
			}
		}
		onProgress(p)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取拉取输出失败: %w", err)
	}
	if lastStatus == "" {
		lastStatus = "Pull complete"
	}

	return &PullResult{ImageRef: imageRef, Status: lastStatus}, nil
}

// ImagePullOptionsWithAuth 构造带认证的拉取选项
func ImagePullOptionsWithAuth(registryAuth string) types.ImagePullOptions {
	return types.ImagePullOptions{RegistryAuth: registryAuth}
}
