package api

import (
	"strings"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	authsvc "cloudpivot/internal/service/auth"

	"github.com/gin-gonic/gin"
)

// TOTPHandler TOTP API Handler
type TOTPHandler struct {
	svc *authsvc.Service
}

// NewTOTPHandler 创建 TOTP Handler
func NewTOTPHandler(svc *authsvc.Service) *TOTPHandler {
	return &TOTPHandler{svc: svc}
}

// RegisterTOTPRoutes 注册 TOTP 路由
// 注意：/auth/totp/login 接收 temp_token（type=totp_temp 的 JWT），通过 JWTAuth 中间件后由 handler 校验类型
func RegisterTOTPRoutes(r *gin.RouterGroup) {
	h := NewTOTPHandler(globalAuthService)

	r.POST("/auth/totp/bind", h.Bind)
	r.POST("/auth/totp/verify", h.Verify)
	r.POST("/auth/totp/disable", h.Disable)
	r.GET("/auth/totp/status", h.Status)
	r.POST("/auth/totp/login", h.Login)
}

// Bind POST /auth/totp/bind — 生成密钥+二维码URL
func (h *TOTPHandler) Bind(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.svc.GetUserInfo(userID)
	if err != nil {
		common.Fail(c, common.CodeNotFound, err.Error())
		return
	}

	if user.TOTPSecret != "" {
		common.Fail(c, common.CodeParamError, "已绑定 TOTP，请先解绑")
		return
	}

	secret, otpauthURL, err := authsvc.GenerateTOTPSecret(user.Username)
	if err != nil {
		common.Fail(c, common.CodeServerError, "生成 TOTP 密钥失败: "+err.Error())
		return
	}

	common.Success(c, dto.TOTPBindResponse{
		Secret:     secret,
		OTPAuthURL: otpauthURL,
	})
}

// Verify POST /auth/totp/verify — 验证6位码，成功则绑定并返回恢复码
func (h *TOTPHandler) Verify(c *gin.Context) {
	var req dto.TOTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	if !authsvc.VerifyTOTPCode(req.Secret, req.Code) {
		common.Fail(c, common.CodeParamError, "TOTP 验证码错误")
		return
	}

	userID := middleware.GetUserID(c)
	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		common.Fail(c, common.CodeNotFound, "用户不存在")
		return
	}

	if user.TOTPSecret != "" {
		common.Fail(c, common.CodeParamError, "已绑定 TOTP")
		return
	}

	// 保存密钥
	if err := model.DB.Model(&user).Update("totp_secret", req.Secret).Error; err != nil {
		common.Fail(c, common.CodeServerError, "保存 TOTP 密钥失败")
		return
	}

	// 生成恢复码
	codes, err := authsvc.GenerateRecoveryCodes()
	if err != nil {
		common.Fail(c, common.CodeServerError, "生成恢复码失败: "+err.Error())
		return
	}
	if err := h.svc.SaveRecoveryCodes(userID, codes); err != nil {
		common.Fail(c, common.CodeServerError, "保存恢复码失败: "+err.Error())
		return
	}

	common.Success(c, dto.TOTPVerifyResponse{
		RecoveryCodes: codes,
	})
}

// Disable POST /auth/totp/disable — 验证当前 TOTP 码后解绑
func (h *TOTPHandler) Disable(c *gin.Context) {
	var req dto.TOTPDisableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	user, err := h.svc.GetUserInfo(userID)
	if err != nil {
		common.Fail(c, common.CodeNotFound, err.Error())
		return
	}

	if user.TOTPSecret == "" {
		common.Fail(c, common.CodeParamError, "未绑定 TOTP")
		return
	}

	// 验证 TOTP 码
	if !authsvc.VerifyTOTPCode(user.TOTPSecret, req.Code) {
		common.Fail(c, common.CodeParamError, "TOTP 验证码错误")
		return
	}

	// 清除密钥
	if err := model.DB.Model(&model.User{}).Where("id = ?", userID).Update("totp_secret", "").Error; err != nil {
		common.Fail(c, common.CodeServerError, "解绑 TOTP 失败")
		return
	}

	// 删除恢复码
	h.svc.DeleteRecoveryCodes(userID)

	common.Success(c, gin.H{"disabled": true})
}

// Status GET /auth/totp/status — 当前用户 TOTP 绑定状态
func (h *TOTPHandler) Status(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.svc.GetUserInfo(userID)
	if err != nil {
		common.Fail(c, common.CodeNotFound, err.Error())
		return
	}
	common.Success(c, dto.TOTPStatusResponse{
		Enabled: user.TOTPSecret != "",
	})
}

// Login POST /auth/totp/login — temp_token + totp_code 完成二次验证，返回正式 JWT
func (h *TOTPHandler) Login(c *gin.Context) {
	// 校验 token 类型必须为 totp_temp
	tokenType := middleware.GetTokenType(c)
	if tokenType != "totp_temp" {
		common.Fail(c, common.CodeUnauthorized, "临时令牌无效")
		return
	}

	var req dto.TOTPLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	result, err := h.svc.LoginWithTOTP(userID, req.TOTPCode)
	if err != nil {
		common.Fail(c, common.CodeUnauthorized, err.Error())
		return
	}

	// 将 temp_token 加入黑名单
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			middleware.BlacklistToken(parts[1], time.Now().Add(10*time.Minute))
		}
	}

	common.Success(c, dto.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    "Bearer",
	})
}
