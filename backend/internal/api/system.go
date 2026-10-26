package api

import (
	"fmt"
	"strconv"

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

	// 安全评分（D-007，Dashboard 展示）
	r.GET("/dashboard/security-score", middleware.RequirePermission(model.PermMonitorView), h.GetSecurityScore)

	// 系统安全更新（SY-005）与 NTP（SY-006）
	r.GET("/system/updates", middleware.RequirePermission(model.PermSettingView), h.CheckSystemUpdates)
	r.POST("/system/updates/install", middleware.RequirePermission(model.PermSettingUpdate), h.InstallSystemUpdates)
	r.GET("/system/ntp", middleware.RequirePermission(model.PermSettingView), h.GetNTPStatus)
	r.PUT("/system/ntp", middleware.RequirePermission(model.PermSettingUpdate), h.SetNTPStatus)

	// RBAC 用户/角色管理 + TOTP（通过已有路由组注册，无需修改 main.go）
	RegisterUserRoutes(r)
	RegisterRoleRoutes(r)
	RegisterTOTPRoutes(r)

	// API Key（U-007）+ 在线会话管理（U-006）
	RegisterAPIKeyRoutes(r)

	// DNS 云厂商管理
	RegisterDNSRoutes(r)
}

// sensitiveSettingKeys 这些键是密码类，GET 响应中一律脱敏，避免明文泄漏。
// 任何持有 PermSettingView 的账号（RBAC 下可授给非管理员）都能读设置，
// 因此绝不能回传 smtp_pass / docker_registry_pass 真值。
var sensitiveSettingKeys = map[string]bool{
	"smtp_pass":            true,
	"docker_registry_pass": true,
}

// settingSentinel GET 响应中密码字段的占位值。前端用它区分「用户未改动密码」，
// PUT 时原样回传此值则被 stripSentinelSettings 剔除，不覆盖存储的真实密码。
const settingSentinel = "********"

// maskSensitiveSettings 脱敏密码类字段：非空密码替换为哨兵值，空值保持不变
// （空值表示该项未配置，前端据此判断是否需要填写）。
func maskSensitiveSettings(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		if sensitiveSettingKeys[k] && v != "" {
			out[k] = settingSentinel
			continue
		}
		out[k] = v
	}
	return out
}

// stripSentinelSettings 剔除回传的哨兵值，避免把存储的真实密码覆盖成占位串。
func stripSentinelSettings(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		if sensitiveSettingKeys[k] && v == settingSentinel {
			continue
		}
		out[k] = v
	}
	return out
}

// GetSettings GET /system/settings - 获取设置
func (h *SystemHandler) GetSettings(c *gin.Context) {
	settings, err := h.svc.GetSettings()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取系统设置失败: "+err.Error())
		return
	}

	common.Success(c, dto.SettingsResponse{Settings: maskSensitiveSettings(settings)})
}

// UpdateSettings PUT /system/settings - 更新设置
func (h *SystemHandler) UpdateSettings(c *gin.Context) {
	var req dto.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.UpdateSettings(stripSentinelSettings(req.Settings)); err != nil {
		common.Fail(c, common.CodeServerError, "更新系统设置失败: "+err.Error())
		return
	}

	// 面板端口真生效（方案 A）：panel_port 变更时写回 config.yaml，重启面板后监听新端口
	portChanged, portMsg := false, ""
	if pp, ok := req.Settings["panel_port"]; ok {
		if newPort, err := strconv.Atoi(pp); err == nil && newPort > 0 {
			current, changed, err := h.svc.UpdatePanelPort(newPort)
			if err != nil {
				common.Fail(c, common.CodeServerError, "端口修改失败: "+err.Error())
				return
			}
			if changed {
				portChanged = true
				portMsg = fmt.Sprintf("面板端口将由 %d 改为 %d，重启面板（systemctl restart rhinopanel）后生效", current, newPort)
			}
		}
	}

	common.Success(c, gin.H{"updated": true, "port_changed": portChanged, "port_message": portMsg})
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
	common.Success(c, maskSensitiveSettings(config))
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

// CheckSystemUpdates GET /system/updates - 检查系统安全更新（SY-005）
func (h *SystemHandler) CheckSystemUpdates(c *gin.Context) {
	updates, pm, err := h.svc.CheckSystemUpdates()
	if err != nil {
		common.Fail(c, common.CodeServerError, "检查系统更新失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{
		"package_manager": pm,
		"items":           updates,
		"total":           len(updates),
	})
}

// InstallSystemUpdates POST /system/updates/install - 一键安装系统更新（SY-005，需 confirm）
func (h *SystemHandler) InstallSystemUpdates(c *gin.Context) {
	var req struct {
		Confirm bool `json:"confirm"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	output, err := h.svc.InstallSystemUpdates(req.Confirm)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, gin.H{"installed": true, "output": output})
}

// GetNTPStatus GET /system/ntp - NTP 状态（SY-006）
func (h *SystemHandler) GetNTPStatus(c *gin.Context) {
	status, err := h.svc.GetNTPStatus()
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, status)
}

// SetNTPStatus PUT /system/ntp - 开关 NTP 同步（SY-006）
func (h *SystemHandler) SetNTPStatus(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.SetNTPEnabled(req.Enabled); err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, gin.H{"enabled": req.Enabled})
}

// GetSecurityScore GET /dashboard/security-score - 安全评分（D-007）
func (h *SystemHandler) GetSecurityScore(c *gin.Context) {
	score, err := h.svc.GetSecurityScore()
	if err != nil {
		common.Fail(c, common.CodeServerError, "计算安全评分失败: "+err.Error())
		return
	}
	common.Success(c, score)
}
