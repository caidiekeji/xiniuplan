package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"cloudpivot/internal/config"
	"cloudpivot/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Service 认证服务
type Service struct {
	db         *gorm.DB
	privKey    *rsa.PrivateKey
	pubKey     *rsa.PublicKey
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewService 创建认证服务
func NewService(db *gorm.DB, privKey *rsa.PrivateKey, pubKey *rsa.PublicKey, cfg *config.Config) *Service {
	return &Service{
		db:         db,
		privKey:    privKey,
		pubKey:     pubKey,
		accessTTL:  time.Duration(cfg.JWT.AccessTokenTTL) * time.Second,
		refreshTTL: time.Duration(cfg.JWT.RefreshTokenTTL) * time.Second,
	}
}

// InitAdmin 初始化管理员账号
func (s *Service) InitAdmin() error {
	var user model.User
	err := s.db.Where("username = ?", "admin").First(&user).Error
	if err == nil {
		// 已存在 admin：确保其为超级管理员角色
		if user.Role != "super_admin" {
			s.db.Model(&user).Update("role", "super_admin")
		}
		// 同步 RoleID
		var superRole model.Role
		if rErr := s.db.Where("code = ?", "super_admin").First(&superRole).Error; rErr == nil {
			if user.RoleID != superRole.ID {
				s.db.Model(&user).Update("role_id", superRole.ID)
			}
		}
		return nil
	}

	randomPassword := generateRandomPassword(16)
	hash, err := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// 查询超级管理员角色 ID
	var roleID uint
	var superRole model.Role
	if rErr := s.db.Where("code = ?", "super_admin").First(&superRole).Error; rErr == nil {
		roleID = superRole.ID
	}

	admin := model.User{
		Username:           "admin",
		PasswordHash:       string(hash),
		Role:               "super_admin",
		RoleID:             roleID,
		Status:             "active",
		MustChangePassword: true,
	}
	if err := s.db.Create(&admin).Error; err != nil {
		return fmt.Errorf("create admin: %w", err)
	}

	log.Printf("[犀牛面板] 初始管理员已创建: admin / %s，请立即修改密码！首次登录后将强制修改", randomPassword)
	return nil
}

// randomPasswordCharset 随机密码字符集（避免易混淆字符 0/O/1/l/I）
const randomPasswordCharset = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#$%^&*"

// generateRandomPassword 生成指定长度的随机密码
func generateRandomPassword(length int) string {
	if length <= 0 {
		length = 16
	}
	result := make([]byte, length)
	max := big.NewInt(int64(len(randomPasswordCharset)))
	for i := range result {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			// 极端情况下回退到固定字符
			result[i] = randomPasswordCharset[0]
			continue
		}
		result[i] = randomPasswordCharset[n.Int64()]
	}
	return string(result)
}

// LoginResult 登录结果
type LoginResult struct {
	AccessToken           string
	RefreshToken          string
	ExpiresIn             int
	RequireTOTP           bool   // 是否需要二次验证
	TempToken             string // TOTP 临时 token（RequireTOTP=true 时有效）
	RequirePasswordChange bool   // 是否需要强制修改密码
}

// Login 用户登录
func (s *Service) Login(username, password string) (*LoginResult, error) {
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	if user.Status != "active" {
		return nil, errors.New("账号已被禁用")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 更新最后登录时间
	now := time.Now()
	s.db.Model(&user).Update("last_login_at", &now)

	// 如果已绑定 TOTP，返回临时 token 要求二次验证
	if user.TOTPSecret != "" {
		tempToken, err := s.generateTempToken(user)
		if err != nil {
			return nil, fmt.Errorf("generate temp token: %w", err)
		}
		return &LoginResult{
			RequireTOTP: true,
			TempToken:   tempToken,
		}, nil
	}

	// 生成正式 token
	return s.generateTokens(user)
}

// Refresh 刷新 token
func (s *Service) Refresh(refreshToken string) (*LoginResult, error) {
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, errors.New("刷新令牌无效")
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return nil, errors.New("令牌类型错误")
	}

	userIDFloat, _ := claims["user_id"].(float64)
	var user model.User
	if err := s.db.First(&user, uint(userIDFloat)).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	if user.Status != "active" {
		return nil, errors.New("账号已被禁用")
	}

	return s.generateTokens(user)
}

// GetUserInfo 获取用户信息
func (s *Service) GetUserInfo(userID uint) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}

// generateToken 生成 JWT token
func (s *Service) generateToken(user model.User, ttl time.Duration, tokenType string) (string, error) {
	now := time.Now()

	// 查询用户角色的权限列表
	permissions := s.getUserPermissions(user)

	claims := jwt.MapClaims{
		"user_id":              user.ID,
		"username":             user.Username,
		"role":                 user.Role,
		"permissions":          permissions,
		"must_change_password": user.MustChangePassword,
		"type":                 tokenType,
		"iat":                  now.Unix(),
		"exp":                  now.Add(ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privKey)
}

// getUserPermissions 查询用户角色对应的权限列表
func (s *Service) getUserPermissions(user model.User) []string {
	var role model.Role
	var err error
	if user.RoleID > 0 {
		err = s.db.First(&role, user.RoleID).Error
	} else {
		err = s.db.Where("code = ?", user.Role).First(&role).Error
	}
	if err != nil || len(role.Permissions) == 0 {
		return []string{}
	}
	return role.Permissions
}

// parseToken 解析并验证 token
func (s *Service) parseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.pubKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// GetPublicKey 获取公钥（供中间件使用）
func (s *Service) GetPublicKey() *rsa.PublicKey {
	return s.pubKey
}
