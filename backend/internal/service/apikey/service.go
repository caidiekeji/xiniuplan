package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// Service API Key 管理服务（U-007）
type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// List 列出 API Key（指定用户的；userID=0 返回全部，仅管理员）
func (s *Service) List(userID uint) ([]model.APIKey, error) {
	items := make([]model.APIKey, 0)
	query := s.db.Order("id DESC")
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("query api keys: %w", err)
	}
	return items, nil
}

// Create 创建 API Key，secret 明文仅在返回值中出现一次
func (s *Service) Create(name string, userID uint, permissions string, expiresInDays int) (*model.APIKey, error) {
	if name == "" {
		return nil, fmt.Errorf("名称不能为空")
	}

	// keyID：8 字节随机 base64url（去填充）；secret：32 字节随机
	kb := make([]byte, 8)
	if _, err := rand.Read(kb); err != nil {
		return nil, fmt.Errorf("generate key id: %w", err)
	}
	keyID := base64.RawURLEncoding.EncodeToString(kb)

	sb := make([]byte, 32)
	if _, err := rand.Read(sb); err != nil {
		return nil, fmt.Errorf("generate secret: %w", err)
	}
	secret := base64.RawURLEncoding.EncodeToString(sb)

	sum := sha256.Sum256([]byte(secret))
	apiKey := &model.APIKey{
		Name:        name,
		KeyID:       keyID,
		KeyHash:     hex.EncodeToString(sum[:]),
		UserID:      userID,
		Permissions: permissions,
		Status:      "active",
	}
	if expiresInDays > 0 {
		exp := time.Now().Add(time.Duration(expiresInDays) * 24 * time.Hour)
		apiKey.ExpiresAt = &exp
	}
	if err := s.db.Create(apiKey).Error; err != nil {
		return nil, fmt.Errorf("create api key: %w", err)
	}
	apiKey.Secret = keyID + "." + secret // 唯一一次明文返回
	return apiKey, nil
}

// Update 更新状态/权限
func (s *Service) Update(id uint, callerID uint, status, permissions string) (*model.APIKey, error) {
	var apiKey model.APIKey
	if err := s.db.First(&apiKey, id).Error; err != nil {
		return nil, fmt.Errorf("API Key 不存在")
	}
	if callerID > 0 && apiKey.UserID != callerID {
		return nil, fmt.Errorf("无权操作他人的 API Key")
	}
	if status != "" {
		if status != "active" && status != "disabled" {
			return nil, fmt.Errorf("无效状态: %s", status)
		}
		apiKey.Status = status
	}
	if permissions != "" {
		apiKey.Permissions = permissions
	}
	if err := s.db.Save(&apiKey).Error; err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// Delete 删除 API Key
func (s *Service) Delete(id uint, callerID uint) error {
	var apiKey model.APIKey
	if err := s.db.First(&apiKey, id).Error; err != nil {
		return fmt.Errorf("API Key 不存在")
	}
	if callerID > 0 && apiKey.UserID != callerID {
		return fmt.Errorf("无权操作他人的 API Key")
	}
	return s.db.Delete(&apiKey).Error
}
