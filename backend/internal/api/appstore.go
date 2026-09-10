package api

import (
	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	appstoresvc "cloudpivot/internal/service/appstore"

	"github.com/gin-gonic/gin"
)

// AppStoreHandler 应用商店 API Handler
type AppStoreHandler struct {
	svc *appstoresvc.Service
}

// NewAppStoreHandler 创建应用商店 Handler
func NewAppStoreHandler() *AppStoreHandler {
	return &AppStoreHandler{
		svc: appstoresvc.NewService(model.DB),
	}
}

// RegisterAppStoreRoutes 注册应用商店路由
func RegisterAppStoreRoutes(r *gin.RouterGroup) {
	h := NewAppStoreHandler()

	r.GET("/apps", middleware.RequirePermission(model.PermAppStoreView), h.ListApps)
	r.GET("/apps/installed", middleware.RequirePermission(model.PermAppStoreView), h.ListInstalled)
	r.GET("/apps/:key", middleware.RequirePermission(model.PermAppStoreView), h.GetApp)
	r.GET("/apps/:key/status", middleware.RequirePermission(model.PermAppStoreView), h.GetAppStatus)
	r.POST("/apps/:key/install", middleware.RequirePermission(model.PermAppStoreInstall), h.InstallApp)
	r.POST("/apps/:key/uninstall", middleware.RequirePermission(model.PermAppStoreInstall), h.UninstallApp)

	// 应用源管理路由
	RegisterAppSourceRoutes(r)
}

// ListApps GET /apps?category=&keyword=&source_id=
func (h *AppStoreHandler) ListApps(c *gin.Context) {
	category := c.Query("category")
	keyword := c.Query("keyword")
	sourceID := c.Query("source_id")

	apps, err := h.svc.ListApps(category, keyword, sourceID)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取应用列表失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"items": apps, "total": len(apps)})
}

// GetApp GET /apps/:key
func (h *AppStoreHandler) GetApp(c *gin.Context) {
	key := c.Param("key")
	app, err := h.svc.GetApp(key)
	if err != nil {
		common.Fail(c, common.CodeNotFound, err.Error())
		return
	}
	common.Success(c, app)
}

// ListInstalled GET /apps/installed
func (h *AppStoreHandler) ListInstalled(c *gin.Context) {
	apps, err := h.svc.ListInstalled()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取已安装应用失败: "+err.Error())
		return
	}
	common.Success(c, apps)
}

// InstallApp POST /apps/:key/install
func (h *AppStoreHandler) InstallApp(c *gin.Context) {
	key := c.Param("key")

	var req dto.InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空 body（无环境变量）
		req = dto.InstallRequest{EnvParams: make(map[string]string)}
	}
	if req.EnvParams == nil {
		req.EnvParams = make(map[string]string)
	}

	installed, err := h.svc.InstallApp(key, req.SourceID, req.EnvParams)
	if err != nil {
		common.Fail(c, common.CodeServerError, "安装应用失败: "+err.Error())
		return
	}

	containerStatus := ""
	if installed.Status == "running" {
		containerStatus = "running"
	} else if installed.Status == "error" {
		containerStatus = "error"
	}

	common.Success(c, dto.InstalledAppResponse{
		ID:              installed.ID,
		Key:             installed.Key,
		Name:            installed.Name,
		Version:         installed.Version,
		InstallPath:     installed.InstallPath,
		Status:          installed.Status,
		EnvParams:       req.EnvParams,
		ContainerStatus: containerStatus,
		CreatedAt:       installed.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       installed.UpdatedAt.Format("2006-01-02 15:04:05"),
	})
}

// UninstallApp POST /apps/:key/uninstall
func (h *AppStoreHandler) UninstallApp(c *gin.Context) {
	key := c.Param("key")
	if err := h.svc.UninstallApp(key); err != nil {
		common.Fail(c, common.CodeServerError, "卸载应用失败: "+err.Error())
		return
	}
	common.Success(c, nil)
}

// GetAppStatus GET /apps/:key/status
func (h *AppStoreHandler) GetAppStatus(c *gin.Context) {
	key := c.Param("key")
	status, err := h.svc.GetInstallStatus(key)
	if err != nil {
		common.Fail(c, common.CodeNotFound, err.Error())
		return
	}
	common.Success(c, status)
}
