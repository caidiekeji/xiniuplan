package api

import (
	"cloudpivot/internal/common"
	monitorsvc "cloudpivot/internal/service/monitor"

	"github.com/gin-gonic/gin"
)

// DashboardHandler Dashboard API Handler
type DashboardHandler struct {
	svc *monitorsvc.Service
}

// NewDashboardHandler 创建 Dashboard Handler
func NewDashboardHandler(svc *monitorsvc.Service) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// Overview GET /api/v1/dashboard/overview
func (h *DashboardHandler) Overview(c *gin.Context) {
	data, err := h.svc.GetOverview(c.Request.Context())
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取资源概览失败: "+err.Error())
		return
	}
	common.Success(c, data)
}

// Metrics GET /api/v1/dashboard/metrics?type=cpu|memory|disk|network&range=1h|6h|24h|7d
func (h *DashboardHandler) Metrics(c *gin.Context) {
	metricType := c.DefaultQuery("type", "cpu")
	timeRange := c.DefaultQuery("range", "24h")

	data, err := h.svc.GetMetrics(metricType, timeRange)
	if err != nil {
		common.Fail(c, common.CodeParamError, err.Error())
		return
	}
	common.Success(c, data)
}

// AuditLogs GET /api/v1/dashboard/audit-logs
func (h *DashboardHandler) AuditLogs(c *gin.Context) {
	logs, err := h.svc.GetRecentAuditLogs()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取审计日志失败: "+err.Error())
		return
	}
	common.Success(c, logs)
}
