package nginx

import (
	"fmt"
	"strings"

	"cloudpivot/internal/model"
)

// WAF 基础版（roadmap #10）：nginx 层 CC 防护 + UA 黑名单
// 原理：limit_req_zone 定义在 http 上下文（conf.d 由 nginx include 在 http{} 内），
// server 块内引用 zone 并对 UA 黑名单命中直接 403。

// GenerateWAFZones 生成 http 上下文的 WAF zone 定义文件内容（00-waf-zones.conf）
// 所有开启 WAF 的站点共用一个 UA 黑名单 map（合并去重），每个站点独立一个限速 zone
func GenerateWAFZones(sites []model.Site) (string, bool) {
	uaSet := map[string]bool{}
	hasWAF := false
	var zones strings.Builder

	for _, site := range sites {
		if !site.WafEnabled || site.PrimaryDomain == "" {
			continue
		}
		hasWAF = true
		zoneName := wafZoneName(site.PrimaryDomain)
		rate := site.CCRpm
		if rate <= 0 {
			rate = 120
		}
		zones.WriteString(fmt.Sprintf("limit_req_zone $binary_remote_addr zone=%s:10m rate=%dr/s;\n",
			zoneName, rate))
		for _, ua := range strings.Split(site.UABlacklist, "\n") {
			ua = strings.TrimSpace(ua)
			if ua != "" {
				uaSet[strings.ToLower(ua)] = true
			}
		}
	}
	if !hasWAF {
		return "", false
	}

	var buf strings.Builder
	buf.WriteString("# 由面板自动生成（WAF zone 定义），请勿手动编辑\n")
	buf.WriteString(zones.String())

	if len(uaSet) > 0 {
		buf.WriteString("map $http_user_agent $waf_bad_ua {\n    default 0;\n")
		for ua := range uaSet {
			// map 值为正则时自动支持关键词匹配
			buf.WriteString(fmt.Sprintf("    ~*%s 1;\n", strings.ReplaceAll(ua, " ", "")))
		}
		buf.WriteString("}\n")
	}
	return buf.String(), true
}

// BuildWAFServerDirectives 生成 server 块内的 WAF 指令（limit_req + UA 拒绝）
func BuildWAFServerDirectives(site *model.Site) string {
	if !site.WafEnabled {
		return ""
	}
	zoneName := wafZoneName(site.PrimaryDomain)
	rate := site.CCRpm
	if rate <= 0 {
		rate = 120
	}
	var buf strings.Builder
	buf.WriteString("\n    # WAF: CC 防护 + UA 黑名单（面板生成）\n")
	if rate > 0 {
		burst := rate / 6 // 突发容量 = 10 秒配额
		if burst < 10 {
			burst = 10
		}
		buf.WriteString(fmt.Sprintf("    limit_req zone=%s burst=%d nodelay;\n", zoneName, burst))
		buf.WriteString(fmt.Sprintf("    limit_req_status 429;\n"))
	}
	// UA 黑名单需要 site 有自定义名单（全局 map 在 zones 文件中）
	if strings.TrimSpace(site.UABlacklist) != "" {
		buf.WriteString("    if ($waf_bad_ua) { return 403; }\n")
	}
	return buf.String()
}

func wafZoneName(domain string) string {
	// zone 名只能含字母数字下划线点：域名中的 . - 替换为 _
	replacer := strings.NewReplacer(".", "_", "-", "_", ":", "_")
	return "waf_" + replacer.Replace(domain)
}
