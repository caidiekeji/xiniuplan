package api

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	backupsvc "cloudpivot/internal/service/backup"

	"github.com/gin-gonic/gin"
)

// RemoteStorageHandler 远程备份存储 API Handler
type RemoteStorageHandler struct {
	svc *backupsvc.RemoteStorageService
}

// NewRemoteStorageHandler 创建 Handler
func NewRemoteStorageHandler() *RemoteStorageHandler {
	return &RemoteStorageHandler{svc: backupsvc.NewRemoteStorageService(model.DB)}
}

// RegisterRemoteStorageRoutes 注册远程存储路由（挂载在认证组下）
func RegisterRemoteStorageRoutes(r *gin.RouterGroup) {
	h := NewRemoteStorageHandler()

	r.GET("/remote-storages", middleware.RequirePermission(model.PermBackupView), h.List)
	r.POST("/remote-storages", middleware.RequirePermission(model.PermBackupCreate), h.Create)
	r.PUT("/remote-storages/:id", middleware.RequirePermission(model.PermBackupCreate), h.Update)
	r.DELETE("/remote-storages/:id", middleware.RequirePermission(model.PermBackupDelete), h.Delete)
	r.POST("/remote-storages/:id/test", middleware.RequirePermission(model.PermBackupCreate), h.Test)

	// 备份推送
	r.POST("/backups/:id/push", middleware.RequirePermission(model.PermBackupCreate), h.PushBackup)
}

func toRemoteResponse(r *model.RemoteStorage, svc *backupsvc.RemoteStorageService) dto.RemoteStorageResponse {
	return dto.RemoteStorageResponse{
		ID:        r.ID,
		Name:      r.Name,
		Type:      r.Type,
		Config:    svc.MaskedConfig(r),
		Enabled:   r.Enabled,
		CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// List GET /remote-storages
func (h *RemoteStorageHandler) List(c *gin.Context) {
	items, err := h.svc.List()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取远程存储列表失败: "+err.Error())
		return
	}
	result := make([]dto.RemoteStorageResponse, 0, len(items))
	for i := range items {
		result = append(result, toRemoteResponse(&items[i], h.svc))
	}
	common.Success(c, result)
}

// Create POST /remote-storages
func (h *RemoteStorageHandler) Create(c *gin.Context) {
	var req dto.RemoteStorageConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	r, err := h.svc.Create(req.Name, req.Type, req.Config, req.Enabled != nil && *req.Enabled)
	if err != nil {
		common.Fail(c, common.CodeServerError, "创建失败: "+err.Error())
		return
	}
	common.Success(c, toRemoteResponse(r, h.svc))
}

// Update PUT /remote-storages/:id
func (h *RemoteStorageHandler) Update(c *gin.Context) {
	id, ok := parseBackupID(c)
	if !ok {
		return
	}
	var req dto.RemoteStorageConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	r, err := h.svc.Update(id, req.Name, req.Type, req.Config, req.Enabled)
	if err != nil {
		common.Fail(c, common.CodeServerError, "更新失败: "+err.Error())
		return
	}
	common.Success(c, toRemoteResponse(r, h.svc))
}

// Delete DELETE /remote-storages/:id
func (h *RemoteStorageHandler) Delete(c *gin.Context) {
	id, ok := parseBackupID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(id); err != nil {
		common.Fail(c, common.CodeServerError, "删除失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"deleted": true})
}

// Test POST /remote-storages/:id/test
func (h *RemoteStorageHandler) Test(c *gin.Context) {
	id, ok := parseBackupID(c)
	if !ok {
		return
	}
	if err := h.svc.Test(id); err != nil {
		common.Fail(c, common.CodeServerError, "连接测试失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"ok": true})
}

// PushBackup POST /backups/:id/push
func (h *RemoteStorageHandler) PushBackup(c *gin.Context) {
	id, ok := parseBackupID(c)
	if !ok {
		return
	}
	var req dto.PushBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	svc := backupsvc.NewService(model.DB)
	if err := svc.PushBackup(id, req.RemoteID); err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	backup, _ := svc.GetBackup(id)
	common.Success(c, backupsvc.ToResponse(backup))
}

// --- 面板间备份接收端（BK-008 异地容灾） ---

// RegisterBackupReceiveRoutes 注册公开的备份接收路由（用 X-Panel-Token 认证，不走 JWT）
func RegisterBackupReceiveRoutes(r *gin.RouterGroup) {
	r.GET("/backup/receive/ping", receivePing)
	r.POST("/backup/receive", receiveBackup)
}

func checkPanelToken(c *gin.Context) bool {
	token := c.GetHeader("X-Panel-Token")
	if token == "" {
		common.Fail(c, common.CodeUnauthorized, "缺少 X-Panel-Token")
		return false
	}
	var setting model.Setting
	if err := model.DB.Where("`key` = ?", "backup_receive_token").First(&setting).Error; err != nil || setting.Value == "" {
		common.Fail(c, common.CodeForbidden, "本机未启用备份接收功能（未设置 backup_receive_token）")
		return false
	}
	if setting.Value != token {
		common.Fail(c, common.CodeUnauthorized, "X-Panel-Token 无效")
		return false
	}
	return true
}

// receivePing GET /backup/receive/ping
func receivePing(c *gin.Context) {
	if !checkPanelToken(c) {
		return
	}
	common.Success(c, gin.H{"pong": true, "time": time.Now().Format(time.RFC3339)})
}

// receiveBackup POST /backup/receive - 接收对端面板推送的备份文件
func receiveBackup(c *gin.Context) {
	if !checkPanelToken(c) {
		return
	}
	fileName := c.GetHeader("X-Backup-Filename")
	if fileName == "" {
		common.Fail(c, common.CodeParamError, "缺少 X-Backup-Filename")
		return
	}
	// 清洗文件名，防路径穿越
	fileName = filepath.Base(fileName)

	dir := filepath.Join("./backup", "remote")
	if err := os.MkdirAll(dir, 0755); err != nil {
		common.Fail(c, common.CodeServerError, "创建接收目录失败: "+err.Error())
		return
	}
	outPath := filepath.Join(dir, fmt.Sprintf("%d_%s", time.Now().Unix(), fileName))
	f, err := os.Create(outPath)
	if err != nil {
		common.Fail(c, common.CodeServerError, "创建文件失败: "+err.Error())
		return
	}
	defer f.Close()
	if _, err := io.Copy(f, io.LimitReader(c.Request.Body, 20<<30)); err != nil { // 最大 20GB
		common.Fail(c, common.CodeServerError, "写入失败: "+err.Error())
		return
	}

	// 记录到备份列表
	size := int64(0)
	if info, err := os.Stat(outPath); err == nil {
		size = info.Size()
	}
	record := &model.Backup{
		Type:     "remote",
		Target:   c.ClientIP(),
		FilePath: outPath,
		Size:     size,
		Status:   "success",
		Remark:   "异地面板推送",
	}
	model.DB.Create(record)
	common.Success(c, gin.H{"received": true, "backup_id": record.ID})
}
