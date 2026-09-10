package main

import (
	"context"
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
	alertservice "cloudpivot/internal/service/alert"
	authsvc "cloudpivot/internal/service/auth"
	cronsvc "cloudpivot/internal/service/cron"
	monitorsvc "cloudpivot/internal/service/monitor"

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

	// 监控采集器
	collector := monitorsvc.NewCollector(db, cfg.Monitor.Interval)
	collector.Start()
	defer collector.Stop()

	monitorService := monitorsvc.NewService(db, collector)

	// 计划任务引擎
	cronEngine := cronsvc.InitCronEngine(db)
	defer cronEngine.Stop()
	slog.Info("[犀牛面板] 计划任务引擎已启动")

	// 告警引擎
	alertEngine := alertservice.InitAlertEngine(db)
	alertEngine.Start()
	defer alertEngine.Stop()
	slog.Info("[犀牛面板] 告警引擎已启动")

	// 5. 初始化路由
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

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
		if strings.HasPrefix(p, "/api/") || p == "/health" {
			c.JSON(404, gin.H{"code": 404, "message": "not found"})
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
		index, err2 := fs.ReadFile(distFS, "index.html")
		if err2 != nil {
			c.String(500, "frontend assets unavailable")
			return
		}
		c.Data(200, "text/html; charset=utf-8", index)
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
