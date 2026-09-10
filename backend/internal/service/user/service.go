package user

import (
	"errors"
	"fmt"
	"time"

	"cloudpivot/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Service 用户管理服务
type Service struct {
	db *gorm.DB
}

// NewService 创建用户管理服务
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// UserListItem 用户列表项（含角色名）
type UserListItem struct {
	ID          uint       `json:"id"`
	Username    string     `json:"username"`
	Role        string     `json:"role"`
	RoleID      uint       `json:"role_id"`
	RoleName    string     `json:"role_name"`
	Status      string     `json:"status"`
	TOTPEnabled bool       `json:"totp_enabled"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ListUsers 分页查询用户
func (s *Service) ListUsers(page, pageSize int, keyword string) ([]UserListItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := s.db.Model(&model.User{})
	if keyword != "" {
		query = query.Where("username LIKE ?", "%"+keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	users := make([]model.User, 0)
	if err := query.Order("id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}

	// 批量查询角色名
	roleIDs := make([]uint, 0, len(users))
	for _, u := range users {
		if u.RoleID > 0 {
			roleIDs = append(roleIDs, u.RoleID)
		}
	}
	roleMap := make(map[uint]string)
	if len(roleIDs) > 0 {
		var roles []model.Role
		s.db.Where("id IN ?", roleIDs).Find(&roles)
		for _, r := range roles {
			roleMap[r.ID] = r.Name
		}
	}

	items := make([]UserListItem, 0, len(users))
	for _, u := range users {
		items = append(items, UserListItem{
			ID:          u.ID,
			Username:    u.Username,
			Role:        u.Role,
			RoleID:      u.RoleID,
			RoleName:    roleMap[u.RoleID],
			Status:      u.Status,
			TOTPEnabled: u.TOTPSecret != "",
			LastLoginAt: u.LastLoginAt,
			CreatedAt:   u.CreatedAt,
		})
	}
	return items, total, nil
}

// CreateUser 创建用户
func (s *Service) CreateUser(username, password string, roleID uint, status string) (*model.User, error) {
	// 检查用户名是否已存在
	var count int64
	s.db.Model(&model.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	// 验证角色存在
	var role model.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		return nil, errors.New("角色不存在")
	}

	if status == "" {
		status = "active"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         role.Code,
		RoleID:       role.ID,
		Status:       status,
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &user, nil
}

// GetUserDetail 获取用户详情
func (s *Service) GetUserDetail(id uint) (*UserListItem, error) {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	var roleName string
	if user.RoleID > 0 {
		var role model.Role
		if err := s.db.First(&role, user.RoleID).Error; err == nil {
			roleName = role.Name
		}
	}

	return &UserListItem{
		ID:          user.ID,
		Username:    user.Username,
		Role:        user.Role,
		RoleID:      user.RoleID,
		RoleName:    roleName,
		Status:      user.Status,
		TOTPEnabled: user.TOTPSecret != "",
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
	}, nil
}

// UpdateUser 更新用户角色/状态
func (s *Service) UpdateUser(id uint, roleID *uint, status *string) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	updates := map[string]interface{}{}
	if roleID != nil {
		var role model.Role
		if err := s.db.First(&role, *roleID).Error; err != nil {
			return nil, errors.New("角色不存在")
		}
		updates["role_id"] = role.ID
		updates["role"] = role.Code
	}
	if status != nil {
		if *status != "active" && *status != "disabled" {
			return nil, errors.New("状态值无效")
		}
		updates["status"] = *status
	}

	if len(updates) > 0 {
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update user: %w", err)
		}
	}

	s.db.First(&user, id)
	return &user, nil
}

// DeleteUser 删除用户（不能删自己，不能删 super_admin）
func (s *Service) DeleteUser(id, currentUserID uint) error {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}
	if user.ID == currentUserID {
		return errors.New("不能删除当前登录用户")
	}
	if user.Role == "super_admin" {
		return errors.New("不能删除超级管理员")
	}

	// 删除用户的恢复码
	s.db.Where("user_id = ?", id).Delete(&model.RecoveryCode{})

	return s.db.Delete(&user).Error
}

// UpdatePassword 修改用户密码
// 如果 oldPassword 为空，表示管理员重置他人密码；否则需验证旧密码
func (s *Service) UpdatePassword(id uint, oldPassword, newPassword string, isSelf bool) error {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}

	if isSelf {
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
			return errors.New("原密码错误")
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	// 修改密码后清除"强制修改密码"标志，避免首次登录死循环
	return s.db.Model(&user).Updates(map[string]interface{}{
		"password_hash":        string(hash),
		"must_change_password": false,
	}).Error
}

// DisableUser 禁用用户
func (s *Service) DisableUser(id uint) error {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}
	if user.Role == "super_admin" {
		return errors.New("不能禁用超级管理员")
	}
	return s.db.Model(&user).Update("status", "disabled").Error
}

// EnableUser 启用用户
func (s *Service) EnableUser(id uint) error {
	var user model.User
	if err := s.db.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}
	return s.db.Model(&user).Update("status", "active").Error
}
