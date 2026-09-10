package api

import (
	"strconv"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RoleHandler 角色管理 API Handler
type RoleHandler struct {
	db *gorm.DB
}

// NewRoleHandler 创建角色管理 Handler
func NewRoleHandler(db *gorm.DB) *RoleHandler {
	return &RoleHandler{db: db}
}

// RegisterRoleRoutes 注册角色管理路由
func RegisterRoleRoutes(r *gin.RouterGroup) {
	h := NewRoleHandler(model.DB)
	g := r.Group("")
	{
		g.GET("/roles", middleware.RequirePermission(model.PermUserView), h.ListRoles)
		g.POST("/roles", middleware.RequirePermission(model.PermUserCreate), h.CreateRole)
		g.PUT("/roles/:id", middleware.RequirePermission(model.PermUserUpdate), h.UpdateRole)
		g.DELETE("/roles/:id", middleware.RequirePermission(model.PermUserDelete), h.DeleteRole)
		g.GET("/permissions", middleware.RequirePermission(model.PermUserView), h.ListPermissions)
	}
}

func toRoleResponse(role *model.Role) dto.RoleResponse {
	perms := []string{}
	if role.Permissions != nil {
		perms = role.Permissions
	}
	return dto.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		Permissions: perms,
		IsSystem:    role.IsSystem,
		CreatedAt:   role.CreatedAt.Format(time.RFC3339),
	}
}

// ListRoles GET /roles
func (h *RoleHandler) ListRoles(c *gin.Context) {
	var roles []model.Role
	if err := h.db.Order("id ASC").Find(&roles).Error; err != nil {
		common.Fail(c, common.CodeServerError, "查询角色失败: "+err.Error())
		return
	}
	resp := make([]dto.RoleResponse, 0, len(roles))
	for i := range roles {
		resp = append(resp, toRoleResponse(&roles[i]))
	}
	common.Success(c, resp)
}

// CreateRole POST /roles
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	var count int64
	h.db.Model(&model.Role{}).Where("code = ?", req.Code).Count(&count)
	if count > 0 {
		common.Fail(c, common.CodeParamError, "角色 code 已存在")
		return
	}
	perms := model.StringSlice(req.Permissions)
	if perms == nil {
		perms = model.StringSlice{}
	}
	role := model.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Permissions: perms,
		IsSystem:    false,
	}
	if err := h.db.Create(&role).Error; err != nil {
		common.Fail(c, common.CodeServerError, "创建角色失败: "+err.Error())
		return
	}
	common.Success(c, toRoleResponse(&role))
}

// UpdateRole PUT /roles/:id
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		common.Fail(c, common.CodeParamError, "角色 ID 无效")
		return
	}
	var role model.Role
	if err := h.db.First(&role, uint(id)).Error; err != nil {
		common.Fail(c, common.CodeNotFound, "角色不存在")
		return
	}
	if role.IsSystem {
		common.Fail(c, common.CodeForbidden, "内置角色不可修改")
		return
	}
	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Permissions != nil {
		updates["permissions"] = model.StringSlice(*req.Permissions)
	}
	if len(updates) > 0 {
		if err := h.db.Model(&role).Updates(updates).Error; err != nil {
			common.Fail(c, common.CodeServerError, "更新角色失败: "+err.Error())
			return
		}
	}
	h.db.First(&role, uint(id))
	common.Success(c, toRoleResponse(&role))
}

// DeleteRole DELETE /roles/:id
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		common.Fail(c, common.CodeParamError, "角色 ID 无效")
		return
	}
	var role model.Role
	if err := h.db.First(&role, uint(id)).Error; err != nil {
		common.Fail(c, common.CodeNotFound, "角色不存在")
		return
	}
	if role.IsSystem {
		common.Fail(c, common.CodeForbidden, "内置角色不可删除")
		return
	}
	var userCount int64
	h.db.Model(&model.User{}).Where("role_id = ?", uint(id)).Count(&userCount)
	if userCount > 0 {
		common.Fail(c, common.CodeForbidden, "该角色下仍有用户，无法删除")
		return
	}
	if err := h.db.Delete(&role).Error; err != nil {
		common.Fail(c, common.CodeServerError, "删除角色失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"deleted": true})
}

// ListPermissions GET /permissions
func (h *RoleHandler) ListPermissions(c *gin.Context) {
	common.Success(c, model.PermissionTree())
}
