package api

import (
	"path/filepath"
	"strconv"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	backupsvc "cloudpivot/internal/service/backup"

	"github.com/gin-gonic/gin"
)

// BackupHandler 备份管理 API Handler
type BackupHandler struct {
	svc *backupsvc.Service
}

// NewBackupHandler 创建备份管理 Handler
func NewBackupHandler(svc *backupsvc.Service) *BackupHandler {
	return &BackupHandler{svc: svc}
}

// RegisterBackupRoutes 注册备份管理路由
func RegisterBackupRoutes(r *gin.RouterGroup) {
	svc := backupsvc.NewService(model.DB)
	h := NewBackupHandler(svc)

	r.GET("/backups", middleware.RequirePermission(model.PermBackupView), h.List)
	r.POST("/backups", middleware.RequirePermission(model.PermBackupCreate), h.Create)
	r.GET("/backups/:id", middleware.RequirePermission(model.PermBackupView), h.Get)
	r.DELETE("/backups/:id", middleware.RequirePermission(model.PermBackupDelete), h.Delete)
	r.POST("/backups/:id/restore", middleware.RequirePermission(model.PermBackupCreate), h.Restore)
	r.GET("/backups/:id/download", middleware.RequirePermission(model.PermBackupView), h.Download)

	// 远程存储配置与推送
	RegisterRemoteStorageRoutes(r)
}

// parseBackupID 从路径参数解析备份 ID
func parseBackupID(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.Fail(c, common.CodeParamError, "无效的备份 ID")
		return 0, false
	}
	return uint(id), true
}

// List GET /backups - 备份列表（分页）
func (h *BackupHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	backupType := c.Query("type")

	backups, total, err := h.svc.ListBackups(page, pageSize, backupType)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取备份列表失败: "+err.Error())
		return
	}

	items := make([]dto.BackupResponse, 0, len(backups))
	for i := range backups {
		items = append(items, backupsvc.ToResponse(&backups[i]))
	}

	common.Page(c, items, total, page, pageSize)
}

// Create POST /backups - 创建备份
func (h *BackupHandler) Create(c *gin.Context) {
	var req dto.CreateBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	backup, err := h.svc.CreateBackup(req)
	if err != nil {
		common.Fail(c, common.CodeServerError, "创建备份失败: "+err.Error())
		return
	}

	common.Success(c, backupsvc.ToResponse(backup))
}

// Get GET /backups/:id - 备份详情
func (h *BackupHandler) Get(c *gin.Context) {
	id, ok := parseBackupID(c)
	if !ok {
		return
	}

	backup, err := h.svc.GetBackup(id)
	if err != nil {
		common.Fail(c, common.CodeNotFound, err.Error())
		return
	}

	common.Success(c, backupsvc.ToResponse(backup))
}

// Delete DELETE /backups/:id - 删除备份
func (h *BackupHandler) Delete(c *gin.Context) {
	id, ok := parseBackupID(c)
	if !ok {
		return
	}

	if err := h.svc.DeleteBackup(id); err != nil {
		common.Fail(c, common.CodeServerError, "删除备份失败: "+err.Error())
		return
	}

	common.Success(c, gin.H{"deleted": true})
}

// Restore POST /backups/:id/restore - 恢复备份
func (h *BackupHandler) Restore(c *gin.Context) {
	id, ok := parseBackupID(c)
	if !ok {
		return
	}

	if err := h.svc.RestoreBackup(id); err != nil {
		common.Fail(c, common.CodeServerError, "恢复备份失败: "+err.Error())
		return
	}

	common.Success(c, gin.H{"restored": true})
}

// Download GET /backups/:id/download - 下载备份文件（文件流）
func (h *BackupHandler) Download(c *gin.Context) {
	id, ok := parseBackupID(c)
	if !ok {
		return
	}

	filePath, err := h.svc.DownloadBackup(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, "下载失败: "+err.Error())
		return
	}

	fileName := filepath.Base(filePath)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/octet-stream")
	c.File(filePath)
}
