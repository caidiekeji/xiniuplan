package api

import (
	"strconv"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	sitesvc "cloudpivot/internal/service/site"

	"github.com/gin-gonic/gin"
)

// SiteHandler 网站管理 API Handler
type SiteHandler struct {
	svc *sitesvc.Service
}

// NewSiteHandler 创建网站管理 Handler
func NewSiteHandler(svc *sitesvc.Service) *SiteHandler {
	return &SiteHandler{svc: svc}
}

// RegisterSiteRoutes 注册网站管理路由
func RegisterSiteRoutes(r *gin.RouterGroup) {
	svc := sitesvc.NewService(model.DB)
	h := NewSiteHandler(svc)

	r.GET("/sites", middleware.RequirePermission(model.PermSiteView), h.List)
	r.POST("/sites", middleware.RequirePermission(model.PermSiteCreate), h.Create)
	r.GET("/sites/:id", middleware.RequirePermission(model.PermSiteView), h.Get)
	r.PUT("/sites/:id", middleware.RequirePermission(model.PermSiteUpdate), h.Update)
	r.DELETE("/sites/:id", middleware.RequirePermission(model.PermSiteDelete), h.Delete)
	r.POST("/sites/:id/start", middleware.RequirePermission(model.PermSiteUpdate), h.Start)
	r.POST("/sites/:id/stop", middleware.RequirePermission(model.PermSiteUpdate), h.Stop)
	r.POST("/sites/:id/restart", middleware.RequirePermission(model.PermSiteUpdate), h.Restart)
	r.GET("/sites/:id/config", middleware.RequirePermission(model.PermSiteView), h.GetConfig)
	r.PUT("/sites/:id/config", middleware.RequirePermission(model.PermSiteUpdate), h.UpdateConfig)
	r.POST("/sites/:id/ssl/apply", middleware.RequirePermission(model.PermSiteUpdate), h.ApplySSL)
	r.GET("/sites/:id/ssl", middleware.RequirePermission(model.PermSiteView), h.GetSSL)
	r.GET("/sites/:id/logs", middleware.RequirePermission(model.PermSiteView), h.GetLogs)
}

// parseSiteID 从路径参数解析站点 ID
func parseSiteID(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.Fail(c, common.CodeParamError, "无效的站点 ID")
		return 0, false
	}
	return uint(id), true
}

// List GET /sites - 网站列表（分页）
func (h *SiteHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	sites, total, err := h.svc.ListSites(page, pageSize)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取网站列表失败: "+err.Error())
		return
	}

	items := make([]dto.SiteResponse, 0, len(sites))
	for i := range sites {
		items = append(items, sitesvc.ToResponse(&sites[i]))
	}

	common.Page(c, items, total, page, pageSize)
}

// Create POST /sites - 创建网站
func (h *SiteHandler) Create(c *gin.Context) {
	var req dto.CreateSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	site, err := h.svc.CreateSite(req)
	if err != nil {
		common.Fail(c, common.CodeServerError, "创建网站失败: "+err.Error())
		return
	}

	common.Success(c, sitesvc.ToResponse(site))
}

// Get GET /sites/:id - 网站详情
func (h *SiteHandler) Get(c *gin.Context) {
	id, ok := parseSiteID(c)
	if !ok {
		return
	}

	site, err := h.svc.GetSite(id)
	if err != nil {
		common.Fail(c, common.CodeNotFound, err.Error())
		return
	}

	common.Success(c, sitesvc.ToResponse(site))
}

// Update PUT /sites/:id - 更新网站
func (h *SiteHandler) Update(c *gin.Context) {
	id, ok := parseSiteID(c)
	if !ok {
		return
	}

	var req dto.UpdateSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	site, err := h.svc.UpdateSite(id, req)
	if err != nil {
		common.Fail(c, common.CodeServerError, "更新网站失败: "+err.Error())
		return
	}

	common.Success(c, sitesvc.ToResponse(site))
}

// Delete DELETE /sites/:id - 删除网站
func (h *SiteHandler) Delete(c *gin.Context) {
	id, ok := parseSiteID(c)
	if !ok {
		return
	}

	deleteFiles := c.Query("delete_files") == "true"

	if err := h.svc.DeleteSite(id, deleteFiles); err != nil {
		common.Fail(c, common.CodeServerError, "删除网站失败: "+err.Error())
		return
	}

	common.Success(c, gin.H{"deleted": true, "delete_files": deleteFiles})
}

