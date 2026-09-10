package site

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloudpivot/internal/dto"
	"cloudpivot/internal/model"
	"cloudpivot/pkg/nginx"
)

// ApplyManualSSL 手动上传 SSL 证书
func (s *Service) ApplyManualSSL(siteID uint, certContent, keyContent string) (*model.Site, error) {
	site, err := s.GetSite(siteID)
	if err != nil {
		return nil, err
	}

	certContent = strings.TrimSpace(certContent)
	keyContent = strings.TrimSpace(keyContent)

	if certContent == "" || keyContent == "" {
		return nil, errors.New("certificate and key content cannot be empty")
	}

	// 解析证书获取有效期
	expireAt, issuer, err := parseCertInfo(certContent)
	if err != nil {
		return nil, fmt.Errorf("parse certificate failed: %w", err)
	}

	// 写入证书文件
	certDir := sslCertDir(site.PrimaryDomain)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return nil, fmt.Errorf("create ssl cert dir: %w", err)
	}

	certPath := filepath.Join(certDir, "fullchain.pem")
	keyPath := filepath.Join(certDir, "privkey.pem")

	if err := os.WriteFile(certPath, []byte(certContent), 0644); err != nil {
		return nil, fmt.Errorf("write cert file: %w", err)
	}
	if err := os.WriteFile(keyPath, []byte(keyContent), 0600); err != nil {
		return nil, fmt.Errorf("write key file: %w", err)
	}

	// 更新 Site SSL 字段
	site.SSLEnabled = true
	site.SSLProvider = "manual"
	site.SSLCertPath = certPath
	site.SSLKeyPath = keyPath
	site.SSLExpireAt = expireAt
	site.SSLAutoRenew = false

	if err := s.db.Save(site).Error; err != nil {
		return nil, fmt.Errorf("update site ssl fields: %w", err)
	}

	// 重新生成 Nginx 配置（含 SSL）
	confContent, err := nginx.Generate(site)
	if err != nil {
		return nil, fmt.Errorf("generate nginx config: %w", err)
	}
	if err := os.WriteFile(site.NginxConf, []byte(confContent), 0644); err != nil {
		return nil, fmt.Errorf("write nginx config: %w", err)
	}

	// 重载 Nginx
	if err := s.reloadNginx(context.Background()); err != nil {
		fmt.Printf("[site-service] nginx reload after ssl apply failed: %v\n", err)
	}

	_ = issuer // issuer 已在 GetSSLInfo 中使用
	return site, nil
}

// ApplyLetsEncrypt 申请 Let's Encrypt 证书（MVP 阶段模拟执行）
func (s *Service) ApplyLetsEncrypt(siteID uint, email string) (*model.Site, string, error) {
	site, err := s.GetSite(siteID)
	if err != nil {
		return nil, "", err
	}

	if email == "" {
		return nil, "", errors.New("email is required for Let's Encrypt")
	}

	// MVP 阶段：模拟 certbot 执行，记录日志，返回待完善状态
	// 生产环境应通过 CommandExecutor 执行 certbot 命令
	siteRoot := siteRootDir(site.PrimaryDomain)

	logMsg := fmt.Sprintf(
		"[Let's Encrypt] MVP 模拟执行: certbot certonly --webroot -w %s -d %s --email %s --non-interactive --agree-tos\n",
		siteRoot, site.PrimaryDomain, email,
	)
	logMsg += "[Let's Encrypt] 注意：当前为 MVP 阶段，证书申请功能待完善，请使用手动上传证书方式。\n"

	// 写入申请日志
	certDir := sslCertDir(site.PrimaryDomain)
	os.MkdirAll(certDir, 0755)
	logPath := filepath.Join(certDir, "letsencrypt.log")
	os.WriteFile(logPath, []byte(logMsg), 0644)

	// 更新状态为待完善
	site.SSLProvider = "letsencrypt"
	site.SSLAutoRenew = true
	s.db.Save(site)

	status := "pending_implementation"
	return site, status, nil
}

// GetSSLInfo 获取 SSL 证书信息
func (s *Service) GetSSLInfo(siteID uint) (*dto.SiteSSLInfo, error) {
	site, err := s.GetSite(siteID)
	if err != nil {
		return nil, err
	}

	info := &dto.SiteSSLInfo{
		Enabled:   site.SSLEnabled,
		Provider:  site.SSLProvider,
		CertPath:  site.SSLCertPath,
		KeyPath:   site.SSLKeyPath,
		AutoRenew: site.SSLAutoRenew,
	}

	if site.SSLExpireAt != nil {
		info.ExpireAt = site.SSLExpireAt.Format(time.RFC3339)
		info.DaysLeft = int(time.Until(*site.SSLExpireAt).Hours() / 24)
	}

	// 尝试从证书文件解析颁发者
	if site.SSLCertPath != "" {
		if certData, err := os.ReadFile(site.SSLCertPath); err == nil {
			if _, issuer, err := parseCertInfo(string(certData)); err == nil {
				info.Issuer = issuer
			}
		}
	}

	return info, nil
}

// parseCertInfo 解析 PEM 证书，返回有效期和颁发者
func parseCertInfo(certPEM string) (*time.Time, string, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, "", errors.New("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, "", fmt.Errorf("parse certificate: %w", err)
	}

	expireAt := cert.NotAfter
	issuer := cert.Issuer.CommonName
	if issuer == "" {
		issuer = cert.Issuer.Organization[0]
		if len(cert.Issuer.Organization) == 0 {
			issuer = "Unknown"
		}
	}

	return &expireAt, issuer, nil
}
