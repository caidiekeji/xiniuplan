package dto

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6"`
	RoleID   uint   `json:"role_id" binding:"required"`
	Status   string `json:"status"` // active | disabled，默认 active
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	RoleID *uint   `json:"role_id"`
	Status *string `json:"status"`
}

// UpdatePasswordRequest 修改密码请求
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UserResponse 用户信息响应
type UserResponse struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Role        string `json:"role"`
	RoleID      uint   `json:"role_id"`
	RoleName    string `json:"role_name"`
	Status      string `json:"status"`
	TOTPEnabled bool   `json:"totp_enabled"`
	LastLoginAt string `json:"last_login_at"`
	CreatedAt   string `json:"created_at"`
}
