package api

import (
	"strconv"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	cronsvc "cloudpivot/internal/service/cron"

	"github.com/gin-gonic/gin"
)

// CronHandler 计划任务 API Handler
type CronHandler struct {
	svc *cronsvc.Service
}

// NewCronHandler 创建计划任务 Handler
func NewCronHandler(svc *cronsvc.Service) *CronHandler {
	return &CronHandler{svc: svc}
}

// RegisterCronRoutes 注册计划任务路由
func RegisterCronRoutes(r *gin.RouterGroup) {
	svc := cronsvc.NewService(model.DB)
	h := NewCronHandler(svc)

	r.GET("/cron/tasks", middleware.RequirePermission(model.PermCronView), h.List)
	r.POST("/cron/tasks", middleware.RequirePermission(model.PermCronCreate), h.Create)
	r.PUT("/cron/tasks/:id", middleware.RequirePermission(model.PermCronUpdate), h.Update)
	r.DELETE("/cron/tasks/:id", middleware.RequirePermission(model.PermCronDelete), h.Delete)
	r.POST("/cron/tasks/:id/enable", middleware.RequirePermission(model.PermCronUpdate), h.Enable)
	r.POST("/cron/tasks/:id/disable", middleware.RequirePermission(model.PermCronUpdate), h.Disable)
	r.POST("/cron/tasks/:id/trigger", middleware.RequirePermission(model.PermCronUpdate), h.Trigger)
	r.GET("/cron/tasks/:id/logs", middleware.RequirePermission(model.PermCronView), h.Logs)
}

// parseCronTaskID 从路径参数解析任务 ID
func parseCronTaskID(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.Fail(c, common.CodeParamError, "无效的任务 ID")
		return 0, false
	}
	return uint(id), true
}

// List GET /cron/tasks - 任务列表（分页）
func (h *CronHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	tasks, total, err := h.svc.ListTasks(page, pageSize)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取任务列表失败: "+err.Error())
		return
	}

	items := make([]dto.CronTaskResponse, 0, len(tasks))
	for i := range tasks {
		items = append(items, cronsvc.ToTaskResponse(&tasks[i]))
	}

	common.Page(c, items, total, page, pageSize)
}

// Create POST /cron/tasks - 创建任务
func (h *CronHandler) Create(c *gin.Context) {
	var req dto.CreateCronRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	task, err := h.svc.CreateTask(req)
	if err != nil {
		common.Fail(c, common.CodeServerError, "创建任务失败: "+err.Error())
		return
	}

	common.Success(c, cronsvc.ToTaskResponse(task))
}

// Update PUT /cron/tasks/:id - 更新任务
func (h *CronHandler) Update(c *gin.Context) {
	id, ok := parseCronTaskID(c)
	if !ok {
		return
	}

	var req dto.UpdateCronRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	task, err := h.svc.UpdateTask(id, req)
	if err != nil {
		common.Fail(c, common.CodeServerError, "更新任务失败: "+err.Error())
		return
	}

	common.Success(c, cronsvc.ToTaskResponse(task))
}

// Delete DELETE /cron/tasks/:id - 删除任务
func (h *CronHandler) Delete(c *gin.Context) {
	id, ok := parseCronTaskID(c)
	if !ok {
		return
	}

	if err := h.svc.DeleteTask(id); err != nil {
		common.Fail(c, common.CodeServerError, "删除任务失败: "+err.Error())
		return
	}

	common.Success(c, gin.H{"deleted": true})
}

// Enable POST /cron/tasks/:id/enable - 启用任务
func (h *CronHandler) Enable(c *gin.Context) {
	id, ok := parseCronTaskID(c)
	if !ok {
		return
	}

	task, err := h.svc.EnableTask(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, "启用任务失败: "+err.Error())
		return
	}

	common.Success(c, cronsvc.ToTaskResponse(task))
}

// Disable POST /cron/tasks/:id/disable - 禁用任务
func (h *CronHandler) Disable(c *gin.Context) {
	id, ok := parseCronTaskID(c)
	if !ok {
		return
	}

	task, err := h.svc.DisableTask(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, "禁用任务失败: "+err.Error())
		return
	}

	common.Success(c, cronsvc.ToTaskResponse(task))
}

// Trigger POST /cron/tasks/:id/trigger - 手动触发
func (h *CronHandler) Trigger(c *gin.Context) {
	id, ok := parseCronTaskID(c)
	if !ok {
		return
	}

	if err := h.svc.TriggerTask(id); err != nil {
		common.Fail(c, common.CodeServerError, "触发任务失败: "+err.Error())
		return
	}

	common.Success(c, gin.H{"triggered": true})
}

// Logs GET /cron/tasks/:id/logs - 执行日志（分页）
func (h *CronHandler) Logs(c *gin.Context) {
	id, ok := parseCronTaskID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// 获取任务名称
	task, err := h.svc.GetTask(id)
	if err != nil {
		common.Fail(c, common.CodeNotFound, err.Error())
		return
	}

	logs, total, err := h.svc.ListLogs(id, page, pageSize)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取日志失败: "+err.Error())
		return
	}

	items := make([]dto.CronLogResponse, 0, len(logs))
	for i := range logs {
		items = append(items, cronsvc.ToLogResponse(&logs[i], task.Name))
	}

	common.Page(c, items, total, page, pageSize)
}
