package nginx

import (
	"strings"
	"testing"

	"cloudpivot/internal/model"
)

// TestGenerate_Static 测试静态站点配置生成
func TestGenerate_Static(t *testing.T) {
	site := &model.Site{
		PrimaryDomain: "example.com",
		Type:          "static",
		RootDir:       "/www/example.com",
	}

	config, err := Generate(site)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(config, "server_name example.com") {
		t.Error("config should contain server_name")
	}
	if !strings.Contains(config, "listen 80") {
		t.Error("config should contain listen 80")
	}
	if !strings.Contains(config, "root /www/example.com") {
		t.Error("config should contain root directive")
	}
	if !strings.Contains(config, "try_files $uri $uri/ =404") {
		t.Error("static config should contain try_files")
	}
}

// TestGenerate_PHPRuntime 测试 PHP 运行时站点配置生成
func TestGenerate_PHPRuntime(t *testing.T) {
	site := &model.Site{
		PrimaryDomain: "php.example.com",
		Type:          "runtime",
		Runtime:       "php",
		RootDir:       "/www/php.example.com",
	}

	config, err := Generate(site)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(config, "fastcgi_pass 127.0.0.1:9000") {
		t.Error("PHP config should contain fastcgi_pass")
	}
	if !strings.Contains(config, "index index.php index.html") {
		t.Error("PHP config should contain index index.php")
	}
	if !strings.Contains(config, `location ~ \.php$`) {
		t.Error("PHP config should contain php location block")
	}
}

// TestGenerate_NodeRuntime 测试 Node 运行时站点配置生成
func TestGenerate_NodeRuntime(t *testing.T) {
	site := &model.Site{
		PrimaryDomain: "node.example.com",
		Type:          "runtime",
		Runtime:       "node",
	}

	config, err := Generate(site)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(config, "proxy_pass http://127.0.0.1:3000") {
		t.Error("Node config should contain proxy_pass to port 3000")
	}
	if !strings.Contains(config, "proxy_set_header Host $host") {
		t.Error("config should contain proxy_set_header Host")
	}
}

// TestGenerate_ReverseProxy 测试反向代理站点配置生成
func TestGenerate_ReverseProxy(t *testing.T) {
	site := &model.Site{
		PrimaryDomain: "proxy.example.com",
		Type:          "reverse_proxy",
		RootDir:       "http://127.0.0.1:8080",
	}

	config, err := Generate(site)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(config, "proxy_pass http://127.0.0.1:8080") {
		t.Error("reverse proxy config should contain proxy_pass")
	}
	if !strings.Contains(config, "proxy_http_version 1.1") {
		t.Error("reverse proxy config should contain HTTP/1.1 for WebSocket")
	}
	if !strings.Contains(config, `proxy_set_header Upgrade $http_upgrade`) {
		t.Error("reverse proxy config should contain WebSocket upgrade header")
	}
}

// TestGenerate_SSL 测试 SSL 配置生成
func TestGenerate_SSL(t *testing.T) {
	site := &model.Site{
		PrimaryDomain: "ssl.example.com",
		Type:          "static",
		SSLEnabled:    true,
		SSLCertPath:   "/etc/ssl/cert.pem",
		SSLKeyPath:    "/etc/ssl/key.pem",
	}

	config, err := Generate(site)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(config, "listen 443 ssl http2") {
		t.Error("SSL config should contain listen 443 ssl http2")
	}
	if !strings.Contains(config, "ssl_certificate /etc/ssl/cert.pem") {
		t.Error("SSL config should contain ssl_certificate")
	}
	if !strings.Contains(config, "return 301 https://$host$request_uri") {
		t.Error("SSL config should contain HTTP to HTTPS redirect")
	}
}

// TestGenerate_NilSite 测试 nil site 返回错误
func TestGenerate_NilSite(t *testing.T) {
	_, err := Generate(nil)
	if err == nil {
		t.Error("expected error for nil site")
	}
}

// TestGenerate_EmptyDomain 测试空域名返回错误
func TestGenerate_EmptyDomain(t *testing.T) {
	site := &model.Site{Type: "static"}
	_, err := Generate(site)
	if err == nil {
		t.Error("expected error for empty primary_domain")
	}
}

// TestValidate_ValidConfig 测试合法配置通过校验
func TestValidate_ValidConfig(t *testing.T) {
	config := `server {
    listen 80;
    server_name example.com;
    root /www/example.com;
}`
	if err := Validate(config); err != nil {
		t.Errorf("valid config should pass validation: %v", err)
	}
}

// TestValidate_EmptyConfig 测试空配置报错
func TestValidate_EmptyConfig(t *testing.T) {
	if err := Validate(""); err == nil {
		t.Error("empty config should return error")
	}
	if err := Validate("   "); err == nil {
		t.Error("whitespace-only config should return error")
	}
}

// TestValidate_UnbalancedBraces 测试括号不匹配报错
func TestValidate_UnbalancedBraces(t *testing.T) {
	// 缺少右括号
	config := `server {
    listen 80;
    server_name example.com;
`
	if err := Validate(config); err == nil {
		t.Error("config with unclosed brace should return error")
	}

	// 多余右括号
	config2 := `server {
    listen 80;
}}`
	if err := Validate(config2); err == nil {
		t.Error("config with extra closing brace should return error")
	}
}

// TestValidate_MissingDirective 测试缺少必需指令
func TestValidate_MissingDirective(t *testing.T) {
	// 缺少 server_name
	config := `server {
    listen 80;
    root /www/example.com;
}`
	if err := Validate(config); err == nil {
		t.Error("config missing server_name should return error")
	}
}
