package api

import (
	"strconv"

	"cloudpivot/internal/common"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	monitorsvc "cloudpivot/internal/service/monitor"

	"github.com/gin-gonic/gin"
)

// MonitorHandler 监控增强 API Handler
type MonitorHandler struct {
	svc *monitorsvc.Service
}

// NewMonitorHandler 创建监控增强 Handler
func NewMonitorHandler(svc *monitorsvc.Service) *MonitorHandler {
	return &MonitorHandler{svc: svc}
}

// RegisterMonitorRoutes 注册监控增强路由
func RegisterMonitorRoutes(r *gin.RouterGroup) {
	collector := monitorsvc.NewCollector(model.DB, 15)
	svc := monitorsvc.NewService(model.DB, collector)
	h := NewMonitorHandler(svc)

	r.GET("/monitor/processes", middleware.RequirePermission(model.PermMonitorView), h.ListProcesses)
	r.GET("/monitor/disk", middleware.RequirePermission(model.PermMonitorView), h.DiskAnalysis)
	r.GET("/monitor/network", middleware.RequirePermission(model.PermMonitorView), h.NetworkConnections)
	r.GET("/monitor/history", middleware.RequirePermission(model.PermMonitorView), h.History)
}

// ListProcesses GET /monitor/processes?top=20&sort=cpu - 进程列表
func (h *MonitorHandler) ListProcesses(c *gin.Context) {
	top, _ := strconv.Atoi(c.DefaultQuery("top", "20"))
	sortBy := c.DefaultQuery("sort", "cpu")

	processes, err := h.svc.ListProcesses(top, sortBy)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取进程列表失败: "+err.Error())
		return
	}

	common.Success(c, processes)
}

// DiskAnalysis GET /monitor/disk - 磁盘分析
func (h *MonitorHandler) DiskAnalysis(c *gin.Context) {
	usages, err := h.svc.GetDiskAnalysis()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取磁盘分析失败: "+err.Error())
		return
	}

	common.Success(c, usages)
}

// NetworkConnections GET /monitor/network - 网络连接统计
func (h *MonitorHandler) NetworkConnections(c *gin.Context) {
	stats, err := h.svc.GetNetworkConnections()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取网络连接统计失败: "+err.Error())
		return
	}

	common.Success(c, stats)
}

// History GET /monitor/history?type=cpu&range=24h - 历史指标
func (h *MonitorHandler) History(c *gin.Context) {
	metricType := c.DefaultQuery("type", "cpu")
	timeRange := c.DefaultQuery("range", "24h")

	data, err := h.svc.GetMetrics(metricType, timeRange)
	if err != nil {
		common.Fail(c, common.CodeParamError, err.Error())
		return
	}

	common.Success(c, data)
}
