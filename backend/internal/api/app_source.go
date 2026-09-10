package api

import (
	"strconv"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	appstoresvc "cloudpivot/internal/service/appstore"

	"github.com/gin-gonic/gin"
)

// AppSourceHandler 应用源 API Handler
type AppSourceHandler struct {
	sourceSvc *appstoresvc.SourceService
}

// NewAppSourceHandler 创建应用源 Handler
func NewAppSourceHandler() *AppSourceHandler {
	return &AppSourceHandler{
		sourceSvc: appstoresvc.GetSourceService(model.DB),
	}
}

// RegisterAppSourceRoutes 注册应用源路由
func RegisterAppSourceRoutes(r *gin.RouterGroup) {
	h := NewAppSourceHandler()

	r.GET("/app-sources", middleware.RequirePermission(model.PermAppStoreView), h.ListSources)
	r.POST("/app-sources", middleware.RequirePermission(model.PermAppStoreInstall), h.CreateSource)
	r.PUT("/app-sources/:id", middleware.RequirePermission(model.PermAppStoreInstall), h.UpdateSource)
	r.DELETE("/app-sources/:id", middleware.RequirePermission(model.PermAppStoreInstall), h.DeleteSource)
	r.POST("/app-sources/:id/toggle", middleware.RequirePermission(model.PermAppStoreInstall), h.ToggleSource)
	r.POST("/app-sources/:id/sync", middleware.RequirePermission(model.PermAppStoreInstall), h.SyncSource)
}

// ListSources GET /app-sources
func (h *AppSourceHandler) ListSources(c *gin.Context) {
	sources, err := h.sourceSvc.ListSources()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取应用源列表失败: "+err.Error())
		return
	}

	responses := make([]dto.SourceResponse, 0, len(sources))
	for _, s := range sources {
		responses = append(responses, toSourceResponse(&s))
	}
	common.Success(c, gin.H{"items": responses, "total": len(responses)})
}

// CreateSource POST /app-sources
func (h *AppSourceHandler) CreateSource(c *gin.Context) {
	var req dto.CreateSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	source, err := h.sourceSvc.CreateSource(req)
	if err != nil {
		common.Fail(c, common.CodeServerError, "创建应用源失败: "+err.Error())
		return
	}
	common.Success(c, toSourceResponse(source))
}

// UpdateSource PUT /app-sources/:id
func (h *AppSourceHandler) UpdateSource(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.Fail(c, common.CodeParamError, "无效的源 ID")
		return
	}

	var req dto.UpdateSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	source, err := h.sourceSvc.UpdateSource(uint(id), req)
	if err != nil {
		common.Fail(c, common.CodeServerError, "更新应用源失败: "+err.Error())
		return
	}
	common.Success(c, toSourceResponse(source))
}

// DeleteSource DELETE /app-sources/:id
func (h *AppSourceHandler) DeleteSource(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.Fail(c, common.CodeParamError, "无效的源 ID")
		return
	}

	if err := h.sourceSvc.DeleteSource(uint(id)); err != nil {
		common.Fail(c, common.CodeServerError, "删除应用源失败: "+err.Error())
		return
	}
	common.Success(c, nil)
}

// ToggleSource POST /app-sources/:id/toggle
func (h *AppSourceHandler) ToggleSource(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.Fail(c, common.CodeParamError, "无效的源 ID")
		return
	}

	source, err := h.sourceSvc.ToggleSource(uint(id))
	if err != nil {
		common.Fail(c, common.CodeServerError, "切换源状态失败: "+err.Error())
		return
	}
	common.Success(c, toSourceResponse(source))
}

// SyncSource POST /app-sources/:id/sync
func (h *AppSourceHandler) SyncSource(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.Fail(c, common.CodeParamError, "无效的源 ID")
		return
	}

	result, err := h.sourceSvc.SyncSource(uint(id))
	if err != nil {
		common.Fail(c, common.CodeServerError, "同步应用源失败: "+err.Error())
		return
	}
	common.Success(c, result)
}

// toSourceResponse 转换为响应 DTO
func toSourceResponse(s *model.AppSource) dto.SourceResponse {
	return dto.SourceResponse{
		ID:         s.ID,
		Name:       s.Name,
		Type:       s.Type,
		URL:        s.URL,
		Path:       s.Path,
		Enabled:    s.Enabled,
		LastSyncAt: s.LastSyncAt,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}
