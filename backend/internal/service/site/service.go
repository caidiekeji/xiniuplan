package site

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/model"
	"cloudpivot/pkg/nginx"

	"gorm.io/gorm"
)

// 域名格式校验
var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// Service 网站管理服务
type Service struct {
	db       *gorm.DB
	executor *common.CommandExecutor
}

// NewService 创建网站管理服务
func NewService(db *gorm.DB) *Service {
	exec := common.NewCommandExecutor(30 * time.Second)

	// 注册 nginx reload 白名单命令
	exec.Register(&common.CommandTemplate{
		Name:        "nginx-reload",
		Description: "Reload nginx configuration",
		Bin:         "nginx",
		Args:        []string{"-s", "reload"},
	})

	// 注册 certbot 白名单命令
	exec.Register(&common.CommandTemplate{
		Name:        "certbot-certonly",
		Description: "Obtain SSL certificate via certbot",
		Bin:         "certbot",
		Args:        []string{"certonly", "--webroot", "-w", "{{webroot}}", "-d", "{{domain}}", "--email", "{{email}}", "--non-interactive", "--agree-tos"},
		AllowedArgs: map[string]bool{"webroot": true, "domain": true, "email": true},
	})

	return &Service{
		db:       db,
		executor: exec,
	}
}

// --- 路径辅助 ---

// nginxConfDir Nginx 配置目录
func nginxConfDir() string {
	if runtime.GOOS == "windows" {
		return "./runtime/nginx/conf.d"
	}
	return "/opt/cloudpivot/runtime/nginx/conf.d"
}

// siteRootDir 网站根目录
func siteRootDir(domain string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join("./runtime/www", domain)
	}
	return filepath.Join("/www", domain)
}

// sslCertDir SSL 证书目录
func sslCertDir(domain string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join("./runtime/ssl", domain)
	}
	return filepath.Join("/opt/cloudpivot/runtime/ssl", domain)
}

// nginxConfPath Nginx 配置文件路径
func nginxConfPath(domain string) string {
	return filepath.Join(nginxConfDir(), domain+".conf")
}

// validateDomain 校验域名格式及安全性
func validateDomain(domain string) error {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return errors.New("domain is empty")
	}
	if !domainRegex.MatchString(domain) {
		return fmt.Errorf("invalid domain format: %s", domain)
	}
	if err := common.ValidateInput(domain); err != nil {
		return fmt.Errorf("domain contains forbidden characters: %w", err)
	}
	return nil
}