// Start POST /sites/:id/start - 启动网站
func (h *SiteHandler) Start(c *gin.Context) {
	id, ok := parseSiteID(c)
	if !ok {
		return
	}

	site, err := h.svc.StartSite(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, "启动网站失败: "+err.Error())
		return
	}

	common.Success(c, sitesvc.ToResponse(site))
}

// Stop POST /sites/:id/stop - 停止网站
func (h *SiteHandler) Stop(c *gin.Context) {
	id, ok := parseSiteID(c)
	if !ok {
		return
	}

	site, err := h.svc.StopSite(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, "停止网站失败: "+err.Error())
		return
	}

	common.Success(c, sitesvc.ToResponse(site))
}

// Restart POST /sites/:id/restart - 重启网站
func (h *SiteHandler) Restart(c *gin.Context) {
	id, ok := parseSiteID(c)
	if !ok {
		return
	}

	site, err := h.svc.RestartSite(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, "重启网站失败: "+err.Error())
		return
	}

	common.Success(c, sitesvc.ToResponse(site))
}

// GetConfig GET /sites/:id/config - 获取 Nginx 配置
func (h *SiteHandler) GetConfig(c *gin.Context) {
	id, ok := parseSiteID(c)
	if !ok {
		return
	}

	content, err := h.svc.GetNginxConfig(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取配置失败: "+err.Error())
		return
	}

	site, _ := h.svc.GetSite(id)
	common.Success(c, dto.NginxConfigResponse{
		Domain:  site.PrimaryDomain,
		Content: content,
	})
}

// UpdateConfig PUT /sites/:id/config - 更新 Nginx 配置
func (h *SiteHandler) UpdateConfig(c *gin.Context) {
	id, ok := parseSiteID(c)
	if !ok {
		return
	}

	var req dto.NginxConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.UpdateNginxConfig(id, req.Content); err != nil {
		common.Fail(c, common.CodeServerError, "更新配置失败: "+err.Error())
		return
	}

	common.Success(c, gin.H{"updated": true})
}

// ApplySSL POST /sites/:id/ssl/apply - 申请 SSL
func (h *SiteHandler) ApplySSL(c *gin.Context) {
	id, ok := parseSiteID(c)
	if !ok {
		return
	}

	var req dto.SSLApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	switch req.Provider {
	case "manual":
		if req.Cert == "" || req.Key == "" {
			common.Fail(c, common.CodeParamError, "手动上传证书需要 cert 和 key 字段")
			return
		}
		site, err := h.svc.ApplyManualSSL(id, req.Cert, req.Key)
		if err != nil {
			common.Fail(c, common.CodeServerError, "应用证书失败: "+err.Error())
			return
		}
		common.Success(c, sitesvc.ToResponse(site))

	case "letsencrypt":
		if req.Email == "" {
			common.Fail(c, common.CodeParamError, "Let's Encrypt 需要 email 字段")
			return
		}
		site, status, err := h.svc.ApplyLetsEncrypt(id, req.Email)
		if err != nil {
			common.Fail(c, common.CodeServerError, "申请证书失败: "+err.Error())
			return
		}
		common.Success(c, gin.H{
			"site":   sitesvc.ToResponse(site),
			"status": status,
		})

	default:
		common.Fail(c, common.CodeParamError, "不支持的 SSL 提供商: "+req.Provider)
	}
}

// GetSSL GET /sites/:id/ssl - 获取 SSL 信息
func (h *SiteHandler) GetSSL(c *gin.Context) {
	id, ok := parseSiteID(c)
	if !ok {
		return
	}

	info, err := h.svc.GetSSLInfo(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取 SSL 信息失败: "+err.Error())
		return
	}

	common.Success(c, info)
}

// GetLogs GET /sites/:id/logs - 访问日志
func (h *SiteHandler) GetLogs(c *gin.Context) {
	id, ok := parseSiteID(c)
	if !ok {
		return
	}

	logType := c.DefaultQuery("type", "access")
	if logType != "access" && logType != "error" {
		common.Fail(c, common.CodeParamError, "日志类型只能是 access 或 error")
		return
	}

	lines, _ := strconv.Atoi(c.DefaultQuery("lines", "100"))

	content, err := h.svc.GetSiteLogs(id, logType, lines)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取日志失败: "+err.Error())
		return
	}

	common.Success(c, dto.SiteLogResponse{
		Type:    logType,
		Lines:   lines,
		Content: content,
	})
}
