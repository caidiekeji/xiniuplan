package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"cloudpivot/internal/model"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const (
	totpIssuer       = "犀牛面板"
	totpTempTokenTTL = 5 * time.Minute
)

// GenerateTOTPSecret 生成 TOTP 密钥，返回 secret 和 otpauth:// URL（用于二维码）
func GenerateTOTPSecret(username string) (secret string, otpauthURL string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuer,
		AccountName: username,
	})
	if err != nil {
		return "", "", fmt.Errorf("generate totp key: %w", err)
	}
	return key.Secret(), key.URL(), nil
}

// VerifyTOTPCode 验证 6 位动态码（允许 ±1 时间窗口）
func VerifyTOTPCode(secret, code string) bool {
	opts := totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	}
	valid, err := totp.ValidateCustom(code, secret, time.Now(), opts)
	return err == nil && valid
}

// GenerateRecoveryCodes 生成 8 个备用恢复码（16 位十六进制字符串）
func GenerateRecoveryCodes() ([]string, error) {
	codes := make([]string, 8)
	for i := range codes {
		b := make([]byte, 8) // 8 bytes -> 16 hex chars
		if _, err := rand.Read(b); err != nil {
			return nil, fmt.Errorf("generate recovery code: %w", err)
		}
		codes[i] = hex.EncodeToString(b)
	}
	return codes, nil
}

// hashRecoveryCode 对恢复码做 SHA-256 哈希
func hashRecoveryCode(code string) string {
	h := sha256.Sum256([]byte(code))
	return hex.EncodeToString(h[:])
}

// SaveRecoveryCodes 保存恢复码哈希到数据库
func (s *Service) SaveRecoveryCodes(userID uint, codes []string) error {
	for _, code := range codes {
		rc := model.RecoveryCode{
			UserID:   userID,
			CodeHash: hashRecoveryCode(code),
		}
		if err := s.db.Create(&rc).Error; err != nil {
			return fmt.Errorf("save recovery code: %w", err)
		}
	}
	return nil
}

// VerifyRecoveryCode 验证恢复码，使用后标记为已用
func (s *Service) VerifyRecoveryCode(userID uint, code string) (bool, error) {
	hash := hashRecoveryCode(code)
	var rc model.RecoveryCode
	if err := s.db.Where("user_id = ? AND code_hash = ? AND used = ?", userID, hash, false).
		First(&rc).Error; err != nil {
		return false, nil
	}
	// 标记为已使用
	if err := s.db.Model(&rc).Update("used", true).Error; err != nil {
		return false, fmt.Errorf("mark recovery code used: %w", err)
	}
	return true, nil
}

// DeleteRecoveryCodes 删除用户所有恢复码
func (s *Service) DeleteRecoveryCodes(userID uint) error {
	return s.db.Where("user_id = ?", userID).Delete(&model.RecoveryCode{}).Error
}

// generateTempToken 生成 TOTP 二次验证用的临时 token
func (s *Service) generateTempToken(user model.User) (string, error) {
	return s.generateToken(user, totpTempTokenTTL, "totp_temp")
}

// LoginWithTOTP 使用 TOTP 码或恢复码完成二次验证，返回正式 JWT
func (s *Service) LoginWithTOTP(userID uint, code string) (*LoginResult, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	if user.Status != "active" {
		return nil, errors.New("账号已被禁用")
	}

	if user.TOTPSecret == "" {
		return nil, errors.New("用户未绑定 TOTP")
	}

	// 优先验证 TOTP 动态码
	if VerifyTOTPCode(user.TOTPSecret, code) {
		return s.generateTokens(user)
	}

	// 再验证恢复码
	ok, err := s.VerifyRecoveryCode(userID, code)
	if err != nil {
		return nil, err
	}
	if ok {
		return s.generateTokens(user)
	}

	return nil, errors.New("TOTP 验证码或恢复码错误")
}

// generateTokens 生成 access + refresh token
func (s *Service) generateTokens(user model.User) (*LoginResult, error) {
	accessToken, err := s.generateToken(user, s.accessTTL, "access")
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.generateToken(user, s.refreshTTL, "refresh")
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		ExpiresIn:             int(s.accessTTL.Seconds()),
		RequirePasswordChange: user.MustChangePassword,
	}, nil
}
