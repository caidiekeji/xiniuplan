package dto

// TOTPBindResponse 绑定 TOTP 响应
type TOTPBindResponse struct {
	Secret     string `json:"secret"`
	OTPAuthURL string `json:"otpauth_url"`
}

// TOTPVerifyRequest 验证 TOTP 绑定请求
type TOTPVerifyRequest struct {
	Secret string `json:"secret" binding:"required"`
	Code   string `json:"code" binding:"required"`
}

// TOTPVerifyResponse TOTP 绑定成功响应（含恢复码）
type TOTPVerifyResponse struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

// TOTPDisableRequest 解绑 TOTP 请求
type TOTPDisableRequest struct {
	Code string `json:"code" binding:"required"`
}

// TOTPStatusResponse TOTP 状态响应
type TOTPStatusResponse struct {
	Enabled bool `json:"enabled"`
}

// TOTPLoginRequest TOTP 二次登录请求
type TOTPLoginRequest struct {
	TOTPCode string `json:"totp_code" binding:"required"`
}
