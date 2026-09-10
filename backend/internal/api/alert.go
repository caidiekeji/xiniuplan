package api

import (
	"strconv"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	alertsvc "cloudpivot/internal/service/alert"

	"github.com/gin-gonic/gin"
)

// AlertHandler 告警通知 API Handler
type AlertHandler struct {
	ruleSvc    *alertsvc.RuleService
	channelSvc *alertsvc.ChannelService
	engine     *alertsvc.AlertEngine
}

// NewAlertHandler 创建告警通知 Handler
func NewAlertHandler(ruleSvc *alertsvc.RuleService, channelSvc *alertsvc.ChannelService, engine *alertsvc.AlertEngine) *AlertHandler {
	return &AlertHandler{
		ruleSvc:    ruleSvc,
		channelSvc: channelSvc,
		engine:     engine,
	}
}

// RegisterAlertRoutes 注册告警通知路由
func RegisterAlertRoutes(r *gin.RouterGroup) {
	engine := alertsvc.InitAlertEngine(model.DB)
	ruleSvc := alertsvc.NewRuleService(model.DB, engine)
	channelSvc := alertsvc.NewChannelService(model.DB)
	h := NewAlertHandler(ruleSvc, channelSvc, engine)

	// 规则路由
	r.GET("/alert/rules", middleware.RequirePermission(model.PermAlertView), h.ListRules)
	r.POST("/alert/rules", middleware.RequirePermission(model.PermAlertCreate), h.CreateRule)
	r.GET("/alert/rules/:id", middleware.RequirePermission(model.PermAlertView), h.GetRule)
	r.PUT("/alert/rules/:id", middleware.RequirePermission(model.PermAlertUpdate), h.UpdateRule)
	r.DELETE("/alert/rules/:id", middleware.RequirePermission(model.PermAlertDelete), h.DeleteRule)
	r.POST("/alert/rules/:id/enable", middleware.RequirePermission(model.PermAlertUpdate), h.EnableRule)
	r.POST("/alert/rules/:id/disable", middleware.RequirePermission(model.PermAlertUpdate), h.DisableRule)
	r.POST("/alert/rules/:id/test", middleware.RequirePermission(model.PermAlertUpdate), h.TestRule)

	// 渠道路由
	r.GET("/alert/channels", middleware.RequirePermission(model.PermAlertView), h.ListChannels)
	r.POST("/alert/channels", middleware.RequirePermission(model.PermAlertCreate), h.CreateChannel)
	r.GET("/alert/channels/:id", middleware.RequirePermission(model.PermAlertView), h.GetChannel)
	r.PUT("/alert/channels/:id", middleware.RequirePermission(model.PermAlertUpdate), h.UpdateChannel)
	r.DELETE("/alert/channels/:id", middleware.RequirePermission(model.PermAlertDelete), h.DeleteChannel)
	r.POST("/alert/channels/:id/test", middleware.RequirePermission(model.PermAlertUpdate), h.TestChannel)

	// 历史路由
	r.GET("/alert/history", middleware.RequirePermission(model.PermAlertView), h.ListHistory)
}

// parseAlertID 从路径参数解析 ID
func parseAlertID(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.Fail(c, common.CodeParamError, "无效的 ID")
		return 0, false
	}
	return uint(id), true
}

// ---------------------------------------------------------------------------
// 规则 Handler
// ---------------------------------------------------------------------------

// ListRules GET /alert/rules - 规则列表（分页）
func (h *AlertHandler) ListRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	rules, total, err := h.ruleSvc.ListRules(page, pageSize)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取规则列表失败: "+err.Error())
		return
	}

	items := make([]dto.AlertRuleResponse, 0, len(rules))
	for i := range rules {
		items = append(items, alertsvc.ToRuleResponse(&rules[i]))
	}
	common.Page(c, items, total, page, pageSize)
}

// GetRule GET /alert/rules/:id - 规则详情
func (h *AlertHandler) GetRule(c *gin.Context) {
	id, ok := parseAlertID(c)
	if !ok {
		return
	}
	rule, err := h.ruleSvc.GetRule(id)
	if err != nil {
		common.Fail(c, common.CodeNotFound, err.Error())
		return
	}
	common.Success(c, alertsvc.ToRuleResponse(rule))
}

// CreateRule POST /alert/rules - 创建规则
func (h *AlertHandler) CreateRule(c *gin.Context) {
	var req dto.CreateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	rule, err := h.ruleSvc.CreateRule(req)
	if err != nil {
		common.Fail(c, common.CodeServerError, "创建规则失败: "+err.Error())
		return
	}
	common.Success(c, alertsvc.ToRuleResponse(rule))
}

// UpdateRule PUT /alert/rules/:id - 更新规则
func (h *AlertHandler) UpdateRule(c *gin.Context) {
	id, ok := parseAlertID(c)
	if !ok {
		return
	}
	var req dto.UpdateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	rule, err := h.ruleSvc.UpdateRule(id, req)
	if err != nil {
		common.Fail(c, common.CodeServerError, "更新规则失败: "+err.Error())
		return
	}
	common.Success(c, alertsvc.ToRuleResponse(rule))
}

