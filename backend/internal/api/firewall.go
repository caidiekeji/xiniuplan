package api

import (
	"strconv"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	firewallsvc "cloudpivot/internal/service/firewall"

	"github.com/gin-gonic/gin"
)

// FirewallHandler 防火墙管理 API Handler
type FirewallHandler struct {
	svc *firewallsvc.Service
}

// NewFirewallHandler 创建防火墙 Handler
func NewFirewallHandler(svc *firewallsvc.Service) *FirewallHandler {
	return &FirewallHandler{svc: svc}
}

// RegisterFirewallRoutes 注册防火墙管理路由
func RegisterFirewallRoutes(r *gin.RouterGroup) {
	svc := firewallsvc.NewService(model.DB)
	h := NewFirewallHandler(svc)

	r.GET("/firewall/rules", middleware.RequirePermission(model.PermFirewallView), h.ListRules)
	r.POST("/firewall/rules", middleware.RequirePermission(model.PermFirewallUpdate), h.AddRule)
	r.DELETE("/firewall/rules/:id", middleware.RequirePermission(model.PermFirewallUpdate), h.DeleteRule)
	r.POST("/firewall/rules/:id/toggle", middleware.RequirePermission(model.PermFirewallUpdate), h.ToggleRule)
	r.GET("/firewall/suggestions", middleware.RequirePermission(model.PermFirewallView), h.Suggestions)
}

// parseFirewallRuleID 从路径参数解析规则 ID
func parseFirewallRuleID(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.Fail(c, common.CodeParamError, "无效的规则 ID")
		return 0, false
	}
	return uint(id), true
}

// ListRules GET /firewall/rules - 规则列表（分页）
func (h *FirewallHandler) ListRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	ruleType := c.Query("type")

	rules, total, err := h.svc.ListRules(page, pageSize, ruleType)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取防火墙规则失败: "+err.Error())
		return
	}

	items := make([]dto.RuleResponse, 0, len(rules))
	for i := range rules {
		items = append(items, firewallsvc.ToResponse(&rules[i]))
	}

	common.Page(c, items, total, page, pageSize)
}

// AddRule POST /firewall/rules - 添加规则
func (h *FirewallHandler) AddRule(c *gin.Context) {
	var req dto.CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	rule, err := h.svc.AddRule(req)
	if err != nil {
		common.Fail(c, common.CodeServerError, "添加防火墙规则失败: "+err.Error())
		return
	}

	common.Success(c, firewallsvc.ToResponse(rule))
}

// DeleteRule DELETE /firewall/rules/:id - 删除规则
func (h *FirewallHandler) DeleteRule(c *gin.Context) {
	id, ok := parseFirewallRuleID(c)
	if !ok {
		return
	}

	if err := h.svc.DeleteRule(id); err != nil {
		common.Fail(c, common.CodeServerError, "删除防火墙规则失败: "+err.Error())
		return
	}

	common.Success(c, gin.H{"deleted": true})
}

// ToggleRule POST /firewall/rules/:id/toggle - 启用/禁用
func (h *FirewallHandler) ToggleRule(c *gin.Context) {
	id, ok := parseFirewallRuleID(c)
	if !ok {
		return
	}

	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	rule, err := h.svc.ToggleRule(id, body.Enabled)
	if err != nil {
		common.Fail(c, common.CodeServerError, "切换规则状态失败: "+err.Error())
		return
	}

	common.Success(c, firewallsvc.ToResponse(rule))
}

// Suggestions GET /firewall/suggestions - 安全建议
func (h *FirewallHandler) Suggestions(c *gin.Context) {
	suggestions, err := h.svc.GetSecuritySuggestions()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取安全建议失败: "+err.Error())
		return
	}

	common.Success(c, suggestions)
}
