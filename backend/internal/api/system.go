package api

import (
	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	systemsvc "cloudpivot/internal/service/system"

	"github.com/gin-gonic/gin"
)

// SystemHandler 系统设置 API Handler
type SystemHandler struct {
	svc *systemsvc.Service
}

// NewSystemHandler 创建系统设置 Handler
func NewSystemHandler(svc *systemsvc.Service) *SystemHandler {
	return &SystemHandler{svc: svc}
}

// RegisterSystemRoutes 注册系统设置路由
func RegisterSystemRoutes(r *gin.RouterGroup) {
	svc := systemsvc.NewService(model.DB)
	h := NewSystemHandler(svc)

	// 系统设置：需要 setting 权限
	r.GET("/system/settings", middleware.RequirePermission(model.PermSettingView), h.GetSettings)
	r.PUT("/system/settings", middleware.RequirePermission(model.PermSettingUpdate), h.UpdateSettings)
	r.GET("/system/version", middleware.RequirePermission(model.PermSettingView), h.GetVersion)
	r.POST("/system/version/check", middleware.RequirePermission(model.PermSettingView), h.CheckUpdate)
	r.POST("/system/clean", middleware.RequirePermission(model.PermSettingUpdate), h.CleanSystem)
	r.GET("/system/info", middleware.RequirePermission(model.PermSettingView), h.GetSystemInfo)
	r.GET("/system/smtp", middleware.RequirePermission(model.PermSettingView), h.GetSmtpConfig)
	r.PUT("/system/smtp", middleware.RequirePermission(model.PermSettingUpdate), h.UpdateSmtpConfig)
	r.POST("/system/smtp/test", middleware.RequirePermission(model.PermSettingUpdate), h.SendTestEmail)

	// RBAC 用户/角色管理 + TOTP（通过已有路由组注册，无需修改 main.go）
	RegisterUserRoutes(r)
	RegisterRoleRoutes(r)
	RegisterTOTPRoutes(r)

	// DNS 云厂商管理
	RegisterDNSRoutes(r)
}

// GetSettings GET /system/settings - 获取设置
func (h *SystemHandler) GetSettings(c *gin.Context) {
	settings, err := h.svc.GetSettings()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取系统设置失败: "+err.Error())
		return
	}

	common.Success(c, dto.SettingsResponse{Settings: settings})
}

// UpdateSettings PUT /system/settings - 更新设置
func (h *SystemHandler) UpdateSettings(c *gin.Context) {
	var req dto.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.UpdateSettings(req.Settings); err != nil {
		common.Fail(c, common.CodeServerError, "更新系统设置失败: "+err.Error())
		return
	}

	common.Success(c, gin.H{"updated": true})
}

// GetVersion GET /system/version - 版本信息
func (h *SystemHandler) GetVersion(c *gin.Context) {
	version, err := h.svc.GetVersion()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取版本信息失败: "+err.Error())
		return
	}

	common.Success(c, version)
}

// CleanSystem POST /system/clean - 清理工具
func (h *SystemHandler) CleanSystem(c *gin.Context) {
	var req dto.CleanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.CleanSystem(req.CleanType)
	if err != nil {
		common.Fail(c, common.CodeServerError, "清理失败: "+err.Error())
		return
	}

	common.Success(c, gin.H{"clean_type": req.CleanType, "result": result})
}

// GetSystemInfo GET /system/info - 详细系统信息
func (h *SystemHandler) GetSystemInfo(c *gin.Context) {
	info, err := h.svc.GetSystemInfo()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取系统信息失败: "+err.Error())
		return
	}

	common.Success(c, info)
}

// GetSmtpConfig GET /system/smtp - 获取 SMTP 配置
func (h *SystemHandler) GetSmtpConfig(c *gin.Context) {
	config, err := h.svc.GetSmtpConfig()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取 SMTP 配置失败: "+err.Error())
		return
	}
	common.Success(c, config)
}

// UpdateSmtpConfig PUT /system/smtp - 更新 SMTP 配置
func (h *SystemHandler) UpdateSmtpConfig(c *gin.Context) {
	var config map[string]string
	if err := c.ShouldBindJSON(&config); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdateSmtpConfig(config); err != nil {
		common.Fail(c, common.CodeServerError, "更新 SMTP 配置失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"updated": true})
}

// SendTestEmail POST /system/smtp/test - 发送测试邮件
func (h *SystemHandler) SendTestEmail(c *gin.Context) {
	var req struct {
		To string `json:"to" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.SendTestEmail(req.To); err != nil {
		common.Fail(c, common.CodeServerError, "发送测试邮件失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"sent": true})
}

// CheckUpdate POST /system/version/check - 检查版本更新
func (h *SystemHandler) CheckUpdate(c *gin.Context) {
	version, err := h.svc.CheckUpdate()
	if err != nil {
		common.Fail(c, common.CodeServerError, "检查更新失败: "+err.Error())
		return
	}
	common.Success(c, version)
}
