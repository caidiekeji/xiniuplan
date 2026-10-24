package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"

	"github.com/docker/docker/api/types"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var execUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// RegisterContainerExtraRoutes 注册容器增强路由：exec 终端 / 网络 / 存储卷 / 镜像导入导出
func RegisterContainerExtraRoutes(r *gin.RouterGroup) {
	h := NewContainerHandler()

	// 容器终端 exec WebSocket
	r.GET("/ws/exec", middleware.RequirePermission(model.PermContainerUpdate), h.ExecWS)

	// 网络管理（C-008）
	r.GET("/networks", middleware.RequirePermission(model.PermContainerView), h.ListNetworks)
	r.POST("/networks", middleware.RequirePermission(model.PermContainerUpdate), h.CreateNetwork)
	r.DELETE("/networks/:id", middleware.RequirePermission(model.PermContainerUpdate), h.RemoveNetwork)

	// 存储卷管理（C-009）
	r.GET("/volumes", middleware.RequirePermission(model.PermContainerView), h.ListVolumes)
	r.POST("/volumes", middleware.RequirePermission(model.PermContainerUpdate), h.CreateVolume)
	r.DELETE("/volumes/:name", middleware.RequirePermission(model.PermContainerUpdate), h.RemoveVolume)

	// 镜像导入导出/删除（C-006 补全）
	r.GET("/images/:id/save", middleware.RequirePermission(model.PermContainerView), h.SaveImage)
	r.POST("/images/load", middleware.RequirePermission(model.PermContainerUpdate), h.LoadImage)
	r.DELETE("/images/:id", middleware.RequirePermission(model.PermContainerUpdate), h.RemoveImage)
}

// ExecWS GET /ws/exec?cid=<container_id>&token=... - 容器内终端（Docker exec）
func (h *ContainerHandler) ExecWS(c *gin.Context) {
	containerID := c.Query("cid")
	if containerID == "" {
		common.Fail(c, common.CodeParamError, "缺少容器 ID (cid)")
		return
	}

	conn, err := execUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 创建 exec（优先 bash，回退 sh）
	execID, err := h.svc.ExecCreate(containerID, []string{"/bin/sh", "-c", "if command -v bash >/dev/null 2>&1; then exec bash; else exec sh; fi"})
	if err != nil {
		writeWSJSON(conn, map[string]string{"type": "error", "message": "创建 exec 失败: " + err.Error()})
		return
	}

	hijacked, err := h.svc.ExecAttach(execID)
	if err != nil {
		writeWSJSON(conn, map[string]string{"type": "error", "message": "附加 exec 失败: " + err.Error()})
		return
	}
	defer hijacked.Close()

	writeWSJSON(conn, map[string]string{"type": "connected", "exec_id": execID[:12]})

	// docker 输出 → WebSocket（tty 模式为纯流，直接转发）
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 8192)
		for {
			n, err := hijacked.Reader.Read(buf)
			if n > 0 {
				if wErr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); wErr != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()

	// WebSocket 输入 → docker
	readErr := make(chan error, 1)
	go func() {
		for {
			msgType, payload, err := conn.ReadMessage()
			if err != nil {
				readErr <- err
				return
			}
			switch msgType {
			case websocket.TextMessage:
				// 控制协议：{"type":"resize","cols":80,"rows":24}
				var ctrl struct {
					Type string `json:"type"`
					Rows uint   `json:"rows"`
					Cols uint   `json:"cols"`
				}
				if json.Unmarshal(payload, &ctrl) == nil && ctrl.Type == "resize" {
					_ = h.svc.ExecResize(execID, ctrl.Rows, ctrl.Cols)
				} else {
					_, _ = hijacked.Conn.Write(payload)
				}
			case websocket.BinaryMessage:
				_, _ = hijacked.Conn.Write(payload)
			}
		}
	}()

	select {
	case <-done:
	case err := <-readErr:
		_ = err
	case <-time.After(12 * time.Hour):
	}
}

