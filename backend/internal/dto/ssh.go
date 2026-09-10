package dto

// SSHConfig SSH 服务配置
type SSHConfig struct {
	Port                int    `json:"port"`
	PermitRootLogin     string `json:"permit_root_login"`     // yes/no/without-password/prohibit-password
	PermitPasswordLogin string `json:"permit_password_login"` // yes/no
	MaxAuthTries        int    `json:"max_auth_tries"`
	ClientAliveInterval int    `json:"client_alive_interval"`
	AllowUsers          string `json:"allow_users"` // 空格分隔的用户列表
}

// UpdateSSHConfigRequest 修改 SSH 配置请求
type UpdateSSHConfigRequest struct {
	Port                int    `json:"port"`
	PermitRootLogin     string `json:"permit_root_login"`
	PermitPasswordLogin string `json:"permit_password_login"`
	MaxAuthTries        int    `json:"max_auth_tries"`
	ClientAliveInterval int    `json:"client_alive_interval"`
	AllowUsers          string `json:"allow_users"`
}

// SSHKey SSH 公钥信息
type SSHKey struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`        // rsa/ed25519/ecdsa
	Comment     string `json:"comment"`     // 密钥备注
	Fingerprint string `json:"fingerprint"` // SHA256:xxxx
	CreatedAt   string `json:"created_at"`  // 文件修改时间近似值
}

// AddSSHKeyRequest 添加公钥请求
type AddSSHKeyRequest struct {
	PublicKey string `json:"public_key" binding:"required"`
}

// SSHLoginLog SSH 登录日志
type SSHLoginLog struct {
	Time   string `json:"time"`
	User   string `json:"user"`
	IP     string `json:"ip"`
	Status string `json:"status"` // success/failed
	Method string `json:"method"` // password/publickey/keyboard-interactive
}

// Fail2banConfig Fail2ban 配置
type Fail2banConfig struct {
	Enabled  bool `json:"enabled"`
	MaxRetry int  `json:"max_retry"`
	BanTime  int  `json:"ban_time"`  // 秒
	FindTime int  `json:"find_time"` // 秒
}

// UpdateFail2banConfigRequest 修改 Fail2ban 配置请求
type UpdateFail2banConfigRequest struct {
	Enabled  bool `json:"enabled"`
	MaxRetry int  `json:"max_retry"`
	BanTime  int  `json:"ban_time"`
	FindTime int  `json:"find_time"`
}

// BannedIP 被封禁 IP
type BannedIP struct {
	IP       string `json:"ip"`
	Jail     string `json:"jail"`
	BannedAt string `json:"banned_at"`
	Reason   string `json:"reason"`
}

// Fail2banStatus Fail2ban 状态
type Fail2banStatus struct {
	Installed       bool       `json:"installed"`
	Enabled         bool       `json:"enabled"`
	CurrentlyBanned int        `json:"currently_banned"`
	TotalBanned     int        `json:"total_banned"`
	CurrentlyFailed int        `json:"currently_failed"`
	TotalFailed     int        `json:"total_failed"`
	BannedIPs       []BannedIP `json:"banned_ips"`
}

// IPActionRequest IP 操作请求（封禁/解封）
type IPActionRequest struct {
	IP string `json:"ip" binding:"required"`
}
