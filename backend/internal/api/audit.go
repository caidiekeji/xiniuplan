package api

import (
	"strconv"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	auditsvc "cloudpivot/internal/service/audit"

	"github.com/gin-gonic/gin"
)

// AuditHandler 审计日志 API Handler
type AuditHandler struct {
	svc *auditsvc.Service
}

// NewAuditHandler 创建审计日志 Handler
func NewAuditHandler(svc *auditsvc.Service) *AuditHandler {
	return &AuditHandler{svc: svc}
}

// RegisterAuditRoutes 注册审计日志路由
func RegisterAuditRoutes(r *gin.RouterGroup) {
	svc := auditsvc.NewService(model.DB)
	h := NewAuditHandler(svc)

	r.GET("/audit/logs", middleware.RequirePermission(model.PermAuditView), h.ListLogs)
	r.GET("/audit/login-logs", middleware.RequirePermission(model.PermAuditView), h.ListLoginLogs)
	r.GET("/audit/export", middleware.RequirePermission(model.PermAuditView), h.Export)
}

// ListLogs GET /audit/logs?user_id=&action=&result=&page=&pageSize= - 操作日志（分页）
func (h *AuditHandler) ListLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	var userID uint
	if uidStr := c.Query("user_id"); uidStr != "" {
		if uid, err := strconv.ParseUint(uidStr, 10, 32); err == nil {
			userID = uint(uid)
		}
	}

	action := c.Query("action")
	result := c.Query("result")

	logs, total, err := h.svc.ListAuditLogs(page, pageSize, userID, action, result)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取审计日志失败: "+err.Error())
		return
	}

	items := make([]dto.AuditLogItem, 0, len(logs))
	for i := range logs {
		items = append(items, auditsvc.ToResponse(&logs[i]))
	}

	common.Page(c, items, total, page, pageSize)
}

// ListLoginLogs GET /audit/login-logs?page=&pageSize= - 登录日志
func (h *AuditHandler) ListLoginLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	logs, total, err := h.svc.ListLoginLogs(page, pageSize)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取登录日志失败: "+err.Error())
		return
	}

	items := make([]dto.AuditLogItem, 0, len(logs))
	for i := range logs {
		items = append(items, auditsvc.ToResponse(&logs[i]))
	}

	common.Page(c, items, total, page, pageSize)
}

// Export GET /audit/export?start=&end= - 导出 CSV（文件流）
func (h *AuditHandler) Export(c *gin.Context) {
	start := c.Query("start")
	end := c.Query("end")

	csvContent, err := h.svc.ExportAuditLogs(start, end)
	if err != nil {
		common.Fail(c, common.CodeParamError, "导出失败: "+err.Error())
		return
	}

	fileName := "audit_logs_" + time.Now().Format("20060102_150405") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.String(200, "\uFEFF"+csvContent) // BOM 确保 Excel 正确识别 UTF-8
}
