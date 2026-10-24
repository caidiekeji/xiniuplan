package terminal

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

// TerminalSession 表示一个 WebSocket + SSH 终端会话
type TerminalSession struct {
	ID         string
	Conn       *websocket.Conn
	SSHClient  *ssh.Client
	SSHSession *ssh.Session
	Stdin      io.WriteCloser
	// RawMode（T-005 ZMODEM）：true 时输出以 BinaryMessage 原样转发，不做 JSON 封装
	RawMode bool
	// RecordFile（T-006 会话录制）：非 nil 时输出同步写入录制文件
	RecordFile io.WriteCloser
	mu         sync.Mutex
	closed     bool
	lastActive time.Time
}

// SessionManager 管理所有终端会话
type SessionManager struct {
	sessions map[string]*TerminalSession
	mu       sync.RWMutex
}

// NewSessionManager 创建会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*TerminalSession),
	}
}

// generateID 生成随机会话 ID
func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Create 创建新会话（仅注册，SSH 连接由调用方建立后注入）
func (sm *SessionManager) Create() *TerminalSession {
	s := &TerminalSession{
		ID:         generateID(),
		lastActive: time.Now(),
	}
	sm.mu.Lock()
	sm.sessions[s.ID] = s
	sm.mu.Unlock()
	return s
}

// Get 根据 ID 获取会话
func (sm *SessionManager) Get(id string) (*TerminalSession, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	s, ok := sm.sessions[id]
	return s, ok
}

// Close 关闭并清理会话
func (sm *SessionManager) Close(id string) {
	sm.mu.Lock()
	s, ok := sm.sessions[id]
	if ok {
		delete(sm.sessions, id)
	}
	sm.mu.Unlock()

	if ok {
		s.close()
	}
}

// close 关闭会话底层资源
func (s *TerminalSession) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true

	if s.SSHSession != nil {
		_ = s.SSHSession.Close()
	}
	if s.Stdin != nil {
		_ = s.Stdin.Close()
	}
	if s.SSHClient != nil {
		_ = s.SSHClient.Close()
	}
	if s.Conn != nil {
		_ = s.Conn.Close()
	}
}

// IsClosed 检查会话是否已关闭
func (s *TerminalSession) IsClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

// Lock 加锁（供外部包在操作会话字段时使用）
func (s *TerminalSession) Lock() {
	s.mu.Lock()
}

// Unlock 解锁
func (s *TerminalSession) Unlock() {
	s.mu.Unlock()
}

// SetConn 设置 WebSocket 连接（线程安全）
func (s *TerminalSession) SetConn(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Conn != nil {
		_ = s.Conn.Close()
	}
	s.Conn = conn
}

// Touch 更新最后活跃时间
func (s *TerminalSession) Touch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastActive = time.Now()
}

// LastActive 获取最后活跃时间
func (s *TerminalSession) LastActive() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastActive
}

// CleanupStale 清理超过 timeout 未活跃的会话
func (sm *SessionManager) CleanupStale(timeout time.Duration) {
	cutoff := time.Now().Add(-timeout)
	var stale []string

	sm.mu.RLock()
	for id, s := range sm.sessions {
		if s.LastActive().Before(cutoff) {
			stale = append(stale, id)
		}
	}
	sm.mu.RUnlock()

	for _, id := range stale {
		sm.Close(id)
	}
}

// StartCleanupLoop 启动后台定时清理
func (sm *SessionManager) StartCleanupLoop(interval, timeout time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			sm.CleanupStale(timeout)
		}
	}()
}

// Count 返回当前活跃会话数
func (sm *SessionManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// String 用于日志
func (s *TerminalSession) String() string {
	return fmt.Sprintf("TerminalSession(%s)", s.ID)
}