func writeWSJSON(conn *websocket.Conn, payload map[string]string) {
	b, _ := json.Marshal(payload)
	_ = conn.WriteMessage(websocket.TextMessage, b)
}

// ListNetworks GET /networks
func (h *ContainerHandler) ListNetworks(c *gin.Context) {
	networks, err := h.svc.ListNetworks()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取网络列表失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"items": networks, "total": len(networks)})
}

// CreateNetwork POST /networks
func (h *ContainerHandler) CreateNetwork(c *gin.Context) {
	var req struct {
		Name    string `json:"name" binding:"required"`
		Driver  string `json:"driver"`
		Subnet  string `json:"subnet"`
		Gateway string `json:"gateway"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.CreateNetwork(req.Name, req.Driver, req.Subnet, req.Gateway); err != nil {
		common.Fail(c, common.CodeServerError, "创建网络失败: "+err.Error())
		return
	}
	common.Success(c, nil)
}

// RemoveNetwork DELETE /networks/:id
func (h *ContainerHandler) RemoveNetwork(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.RemoveNetwork(id); err != nil {
		common.Fail(c, common.CodeServerError, "删除网络失败: "+err.Error())
		return
	}
	common.Success(c, nil)
}

// ListVolumes GET /volumes
func (h *ContainerHandler) ListVolumes(c *gin.Context) {
	volumes, err := h.svc.ListVolumes()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取存储卷列表失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"items": volumes, "total": len(volumes)})
}

// CreateVolume POST /volumes
func (h *ContainerHandler) CreateVolume(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		Driver string `json:"driver"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.CreateVolume(req.Name, req.Driver); err != nil {
		common.Fail(c, common.CodeServerError, "创建存储卷失败: "+err.Error())
		return
	}
	common.Success(c, nil)
}

// RemoveVolume DELETE /volumes/:name?force=true
func (h *ContainerHandler) RemoveVolume(c *gin.Context) {
	name := c.Param("name")
	force := c.DefaultQuery("force", "false") == "true"
	if err := h.svc.RemoveVolume(name, force); err != nil {
		common.Fail(c, common.CodeServerError, "删除存储卷失败: "+err.Error())
		return
	}
	common.Success(c, nil)
}

// SaveImage GET /images/:id/save?ref=image:tag - 导出镜像 tar 流
func (h *ContainerHandler) SaveImage(c *gin.Context) {
	ref := c.Query("ref")
	if ref == "" {
		ref = c.Param("id")
	}
	reader, err := h.svc.SaveImages([]string{ref})
	if err != nil {
		common.Fail(c, common.CodeServerError, "导出镜像失败: "+err.Error())
		return
	}
	defer reader.Close()
	fileName := fmt.Sprintf("image_%s.tar", sanitizeFilePart(ref))
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/x-tar")
	_, _ = io.Copy(c.Writer, reader)
}

// LoadImage POST /images/load - 从上传的 tar 导入镜像
func (h *ContainerHandler) LoadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		common.Fail(c, common.CodeParamError, "请上传镜像 tar 文件 (form field: file)")
		return
	}
	src, err := file.Open()
	if err != nil {
		common.Fail(c, common.CodeServerError, "读取上传文件失败: "+err.Error())
		return
	}
	defer src.Close()
	output, err := h.svc.LoadImages(src)
	if err != nil {
		common.Fail(c, common.CodeServerError, "导入镜像失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"output": output})
}

// RemoveImage DELETE /images/:id?force=true
func (h *ContainerHandler) RemoveImage(c *gin.Context) {
	id := c.Param("id")
	force := c.DefaultQuery("force", "false") == "true"
	if err := h.svc.RemoveImage(id, force); err != nil {
		common.Fail(c, common.CodeServerError, "删除镜像失败: "+err.Error())
		return
	}
	common.Success(c, nil)
}

// sanitizeFilePart 清理镜像引用中的文件名不安全字符
func sanitizeFilePart(ref string) string {
	replacer := strings.NewReplacer("/", "_", ":", "_", "\\", "_", "..", "_")
	return replacer.Replace(ref)
}

var _ = types.HijackedResponse{} // 保留 types 引用
