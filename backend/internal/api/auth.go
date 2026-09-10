package api

import (
	"strings"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	authsvc "cloudpivot/internal/service/auth"
	usersvc "cloudpivot/internal/service/user"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证 API Handler
type AuthHandler struct {
	svc *authsvc.Service
}

// globalAuthService 包级 auth service，供 TOTP 等路由注册使用（main.go 不修改的前提下）
var globalAuthService *authsvc.Service

// NewAuthHandler 创建认证 Handler
func NewAuthHandler(svc *authsvc.Service) *AuthHandler {
	globalAuthService = svc
	return &AuthHandler{svc: svc}
}

// Login POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		common.Fail(c, common.CodeUnauthorized, err.Error())
		return
	}

	// 需要 TOTP 二次验证
	if result.RequireTOTP {
		common.Success(c, gin.H{
			"require_totp": true,
			"temp_token":   result.TempToken,
		})
		return
	}

	common.Success(c, dto.LoginResponse{
		AccessToken:           result.AccessToken,
		RefreshToken:          result.RefreshToken,
		ExpiresIn:             result.ExpiresIn,
		TokenType:             "Bearer",
		RequirePasswordChange: result.RequirePasswordChange,
	})
}

// Logout POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			// 将 access token 加入黑名单，2小时后自动过期
			middleware.BlacklistToken(parts[1], time.Now().Add(2*time.Hour))
		}
	}
	common.Success(c, nil)
}

// Refresh POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.Refresh(req.RefreshToken)
	if err != nil {
		common.Fail(c, common.CodeUnauthorized, err.Error())
		return
	}

	common.Success(c, dto.LoginResponse{
		AccessToken:           result.AccessToken,
		RefreshToken:          result.RefreshToken,
		ExpiresIn:             result.ExpiresIn,
		TokenType:             "Bearer",
		RequirePasswordChange: result.RequirePasswordChange,
	})
}

// ChangePassword PUT /api/v1/auth/change-password
// 自助修改密码：任何已登录用户可用（不受 PermUserUpdate 限制），
// 供首次登录强制改密和普通用户自助改密使用。
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		common.Fail(c, common.CodeUnauthorized, "未认证")
		return
	}

	// isSelf=true：必须校验原密码
	svc := usersvc.NewService(model.DB)
	if err := svc.UpdatePassword(userID, req.OldPassword, req.NewPassword, true); err != nil {
		common.Fail(c, common.CodeParamError, err.Error())
		return
	}

	// 改密成功后将当前 token 加入黑名单，强制重新登录获取无标志的新 token
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		if parts := strings.SplitN(authHeader, " ", 2); len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			middleware.BlacklistToken(parts[1], time.Now().Add(2*time.Hour))
		}
	}

	common.Success(c, gin.H{"updated": true})
}

// Info GET /api/v1/auth/info
func (h *AuthHandler) Info(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.svc.GetUserInfo(userID)
	if err != nil {
		common.Fail(c, common.CodeNotFound, err.Error())
		return
	}

	lastLogin := ""
	if user.LastLoginAt != nil {
		lastLogin = user.LastLoginAt.Format(time.RFC3339)
	}

	common.Success(c, dto.UserInfoResponse{
		ID:                 user.ID,
		Username:           user.Username,
		Role:               user.Role,
		Status:             user.Status,
		TOTPEnabled:        user.TOTPSecret != "",
		MustChangePassword: user.MustChangePassword,
		LastLoginAt:        lastLogin,
		CreatedAt:          user.CreatedAt.Format(time.RFC3339),
	})
}

// ForgotPassword POST /api/v1/auth/forgot-password
// 忘记密码：用户名+邮箱匹配则发送重置邮件。无论匹配与否统一返回成功（防账号枚举）。
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	// 构造面板访问来源（含安全入口前缀，链接可直接打开）
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	origin := scheme + "://" + c.Request.Host
	if entry := middleware.GetSecurityEntry(); entry != "" {
		origin += "/" + entry
	}

	if err := h.svc.RequestPasswordReset(req.Username, req.Email, origin); err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, gin.H{"sent": true})
}

// ResetPassword POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.ResetPassword(req.Token, req.NewPassword); err != nil {
		common.Fail(c, common.CodeParamError, err.Error())
		return
	}
	common.Success(c, gin.H{"reset": true})
}
