package nginx

import (
	"fmt"
	"runtime"
	"strings"

	"cloudpivot/internal/model"
)

// SnippetDir 站点配置片段目录（server 块内 include 的 *.conf）
// 注意：与服务层 nginxConfDir() 保持一致的路径约定
func SnippetDir(domain string) string {
	base := "/opt/cloudpivot/runtime/nginx/conf.d"
	if runtime.GOOS == "windows" {
		base = "./runtime/nginx/conf.d"
	}
	return base + "/" + domain + ".d"
}

// snippetIncludeInclude 生成 server 块内的 include 指令
func buildSnippetInclude(site *model.Site) string {
	if site.PrimaryDomain == "" {
		return ""
	}
	return fmt.Sprintf("\n    # 云枢面板：站点增强配置片段（伪静态/防盗链/重定向/访问控制）\n    include %s/*.conf;\n", SnippetDir(site.PrimaryDomain))
}

// RewriteTemplates 内置伪静态规则模板（W-006）
var RewriteTemplates = map[string]string{
	"wordpress": `location / {
    try_files $uri $uri/ /index.php?$args;
}
rewrite /wp-admin$ $scheme://$host$uri/ permanent;`,
	"typecho": `location / {
    index index.html index.php;
    if (!-e $request_filename) {
        rewrite ^(.*)$ /index.php$1 last;
    }
}`,
	"thinkphp": `location / {
    if (!-e $request_filename) {
        rewrite ^(.*)$ /index.php?s=$1 last;
        break;
    }
}`,
	"laravel": `location / {
    try_files $uri $uri/ /index.php?$query_string;
}`,
	"discuz": `location / {
    rewrite ^([^\.]*)/topic-(.+)\.html$ $1/portal.php?mod=topic&topic=$2 last;
    rewrite ^([^\.]*)/article-([0-9]+)-([0-9]+)\.html$ $1/portal.php?mod=view&aid=$2&page=$3 last;
    rewrite ^([^\.]*)/forum-(\w+)-([0-9]+)\.html$ $1/forum.php?mod=forumdisplay&fid=$2&page=$3 last;
    rewrite ^([^\.]*)/thread-([0-9]+)-([0-9]+)-([0-9]+)\.html$ $1/forum.php?mod=viewthread&tid=$2&extra=page%3D$4&page=$3 last;
    rewrite ^([^\.]*)/group-([0-9]+)-([0-9]+)\.html$ $1/forum.php?mod=group&fid=$2&page=$3 last;
    rewrite ^([^\.]*)/blog-([0-9]+)-([0-9]+)\.html$ $1/home.php?mod=space&uid=$2&do=blog&id=$3 last;
    rewrite ^([^\.]*)/(fid|tid)-([0-9]+)\.html$ $1/index.php?action=$2&value=$3 last;
}`,
	"default": `location / {
    try_files $uri $uri/ /index.php?$query_string;
}`,
}

// BuildRewriteSnippet 伪静态片段（W-006）
func BuildRewriteSnippet(templateKey string) (string, error) {
	tpl, ok := RewriteTemplates[templateKey]
	if !ok {
		return "", fmt.Errorf("未知伪静态模板: %s", templateKey)
	}
	return "# 伪静态规则模板: " + templateKey + "\n" + tpl + "\n", nil
}

// BuildHotlinkSnippet 防盗链片段（W-007）：referer 白名单
func BuildHotlinkSnippet(allowedDomains []string, extensions string, allowEmpty bool) (string, error) {
	if len(allowedDomains) == 0 {
		return "", fmt.Errorf("至少指定一个允许的域名")
	}
	if extensions == "" {
		extensions = "jpg|jpeg|png|gif|webp|css|js|woff|woff2|ttf|svg|mp4"
	}
	var buf strings.Builder
	buf.WriteString("# 防盗链：referer 白名单\nlocation ~* \\.(" + extensions + ")$ {\n")
	buf.WriteString("    valid_referers")
	if allowEmpty {
		buf.WriteString(" none blocked")
	}
	for _, d := range allowedDomains {
		d = strings.TrimSpace(d)
		if d == "" || strings.ContainsAny(d, ";{}\"'\\") {
			continue
		}
		buf.WriteString(" " + d)
	}
	buf.WriteString(";\n")
	buf.WriteString("    if ($invalid_referer) {\n        return 403;\n    }\n}\n")
	return buf.String(), nil
}

// BuildRedirectSnippet 重定向规则片段（W-008）
// from 为路径前缀或 location 匹配，to 为目标地址，permanent=true 时 301 否则 302
func BuildRedirectSnippet(from, to string, permanent bool) (string, error) {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	if from == "" || to == "" {
		return "", fmt.Errorf("重定向来源和目标不能为空")
	}
	if strings.ContainsAny(from, "{}\"'\\") || strings.ContainsAny(to, "{}\"'\\") {
		return "", fmt.Errorf("重定向参数包含非法字符")
	}
	code := "302"
	if permanent {
		code = "301"
	}
	return fmt.Sprintf("# 重定向规则: %s -> %s\nrewrite ^%s(.*)$ %s$1 %s;\n", from, to, from, to, code), nil
}

// BuildAccessControlSnippet IP 黑白名单片段（W-009）
// mode: whitelist（仅允许列表内）| blacklist（拒绝列表内）
func BuildAccessControlSnippet(mode string, entries []string) (string, error) {
	if len(entries) == 0 {
		return "", fmt.Errorf("IP 列表不能为空")
	}
	var buf strings.Builder
	if mode == "whitelist" {
		buf.WriteString("# 访问控制：IP 白名单（仅允许列表内地址）\n")
		for _, e := range entries {
			e = strings.TrimSpace(e)
			if e == "" || strings.ContainsAny(e, ";{}\"'\\") {
				continue
			}
			buf.WriteString("allow " + e + ";\n")
		}
		buf.WriteString("deny all;\n")
	} else {
		buf.WriteString("# 访问控制：IP 黑名单\n")
		for _, e := range entries {
			e = strings.TrimSpace(e)
			if e == "" || strings.ContainsAny(e, ";{}\"'\\") {
				continue
			}
			buf.WriteString("deny " + e + ";\n")
		}
	}
	return buf.String(), nil
}

// BuildBasicAuthSnippet 基础认证片段（W-009）
func BuildBasicAuthSnippet(realm, htpasswdPath string) (string, error) {
	if htpasswdPath == "" {
		return "", fmt.Errorf("缺少 htpasswd 文件路径")
	}
	if realm == "" {
		realm = "Restricted"
	}
	return fmt.Sprintf("# 访问控制：Basic Auth\nauth_basic \"%s\";\nauth_basic_user_file %s;\n", realm, htpasswdPath), nil
}
