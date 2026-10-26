package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cloudpivot/web"

	"cloudpivot/internal/api"
	"cloudpivot/internal/config"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	aisvc "cloudpivot/internal/service/ai"
	alertservice "cloudpivot/internal/service/alert"
	authsvc "cloudpivot/internal/service/auth"

	containersvc "cloudpivot/internal/service/container"
	cronsvc "cloudpivot/internal/service/cron"
	daemonsvc "cloudpivot/internal/service/daemon"
	deploysvc "cloudpivot/internal/service/deploy"
	monitorsvc "cloudpivot/internal/service/monitor"
	nodesvc "cloudpivot/internal/service/node"
	sitesvc "cloudpivot/internal/service/site"
	systemsvc "cloudpivot/internal/service/system"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化结构化日志
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	// 1. 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		os.Exit(1)
	}
	slog.Info("[犀牛面板] 配置加载成功", "port", cfg.Server.Port, "mode", cfg.Server.Mode)

	// 2. 初始化数据库
	db, err := model.Init(cfg.Database.Path)
	if err != nil {
		slog.Error("数据库初始化失败", "error", err)
		os.Exit(1)
	}
	slog.Info("[犀牛面板] 数据库初始化成功", "path", cfg.Database.Path)

	// 3. 生成/加载 JWT RSA 密钥对
	privKey, pubKey, err := authsvc.LoadOrGenerateKeys(cfg.JWT.KeyDir)
	if err != nil {
		slog.Error("JWT 密钥初始化失败", "error", err)
		os.Exit(1)
	}
	slog.Info("[犀牛面板] JWT RSA 密钥对就绪")

	// 4. 初始化服务
	authService := authsvc.NewService(db, privKey, pubKey, cfg)

	// 初始化管理员
	if err := authService.InitAdmin(); err != nil {
		slog.Error("初始化管理员失败", "error", err)
		os.Exit(1)
	}

	// 安全入口种子：安装脚本（scripts/install.sh）通过 systemd 环境变量传入
	// 随机入口路径，首次启动时写入 settings。仅在「变量非空且库里尚未配置」时生效，
	// 之后的修改一律走设置页 —— 保证环境变量不会覆盖用户在界面上的手动配置。
	if entry := strings.TrimSpace(os.Getenv("SECURITY_ENTRY")); entry != "" {
		if err := seedSecurityEntry(db, entry); err != nil {
			slog.Warn("安全入口种子写入失败（可忽略，之后可在设置页手动配置）", "error", err)
		}
	}

	// 监控采集器
	collector := monitorsvc.NewCollector(db, cfg.Monitor.Interval)
	collector.Start()
	defer collector.Stop()

	monitorService := monitorsvc.NewService(db, collector)

	// 计划任务引擎（先补齐内置任务：SSL 检查/Docker 清理/回收站清理/防篡改/病毒扫描）
	cronsvc.EnsureDefaultTasks(db)
	cronEngine := cronsvc.InitCronEngine(db)
	defer cronEngine.Stop()
	slog.Info("[犀牛面板] 计划任务引擎已启动")

	// 告警引擎
	alertEngine := alertservice.InitAlertEngine(db)
	alertEngine.Start()
	defer alertEngine.Stop()
	slog.Info("[犀牛面板] 告警引擎已启动")

	notifier := alertservice.NewNotifierManager(db)

	// 拨测服务（roadmap #6）
	probeSvc := monitorsvc.NewProbeService(db, notifier)
	probeSvc.Start()
	defer probeSvc.Stop()

	// 多节点管理（roadmap #13）
	nodeSvc := nodesvc.NewService(db)
	nodeSvc.Start()
	defer nodeSvc.Stop()

	// 进程守护（roadmap #11）：崩溃自恢复 + 面板启动时恢复
	daemonSvc := daemonsvc.NewService(db, "")
	daemonSvc.StartService()

	// 新功能服务装配（拨测/守护/Git部署/Compose/构建/文件/AI/防篡改）
	deploySvc := deploysvc.NewService(db, notifier)
	siteSvc := sitesvc.NewService(db)
	// 站点目录按配置注入（默认相对面板工作目录，保证默认可写；
	// 此前硬编码 /www 与 /opt/cloudpivot/runtime/... 在服务器上不存在）
	sitesvc.ConfigureSitePaths(cfg.Site.RootDir, cfg.Site.NginxConfDir, cfg.Site.SSLDir)
	api.SetFeatureServices(
		probeSvc,
		daemonSvc,
		deploySvc,
		containersvc.NewService(),
		api.SharedFileManager(),
		systemsvc.NewService(db),
		siteSvc,
	)
	api.SetNotifier(notifier)
	slog.Info("[犀牛面板] 新功能服务已装配")

	// 5. 初始化路由
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()
	r.Use(middleware.SanitizedLogger(), gin.Recovery()) // 安全日志：query 中 token 脱敏

	// 中间件
	r.Use(middleware.CORS())
	r.Use(middleware.SecurityHeaders())

	// 限流器：登录 10次/分钟，其他 100次/分钟
	limiter := middleware.NewRateLimiter(10, 100)

	// API 路由组
	v1 := r.Group("/api/v1")
	v1.Use(limiter.Limit())
	v1.Use(middleware.Audit())

	// 公开路由（无需认证）
	authHandler := api.NewAuthHandler(authService)
	v1.POST("/auth/login", authHandler.Login)
	v1.POST("/auth/refresh", authHandler.Refresh)
	v1.POST("/auth/forgot-password", authHandler.ForgotPassword)
	v1.POST("/auth/reset-password", authHandler.ResetPassword)

	// 面板间备份接收（X-Panel-Token 认证，BK-008 异地容灾）
	api.RegisterBackupReceiveRoutes(v1)
	// 入口探测：前端用于判断是否启用了安全入口
	v1.GET("/panel/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true, "version": monitorsvc.PanelVersion})
	})

	// 需认证路由
	authGroup := v1.Group("")
	authGroup.Use(middleware.JWTAuth(pubKey))
	// 首次登录强制改密：标志未清除前仅放行改密/信息/登出接口
	authGroup.Use(middleware.EnsurePasswordChanged())
	{
		// 认证
		authGroup.POST("/auth/logout", authHandler.Logout)
		authGroup.GET("/auth/info", authHandler.Info)
		authGroup.PUT("/auth/change-password", authHandler.ChangePassword)

		// Dashboard
		dashboardHandler := api.NewDashboardHandler(monitorService)
		authGroup.GET("/dashboard/overview", middleware.RequirePermission(model.PermMonitorView), dashboardHandler.Overview)
		authGroup.GET("/dashboard/metrics", middleware.RequirePermission(model.PermMonitorView), dashboardHandler.Metrics)
		authGroup.GET("/dashboard/audit-logs", middleware.RequirePermission(model.PermAuditView), dashboardHandler.AuditLogs)

		// 网站管理
		api.RegisterSiteRoutes(authGroup)

		// 数据库管理
		api.RegisterDatabaseRoutes(authGroup)

		// 文件管理器
		api.RegisterFileRoutes(authGroup)

		// Web 终端 (WebSocket)
		api.RegisterTerminalWSRoutes(authGroup)

		// 容器管理
		api.RegisterContainerRoutes(authGroup)

		// 镜像拉取（自定义加速源 / 私有仓库）
		api.RegisterImagePullRoutes(authGroup)

		// 应用商店
		api.RegisterAppStoreRoutes(authGroup)

		// 插件管理
		api.RegisterPluginRoutes(authGroup)

		// 备份恢复
		api.RegisterBackupRoutes(authGroup)

		// 计划任务
		api.RegisterCronRoutes(authGroup)

		// 防火墙
		api.RegisterFirewallRoutes(authGroup)

		// 监控增强
		api.RegisterMonitorRoutes(authGroup)

		// 审计日志
		api.RegisterAuditRoutes(authGroup)

		// 系统设置（含用户管理/角色权限/TOTP）
		api.RegisterSystemRoutes(authGroup)

		// Redis 管理
		api.RegisterRedisRoutes(authGroup)

		// PostgreSQL 管理
		api.RegisterPostgresRoutes(authGroup)

		// 告警通知
		api.RegisterAlertRoutes(authGroup)

		// SSH 安全
		api.RegisterSSHRoutes(authGroup)

		// ===== roadmap 新功能路由 =====
		// 网站拨测
		(&api.ProbeAPI{Svc: probeSvc}).Register(authGroup)
		// 进程守护
		(&api.DaemonAPI{Svc: daemonSvc}).Register(authGroup)
		// Git 推送部署（webhook 免登录，sign 校验）
		(&api.DeployAPI{Svc: deploySvc}).Register(authGroup, v1)
		// Compose 编排 + 镜像构建
		(&api.ComposeAPI{}).Register(authGroup)
		(&api.BuildAPI{}).Register(authGroup)
		// AI 运维助手
		(&api.AIAPI{Svc: aisvc.NewService(db)}).Register(authGroup)
		// 网站防篡改
		(&api.IntegrityAPI{Svc: siteSvc}).Register(authGroup)
		// 文件回收站
		(&api.RecycleAPI{}).Register(authGroup)
		// 面板一键自升级
		(&api.UpgradeAPI{}).Register(authGroup)
		// 多节点管理
		api.RegisterNodeRoutes(authGroup, nodeSvc)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "version": monitorsvc.PanelVersion})
	})

	// 前端 SPA：服务嵌入的静态文件，未匹配路径回退 index.html（history 路由）
	// /api 前缀的未知路径仍返回 404，不回退
	distFS, err := web.Dist()
	if err != nil {
		slog.Error("[犀牛面板] 嵌入前端资源不可用", "error", err)
		os.Exit(1)
	}
	fileServer := http.FileServer(http.FS(distFS))
	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path
		// 含 /api/ 段的路径（可能是安全入口前缀后的 API，如 /admin-xyz/api/v1/...）
		// 一律返回 404 JSON，不做 SPA 回退——否则前端的安全入口探测请求会拿到
		// 一份 index.html，解析 JSON 失败后再串行重试，白等一个 RTT 往返
		// （2026-10-23 实测 /dashboard/api/v1/panel/ping 回退 HTML，首屏探测阻塞 2 秒）
		if isAPIPath(p) || p == "/health" {
			c.JSON(404, gin.H{"code": 404, "message": "not found"})
			return
		}
		// index.html（含 SPA 回退）：注入访问前缀，前端免去运行时探测请求
		if p == "/index.html" {
			c.Data(200, "text/html; charset=utf-8", panelIndexHTML(distFS))
			return
		}
		// 尝试提供真实文件（assets、favicon 等）
		f, err := distFS.Open(strings.TrimPrefix(p, "/"))
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		// SPA 回退
		c.Data(200, "text/html; charset=utf-8", panelIndexHTML(distFS))
	})

	// 6. 启动 HTTP/HTTPS 服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      middleware.SecurityEntryWrapper(r), // 安全入口：路由匹配前剥离/拦截
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	go func() {
		if cfg.Server.TLSEnable && cfg.Server.TLSCertFile != "" && cfg.Server.TLSKeyFile != "" {
			slog.Info("[犀牛面板] HTTPS 服务启动", "addr", addr, "tls", true)
			if err := srv.ListenAndServeTLS(cfg.Server.TLSCertFile, cfg.Server.TLSKeyFile); err != nil && err != http.ErrServerClosed {
				slog.Error("HTTPS 服务错误", "error", err)
				os.Exit(1)
			}
		} else {
			slog.Info("[犀牛面板] 服务启动", "addr", addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("HTTP 服务错误", "error", err)
				os.Exit(1)
			}
		}
	}()

	// 7. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("[犀牛面板] 正在关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("服务强制关闭", "error", err)
		os.Exit(1)
	}

	slog.Info("[犀牛面板] 服务已优雅关闭")
}

