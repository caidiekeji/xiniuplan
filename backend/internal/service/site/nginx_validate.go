package site

import (
	"fmt"
	"regexp"
	"strings"
)

// --- nginx 配置内容安全校验 ---

// dangerousNginxDirectives 危险指令黑名单：
// 面板用户提交的 nginx 配置/片段会写入 conf.d 并被 reload 加载，
// 以下指令可导致任意文件读取（include/alias 指向敏感路径）、
// 任意代码执行（load_module / lua_*）、或破坏 nginx 主配置上下文。
var dangerousNginxDirectives = []struct {
	pattern *regexp.Regexp
	reason  string
}{
	{regexp.MustCompile(`(?mi)^\s*include\s`), "include 可加载任意文件为配置（改用面板片段机制）"},
	{regexp.MustCompile(`(?mi)\binclude\s+["']?/`), "include 绝对路径禁止"},
	{regexp.MustCompile(`(?mi)^\s*load_module\s`), "load_module 可加载动态模块，存在代码执行风险"},
	{regexp.MustCompile(`(?mi)\blua_(load_package|code_cache)\b`), "lua 指令存在代码执行风险"},
	{regexp.MustCompile(`(?mi)^\s*(user|pid|daemon|master_process|worker_processes|worker_rlimit)\s`), "主配置级指令不允许出现在站点配置中"},
	{regexp.MustCompile(`(?mi)^\s*events\s*\{`), "events 块属于主配置上下文"},
	{regexp.MustCompile(`(?mi)^\s*http\s*\{`), "http 块属于主配置上下文"},
	{regexp.MustCompile(`(?mi)\berror_log\s+/dev/`), "error_log 不允许指向设备文件"},
	{regexp.MustCompile(`(?mi)\baccess_log\s*\|`), "access_log 管道写法存在命令执行风险"},
	{regexp.MustCompile(`(?mi)\bperl_set\b|\bjs_set\b.*require\(`), "脚本类指令存在代码执行风险"},
}

// ValidateNginxContent 校验用户提交的 nginx 配置内容（整段 server 配置或片段）。
// 返回 error 即拒绝写入 conf.d。
func ValidateNginxContent(content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("配置内容不能为空")
	}
	// 长度上限，防止超长配置拖垮 nginx reload
	if len(content) > 256*1024 {
		return fmt.Errorf("配置内容超过 256KB 上限")
	}
	for _, rule := range dangerousNginxDirectives {
		if rule.pattern.MatchString(content) {
			return fmt.Errorf("配置包含禁止的指令: %s", rule.reason)
		}
	}
	// 花括号必须配对，防止残缺配置导致 nginx reload 失败（服务中断）
	if open, close := strings.Count(content, "{"), strings.Count(content, "}"); open != close {
		return fmt.Errorf("配置花括号不配对: { %d 个, } %d 个", open, close)
	}
	return nil
}
