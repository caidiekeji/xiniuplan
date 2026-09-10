package api

import (
	"strconv"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	usersvc "cloudpivot/internal/service/user"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户管理 API Handler
type UserHandler struct {
	svc *usersvc.Service
}

// NewUserHandler 创建用户管理 Handler
func NewUserHandler(svc *usersvc.Service) *UserHandler {
	return &UserHandler{svc: svc}
}

// RegisterUserRoutes 注册用户管理路由
func RegisterUserRoutes(r *gin.RouterGroup) {
	svc := usersvc.NewService(model.DB)
	h := NewUserHandler(svc)

	g := r.Group("/users")
	g.Use(middleware.RequirePermission(model.PermUserView))
	{
		g.GET("", h.ListUsers)
		g.GET("/:id", h.GetUser)
		g.POST("", middleware.RequirePermission(model.PermUserCreate), h.CreateUser)
		g.PUT("/:id", middleware.RequirePermission(model.PermUserUpdate), h.UpdateUser)
		g.DELETE("/:id", middleware.RequirePermission(model.PermUserDelete), h.DeleteUser)
		g.PUT("/:id/password", middleware.RequirePermission(model.PermUserUpdate), h.UpdatePassword)
		g.POST("/:id/disable", middleware.RequirePermission(model.PermUserUpdate), h.DisableUser)
		g.POST("/:id/enable", middleware.RequirePermission(model.PermUserUpdate), h.EnableUser)
	}
}

func parseUserID(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		common.Fail(c, common.CodeParamError, "用户 ID 无效")
		return 0, false
	}
	return uint(id), true
}

func toUserResponse(item *usersvc.UserListItem) dto.UserResponse {
	lastLogin := ""
	if item.LastLoginAt != nil {
		lastLogin = item.LastLoginAt.Format(time.RFC3339)
	}
	return dto.UserResponse{
		ID:          item.ID,
		Username:    item.Username,
		Role:        item.Role,
		RoleID:      item.RoleID,
		RoleName:    item.RoleName,
		Status:      item.Status,
		TOTPEnabled: item.TOTPEnabled,
		LastLoginAt: lastLogin,
		CreatedAt:   item.CreatedAt.Format(time.RFC3339),
	}
}

// ListUsers GET /users
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")

	items, total, err := h.svc.ListUsers(page, pageSize, keyword)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}

	resp := make([]dto.UserResponse, 0, len(items))
	for i := range items {
		resp = append(resp, toUserResponse(&items[i]))
	}
	common.Page(c, resp, total, page, pageSize)
}

// CreateUser POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	user, err := h.svc.CreateUser(req.Username, req.Password, req.RoleID, req.Status)
	if err != nil {
		common.Fail(c, common.CodeParamError, err.Error())
		return
	}

	common.Success(c, gin.H{"id": user.ID, "username": user.Username})
}

// GetUser GET /users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}

	item, err := h.svc.GetUserDetail(id)
	if err != nil {
		common.Fail(c, common.CodeNotFound, err.Error())
		return
	}
	common.Success(c, toUserResponse(item))
}

// UpdateUser PUT /users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	user, err := h.svc.UpdateUser(id, req.RoleID, req.Status)
	if err != nil {
		common.Fail(c, common.CodeParamError, err.Error())
		return
	}
	common.Success(c, gin.H{"id": user.ID, "updated": true})
}

// DeleteUser DELETE /users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}

	currentUserID := middleware.GetUserID(c)
	if err := h.svc.DeleteUser(id, currentUserID); err != nil {
		common.Fail(c, common.CodeForbidden, err.Error())
		return
	}
	common.Success(c, gin.H{"deleted": true})
}

// UpdatePassword PUT /users/:id/password
func (h *UserHandler) UpdatePassword(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}

	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	currentUserID := middleware.GetUserID(c)
	isSelf := id == currentUserID

	if err := h.svc.UpdatePassword(id, req.OldPassword, req.NewPassword, isSelf); err != nil {
		common.Fail(c, common.CodeParamError, err.Error())
		return
	}
	common.Success(c, gin.H{"updated": true})
}

// DisableUser POST /users/:id/disable
func (h *UserHandler) DisableUser(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}
	if err := h.svc.DisableUser(id); err != nil {
		common.Fail(c, common.CodeForbidden, err.Error())
		return
	}
	common.Success(c, gin.H{"disabled": true})
}

// EnableUser POST /users/:id/enable
func (h *UserHandler) EnableUser(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}
	if err := h.svc.EnableUser(id); err != nil {
		common.Fail(c, common.CodeParamError, err.Error())
		return
	}
	common.Success(c, gin.H{"enabled": true})
}