// DeleteRule DELETE /alert/rules/:id - 删除规则
func (h *AlertHandler) DeleteRule(c *gin.Context) {
	id, ok := parseAlertID(c)
	if !ok {
		return
	}
	if err := h.ruleSvc.DeleteRule(id); err != nil {
		common.Fail(c, common.CodeServerError, "删除规则失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"deleted": true})
}

// EnableRule POST /alert/rules/:id/enable - 启用规则
func (h *AlertHandler) EnableRule(c *gin.Context) {
	id, ok := parseAlertID(c)
	if !ok {
		return
	}
	if err := h.ruleSvc.EnableRule(id); err != nil {
		common.Fail(c, common.CodeServerError, "启用规则失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"enabled": true})
}

// DisableRule POST /alert/rules/:id/disable - 禁用规则
func (h *AlertHandler) DisableRule(c *gin.Context) {
	id, ok := parseAlertID(c)
	if !ok {
		return
	}
	if err := h.ruleSvc.DisableRule(id); err != nil {
		common.Fail(c, common.CodeServerError, "禁用规则失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"disabled": true})
}

// TestRule POST /alert/rules/:id/test - 手动测试规则
func (h *AlertHandler) TestRule(c *gin.Context) {
	id, ok := parseAlertID(c)
	if !ok {
		return
	}
	if err := h.ruleSvc.TestRule(id); err != nil {
		common.Fail(c, common.CodeServerError, "测试规则失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"tested": true})
}

// ---------------------------------------------------------------------------
// 渠道 Handler
// ---------------------------------------------------------------------------

// ListChannels GET /alert/channels - 渠道列表（分页）
func (h *AlertHandler) ListChannels(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	channels, total, err := h.channelSvc.ListChannels(page, pageSize)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取渠道列表失败: "+err.Error())
		return
	}

	items := make([]dto.AlertChannelResponse, 0, len(channels))
	for i := range channels {
		items = append(items, alertsvc.ToChannelResponse(&channels[i]))
	}
	common.Page(c, items, total, page, pageSize)
}

// GetChannel GET /alert/channels/:id - 渠道详情
func (h *AlertHandler) GetChannel(c *gin.Context) {
	id, ok := parseAlertID(c)
	if !ok {
		return
	}
	channel, err := h.channelSvc.GetChannel(id)
	if err != nil {
		common.Fail(c, common.CodeNotFound, err.Error())
		return
	}
	common.Success(c, alertsvc.ToChannelResponse(channel))
}

// CreateChannel POST /alert/channels - 创建渠道
func (h *AlertHandler) CreateChannel(c *gin.Context) {
	var req dto.CreateAlertChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	channel, err := h.channelSvc.CreateChannel(req)
	if err != nil {
		common.Fail(c, common.CodeServerError, "创建渠道失败: "+err.Error())
		return
	}
	common.Success(c, alertsvc.ToChannelResponse(channel))
}

// UpdateChannel PUT /alert/channels/:id - 更新渠道
func (h *AlertHandler) UpdateChannel(c *gin.Context) {
	id, ok := parseAlertID(c)
	if !ok {
		return
	}
	var req dto.UpdateAlertChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	channel, err := h.channelSvc.UpdateChannel(id, req)
	if err != nil {
		common.Fail(c, common.CodeServerError, "更新渠道失败: "+err.Error())
		return
	}
	common.Success(c, alertsvc.ToChannelResponse(channel))
}

// DeleteChannel DELETE /alert/channels/:id - 删除渠道
func (h *AlertHandler) DeleteChannel(c *gin.Context) {
	id, ok := parseAlertID(c)
	if !ok {
		return
	}
	if err := h.channelSvc.DeleteChannel(id); err != nil {
		common.Fail(c, common.CodeServerError, "删除渠道失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"deleted": true})
}

// TestChannel POST /alert/channels/:id/test - 测试渠道
func (h *AlertHandler) TestChannel(c *gin.Context) {
	id, ok := parseAlertID(c)
	if !ok {
		return
	}
	var req dto.TestChannelRequest
	// 允许空 body，使用默认测试消息
	_ = c.ShouldBindJSON(&req)

	if err := h.channelSvc.TestChannel(id, req); err != nil {
		common.Fail(c, common.CodeServerError, "测试渠道失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"sent": true})
}

// ---------------------------------------------------------------------------
// 历史 Handler
// ---------------------------------------------------------------------------

// ListHistory GET /alert/history?rule_id=&status=&page=&pageSize= - 告警历史
func (h *AlertHandler) ListHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	ruleID := c.Query("rule_id")
	status := c.Query("status")

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	query := model.DB.Model(&model.AlertHistory{})
	if ruleID != "" {
		query = query.Where("rule_id = ?", ruleID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		common.Fail(c, common.CodeServerError, "查询告警历史失败: "+err.Error())
		return
	}

	var histories []model.AlertHistory
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&histories).Error; err != nil {
		common.Fail(c, common.CodeServerError, "查询告警历史失败: "+err.Error())
		return
	}

	items := make([]dto.AlertHistoryResponse, 0, len(histories))
	for i := range histories {
		items = append(items, alertsvc.ToHistoryResponse(&histories[i]))
	}
	common.Page(c, items, total, page, pageSize)
}
