package api

import (
	"strconv"

	"cloudpivot/internal/common"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	apikeysvc "cloudpivot/internal/service/apikey"

	"github.com/gin-gonic/gin"
)

// APIKeyHandler API Key 管理 Handler（U-007）
type APIKeyHandler struct {
	svc *apikeysvc.Service
}

func NewAPIKeyHandler() *APIKeyHandler {
	return &APIKeyHandler{svc: apikeysvc.NewService(model.DB)}
}

// RegisterAPIKeyRoutes 注册 API Key 路由（认证组下）
func RegisterAPIKeyRoutes(r *gin.RouterGroup) {
	h := NewAPIKeyHandler()

	r.GET("/auth/apikeys", h.List)
	r.POST("/auth/apikeys", h.Create)
	r.PUT("/auth/apikeys/:id", h.Update)
	r.DELETE("/auth/apikeys/:id", h.Delete)

	// 在线会话管理（U-006）
	r.GET("/auth/sessions", h.ListSessions)
	r.DELETE("/auth/sessions/:hash", h.RevokeSession)
}

// List GET /auth/apikeys
func (h *APIKeyHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	// 管理员可看全部，普通用户只看自己的
	if middleware.GetRole(c) != "super_admin" {
		_ = userID
	} else {
		userID = 0
	}
	items, err := h.svc.List(userID)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取 API Key 列表失败: "+err.Error())
		return
	}
	common.Success(c, items)
}

// Create POST /auth/apikeys
func (h *APIKeyHandler) Create(c *gin.Context) {
	var req struct {
		Name          string `json:"name" binding:"required"`
		Permissions   string `json:"permissions"` // 逗号分隔权限点，空=继承用户角色
		ExpiresInDays int    `json:"expires_in_days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	apiKey, err := h.svc.Create(req.Name, middleware.GetUserID(c), req.Permissions, req.ExpiresInDays)
	if err != nil {
		common.Fail(c, common.CodeServerError, "创建 API Key 失败: "+err.Error())
		return
	}
	common.Success(c, apiKey) // secret 仅此一次返回
}

// Update PUT /auth/apikeys/:id
func (h *APIKeyHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.Fail(c, common.CodeParamError, "无效的 ID")
		return
	}
	var req struct {
		Status      string `json:"status"`
		Permissions string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	callerID := middleware.GetUserID(c)
	if middleware.GetRole(c) == "super_admin" {
		callerID = 0 // 管理员可操作所有 key
	}
	apiKey, err := h.svc.Update(uint(id), callerID, req.Status, req.Permissions)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, apiKey)
}

// Delete DELETE /auth/apikeys/:id
func (h *APIKeyHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.Fail(c, common.CodeParamError, "无效的 ID")
		return
	}
	callerID := middleware.GetUserID(c)
	if middleware.GetRole(c) == "super_admin" {
		callerID = 0
	}
	if err := h.svc.Delete(uint(id), callerID); err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, gin.H{"deleted": true})
}

// ListSessions GET /auth/sessions - 在线会话列表
func (h *APIKeyHandler) ListSessions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if middleware.GetRole(c) == "super_admin" {
		userID = 0 // 管理员看全部
	}
	common.Success(c, middleware.ListSessions(userID))
}

// RevokeSession DELETE /auth/sessions/:hash - 踢出会话
func (h *APIKeyHandler) RevokeSession(c *gin.Context) {
	hash := c.Param("hash")
	callerID := middleware.GetUserID(c)
	if middleware.GetRole(c) == "super_admin" {
		callerID = 0 // 管理员可踢任何人
	}
	if err := middleware.RevokeSession(hash, callerID); err != nil {
		common.Fail(c, common.CodeForbidden, err.Error())
		return
	}
	common.Success(c, gin.H{"revoked": true})
}
