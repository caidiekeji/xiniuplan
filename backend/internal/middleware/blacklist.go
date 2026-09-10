package middleware

import (
	"sync"
	"time"
)

// tokenBlacklist 内存 token 黑名单（MVP 阶段）
var (
	blacklist   = make(map[string]time.Time)
	blacklistMu sync.RWMutex
)

// BlacklistToken 将 token 加入黑名单
func BlacklistToken(token string, expireAt time.Time) {
	blacklistMu.Lock()
	defer blacklistMu.Unlock()
	blacklist[token] = expireAt
}

// IsTokenBlacklisted 检查 token 是否在黑名单
func IsTokenBlacklisted(token string) bool {
	blacklistMu.RLock()
	defer blacklistMu.RUnlock()
	expireAt, exists := blacklist[token]
	if !exists {
		return false
	}
	// 已过期的自动清理
	if time.Now().After(expireAt) {
		go func() {
			blacklistMu.Lock()
			delete(blacklist, token)
			blacklistMu.Unlock()
		}()
		return false
	}
	return true
}

// CleanExpiredBlacklist 清理过期黑名单条目
func CleanExpiredBlacklist() {
	blacklistMu.Lock()
	defer blacklistMu.Unlock()
	now := time.Now()
	for token, expireAt := range blacklist {
		if now.After(expireAt) {
			delete(blacklist, token)
		}
	}
}
