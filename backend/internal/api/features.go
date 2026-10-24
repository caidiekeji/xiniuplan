package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	"cloudpivot/internal/service/ai"
	"cloudpivot/internal/service/alert"
	"cloudpivot/internal/service/container"
	"cloudpivot/internal/service/daemon"
	"cloudpivot/internal/service/deploy"
	filepkg "cloudpivot/internal/service/file"
	"cloudpivot/internal/service/monitor"
	sitesvc "cloudpivot/internal/service/site"
	"cloudpivot/internal/service/system"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ===== 服务实例注入（main.go 装配时设置）=====
var (
	probeSvc        *monitor.ProbeService
	daemonSvc       *daemon.Service
	deploySvc       *deploy.Service
	containerSvcRef *container.Service
	fileSvcRef      *filepkg.Manager
	systemSvcRef    *system.Service
	siteSvcRef      *sitesvc.Service
)

// SetFeatureServices main.go 装配
func SetFeatureServices(probe *monitor.ProbeService, dm *daemon.Service, dp *deploy.Service,
	container *container.Service, file *filepkg.Manager, system *system.Service, site *sitesvc.Service) {
	probeSvc, daemonSvc, deploySvc = probe, dm, dp
	containerSvcRef, fileSvcRef, systemSvcRef, siteSvcRef = container, file, system, site
}

// SetNotifier 注入通知管理器（防篡改即时告警等场景）
func SetNotifier(nm *alert.NotifierManager) { notifierMgr = nm }

var notifierMgr *alert.NotifierManager

func containerSvc() *container.Service { return containerSvcRef }
func fileSvc() *filepkg.Manager        { return fileSvcRef }
func systemSvc() *system.Service       { return systemSvcRef }

func findSite(id uint) (*model.Site, error) {
	var site model.Site
	if err := model.DB.First(&site, id).Error; err != nil {
		return nil, fmt.Errorf("站点不存在")
	}
	return &site, nil
}

// ============ 新功能路由注册（roadmap v1.1/v1.2 全量实现） ============
// 拨测 / 进程守护 / Git 部署 / Compose 编排 / 镜像构建 / AI 诊断 / 防篡改 / 回收站 / 面板升级

// ProbeAPI 拨测
type ProbeAPI struct{ Svc *monitor.ProbeService }

// RegisterProbeRoutes 注册拨测路由
func (h *ProbeAPI) Register(r *gin.RouterGroup) {
	g := r.Group("/probes", middleware.RequirePermission(model.PermMonitorView))
	g.GET("", func(c *gin.Context) {
		list, err := h.Svc.List()
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"items": list})
	})
	g.POST("", middleware.RequirePermission(model.PermMonitorUpdate), func(c *gin.Context) {
		var p model.ProbeMonitor
		if err := c.ShouldBindJSON(&p); err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		if err := h.Svc.Create(&p); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, p)
	})
	g.PUT("/:id", middleware.RequirePermission(model.PermMonitorUpdate), func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var p model.ProbeMonitor
		if err := c.ShouldBindJSON(&p); err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		if err := h.Svc.Update(uint(id), &p); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"updated": true})
	})
	g.DELETE("/:id", middleware.RequirePermission(model.PermMonitorUpdate), func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := h.Svc.Delete(uint(id)); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"deleted": true})
	})
	g.POST("/:id/check", middleware.RequirePermission(model.PermMonitorUpdate), func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := h.Svc.CheckNow(uint(id)); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"checked": true})
	})
	g.GET("/:id/results", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
		list, err := h.Svc.Results(uint(id), limit)
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"items": list})
	})
}

// DaemonAPI 进程守护
type DaemonAPI struct{ Svc *daemon.Service }

// RegisterDaemonRoutes 注册进程守护路由
func (h *DaemonAPI) Register(r *gin.RouterGroup) {
	g := r.Group("/daemon", middleware.RequirePermission(model.PermMonitorView))
	g.GET("", func(c *gin.Context) {
		list, err := h.Svc.List()
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"items": list})
	})
	g.POST("", middleware.RequirePermission(model.PermMonitorUpdate), func(c *gin.Context) {
		var p model.DaemonProcess
		if err := c.ShouldBindJSON(&p); err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		if err := h.Svc.Create(&p); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, p)
	})
	g.PUT("/:id", middleware.RequirePermission(model.PermMonitorUpdate), func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var p model.DaemonProcess
		if err := c.ShouldBindJSON(&p); err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		if err := h.Svc.Update(uint(id), &p); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"updated": true})
	})
	g.DELETE("/:id", middleware.RequirePermission(model.PermMonitorUpdate), func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := h.Svc.Delete(uint(id)); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"deleted": true})
	})
	g.POST("/:id/start", middleware.RequirePermission(model.PermMonitorUpdate), func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		p, err := h.Svc.Launch(uint(id))
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, p)
	})
	g.POST("/:id/stop", middleware.RequirePermission(model.PermMonitorUpdate), func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := h.Svc.StopProcess(uint(id)); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"stopped": true})
	})
	g.GET("/:id/logs", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		lines, _ := strconv.Atoi(c.DefaultQuery("lines", "200"))
		logs, err := h.Svc.Logs(uint(id), lines)
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"content": logs})
	})
}

