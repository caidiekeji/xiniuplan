package api

import (
	"strconv"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	sshsvc "cloudpivot/internal/service/ssh"

	"github.com/gin-gonic/gin"
)

// SSHHandler SSH 安全加固 API Handler
type SSHHandler struct {
	svc *sshsvc.Service
}

// NewSSHHandler 创建 SSH Handler
func NewSSHHandler(svc *sshsvc.Service) *SSHHandler {
	return &SSHHandler{svc: svc}
}

// RegisterSSHRoutes 注册 SSH 安全加固路由
func RegisterSSHRoutes(r *gin.RouterGroup) {
	svc := sshsvc.NewService(model.DB)
	h := NewSSHHandler(svc)

	r.GET("/ssh/config", middleware.RequirePermission(model.PermSSHView), h.GetConfig)
	r.PUT("/ssh/config", middleware.RequirePermission(model.PermSSHManage), h.UpdateConfig)
	r.GET("/ssh/keys", middleware.RequirePermission(model.PermSSHView), h.ListKeys)
	r.POST("/ssh/keys", middleware.RequirePermission(model.PermSSHManage), h.AddKey)
	r.DELETE("/ssh/keys/:fingerprint", middleware.RequirePermission(model.PermSSHManage), h.DeleteKey)
	r.GET("/ssh/logs", middleware.RequirePermission(model.PermSSHView), h.GetLoginLogs)
	r.GET("/ssh/fail2ban/status", middleware.RequirePermission(model.PermSSHView), h.GetFail2banStatus)
	r.PUT("/ssh/fail2ban/config", middleware.RequirePermission(model.PermSSHManage), h.UpdateFail2banConfig)
	r.POST("/ssh/fail2ban/unban", middleware.RequirePermission(model.PermSSHManage), h.UnbanIP)
	r.POST("/ssh/fail2ban/ban", middleware.RequirePermission(model.PermSSHManage), h.BanIP)
}

// GetConfig GET /ssh/config — 读取当前 SSH 配置
func (h *SSHHandler) GetConfig(c *gin.Context) {
	cfg, err := h.svc.GetConfig()
	if err != nil {
		common.Fail(c, common.CodeServerError, "读取 SSH 配置失败: "+err.Error())
		return
	}
	common.Success(c, cfg)
}

// UpdateConfig PUT /ssh/config — 修改 SSH 配置
func (h *SSHHandler) UpdateConfig(c *gin.Context) {
	var req dto.UpdateSSHConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	// 基本参数校验
	if req.Port < 1 || req.Port > 65535 {
		common.Fail(c, common.CodeParamError, "端口号必须在 1-65535 之间")
		return
	}
	if req.MaxAuthTries < 1 || req.MaxAuthTries > 20 {
		common.Fail(c, common.CodeParamError, "最大认证尝试次数必须在 1-20 之间")
		return
	}

	if err := h.svc.UpdateConfig(req); err != nil {
		common.Fail(c, common.CodeServerError, "修改 SSH 配置失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"updated": true})
}

// ListKeys GET /ssh/keys — 列出 authorized_keys
func (h *SSHHandler) ListKeys(c *gin.Context) {
	keys, err := h.svc.ListKeys()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取 SSH 密钥列表失败: "+err.Error())
		return
	}
	common.Success(c, keys)
}

// AddKey POST /ssh/keys — 添加公钥
func (h *SSHHandler) AddKey(c *gin.Context) {
	var req dto.AddSSHKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	key, err := h.svc.AddKey(req.PublicKey)
	if err != nil {
		common.Fail(c, common.CodeServerError, "添加 SSH 公钥失败: "+err.Error())
		return
	}
	common.Success(c, key)
}

// DeleteKey DELETE /ssh/keys/:fingerprint — 删除公钥
func (h *SSHHandler) DeleteKey(c *gin.Context) {
	fingerprint := c.Param("fingerprint")
	if fingerprint == "" {
		common.Fail(c, common.CodeParamError, "fingerprint 不能为空")
		return
	}

	if err := h.svc.DeleteKey(fingerprint); err != nil {
		common.Fail(c, common.CodeServerError, "删除 SSH 公钥失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"deleted": true})
}

// GetLoginLogs GET /ssh/logs — 登录日志（分页）
func (h *SSHHandler) GetLoginLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	logs, total, err := h.svc.GetLoginLogs(page, pageSize)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取 SSH 登录日志失败: "+err.Error())
		return
	}
	common.Page(c, logs, total, page, pageSize)
}

// GetFail2banStatus GET /ssh/fail2ban/status — Fail2ban 状态+封禁列表
func (h *SSHHandler) GetFail2banStatus(c *gin.Context) {
	status, err := h.svc.GetFail2banStatus()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取 Fail2ban 状态失败: "+err.Error())
		return
	}
	common.Success(c, status)
}

// UpdateFail2banConfig PUT /ssh/fail2ban/config — 修改封禁策略
func (h *SSHHandler) UpdateFail2banConfig(c *gin.Context) {
	var req dto.UpdateFail2banConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	if req.MaxRetry < 1 || req.MaxRetry > 20 {
		common.Fail(c, common.CodeParamError, "最大重试次数必须在 1-20 之间")
		return
	}
	if req.BanTime < 60 || req.BanTime > 86400*365 {
		common.Fail(c, common.CodeParamError, "封禁时长必须在 60 秒 - 1 年之间")
		return
	}
	if req.FindTime < 60 || req.FindTime > 86400 {
		common.Fail(c, common.CodeParamError, "检测窗口必须在 60 秒 - 1 天之间")
		return
	}

	if err := h.svc.UpdateFail2banConfig(req); err != nil {
		common.Fail(c, common.CodeServerError, "修改 Fail2ban 配置失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"updated": true})
}

// UnbanIP POST /ssh/fail2ban/unban — 解封 IP
func (h *SSHHandler) UnbanIP(c *gin.Context) {
	var req dto.IPActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.UnbanIP(req.IP); err != nil {
		common.Fail(c, common.CodeServerError, "解封 IP 失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"unbanned": req.IP})
}

// BanIP POST /ssh/fail2ban/ban — 手动封禁
func (h *SSHHandler) BanIP(c *gin.Context) {
	var req dto.IPActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.BanIP(req.IP); err != nil {
		common.Fail(c, common.CodeServerError, "封禁 IP 失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"banned": req.IP})
}