// reloadNginx 重新加载 Nginx 配置
func (s *Service) reloadNginx(ctx context.Context) error {
	result, err := s.executor.Execute(ctx, "nginx-reload", nil)
	if err != nil {
		// Windows 开发环境无 nginx，记录日志但不阻断
		if runtime.GOOS == "windows" {
			fmt.Printf("[site-service] nginx reload skipped (windows dev): %v\n", err)
			return nil
		}
		return fmt.Errorf("nginx reload failed: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("nginx reload exit code %d: %s", result.ExitCode, result.Stderr)
	}
	return nil
}

// --- CRUD ---

// CreateSite 创建网站
func (s *Service) CreateSite(req dto.CreateSiteRequest) (*model.Site, error) {
	// 校验主域名
	if err := validateDomain(req.PrimaryDomain); err != nil {
		return nil, err
	}

	// 校验附加域名
	for _, d := range req.Domains {
		if d != "" {
			if err := validateDomain(d); err != nil {
				return nil, fmt.Errorf("domain %s: %w", d, err)
			}
		}
	}

	// 检查域名是否已存在
	var count int64
	if err := s.db.Model(&model.Site{}).Where("primary_domain = ?", req.PrimaryDomain).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("check domain exists: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("domain already exists: %s", req.PrimaryDomain)
	}

	// 确定网站根目录
	rootDir := req.RootDir
	if rootDir == "" {
		rootDir = siteRootDir(req.PrimaryDomain)
	}

	// reverse_proxy 类型：RootDir 存储代理目标地址
	if req.Type == "reverse_proxy" && req.ProxyTarget != "" {
		rootDir = req.ProxyTarget
	}

	// 创建网站根目录（reverse_proxy 也创建日志目录）
	siteRoot := siteRootDir(req.PrimaryDomain)
	logsDir := filepath.Join(siteRoot, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("create site logs dir: %w", err)
	}

	// 静态站点生成默认 index.html
	if req.Type == "static" {
		indexPath := filepath.Join(siteRoot, "index.html")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			defaultIndex := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); }
        .container { text-align: center; color: white; }
        h1 { font-size: 2.5em; margin-bottom: 0.5em; }
        p { font-size: 1.2em; opacity: 0.9; }
    </style>
</head>
<body>
    <div class="container">
        <h1>%s</h1>
        <p>网站已成功创建，由犀牛面板提供支持</p>
    </div>
</body>
</html>`, req.PrimaryDomain, req.PrimaryDomain)
			if err := os.WriteFile(indexPath, []byte(defaultIndex), 0644); err != nil {
				return nil, fmt.Errorf("write default index.html: %w", err)
			}
		}
	}

	// 构建 Site 模型
	domains := model.StringSlice{}
	for _, d := range req.Domains {
		if d != "" {
			domains = append(domains, d)
		}
	}

	site := &model.Site{
		PrimaryDomain: req.PrimaryDomain,
		Domains:       domains,
		Type:          req.Type,
		Runtime:       req.Runtime,
		RootDir:       rootDir,
		NginxConf:     nginxConfPath(req.PrimaryDomain),
		Status:        "running",
		Remark:        req.Remark,
	}

	// 生成 Nginx 配置
	confContent, err := nginx.Generate(site)
	if err != nil {
		return nil, fmt.Errorf("generate nginx config: %w", err)
	}

	// 写入 Nginx 配置文件
	if err := os.MkdirAll(nginxConfDir(), 0755); err != nil {
		return nil, fmt.Errorf("create nginx conf dir: %w", err)
	}
	if err := os.WriteFile(site.NginxConf, []byte(confContent), 0644); err != nil {
		return nil, fmt.Errorf("write nginx config: %w", err)
	}

	// 保存到数据库
	if err := s.db.Create(site).Error; err != nil {
		return nil, fmt.Errorf("create site in db: %w", err)
	}

	// 重载 Nginx
	if err := s.reloadNginx(context.Background()); err != nil {
		// 不阻断创建，仅记录
		fmt.Printf("[site-service] nginx reload after create failed: %v\n", err)
	}

	return site, nil
}

// GetSite 获取网站详情
func (s *Service) GetSite(id uint) (*model.Site, error) {
	var site model.Site
	if err := s.db.First(&site, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("site not found")
		}
		return nil, fmt.Errorf("query site: %w", err)
	}
	return &site, nil
}

// ListSites 分页获取网站列表
func (s *Service) ListSites(page, pageSize int) ([]model.Site, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	sites := make([]model.Site, 0)
	var total int64

	if err := s.db.Model(&model.Site{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count sites: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := s.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&sites).Error; err != nil {
		return nil, 0, fmt.Errorf("query sites: %w", err)
	}

	return sites, total, nil
}

// UpdateSite 更新网站
func (s *Service) UpdateSite(id uint, req dto.UpdateSiteRequest) (*model.Site, error) {
	site, err := s.GetSite(id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.PrimaryDomain != nil {
		if err := validateDomain(*req.PrimaryDomain); err != nil {
			return nil, err
		}
		// 检查新域名是否被其他站点使用
		var count int64
		s.db.Model(&model.Site{}).Where("primary_domain = ? AND id != ?", *req.PrimaryDomain, id).Count(&count)
		if count > 0 {
			return nil, fmt.Errorf("domain already exists: %s", *req.PrimaryDomain)
		}
		site.PrimaryDomain = *req.PrimaryDomain
	}

	if req.Domains != nil {
		domains := model.StringSlice{}
		for _, d := range req.Domains {
			if d != "" {
				if err := validateDomain(d); err != nil {
					return nil, fmt.Errorf("domain %s: %w", d, err)
				}
				domains = append(domains, d)
			}
		}
		site.Domains = domains
	}

	if req.Type != nil {
		site.Type = *req.Type
	}

	if req.Runtime != nil {
		site.Runtime = *req.Runtime
	}

	if req.Remark != nil {
		site.Remark = *req.Remark
	}

	// 处理 root_dir / proxy_target
	if req.Type != nil && *req.Type == "reverse_proxy" {
		if req.ProxyTarget != nil {
			site.RootDir = *req.ProxyTarget
		}
	} else if req.RootDir != nil {
		site.RootDir = *req.RootDir
	}

	// 更新 Nginx 配置路径（域名可能变了）
	site.NginxConf = nginxConfPath(site.PrimaryDomain)

	// 重新生成 Nginx 配置
	confContent, err := nginx.Generate(site)
	if err != nil {
		return nil, fmt.Errorf("generate nginx config: %w", err)
	}
	if err := os.MkdirAll(nginxConfDir(), 0755); err != nil {
		return nil, fmt.Errorf("create nginx conf dir: %w", err)
	}
	if err := os.WriteFile(site.NginxConf, []byte(confContent), 0644); err != nil {
		return nil, fmt.Errorf("write nginx config: %w", err)
	}

	// 保存到数据库
	if err := s.db.Save(site).Error; err != nil {
		return nil, fmt.Errorf("update site in db: %w", err)
	}

	// 重载 Nginx
	if err := s.reloadNginx(context.Background()); err != nil {
		fmt.Printf("[site-service] nginx reload after update failed: %v\n", err)
	}

	return site, nil
}

// DeleteSite 删除网站
func (s *Service) DeleteSite(id uint, deleteFiles bool) error {
	site, err := s.GetSite(id)
	if err != nil {
		return err
	}

	// 删除 Nginx 配置文件
	if site.NginxConf != "" {
		os.Remove(site.NginxConf)
		// 同时删除可能存在的 .disabled 文件
		os.Remove(site.NginxConf + ".disabled")
	}

	// 删除网站文件
	if deleteFiles {
		siteRoot := siteRootDir(site.PrimaryDomain)
		os.RemoveAll(siteRoot)

		// 删除 SSL 证书
		sslDir := sslCertDir(site.PrimaryDomain)
		os.RemoveAll(sslDir)
	}

	// 删除数据库记录
	if err := s.db.Delete(site).Error; err != nil {
		return fmt.Errorf("delete site from db: %w", err)
	}

	// 重载 Nginx
	if err := s.reloadNginx(context.Background()); err != nil {
		fmt.Printf("[site-service] nginx reload after delete failed: %v\n", err)
	}

	return nil
}

// --- 启停 ---

// StartSite 启动网站
func (s *Service) StartSite(id uint) (*model.Site, error) {
	site, err := s.GetSite(id)
	if err != nil {
		return nil, err
	}

	// 如果配置被禁用，恢复它
	disabledPath := site.NginxConf + ".disabled"
	if _, err := os.Stat(disabledPath); err == nil {
		if err := os.Rename(disabledPath, site.NginxConf); err != nil {
			return nil, fmt.Errorf("enable nginx config: %w", err)
		}
	}

	site.Status = "running"
	if err := s.db.Save(site).Error; err != nil {
		return nil, fmt.Errorf("update site status: %w", err)
	}

	if err := s.reloadNginx(context.Background()); err != nil {
		return nil, err
	}

	return site, nil
}

// StopSite 停止网站
func (s *Service) StopSite(id uint) (*model.Site, error) {
	site, err := s.GetSite(id)
	if err != nil {
		return nil, err
	}

	// 禁用配置文件（重命名为 .disabled）
	if _, err := os.Stat(site.NginxConf); err == nil {
		if err := os.Rename(site.NginxConf, site.NginxConf+".disabled"); err != nil {
			return nil, fmt.Errorf("disable nginx config: %w", err)
		}
	}

	site.Status = "stopped"
	if err := s.db.Save(site).Error; err != nil {
		return nil, fmt.Errorf("update site status: %w", err)
	}

	if err := s.reloadNginx(context.Background()); err != nil {
		return nil, err
	}

	return site, nil
}

// RestartSite 重启网站
func (s *Service) RestartSite(id uint) (*model.Site, error) {
	site, err := s.GetSite(id)
	if err != nil {
		return nil, err
	}

	if err := s.reloadNginx(context.Background()); err != nil {
		return nil, err
	}

	site.Status = "running"
	s.db.Save(site)

	return site, nil
}

// --- Nginx 配置管理 ---

// GetNginxConfig 获取 Nginx 配置内容
func (s *Service) GetNginxConfig(id uint) (string, error) {
	site, err := s.GetSite(id)
	if err != nil {
		return "", err
	}

	confPath := site.NginxConf
	// 如果配置被禁用，读取 .disabled 文件
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		confPath = confPath + ".disabled"
	}

	content, err := os.ReadFile(confPath)
	if err != nil {
		return "", fmt.Errorf("read nginx config: %w", err)
	}

	return string(content), nil
}

// UpdateNginxConfig 更新 Nginx 配置
func (s *Service) UpdateNginxConfig(id uint, content string) error {
	site, err := s.GetSite(id)
	if err != nil {
		return err
	}

	// 校验配置
	if err := nginx.Validate(content); err != nil {
		return fmt.Errorf("nginx config validation failed: %w", err)
	}

	// 写入配置
	confPath := site.NginxConf
	if err := os.MkdirAll(filepath.Dir(confPath), 0755); err != nil {
		return fmt.Errorf("create nginx conf dir: %w", err)
	}
	if err := os.WriteFile(confPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write nginx config: %w", err)
	}

	// 重载 Nginx
	if err := s.reloadNginx(context.Background()); err != nil {
		return err
	}

	return nil
}

// --- 日志 ---

// GetSiteLogs 读取网站日志尾部
func (s *Service) GetSiteLogs(id uint, logType string, lines int) (string, error) {
	site, err := s.GetSite(id)
	if err != nil {
		return "", err
	}

	if lines <= 0 {
		lines = 100
	}
	if lines > 1000 {
		lines = 1000
	}

	logFile := filepath.Join(siteRootDir(site.PrimaryDomain), "logs", logType+".log")

	// 如果文件不存在，返回空
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return "", nil
	}

	content, err := tailFile(logFile, lines)
	if err != nil {
		return "", fmt.Errorf("read log file: %w", err)
	}

	return content, nil
}

// tailFile 读取文件最后 N 行
func tailFile(path string, n int) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}

// ToResponse 将 model.Site 转换为 dto.SiteResponse
func ToResponse(site *model.Site) dto.SiteResponse {
	resp := dto.SiteResponse{
		ID:            site.ID,
		PrimaryDomain: site.PrimaryDomain,
		Domains:       []string(site.Domains),
		Type:          site.Type,
		Runtime:       site.Runtime,
		RootDir:       site.RootDir,
		NginxConf:     site.NginxConf,
		Status:        site.Status,
		ContainerID:   site.ContainerID,
		Remark:        site.Remark,
		CreatedAt:     site.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     site.UpdatedAt.Format(time.RFC3339),
		SSL: dto.SiteSSLInfo{
			Enabled:   site.SSLEnabled,
			Provider:  site.SSLProvider,
			CertPath:  site.SSLCertPath,
			KeyPath:   site.SSLKeyPath,
			AutoRenew: site.SSLAutoRenew,
		},
	}

	if site.SSLExpireAt != nil {
		resp.SSL.ExpireAt = site.SSLExpireAt.Format(time.RFC3339)
		resp.SSL.DaysLeft = int(time.Until(*site.SSLExpireAt).Hours() / 24)
	}

	return resp
}