// DeployAPI Git 部署
type DeployAPI struct{ Svc *deploy.Service }

// RegisterDeployRoutes 注册 Git 部署路由
func (h *DeployAPI) Register(authed *gin.RouterGroup, public *gin.RouterGroup) {
	g := authed.Group("/deploy/apps", middleware.RequirePermission(model.PermSiteView))
	g.GET("", func(c *gin.Context) {
		list, err := h.Svc.List()
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"items": list})
	})
	g.POST("", middleware.RequirePermission(model.PermSiteUpdate), func(c *gin.Context) {
		var app model.DeployApp
		if err := c.ShouldBindJSON(&app); err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		if err := h.Svc.Create(&app); err != nil {
			// 创建失败都是参数问题（空字段、非法路径、非法类型），
			// 返回 400 让前端能展示具体原因，而不是系统异常
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		common.Success(c, app)
	})
	g.PUT("/:id", middleware.RequirePermission(model.PermSiteUpdate), func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		// 用专门的更新 DTO：Enabled 为 *bool，未传时不会被零值 false 覆盖
		var req dto.DeployAppUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		if err := h.Svc.Update(uint(id), &req); err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		common.Success(c, gin.H{"updated": true})
	})
	g.DELETE("/:id", middleware.RequirePermission(model.PermSiteUpdate), func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := h.Svc.Delete(uint(id)); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"deleted": true})
	})
	g.POST("/:id/run", middleware.RequirePermission(model.PermSiteUpdate), func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		output, err := h.Svc.Deploy(uint(id))
		if err != nil {
			common.Fail(c, common.CodeServerError, output+"\n"+err.Error())
			return
		}
		common.Success(c, gin.H{"output": output})
	})
	g.GET("/:id/logs", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		list, err := h.Svc.Logs(uint(id), limit)
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"items": list})
	})
	// Webhook（免登录，secret 校验）：/deploy/webhook/:id?sign=xxx
	public.POST("/deploy/webhook/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var app model.DeployApp
		if err := db().First(&app, id).Error; err != nil {
			common.Fail(c, common.CodeParamError, "应用不存在")
			return
		}
		sign := c.Query("sign")
		if sign == "" {
			sign = c.GetHeader("X-Webhook-Sign")
		}
		body, _ := c.GetRawData()
		if !h.Svc.VerifyWebhook(&app, body, sign) {
			common.Fail(c, common.CodeUnauthorized, "webhook 签名校验失败")
			return
		}
		// 异步部署，立即响应
		go h.Svc.Deploy(uint(id))
		common.Success(c, gin.H{"triggered": true})
	})
}

// db 供 webhook 查 app（避免循环依赖，直接用全局）
func db() *gorm.DB { return model.DB }

// ComposeAPI Compose 编排
type ComposeAPI struct{}

// RegisterComposeRoutes 注册 Compose 编排路由
func (h *ComposeAPI) Register(r *gin.RouterGroup) {
	g := r.Group("/compose", middleware.RequirePermission(model.PermContainerView))
	g.GET("", func(c *gin.Context) {
		list, err := containerSvc().ListComposeStacks()
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"items": list})
	})
	g.GET("/:name", func(c *gin.Context) {
		info, err := containerSvc().GetComposeStack(c.Param("name"))
		if err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		common.Success(c, info)
	})
	g.POST("", middleware.RequirePermission(model.PermContainerUpdate), func(c *gin.Context) {
		var req struct{ Name, Content string }
		if err := c.ShouldBindJSON(&req); err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		if err := containerSvc().SaveComposeStack(req.Name, req.Content); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"saved": true})
	})
	g.DELETE("/:name", middleware.RequirePermission(model.PermContainerUpdate), func(c *gin.Context) {
		out, err := containerSvc().DeleteComposeStack(c.Param("name"))
		if err != nil {
			common.Fail(c, common.CodeServerError, out+"\n"+err.Error())
			return
		}
		common.Success(c, gin.H{"output": out})
	})
	g.POST("/:name/:action", middleware.RequirePermission(model.PermContainerUpdate), func(c *gin.Context) {
		out, err := containerSvc().ComposeAction(c.Param("name"), c.Param("action"))
		if err != nil {
			common.Fail(c, common.CodeServerError, out+"\n"+err.Error())
			return
		}
		common.Success(c, gin.H{"output": out})
	})
}

