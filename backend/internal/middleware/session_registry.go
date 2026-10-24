package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 会话操作错误
var (
	errSessionNotFound  = errors.New("会话不存在或已过期")
	errSessionForbidden = errors.New("无权踢出他人的会话")
)

// SessionInfo 在线会话信息（U-006 会话管理）
type SessionInfo struct {
	TokenHash  string    `json:"token_hash"` // SHA-256(token)，用作踢出操作的会话 ID
	UserID     uint      `json:"user_id"`
	Username   string    `json:"username"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	TokenType  string    `json:"token_type"` // access | refresh
	LoginAt    time.Time `json:"login_at"`
	LastActive time.Time `json:"last_active"`
	ExpiresAt  time.Time `json:"expires_at"`
}

var (
	sessionMu   sync.RWMutex
	sessionByID = make(map[string]*sessionEntry) // tokenHash -> entry
)

type sessionEntry struct {
	info       SessionInfo
	rawToken   string // 仅驻留内存，用于踢出时加入黑名单
	lastStored time.Time
}

// hashToken 计算令牌指纹（不落明文）
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// RegisterSession 在 JWT 校验通过后登记会话（幂等，同一 token 只更新活跃时间）
func RegisterSession(c *gin.Context, token string, userID uint, username, tokenType string, expiresAt time.Time) {
	hash := hashToken(token)
	now := time.Now()

	sessionMu.Lock()
	defer sessionMu.Unlock()

	if e, ok := sessionByID[hash]; ok {
		e.info.LastActive = now
		return
	}
	// 单用户会话上限保护：清理过期会话，防止内存膨胀
	if len(sessionByID) > 10000 {
		cleanSessionsLocked()
	}
	sessionByID[hash] = &sessionEntry{
		info: SessionInfo{
			TokenHash:  hash,
			UserID:     userID,
			Username:   username,
			IP:         c.ClientIP(),
			UserAgent:  c.GetHeader("User-Agent"),
			TokenType:  tokenType,
			LoginAt:    now,
			LastActive: now,
			ExpiresAt:  expiresAt,
		},
		rawToken:   token,
		lastStored: now,
	}
}

// ListSessions 列出在线会话。userID>0 时只返回该用户的会话，0 返回全部（管理员）
func ListSessions(userID uint) []SessionInfo {
	now := time.Now()
	result := make([]SessionInfo, 0)
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	for _, e := range sessionByID {
		if now.After(e.info.ExpiresAt) {
			continue // 过期会话不展示
		}
		if userID > 0 && e.info.UserID != userID {
			continue
		}
		result = append(result, e.info)
	}
	return result
}

// RevokeSession 踢出会话：将底层 token 加入黑名单并从注册表移除
// callerID > 0 时校验只能踢自己的会话（管理员传 0 或自身校验在 handler 做）
func RevokeSession(tokenHash string, callerID uint) error {
	sessionMu.Lock()
	e, ok := sessionByID[tokenHash]
	if !ok {
		sessionMu.Unlock()
		return errSessionNotFound
	}
	if callerID > 0 && e.info.UserID != callerID {
		sessionMu.Unlock()
		return errSessionForbidden
	}
	raw := e.rawToken
	exp := e.info.ExpiresAt
	delete(sessionByID, tokenHash)
	sessionMu.Unlock()

	BlacklistToken(raw, exp)
	return nil
}

// KickUserSessions 踢出某用户的全部会话（用户被禁用/删除时调用）
func KickUserSessions(userID uint) int {
	sessionMu.Lock()
	var tokens []struct {
		raw string
		exp time.Time
	}
	for hash, e := range sessionByID {
		if e.info.UserID == userID {
			tokens = append(tokens, struct {
				raw string
				exp time.Time
			}{e.rawToken, e.info.ExpiresAt})
			delete(sessionByID, hash)
		}
	}
	sessionMu.Unlock()

	for _, t := range tokens {
		BlacklistToken(t.raw, t.exp)
	}
	return len(tokens)
}

// RemoveSessionFromRegistry 仅从注册表移除（登出时），同时清黑名单由调用方处理
func RemoveSessionFromRegistry(token string) {
	hash := hashToken(token)
	sessionMu.Lock()
	delete(sessionByID, hash)
	sessionMu.Unlock()
}

func cleanSessionsLocked() {
	now := time.Now()
	for hash, e := range sessionByID {
		if now.After(e.info.ExpiresAt) {
			delete(sessionByID, hash)
		}
	}
}

// CleanExpiredSessions 定期清理过期会话
func CleanExpiredSessions() {
	sessionMu.Lock()
	cleanSessionsLocked()
	sessionMu.Unlock()
}
