package api

import (
	"bufio"
	"io"
	"strconv"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	containersvc "cloudpivot/internal/service/container"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gin-gonic/gin"
)

// ContainerHandler 容器管理 API Handler
type ContainerHandler struct {
	svc *containersvc.Service
}

// NewContainerHandler 创建容器管理 Handler
func NewContainerHandler() *ContainerHandler {
	return &ContainerHandler{
		svc: containersvc.NewService(),
	}
}

// RegisterContainerRoutes 注册容器管理路由
func RegisterContainerRoutes(r *gin.RouterGroup) {
	h := NewContainerHandler()

	r.GET("/containers", middleware.RequirePermission(model.PermContainerView), h.ListContainers)
	r.GET("/containers/:id", middleware.RequirePermission(model.PermContainerView), h.GetContainer)
	r.POST("/containers/:id/start", middleware.RequirePermission(model.PermContainerUpdate), h.StartContainer)
	r.POST("/containers/:id/stop", middleware.RequirePermission(model.PermContainerUpdate), h.StopContainer)
	r.POST("/containers/:id/restart", middleware.RequirePermission(model.PermContainerUpdate), h.RestartContainer)
	r.DELETE("/containers/:id", middleware.RequirePermission(model.PermContainerDelete), h.RemoveContainer)
	r.GET("/containers/:id/logs", middleware.RequirePermission(model.PermContainerView), h.GetContainerLogs)
	r.GET("/containers/:id/stats", middleware.RequirePermission(model.PermContainerView), h.GetContainerStats)
	r.GET("/images", middleware.RequirePermission(model.PermContainerView), h.ListImages)

	// 容器增强：exec 终端 / 网络 / 存储卷 / 镜像导入导出
	RegisterContainerExtraRoutes(r)
}

// ListContainers GET /containers?all=true
func (h *ContainerHandler) ListContainers(c *gin.Context) {
	all := c.DefaultQuery("all", "true") == "true"

	containers, err := h.svc.ListContainers(all)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取容器列表失败: "+err.Error())
		return
	}

	responses := make([]dto.ContainerResponse, 0, len(containers))
	for _, ct := range containers {
		responses = append(responses, dto.ContainerResponse{
			ID:      ct.ID,
			Name:    ct.Name,
			Image:   ct.Image,
			Status:  ct.Status,
			Ports:   ct.Ports,
			Created: ct.Created,
			State:   ct.State,
			Labels:  ct.Labels,
		})
	}
	common.Success(c, gin.H{
		"items": responses,
		"total": len(responses),
	})
}

// GetContainer GET /containers/:id
func (h *ContainerHandler) GetContainer(c *gin.Context) {
	id := c.Param("id")

	ct, err := h.svc.GetContainer(id)
	if err != nil {
		common.Fail(c, common.CodeNotFound, "获取容器详情失败: "+err.Error())
		return
	}

	common.Success(c, dto.ContainerResponse{
		ID:      ct.ID,
		Name:    ct.Name,
		Image:   ct.Image,
		Status:  ct.Status,
		Ports:   ct.Ports,
		Created: ct.Created,
		State:   ct.State,
		Labels:  ct.Labels,
	})
}

// StartContainer POST /containers/:id/start
func (h *ContainerHandler) StartContainer(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.StartContainer(id); err != nil {
		common.Fail(c, common.CodeServerError, "启动容器失败: "+err.Error())
		return
	}
	common.Success(c, nil)
}

// StopContainer POST /containers/:id/stop
func (h *ContainerHandler) StopContainer(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.StopContainer(id); err != nil {
		common.Fail(c, common.CodeServerError, "停止容器失败: "+err.Error())
		return
	}
	common.Success(c, nil)
}

// RestartContainer POST /containers/:id/restart
func (h *ContainerHandler) RestartContainer(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.RestartContainer(id); err != nil {
		common.Fail(c, common.CodeServerError, "重启容器失败: "+err.Error())
		return
	}
	common.Success(c, nil)
}

// RemoveContainer DELETE /containers/:id?force=true
func (h *ContainerHandler) RemoveContainer(c *gin.Context) {
	id := c.Param("id")
	force := c.DefaultQuery("force", "false") == "true"

	if err := h.svc.RemoveContainer(id, force); err != nil {
		common.Fail(c, common.CodeServerError, "删除容器失败: "+err.Error())
		return
	}
	common.Success(c, nil)
}

// GetContainerLogs GET /containers/:id/logs?tail=100&follow=false
func (h *ContainerHandler) GetContainerLogs(c *gin.Context) {
	id := c.Param("id")
	tail, _ := strconv.Atoi(c.DefaultQuery("tail", "100"))
	if tail <= 0 {
		tail = 100
	}
	follow := c.DefaultQuery("follow", "false") == "true"

	if !follow {
		// 非流式：读取全部日志返回
		content, err := h.svc.GetContainerLogsContent(id, tail)
		if err != nil {
			common.Fail(c, common.CodeServerError, "获取日志失败: "+err.Error())
			return
		}
		common.Success(c, dto.ContainerLogResponse{Content: content})
		return
	}

	// 流式：通过 SSE 推送日志
	reader, cancel, err := h.svc.GetContainerLogs(id, tail, true)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取日志流失败: "+err.Error())
		return
	}
	defer reader.Close()
	if cancel != nil {
		defer cancel()
	}

	// Docker 日志是多路复用格式，用 stdcopy + pipe 解析后再逐行推送
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		stdcopy.StdCopy(pw, pw, reader)
	}()
	defer pr.Close()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, ok := c.Writer.(gin.ResponseWriter)
	if !ok {
		common.Fail(c, common.CodeServerError, "streaming not supported")
		return
	}

	scanner := bufio.NewScanner(pr)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		c.Writer.WriteString("data: " + line + "\n\n")
		flusher.Flush()
	}
}

// GetContainerStats GET /containers/:id/stats
func (h *ContainerHandler) GetContainerStats(c *gin.Context) {
	id := c.Param("id")
	stats, err := h.svc.GetContainerStats(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取容器状态失败: "+err.Error())
		return
	}
	common.Success(c, dto.ContainerStatsResponse{
		CPUPercent:    stats.CPUPercent,
		MemoryUsage:   stats.MemoryUsage,
		MemoryLimit:   stats.MemoryLimit,
		MemoryPercent: stats.MemoryPercent,
	})
}

// ListImages GET /images
func (h *ContainerHandler) ListImages(c *gin.Context) {
	images, err := h.svc.ListImages()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取镜像列表失败: "+err.Error())
		return
	}

	responses := make([]dto.ImageResponse, 0, len(images))
	for _, img := range images {
		responses = append(responses, dto.ImageResponse{
			ID:      img.ID,
			Tags:    img.Tags,
			Size:    img.Size,
			Created: img.Created,
		})
	}
	common.Success(c, responses)
}
