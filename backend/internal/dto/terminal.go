package dto

// TerminalMessage WebSocket 终端消息协议
type TerminalMessage struct {
	Type string `json:"type"` // input | resize | output | close | error
	Data string `json:"data,omitempty"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
}

// CreateSessionRequest 创建 SSH 会话请求
type CreateSessionRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	KeyPath  string `json:"key_path,omitempty"`
}

// SessionInfo 终端会话信息
type SessionInfo struct {
	SessionID string `json:"session_id"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
}
