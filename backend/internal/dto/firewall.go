package dto

// CreateRuleRequest 创建防火墙规则请求
type CreateRuleRequest struct {
	Type      string `json:"type" binding:"required,oneof=port ip forward"`
	Direction string `json:"direction" binding:"required,oneof=inbound outbound"`
	Protocol  string `json:"protocol" binding:"required,oneof=tcp udp icmp all"`
	Port      int    `json:"port"`
	IPAddress string `json:"ip_address"`
	DestIP    string `json:"dest_ip"`   // type=forward：目标地址
	DestPort  int    `json:"dest_port"` // type=forward：目标端口
	Action    string `json:"action" binding:"required,oneof=accept drop"`
	Remark    string `json:"remark"`
}

// RuleResponse 防火墙规则响应
type RuleResponse struct {
	ID        uint   `json:"id"`
	Type      string `json:"type"`
	Direction string `json:"direction"`
	Protocol  string `json:"protocol"`
	Port      int    `json:"port"`
	IPAddress string `json:"ip_address"`
	DestIP    string `json:"dest_ip"`
	DestPort  int    `json:"dest_port"`
	Action    string `json:"action"`
	Status    string `json:"status"`
	Remark    string `json:"remark"`
	CreatedAt string `json:"created_at"`
}

// SecuritySuggestion 安全建议
type SecuritySuggestion struct {
	Port       int    `json:"port"`
	Service    string `json:"service"`
	RiskLevel  string `json:"risk_level"` // high | medium | low
	Suggestion string `json:"suggestion"`
}
