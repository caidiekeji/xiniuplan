package api

import (
	"regexp"
	"strconv"
	"strings"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	sitesvc "cloudpivot/internal/service/site"
	nginxpkg "cloudpivot/pkg/nginx"

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
	// 禁用并清除站点 SSL（此前前端用「申请空证书」来变相删除，必被 400 拒绝）
	r.DELETE("/sites/:id/ssl", middleware.RequirePermission(model.PermSiteUpdate), h.DeleteSSL)
	r.GET("/sites/:id/logs", middleware.RequirePermission(model.PermSiteView), h.GetLogs)

	// 站点配置增强（W-006~009）：伪静态/防盗链/重定向/访问控制片段
	r.GET("/sites/rewrite-templates", middleware.RequirePermission(model.PermSiteView), h.ListRewriteTemplates)
	r.GET("/sites/:id/snippets", middleware.RequirePermission(model.PermSiteView), h.ListSnippets)
	r.GET("/sites/:id/traffic", middleware.RequirePermission(model.PermSiteView), h.GetTraffic)
	r.POST("/sites/:id/snippets", middleware.RequirePermission(model.PermSiteUpdate), h.SetSnippet)
	r.DELETE("/sites/:id/snippets/:name", middleware.RequirePermission(model.PermSiteUpdate), h.DeleteSnippet)
	r.POST("/sites/:id/basic-auth", middleware.RequirePermission(model.PermSiteUpdate), h.AddBasicAuthUser)
}

// ListRewriteTemplates GET /sites/rewrite-templates - 伪静态模板列表（W-006）
func (h *SiteHandler) ListRewriteTemplates(c *gin.Context) {
	templates := make([]gin.H, 0, len(nginxpkg.RewriteTemplates))
	for name, content := range nginxpkg.RewriteTemplates {
		templates = append(templates, gin.H{"name": name, "content": content})
	}
	common.Success(c, templates)
}

// ListSnippets GET /sites/:id/snippets
func (h *SiteHandler) ListSnippets(c *gin.Context) {
	id, ok := parseSiteID(c)
	if !ok {
		return
	}
	items, err := h.svc.ListSnippets(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, items)
}

// SetSnippet POST /sites/:id/snippets
// body: {name, content} 或 {type, params...} 由面板生成内容
// type: rewrite(template) | hotlink(domains,extensions,allow_empty) | redirect(from,to,permanent) | acl(mode,entries)
func (h *SiteHandler) SetSnippet(c *gin.Context) {
	id, ok := parseSiteID(c)
	if !ok {
		return
	}
	var req struct {
		Name       string   `json:"name"`
		Content    string   `json:"content"`
		Type       string   `json:"type"`
		Template   string   `json:"template"`
		Domains    []string `json:"domains"`
		Extensions string   `json:"extensions"`
		AllowEmpty bool     `json:"allow_empty"`
		From       string   `json:"from"`
		To         string   `json:"to"`
		Permanent  bool     `json:"permanent"`
		Mode       string   `json:"mode"`
		Entries    []string `json:"entries"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	name := req.Name
	content := req.Content
	switch req.Type {
	case "":
		// 直接写入 content，name 必填
		if name == "" || content == "" {
			common.Fail(c, common.CodeParamError, "name 和 content 不能为空")
			return
		}
	case "rewrite":
		if name == "" {
			name = "rewrite"
		}
		cn, err := nginxpkg.BuildRewriteSnippet(req.Template)
		if err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		content = cn
	case "hotlink":
		cn, err := nginxpkg.BuildHotlinkSnippet(req.Domains, req.Extensions, req.AllowEmpty)
		if err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		content = cn
		if name == "" {
			name = "hotlink"
		}
	case "redirect":
		cn, err := nginxpkg.BuildRedirectSnippet(req.From, req.To, req.Permanent)
		if err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		content = cn
		if name == "" {
			name = "redirect_" + sanitizeSnippetName(req.From)
		}
	case "acl":
		cn, err := nginxpkg.BuildAccessControlSnippet(req.Mode, req.Entries)
		if err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		content = cn
		if name == "" {
			name = "access_control"
		}
	default:
		common.Fail(c, common.CodeParamError, "不支持的片段类型: "+req.Type)
		return
	}

	if err := h.svc.SetSnippet(id, name, content); err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, gin.H{"name": name})
}

// DeleteSnippet DELETE /sites/:id/snippets/:name
func (h *SiteHandler) DeleteSnippet(c *gin.Context) {
	id, ok := parseSiteID(c)
	if !ok {
		return
	}
	name := c.Param("name")
	if err := h.svc.DeleteSnippet(id, name); err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, gin.H{"deleted": true})
}

// AddBasicAuthUser POST /sites/:id/basic-auth（W-009 Basic Auth）
func (h *SiteHandler) AddBasicAuthUser(c *gin.Context) {
	id, ok := parseSiteID(c)
	if !ok {
		return
	}
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	line, err := sitesvc.CreateBasicAuthUser(req.Username, req.Password)
	if err != nil {
		common.Fail(c, common.CodeParamError, err.Error())
		return
	}
	_ = line
	if err := h.svc.AppendBasicAuthUser(id, req.Username, req.Password); err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, gin.H{"added": true})
}

// sanitizeSnippetName 清洗片段名
func sanitizeSnippetName(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	return re.ReplaceAllString(strings.Trim(s, "/"), "_")
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

// List GET /sites - 网站列表（分页 + 筛选）
func (h *SiteHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	// 三个筛选参数此前前端一直在传、后端却没读，导致筛选器全部无效（2026-10-23 审计）
	filter := sitesvc.SiteFilter{
		Domain: c.Query("domain"),
		Type:   c.Query("type"),
		Status: c.Query("status"),
	}

	sites, total, err := h.svc.ListSites(page, pageSize, filter)
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

// DeleteSSL DELETE /sites/:id/ssl - 禁用并清除站点 SSL
func (h *SiteHandler) DeleteSSL(c *gin.Context) {
	id, ok := parseSiteID(c)
	if !ok {
		return
	}
	site, err := h.svc.RemoveSSL(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, "禁用证书失败: "+err.Error())
		return
	}
	common.Success(c, sitesvc.ToResponse(site))
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

// GetTraffic GET /sites/:id/traffic?range=24h|7d|30d - 流量统计（W-018）
func (h *SiteHandler) GetTraffic(c *gin.Context) {
	id, ok := parseSiteID(c)
	if !ok {
		return
	}
	summary, err := h.svc.GetTraffic(id, c.DefaultQuery("range", "24h"))
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, summary)
}
