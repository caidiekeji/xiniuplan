package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"cloudpivot/internal/model"
)

// securityEntryCache 安全入口设置的本地缓存，避免每个请求都查库
type securityEntryCache struct {
	mu       sync.RWMutex
	value    string
	loadedAt time.Time
	ttl      time.Duration
}

var entryCache = &securityEntryCache{ttl: 30 * time.Second}

// GetSecurityEntry 读取安全入口设置（带 30s 缓存，设置变更最迟 30s 生效）
func GetSecurityEntry() string {
	entryCache.mu.RLock()
	if time.Since(entryCache.loadedAt) < entryCache.ttl {
		v := entryCache.value
		entryCache.mu.RUnlock()
		return v
	}
	entryCache.mu.RUnlock()

	entryCache.mu.Lock()
	defer entryCache.mu.Unlock()
	// 双重检查：拿写锁期间可能已被其他请求刷新
	if time.Since(entryCache.loadedAt) < entryCache.ttl {
		return entryCache.value
	}

	v := ""
	if model.DB != nil {
		var setting model.Setting
		if err := model.DB.Where("`key` = ?", "security_entry").First(&setting).Error; err == nil {
			v = strings.Trim(strings.TrimSpace(setting.Value), "/")
		}
	}
	entryCache.value = v
	entryCache.loadedAt = time.Now()
	if v != "" {
		slog.Info("[安全入口] 已启用", "entry", "/"+v)
	}
	return v
}

// invalidateSecurityEntryCache 设置变更后主动失效缓存
func InvalidateSecurityEntryCache() {
	entryCache.mu.Lock()
	entryCache.loadedAt = time.Time{}
	entryCache.mu.Unlock()
}

// SecurityEntryWrapper 安全入口包装层。
// 必须包在 gin Engine 外层（gin 的 r.Use 中间件在路由匹配之后才执行，
// 无法在匹配前改写路径）。
//
// 未配置入口：所有请求原样放行。
// 已配置入口（如 admin-xyz）：
//   - /{entry} 与 /{entry}/... → 剥离前缀后交给 gin（SPA 静态文件由 nginx 兜底）
//   - /health                  → 始终放行（探活）
//   - 其余路径（含裸 /api/v1） → 404，避免暴露面板存在
func SecurityEntryWrapper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entry := GetSecurityEntry()
		if entry == "" {
			next.ServeHTTP(w, r)
			return
		}

		path := r.URL.Path

		// 健康检查始终放行
		if path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// 通过安全入口访问：剥离前缀
		if path == "/"+entry || strings.HasPrefix(path, "/"+entry+"/") {
			trimmed := strings.TrimPrefix(path, "/"+entry)
			if trimmed == "" {
				trimmed = "/"
			}
			r.URL.Path = trimmed
			next.ServeHTTP(w, r)
			return
		}

		// 未通过安全入口：伪装成不存在的站点
		http.NotFound(w, r)
	})
}