// isAPIPath 判断路径中是否包含 /api/ 段。
//
// 除根路径的 /api/ 外，安全入口模式下的 /{entry}/api/v1/... 也属于 API 路径，
// 不应被 SPA 回退成 index.html（前端安全入口探测依赖该路径返回 JSON）。
func isAPIPath(p string) bool {
	if strings.HasPrefix(p, "/api/") {
		return true
	}
	return strings.Contains(p, "/api/")
}

// panelBasePlaceholder index.html 中的访问前缀占位符
const panelBasePlaceholder = "<!--__PANEL_BASE__-->"

// panelIndexHTML 读取 index.html 并注入访问前缀。
//
// 后端本就持有安全入口配置（middleware.GetSecurityEntry），把前缀直接写进
// HTML，前端就不必在首屏发探测请求去猜——那次探测是串行阻塞点，
// 实测可达 1.5~2 秒（2026-10-23）。
func panelIndexHTML(distFS fs.FS) []byte {
	index, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		return []byte("<!doctype html><html><body>frontend assets unavailable</body></html>")
	}
	base := ""
	if entry := middleware.GetSecurityEntry(); entry != "" {
		base = "/" + entry
	}
	injected := `<script>window.__PANEL_BASE__="` + base + `";</script>`
	return []byte(strings.Replace(string(index), panelBasePlaceholder, injected, 1))
}

// seedSecurityEntry 首次启动时把安装脚本传入的安全入口写入 settings
//
// 库里已有值（含空串即用户显式关闭过入口）则不覆盖。
func seedSecurityEntry(db *gorm.DB, entry string) error {
	// 规整：去掉首尾斜杠，避免 /panel-x 与 panel-x 混存
	entry = strings.Trim(entry, "/")
	if entry == "" {
		return nil
	}

	var setting model.Setting
	err := db.Where("`key` = ?", "security_entry").First(&setting).Error
	if err == nil {
		// 已存在：无论值是什么都不覆盖（用户显式配置优先）
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("query security_entry: %w", err)
	}

	setting = model.Setting{
		Key:         "security_entry",
		Value:       entry,
		Description: "安全入口",
		UpdatedAt:   time.Now(),
	}
	if err := db.Create(&setting).Error; err != nil {
		return fmt.Errorf("create security_entry: %w", err)
	}
	slog.Info("[犀牛面板] 安全入口已按安装配置启用", "entry", "/"+entry)
	return nil
}