// BuildAPI 镜像构建（流式日志）
type BuildAPI struct{}

// RegisterBuildRoutes 注册镜像构建路由
func (h *BuildAPI) Register(r *gin.RouterGroup) {
	r.POST("/images/build", middleware.RequirePermission(model.PermContainerUpdate), func(c *gin.Context) {
		var req struct{ Tag, Dockerfile string }
		if err := c.ShouldBindJSON(&req); err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		// 流式响应
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.Header("X-Accel-Buffering", "no")
		flusher, canFlush := c.Writer.(http.Flusher)
		err := containerSvc().BuildImage(req.Tag, req.Dockerfile, func(line string) {
			fmt.Fprintln(c.Writer, line)
			if canFlush {
				flusher.Flush()
			}
		})
		if err != nil {
			fmt.Fprintln(c.Writer, "[ERROR] "+err.Error())
		} else {
			fmt.Fprintln(c.Writer, "[DONE] 构建完成: "+req.Tag)
		}
		if canFlush {
			flusher.Flush()
		}
	})
}

// AIAPI AI 诊断
type AIAPI struct{ Svc *ai.Service }

// RegisterAIRoutes 注册 AI 诊断路由
func (h *AIAPI) Register(r *gin.RouterGroup) {
	r.POST("/ai/diagnose", middleware.RequirePermission(model.PermSettingUpdate), func(c *gin.Context) {
		var req ai.DiagnoseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		result, err := h.Svc.Diagnose(&req)
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, result)
	})
}

// IntegrityAPI 网站防篡改（复用 site 包 tamper 实现，cron tamper_check 同源）
type IntegrityAPI struct{ Svc *sitesvc.Service }

// RegisterIntegrityRoutes 注册防篡改路由（挂在 sites 组）
func (h *IntegrityAPI) Register(r *gin.RouterGroup) {
	g := r.Group("/sites/:id/integrity", middleware.RequirePermission(model.PermSiteUpdate))
	g.GET("/status", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var count int64
		db().Model(&model.SiteTamperBaseline{}).Where("site_id = ?", id).Count(&count)
		common.Success(c, gin.H{"has_baseline": count > 0, "files": count})
	})
	g.POST("/baseline", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if _, err := findSite(uint(id)); err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		count, err := h.Svc.BuildTamperBaseline(uint(id))
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"files": count, "message": fmt.Sprintf("基线已建立，共 %d 个文件", count)})
	})
	g.GET("/check", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if _, err := findSite(uint(id)); err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		changes, err := h.Svc.VerifyTamper(uint(id))
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		// 可选：差异即时告警（notify=1）
		if c.Query("notify") == "1" && len(changes) > 0 {
			site, _ := findSite(uint(id))
			lines := make([]string, 0, len(changes))
			for _, ch := range changes {
				lines = append(lines, fmt.Sprintf("  %s (%s)", ch.Path, ch.Status))
			}
			_ = notifierMgr.SendToAll(fmt.Sprintf("网站防篡改告警：%s", site.PrimaryDomain),
				fmt.Sprintf("发现 %d 处文件变更：\n%s", len(changes), strings.Join(lines, "\n")))
		}
		common.Success(c, gin.H{"changes": changes, "count": len(changes)})
	})
}

// RecycleAPI 文件回收站
type RecycleAPI struct{}

// RegisterRecycleRoutes 注册回收站路由
func (h *RecycleAPI) Register(r *gin.RouterGroup) {
	g := r.Group("/files/recycle", middleware.RequirePermission(model.PermFileView))
	g.GET("", func(c *gin.Context) {
		items, err := fileSvc().ListRecycle(model.DB)
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"items": items})
	})
	g.POST("/restore", middleware.RequirePermission(model.PermFileUpdate), func(c *gin.Context) {
		var req struct{ ID uint }
		if err := c.ShouldBindJSON(&req); err != nil || req.ID == 0 {
			common.Fail(c, common.CodeParamError, "缺少回收站条目 ID")
			return
		}
		if err := fileSvc().RestoreFromRecycle(req.ID, model.DB); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"restored": true})
	})
	g.DELETE("", middleware.RequirePermission(model.PermFileUpdate), func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
		if id == 0 {
			common.Fail(c, common.CodeParamError, "缺少回收站条目 ID")
			return
		}
		if err := fileSvc().PurgeRecycle(uint(id), model.DB); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"purged": true})
	})
}

// UpgradeAPI 面板自升级
type UpgradeAPI struct{}

// RegisterUpgradeRoutes 注册面板升级路由
func (h *UpgradeAPI) Register(r *gin.RouterGroup) {
	r.POST("/system/upgrade", middleware.RequirePermission(model.PermSettingUpdate), func(c *gin.Context) {
		result, err := systemSvc().Upgrade()
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, result)
	})
}
