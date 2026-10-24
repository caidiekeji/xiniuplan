package nginx

import (
	"errors"
	"fmt"
	"strings"

	"cloudpivot/internal/model"
)

// SSLConfig SSL 配置
type SSLConfig struct {
	Enabled bool
	Cert    string
	Key     string
}

// Location Nginx location 块
type Location struct {
	Path      string
	ProxyPass string
	Root      string
	FastCGI   string
	TryFiles  string
}

// Config Nginx 站点配置
type Config struct {
	ServerName []string
	Root       string
	Index      string
	ProxyPass  string
	SSL        SSLConfig
	Locations  []Location
	Redirects  []string
	AccessLog  string
	ErrorLog   string
}

// runtime 默认端口映射
var runtimePorts = map[string]string{
	"node":   "3000",
	"python": "8000",
	"go":     "8080",
}

// Generate 根据 Site 模型生成完整 Nginx server 块配置
func Generate(site *model.Site) (string, error) {
	if site == nil {
		return "", errors.New("site is nil")
	}
	if site.PrimaryDomain == "" {
		return "", errors.New("primary_domain is empty")
	}

	serverNames := buildServerNames(site)
	accessLog := fmt.Sprintf("/www/%s/logs/access.log", site.PrimaryDomain)
	errorLog := fmt.Sprintf("/www/%s/logs/error.log", site.PrimaryDomain)

	var buf strings.Builder

	if site.SSLEnabled && site.SSLCertPath != "" && site.SSLKeyPath != "" {
		// 80 端口跳转到 443
		buf.WriteString(buildRedirectServer(serverNames))
		// 443 SSL server 块
		buf.WriteString(buildServerBlock(serverNames, "443 ssl http2", site, accessLog, errorLog, true))
	} else {
		buf.WriteString(buildServerBlock(serverNames, "80", site, accessLog, errorLog, false))
	}

	return buf.String(), nil
}

// buildServerNames 构建 server_name 指令值
func buildServerNames(site *model.Site) string {
	names := []string{site.PrimaryDomain}
	for _, d := range site.Domains {
		if d != "" && d != site.PrimaryDomain {
			names = append(names, d)
		}
	}
	return strings.Join(names, " ")
}

// buildRedirectServer 生成 80 端口跳转配置
func buildRedirectServer(serverNames string) string {
	return fmt.Sprintf(`server {
    listen 80;
    server_name %s;
    return 301 https://$host$request_uri;
}

`, serverNames)
}

// buildServerBlock 生成主 server 块
func buildServerBlock(serverNames, listen string, site *model.Site, accessLog, errorLog string, ssl bool) string {
	var buf strings.Builder

	buf.WriteString("server {\n")
	buf.WriteString(fmt.Sprintf("    listen %s;\n", listen))
	buf.WriteString(fmt.Sprintf("    server_name %s;\n", serverNames))

	if ssl {
		buf.WriteString(fmt.Sprintf("    ssl_certificate %s;\n", site.SSLCertPath))
		buf.WriteString(fmt.Sprintf("    ssl_certificate_key %s;\n", site.SSLKeyPath))
		buf.WriteString("    ssl_protocols TLSv1.2 TLSv1.3;\n")
		buf.WriteString("    ssl_ciphers HIGH:!aNULL:!MD5;\n")
		buf.WriteString("    ssl_prefer_server_ciphers on;\n")
	}

	buf.WriteString(fmt.Sprintf("    access_log %s;\n", accessLog))
	buf.WriteString(fmt.Sprintf("    error_log %s;\n", errorLog))

	// WAF（roadmap #10）：CC 防护 + UA 黑名单
	buf.WriteString(BuildWAFServerDirectives(site))

	// 站点增强配置片段（伪静态/防盗链/重定向/访问控制，W-006~009）
	buf.WriteString(buildSnippetInclude(site))

	// 根据类型生成 location 块
	switch site.Type {
	case "static":
		buf.WriteString(buildStaticLocations(site))
	case "runtime":
		buf.WriteString(buildRuntimeLocations(site))
	case "reverse_proxy":
		buf.WriteString(buildReverseProxyLocations(site))
	default:
		buf.WriteString(buildStaticLocations(site))
	}

	buf.WriteString("}\n")
	return buf.String()
}

// buildStaticLocations 静态站点 location
func buildStaticLocations(site *model.Site) string {
	root := site.RootDir
	if root == "" {
		root = fmt.Sprintf("/www/%s", site.PrimaryDomain)
	}
	return fmt.Sprintf(`
    root %s;
    index index.html;

    location / {
        try_files $uri $uri/ =404;
    }
`, root)
}

// buildRuntimeLocations 运行时站点 location
func buildRuntimeLocations(site *model.Site) string {
	root := site.RootDir
	if root == "" {
		root = fmt.Sprintf("/www/%s", site.PrimaryDomain)
	}

	var buf strings.Builder

	switch strings.ToLower(site.Runtime) {
	case "php":
		buf.WriteString(fmt.Sprintf(`
    root %s;
    index index.php index.html;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        fastcgi_pass 127.0.0.1:9000;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
`, root))

	case "node", "python", "go":
		port := runtimePorts[strings.ToLower(site.Runtime)]
		buf.WriteString(fmt.Sprintf(`
    location / {
        proxy_pass http://127.0.0.1:%s;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
`, port))

	default:
		// 未知 runtime，回退到静态
		buf.WriteString(buildStaticLocations(site))
	}

	return buf.String()
}

// buildReverseProxyLocations 反向代理 location（含 WebSocket 升级头）
func buildReverseProxyLocations(site *model.Site) string {
	// reverse_proxy 类型下，RootDir 字段存储代理目标地址
	target := site.RootDir
	if target == "" {
		target = "http://127.0.0.1:8080"
	}

	return fmt.Sprintf(`
    location / {
        proxy_pass %s;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket 升级
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 86400;
    }
`, target)
}

// Validate 基本语法检查：括号匹配、必需指令
func Validate(config string) error {
	if strings.TrimSpace(config) == "" {
		return errors.New("config is empty")
	}

	// 括号匹配检查
	var depth int
	for _, ch := range config {
		switch ch {
		case '{':
			depth++
		case '}':
			depth--
			if depth < 0 {
				return errors.New("unbalanced braces: unexpected '}'")
			}
		}
	}
	if depth != 0 {
		return fmt.Errorf("unbalanced braces: %d unclosed '{'", depth)
	}

	// 必需指令检查
	required := []string{"server_name", "listen"}
	for _, directive := range required {
		if !strings.Contains(config, directive+" ") && !strings.Contains(config, directive+"\t") && !strings.Contains(config, directive+";") {
			return fmt.Errorf("missing required directive: %s", directive)
		}
	}

	// 必须包含至少一个 server 块
	if !strings.Contains(config, "server") {
		return errors.New("missing server block")
	}

	return nil
}
